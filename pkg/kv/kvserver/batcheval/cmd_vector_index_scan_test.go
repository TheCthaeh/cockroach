// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package batcheval

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv/kvpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/vecindex/quantize"
	"github.com/cockroachdb/cockroach/pkg/sql/vecindex/testutils"
	"github.com/cockroachdb/cockroach/pkg/sql/vecindex/vecstore"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/vector"
	"github.com/stretchr/testify/require"
)

func TestVectorIndexScan(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	settings := cluster.MakeTestingClusterSettings()
	eng := storage.NewDefaultInMemForTesting()
	defer eng.Close()

	rnd, seed := randutil.NewTestRand()
	t.Logf("random seed: %v", seed)
	dims := rnd.Intn(100) + 1
	set := testutils.RandomVectorSet(rnd, dims)
	indexSeed := rnd.Int63()
	for _, quantizer := range []quantize.Quantizer{
		quantize.NewUnQuantizer(dims),
		quantize.NewRaBitQuantizer(dims, indexSeed),
	} {
		name := strings.TrimPrefix(fmt.Sprintf("%T", quantizer), "*quantize.")
		quantizedSet := quantizer.Quantize(ctx, set)
		t.Run(name, func(t *testing.T) {
			nonLeafLevel := vecstore.Level(rnd.Intn(10)) + vecstore.SecondLevel
			for _, level := range []vecstore.Level{vecstore.LeafLevel, nonLeafLevel} {
				t.Run(fmt.Sprintf("level=%d", level), func(t *testing.T) {
					var tsCounter int64
					ts := func() hlc.Timestamp {
						tsCounter++
						return hlc.Timestamp{WallTime: tsCounter}
					}
					startKey := putQuantizedVectorSet(t, ctx, rnd, eng, quantizedSet, level, ts)
					endKey := startKey.PrefixEnd()
					quantizeMethod := quantize.MethodNotQuantized
					if _, ok := quantizedSet.(*quantize.RaBitQuantizedVectorSet); ok {
						quantizeMethod = quantize.MethodRaBitQ
					}
					for range 10 {
						// Generate a random query vector and perform a vector index scan.
						queryVector := vector.Random(rnd, dims)
						req := &kvpb.VectorIndexScanRequest{
							QuantizeMethod: quantizeMethod,
							Dims:           uint64(dims),
							Seed:           indexSeed,
							QueryVector:    queryVector,
						}
						req.SetHeader(kvpb.RequestHeader{Key: startKey, EndKey: endKey})
						resp := &kvpb.VectorIndexScanResponse{}
						cArgs := CommandArgs{
							Args:    req,
							Header:  kvpb.Header{Timestamp: ts()},
							EvalCtx: (&MockEvalCtx{ClusterSettings: settings}).EvalContext(),
						}
						_, err := VectorIndexScan(ctx, eng, cArgs, resp)
						require.NoError(t, err)

						// Verify the response.
						squaredDistances := make([]float32, set.Count)
						errorBounds := make([]float32, set.Count)
						quantizer.EstimateSquaredDistances(
							ctx, quantizedSet, queryVector, squaredDistances, errorBounds,
						)
						require.Equal(t, uint64(set.Count), resp.Count)
						require.Equal(t, squaredDistances, resp.SquaredDistances)
						require.Equal(t, errorBounds, resp.ErrorBounds)
						if level == vecstore.LeafLevel {
							require.Zero(t, len(resp.ChildPartitionKeys))
							require.Equal(t, set.Count, len(resp.ChildPrimaryKeys))
							for i := range set.Count {
								require.Equal(t, roachpb.Key(getEncPrimaryKey(i)), resp.ChildPrimaryKeys[i])
							}
						} else {
							require.Zero(t, len(resp.ChildPrimaryKeys))
							require.Equal(t, set.Count, len(resp.ChildPartitionKeys))
							for i := range set.Count {
								require.Equal(t, uint64(i), resp.ChildPartitionKeys[i])
							}
						}
					}
				})
			}
		})
	}
}

// putQuantizedVectorSet encodes the given quantized vector set and writes it to
// the engine. It returns the randomly generated partition key.
//
// The key suffix and child key for each vector is the index of the vector
// within the partition.
func putQuantizedVectorSet(
	t *testing.T,
	ctx context.Context,
	rnd *rand.Rand,
	eng storage.Engine,
	set quantize.QuantizedVectorSet,
	level vecstore.Level,
	ts func() hlc.Timestamp,
) (partitionKey roachpb.Key) {
	partitionKey = encoding.EncodeUint64Ascending(nil, randutil.RandUint64n(rnd, 1000))
	vectorKeyPrefix := partitionKey[:len(partitionKey):len(partitionKey)]
	vectorKey := func(i int) roachpb.Key {
		var childKey vecstore.ChildKey
		if level == vecstore.LeafLevel {
			childKey.PrimaryKey = getEncPrimaryKey(i)
		} else {
			childKey.PartitionKey = vecstore.PartitionKey(i)
		}
		return vecstore.EncodeChildKey(vectorKeyPrefix, childKey)
	}
	encMetadata, err := vecstore.EncodePartitionMetadata(level, set.GetCentroid())
	require.NoError(t, err)
	_, err = storage.MVCCPut(
		ctx, eng, partitionKey, ts(), roachpb.MakeValueFromBytes(encMetadata),
		storage.MVCCWriteOptions{},
	)
	require.NoError(t, err)
	for i := 0; i < set.GetCount(); i++ {
		var encVector []byte
		switch quantizedSet := set.(type) {
		case *quantize.UnQuantizedVectorSet:
			encVector, err = vecstore.EncodeUnquantizedVector(nil,
				quantizedSet.CentroidDistances[i], quantizedSet.Vectors.At(i),
			)
			require.NoError(t, err)
		case *quantize.RaBitQuantizedVectorSet:
			encVector = vecstore.EncodeRaBitQVector(nil,
				quantizedSet.CodeCounts[i], quantizedSet.CentroidDistances[i],
				quantizedSet.DotProducts[i], quantizedSet.Codes.At(i),
			)
		}
		_, err = storage.MVCCPut(
			ctx, eng, vectorKey(i), ts(), roachpb.MakeValueFromBytes(encVector),
			storage.MVCCWriteOptions{},
		)
		require.NoError(t, err)
	}
	return partitionKey
}

func getEncPrimaryKey(i int) []byte {
	pk := encoding.EncodeUint64Ascending(nil, uint64(i))
	rnd := rand.New(rand.NewSource(int64(i)))
	return encoding.EncodeBytesAscending(pk, randutil.RandBytes(rnd, 16))
}

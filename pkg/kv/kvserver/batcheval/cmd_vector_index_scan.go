// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package batcheval

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/vecindex/quantize"
	"github.com/cockroachdb/cockroach/pkg/sql/vecindex/vecstore"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/vector"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(kvpb.VectorIndexScan, DefaultDeclareIsolatedKeys, VectorIndexScan)
}

func VectorIndexScan(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp kvpb.Response,
) (result.Result, error) {
	args := cArgs.Args.(*kvpb.VectorIndexScanRequest)
	h := cArgs.Header
	reply := resp.(*kvpb.VectorIndexScanResponse)

	readCategory := ScanReadCategory(cArgs.EvalCtx.AdmissionHeader())
	opts := storage.MVCCScanOptions{
		Inconsistent:            h.ReadConsistency != kvpb.CONSISTENT,
		Txn:                     h.Txn,
		ScanStats:               cArgs.ScanStats,
		Uncertainty:             cArgs.Uncertainty,
		MaxKeys:                 h.MaxSpanRequestKeys,
		MaxLockConflicts:        storage.MaxConflictsPerLockConflictError.Get(&cArgs.EvalCtx.ClusterSettings().SV),
		TargetLockConflictBytes: storage.TargetBytesPerLockConflictError.Get(&cArgs.EvalCtx.ClusterSettings().SV),
		TargetBytes:             h.TargetBytes,
		AllowEmpty:              h.AllowEmpty,
		WholeRowsOfSize:         h.WholeRowsOfSize,
		Reverse:                 false,
		MemoryAccount:           cArgs.EvalCtx.GetResponseMemoryAccount(),
		DontInterleaveIntents:   cArgs.DontInterleaveIntents,
		ReadCategory:            readCategory,
	}

	getEncodedVal := func(rawBytes []byte) ([]byte, error) {
		mvccVal, err := storage.DecodeMVCCValue(rawBytes)
		if err != nil {
			return nil, err
		}
		// TODO(drewk,mw5h): this is assuming we're using ValueType_BYTES encoding.
		// Verify that this will always be the case.
		return mvccVal.Value.GetBytes()
	}

	// TODO(drewk,mw5h): We could push the distance calculation logic deeper and
	// avoid materializing the KVs in intermediate buffers. E.g., we could
	// implement the results interface.
	scanRes, err := storage.MVCCScan(ctx, reader, args.Key, args.EndKey, h.Timestamp, opts)
	if err != nil {
		return result.Result{}, err
	}
	if len(scanRes.KVs) == 0 {
		return result.Result{}, vecstore.ErrPartitionNotFound
	}
	reply.NumKeys = scanRes.NumKeys
	reply.NumBytes = scanRes.NumBytes
	vectorKVs := scanRes.KVs[1:]
	numVectors := len(vectorKVs)
	reply.Count = uint64(numVectors)

	// The metadata row is always the first entry in the partition's KV span.
	encMetadata, err := getEncodedVal(scanRes.KVs[0].Value.RawBytes)
	if err != nil {
		return result.Result{}, err
	}
	level, centroid, err := vecstore.DecodePartitionMetadata(encMetadata)
	if err != nil {
		return result.Result{}, err
	}
	// The key for each vector is prefixed by the partition key, followed by the
	// child key.
	prefixLen := len(args.Key)
	switch level {
	case vecstore.InvalidLevel:
		return result.Result{}, errors.AssertionFailedf("vector index level cannot be zero")
	case vecstore.LeafLevel:
		// At the leaf level, index entries point to the primary key.
		reply.ChildPrimaryKeys = make([]roachpb.Key, numVectors)
		for i := range vectorKVs {
			reply.ChildPrimaryKeys[i] = vectorKVs[i].Key[prefixLen:]
		}
	default:
		// At non-leaf levels, index entries point to the next level.
		reply.ChildPartitionKeys = make([]uint64, numVectors)
		for i := range vectorKVs {
			_, childPartitionKey, err := encoding.DecodeUint64Ascending(vectorKVs[i].Key[prefixLen:])
			if err != nil {
				return result.Result{}, err
			}
			reply.ChildPartitionKeys[i] = childPartitionKey
		}
	}

	// TODO(drewk,mw5h): we may have to handle column families causing there to be
	// multiple KVs per indexed row. For now, just assume one KV per row.
	//
	// TODO(drewk,mw5h): we could avoid allocating the quantized sets by
	// performing the distance calculations directly on the encoded vectors.
	reply.SquaredDistances = make([]float32, numVectors)
	reply.ErrorBounds = make([]float32, numVectors)
	switch args.QuantizeMethod {
	case quantize.MethodNotQuantized:
		quantizedSet := quantize.UnQuantizedVectorSet{
			Centroid:          centroid,
			CentroidDistances: make([]float32, 0, numVectors),
			Vectors:           vector.MakeSet(int(args.Dims)),
		}
		for i := range vectorKVs {
			encVal, err := getEncodedVal(vectorKVs[i].Value.RawBytes)
			if err != nil {
				return result.Result{}, err
			}
			err = vecstore.DecodeUnquantizedVectorToSet(encVal, &quantizedSet)
			if err != nil {
				return result.Result{}, err
			}
		}
		quantizer := quantize.NewUnQuantizer(int(args.Dims))
		quantizer.EstimateSquaredDistances(
			ctx, &quantizedSet, args.QueryVector, reply.SquaredDistances, reply.ErrorBounds,
		)
	case quantize.MethodRaBitQ:
		quantizedSet := quantize.RaBitQuantizedVectorSet{
			Centroid:          centroid,
			Codes:             quantize.MakeRaBitQCodeSet(int(args.Dims)),
			CodeCounts:        make([]uint32, 0, numVectors),
			CentroidDistances: make([]float32, 0, numVectors),
			DotProducts:       make([]float32, 0, numVectors),
		}
		quantizedSet.Codes.Data = make([]uint64, 0, numVectors*quantizedSet.Codes.Width)
		for i := range vectorKVs {
			encVal, err := getEncodedVal(vectorKVs[i].Value.RawBytes)
			if err != nil {
				return result.Result{}, err
			}
			if err = vecstore.DecodeRaBitQVectorToSet(encVal, &quantizedSet); err != nil {
				return result.Result{}, err
			}
		}
		// TODO(drewk,mw5h): this setup step is expensive. We should cache the
		// quantizer, and/or the ROT matrix.
		quantizer := quantize.NewRaBitQuantizer(int(args.Dims), args.Seed)
		quantizer.EstimateSquaredDistances(
			ctx, &quantizedSet, args.QueryVector, reply.SquaredDistances, reply.ErrorBounds,
		)
	default:
		return result.Result{}, errors.AssertionFailedf(
			"unknown quantize method: %d", args.QuantizeMethod)
	}

	var res result.Result
	res.Local.EncounteredIntents = scanRes.Intents
	return res, nil
}

// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kvstreamer

import (
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv/kvpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

const (
	intSliceOverhead          = int64(unsafe.Sizeof([]int{}))
	intSize                   = int64(unsafe.Sizeof(int(0)))
	int32SliceOverhead        = int64(unsafe.Sizeof([]int32{}))
	int32Size                 = int64(unsafe.Sizeof(int32(0)))
	requestUnionSliceOverhead = int64(unsafe.Sizeof([]kvpb.RequestUnion{}))
	requestUnionOverhead      = int64(unsafe.Sizeof(kvpb.RequestUnion{}))
	getRequestOverhead        = int64(unsafe.Sizeof(kvpb.RequestUnion_Get{}) +
		unsafe.Sizeof(kvpb.GetRequest{}))
	scanRequestOverhead = int64(unsafe.Sizeof(kvpb.RequestUnion_Scan{}) +
		unsafe.Sizeof(kvpb.ScanRequest{}))
	responseUnionOverhead = int64(unsafe.Sizeof(kvpb.ResponseUnion_Get{}))
	getResponseOverhead   = int64(unsafe.Sizeof(kvpb.GetResponse{}))
	scanResponseOverhead  = int64(unsafe.Sizeof(kvpb.ScanResponse{}))
)

var zeroInt32Slice []int32

func init() {
	reverseScanRequestOverhead := int64(unsafe.Sizeof(kvpb.RequestUnion_ReverseScan{}) +
		unsafe.Sizeof(kvpb.ReverseScanRequest{}))
	if reverseScanRequestOverhead != scanRequestOverhead {
		panic(fmt.Sprintf(
			"ReverseScanRequest and ScanRequest have different overheads %d and scan req %d",
			reverseScanRequestOverhead, scanRequestOverhead,
		))
	}
	scanResponseUnionOverhead := int64(unsafe.Sizeof(kvpb.ResponseUnion_Scan{}))
	if responseUnionOverhead != scanResponseUnionOverhead {
		panic("ResponseUnion_Get and ResponseUnion_Scan have different overheads")
	}
	zeroInt32Slice = make([]int32, 1<<10)
}

// Note that we cannot use Size() methods that are automatically generated by
// the protobuf library because
// - they calculate the size of the serialized message whereas we're interested
// in the deserialized in-memory footprint.
// - they account for things differently from how the memory usage is accounted
// for by the KV layer for the purposes of tracking TargetBytes limit.

// scanRequestSize calculates the footprint of a {,Reverse}Scan request,
// including the overhead. key and endKey are the keys from the span of the
// request header (we choose to avoid taking in a roachpb.Span in order to
// reduce allocations).
func scanRequestSize(key, endKey roachpb.Key) int64 {
	return scanRequestOverhead + int64(cap(key)) + int64(cap(endKey))
}

// getRequestSize calculates the footprint of a Get request for a given key,
// including its overhead.
func getRequestSize(key roachpb.Key) int64 {
	return getRequestOverhead + int64(cap(key))
}

func requestsMemUsage(reqs []kvpb.RequestUnion) (memUsage int64) {
	for _, r := range reqs {
		req := r.GetInner()
		h := req.Header()
		switch req.Method() {
		case kvpb.Get:
			memUsage += getRequestSize(h.Key)
		case kvpb.Scan, kvpb.ReverseScan:
			memUsage += scanRequestSize(h.Key, h.EndKey)
		default:
			panic(fmt.Sprintf("unexpected request type %s", r.GetInner()))
		}
	}
	return memUsage
}

// getResponseSize calculates the size of the GetResponse similar to how it is
// accounted for TargetBytes parameter by the KV layer.
func getResponseSize(get *kvpb.GetResponse) int64 {
	if get.Value == nil {
		return 0
	}
	return int64(len(get.Value.RawBytes))
}

// scanResponseSize calculates the size of a ScanResponse or ReverseScanResponse
// similar to how it is accounted for TargetBytes parameter by the KV layer.
func scanResponseSize(scan kvpb.Response) int64 {
	switch response := scan.(type) {
	case *kvpb.ScanResponse:
		return response.NumBytes
	case *kvpb.ReverseScanResponse:
		return response.NumBytes
	}
	panic(errors.AssertionFailedf("unexpected response type: %v", scan))
}

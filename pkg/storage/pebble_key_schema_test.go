// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package storage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/testutils/datapathutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/crlib/crbytes"
	"github.com/cockroachdb/crlib/crstrings"
	"github.com/cockroachdb/datadriven"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/sstable/block"
	"github.com/cockroachdb/pebble/sstable/colblk"
	"github.com/olekukonko/tablewriter"
)

func TestKeySchema_KeyWriter(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	var kw colblk.KeyWriter
	var row int
	var buf bytes.Buffer
	var keyBuf []byte
	datadriven.RunTest(t, datapathutils.TestDataPath(t, "key_schema_key_writer"), func(t *testing.T, td *datadriven.TestData) string {
		buf.Reset()
		switch td.Cmd {
		case "init":
			// Exercise both resetting and retrieving a new writer.
			if kw != nil && rand.Intn(2) == 1 {
				kw.Reset()
			} else {
				kw = keySchema.NewKeyWriter()
			}
			row = 0
			keyBuf = keyBuf[:0]
			return ""
		case "write":
			for i, line := range crstrings.Lines(td.Input) {
				k, err := parseTestKey(line)
				if err != nil {
					t.Fatalf("bad test key %q on line %d: %s", line, i, err)
				}
				fmt.Fprintf(&buf, "Parse(%q) = hex:%x\n", line, k)
				kcmp := kw.ComparePrev(k)
				if v := EngineKeyCompare(k, keyBuf); v < 0 {
					t.Fatalf("line %d: EngineKeyCompare(%q, hex:%x) = %d", i, line, keyBuf, v)
				} else if v != int(kcmp.UserKeyComparison) {
					t.Fatalf("line %d: EngineKeyCompare(%q, hex:%x) = %d; kcmp.UserKeyComparison = %d",
						i, line, keyBuf, v, kcmp.UserKeyComparison)
				}

				fmt.Fprintf(&buf, "%02d: ComparePrev(%q): PrefixLen=%d; CommonPrefixLen=%d; UserKeyComparison=%d\n",
					i, line, kcmp.PrefixLen, kcmp.CommonPrefixLen, kcmp.UserKeyComparison)
				kw.WriteKey(row, k, kcmp.PrefixLen, kcmp.CommonPrefixLen)
				fmt.Fprintf(&buf, "%02d: WriteKey(%d, %q, PrefixLen=%d, CommonPrefixLen=%d)\n",
					i, row, line, kcmp.PrefixLen, kcmp.CommonPrefixLen)

				keyBuf = kw.MaterializeKey(keyBuf[:0], row)
				if !EngineKeyEqual(k, keyBuf) {
					t.Fatalf("line %d: EngineKeyEqual(hex:%x, hex:%x) == false", i, k, keyBuf)
				}
				if v := EngineKeyCompare(k, keyBuf); v != 0 {
					t.Fatalf("line %d: EngineKeyCompare(hex:%x, hex:%x) = %d", i, k, keyBuf, v)
				}

				fmt.Fprintf(&buf, "%02d: MaterializeKey(_, %d) = hex:%x\n", i, row, keyBuf)
				row++
			}
			return buf.String()
		case "finish":
			b := crbytes.AllocAligned(int(kw.Size(row, 0) + 1))
			offs := make([]uint32, kw.NumColumns()+1)
			for i := 0; i < kw.NumColumns(); i++ {
				offs[i+1] = kw.Finish(i, row, offs[i], b)
			}
			roachKeys, _ := colblk.DecodePrefixBytes(b, offs[cockroachColRoachKey], row)
			mvccWallTimes, _ := colblk.DecodeUnsafeUints(b, offs[cockroachColMVCCWallTime], row)
			mvccLogicalTimes, _ := colblk.DecodeUnsafeUints(b, offs[cockroachColMVCCLogical], row)
			untypedVersions, _ := colblk.DecodeRawBytes(b, offs[cockroachColUntypedVersion], row)
			tbl := tablewriter.NewWriter(&buf)
			tbl.SetHeader([]string{"Key", "Wall", "Logical", "Untyped"})
			for i := 0; i < row; i++ {
				tbl.Append([]string{
					asciiOrHex(roachKeys.At(i)),
					fmt.Sprintf("%d", mvccWallTimes.At(i)),
					fmt.Sprintf("%d", mvccLogicalTimes.At(i)),
					fmt.Sprintf("%x", untypedVersions.At(i)),
				})
			}
			tbl.Render()
			return buf.String()
		default:
			panic(fmt.Sprintf("unrecognized command %q", td.Cmd))
		}
	})
}

func TestKeySchema_KeySeeker(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	var buf bytes.Buffer
	var enc colblk.DataBlockEncoder
	var dec colblk.DataBlockDecoder
	var ks colblk.KeySeeker
	var maxKeyLen int
	enc.Init(keySchema)

	initKeySeeker := func() {
		if ks == nil || rand.Intn(2) == 1 {
			if ks != nil {
				ks.Release()
			}
			ks = keySchema.NewKeySeeker()
		}
		if err := ks.Init(&dec); err != nil {
			t.Fatal(err)
		}
	}

	datadriven.RunTest(t, datapathutils.TestDataPath(t, "key_schema_key_seeker"), func(t *testing.T, td *datadriven.TestData) string {
		buf.Reset()
		switch td.Cmd {
		case "define-block":
			enc.Reset()
			maxKeyLen = 0
			var rows int
			for i, line := range crstrings.Lines(td.Input) {
				k, err := parseTestKey(line)
				if err != nil {
					t.Fatalf("bad test key %q on line %d: %s", line, i, err)
				}
				fmt.Fprintf(&buf, "Parse(%q) = hex:%x\n", line, k)
				maxKeyLen = max(maxKeyLen, len(k))
				kcmp := enc.KeyWriter.ComparePrev(k)
				ikey := pebble.InternalKey{
					UserKey: k,
					Trailer: pebble.MakeInternalKeyTrailer(0, pebble.InternalKeyKindSet),
				}
				enc.Add(ikey, k, block.InPlaceValuePrefix(false), kcmp, false /* isObsolete */)
				rows++
			}
			blk, _ := enc.Finish(rows, enc.Size())
			dec.Init(keySchema, blk)
			return buf.String()
		case "is-lower-bound":
			initKeySeeker()
			syntheticSuffix, syntheticSuffixStr, _ := getSyntheticSuffix(t, td)
			for _, line := range crstrings.Lines(td.Input) {
				k, err := parseTestKey(line)
				if err != nil {
					t.Fatalf("bad test key %q: %s", line, err)
				}
				got := ks.IsLowerBound(k, syntheticSuffix)
				fmt.Fprintf(&buf, "IsLowerBound(%q, %q) = %t\n", line, syntheticSuffixStr, got)
			}
			return buf.String()
		case "seek-ge":
			initKeySeeker()
			for _, line := range crstrings.Lines(td.Input) {
				fields := strings.Fields(line)
				k, err := parseTestKey(fields[0])
				if err != nil {
					t.Fatalf("bad test key %q: %s", fields[0], err)
				}
				boundRow := -1
				searchDir := 0
				if len(fields) == 3 {
					boundRow, err = strconv.Atoi(fields[1])
					if err != nil {
						t.Fatalf("bad bound row %q: %s", fields[1], err)
					}
					switch fields[2] {
					case "fwd":
						searchDir = +1
					case "bwd":
						searchDir = -1
					default:
						t.Fatalf("bad search direction %q", fields[2])
					}
				}
				row, equalPrefix := ks.SeekGE(k, boundRow, int8(searchDir))

				fmt.Fprintf(&buf, "SeekGE(%q, boundRow=%d, searchDir=%d) = (row=%d, equalPrefix=%t)",
					line, boundRow, searchDir, row, equalPrefix)
				if row >= 0 && row < dec.BlockDecoder().Rows() {
					var kiter colblk.PrefixBytesIter
					kiter.Buf = make([]byte, maxKeyLen+1)
					key := ks.MaterializeUserKey(&kiter, -1, row)
					fmt.Fprintf(&buf, " [hex:%x]", key)
				}
				fmt.Fprintln(&buf)
			}
			return buf.String()
		case "materialize-user-key":
			initKeySeeker()
			syntheticSuffix, syntheticSuffixStr, syntheticSuffixOk := getSyntheticSuffix(t, td)

			var kiter colblk.PrefixBytesIter
			kiter.Buf = make([]byte, maxKeyLen+len(syntheticSuffix)+1)
			prevRow := -1
			for _, line := range crstrings.Lines(td.Input) {
				row, err := strconv.Atoi(line)
				if err != nil {
					t.Fatalf("bad row number %q: %s", line, err)
				}
				if syntheticSuffixOk {
					key := ks.MaterializeUserKeyWithSyntheticSuffix(&kiter, syntheticSuffix, prevRow, row)
					fmt.Fprintf(&buf, "MaterializeUserKeyWithSyntheticSuffix(%d, %d, %s) = hex:%x\n", prevRow, row, syntheticSuffixStr, key)
				} else {
					key := ks.MaterializeUserKey(&kiter, prevRow, row)
					fmt.Fprintf(&buf, "MaterializeUserKey(%d, %d) = hex:%x\n", prevRow, row, key)
				}
				prevRow = row
			}
			return buf.String()
		default:
			panic(fmt.Sprintf("unrecognized command %q", td.Cmd))
		}
	})

}

func getSyntheticSuffix(t *testing.T, td *datadriven.TestData) ([]byte, string, bool) {
	var syntheticSuffix []byte
	var syntheticSuffixStr string
	cmdArg, ok := td.Arg("synthetic-suffix")
	if ok {
		syntheticSuffixStr = cmdArg.SingleVal(t)
		var err error
		syntheticSuffix, err = parseTestKey(syntheticSuffixStr)
		if err != nil {
			t.Fatalf("parsing synthetic suffix %q: %s", syntheticSuffixStr, err)
		}
		syntheticSuffix = syntheticSuffix[1:] // Trim the separator byte.
	}
	return syntheticSuffix, syntheticSuffixStr, ok
}

func asciiOrHex(b []byte) string {
	if bytes.ContainsFunc(b, func(r rune) bool { return r < ' ' || r > '~' }) {
		return fmt.Sprintf("hex:%x", b)
	}
	return string(b)
}

func parseTestKey(s string) ([]byte, error) {
	if strings.HasPrefix(s, "hex:") {
		b, err := hex.DecodeString(strings.TrimPrefix(s, "hex:"))
		if err != nil {
			return nil, errors.Wrap(err, "parsing hexadecimal literal key")
		}
		return b, nil
	}
	i := strings.IndexByte(s, '@')
	if i == -1 {
		// Return just the roachpb key with the sentinel byte.
		return append([]byte(s), 0x00), nil
	}
	if len(s[i+1:]) == 0 {
		return nil, errors.Newf("key %q has empty suffix", s)
	}
	version := s[i+1:]
	j := strings.IndexByte(version, ',')
	switch version[0:j] {
	case "Shared", "Exclusive", "Intent":
		// This is the lock strength. Parse as a lock table key.
		strength := lock.Intent
		switch version[0:j] {
		case "Shared":
			strength = lock.Shared
		case "Exclusive":
			strength = lock.Exclusive
		}
		txnUUID, err := uuid.FromString(version[j+1:])
		if err != nil {
			return nil, errors.Wrapf(err, "parsing lock table transaction UUID")
		}
		ltk := LockTableKey{
			Key:      []byte(s[:i]),
			Strength: strength,
			TxnUUID:  txnUUID,
		}
		ek, _ := ltk.ToEngineKey(nil)
		return ek.Encode(), nil
	default:
		// Parse as a MVCC key.
		ts, err := hlc.ParseTimestamp(version)
		if err != nil {
			return nil, errors.Wrap(err, "parsing MVCC timestamp")
		}
		return EncodeMVCCKey(MVCCKey{
			Key:       []byte(s[:i]),
			Timestamp: ts,
		}), nil
	}
}

// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package duration

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"

	_ "github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/stretchr/testify/require"
)

type durationTest struct {
	cmpToPrev int
	duration  Duration
	err       bool
}

// positiveDurationTests is used both to check that each duration roundtrips
// through Encode/Decode and that they sort in the expected way. They are not
// required to be listed in ascending order, but for ease of maintenance, it is
// expected that they stay ascending.
//
// The negative tests are generated by prepending everything but the 0 case and
// flipping the sign of cmpToPrev, since they will be getting bigger in absolute
// value and more negative.
//
// TODO(dan): Write more tests with a mixture of positive and negative
// components.
var positiveDurationTests = []durationTest{
	{1, Duration{Months: 0, Days: 0, nanos: 0}, false},
	{1, Duration{Months: 0, Days: 0, nanos: 1}, false},
	{1, Duration{Months: 0, Days: 0, nanos: nanosInDay - 1}, false},
	{1, Duration{Months: 0, Days: 1, nanos: 0}, false},
	{0, Duration{Months: 0, Days: 0, nanos: nanosInDay}, false},
	{1, Duration{Months: 0, Days: 0, nanos: nanosInDay + 1}, false},
	{1, Duration{Months: 0, Days: DaysPerMonth - 1, nanos: 0}, false},
	{1, Duration{Months: 0, Days: 0, nanos: nanosInMonth - 1}, false},
	{1, Duration{Months: 1, Days: 0, nanos: 0}, false},
	{0, Duration{Months: 0, Days: DaysPerMonth, nanos: 0}, false},
	{0, Duration{Months: 0, Days: 0, nanos: nanosInMonth}, false},
	{1, Duration{Months: 0, Days: 0, nanos: nanosInMonth + 1}, false},
	{1, Duration{Months: 0, Days: DaysPerMonth + 1, nanos: 0}, false},
	{1, Duration{Months: 1, Days: 1, nanos: 1}, false},
	{1, Duration{Months: 1, Days: 10, nanos: 0}, false},
	{0, Duration{Months: 0, Days: 40, nanos: 0}, false},
	{1, Duration{Months: 2, Days: 0, nanos: 0}, false},
	// '106751 days 23:47:16.854775' should not overflow.
	{1, Duration{Months: 0, Days: 106751, nanos: 85636854775000}, false},
	// '106751 days 23:47:16.854776' should overflow.
	{1, Duration{Months: 0, Days: 106751, nanos: 85636854776000}, true},
	{1, Duration{Months: math.MaxInt64 - 1, Days: DaysPerMonth - 1, nanos: nanosInDay * 2}, true},
	{1, Duration{Months: math.MaxInt64 - 1, Days: DaysPerMonth * 2, nanos: nanosInDay * 2}, true},
	{1, Duration{Months: math.MaxInt64, Days: math.MaxInt64, nanos: nanosInMonth + nanosInDay}, true},
	{1, Duration{Months: math.MaxInt64, Days: math.MaxInt64, nanos: math.MaxInt64}, true},
}

var mixedDurationTests = []durationTest{
	{-1, Duration{Months: -1, Days: -2, nanos: 123}, false},
	{1, Duration{Months: 0, Days: -1, nanos: 123}, false},
	{1, Duration{Months: 0, Days: 0, nanos: -123}, false},
	{1, Duration{Months: 0, Days: 1, nanos: -123456}, false},
}

func fullDurationTests() []durationTest {
	var ret []durationTest
	for _, test := range positiveDurationTests {
		d := test.duration
		negDuration := Duration{Months: -d.Months, Days: -d.Days, nanos: -d.nanos}
		ret = append(ret, durationTest{cmpToPrev: -test.cmpToPrev, duration: negDuration, err: test.err})
	}
	ret = append(ret, positiveDurationTests...)
	ret = append(ret, mixedDurationTests...)
	return ret
}

func TestEncodeDecode(t *testing.T) {
	for i, test := range fullDurationTests() {
		sortNanos, months, days, err := test.duration.Encode()
		if test.err && err == nil {
			t.Errorf("%d expected error but didn't get one", i)
		} else if !test.err && err != nil {
			t.Errorf("%d expected no error but got one: %s", i, err)
		}
		if err != nil {
			continue
		}
		sortNanosBig, _, _ := test.duration.EncodeBigInt()
		if sortNanos != sortNanosBig.Int64() {
			t.Errorf("%d Encode and EncodeBig didn't match [%d] vs [%s]", i, sortNanos, sortNanosBig)
		}
		d, err := Decode(sortNanos, months, days)
		if err != nil {
			t.Fatal(err)
		}
		if test.duration != d {
			t.Errorf("%d encode/decode mismatch [%v] vs [%v]", i, test, d)
		}
	}
}

func TestCompare(t *testing.T) {
	prev := Duration{nanos: 1} // It's expected that we start with something greater than 0.
	for i, test := range fullDurationTests() {
		cmp := test.duration.Compare(prev)
		if cmp != test.cmpToPrev {
			t.Errorf("%d did not compare correctly got %d expected %d [%s] vs [%s]", i, cmp, test.cmpToPrev, prev, test.duration)
		}
		prev = test.duration
	}
}

func TestNormalize(t *testing.T) {
	for i, test := range fullDurationTests() {
		nanos, _, _ := test.duration.EncodeBigInt()
		normalized := test.duration.normalize()
		normalizedNanos, _, _ := normalized.EncodeBigInt()
		if nanos.Cmp(normalizedNanos) != 0 {
			t.Errorf("%d effective nanos were changed [%s] [%s]", i, test.duration, normalized)
		}
		if normalized.Days > DaysPerMonth && normalized.Months != math.MaxInt64 ||
			normalized.Days < -DaysPerMonth && normalized.Months != math.MinInt64 {
			t.Errorf("%d days were not normalized [%s]", i, normalized)
		}
		if normalized.nanos > nanosInDay && normalized.Days != math.MaxInt64 ||
			normalized.nanos < -nanosInDay && normalized.Days != math.MinInt64 {
			t.Errorf("%d nanos were not normalized [%s]", i, normalized)
		}
	}
}

func TestDiffMicros(t *testing.T) {
	tests := []struct {
		t1, t2  time.Time
		expDiff int64
	}{
		{
			t1:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 0,
		},
		{
			t1:      time.Date(1, 8, 15, 12, 30, 45, 0, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: -63062710155000000,
		},
		{
			t1:      time.Date(1994, 8, 15, 12, 30, 45, 0, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: -169730955000000,
		},
		{
			t1:      time.Date(2012, 8, 15, 12, 30, 45, 0, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 398349045000000,
		},
		{
			t1:      time.Date(8012, 8, 15, 12, 30, 45, 0, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 189740061045000000,
		},
		// Test if the nanoseconds round correctly.
		{
			t1:      time.Date(2000, 1, 1, 0, 0, 0, 499, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 0,
		},
		{
			t1:      time.Date(1999, 12, 31, 23, 59, 59, 999999501, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 0,
		},
		{
			t1:      time.Date(2000, 1, 1, 0, 0, 0, 500, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: 1,
		},
		{
			t1:      time.Date(1999, 12, 31, 23, 59, 59, 999999500, time.UTC),
			t2:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expDiff: -1,
		},
	}

	for i, test := range tests {
		if res := DiffMicros(test.t1, test.t2); res != test.expDiff {
			t.Errorf("%d: expected DiffMicros(%v, %v) = %d, found %d",
				i, test.t1, test.t2, test.expDiff, res)
		} else {
			// Swap order and make sure the results are mirrored.
			exp := -test.expDiff
			if res := DiffMicros(test.t2, test.t1); res != exp {
				t.Errorf("%d: expected DiffMicros(%v, %v) = %d, found %d",
					i, test.t2, test.t1, exp, res)
			}
		}
	}
}

// TestAdd looks at various rounding cases, comparing our date math
// to behavior observed in PostgreSQL 10.
func TestAdd(t *testing.T) {

	bucharest, err := timeutil.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}

	tests := []struct {
		t   time.Time
		d   Duration
		exp time.Time
	}{
		// Year wraparound
		{
			t:   time.Date(1993, 10, 01, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 3},
			exp: time.Date(1994, 1, 01, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(1992, 10, 01, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 15},
			exp: time.Date(1994, 1, 01, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 28, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 12},
			exp: time.Date(2021, time.July, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 18},
			exp: time.Date(2022, time.January, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 30},
			exp: time.Date(2023, time.January, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 12},
			exp: time.Date(2021, time.July, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 12, nanos: int64(10 * time.Second)},
			exp: time.Date(2021, time.July, 29, 0, 0, 10, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 17, nanos: int64(10 * time.Second)},
			exp: time.Date(2021, time.December, 29, 0, 0, 10, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 18, nanos: int64(10 * time.Second)},
			exp: time.Date(2022, time.January, 29, 0, 0, 10, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 19, nanos: int64(10 * time.Second)},
			exp: time.Date(2022, time.February, 28, 0, 0, 10, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 20, nanos: int64(10 * time.Second)},
			exp: time.Date(2022, time.March, 29, 0, 0, 10, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 30, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 12},
			exp: time.Date(2021, time.July, 30, 0, 0, 0, 0, time.UTC),
		},
		// Check DST boundaries.
		{
			// Regression for #64772.
			t:   time.Date(2020, 10, 25, 0, 0, 0, 0, bucharest).Add(3 * time.Hour), // 03:00+03
			d:   Duration{},
			exp: time.Date(2020, 10, 25, 0, 0, 0, 0, bucharest).Add(3 * time.Hour), // 03:00+03
		},
		{
			// Regression for #64772.
			t:   time.Date(2020, 10, 25, 0, 0, 0, 0, bucharest).Add(3 * time.Hour), // 03:00+03
			d:   Duration{nanos: int64(time.Hour)},
			exp: time.Date(2020, 10, 25, 3, 0, 0, 0, bucharest), // 03:00+02
		},
		{
			t:   time.Date(2020, 10, 24, 3, 0, 0, 0, bucharest),
			d:   Duration{Days: 1},
			exp: time.Date(2020, 10, 25, 3, 0, 0, 0, bucharest),
		},

		// Check leap behaviors
		{
			t:   time.Date(1996, 02, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 12},
			exp: time.Date(1997, 02, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(1996, 02, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 48},
			exp: time.Date(2000, 02, 29, 0, 0, 0, 0, time.UTC),
		},

		// This pair shows something one might argue is weird:
		// that two different times plus the same duration results
		// in the same result.
		{
			t:   time.Date(1996, 01, 30, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 1, Days: 1},
			exp: time.Date(1996, 03, 01, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(1996, 01, 31, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: 1, Days: 1},
			exp: time.Date(1996, 03, 01, 0, 0, 0, 0, time.UTC),
		},

		// Check negative operations
		{
			t:   time.Date(2016, 02, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1},
			exp: time.Date(2016, 01, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2016, 02, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1, Days: -1},
			exp: time.Date(2016, 01, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2016, 03, 31, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1},
			exp: time.Date(2016, 02, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2016, 03, 31, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1, Days: -1},
			exp: time.Date(2016, 02, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2016, 02, 01, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1},
			exp: time.Date(2016, 01, 01, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2016, 02, 01, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -1, Days: -1},
			exp: time.Date(2015, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 28, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -12},
			exp: time.Date(2019, time.July, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -18},
			exp: time.Date(2019, time.January, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -30},
			exp: time.Date(2018, time.January, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -12},
			exp: time.Date(2019, time.July, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -12, nanos: int64(-10 * time.Second)},
			exp: time.Date(2019, time.July, 28, 23, 59, 50, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -19, nanos: int64(-10 * time.Second)},
			exp: time.Date(2018, time.December, 28, 23, 59, 50, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 29, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -20, nanos: int64(-10 * time.Second)},
			exp: time.Date(2018, time.November, 28, 23, 59, 50, 0, time.UTC),
		},
		{
			t:   time.Date(2020, time.July, 30, 0, 0, 0, 0, time.UTC),
			d:   Duration{Months: -12},
			exp: time.Date(2019, time.July, 30, 0, 0, 0, 0, time.UTC),
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.t, test.d), func(t *testing.T) {
			if res := Add(test.t, test.d); !test.exp.Equal(res) {
				t.Errorf("%d: expected Add(%s, %s) = %s, found %s",
					i, test.t, test.d, test.exp, res)
			}
		})
	}
}

func TestAddMicros(t *testing.T) {
	tests := []struct {
		t   time.Time
		d   int64
		exp time.Time
	}{
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   0,
			exp: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   123456789,
			exp: time.Date(2000, 1, 1, 0, 2, 3, 456789000, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 999, time.UTC),
			d:   123456789,
			exp: time.Date(2000, 1, 1, 0, 2, 3, 456789999, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   12345678987654321,
			exp: time.Date(2391, 03, 21, 19, 16, 27, 654321000, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   math.MaxInt64 / 10,
			exp: time.Date(31227, 9, 14, 2, 48, 05, 477580000, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   -123456789,
			exp: time.Date(1999, 12, 31, 23, 57, 56, 543211000, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 999, time.UTC),
			d:   -123456789,
			exp: time.Date(1999, 12, 31, 23, 57, 56, 543211999, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   -12345678987654321,
			exp: time.Date(1608, 10, 12, 04, 43, 32, 345679000, time.UTC),
		},
		{
			t:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			d:   -math.MaxInt64 / 10,
			exp: time.Date(-27228, 04, 18, 21, 11, 54, 522420000, time.UTC),
		},
	}

	for i, test := range tests {
		if res := AddMicros(test.t, test.d); !test.exp.Equal(res) {
			t.Errorf("%d: expected AddMicros(%v, %d) = %v, found %v",
				i, test.t, test.d, test.exp, res)
		}
	}
}

func TestFloatMath(t *testing.T) {
	const nanosInMinute = nanosInSecond * 60
	const nanosInHour = nanosInMinute * 60
	durationOutOfRange := MakeDuration(math.MaxInt64, 999999999999, 999999999999)
	negatedDurationOutOfRange := MakeDuration(-math.MaxInt64, -999999999999, -999999999999)

	tests := []struct {
		d   Duration
		f   float64
		mul Duration
		div Duration
	}{
		{
			Duration{Months: 1, Days: 2, nanos: nanosInHour * 2},
			0.15,
			Duration{Days: 4, nanos: nanosInHour*19 + nanosInMinute*30},
			Duration{Months: 6, Days: 33, nanos: nanosInHour*21 + nanosInMinute*20},
		},
		{
			Duration{Months: 1, Days: 2, nanos: nanosInHour * 2},
			0.3,
			Duration{Days: 9, nanos: nanosInHour * 15},
			Duration{Months: 3, Days: 16, nanos: nanosInHour*22 + nanosInMinute*40},
		},
		{
			Duration{Months: 1, Days: 2, nanos: nanosInHour * 2},
			0.5,
			Duration{Days: 16, nanos: nanosInHour * 1},
			Duration{Months: 2, Days: 4, nanos: nanosInHour * 4},
		},
		{
			Duration{Months: 1, Days: 2, nanos: nanosInHour * 2},
			0.8,
			Duration{Days: 25, nanos: nanosInHour * 16},
			Duration{Months: 1, Days: 10, nanos: nanosInHour*2 + nanosInMinute*30},
		},
		{
			Duration{Months: 1, Days: 17, nanos: nanosInHour * 2},
			2.0,
			Duration{Months: 2, Days: 34, nanos: nanosInHour * 4},
			Duration{Days: 23, nanos: nanosInHour * 13},
		},
		{
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0.253000},
			3.2,
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0.8096},
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0.079062},
		},
		{
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0.000001},
			2.0,
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0.000002},
			Duration{Months: 0, Days: 0, nanos: nanosInSecond * 0},
		},
		{
			durationOutOfRange,
			1.0,
			durationOutOfRange,
			durationOutOfRange,
		},
		{
			durationOutOfRange,
			-1.0,
			negatedDurationOutOfRange,
			negatedDurationOutOfRange,
		},
	}

	for i, test := range tests {
		if res := test.d.MulFloat(test.f); test.mul != res {
			t.Errorf(
				"%d: expected %v.MulFloat(%f) = %v, found %v",
				i,
				test.d,
				test.f,
				test.mul,
				res)
		}
		if res := test.d.DivFloat(test.f); test.div != res {
			t.Errorf(
				"%d: expected %v.DivFloat(%f) = %v, found %v",
				i,
				test.d,
				test.f,
				test.div,
				res)
		}
	}
}

func TestTruncate(t *testing.T) {
	zero := time.Duration(0).String()
	testCases := []struct {
		d, r time.Duration
		s    string
	}{
		{0, 1, zero},
		{0, 1, zero},
		{time.Second, 1, "1s"},
		{time.Second, 2 * time.Second, zero},
		{time.Second + 1, time.Second, "1s"},
		{11 * time.Nanosecond, 10 * time.Nanosecond, "10ns"},
		{time.Hour + time.Nanosecond + 3*time.Millisecond + time.Second, time.Millisecond, "1h0m1.003s"},
	}
	for i, tc := range testCases {
		if s := Truncate(tc.d, tc.r).String(); s != tc.s {
			t.Errorf("%d: (%s,%s) should give %s, but got %s", i, tc.d, tc.r, tc.s, s)
		}
	}
}

// TestNanos verifies that nanoseconds can only be present after Decode and
// that any operation will remove them.
func TestNanos(t *testing.T) {
	d, err := Decode(1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if expect, actual := int64(1), d.nanos; expect != actual {
		t.Fatalf("expected %d, got %d", expect, actual)
	}
	if expect, actual := "00:00:00+1ns", d.StringNanos(); expect != actual {
		t.Fatalf("expected %s, got %s", expect, actual)
	}
	// Add, even of a 0-duration interval, should call round.
	d = d.Add(Duration{})
	if expect, actual := int64(0), d.nanos; expect != actual {
		t.Fatalf("expected %d, got %d", expect, actual)
	}
	d, err = Decode(500, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if expect, actual := int64(500), d.nanos; expect != actual {
		t.Fatalf("expected %d, got %d", expect, actual)
	}
	if expect, actual := "00:00:00+500ns", d.StringNanos(); expect != actual {
		t.Fatalf("expected %s, got %s", expect, actual)
	}
	d = d.Add(Duration{})
	if expect, actual := int64(1000), d.nanos; expect != actual {
		t.Fatalf("expected %d, got %d", expect, actual)
	}
	if expect, actual := "00:00:00.000001", d.StringNanos(); expect != actual {
		t.Fatalf("expected %s, got %s", expect, actual)
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		duration    Duration
		postgres    string
		iso8601     string
		sqlStandard string
	}{
		{
			duration:    Duration{},
			postgres:    "00:00:00",
			iso8601:     "PT0S",
			sqlStandard: "0",
		},
		// all negative

		// plural
		{
			duration:    Duration{Months: -35, Days: -2, nanos: -12345678000000},
			postgres:    "-2 years -11 mons -2 days -03:25:45.678",
			iso8601:     "P-2Y-11M-2DT-3H-25M-45.678S",
			sqlStandard: "-2-11 -2 -3:25:45.678",
		},
		// singular
		{
			duration:    Duration{Months: -13, Days: -1, nanos: -1000000},
			postgres:    "-1 years -1 mons -1 days -00:00:00.001",
			iso8601:     "P-1Y-1M-1DT-0.001S",
			sqlStandard: "-1-1 -1 -0:00:00.001",
		},
		// years only
		{
			duration:    Duration{Months: -12},
			postgres:    "-1 years",
			iso8601:     "P-1Y",
			sqlStandard: "-1-0",
		},
		// months only
		{
			duration:    Duration{Months: -1},
			postgres:    "-1 mons",
			iso8601:     "P-1M",
			sqlStandard: "-0-1",
		},
		// days only
		{
			duration:    Duration{Days: -1},
			postgres:    "-1 days",
			iso8601:     "P-1D",
			sqlStandard: "-1 0:00:00",
		},
		// hours only
		{
			duration:    Duration{nanos: -int64(time.Hour)},
			postgres:    "-01:00:00",
			iso8601:     "PT-1H",
			sqlStandard: "-1:00:00",
		},
		// minutes only
		{
			duration:    Duration{nanos: -int64(time.Minute)},
			postgres:    "-00:01:00",
			iso8601:     "PT-1M",
			sqlStandard: "-0:01:00",
		},
		// seconds only
		{
			duration:    Duration{nanos: -int64(time.Second)},
			postgres:    "-00:00:01",
			iso8601:     "PT-1S",
			sqlStandard: "-0:00:01",
		},
		// milliseconds only
		{
			duration:    Duration{nanos: -int64(time.Millisecond)},
			postgres:    "-00:00:00.001",
			iso8601:     "PT-0.001S",
			sqlStandard: "-0:00:00.001",
		},
		// without time
		{
			duration:    Duration{Months: -35, Days: -1, nanos: 0},
			postgres:    "-2 years -11 mons -1 days",
			iso8601:     "P-2Y-11M-1D",
			sqlStandard: "-2-11 -1 +0:00:00",
		},
		// without years
		{
			duration:    Duration{Months: -11, Days: -1, nanos: -1000000},
			postgres:    "-11 mons -1 days -00:00:00.001",
			iso8601:     "P-11M-1DT-0.001S",
			sqlStandard: "-0-11 -1 -0:00:00.001",
		},
		// without months
		{
			duration:    Duration{Months: -12, Days: -1, nanos: -1000000},
			postgres:    "-1 years -1 days -00:00:00.001",
			iso8601:     "P-1Y-1DT-0.001S",
			sqlStandard: "-1-0 -1 -0:00:00.001",
		},
		// without years, months
		{
			duration:    Duration{Months: 0, Days: -1, nanos: -1000000},
			postgres:    "-1 days -00:00:00.001",
			iso8601:     "P-1DT-0.001S",
			sqlStandard: "-1 0:00:00.001",
		},
		// without days
		{
			duration:    Duration{Months: -35, Days: 0, nanos: -1000000},
			postgres:    "-2 years -11 mons -00:00:00.001",
			iso8601:     "P-2Y-11MT-0.001S",
			sqlStandard: "-2-11 +0 -0:00:00.001",
		},

		// all positive

		// plural
		{
			duration:    Duration{Months: 35, Days: 2, nanos: 12345678000000},
			postgres:    "2 years 11 mons 2 days 03:25:45.678",
			iso8601:     "P2Y11M2DT3H25M45.678S",
			sqlStandard: "+2-11 +2 +3:25:45.678",
		},
		// singular
		{
			duration:    Duration{Months: 13, Days: 1, nanos: 1000000},
			postgres:    "1 year 1 mon 1 day 00:00:00.001",
			iso8601:     "P1Y1M1DT0.001S",
			sqlStandard: "+1-1 +1 +0:00:00.001",
		},
		// without time
		{
			duration:    Duration{Months: 35, Days: 1, nanos: 0},
			postgres:    "2 years 11 mons 1 day",
			iso8601:     "P2Y11M1D",
			sqlStandard: "+2-11 +1 +0:00:00",
		},
		// without years
		{
			duration:    Duration{Months: 11, Days: 1, nanos: 1000000},
			postgres:    "11 mons 1 day 00:00:00.001",
			iso8601:     "P11M1DT0.001S",
			sqlStandard: "+0-11 +1 +0:00:00.001",
		},
		// without months
		{
			duration:    Duration{Months: 12, Days: 1, nanos: 1000000},
			postgres:    "1 year 1 day 00:00:00.001",
			iso8601:     "P1Y1DT0.001S",
			sqlStandard: "+1-0 +1 +0:00:00.001",
		},
		// without years, months
		{
			duration:    Duration{Months: 0, Days: 1, nanos: 1000000},
			postgres:    "1 day 00:00:00.001",
			iso8601:     "P1DT0.001S",
			sqlStandard: "1 0:00:00.001",
		},
		// without days
		{
			duration:    Duration{Months: 35, Days: 0, nanos: 1000000},
			postgres:    "2 years 11 mons 00:00:00.001",
			iso8601:     "P2Y11MT0.001S",
			sqlStandard: "+2-11 +0 +0:00:00.001",
		},

		// mixed positive and negative units

		// PG prints '+' when a time unit changes the sign compared to the previous
		// unit, i.e. below CRDB should print +2 days (in 'postgres' style).
		{
			duration:    Duration{Months: -35, Days: 2, nanos: -12345678000000},
			postgres:    "-2 years -11 mons +2 days -03:25:45.678",
			iso8601:     "P-2Y-11M2DT-3H-25M-45.678S",
			sqlStandard: "-2-11 +2 -3:25:45.678",
		},
		{
			duration:    Duration{Months: 35, Days: -2, nanos: 12345678000000},
			postgres:    "2 years 11 mons -2 days +03:25:45.678",
			iso8601:     "P2Y11M-2DT3H25M45.678S",
			sqlStandard: "+2-11 -2 +3:25:45.678",
		},
		{
			duration:    Duration{Months: -35, Days: 2, nanos: 12345678000000},
			postgres:    "-2 years -11 mons +2 days 03:25:45.678",
			iso8601:     "P-2Y-11M2DT3H25M45.678S",
			sqlStandard: "-2-11 +2 +3:25:45.678",
		},
		// no year/month
		{
			duration:    Duration{Days: -2, nanos: 12345678000000},
			postgres:    "-2 days +03:25:45.678",
			iso8601:     "P-2DT3H25M45.678S",
			sqlStandard: "-2 +3:25:45.678",
		},
		// no days
		{
			duration:    Duration{Months: -35, nanos: 12345678000000},
			postgres:    "-2 years -11 mons +03:25:45.678",
			iso8601:     "P-2Y-11MT3H25M45.678S",
			sqlStandard: "-2-11 +0 +3:25:45.678",
		},
		{
			duration:    Duration{Days: -1, nanos: -123456789000000},
			postgres:    "-1 days -34:17:36.789",
			iso8601:     "P-1DT-34H-17M-36.789S",
			sqlStandard: "-1 34:17:36.789",
		},
		{
			duration:    Duration{Months: -12, Days: 3, nanos: -12345678000000},
			postgres:    "-1 years +3 days -03:25:45.678",
			iso8601:     "P-1Y3DT-3H-25M-45.678S",
			sqlStandard: "-1-0 +3 -3:25:45.678",
		},
		{
			duration:    Duration{Months: -12, Days: -3, nanos: 12345678000000},
			postgres:    "-1 years -3 days +03:25:45.678",
			iso8601:     "P-1Y-3DT3H25M45.678S",
			sqlStandard: "-1-0 -3 +3:25:45.678",
		},
		{
			duration:    Duration{Months: -12, Days: -3, nanos: 12345678000000},
			postgres:    "-1 years -3 days +03:25:45.678",
			iso8601:     "P-1Y-3DT3H25M45.678S",
			sqlStandard: "-1-0 -3 +3:25:45.678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.postgres, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.duration.FormatWithStyle(buf, IntervalStyle_POSTGRES)
			require.Equal(t, tt.postgres, buf.String())

			buf.Reset()
			tt.duration.Format(buf)
			require.Equal(t, tt.postgres, buf.String())

			buf.Reset()
			tt.duration.FormatWithStyle(buf, IntervalStyle_ISO_8601)
			require.Equal(t, tt.iso8601, buf.String())

			buf.Reset()
			tt.duration.FormatWithStyle(buf, IntervalStyle_SQL_STANDARD)
			require.Equal(t, tt.sqlStandard, buf.String())
		})
	}
}

func BenchmarkAdd(b *testing.B) {
	b.Run("fast-path-by-no-months-in-duration", func(b *testing.B) {
		s := time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC)
		d := Duration{Days: 1}
		for i := 0; i < b.N; i++ {
			Add(s, d)
		}
	})
	b.Run("fast-path-by-day-number", func(b *testing.B) {
		s := time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC)
		d := Duration{Months: 1}
		for i := 0; i < b.N; i++ {
			Add(s, d)
		}
	})
	b.Run("no-adjustment", func(b *testing.B) {
		s := time.Date(2018, 01, 31, 0, 0, 0, 0, time.UTC)
		d := Duration{Months: 2}
		for i := 0; i < b.N; i++ {
			Add(s, d)
		}
	})
	b.Run("with-adjustment", func(b *testing.B) {
		s := time.Date(2018, 01, 31, 0, 0, 0, 0, time.UTC)
		d := Duration{Months: 1}
		for i := 0; i < b.N; i++ {
			Add(s, d)
		}
	})
}

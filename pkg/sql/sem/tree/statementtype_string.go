// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// Code generated by "stringer"; DO NOT EDIT.

package tree

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeDDL-0]
	_ = x[TypeDML-1]
	_ = x[TypeDCL-2]
	_ = x[TypeTCL-3]
}

func (i StatementType) String() string {
	switch i {
	case TypeDDL:
		return "TypeDDL"
	case TypeDML:
		return "TypeDML"
	case TypeDCL:
		return "TypeDCL"
	case TypeTCL:
		return "TypeTCL"
	default:
		return "StatementType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

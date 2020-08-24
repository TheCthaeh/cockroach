// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package norm

import (
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
)

// ConcatLeftDeepAnds concatenates any left-deep And expressions in the right
// expression with any left-deep And expressions in the left expression. The
// result is a combined left-deep And expression. Note that NormalizeNestedAnds
// has already guaranteed that both inputs will already be left-deep.
func (c *CustomFuncs) ConcatLeftDeepAnds(left, right opt.ScalarExpr) opt.ScalarExpr {
	if and, ok := right.(*memo.AndExpr); ok {
		return c.f.ConstructAnd(c.ConcatLeftDeepAnds(left, and.Left), and.Right)
	}
	return c.f.ConstructAnd(left, right)
}

// NegateComparison negates a comparison op like:
//   a.x = 5
// to:
//   a.x <> 5
func (c *CustomFuncs) NegateComparison(
	cmp opt.Operator, left, right opt.ScalarExpr,
) opt.ScalarExpr {
	negate := opt.NegateOpMap[cmp]
	return c.f.DynamicConstruct(negate, left, right).(opt.ScalarExpr)
}

// FindRedundantConjunct takes the left and right operands of an Or operator as
// input. It examines each conjunct from the left expression and determines
// whether it appears as a conjunct in the right expression. If so, it returns
// the matching conjunct. Otherwise, it returns nil. For example:
//
//   A OR A                               =>  A
//   B OR A                               =>  nil
//   A OR (A AND B)                       =>  A
//   (A AND B) OR (A AND C)               =>  A
//   (A AND B AND C) OR (A AND (D OR E))  =>  A
//
// Once a redundant conjunct has been found, it is extracted via a call to the
// ExtractRedundantConjunct function. Redundant conjuncts are extracted from
// multiple nested Or operators by repeated application of these functions.
func (c *CustomFuncs) FindRedundantConjunct(left, right opt.ScalarExpr) opt.ScalarExpr {
	// Recurse over each conjunct from the left expression and determine whether
	// it's redundant.
	for {
		// Assume a left-deep And expression tree normalized by NormalizeNestedAnds.
		if and, ok := left.(*memo.AndExpr); ok {
			if c.isConjunct(and.Right, right) {
				return and.Right
			}
			left = and.Left
		} else {
			if c.isConjunct(left, right) {
				return left
			}
			return nil
		}
	}
}

// isConjunct returns true if the candidate expression is a conjunct within the
// given conjunction. The conjunction is assumed to be left-deep (normalized by
// the NormalizeNestedAnds rule).
func (c *CustomFuncs) isConjunct(candidate, conjunction opt.ScalarExpr) bool {
	for {
		if and, ok := conjunction.(*memo.AndExpr); ok {
			if and.Right == candidate {
				return true
			}
			conjunction = and.Left
		} else {
			return conjunction == candidate
		}
	}
}

// ExtractRedundantConjunct extracts a redundant conjunct from an Or expression,
// and returns an And of the conjunct with the remaining Or expression (a
// logically equivalent expression). For example:
//
//   (A AND B) OR (A AND C)  =>  A AND (B OR C)
//
// If extracting the conjunct from one of the OR conditions would result in an
// empty condition, the conjunct itself is returned (a logically equivalent
// expression). For example:
//
//   A OR (A AND B)  =>  A
//
// These transformations are useful for finding a conjunct that can be pushed
// down in the query tree. For example, if the redundant conjunct A is fully
// bound by one side of a join, it can be pushed through the join, even if B and
// C cannot.
func (c *CustomFuncs) ExtractRedundantConjunct(
	conjunct, left, right opt.ScalarExpr,
) opt.ScalarExpr {
	if conjunct == left || conjunct == right {
		return conjunct
	}

	return c.f.ConstructAnd(
		conjunct,
		c.f.ConstructOr(
			c.extractConjunct(conjunct, left.(*memo.AndExpr)),
			c.extractConjunct(conjunct, right.(*memo.AndExpr)),
		),
	)
}

// extractConjunct traverses the And subtree looking for the given conjunct,
// which must be present. Once it's located, it's removed from the tree, and
// the remaining expression is returned.
func (c *CustomFuncs) extractConjunct(conjunct opt.ScalarExpr, and *memo.AndExpr) opt.ScalarExpr {
	if and.Right == conjunct {
		return and.Left
	}
	if and.Left == conjunct {
		return and.Right
	}
	return c.f.ConstructAnd(c.extractConjunct(conjunct, and.Left.(*memo.AndExpr)), and.Right)
}

// FindConstrainedColumn takes the left and right operands of an Or operator
// as input. It examines each conjunct from the left expression and determines
// whether it contains a single outer column, which also appears in a conjunct
// as a single outer column in the right expression. If so, it returns the
// matching column ID. Otherwise, it returns 0. For example, suppose the column
// ID of x is 1:
//
//   x=1 OR x=2                                       =>  1
//   x=1 OR y=2                                       =>  0
//   x=1 OR (x=2 AND y=2)                             =>  1
//   (x=1 AND y=2) OR (x=2 AND z=3)                   =>  1
//   (x=1 AND y=2 AND z=3) OR (x=2 AND (a=1 OR b=2))  =>  1
//
// Once a constrained column has been found, a conjunct is extracted via a
// call to the ExtractConstrainedColumnConjunct function. Matching conjuncts
// are extracted from multiple nested Or operators by repeated application of
// these functions.
func (c *CustomFuncs) FindConstrainedColumn(
	left, right opt.ScalarExpr, unfilteredCols opt.ColSet,
) opt.ColSet {
	// Recurse over each conjunct from the left expression and determine whether
	// it has a single outer column that matches a conjunct on the right.
	for {
		// Assume a left-deep And expression tree normalized by NormalizeNestedAnds.
		if and, ok := left.(*memo.AndExpr); ok {
			if colSet, ok := c.isSingleVariableConjunct(and.Right, right); ok {
				if colSet.Intersects(unfilteredCols) {
					// Only return if the column is unfiltered to avoid pushing down the
					// new conjunct multiple times.
					return colSet
				}
			}
			left = and.Left
		} else {
			if colSet, ok := c.isSingleVariableConjunct(left, right); ok {
				if colSet.Intersects(unfilteredCols) {
					// Only return if the column is unfiltered to avoid pushing down the
					// new conjunct multiple times.
					return colSet
				}
			}
			return opt.ColSet{}
		}
	}
}

// isSingleVariableConjunct returns true if the candidate expression is a
// single-variable comparison within the given conjunction. The conjunction
// is assumed to be left-deep (normalized by the NormalizeNestedAnds rule).
func (c *CustomFuncs) isSingleVariableConjunct(
	candidate, conjunction opt.ScalarExpr,
) (_ opt.ColSet, ok bool) {
	outerCols := c.ScalarExprOuterCols(candidate)
	if outerCols.Len() != 1 {
		return opt.ColSet{}, false
	}
	for {
		if and, ok := conjunction.(*memo.AndExpr); ok {
			if c.ScalarExprOuterCols(and.Right).Equals(outerCols) {
				return outerCols, true
			}
			conjunction = and.Left
		} else {
			if c.ScalarExprOuterCols(conjunction).Equals(outerCols) {
				return outerCols, true
			}
			return opt.ColSet{}, false
		}
	}
}

// ScalarExprOuterCols calculates the set of outer columns from scratch for
// expressions that do not implement ScalarPropsExpr.
func (c *CustomFuncs) ScalarExprOuterCols(e opt.Expr) opt.ColSet {
	var cols opt.ColSet
	switch t := e.(type) {
	case *memo.VariableExpr:
		cols.Add(t.Col)
		return cols
	}
	for i, n := 0, e.ChildCount(); i < n; i++ {
		cols.UnionWith(c.ScalarExprOuterCols(e.Child(i)))
	}
	return cols
}

// ExtractConstrainedColumnConjunct extracts a conjunct containing a single
// variable from an Or expression, and returns the conjunct. For example:
//
//   (x=1 AND y=2) OR (x=2 AND z=3)  =>  (x=1 OR x=2)
//
// These transformations are useful for finding a conjunct that can be pushed
// down in the query tree. For example, if the conjunct (x=1 OR x=2) is fully
// bound by one side of a join, it can be pushed through the join, even if the
// original OR expression cannot.
func (c *CustomFuncs) ExtractConstrainedColumnConjunct(
	colSet opt.ColSet, left, right opt.ScalarExpr,
) opt.ScalarExpr {
	return c.f.ConstructOr(
		c.extractConjunctWithOuterCols(colSet, left),
		c.extractConjunctWithOuterCols(colSet, right),
	)
}

// extractConjunctWithOuterCols traverses the subtree looking for all
// conjuncts with the given set of outer columns. It returns them all
// as a left-deep tree of nested And expressions.
func (c *CustomFuncs) extractConjunctWithOuterCols(
	colSet opt.ColSet, expr opt.ScalarExpr,
) opt.ScalarExpr {
	if c.ScalarExprOuterCols(expr).Equals(colSet) {
		return expr
	}
	and, ok := expr.(*memo.AndExpr)
	if !ok {
		return nil
	}
	left := c.extractConjunctWithOuterCols(colSet, and.Left)
	if c.ScalarExprOuterCols(and.Right).Equals(colSet) {
		if left != nil {
			return c.f.ConstructAnd(left, and.Right)
		}
		return and.Right
	}
	return left
}

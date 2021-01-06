// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package optbuilder

import (
	"context"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

// udf represents a UDF expression in an expression tree after it has been
// type-checked and added to the memo.
type udf struct {
	// The resolved function expression.
	*tree.FuncExpr

	// cols contains the output columns of the UDF.
	cols []scopeColumn

	// expr is the memo expression defined by the UDF.
	expr memo.RelExpr

	// isScalar indicates whether the UDF returns a scalar output.
	isScalar bool
}

// Walk is part of the tree.Expr interface.
func (u *udf) Walk(v tree.Visitor) tree.Expr {
	return u
}

// TypeCheck is part of the tree.Expr interface.
func (u *udf) TypeCheck(
	_ context.Context, ctx *tree.SemaContext, desired *types.T,
) (tree.TypedExpr, error) {
  return u, nil
}

// Eval is part of the tree.TypedExpr interface.
func (u *udf) Eval(_ *tree.EvalContext) (tree.Datum, error) {
	panic(errors.AssertionFailedf("udf must be replaced before evaluation"))
}

var _ tree.Expr = &udf{}
var _ tree.TypedExpr = &udf{}

// buildScalarUdfJoins performs a series of lateral cross joins between the
// input of the given scope and each scalar UDF of the scope.
func (b *Builder) buildScalarUdfJoins(inScope *scope) {
	if len(inScope.udfs) == 0 {
		return
	}
	for i := range inScope.udfs {
		if !inScope.udfs[i].isScalar {
			continue
		}
		if inScope.expr == nil {
			inScope.expr = inScope.udfs[i].expr
			continue
		}
		inScope.expr = b.factory.ConstructInnerJoinApply(
			inScope.expr, inScope.udfs[i].expr, memo.TrueFilter, memo.EmptyJoinPrivate)
	}
}

// buildUdfZip returns an expression that reflects a functional zip of all
// UDFs in the given slice. If ignoreScalarUdfs is true, any scalar UDFs are not
// included in the zip. This is necessary because scalar and non-scalar UDFs
// behave differently when in the SELECT list.
func (b *Builder) buildUdfZip(udfs []*udf, ignoreScalarUdfs bool) memo.RelExpr {
	if len(udfs) == 0 {
		return nil
	}
	var result memo.RelExpr
	for i := range udfs {
		if ignoreScalarUdfs && udfs[i].isScalar {
			continue
		}
		if result == nil {
			result = udfs[i].expr
			continue
		}
		result = b.buildFunctionalZip(result, udfs[i].expr)
	}
	return result
}

// buildFunctionalZip returns an expression that generates a functional zip
// between the left and right expressions.
func (b *Builder) buildFunctionalZip(left, right memo.RelExpr) memo.RelExpr {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	left, leftOrdCol := b.wrapInOrdinality(left)
	right, rightOrdCol := b.wrapInOrdinality(right)
	joinCondition := memo.FiltersExpr{
		b.factory.ConstructFiltersItem(
			b.factory.ConstructEq(
				b.factory.ConstructVariable(leftOrdCol),
				b.factory.ConstructVariable(rightOrdCol),
			),
		),
	}
	return b.factory.ConstructFullJoin(left, right, joinCondition, memo.EmptyJoinPrivate)
}

// wrapInOrdinality constructs an Ordinality operator around the given
// expression and returns the resulting expression, as well as a column
// corresponding to the ordinality column.
func (b *Builder) wrapInOrdinality(inExpr memo.RelExpr) (memo.RelExpr, opt.ColumnID) {
	ordCol := b.factory.Metadata().AddColumn("rownum", types.Int)
	private := memo.OrdinalityPrivate{ColID: ordCol}
	retExpr := b.factory.ConstructOrdinality(inExpr, &private)
	return retExpr, ordCol
}

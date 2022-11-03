package testutils

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/norm"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/optgen/exprgen"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/testutils/testcat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/eval"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type testVal uint8
type testRow []testVal

type evalResult struct {
	cols opt.ColList
	rows []testRow
}

type testExprEvaluator struct {
	rnd *rand.Rand
}

func (ev *testExprEvaluator) init(mem *memo.Memo) {
	ev.rnd, _ = randutil.NewTestRand()
}

func (ev *testExprEvaluator) evaluateExpr(expr memo.RelExpr) (res evalResult, err error) {
	switch t := expr.(type) {
	case *memo.ScanExpr:
		return ev.evaluateScan(t)
	case *memo.InnerJoinExpr:
		return ev.evaluateJoin(t)
	}
	return evalResult{}, errors.AssertionFailedf("unexpected expression: %v", expr.Op())
}

func (ev *testExprEvaluator) evaluateScan(expr *memo.ScanExpr) (res evalResult, err error) {
	res.cols = expr.Cols.ToList()
	res.rows = ev.genRandTestRows(len(res.cols))
	return res, nil
}

func (ev *testExprEvaluator) evaluateJoin(expr memo.RelExpr) (res evalResult, err error) {
	leftExpr, rightExpr := expr.Child(0).(memo.RelExpr), expr.Child(1).(memo.RelExpr)
	joinFilters := expr.Child(2).(*memo.FiltersExpr)
	var left, right evalResult
	if left, err = ev.evaluateExpr(leftExpr); err != nil {
		return evalResult{}, err
	}
	if right, err = ev.evaluateExpr(rightExpr); err != nil {
		return evalResult{}, err
	}
	res.cols = append(left.cols, right.cols...)
	for _, leftRow := range left.rows {
		for _, rightRow := range right.rows {
			joinRow := make(testRow, len(leftRow)+len(rightRow))
			copy(joinRow, leftRow)
			copy(joinRow[:len(leftRow)], rightRow)
			matched, err := ev.evaluateFilters(joinRow, res.cols, *joinFilters)
			if err != nil {
				return evalResult{}, err
			}

			switch expr.Op() {
			case opt.InnerJoinOp:
				res.rows
			}
		}
	}
	return res, nil
}

func (ev *testExprEvaluator) evaluateFilters(
	row testRow, cols opt.ColList, filters memo.FiltersExpr,
) (matched bool, err error) {
	getVal := func(id opt.ColumnID) (val testVal, err error) {
		for i, col := range cols {
			if col == id {
				return row[i], nil
			}
		}
		return 0, errors.AssertionFailedf("couldn't find value for column %d", id)
	}

	for i := range filters {
		eqExpr, ok := filters[i].Condition.(*memo.EqExpr)
		if !ok {
			return false, errors.AssertionFailedf("testExprEvaluator only handles equality conditions")
		}
		leftVar, leftOk := eqExpr.Left.(*memo.VariableExpr)
		rightVar, rightOk := eqExpr.Right.(*memo.VariableExpr)
		if !leftOk || !rightOk {
			return false, errors.AssertionFailedf("testExprEvaluator only handles simple comparisons")
		}
		var leftVal, rightVal testVal
		if leftVal, err = getVal(leftVar.Col); err != nil {
			return false, err
		}
		if rightVal, err = getVal(rightVar.Col); err != nil {
			return false, err
		}
		if leftVal != rightVal {
			return false, nil
		}
	}
}

const maxRows = 1000
const maxVal = 50

// TODO(drewk): allow for keys and other FDs
func (ev *testExprEvaluator) genRandTestRows(numCols int) (rows []testRow) {
	rowCount := ev.rnd.Intn(maxRows + 1)
	rows = make([]testRow, rowCount)
	for rowIdx := range rows {
		rows[rowIdx] = make(testRow, numCols)
		for colIdx := range rows[rowIdx] {
			rows[rowIdx][colIdx] = testVal(ev.rnd.Intn(maxVal + 1))
		}
	}
	return rows
}

func TestStuff(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	catalog := testcat.New()
	evalCtx := eval.MakeTestingEvalContext(cluster.MakeTestingClusterSettings())

	var f norm.Factory
	f.Init(ctx, &evalCtx, catalog)

	var ev testExprEvaluator
	ev.init(f.Memo())

	_, err := catalog.ExecuteDDL("CREATE TABLE abc (a INT, b INT, c INT, INDEX ab(a, b))")
	require.NoError(t, err)
	_, err = catalog.ExecuteDDL("CREATE TABLE def (d INT, e INT, f INT)")
	require.NoError(t, err)

	expr, err := exprgen.Build(ctx, catalog, &f, `
	(InnerJoin
  	(Scan [ (Table "abc") (Cols "a,b,c") ])
  	(Scan [ (Table "def") (Cols "d,e,f") ])
  	[ (Eq (Var "a") (Var "d")) ]
  	[ ]
	)`)
	require.NoError(t, err)
	fmt.Println(expr)
	res, err := ev.evaluateExpr(expr.(memo.RelExpr))
	require.NoError(t, err)
	fmt.Println(res.cols.ToSet())
	fmt.Println("=================")
	for i := range res.rows {
		fmt.Println(res.rows[i])
	}
}

func TestMain(m *testing.M) {
	randutil.SeedForTests()
	os.Exit(m.Run())
}

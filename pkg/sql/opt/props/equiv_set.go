// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package props

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

// EquivSet describes a set of equivalence groups of columns. It can answer
// queries about which columns are equivalent to one another. Equivalence groups
// are always non-empty and disjoint. The ColSets used to represent equivalence
// groups should be considered immutable; in order to update them, they should
// be replaced with different ColSets.
type EquivSet struct {
	buf    [equalityBufferSize]opt.ColSet
	groups []opt.ColSet
}

const equalityBufferSize = 1

// NewEquivSet returns a new equality set with a starting capacity of one
// equivalence group. This optimizes for the common case when only one
// equivalence group is stored.
func NewEquivSet() EquivSet {
	set := EquivSet{}
	set.groups = set.buf[:0]
	return set
}

// Reset prepares the EquivSet for reuse, maintaining references to any
// allocated slice memory.
func (eq *EquivSet) Reset() {
	for i := range eq.groups {
		// Release any references to the large portion of ColSets.
		eq.groups[i] = opt.ColSet{}
	}
	eq.groups = eq.groups[:0]
}

// Add adds the given equivalent columns to the EquivSet. If possible, the
// columns are added to an existing group. Otherwise, a new one is created.
func (eq *EquivSet) Add(equivCols opt.ColSet) {
	if equivCols.Len() <= 1 {
		return
	}

	// Attempt to add the equivalence to an existing group.
	for i := range eq.groups {
		if eq.groups[i].Intersects(equivCols) {
			if equivCols.SubsetOf(eq.groups[i]) {
				// No-op
				return
			}
			eq.groups[i] = eq.groups[i].Union(equivCols)
			eq.tryMergeGroups(i)
			return
		}
	}
	// Make a new equivalence group.
	eq.groups = append(eq.groups, equivCols.Copy())
}

// AddFromFDs adds all equivalence relations from the given FuncDepSet to the
// EquivSet.
func (eq *EquivSet) AddFromFDs(fdset *FuncDepSet) {
	eq.AddFrom(&fdset.equiv)
}

// AddFrom adds all equivalence relations from the given EquivSet.
func (eq *EquivSet) AddFrom(other *EquivSet) {
	for i := range other.groups {
		eq.Add(other.groups[i])
	}
}

// AreColsEquiv indicates whether the given columns are equivalent.
func (eq *EquivSet) AreColsEquiv(left, right opt.ColumnID) bool {
	if left == right {
		return true
	}
	return eq.areColsInSetEquiv(opt.MakeColSet(left, right))
}

func (eq *EquivSet) areColsInSetEquiv(cols opt.ColSet) bool {
	if cols.Len() <= 1 {
		return true
	}
	for i := range eq.groups {
		if cols.SubsetOf(eq.groups[i]) {
			return true
		}
	}
	return false
}

// computeEquivClosure returns the equivalence closure of the given columns.
// This includes the given columns, as well as all columns that are equal to any
// of the given columns.
func (eq *EquivSet) computeEquivClosure(cols opt.ColSet) opt.ColSet {
	closure := cols.Copy()
	for i := range eq.groups {
		if eq.groups[i].Intersects(cols) {
			closure.UnionWith(eq.groups[i])
		}
	}
	return closure
}

// equivReps returns one "representative" column from each equivalency group.
func (eq *EquivSet) equivReps() opt.ColSet {
	var reps opt.ColSet
	for i := range eq.groups {
		id, ok := eq.groups[i].Next(0)
		if !ok {
			panic(errors.AssertionFailedf("empty equivalence group"))
		}
		reps.Add(id)
	}
	return reps
}

// makeDisjoint splits each equivalency group into two disjoint equivalency
// groups: the intersection and difference with the given ColSet respectively.
// It is used by FuncDepSet when modifying the FDs to reflect the impact of
// null-extending an input of a join.
func (eq *EquivSet) makeDisjoint(cols opt.ColSet) {
	for i := len(eq.groups) - 1; i >= 0; i-- {
		if !eq.groups[i].Intersects(cols) || eq.groups[i].SubsetOf(cols) {
			// No-op case - the equivalency group is unchanged by the split.
			continue
		}
		intersection := eq.groups[i].Intersection(cols)
		difference := eq.groups[i].Difference(cols)
		intersectionTrivial := intersection.Len() <= 1
		differenceTrivial := difference.Len() <= 1
		if intersectionTrivial && differenceTrivial {
			// Remove the current group entirely.
			eq.groups[i] = eq.groups[len(eq.groups)-1]
			eq.groups = eq.groups[:len(eq.groups)-1]
			continue
		}
		if intersectionTrivial {
			// Replace the current group with the difference group.
			eq.groups[i] = difference
			continue
		}
		if differenceTrivial {
			// Replace the current group with the intersection group.
			eq.groups[i] = intersection
			continue
		}
		// Both new equivalency groups are nontrivial.
		eq.groups[i] = intersection
		eq.groups = append(eq.groups, difference)
	}
}

// projectCols restricts the EquivSet to only refer to columns from the given
// set. Some equivalence groups may be removed if they become trivial after the
// projection (e.g. zero or one columns).
func (eq *EquivSet) projectCols(cols opt.ColSet) {
	for i := len(eq.groups) - 1; i >= 0; i-- {
		if eq.groups[i].SubsetOf(cols) {
			continue
		}
		eq.groups[i] = eq.groups[i].Intersection(cols)
		if eq.groups[i].Len() <= 1 {
			// This equivalence group is now trivial. Remove it by replacing it with
			// the last equivalence group and reslicing. Skip the index increment,
			// since we also want to project the group we just moved.
			eq.groups[i] = eq.groups[len(eq.groups)-1]
			eq.groups = eq.groups[:len(eq.groups)-1]
		}
	}
}

// tryMergeGroups attempts to merge the equality group at the given index with
// any of the *following* groups. If a group can be merged, it is removed after
// its columns are added to the given group. It is safe to modify the equality
// group, since it will have been replaced by Add.
func (eq *EquivSet) tryMergeGroups(idx int) {
	for i := idx + 1; i < len(eq.groups); i++ {
		if eq.groups[idx].Intersects(eq.groups[i]) {
			eq.groups[idx].UnionWith(eq.groups[i])
			eq.groups[i] = eq.groups[len(eq.groups)-1]
			eq.groups[len(eq.groups)-1] = opt.ColSet{}
			eq.groups = eq.groups[:len(eq.groups)-1]
		}
	}
}

// verify checks the invariants that the equivalency groups are disjoint and
// non-empty.
func (eq *EquivSet) verify() {
	var seenCols opt.ColSet
	for i := range eq.groups {
		if eq.groups[i].Len() <= 1 {
			panic(errors.AssertionFailedf(
				"expected non-trivial equivalency group:",
				redact.Safe(eq.groups[i]),
			))
		}
		if eq.groups[i].Intersects(seenCols) {
			panic(errors.AssertionFailedf(
				"expected columns to be in at most one equivalency group:",
				redact.Safe(eq.groups[i].Intersection(seenCols)),
			))
		}
	}
}

func (eq *EquivSet) format(b *strings.Builder) {
	isFirst := true
	for i := range eq.groups {
		group := eq.groups[i]
		for col, ok := group.Next(0); ok; col, ok = group.Next(col + 1) {
			if !isFirst {
				b.WriteString(", ")
			}
			isFirst = false
			from := opt.MakeColSet(col)
			to := group.Copy()
			to.Remove(col)
			fmt.Fprintf(b, "%s==%s", from, to)
		}
	}
}

func (eq *EquivSet) String() string {
	var b strings.Builder
	eq.format(&b)
	return b.String()
}

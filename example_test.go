// Copyright 2015 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prefixtree_test

import (
	"fmt"

	"github.com/beevik/prefixtree"
)

// Build a prefix tree and search it for various prefix matches.
func ExampleTree_usage() {
	// Create the tree. Add 5 strings, each with an associated
	// integer.
	tree := prefixtree.New()
	for i, s := range []string{
		"apple",
		"orange",
		"apple pie",
		"lemon meringue",
		"lemon",
	} {
		tree.Add(s, i)
	}

	// Attempt to find various prefixes in the tree, and output
	// the result.
	fmt.Printf("%-18s %-8s %s\n", "prefix", "data", "error")
	fmt.Printf("%-18s %-8s %s\n", "------", "----", "-----")
	for _, s := range []string{
		"a",
		"appl",
		"apple",
		"apple p",
		"apple pie",
		"apple pies",
		"o",
		"orang",
		"orange",
		"oranges",
		"lemo",
		"lemon",
		"lemon m",
		"lemon meringue",
		"lemon meringues",
	} {
		data, err := tree.Find(s)
		fmt.Printf("%-18s %-8v %v\n", s, data, err)
	}

	// Output:
	// prefix             data     error
	// ------             ----     -----
	// a                  <nil>    prefixtree: prefix ambiguous
	// appl               <nil>    prefixtree: prefix ambiguous
	// apple              0        <nil>
	// apple p            2        <nil>
	// apple pie          2        <nil>
	// apple pies         <nil>    prefixtree: prefix not found
	// o                  1        <nil>
	// orang              1        <nil>
	// orange             1        <nil>
	// oranges            <nil>    prefixtree: prefix not found
	// lemo               <nil>    prefixtree: prefix ambiguous
	// lemon              4        <nil>
	// lemon m            3        <nil>
	// lemon meringue     3        <nil>
	// lemon meringues    <nil>    prefixtree: prefix not found
}

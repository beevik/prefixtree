// Copyright 2015-2023 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prefixtree

import (
	"bufio"
	"math/rand"
	"os"
	"testing"
)

type entry struct {
	key   string
	value any
}

type result struct {
	key   string
	value any
	err   error
}

func test(t *testing.T, entries []entry, results []result) *Tree {
	// Run 256 iterations of build/find using random tree entry
	// insertion orders.
	var tree *Tree
	for i := 0; i < 256; i++ {
		tree = New()
		for _, i := range rand.Perm(len(entries)) {
			tree.Add(entries[i].key, entries[i].value)
		}

		fail := false
		for _, expected := range results {
			value, err := tree.Find(expected.key)
			if expected.err != nil {
				if err != expected.err {
					fail = true
					t.Errorf("Find(\"%s\") returned error [%v], expected error [%v].\n",
						expected.key, err, expected.err)
				}
			} else {
				if err != nil {
					fail = true
					t.Errorf("Find(\"%s\") returned error [%v], expected value %d.\n",
						expected.key, err, expected.value)
				} else if value != expected.value {
					fail = true
					t.Errorf("Find(\"%s\") returned value %d, expected value %d.\n",
						expected.key, value, expected.value)
				}
			}
		}

		if fail {
			t.Errorf("Failures listed above occurred during iteration %d.\n", i)
			break
		}
	}

	// Return the last tree created in case further tests are to be done with
	// it.
	return tree
}

func TestAdd(t *testing.T) {
	test(
		t,
		[]entry{
			{"apple", 1},
			{"applepie", 2},
			{"a", 3},
			{"armor", 4},
		},
		[]result{
			{"a", 3, nil},
			{"ap", 0, ErrPrefixAmbiguous},
			{"app", 0, ErrPrefixAmbiguous},
			{"appl", 0, ErrPrefixAmbiguous},
			{"apps", 0, ErrPrefixNotFound},
			{"apple", 1, nil},
			{"applep", 2, nil},
			{"applepi", 2, nil},
			{"applepie", 2, nil},
			{"applepies", 0, ErrPrefixNotFound},
			{"applepix", 0, ErrPrefixNotFound},
			{"ar", 4, nil},
			{"arm", 4, nil},
			{"armo", 4, nil},
			{"armor", 4, nil},
			{"armors", 0, ErrPrefixNotFound},
			{"armx", 0, ErrPrefixNotFound},
			{"ax", 0, ErrPrefixNotFound},
			{"b", 0, ErrPrefixNotFound},
			{"p", 0, ErrPrefixNotFound},
			{"pple", 0, ErrPrefixNotFound},
			{"", 0, ErrPrefixAmbiguous},
		})
}

func TestSplit(t *testing.T) {
	test(
		t,
		[]entry{
			{"abc", 1},
			{"ab", 2},
		},
		[]result{
			{"a", 0, ErrPrefixAmbiguous},
			{"ab", 2, nil},
			{"abc", 1, nil},
		})
}

func TestLargeDegree(t *testing.T) {
	test(
		t,
		[]entry{
			{"-a", 1},
			{"-b", 2},
			{"-c", 3},
			{"-d", 4},
			{"-e", 5},
			{"-f", 6},
			{"-g", 7},
			{"-h", 8},
			{"-i", 9},
			{"-j", 10},
			{"-k", 11},
			{"-dog", 12},
			{"-l", 13},
			{"-m", 14},
			{"-n", 15},
			{"-o", 16},
			{"-p", 17},
			{"-q", 18},
			{"-r", 19},
			{"-s", 20},
			{"-t", 21},
			{"-u", 22},
			{"-v", 23},
			{"-w", 24},
			{"-x", 25},
			{"-y", 26},
			{"-z", 27},
		},
		[]result{
			{"-", 0, ErrPrefixAmbiguous},
			{"-a", 1, nil},
			{"-b", 2, nil},
			{"-c", 3, nil},
			{"-d", 4, nil},
			{"-e", 5, nil},
			{"-f", 6, nil},
			{"-g", 7, nil},
			{"-h", 8, nil},
			{"-i", 9, nil},
			{"-j", 10, nil},
			{"-k", 11, nil},
			{"-dog", 12, nil},
			{"-do", 12, nil},
			{"-l", 13, nil},
			{"-m", 14, nil},
			{"-n", 15, nil},
			{"-o", 16, nil},
			{"-p", 17, nil},
			{"-q", 18, nil},
			{"-r", 19, nil},
			{"-s", 20, nil},
			{"-t", 21, nil},
			{"-u", 22, nil},
			{"-v", 23, nil},
			{"-w", 24, nil},
			{"-x", 25, nil},
			{"-y", 26, nil},
			{"-z", 27, nil},
		})
}

func TestFindAllValues(t *testing.T) {
	entries := []entry{
		{"apple", 1},
		{"applepie", 2},
		{"a", 3},
		{"arm", 4},
		{"bee", 5},
	}

	tree := New()
	for _, entry := range entries {
		tree.Add(entry.key, entry.value)
	}

	results := []struct {
		key    string
		values []any
	}{
		{"", []any{3, 1, 2, 4, 5}},
		{"a", []any{3, 1, 2, 4}},
		{"ap", []any{1, 2}},
		{"app", []any{1, 2}},
		{"appl", []any{1, 2}},
		{"apple", []any{1, 2}},
		{"applep", []any{2}},
		{"applepi", []any{2}},
		{"applepie", []any{2}},
		{"applepies", []any{}},
		{"ar", []any{4}},
		{"arm", []any{4}},
		{"arms", []any{}},
		{"b", []any{5}},
		{"be", []any{5}},
		{"bee", []any{5}},
		{"bees", []any{}},
		{"c", []any{}},
	}

	for _, expected := range results {
		values := tree.FindAllValues(expected.key)
		match := false
		if len(values) == len(expected.values) {
			match = true
			for i := 0; i < len(values); i++ {
				if values[i] != expected.values[i] {
					match = false
					break
				}
			}
		}

		if !match {
			t.Errorf("FindAllValues(\"%s\") returned %v, expected %v.\n",
				expected.key, values, expected.values)
		}
	}
}

func TestMatchingChars(t *testing.T) {
	type test struct {
		s1     string
		s2     string
		result int
	}
	var tests = []test{
		{"a", "ap", 1},
		{"ap", "ap", 2},
		{"app", "ap", 2},
		{"apple", "ap", 2},
		{"ap", "a", 1},
		{"apple", "a", 1},
		{"apple", "bag", 0},
	}
	for _, test := range tests {
		r := matchingChars(test.s1, test.s2)
		if r != test.result {
			t.Errorf("matchingChars(\"%s\", \"%s\") returned %d, expected %d\n",
				test.s1, test.s2, r, test.result)
		}
	}
}

func TestDictionary(t *testing.T) {
	// Attempt to open the unix words dictionary file. If it doesn't
	// exist, skip this test.
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		return
	}

	// Scan all words from the dictionary into the tree.
	tree := New()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tree.Add(scanner.Text(), nil)
	}
	file.Close()

	// Find some prefixes that should be unambiguous and in the
	// dictionary.
	var keys = []string{
		"zebra",
		"axe",
		"diamond",
		"big",
		"diatribe",
		"diametrical",
		"diametricall",
		"diametrically",
	}
	for _, key := range keys {
		_, err := tree.Find(key)
		if err != nil {
			t.Errorf("Find(\"%s\") encountered error: %v\n", key, err)
		}
	}

	// Find some prefixes that should be ambiguous.
	keys = []string{
		"ab",
		"co",
		"de",
		"dea",
	}
	for _, key := range keys {
		_, err := tree.Find(key)
		if err != ErrPrefixAmbiguous {
			t.Errorf("Find(\"%s\") should have been ambiguous\n", key)
		}
	}
}

func BenchmarkDictionary(b *testing.B) {
	// This benchmark is used to determine the binary
	// search cutoff point for the Find function.
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		return
	}

	// Scan all words from the dictionary into the tree.
	tree := New()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tree.Add(scanner.Text(), nil)
	}
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var keys = []string{
			"zebra",
			"axe",
			"diamond",
			"big",
			"diatribe",
			"diametrical",
			"diametricall",
			"diametrically",
			"scene",
			"altar",
			"pituitary",
			"yellow",
			"target",
			"greedy",
			"oracle",
			"ruddy",
		}
		for _, key := range keys {
			_, err := tree.Find(key)
			if err != nil {
				b.Errorf("Find(\"%s\") encountered error: %v\n", key, err)
			}
		}
	}
}

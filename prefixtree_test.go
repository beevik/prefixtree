// Copyright 2015 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prefixtree

import (
	"bufio"
	"math/rand"
	"os"
	"testing"
)

type testEntry struct {
	s     string
	value int
}

type testFind struct {
	s    string
	data int
	err  error
}

func buildTree(entries []testEntry) *Tree {
	// Add the entries to the tree in random order.
	tree := New()
	for _, i := range rand.Perm(len(entries)) {
		tree.Add(entries[i].s, entries[i].value)
	}
	return tree
}

func testTree(t *testing.T, tree *Tree, tests []testFind) {
	for _, test := range tests {
		data, err := tree.Find(test.s)
		if err != nil && err != test.err {
			t.Errorf("Find(\"%s\") returned [%v], expected [%v]\n",
				test.s, err, test.err)
		}
		if err == nil && data.(int) != test.data {
			t.Errorf("Find(\"%s\") returned %d, expected %d\n",
				test.s, data.(int), test.data)
		}
	}
}

func TestAdd(t *testing.T) {
	tree := buildTree([]testEntry{
		{"apple", 1},
		{"applepie", 2},
		{"a", 3},
		{"armor", 4},
	})
	testTree(t, tree, []testFind{
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
		{"", 0, ErrPrefixAmbiguous},
	})
}

func TestSplit(t *testing.T) {
	tree := buildTree([]testEntry{
		{"abc", 1},
		{"ab", 2},
	})
	find := []testFind{
		{"a", 0, ErrPrefixAmbiguous},
		{"ab", 2, nil},
		{"abc", 1, nil},
	}
	testTree(t, tree, find)
}

func TestLargeDegree(t *testing.T) {
	tree := buildTree([]testEntry{
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
	})
	testTree(t, tree, []testFind{
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
	var tests = []string{
		"zebra",
		"axe",
		"diamond",
		"big",
		"diatribe",
		"diametrical",
		"diametricall",
		"diametrically",
	}
	for _, test := range tests {
		_, err := tree.Find(test)
		if err != nil {
			t.Errorf("Find(\"%s\") encountered error: %v\n", test, err)
		}
	}

	// Find some prefixes that should be ambiguous.
	tests = []string{
		"ab",
		"co",
		"de",
		"dea",
	}
	for _, test := range tests {
		_, err := tree.Find(test)
		if err != ErrPrefixAmbiguous {
			t.Errorf("Find(\"%s\") should have been ambiguous\n", test)
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
		var tests = []string{
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
		for _, test := range tests {
			_, err := tree.Find(test)
			if err != nil {
				b.Errorf("Find(\"%s\") encountered error: %v\n", test, err)
			}
		}
	}
}

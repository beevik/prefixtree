// Copyright 2015 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package prefixtree implements a prefix tree (technically, a trie). A prefix
// tree enables rapid searching for strings that uniquely match a given
// prefix. This implementation allows the user to associate data with each
// string, so it can act as a sort of flexible key-value store.
package prefixtree

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrPrefixNotFound is returned by Find if the prefix being searched for
	// matches zero strings in the prefix tree.
	ErrPrefixNotFound = errors.New("prefixtree: prefix not found")

	// ErrPrefixAmbiguous is returned by Find if the prefix being
	// searched for matches more than one string in the prefix tree.
	ErrPrefixAmbiguous = errors.New("prefixtree: prefix ambiguous")
)

// A Tree represents a prefix tree containing strings and their associated
// data. The tree is implemented as a trie and can be searched efficiently for
// unique prefix matches.
type Tree struct {
	links       []link
	data        any
	terminal    bool
	descendants int
}

type link struct {
	str  string
	tree *Tree
}

// New returns an empty prefix tree.
func New() *Tree {
	return new(Tree)
}

// matchingChars returns the number of shared characters in s1 and s2,
// starting from the beginning of each string.
func matchingChars(s1, s2 string) int {
	i := 0
	for l := minInt(len(s1), len(s2)); i < l; i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return i
}

// Find searches the prefix tree for a string that uniquely matches the
// prefix. If found, the data associated with the string is returned. If not
// found, ErrPrefixNotFound is returned. If the prefix matches more than one
// string in the tree, ErrPrefixAmbiguous is returned.
func (t *Tree) Find(prefix string) (data any, err error) {
outerLoop:
	for {
		// Ran out of prefix? Then return data if this node is terminal.
		if len(prefix) == 0 {
			if t.terminal {
				return t.data, nil
			} else {
				return nil, ErrPrefixAmbiguous
			}
		}

		// Figure out which links to consider. Do a binary search for 2
		// candidate links if the number of links from the node is large-ish
		// (20+). Otherwise search all links. The cutoff point between binary
		// and linear search was determined by benchmarking against the unix
		// english dictionary.
		var start, stop int
		if len(t.links) >= 20 {
			ix := sort.Search(len(t.links),
				func(i int) bool { return t.links[i].str >= prefix })
			start, stop = maxInt(0, ix-1), minInt(ix, len(t.links)-1)
		} else {
			start, stop = 0, len(t.links)-1
		}

		// Perform the check on all candidate links.
	innerLoop:
		for i := start; i <= stop; i++ {
			link := &t.links[i]
			m := matchingChars(prefix, link.str)
			switch {
			case m == 0:
				continue innerLoop
			case m == len(link.str):
				// Full link match, so proceed down subtree.
				t, prefix = link.tree, prefix[m:]
				continue outerLoop
			case m == len(prefix):
				// Remaining prefix fully consumed, so return data unless it's
				// non-terminal or ambiguous.
				switch {
				case link.tree.descendants > 1:
					return nil, ErrPrefixAmbiguous
				case link.tree.terminal:
					return link.tree.data, nil
				default:
					return nil, ErrPrefixNotFound
				}
			}
		}
		return nil, ErrPrefixNotFound
	}
}

// Add a string and its associated data to the prefix tree.
func (t *Tree) Add(s string, data any) {
outerLoop:
	for {
		t.descendants++

		// If we've consumed the entire string, then the tree node is terminal
		// and we're done.
		if len(s) == 0 {
			t.terminal, t.data = true, data
			break outerLoop
		}

		// Find the lexicographical link insertion point.
		ix := sort.Search(len(t.links),
			func(i int) bool { return t.links[i].str >= s })

		// Check the links before and after the insertion point for a matching
		// prefix to see if we need to split one of them.
		var splitLink *link
		var splitIndex int
	innerLoop:
		for li, lm := maxInt(ix-1, 0), minInt(ix, len(t.links)-1); li <= lm; li++ {
			link := &t.links[li]
			m := matchingChars(link.str, s)
			switch {
			case m == len(link.str):
				// Full link match, so proceed down the subtree.
				t, s = link.tree, s[m:]
				continue outerLoop
			case m > 0:
				// Partial match, so we'll need to split this tree node.
				splitLink, splitIndex = link, m
				break innerLoop
			}
		}

		// No split necessary, so insert a new link and subtree.
		if splitLink == nil {
			subtree := &Tree{data: data, terminal: true, descendants: 1}
			t.links = append(t.links[:ix],
				append([]link{{s, subtree}}, t.links[ix:]...)...)
			break outerLoop
		}

		// A split is necessary, so split the current link's string and insert
		// a child tree.
		s1, s2 := splitLink.str[:splitIndex], splitLink.str[splitIndex:]
		child := &Tree{
			links:       []link{{s2, splitLink.tree}},
			descendants: splitLink.tree.descendants,
		}
		splitLink.str, splitLink.tree = s1, child
		t, s = child, s[splitIndex:]
	}
}

// Output the structure of the tree to stdout. This function exists for
// debugging purposes.
func (t *Tree) Output() {
	t.outputNode(0)
}

func (t *Tree) outputNode(level int) {
	fmt.Printf("%sNode: term=%v desc=%d data=%v\n",
		strings.Repeat("  ", level), t.terminal, t.descendants, t.data)
	for i, l := range t.links {
		fmt.Printf("%s Link %d: s=%s\n",
			strings.Repeat("  ", level), i, l.str)
		l.tree.outputNode(level + 1)
	}
}

func minInt(a, b int) int {
	switch {
	case a < b:
		return a
	default:
		return b
	}
}

func maxInt(a, b int) int {
	switch {
	case a > b:
		return a
	default:
		return b
	}
}

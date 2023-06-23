// Copyright 2015-2023 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package prefixtree implements a prefix tree (technically, a trie). A prefix
// tree enables rapid searching for key strings that uniquely match a given
// prefix. This implementation allows the user to associate value data with
// each key string, so it can act as a sort of flexible key-value store.
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
// value data. The tree is implemented as a trie and can be searched
// efficiently for unique prefix matches.
type Tree struct {
	links       []link
	value       any
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

// Find searches the prefix tree for a key string that uniquely matches the
// prefix. If found, the value data associated with the key is returned. If
// not found, ErrPrefixNotFound is returned. If the prefix matches more than
// one key in the tree, ErrPrefixAmbiguous is returned.
func (t *Tree) Find(prefix string) (value any, err error) {
	st, err := t.findSubtree(prefix)
	if err != nil {
		return nil, err
	}
	return st.value, nil
}

// FindAll searches the prefix tree for all key strings prefixed by the
// provided prefix. All matching values are returned.
func (t *Tree) FindAll(prefix string) (values []any) {
	st, err := t.findSubtree(prefix)
	if err == ErrPrefixNotFound {
		return []any{}
	}
	return appendDescendantValues(st, nil)
}

// findSubtree searches the prefix tree for the deepest subtree matching
// the prefix.
func (t *Tree) findSubtree(prefix string) (*Tree, error) {
outerLoop:
	for {
		// Ran out of prefix? Then return value data only if this node is
		// terminal.
		if len(prefix) == 0 {
			if t.terminal {
				return t, nil
			} else {
				return t, ErrPrefixAmbiguous
			}
		}

		// Figure out which links to consider. If the number of links from the
		// node is large-ish (20+), do a binary search for 2 candidate links.
		// Otherwise search all links. The cutoff point between binary and
		// linear search was determined by benchmarking against the unix
		// english dictionary.
		start, stop := 0, len(t.links)-1
		if len(t.links) >= 20 {
			ix := sort.Search(len(t.links),
				func(i int) bool { return t.links[i].str >= prefix })
			start, stop = maxInt(0, ix-1), minInt(ix, stop)
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
				// Full link 1, so proceed down subtree.
				t, prefix = link.tree, prefix[m:]
				continue outerLoop
			case m == len(prefix):
				// Remaining prefix fully consumed, so return value data
				// unless it's non-terminal or ambiguous.
				switch {
				case link.tree.descendants > 1:
					return link.tree, ErrPrefixAmbiguous
				case link.tree.terminal:
					return link.tree, nil
				default:
					return nil, ErrPrefixNotFound
				}
			}
		}
		return nil, ErrPrefixNotFound
	}
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

// appendDescendantValues recursively appends a tree's descendant values
// to an array of values.
func appendDescendantValues(t *Tree, values []any) []any {
	if t.terminal {
		values = append(values, t.value)
	}
	for i := 0; i < len(t.links); i++ {
		values = appendDescendantValues(t.links[i].tree, values)
	}
	return values
}

// Add a key string and its associated value data to the prefix tree.
func (t *Tree) Add(key string, value any) {
outerLoop:
	for {
		t.descendants++

		// If we've consumed the entire string, then the tree node is terminal
		// and we're done.
		if len(key) == 0 {
			t.terminal, t.value = true, value
			break outerLoop
		}

		// Find the lexicographical link insertion point.
		ix := sort.Search(len(t.links),
			func(i int) bool { return t.links[i].str >= key })

		// Check the links before and after the insertion point for a matching
		// prefix to see if we need to split one of them.
		var splitLink *link
		var splitIndex int
	innerLoop:
		for li, lm := maxInt(ix-1, 0), minInt(ix, len(t.links)-1); li <= lm; li++ {
			link := &t.links[li]
			m := matchingChars(link.str, key)
			switch {
			case m == len(link.str):
				// Full link match, so proceed down the subtree.
				t, key = link.tree, key[m:]
				continue outerLoop
			case m > 0:
				// Partial match, so we'll need to split this tree node.
				splitLink, splitIndex = link, m
				break innerLoop
			}
		}

		// No split necessary, so insert a new link and subtree.
		if splitLink == nil {
			subtree := &Tree{value: value, terminal: true, descendants: 1}
			t.links = append(t.links[:ix],
				append([]link{{key, subtree}}, t.links[ix:]...)...)
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
		t, key = child, key[splitIndex:]
	}
}

// Output the structure of the tree to stdout. This function exists for
// debugging purposes.
func (t *Tree) Output() {
	t.outputNode(0)
}

func (t *Tree) outputNode(level int) {
	fmt.Printf("%sNode: term=%v desc=%d value=%v\n",
		strings.Repeat("    ", level), t.terminal, t.descendants, t.value)
	for i, l := range t.links {
		fmt.Printf("%s  Link %d: s=\"%s\"\n",
			strings.Repeat("    ", level), i, l.str)
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

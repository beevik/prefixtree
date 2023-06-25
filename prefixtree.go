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

// A KeyValue type encapsulates a key string and its associated value.
type KeyValue struct {
	Key   string
	Value any
}

// A Tree represents a prefix tree containing strings and their associated
// value data. The tree is implemented as a trie and can be searched
// efficiently for unique prefix matches.
type Tree struct {
	key         string
	value       any
	links       []link
	descendants int
}

type link struct {
	keyseg string
	tree   *Tree
}

// New returns an empty prefix tree.
func New() *Tree {
	return new(Tree)
}

// isTerminal returns true if the tree is a terminal subtree in the
// prefix tree.
func (t *Tree) isTerminal() bool {
	return t.key != ""
}

// Find searches the prefix tree for all key strings prefixed by the
// provided prefix and returns them.
//
// Deprecated: Use FindValue instead.
func (t *Tree) Find(prefix string) (value any, err error) {
	return t.FindValue(prefix)
}

// FindAll searches the prefix tree for all key strings prefixed by the
// provided prefix. All associated values are returned.
//
// Deprecated: Use FindValues instead.
func (t *Tree) FindAll(prefix string) (values []any) {
	return t.FindValues(prefix)
}

// FindKey searches the prefix tree for a key string that uniquely matches the
// prefix. If found, the full matching key is returned. If not found,
// ErrPrefixNotFound is returned. If the prefix matches more than one key in
// the tree, ErrPrefixAmbiguous is returned.
func (t *Tree) FindKey(prefix string) (key string, err error) {
	st, err := t.findSubtree(prefix)
	if err != nil {
		return "", err
	}
	return st.key, nil
}

// FindKeyValue searches the prefix tree for a key string that uniquely
// matches the prefix. If found, the full matching key and its associated
// value is returned. If not found, ErrPrefixNotFound is returned. If the
// prefix matches more than one key in the tree, ErrPrefixAmbiguous is
// returned.
func (t *Tree) FindKeyValue(prefix string) (kv KeyValue, err error) {
	st, err := t.findSubtree(prefix)
	if err != nil {
		return KeyValue{}, err
	}
	return KeyValue{st.key, st.value}, nil
}

// FindKeys searches the prefix tree for all key strings prefixed by the
// provided prefix and returns them.
func (t *Tree) FindKeys(prefix string) (keys []string) {
	st, err := t.findSubtree(prefix)
	if err == ErrPrefixNotFound {
		return []string{}
	}
	return appendDescendantKeys(st, nil)
}

// FindValue searches the prefix tree for a key string that uniquely matches
// the prefix. If found, the value associated with the key is returned. If not
// found, ErrPrefixNotFound is returned. If the prefix matches more than one
// key in the tree, ErrPrefixAmbiguous is returned.
func (t *Tree) FindValue(prefix string) (value any, err error) {
	st, err := t.findSubtree(prefix)
	if err != nil {
		return nil, err
	}
	return st.value, nil
}

// FindKeyValues searches the prefix tree for all key strings prefixed by the
// provided prefix. All discovered keys and their values are returned.
func (t *Tree) FindKeyValues(prefix string) (values []KeyValue) {
	st, err := t.findSubtree(prefix)
	if err == ErrPrefixNotFound {
		return []KeyValue{}
	}
	return appendDescendantKeyValues(st, nil)
}

// FindValues searches the prefix tree for all key strings prefixed by the
// provided prefix. All associated values are returned.
func (t *Tree) FindValues(prefix string) (values []any) {
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
		// Ran out of prefix?
		if len(prefix) == 0 {
			if t.isTerminal() {
				return t, nil
			}
			return t, ErrPrefixAmbiguous
		}

		// Figure out which links to consider. If the number of links from the
		// node is large-ish (20+), do a binary search for 2 candidate links.
		// Otherwise search all links. The cutoff point between binary and
		// linear search was determined by benchmarking against the unix
		// english dictionary.
		start, stop := 0, len(t.links)-1
		if len(t.links) >= 20 {
			ix := sort.Search(len(t.links),
				func(i int) bool { return t.links[i].keyseg >= prefix })
			start, stop = maxInt(0, ix-1), minInt(ix, stop)
		}

		// Perform the check on all candidate links.
	innerLoop:
		for i := start; i <= stop; i++ {
			link := &t.links[i]
			m := matchingChars(prefix, link.keyseg)
			switch {
			case m == 0:
				continue innerLoop
			case m == len(link.keyseg):
				// Full link 1, so proceed down subtree.
				t, prefix = link.tree, prefix[m:]
				continue outerLoop
			case m == len(prefix):
				// Remaining prefix fully consumed, so return value data
				// unless it's non-terminal.
				switch {
				case link.tree.descendants > 1:
					return link.tree, ErrPrefixAmbiguous
				case link.tree.isTerminal():
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

// appendDescendantKeys recursively appends a tree's descendant keys
// to an array of keys.
func appendDescendantKeys(t *Tree, keys []string) []string {
	if t.isTerminal() {
		keys = append(keys, t.key)
	}
	for i := 0; i < len(t.links); i++ {
		keys = appendDescendantKeys(t.links[i].tree, keys)
	}
	return keys
}

// appendDescendantKeyValues recursively appends a tree's descendant keys
// to an array of key/value pairs.
func appendDescendantKeyValues(t *Tree, kv []KeyValue) []KeyValue {
	if t.isTerminal() {
		kv = append(kv, KeyValue{t.key, t.value})
	}
	for i := 0; i < len(t.links); i++ {
		kv = appendDescendantKeyValues(t.links[i].tree, kv)
	}
	return kv
}

// appendDescendantValues recursively appends a tree's descendant values
// to an array of values.
func appendDescendantValues(t *Tree, values []any) []any {
	if t.isTerminal() {
		values = append(values, t.value)
	}
	for i := 0; i < len(t.links); i++ {
		values = appendDescendantValues(t.links[i].tree, values)
	}
	return values
}

// Add a key string and its associated value data to the prefix tree.
func (t *Tree) Add(key string, value any) {
	k := key
outerLoop:
	for {
		t.descendants++

		// If we've consumed the entire string, then the tree node is terminal
		// and we're done.
		if len(k) == 0 {
			t.key, t.value = key, value
			break outerLoop
		}

		// Find the lexicographical link insertion point.
		ix := sort.Search(len(t.links),
			func(i int) bool { return t.links[i].keyseg >= k })

		// Check the links before and after the insertion point for a matching
		// prefix to see if we need to split one of them.
		var splitLink *link
		var splitIndex int
	innerLoop:
		for li, lm := maxInt(ix-1, 0), minInt(ix, len(t.links)-1); li <= lm; li++ {
			link := &t.links[li]
			m := matchingChars(link.keyseg, k)
			switch {
			case m == len(link.keyseg):
				// Full link match, so proceed down the subtree.
				t, k = link.tree, k[m:]
				continue outerLoop
			case m > 0:
				// Partial match, so we'll need to split this tree node.
				splitLink, splitIndex = link, m
				break innerLoop
			}
		}

		// No split necessary, so insert a new link and subtree.
		if splitLink == nil {
			child := &Tree{
				key:         key,
				value:       value,
				links:       nil,
				descendants: 1,
			}
			t.links = append(t.links[:ix],
				append([]link{{k, child}}, t.links[ix:]...)...)
			break outerLoop
		}

		// A split is necessary, so split the current link's string and insert
		// a child tree.
		k1, k2 := splitLink.keyseg[:splitIndex], splitLink.keyseg[splitIndex:]
		child := &Tree{
			key:         "",
			value:       nil,
			links:       []link{{k2, splitLink.tree}},
			descendants: splitLink.tree.descendants,
		}
		splitLink.keyseg, splitLink.tree = k1, child
		t, k = child, k[splitIndex:]
	}
}

// Output the structure of the tree to stdout. This function exists for
// debugging purposes.
func (t *Tree) Output() {
	t.outputNode(0)
}

func (t *Tree) outputNode(level int) {
	fmt.Printf("%sNode: key=\"%s\" term=%v desc=%d value=%v\n",
		strings.Repeat("    ", level), t.key, t.isTerminal(), t.descendants, t.value)
	for i, l := range t.links {
		fmt.Printf("%s  Link %d: ks=\"%s\"\n",
			strings.Repeat("    ", level), i, l.keyseg)
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

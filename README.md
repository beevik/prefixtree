[![GoDoc](https://godoc.org/github.com/beevik/prefixtree?status.svg)](https://godoc.org/github.com/beevik/prefixtree)
[![Go](https://github.com/beevik/prefixtree/actions/workflows/go.yml/badge.svg)](https://github.com/beevik/prefixtree/actions/workflows/go.yml)

prefixtree
==========

The prefixtree package implements a simple prefix trie data structure.
The tree enables rapid searching for strings that uniquely match a given
prefix. The implementation allows the user to associate data with each
string, so it can act as a sort of flexible key-value store where
searches succeed with the shortest unambiguous key prefix.

### Example: Building a prefix tree

The following code adds strings and associated data (in this case an integer)
to a prefix tree.

```go
tree := prefixtree.New()

tree.Add("apple", 10)
tree.Add("orange", 20)
tree.Add("apple pie", 30)
tree.Add("lemon", 40)
tree.Add("lemon meringue pie", 50)
```

### Example: Searching the prefix tree

The following code searches the prefix tree generated by the
previous example and outputs the results.

```go
fmt.Printf("%-18s %-8s %s\n", "prefix", "value", "error")
fmt.Printf("%-18s %-8s %s\n", "------", "-----", "-----")

for _, prefix := range []string{
    "a",
    "appl",
    "apple",
    "apple p",
    "apple pie",
    "apple pies",
    "o",
    "oran",
    "orange",
    "oranges",
    "l",
    "lemo",
    "lemon",
    "lemon m",
    "lemon meringue",
    "pear",
} {
    value, err := tree.Find(prefix)
    fmt.Printf("%-18s %-8v %v\n", prefix, value, err)
}
```

Output:
```
prefix             value    error
------             -----    -----
a                  <nil>    prefixtree: prefix ambiguous
appl               <nil>    prefixtree: prefix ambiguous
apple              10       <nil>
apple p            30       <nil>
apple pie          30       <nil>
apple pies         <nil>    prefixtree: prefix not found
o                  20       <nil>
orang              20       <nil>
orange             20       <nil>
oranges            <nil>    prefixtree: prefix not found
l                  <nil>    prefixtree: prefix ambiguous
lemo               <nil>    prefixtree: prefix ambiguous
lemon              40       <nil>
lemon m            50       <nil>
lemon meringue     50       <nil>
pear               <nil>    prefixtree: prefix not found
```

package encoder

import "sort"

const (
	notUsed = ""
)

type Strings struct {
	valid bool
	tbl   map[string]struct{}
}

type Table struct {
	valid   bool
	tbl     map[string]int32
	strings []string
}

func NewStrings() *Strings {
	s := &Strings{
		valid: true,
		tbl:   make(map[string]struct{}),
	}

	return s
}

func (s *Strings) Add(value string) {
	if !s.valid {
		panic("Strings in an invalid state")
	}

	s.tbl[value] = struct{}{}
}

func (s *Strings) CalcTable() *Table {
	if !s.valid {
		panic("Strings in an invalid state")
	}

	strings := make([]string, 0, len(s.tbl)+1)

	// Index 0 is used by pb.DenseNodes to encode tags.  We insert an empty
	// string that will be at index 0 after the array has been sorted.
	strings = append(strings, notUsed)

	for k := range s.tbl {
		strings = append(strings, k)
	}

	sort.Strings(strings)

	tbl := make(map[string]int32, len(strings))
	for i, k := range strings {
		tbl[k] = int32(i)
	}

	return &Table{
		valid:   true,
		tbl:     tbl,
		strings: strings,
	}
}

func (t *Table) IndexOf(value string) int32 {
	if !t.valid {
		panic("Table is in an invalid state")
	}

	if index, ok := t.tbl[value]; !ok {
		panic("Index does not exist")
	} else {
		return index
	}
}

func (t *Table) AsArray() []string {
	if !t.valid {
		panic("Table is in an invalid state")
	}

	return t.strings
}

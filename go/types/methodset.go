// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements method sets.

package types

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

// A MethodSet is an ordered set of concrete or abstract (interface) methods;
// a method is a MethodVal selection.
// The zero value for a MethodSet is a ready-to-use empty method set.
type MethodSet struct {
	list []*Selection
}

func (s *MethodSet) String() string {
	if s.Len() == 0 {
		return "MethodSet {}"
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "MethodSet {")
	for _, f := range s.list {
		fmt.Fprintf(&buf, "\t%s\n", f)
	}
	fmt.Fprintln(&buf, "}")
	return buf.String()
}

// Len returns the number of methods in s.
func (s *MethodSet) Len() int { return len(s.list) }

// At returns the i'th method in s for 0 <= i < s.Len().
func (s *MethodSet) At(i int) *Selection { return s.list[i] }

// Lookup returns the method with matching package and name, or nil if not found.
func (s *MethodSet) Lookup(pkg *Package, name string) *Selection {
	if s.Len() == 0 {
		return nil
	}

	key := Id(pkg, name)
	i := sort.Search(len(s.list), func(i int) bool {
		m := s.list[i]
		return m.Obj.Id() >= key
	})
	if i < len(s.list) {
		m := s.list[i]
		if m.Obj.Id() == key {
			return m
		}
	}
	return nil
}

// Shared empty method set.
var emptyMethodSet MethodSet

// A cachedMethodSet provides access to a method set
// for a given type by computing it once on demand,
// and then caching it for future use. Threadsafe.
type cachedMethodSet struct {
	mset *MethodSet
	mu   sync.RWMutex // protects mset
}

// Of returns the (possibly cached) method set for typ.
// Threadsafe.
func (c *cachedMethodSet) of(typ Type) *MethodSet {
	c.mu.RLock()
	mset := c.mset
	c.mu.RUnlock()
	if mset == nil {
		mset = NewMethodSet(typ)
		c.mu.Lock()
		c.mset = mset
		c.mu.Unlock()
	}
	return mset
}

// NewMethodSet returns the method set for the given type T.  It
// always returns a non-nil method set, even if it is empty.
//
// A MethodSetCache handles repeat queries more efficiently.
func NewMethodSet(T Type) *MethodSet {
	typ, isPtr := deref(T)
	named, _ := typ.(*Named)

	// If typ is an interface then *typ has no methods.
	if isPtr {
		utyp := typ
		if named != nil {
			utyp = named.UnderlyingT
		}
		if _, ok := utyp.(*Interface); ok {
			return &emptyMethodSet
		}
	}

	sel := selections(T)

	if len(sel) == 0 {
		return &emptyMethodSet
	}

	// collect methods
	var list []*Selection
	for _, s := range sel {
		if s != nil && s.Kind == MethodVal {
			s.Recv = T
			list = append(list, s)
		}
	}
	sort.Sort(byUniqueName(list))
	return &MethodSet{list}
}

// A FieldSet is an ordered set of struct fields;
// a field is a FieldVal selection.
// The zero value for a FieldSet is a ready-to-use empty field set.
type FieldSet struct {
	list []*Selection
}

func (s *FieldSet) String() string {
	if s.Len() == 0 {
		return "FieldSet {}"
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "FieldSet {")
	for _, f := range s.list {
		fmt.Fprintf(&buf, "\t%s\n", f)
	}
	fmt.Fprintln(&buf, "}")
	return buf.String()
}

// Len returns the number of fields in s.
func (s *FieldSet) Len() int { return len(s.list) }

// At returns the i'th field in s for 0 <= i < s.Len().
func (s *FieldSet) At(i int) *Selection { return s.list[i] }

// Lookup returns the field with matching package and name, or nil if not found.
func (s *FieldSet) Lookup(pkg *Package, name string) *Selection {
	if s.Len() == 0 {
		return nil
	}

	key := Id(pkg, name)
	i := sort.Search(len(s.list), func(i int) bool {
		m := s.list[i]
		return m.Obj.Id() >= key
	})
	if i < len(s.list) {
		m := s.list[i]
		if m.Obj.Id() == key {
			return m
		}
	}
	return nil
}

// Shared empty field set.
var emptyFieldSet FieldSet

// NewFieldSet returns the field set for the given type T.  It
// always returns a non-nil field set, even if it is empty.
func NewFieldSet(T Type) *FieldSet {
	sel := selections(T)

	// If T is a named type with underlying type *V then T has all the fields of V but none of its methods.
	// Furthermore, such a T can have no fields or methods of its own, so no need to worry about shadowing.
	if t, _ := T.(*Named); t != nil {
		u := t.UnderlyingT
		if _, ok := u.(*Pointer); ok {
			sel = selections(u)
		}
	}

	if len(sel) == 0 {
		return &emptyFieldSet
	}

	// collect fields
	var list []*Selection
	for _, s := range sel {
		if s != nil && s.Kind == FieldVal {
			s.Recv = T
			list = append(list, s)
		}
	}
	sort.Sort(byUniqueName(list))
	return &FieldSet{list}
}

func selections(T Type) selectionSet {
	// WARNING: The code in this function is extremely subtle - do not modify casually!
	//          This function and lookupFieldOrMethod should be kept in sync.

	// selections up to the current depth, allocated lazily
	var base selectionSet

	typ, isPtr := deref(T)
	named, _ := typ.(*Named)

	// Start with typ as single entry at shallowest depth.
	// If typ is not a named type, insert a nil type instead.
	current := []embeddedType{{named, nil, isPtr, false}}

	// named types that we have seen already, allocated lazily
	var seen map[*Named]bool

	// collect selections at current depth
	for len(current) > 0 {
		var next []embeddedType // embedded types found at current depth

		// selections at current depth, allocated lazily
		var set selectionSet

		for _, e := range current {
			// The very first time only, e.typ may be nil.
			// In this case, we don't have a named type and
			// we simply continue with the underlying type.
			if e.typ != nil {
				if seen[e.typ] {
					// We have seen this type before, at a more shallow depth
					// (note that multiples of this type at the current depth
					// were consolidated before). The type at that depth shadows
					// this same type at the current depth, so we can ignore
					// this one.
					continue
				}
				if seen == nil {
					seen = make(map[*Named]bool)
				}
				seen[e.typ] = true

				set = set.addMethods(e.typ.Methods, e.index, e.indirect, e.multiples)

				// continue with underlying type
				typ = e.typ.UnderlyingT
			}

			switch t := typ.(type) {
			case *Struct:
				for i, f := range t.Fields {
					index := concat(e.index, i)
					set = set.addField(f, index, e.indirect, e.multiples)

					// Embedded fields are always of the form T or *T where
					// T is a named type. If typ appeared multiple times at
					// this depth, f.Type appears multiple times at the next
					// depth.
					if f.Anonymous {
						// Ignore embedded basic types - only user-defined
						// named types can have methods or struct fields.
						typ, isPtr := deref(f.Type)
						if t, _ := typ.(*Named); t != nil {
							next = append(next, embeddedType{t, index, e.indirect || isPtr, e.multiples})
						}
					}
				}

			case *Interface:
				set = set.addMethods(t.allMethods, e.index, true, e.multiples)
			}
		}

		// Add selections and collisions at this depth to base if no entries with matching
		// names exist already.
		for id, s := range set {
			if _, found := base[id]; !found {
				if base == nil {
					base = make(selectionSet)
				}
				base[id] = s
			}
		}

		current = consolidateMultiples(next)
	}

	return base
}

// A selectionSet is a set of selections and name collisions.
// A collision indicates that multiple selections with the
// same unique id appeared.
type selectionSet map[string]*Selection // a nil entry indicates a name collision

// AddField adds field v to the set s.
// If multiples is set, v appears multiple times
// and is treated as a collision.
func (s selectionSet) addField(v *Var, index []int, indirect bool, multiples bool) selectionSet {
	if s == nil {
		s = make(selectionSet)
	}
	id := v.Id()
	// if v is not in the set, add it
	if !multiples {
		if _, found := s[id]; !found {
			s[id] = &Selection{FieldVal, nil, v, index, indirect}
			return s
		}
	}
	s[id] = nil // collision
	return s
}

// AddMethods adds all functions in list to the method set s.
// If multiples is set, every function in list appears multiple times
// and is treated as a collision.
func (s selectionSet) addMethods(list []*Func, index []int, indirect bool, multiples bool) selectionSet {
	if len(list) == 0 {
		return s
	}
	if s == nil {
		s = make(selectionSet)
	}
	for i, f := range list {
		id := f.Id()
		// if f is not in the set, add it
		if !multiples {
			if _, found := s[id]; !found && (indirect || !ptrRecv(f)) {
				s[id] = &Selection{MethodVal, nil, f, concat(index, i), indirect}
				continue
			}
		}
		s[id] = nil // collision
	}
	return s
}

// ptrRecv reports whether the receiver is of the form *T.
// The receiver must exist.
func ptrRecv(f *Func) bool {
	_, isPtr := deref(f.Type.(*Signature).Recv.Type)
	return isPtr
}

// byUniqueName function lists can be sorted by their unique names.
type byUniqueName []*Selection

func (a byUniqueName) Len() int           { return len(a) }
func (a byUniqueName) Less(i, j int) bool { return a[i].Obj.Id() < a[j].Obj.Id() }
func (a byUniqueName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

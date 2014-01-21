// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements Selections.

package types

import (
	"bytes"
	"fmt"
)

// SelectionKind describes the kind of a selector expression x.f.
type SelectionKind int

const (
	FieldVal   SelectionKind = iota // x.f is a struct field selector
	MethodVal                       // x.f is a method selector
	MethodExpr                      // x.f is a method expression
	PackageObj                      // x.f is a qualified identifier
)

// A Selection describes a selector expression x.f.
// For the declarations:
//
//	type T struct{ x int; E }
//	type E struct{}
//	func (e E) m() {}
//	var p *T
//
// the following relations exist:
//
//	Selector    Kind          Recv    Obj    Type               Index     Indirect
//
//	p.x         FieldVal      T       x      int                {0}       true
//	p.m         MethodVal     *T      m      func (e *T) m()    {1, 0}    true
//	T.m         MethodExpr    T       m      func m(_ T)        {1, 0}    false
//	math.Pi     PackageObj    nil     Pi     untyped numeric    nil       false
//
type Selection struct {
	Kind     SelectionKind
	Recv     Type   // type of x, nil if kind == PackageObj
	Obj      Object // object denoted by x.f
	Index    []int  // path from x to x.f, nil if kind == PackageObj
	Indirect bool   // set if there was any pointer indirection on the path, false if kind == PackageObj
}

// Recv returns the type of x in x.f.
// The result is nil if x.f is a qualified identifier (PackageObj).

// Obj returns the object denoted by x.f.
// The following object types may appear:
//
//	Kind          Object
//
//	FieldVal      *Var                          field
//	MethodVal     *Func                         method
//	MethodExpr    *Func                         method
//	PackageObj    *Const, *Type, *Var, *Func    imported const, type, var, or func
//

// Type returns the type of x.f, which may be different from the type of f.
// See Selection for more information.
func (s *Selection) Type() Type {
	switch s.Kind {
	case MethodVal:
		// The type of x.f is a method with its receiver type set
		// to the type of x.
		sig := *s.Obj.(*Func).Type.(*Signature)
		recv := *sig.Recv
		recv.Type = s.Recv
		sig.Recv = &recv
		return &sig

	case MethodExpr:
		// The type of x.f is a function (without receiver)
		// and an additional first argument with the same type as x.
		// TODO(gri) Similar code is already in call.go - factor!
		sig := *s.Obj.(*Func).Type.(*Signature)
		arg0 := *sig.Recv
		arg0.Type = s.Recv
		var params []*Var
		if sig.Params != nil {
			params = sig.Params
		}
		sig.Params = append([]*Var{&arg0}, params...)
		return &sig
	}

	// In all other cases, the type of x.f is the type of x.
	return s.Obj.GetType()
}

// Index describes the path from x to f in x.f.
// The result is nil if x.f is a qualified identifier (PackageObj).
//
// The last index entry is the field or method index of the type declaring f;
// either:
//
//	1) the list of declared methods of a named type; or
//	2) the list of methods of an interface type; or
//	3) the list of fields of a struct type.
//
// The earlier index entries are the indices of the embedded fields implicitly
// traversed to get from (the type of) x to f, starting at embedding depth 0.

// Indirect reports whether any pointer indirection was required to get from
// x to f in x.f.
// The result is false if x.f is a qualified identifier (PackageObj).

func (s *Selection) String() string { return SelectionString(nil, s) }

// SelectionString returns the string form of s.
// Type names are printed package-qualified
// only if they do not belong to this package.
//
func SelectionString(this *Package, s *Selection) string {
	var k string
	switch s.Kind {
	case FieldVal:
		k = "field "
	case MethodVal:
		k = "method "
	case MethodExpr:
		k = "method expr "
	case PackageObj:
		return fmt.Sprintf("qualified ident %s", s.Obj)
	default:
		unreachable()
	}
	var buf bytes.Buffer
	buf.WriteString(k)
	buf.WriteByte('(')
	WriteType(&buf, this, s.Recv)
	fmt.Fprintf(&buf, ") %s", s.Obj.GetName())
	writeSignature(&buf, this, s.Type().(*Signature))
	return buf.String()
}

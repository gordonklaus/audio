// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "fmt"

// A Package describes a Go package.
type Package struct {
	Path     string
	Name     string
	scope    *Scope
	complete bool
	Imports  []*Package
	fake     bool // scope lookup errors are silently dropped if package is fake (internal use only)
}

// NewPackage returns a new Package for the given package path,
// name, and scope. The package is not complete and contains no
// explicit imports.
func NewPackage(path, name string, scope *Scope) *Package {
	return &Package{Path: path, Name: name, scope: scope}
}

// Scope returns the (complete or incomplete) package scope
// holding the objects declared at package level (TypeNames,
// Consts, Vars, and Funcs).
func (pkg *Package) Scope() *Scope { return pkg.scope }

// A package is complete if its scope contains (at least) all
// exported objects; otherwise it is incomplete.
func (pkg *Package) Complete() bool { return pkg.complete }

// MarkComplete marks a package as complete.
func (pkg *Package) MarkComplete() { pkg.complete = true }

func (pkg *Package) String() string {
	return fmt.Sprintf("package %s (%s)", pkg.Name, pkg.Path)
}

// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type browser struct {
	*ViewBase
	options    browserOptions
	typ        types.Type // non-nil if this is a selection browser
	fun        *funcNode
	currentPkg *types.Package
	imports    []*types.Package
	finished   bool
	accepted   func(types.Object)
	canceled   func()

	path, objs objects
	i          int
	newObj     types.Object

	pathTexts, objTexts []*Text
	text                *Text
	typeView            *typeView
	pkgName             *Text
	funcAsVal           bool
}

type browserOptions struct {
	objFilter                             func(types.Object) bool
	acceptTypes, enterTypes, canFuncAsVal bool
}

func newBrowser(options browserOptions, parent View) *browser {
	b := &browser{options: options}
	b.ViewBase = NewView(b)

	// not a very beautiful way to get context but the most comprehensible I could find
loop:
	for v := parent; v != nil; v = Parent(v) {
		switch v := v.(type) {
		case *block: // must use blk.func_() instead of going directly to *funcNode so we don't grab a func literal node
			b.fun = v.func_()
			break loop
		case *fluxWindow:
			t := v.browser.currentObj().(*types.TypeName)
			b.currentPkg = t.Pkg
			b.imports = imports(t)
			break loop
		}
	}
	if b.fun != nil {
		b.currentPkg = b.fun.pkg()
		b.imports = b.fun.imports()
	}

	b.text = NewText("")
	b.text.SetBackgroundColor(noColor)
	b.text.Validate = b.validateText
	b.text.TextChanged = b.textChanged
	b.Add(b.text)

	b.typeView = newTypeView(new(types.Type))
	b.typeView.pkg = b.currentPkg
	b.Add(b.typeView)

	b.pkgName = NewText("")
	b.pkgName.SetBackgroundColor(Color{0, 0, 0, .7})
	b.Add(b.pkgName)

	b.clearText()

	return b
}

func newSelectionBrowser(t types.Type, parent View) *browser {
	b := newBrowser(browserOptions{canFuncAsVal: true}, parent)
	b.typ = t
	b.clearText()
	return b
}

func imports(t *types.TypeName) (x []*types.Package) {
	seen := map[*types.Package]bool{}
	walkType(t.Type.(*types.Named).UnderlyingT, func(n *types.Named) {
		if p := n.Obj.Pkg; p != nil && p != t.Pkg && !seen[p] {
			seen[p] = true
			x = append(x, p)
		}
	})
	return
}

func (b *browser) cancel() {
	if !b.finished {
		b.finished = true
		b.canceled()
	}
}

func (b *browser) clearText() { b.text.SetText("") }

func (b browser) currentObj() types.Object {
	if len(b.objs) > 0 {
		return b.objs[b.i]
	}
	return nil
}

func (b *browser) makeCurrent(obj types.Object) {
	for i, o := range b.objs {
		if o == obj {
			b.i = i
			b.refresh()
			break
		}
	}
}

func (b browser) lastPathText() (*Text, bool) {
	if np := len(b.pathTexts); np > 0 {
		return b.pathTexts[np-1], true
	}
	return nil, false
}

func (b *browser) validateText(text *string) bool {
	if b.newObj != nil {
		// TODO: unique id
		return true
	}
	for _, obj := range b.filteredObjs() {
		name := obj.GetName()
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(*text)) {
			*text = name[:len(*text)]
			return true
		}
	}
	return false
}

func (b *browser) textChanged(text string) {
	objs := objects{}
	i := 0

	for _, obj := range b.filteredObjs() {
		if strings.HasPrefix(strings.ToLower(obj.GetName()), strings.ToLower(text)) {
			objs = append(objs, obj)
			if obj == b.currentObj() {
				i = len(objs) - 1
			}
		}
	}

	if b.newObj != nil {
		switch obj := b.newObj.(type) {
		case *pkgObject:
			obj.name = text
		case *types.TypeName:
			obj.Name = text
		case *types.Func:
			obj.Name = text
		case *types.Var:
			obj.Name = text
		case *types.Const:
			obj.Name = text
		case *localVar:
			obj.Name = text
		}
		i = 0
		for i < len(objs) && objLess(objs[i], b.newObj) {
			i++
		}
		objs = append(objs[:i], append(objects{b.newObj}, objs[i:]...)...)
	}

	b.objs = objs
	b.i = i
	b.refresh()
}

func (b *browser) refresh() {
	cur := b.currentObj()

	if cur != nil {
		f := b.text.TextChanged
		b.text.TextChanged = nil
		b.text.SetText(cur.GetName()[:len(b.text.Text())])
		b.text.TextChanged = f
	}

	xOffset := 0.0
	if t, ok := b.lastPathText(); ok {
		if cur != nil {
			sep := "."
			if _, ok := cur.(*pkgObject); ok {
				sep = "/"
			}
			text := t.Text()
			t.SetText(text[:len(text)-1] + sep)
		}
		xOffset = Pos(t).X + Width(t)
	}

	n := len(b.objs)
	for _, l := range b.objTexts {
		l.Close()
	}
	b.objTexts = nil
	width := 0.0
	for i, obj := range b.objs {
		l := NewText(obj.GetName())
		l.SetTextColor(color(obj, false, b.funcAsVal))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		b.Add(l)
		b.objTexts = append(b.objTexts, l)
		l.Move(Pt(xOffset, float64(n-i-1)*Height(l)))
		if Width(l) > width {
			width = Width(l)
		}
	}
	Raise(b.text)
	Resize(b, Pt(xOffset+width, float64(n)*Height(b.text)))

	yOffset := float64(n-b.i-1) * Height(b.text)
	b.text.Move(Pt(xOffset, yOffset))
	Hide(b.pkgName)
	if pkg, ok := cur.(*pkgObject); ok {
		t := b.pkgName
		t.SetText(pkg.name)
		t.Move(Pt(xOffset+width+16, yOffset-(Height(t)-Height(b.text))/2))
		if pkg.name != pkg.GetName() {
			Show(t)
		}
	}
	Hide(b.typeView)
	if cur != nil {
		b.text.SetTextColor(color(cur, true, b.funcAsVal))
		switch cur := cur.(type) {
		case *types.TypeName:
			if t, ok := cur.Type.(*types.Named); ok && t.UnderlyingT != nil {
				b.typeView.setType(t.UnderlyingT)
				Show(b.typeView)
			}
		case *types.Func, *types.Var, *types.Const, field, *localVar:
			if !isOperator(cur) {
				b.typeView.setType(cur.GetType())
				Show(b.typeView)
			}
		}
		b.typeView.Move(Pt(xOffset+width+16, yOffset-(Height(b.typeView)-Height(b.text))/2))
	}
	for _, p := range b.pathTexts {
		p.Move(Pt(Pos(p).X, yOffset))
	}

	Pan(b, Pt(0, yOffset))
}

var pkgObjects = map[string]*pkgObject{}

func (b browser) filteredObjs() (objs objects) {
	add := func(obj types.Object) {
		if invisible(obj, b.currentPkg) {
			return
		}
		if _, ok := obj.(*pkgObject); ok || b.options.objFilter == nil || b.options.objFilter(obj) {
			objs = append(objs, obj)
		}
	}

	addSubPkgs := func(importPath string) {
		seen := map[string]bool{}
		for _, srcDir := range build.Default.SrcDirs() {
			file, err := os.Open(filepath.Join(srcDir, importPath))
			if err != nil {
				continue
			}
			defer file.Close()
			fileInfos, err := file.Readdir(-1)
			if err != nil {
				continue
			}
			for _, fileInfo := range fileInfos {
				name := fileInfo.Name()
				if !fileInfo.IsDir() || !unicode.IsLetter([]rune(name)[0]) || name == "testdata" || seen[name] {
					continue
				}
				seen[name] = true

				importPath := path.Join(importPath, name)
				pkgObj, ok := pkgObjects[importPath]
				if !ok {
					if pkg, err := build.Import(importPath, "", build.AllowBinary); err == nil {
						name = pkg.Name
					}
					pkgObj = &pkgObject{nil, importPath, name, srcDir}
					pkgObjects[importPath] = pkgObj
				}
				add(pkgObj)
			}
		}
	}

	if b.typ != nil {
		mset := types.NewMethodSet(b.typ)
		for i := 0; i < mset.Len(); i++ {
			m := mset.At(i)
			// m.Type() has the correct receiver for inherited methods (m.Obj does not)
			add(types.NewFunc(0, m.Obj.GetPkg(), m.Obj.GetName(), m.Type().(*types.Signature)))
		}
		fset := types.NewFieldSet(b.typ)
		for i := 0; i < fset.Len(); i++ {
			f := fset.At(i)
			add(field{f.Obj.(*types.Var), f.Recv, f.Indirect})
		}
	} else if len(b.path) > 0 {
		switch obj := b.path[0].(type) {
		case *pkgObject:
			if pkg, err := getPackage(obj.importPath); err == nil {
				for _, obj := range pkg.Scope().Objects {
					add(obj)
				}
			} else {
				if _, ok := err.(*build.NoGoError); !ok {
					fmt.Println(err)
				}
				pkgs[obj.importPath] = types.NewPackage(obj.importPath, obj.name, types.NewScope(types.Universe))
			}
			addSubPkgs(obj.importPath)
		case *types.TypeName:
			for _, m := range intuitiveMethodSet(obj.Type) {
				if types.IsIdentical(m.Obj.(*types.Func).Type.(*types.Signature).Recv.Type, m.Recv) {
					// preserve Object identity for non-inherited methods so that fluxObjs works
					add(m.Obj)
				} else {
					// m.Type() has the correct receiver for inherited methods (m.Obj does not)
					add(types.NewFunc(0, m.Obj.GetPkg(), m.Obj.GetName(), m.Type().(*types.Signature)))
				}
			}
		}
	} else {
		for _, name := range []string{"break", "call", "continue", "convert", "defer", "func", "go", "if", "loop", "select", "typeAssert"} {
			add(special{newVar(name, nil)})
		}
		for _, name := range []string{"=", "*"} {
			add(newVar(name, nil))
		}
		pkgs := b.imports
		if b.currentPkg != nil {
			pkgs = append(pkgs, b.currentPkg)
		}
		for _, p := range pkgs {
			for _, obj := range p.Scope().Objects {
				add(obj)
			}
		}
		for _, obj := range types.Universe.Objects {
			switch obj.GetName() {
			case "nil", "print", "println":
				continue
			}
			add(obj)
		}
		for _, op := range []string{"!", "&&", "||", "+", "-", "*", "/", "%", "&", "|", "^", "&^", "<<", ">>", "==", "!=", "<", "<=", ">", ">=", "[]", "[:]", "<-"} {
			add(types.NewFunc(0, nil, op, nil))
		}
		for _, t := range []*types.TypeName{protoPointer, protoArray, protoSlice, protoMap, protoChan, protoFunc, protoInterface, protoStruct} {
			add(t)
		}
		addSubPkgs("")
		if b.fun != nil {
			b.fun.funcblk.walk(func(blk *block) {
				for v := range blk.localVars {
					add(v)
				}
			}, nil, nil)
		}
	}

	sort.Sort(objs)
	return
}

func isFluxObj(obj types.Object) bool {
	return fluxObjs[obj]
}

func isType(obj types.Object) bool {
	_, ok := obj.(*types.TypeName)
	return ok
}

func isComparableType(obj types.Object) bool {
	switch obj {
	case protoPointer, protoArray, protoChan, protoInterface, protoStruct:
		return true
	}
	t, ok := obj.(*types.TypeName)
	return ok && t.GetType() != nil && types.Comparable(t.GetType())
}

func isCompositeOrPtrType(obj types.Object) bool {
	return isCompositeType(obj) || obj == protoPointer
}

func isCompositeType(obj types.Object) bool {
	switch obj {
	case protoArray, protoSlice, protoMap, protoStruct:
		return true
	}
	if obj, ok := obj.(*types.TypeName); ok {
		switch underlying(obj.GetType()).(type) {
		case *types.Array, *types.Slice, *types.Map, *types.Struct:
			return true
		}
	}
	return false
}

func isMakeableType(obj types.Object) bool {
	switch obj {
	case protoSlice, protoMap, protoChan:
		return true
	}
	if obj, ok := obj.(*types.TypeName); ok {
		switch underlying(obj.GetType()).(type) {
		case *types.Slice, *types.Map, *types.Chan:
			return true
		}
	}
	return false
}

func isGoDeferrable(obj types.Object) bool {
	switch obj := obj.(type) {
	case special:
		return obj.Name == "call"
	case *types.Builtin:
		switch obj.Name {
		case "close", "copy", "delete", "panic", "recover":
			return true
		}
	case *types.Func:
		return !isOperator(obj)
	case *types.TypeName:
		_, ok := obj.GetType().(*types.Named)
		return ok
	}
	return false
}

func (b *browser) LostKeyFocus() { b.cancel() } // TODO: implement this differently.  it interferes with localVar type editing

func (b *browser) KeyPress(event KeyEvent) {
	if b.options.canFuncAsVal && event.Shift != b.funcAsVal {
		b.funcAsVal = event.Shift
		b.refresh()
	}
	switch event.Key {
	case KeyUp:
		if b.newObj == nil {
			b.i--
			if b.i < 0 {
				b.i += len(b.objs)
			}
			b.refresh()
		}
	case KeyDown:
		if b.newObj == nil {
			b.i++
			if b.i >= len(b.objs) {
				b.i -= len(b.objs)
			}
			b.refresh()
		}
	case KeyBackspace:
		if len(b.text.Text()) > 0 {
			b.text.KeyPress(event)
			break
		}
		fallthrough
	case KeyLeft:
		if len(b.path) > 0 && b.newObj == nil {
			parent := b.path[0]
			b.path = b.path[1:]

			i := len(b.pathTexts) - 1
			b.pathTexts[i].Close()
			b.pathTexts = b.pathTexts[:i]

			b.clearText()
			b.makeCurrent(parent)
		}
	case KeyEnter:
		obj := b.currentObj()
		if obj == nil {
			return
		}
		if pkg, ok := obj.(*pkgObject); ok && event.Shift && b.newObj == nil {
			Show(b.pkgName)
			b.pkgName.Accept = func(name string) {
				if pkg.name != name {
					pkg.name = name
					savePackageName(pkg)
				}
				b.refresh()
				SetKeyFocus(b)
			}
			b.pkgName.Reject = func() {
				b.refresh()
				SetKeyFocus(b)
			}
			SetKeyFocus(b.pkgName)
			return
		}

		if b.newObj != nil {
			switch obj := b.newObj.(type) {
			case *pkgObject:
				// all fields are blank except name which represents the final path element, not the package name.
				// fill in srcDir and importPath based on name and, if present, parent.
				if len(b.path) > 0 {
					parent := b.path[0].(*pkgObject)
					obj.srcDir = parent.srcDir
					obj.importPath = path.Join(parent.importPath, obj.name)
				} else {
					dirs := build.Default.SrcDirs()
					obj.srcDir = dirs[len(dirs)-1]
					obj.importPath = obj.name
				}
				pkgObjects[obj.importPath] = obj
				if err := os.Mkdir(filepath.Join(obj.srcDir, obj.importPath), 0777); err != nil {
					panic(err)
				}
			case *types.TypeName, *types.Func, *types.Var, *types.Const:
				if isMethod(obj) {
					recv := b.path[0].(*types.TypeName).Type.(*types.Named)
					recv.Methods = append(recv.Methods, obj.(*types.Func))
				} else {
					pkg := b.currentPkg
					if len(b.path) > 0 {
						pkg = pkgs[b.path[0].(*pkgObject).importPath]
					}
					pkg.Scope().Insert(obj)
				}
			case *localVar:
				b.finished = true // hack to avoid cancelling browser when it loses keyboard focus to typeView
				b.typeView.edit(func() {
					b.finished = false
					if *b.typeView.typ != nil {
						obj.Type = *b.typeView.typ
						b.finished = true
						b.accepted(obj)
					} else {
						b.clearText()
						SetKeyFocus(b)
					}
				})
				b.newObj = nil
				return
			}

			b.makeCurrent(b.newObj)
			b.newObj = nil
		}

		_, isPkg := obj.(*pkgObject)
		_, isType := obj.(*types.TypeName)
		if !(isPkg || isType && !b.options.acceptTypes) {
			b.finished = true
			b.accepted(obj)
			return
		}
		fallthrough
	case KeyRight:
		if b.newObj == nil {
			switch obj := b.currentObj().(type) {
			case *pkgObject, *types.TypeName:
				if t, ok := obj.(*types.TypeName); ok {
					if _, ok = t.Type.(*types.Basic); ok || t.Type == nil || !b.options.enterTypes {
						break
					}
				}
				b.path = append(objects{obj}, b.path...)

				sep := "."
				if _, ok := obj.(*pkgObject); ok {
					sep = "/"
				}
				t := NewText(obj.GetName() + sep)
				t.SetTextColor(color(obj, true, b.funcAsVal))
				t.SetBackgroundColor(Color{0, 0, 0, .7})
				b.Add(t)
				x := 0.0
				if t, ok := b.lastPathText(); ok {
					x = Pos(t).X + Width(t)
				}
				t.Move(Pt(x, 0))
				b.pathTexts = append(b.pathTexts, t)

				b.clearText()
			}
		}
	case KeyEscape:
		if b.newObj != nil {
			b.newObj = nil
			if b.i < len(b.objs)-1 {
				b.i++
			}
			b.clearText()
		} else {
			b.cancel()
		}
	default:
		if b.fun != nil {
			if event.Command && event.Text == "0" {
				b.newObj = &localVar{refs: map[*valueNode]bool{}}
				b.clearText()
			} else {
				b.text.KeyPress(event)
			}
			return
		}

		makeInPkg := false
		var pkg *types.Package
		var recv *types.TypeName
		if len(b.path) > 0 {
			switch obj := b.path[0].(type) {
			case *pkgObject:
				makeInPkg = true
				pkg = pkgs[obj.importPath]
			case *types.TypeName:
				recv = obj
				pkg = obj.Pkg
			}
		}
		makePkgInRoot := len(b.path) == 0 && event.Text == "1"
		makeMethod := recv != nil && event.Text == "3"
		if b.newObj == nil && event.Command && (makePkgInRoot || makeInPkg || makeMethod) {
			switch event.Text {
			case "1":
				b.newObj = &pkgObject{}
			case "2":
				t := types.NewTypeName(0, pkg, "", nil)
				t.Type = &types.Named{Obj: t}
				b.newObj = t
			case "3":
				sig := &types.Signature{}
				if recv != nil {
					sig.Recv = newVar("", &types.Pointer{Elem: recv.Type})
				}
				b.newObj = types.NewFunc(0, pkg, "", sig)
			case "4":
				b.newObj = types.NewVar(0, pkg, "", nil)
			case "5":
				b.newObj = types.NewConst(0, pkg, "", nil, nil)
			default:
				b.text.KeyPress(event)
				return
			}
			b.clearText()
		} else {
			b.text.KeyPress(event)
		}
	}
}

func (b *browser) KeyRelease(event KeyEvent) {
	if b.options.canFuncAsVal && event.Shift != b.funcAsVal {
		b.funcAsVal = event.Shift
		b.refresh()
	}
}

func (b *browser) Paint() {
	rect := ZR
	if b.newObj == nil && len(b.objTexts) > 0 {
		cur := b.objTexts[b.i]
		rect = Rectangle{Pt(0, Pos(cur).Y), Pos(cur).Add(Size(cur))}
	} else {
		rect = RectInParent(b.text)
		rect.Min.X = 0
	}
	SetColor(Color{1, 1, 1, .7})
	FillRect(rect)
}

type pkgObject struct {
	types.Object
	importPath, name, srcDir string
}

func (p pkgObject) GetName() string {
	if p.importPath == "" {
		// during object creation, all fields are blank except p.name which is being edited
		return p.name
	}
	return path.Base(p.importPath)
}

func (p pkgObject) GetPkg() *types.Package { return nil }

var (
	protoPointer   = types.NewTypeName(0, nil, "pointer", nil)
	protoArray     = types.NewTypeName(0, nil, "array", nil)
	protoSlice     = types.NewTypeName(0, nil, "slice", nil)
	protoMap       = types.NewTypeName(0, nil, "map", nil)
	protoChan      = types.NewTypeName(0, nil, "chan", nil)
	protoFunc      = types.NewTypeName(0, nil, "func", nil)
	protoInterface = types.NewTypeName(0, nil, "interface", nil)
	protoStruct    = types.NewTypeName(0, nil, "struct", nil)
)

func newProtoType(t *types.TypeName) types.Type {
	switch t {
	case protoPointer:
		return &types.Pointer{}
	case protoArray:
		return &types.Array{}
	case protoSlice:
		return &types.Slice{}
	case protoMap:
		return &types.Map{}
	case protoChan:
		return &types.Chan{Dir: types.SendRecv}
	case protoFunc:
		return &types.Signature{}
	case protoInterface:
		return &types.Interface{}
	case protoStruct:
		return &types.Struct{}
	}
	panic(fmt.Sprintf("not a proto type %#v", t))
}

type special struct {
	*types.Var
}

type field struct {
	*types.Var
	recv        types.Type
	addressable bool
}

type objects []types.Object

func (o objects) Len() int { return len(o) }
func (o objects) Less(i, j int) bool {
	return objLess(o[i], o[j])
}
func objLess(o1, o2 types.Object) bool {
	n1, n2 := o1.GetName(), o2.GetName()
	switch o1.(type) {
	case special:
		switch o2.(type) {
		case special:
			return n1 < n2
		default:
			return true
		}
	case *types.TypeName:
		switch o2.(type) {
		case special:
			return false
		case *types.TypeName:
			return n1 < n2
		default:
			return true
		}
	case *types.Func, *types.Builtin:
		switch o2.(type) {
		case special, *types.TypeName:
			return false
		case *types.Func, *types.Builtin:
			return n1 < n2
		default:
			return true
		}
	case *types.Var, field, *localVar:
		switch o2.(type) {
		default:
			return false
		case *types.Var, field, *localVar:
			return n1 < n2
		case *types.Const, *pkgObject:
			return true
		}
	case *types.Const:
		switch o2.(type) {
		default:
			return false
		case *types.Const:
			return n1 < n2
		case *pkgObject:
			return true
		}
	case *pkgObject:
		switch o2.(type) {
		default:
			return false
		case *pkgObject:
			return n1 < n2
		}
	}
	panic("unreachable")
}
func (o objects) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

// TODO: use the version from ssa or go/types
func intuitiveMethodSet(T types.Type) []*types.Selection {
	var result []*types.Selection
	mset := types.NewMethodSet(T)
	if _, ok := T.Underlying().(*types.Interface); ok {
		for i, n := 0, mset.Len(); i < n; i++ {
			result = append(result, mset.At(i))
		}
	} else {
		pmset := types.NewMethodSet(types.NewPointer(T))
		for i, n := 0, pmset.Len(); i < n; i++ {
			meth := pmset.At(i)
			if m := mset.Lookup(meth.Obj.GetPkg(), meth.Obj.GetName()); m != nil {
				meth = m
			}
			result = append(result, meth)
		}
	}
	return result
}

package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
	"go/ast"
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
	mode       browserMode
	currentPkg *types.Package
	imports    []*types.Package
	finished   bool
	accepted   func(types.Object)
	canceled   func()

	path    []types.Object
	indices []int
	i       int
	newObj  types.Object

	pathTexts, nameTexts []Text
	text                 *nodeNameText
	typeView             *typeView
	pkgNameText          *TextBase
}

type browserMode int

const (
	browse browserMode = iota
	fluxSourceOnly
	typesOnly
	compositeOrPtrType
	compositeType
	makeableType
)

func newBrowser(mode browserMode, currentPkg *types.Package, imports []*types.Package) *browser {
	b := &browser{mode: mode, currentPkg: currentPkg, imports: imports, accepted: func(types.Object) {}, canceled: func() {}}
	b.ViewBase = NewView(b)

	b.text = newNodeNameText(b)
	b.text.SetBackgroundColor(Color{0, 0, 0, 0})
	b.Add(b.text)

	b.pkgNameText = NewText("")
	b.pkgNameText.SetBackgroundColor(Color{0, 0, 0, .7})
	b.Add(b.pkgNameText)

	b.text.SetText("")

	return b
}

func (b *browser) cancel() {
	if !b.finished {
		b.finished = true
		b.canceled()
	}
}

type special struct {
	types.Object
	name string
}

func (s special) GetName() string { return s.name }

type buildPackage struct {
	types.Object
	*build.Package
}

func (p buildPackage) GetName() string {
	if p.Dir == "" {
		return ""
	}
	return path.Base(p.Dir)
}

type method struct {
	types.Object
	*types.Method
}

func (m method) GetName() string        { return m.Name }
func (m method) GetType() types.Type    { return m.Type }
func (m method) GetPkg() *types.Package { return m.Pkg }

type field struct {
	types.Object
	*types.Field
	recv *types.NamedType
}

func (f field) GetName() string        { return f.Name }
func (f field) GetType() types.Type    { return f.Type }
func (f field) GetPkg() *types.Package { return f.Pkg }

type objects []types.Object

func (o objects) Len() int { return len(o) }
func (o objects) Less(i, j int) bool {
	ni, nj := o[i].GetName(), o[j].GetName()
	switch o[i].(type) {
	case special:
		switch o[j].(type) {
		case special:
			return ni < nj
		default:
			return true
		}
	case *types.TypeName:
		switch o[j].(type) {
		case special:
			return false
		case *types.TypeName:
			return ni < nj
		default:
			return true
		}
	case *types.Func, method:
		switch o[j].(type) {
		case special, *types.TypeName:
			return false
		case *types.Func, method:
			return ni < nj
		default:
			return true
		}
	case *types.Var, field:
		switch o[j].(type) {
		default:
			return false
		case *types.Var, field:
			return ni < nj
		case *types.Const, buildPackage:
			return true
		}
	case *types.Const:
		switch o[j].(type) {
		default:
			return false
		case *types.Const:
			return ni < nj
		case buildPackage:
			return true
		}
	case buildPackage:
		switch o[j].(type) {
		default:
			return false
		case buildPackage:
			return ni < nj
		}
	}
	panic("unreachable")
}
func (o objects) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

var buildPackages = map[string]buildPackage{}

func (b browser) filteredObjs() (objs []types.Object) {
	add := func(obj types.Object) {
		if _, ok := obj.(buildPackage); !ok {
			switch b.mode {
			case fluxSourceOnly:
				if !fluxObjs[obj] {
					return
				}
			case typesOnly:
				if _, ok := obj.(*types.TypeName); !ok {
					return
				}
			case compositeOrPtrType, compositeType:
				switch obj {
				default:
					obj, ok := obj.(*types.TypeName)
					if !ok {
						return
					}
					t, ok := obj.Type.(*types.NamedType)
					if !ok {
						return
					}
					switch t.Underlying.(type) {
					default:
						return
					case *types.Array, *types.Slice, *types.Map, *types.Struct:
					}
				case protoPointer:
					if b.mode == compositeType {
						return
					}
				case protoArray, protoSlice, protoMap, protoStruct:
				}
			case makeableType:
				switch obj {
				default:
					obj, ok := obj.(*types.TypeName)
					if !ok {
						return
					}
					t, ok := obj.Type.(*types.NamedType)
					if !ok {
						return
					}
					switch t.Underlying.(type) {
					default:
						return
					case *types.Slice, *types.Map, *types.Chan:
					}
				case protoSlice, protoMap, protoChan:
				}
			}
			if b.currentPkg != nil {
				if p := obj.GetPkg(); p != nil && p != b.currentPkg && !ast.IsExported(obj.GetName()) {
					return
				}
			}
		}
		objs = append(objs, obj)
	}

	addPkgs := func(dir string) {
		if file, err := os.Open(dir); err == nil {
			if fileInfos, err := file.Readdir(-1); err == nil {
				for _, fileInfo := range fileInfos {
					name := fileInfo.Name()
					if fileInfo.IsDir() && unicode.IsLetter([]rune(name)[0]) && name != "testdata" {
						fullPath := filepath.Join(dir, name)
						buildPkg, ok := buildPackages[fullPath]
						if !ok {
							pkg, _ := build.ImportDir(fullPath, build.AllowBinary)
							buildPkg = buildPackage{nil, pkg}
							buildPackages[fullPath] = buildPkg
						}
						add(buildPkg)
					}
				}
			}
		}
	}

	if len(b.path) == 0 {
		if b.mode == browse {
			for _, name := range []string{"[]", "[]=", "addr", "call", "convert", "func", "if", "indirect", "loop", "typeAssert"} {
				objs = append(objs, special{name: name})
			}
		}
		pkgs := b.imports
		if b.currentPkg != nil {
			pkgs = append(pkgs, b.currentPkg)
		}
		for _, p := range pkgs {
			for _, obj := range p.Scope.Entries {
				add(obj)
			}
		}
		for _, obj := range types.Universe.Entries {
			add(obj)
		}
		for _, s := range []string{"+", "-", "*", "/", "%", "&", "|", "^", "&^", "&&", "||", "!", "==", "!=", "<", "<=", ">", ">="} {
			add(&types.Func{Name: s})
		}
		for _, t := range []*types.TypeName{protoPointer, protoArray, protoSlice, protoMap, protoChan, protoFunc, protoInterface, protoStruct} {
			add(t)
		}
		for _, dir := range build.Default.SrcDirs() {
			addPkgs(dir)
		}
	} else {
		switch obj := b.path[0].(type) {
		case buildPackage:
			if pkg, err := getPackage(obj.ImportPath); err == nil {
				for _, obj := range pkg.Scope.Entries {
					add(obj)
				}
			} else {
				if _, ok := err.(*build.NoGoError); !ok {
					fmt.Println(err)
				}
				pkgs[obj.ImportPath] = &types.Package{Name: obj.GetName(), Path: obj.ImportPath, Scope: &types.Scope{}}
			}
			addPkgs(obj.Dir)
		case *types.TypeName:
			// TODO: use types.NewMethodSet to get the entire method set
			// TODO: shouldn't go/types also have a NewFieldSet?
			if nt, ok := obj.Type.(*types.NamedType); ok {
				for _, m := range nt.Methods {
					add(method{nil, m})
				}
				switch ut := nt.Underlying.(type) {
				case *types.Struct:
					for _, f := range ut.Fields {
						add(field{nil, f, nt})
					}
				case *types.Interface:
					for _, m := range ut.Methods {
						add(method{nil, m})
					}
				}
			}
		}
	}

	// TODO: merge duplicate directories across srcDirs (warn if there is package shadowing)

	sort.Sort(objects(objs))
	return
}

func (b browser) currentObj() types.Object {
	objs := b.filteredObjs()
	if len(b.indices) == 0 || len(objs) == 0 {
		return nil
	}
	return objs[b.indices[b.i]]
}

func (b browser) lastPathText() (Text, bool) {
	if np := len(b.pathTexts); np > 0 {
		return b.pathTexts[np-1], true
	}
	return nil, false
}

func (b *browser) Paint() {
	rect := ZR
	if b.newObj == nil && len(b.nameTexts) > 0 {
		cur := b.nameTexts[b.i]
		rect = Rectangle{Pt(0, Pos(cur).Y), Pos(cur).Add(Size(cur))}
	} else {
		rect = RectInParent(b.text)
		rect.Min.X = 0
	}
	SetColor(Color{1, 1, 1, .7})
	FillRect(rect)
}

type nodeNameText struct {
	*TextBase
	b *browser
}

func newNodeNameText(b *browser) *nodeNameText {
	t := &nodeNameText{}
	t.TextBase = NewTextBase(t, "")
	t.b = b
	return t
}
func (t *nodeNameText) SetText(text string) {
	b := t.b
	if b.newObj == nil {
		if objs := b.filteredObjs(); len(objs) > 0 {
			for _, obj := range objs {
				if strings.HasPrefix(strings.ToLower(obj.GetName()), strings.ToLower(text)) {
					goto ok
				}
			}
			return
		}
	}
ok:
	currentIndex := 0
	n := len(b.indices)
	if n > 0 {
		b.i %= n
		if b.i < 0 {
			b.i += n
		}
		currentIndex = b.indices[b.i]
	}

	objs := b.filteredObjs()
	if b.newObj != nil {
		switch obj := b.newObj.(type) {
		case buildPackage:
			obj.Dir = text
		case *types.TypeName:
			obj.Name = text
		case *types.Func:
			obj.Name = text
		case method:
			obj.Name = text
		case *types.Var:
			obj.Name = text
		case *types.Const:
			obj.Name = text
		}
		newIndex := 0
		for i, obj := range objs {
			if obj.GetName() >= b.newObj.GetName() {
				switch obj.(type) {
				case buildPackage:
					if _, ok := b.newObj.(buildPackage); !ok {
						continue
					}
				case *types.Func:
					if _, ok := b.newObj.(*types.Func); !ok {
						continue
					}
				default:
					continue
				}
				newIndex = i
				break
			}
		}
		objs = append(objs[:newIndex], append([]types.Object{b.newObj}, objs[newIndex:]...)...)
		currentIndex = newIndex
	}

	b.indices = nil
	for i, obj := range objs {
		if strings.HasPrefix(strings.ToLower(obj.GetName()), strings.ToLower(text)) {
			b.indices = append(b.indices, i)
		}
	}
	n = len(b.indices)
	for i, index := range b.indices {
		if index >= currentIndex {
			b.i = i
			break
		}
	}
	if b.i >= n {
		b.i = n - 1
	}

	var cur types.Object
	if n > 0 {
		cur = objs[b.indices[b.i]]
	}
	if cur != nil {
		text = cur.GetName()[:len(text)]
	} else {
		text = ""
	}
	t.TextBase.SetText(text)

	if t, ok := b.lastPathText(); ok && cur != nil {
		sep := ""
		if _, ok := cur.(buildPackage); ok {
			sep = "/"
		} else {
			sep = "."
		}
		text := t.GetText()
		t.SetText(text[:len(text)-1] + sep)
	}
	xOffset := 0.0
	if t, ok := b.lastPathText(); ok {
		xOffset = Pos(t).X + Width(t)
	}

	for _, l := range b.nameTexts {
		l.Close()
	}
	b.nameTexts = []Text{}
	width := 0.0
	for i, activeIndex := range b.indices {
		obj := objs[activeIndex]
		l := NewText(obj.GetName())
		l.SetTextColor(getTextColor(obj, .7))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		b.Add(l)
		b.nameTexts = append(b.nameTexts, l)
		l.Move(Pt(xOffset, float64(n-i-1)*Height(l)))
		if Width(l) > width {
			width = Width(l)
		}
	}
	Raise(b.text)
	Resize(b, Pt(xOffset+width, float64(len(b.nameTexts))*Height(b.text)))

	yOffset := float64(n-b.i-1) * Height(b.text)
	b.text.Move(Pt(xOffset, yOffset))
	if b.typeView != nil {
		b.typeView.Close()
	}
	if pkg, ok := cur.(buildPackage); ok {
		t := b.pkgNameText
		t.SetText(pkg.Name)
		t.Move(Pt(xOffset+width+16, yOffset-(Height(t)-Height(b.text))/2))
		if pkg.Name != path.Base(pkg.Dir) {
			Show(t)
		} else {
			Hide(t)
		}
	} else {
		Hide(b.pkgNameText)
	}
	if cur != nil {
		b.text.SetTextColor(getTextColor(cur, 1))
		switch cur := cur.(type) {
		case *types.TypeName:
			if t, ok := cur.Type.(*types.NamedType); ok {
				b.typeView = newTypeView(&t.Underlying)
				b.Add(b.typeView)
			}
		case *types.Func, method, *types.Var, *types.Const, field:
			t := cur.GetType()
			b.typeView = newTypeView(&t)
			b.Add(b.typeView)
		}
		if b.typeView != nil {
			b.typeView.Move(Pt(xOffset+width+16, yOffset-(Height(b.typeView)-Height(b.text))/2))
		}
	}
	for _, p := range b.pathTexts {
		p.Move(Pt(Pos(p).X, yOffset))
	}

	Pan(b, Pt(0, yOffset))
}
func (t *nodeNameText) LostKeyFocus() { t.b.cancel() }
func (t *nodeNameText) KeyPress(event KeyEvent) {
	b := t.b
	switch event.Key {
	case KeyUp:
		if b.newObj == nil {
			b.i--
			t.SetText(t.GetText())
		}
	case KeyDown:
		if b.newObj == nil {
			b.i++
			t.SetText(t.GetText())
		}
	case KeyBackspace:
		if len(t.GetText()) > 0 {
			t.TextBase.KeyPress(event)
			break
		}
		fallthrough
	case KeyLeft:
		if len(b.path) > 0 && b.newObj == nil {
			previous := b.path[0]
			b.path = b.path[1:]
			b.i = 0
			for i, obj := range b.filteredObjs() {
				if obj == previous {
					b.indices = []int{i}
					break
				}
			}

			i := len(b.pathTexts) - 1
			b.pathTexts[i].Close()
			b.pathTexts = b.pathTexts[:i]

			t.SetText("")
		}
	case KeyEnter:
		cur := b.currentObj()
		if cur == nil && b.newObj == nil {
			return
		}
		if pkg, ok := cur.(buildPackage); ok && event.Shift {
			t := b.pkgNameText
			Show(t)
			t.Accept = func(s string) {
				if s != pkg.Name {
					pkg.Name = s
					savePackageName(pkg.Package)
				}
				b.text.SetText("")
				SetKeyFocus(b.text)
			}
			t.Reject = func() {
				b.text.SetText(b.text.GetText())
				SetKeyFocus(b.text)
			}
			SetKeyFocus(t)
			return
		}

		obj := b.newObj
		existing := false
		if obj == nil {
			obj = cur
		} else if cur != nil && obj.GetName() == cur.GetName() {
			obj = cur
			existing = true
		}
		if b.newObj != nil && !existing {
			switch obj := obj.(type) {
			case buildPackage:
				path := ""
				if len(b.path) > 0 {
					path = b.path[0].(buildPackage).Dir
				} else {
					d := build.Default.SrcDirs()
					path = d[len(d)-1]
				}
				if err := os.Mkdir(filepath.Join(path, obj.GetName()), 0777); err != nil {
					panic(err)
				}
			case *types.TypeName, *types.Func, *types.Var, *types.Const:
				pkg := b.currentPkg
				if len(b.path) > 0 {
					pkg = pkgs[b.path[0].(buildPackage).ImportPath]
				}
				if pkg != nil {
					pkg.Scope.Insert(obj)
				}
			case method:
				t := b.path[0].(*types.TypeName).Type.(*types.NamedType)
				t.Methods = append(t.Methods, obj.Method)
			}

			b.i = 0
			for i, child := range b.filteredObjs() {
				if child == obj {
					b.indices = []int{i}
					break
				}
			}
		}
		b.newObj = nil
		if _, ok := obj.(buildPackage); !ok {
			b.finished = true
			b.accepted(obj)
			return
		}
		fallthrough
	case KeyRight:
		if b.newObj == nil {
			switch obj := b.currentObj().(type) {
			case buildPackage, *types.TypeName:
				if t, ok := obj.(*types.TypeName); ok {
					if _, ok = t.Type.(*types.Basic); ok || t.Type == nil {
						break
					}
				}
				b.path = append([]types.Object{obj}, b.path...)
				b.indices = nil

				sep := ""
				if _, ok := obj.(buildPackage); ok {
					sep = "/"
				} else {
					sep = "."
				}
				pathText := NewText(obj.GetName() + sep)
				pathText.SetTextColor(getTextColor(obj, 1))
				pathText.SetBackgroundColor(Color{0, 0, 0, .7})
				b.Add(pathText)
				x := 0.0
				if t, ok := b.lastPathText(); ok {
					x = Pos(t).X + Width(t)
				}
				pathText.Move(Pt(x, 0))
				b.pathTexts = append(b.pathTexts, pathText)

				t.SetText("")
			}
		}
	case KeyEscape:
		if b.newObj == nil {
			b.cancel()
		} else {
			b.newObj = nil
			t.SetText("")
		}
	default:
		makeInPkg := false
		var pkg *types.Package
		var recv *types.TypeName
		if len(b.path) > 0 {
			switch obj := b.path[0].(type) {
			case buildPackage:
				makeInPkg = true
				pkg = pkgs[obj.ImportPath]
			case *types.TypeName:
				recv = obj
				pkg = obj.Pkg
			}
		}
		makeInRoot := len(b.path) == 0 && (b.currentPkg != nil || event.Text == "1")
		makeInType := recv != nil && event.Text == "3"
		if b.newObj == nil && b.mode != typesOnly && event.Command && (makeInRoot || makeInPkg || makeInType) {
			switch event.Text {
			case "1":
				b.newObj = buildPackage{nil, &build.Package{}}
			case "2":
				t := &types.TypeName{Pkg: pkg}
				t.Type = &types.NamedType{Obj: t}
				b.newObj = t
			case "3":
				if recv != nil {
					b.newObj = method{nil, &types.Method{QualifiedName: types.QualifiedName{Pkg: pkg}, Type: &types.Signature{Recv: &types.Var{Type: &types.Pointer{recv.Type}}}}}
				} else {
					b.newObj = &types.Func{Pkg: pkg, Type: &types.Signature{}}
				}
			case "4":
				b.newObj = &types.Var{Pkg: pkg}
			case "5":
				b.newObj = &types.Const{Pkg: pkg}
			default:
				t.TextBase.KeyPress(event)
				return
			}
			t.SetText("")
		} else {
			t.TextBase.KeyPress(event)
		}
	}
}

func getTextColor(obj types.Object, alpha float64) Color {
	switch obj.(type) {
	case special:
		return Color{1, 1, .6, alpha}
	case buildPackage:
		return Color{1, 1, 1, alpha}
	case *types.TypeName:
		return Color{.6, 1, .6, alpha}
	case *types.Func, method:
		return Color{1, .6, .6, alpha}
	case *types.Var, *types.Const, field:
		return Color{.6, .6, 1, alpha}
	}
	return Color{}
}

var (
	protoPointer   = &types.TypeName{Name: "pointer"}
	protoArray     = &types.TypeName{Name: "array"}
	protoSlice     = &types.TypeName{Name: "slice"}
	protoMap       = &types.TypeName{Name: "map"}
	protoChan      = &types.TypeName{Name: "chan"}
	protoFunc      = &types.TypeName{Name: "func"}
	protoInterface = &types.TypeName{Name: "interface"}
	protoStruct    = &types.TypeName{Name: "struct"}
)

func newProtoType(t *types.TypeName) (p types.Type) {
	switch t {
	case protoPointer:
		p = &types.Pointer{}
	case protoArray:
		p = &types.Array{}
	case protoSlice:
		p = &types.Slice{}
	case protoMap:
		p = &types.Map{}
	case protoChan:
		p = &types.Chan{Dir: ast.SEND | ast.RECV}
	case protoFunc:
		p = &types.Signature{}
	case protoInterface:
		p = &types.Interface{}
	case protoStruct:
		p = &types.Struct{}
	default:
		panic(fmt.Sprintf("not a proto type %#v", t))
	}
	return
}

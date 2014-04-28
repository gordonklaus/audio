package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
)

var (
	lineColor  = Color{.5, .5, .5, 1}
	focusColor = Color{1, 1, 1, .5}
	noColor    = Color{}
)

func color(obj types.Object, bright, funcAsVal bool) Color {
	alpha := .7
	if bright {
		alpha = 1
	}
	switch obj.(type) {
	case special:
		return Color{1, 1, .6, alpha}
	case *pkgObject:
		return Color{1, 1, 1, alpha}
	case *types.TypeName:
		return Color{.6, 1, .6, alpha}
	case *types.Func, *types.Builtin:
		if funcAsVal && obj.GetPkg() != nil { //Pkg==nil == builtin
			return color(&types.Var{}, bright, funcAsVal)
		}
		return Color{1, .6, .6, alpha}
	case *types.Var, *types.Const, field, *localVar:
		return Color{.6, .6, 1, alpha}
	}
	panic(fmt.Sprintf("unknown object type %T", obj))
}

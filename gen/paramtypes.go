package gen

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type SpecParamGenerator interface {
	GenerateSpecParam(pkg *packages.Package, iTracker *importTracker) (specParam, error)
}

type simpleIdent struct {
	ident *ast.Ident
}

func newSimpleIdent(ident *ast.Ident) SpecParamGenerator {
	return simpleIdent{
		ident: ident,
	}
}

func (s simpleIdent) GenerateSpecParam(pkg *packages.Package, iTracker *importTracker) (specParam, error) {
	return getSpecParam(pkg, s.ident, iTracker)
}

type pointerIdent struct {
	ident SpecParamGenerator
}

func newPointerIdent(ident SpecParamGenerator) SpecParamGenerator {
	return pointerIdent{ident: ident}
}

func (p pointerIdent) GenerateSpecParam(pkg *packages.Package, iTracker *importTracker) (specParam, error) {
	sp, err := p.ident.GenerateSpecParam(pkg, iTracker)
	if err != nil {
		return specParam{}, err
	}

	return specParam{
		Type:       fmt.Sprintf("*%s", sp.Type),
		ImportName: sp.ImportName,
	}, nil
}

type arrayIdent struct {
	ident SpecParamGenerator
}

func newArrayIdent(ident SpecParamGenerator) SpecParamGenerator {
	return arrayIdent{ident: ident}
}

func (a arrayIdent) GenerateSpecParam(pkg *packages.Package, iTracker *importTracker) (specParam, error) {
	sp, err := a.ident.GenerateSpecParam(pkg, iTracker)
	if err != nil {
		return specParam{}, err
	}

	return specParam{
		Type:       fmt.Sprintf("[]%s", sp.Type),
		ImportName: sp.ImportName,
	}, nil
}

type mapIdent struct {
	key SpecParamGenerator
	val SpecParamGenerator
}

func newMapIdent(key SpecParamGenerator, val SpecParamGenerator) SpecParamGenerator {
	return mapIdent{
		key: key,
		val: val,
	}
}

func (m mapIdent) GenerateSpecParam(pkg *packages.Package, iTracker *importTracker) (specParam, error) {
	spKey, err := m.key.GenerateSpecParam(pkg, iTracker)
	if err != nil {
		return specParam{}, err
	}

	spValue, err := m.val.GenerateSpecParam(pkg, iTracker)
	if err != nil {
		return specParam{}, err
	}

	typeKey := spKey.Type
	if spKey.ImportName != "" {
		typeKey = fmt.Sprintf("%s.%s", spKey.ImportName, spKey.Type)
	}
	typeValue := spValue.Type
	if spValue.ImportName != "" {
		typeValue = fmt.Sprintf("%s.%s", spValue.ImportName, spValue.Type)
	}

	return specParam{
		Type: fmt.Sprintf("map[%s]%s", typeKey, typeValue),
	}, nil
}

func getSpecParam(pkg *packages.Package, ident *ast.Ident, iTracker *importTracker) (specParam, error) {
	var (
		specParamIdent specParam
		imp            string
	)

	switch tpe := pkg.TypesInfo.TypeOf(ident).(type) {
	case *types.Named:
		specParamIdent = specParam{
			Type: ident.Name,
		}

		if tpe.Obj().Pkg() != nil {
			imp = tpe.Obj().Pkg().Path()
			specParamIdent.ImportName = tpe.Obj().Pkg().Name()
		}

	case *types.Basic:
		specParamIdent = specParam{
			Type: ident.Name,
		}
	default:
		return specParam{}, fmt.Errorf("unexpected field type: %s", pkg.TypesInfo.TypeOf(ident))
	}

	specParamIdent.ImportName = iTracker.getImportAlias(specParamIdent.ImportName, imp)

	return specParamIdent, nil
}

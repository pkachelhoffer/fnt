package gen

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type Spec struct {
	Name        string
	Functions   []SpecFunction
	Package     string
	PackageName string
}

type SpecFunction struct {
	Name    string
	Params  []SpecParam
	Returns []SpecParam
}

type SpecParam struct {
	Type       string
	Import     string
	ImportName string
}

func GetInterfaceSpec(path string, interfaceName string, targetPackageName string) (Spec, ImportTracker, error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  path,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return Spec{}, ImportTracker{}, fmt.Errorf("error loading package: %s", err)
	}

	for _, pkg := range pkgs {
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range gd.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Name.Name != interfaceName {
						continue
					}

					intType, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok {
						continue
					}

					var (
						intSpec Spec
					)

					intSpec.Name = typeSpec.Name.Name
					intSpec.PackageName = targetPackageName
					intSpec.Package = pkg.ID

					iTracker := newImportTracker(intSpec.Package)

					// Process interface functions
					for _, m := range intType.Methods.List {
						if len(m.Names) == 0 {
							continue
						}

						sf, err := processFunction(pkg, m, iTracker)
						if err != nil {
							return Spec{}, ImportTracker{}, err
						}

						intSpec.Functions = append(intSpec.Functions, sf)
					}

					return intSpec, *iTracker, nil
				}
			}
		}
	}

	return Spec{}, ImportTracker{}, fmt.Errorf("invalid interface specified or not found: %s", interfaceName)
}

func processFunction(pkg *packages.Package, field *ast.Field, iTracker *ImportTracker) (SpecFunction, error) {
	fnc, ok := field.Type.(*ast.FuncType)
	if !ok {
		return SpecFunction{}, fmt.Errorf("unexpected field type: %s", field.Type)
	}

	sf := SpecFunction{
		Name: field.Names[0].Name,
	}

	// Process parameters
	for _, p := range fnc.Params.List {
		ident, ok := getIdent(p)
		if !ok {
			continue
		}

		sp, err := processParameter(pkg, ident, iTracker)
		if err != nil {
			return SpecFunction{}, err
		}

		sf.Params = append(sf.Params, sp)
	}

	// Process returns
	if fnc.Results != nil {
		for _, p := range fnc.Results.List {
			ident, ok := getIdent(p)
			if !ok {
				continue
			}

			ret, err := processParameter(pkg, ident, iTracker)
			if err != nil {
				return SpecFunction{}, err
			}

			sf.Returns = append(sf.Returns, ret)
		}
	}

	return sf, nil
}

func getIdent(field *ast.Field) (*ast.Ident, bool) {
	switch tpe := field.Type.(type) {
	case *ast.Ident:
		return tpe, true
	case ast.Expr:
		exp, ok := tpe.(*ast.SelectorExpr)
		if !ok {
			fmt.Println("nope")
		} else {
			return exp.Sel, true
		}
	}
	return nil, false
}

func processParameter(pkg *packages.Package, param *ast.Ident, iTracker *ImportTracker) (SpecParam, error) {
	var specParam SpecParam

	switch tpe := pkg.TypesInfo.TypeOf(param).(type) {
	case *types.Named:
		specParam = SpecParam{
			Type: param.Name,
		}

		if tpe.Obj().Pkg() != nil {
			specParam.Import = tpe.Obj().Pkg().Path()
			specParam.ImportName = tpe.Obj().Pkg().Name()
		}

	case *types.Basic:
		specParam = SpecParam{
			Type:   param.Name,
			Import: "",
		}
	default:
		tt := pkg.TypesInfo.TypeOf(param)
		fmt.Println(tt)
		return SpecParam{}, fmt.Errorf("unexpected field type: %s", pkg.TypesInfo.TypeOf(param))
	}

	specParam.ImportName = iTracker.GetImportAlias(specParam.ImportName, specParam.Import)
	if specParam.Import == pkg.ID {
		specParam.Import = ""
	}

	return specParam, nil

}

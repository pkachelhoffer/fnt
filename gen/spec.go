package gen

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type spec struct {
	Name        string
	Functions   []specFunction
	Package     string
	PackageName string
	FileName    string
}

type specFunction struct {
	Name    string
	Params  []specParam
	Returns []specParam
}

type specParam struct {
	Type       string
	ImportName string
}

type identType struct {
	ident     *ast.Ident
	isPointer bool
	isArray   bool
}

type identDesc struct {
	primaryIdentType identType

	// Maps identities will have two -> Key and Value. The subIdentityType refers to the Map value
	subIdentityType identType
}

func getInterfaceSpec(path string, interfaceName string, targetPackageName string) (spec, importTracker, error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  path,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return spec{}, importTracker{}, fmt.Errorf("error loading package: %s", err)
	}

	for _, pkg := range pkgs {
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, specDecl := range gd.Specs {
					typeSpec, ok := specDecl.(*ast.TypeSpec)
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
						intSpec spec
					)

					intSpec.Name = typeSpec.Name.Name
					intSpec.PackageName = targetPackageName
					if intSpec.PackageName == "" {
						intSpec.PackageName = pkg.Types.Name()
					}
					intSpec.Package = pkg.ID
					intSpec.FileName = getFileName(pkg, syntax)

					iTracker := newImportTracker(intSpec.Package)

					// Process interface functions
					for _, m := range intType.Methods.List {
						if len(m.Names) == 0 {
							continue
						}

						sf, err := processFunction(pkg, m, iTracker)
						if err != nil {
							return spec{}, importTracker{}, err
						}

						intSpec.Functions = append(intSpec.Functions, sf)
					}

					return intSpec, *iTracker, nil
				}
			}
		}
	}

	return spec{}, importTracker{}, fmt.Errorf("invalid interface specified or not found: %s", interfaceName)
}

func getFileName(pkg *packages.Package, syntax *ast.File) string {
	var filePath string
	ast.Inspect(syntax, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				position := pkg.Fset.Position(typeSpec.Pos())
				filePath = position.Filename
			}
		}
		return true
	})

	return filePath
}

func processFunction(pkg *packages.Package, field *ast.Field, iTracker *importTracker) (specFunction, error) {
	fnc, ok := field.Type.(*ast.FuncType)
	if !ok {
		return specFunction{}, fmt.Errorf("unexpected field type: %s", field.Type)
	}

	sf := specFunction{
		Name: field.Names[0].Name,
	}

	// Process parameters
	for _, p := range fnc.Params.List {
		sp, err := processField(pkg, p, iTracker)
		if err != nil {
			return specFunction{}, fmt.Errorf("process field: %s", err.Error())
		}

		sf.Params = append(sf.Params, sp)
	}

	// Process returns
	if fnc.Results != nil {
		for _, p := range fnc.Results.List {
			sp, err := processField(pkg, p, iTracker)
			if err != nil {
				return specFunction{}, fmt.Errorf("process field: %s", err.Error())
			}

			sf.Returns = append(sf.Returns, sp)
		}
	}

	return sf, nil
}

func processField(pkg *packages.Package, p *ast.Field, iTracker *importTracker) (specParam, error) {
	ident, ok := getIdent(p)
	if !ok {
		return specParam{}, fmt.Errorf("failed getting type identity: %s", p.Type)
	}

	sp, err := processParameter(pkg, ident, iTracker)
	if err != nil {
		return specParam{}, err
	}

	return sp, nil
}

func getIdent(field *ast.Field) (identDesc, bool) {
	switch tpe := field.Type.(type) {
	case *ast.Ident:
		return identDesc{
			primaryIdentType: identType{ident: tpe},
		}, true
	case ast.Expr:
		return convertExpr(tpe)
	}
	return identDesc{}, false
}

func convertExpr(expr ast.Expr) (identDesc, bool) {
	switch tpeExpr := expr.(type) {
	case *ast.Ident:
		return identDesc{
			primaryIdentType: identType{ident: tpeExpr},
		}, true
	case *ast.SelectorExpr:
		return identDesc{
			primaryIdentType: identType{ident: tpeExpr.Sel},
		}, true
	case *ast.StarExpr:
		identTpe, ok := tpeExpr.X.(*ast.Ident)
		if ok {
			return identDesc{
				primaryIdentType: identType{ident: identTpe, isPointer: true},
			}, true
		}
	case *ast.ArrayType:
		id, ok := convertExpr(tpeExpr.Elt)
		if ok {
			return identDesc{
				primaryIdentType: identType{
					ident:     id.primaryIdentType.ident,
					isPointer: id.primaryIdentType.isPointer,
					isArray:   true},
			}, true
		}
	case *ast.MapType:
		primary, ok := convertExpr(tpeExpr.Key)
		if ok {
			secondary, ok := convertExpr(tpeExpr.Value)
			if ok {
				return identDesc{
					primaryIdentType: primary.primaryIdentType,
					subIdentityType:  secondary.primaryIdentType,
				}, ok
			}
		}
	}

	return identDesc{}, false
}

func getIdentityType(pkg *packages.Package, identType identType, iTracker *importTracker) (specParam, error) {
	var (
		specParamIdent specParam
		imp            string
	)

	switch tpe := pkg.TypesInfo.TypeOf(identType.ident).(type) {
	case *types.Named:
		specParamIdent = specParam{
			Type: identType.ident.Name,
		}

		if tpe.Obj().Pkg() != nil {
			imp = tpe.Obj().Pkg().Path()
			specParamIdent.ImportName = tpe.Obj().Pkg().Name()
		}

	case *types.Basic:
		specParamIdent = specParam{
			Type: identType.ident.Name,
		}
	default:
		return specParam{}, fmt.Errorf("unexpected field type: %s", pkg.TypesInfo.TypeOf(identType.ident))
	}

	specParamIdent.ImportName = iTracker.getImportAlias(specParamIdent.ImportName, imp)

	// If parameter is pointer type, add star
	if identType.isPointer {
		specParamIdent.Type = fmt.Sprintf("*%s", specParamIdent.Type)
	}

	// Add brackets if array
	if identType.isArray {
		specParamIdent.Type = fmt.Sprintf("[]%s", specParamIdent.Type)
	}

	return specParamIdent, nil
}

func processParameter(pkg *packages.Package, identDesc identDesc, iTracker *importTracker) (specParam, error) {
	var finalSpecParam specParam

	primaryParamIdent, err := getIdentityType(pkg, identDesc.primaryIdentType, iTracker)
	if err != nil {
		return specParam{}, err
	}

	if identDesc.subIdentityType.ident == nil {
		finalSpecParam = primaryParamIdent
	} else {
		subParamIndent, err := getIdentityType(pkg, identDesc.subIdentityType, iTracker)
		if err != nil {
			return specParam{}, err
		}
		type1 := primaryParamIdent.Type
		if primaryParamIdent.ImportName != "" {
			type1 = fmt.Sprintf("%s.%s", primaryParamIdent.ImportName, primaryParamIdent.Type)
		}
		type2 := subParamIndent.Type
		if subParamIndent.ImportName != "" {
			type2 = fmt.Sprintf("%s.%s", subParamIndent.ImportName, subParamIndent.Type)
		}
		// It's a map
		finalSpecParam.Type = fmt.Sprintf("map[%s]%s", type1, type2)
	}

	return finalSpecParam, nil
}

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
	Import     string
	ImportName string
}

type identDesc struct {
	ident     *ast.Ident
	isPointer bool
	isArray   bool
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
		ident, ok := getIdent(p)
		if !ok {
			return specFunction{}, fmt.Errorf("failed getting type identity: %s", p.Type)
		}

		sp, err := processParameter(pkg, ident, iTracker)
		if err != nil {
			return specFunction{}, err
		}

		sf.Params = append(sf.Params, sp)
	}

	// Process returns
	if fnc.Results != nil {
		for _, p := range fnc.Results.List {
			ident, ok := getIdent(p)
			if !ok {
				return specFunction{}, fmt.Errorf("failed getting type identity: %s", p.Type)
			}

			ret, err := processParameter(pkg, ident, iTracker)
			if err != nil {
				return specFunction{}, err
			}

			sf.Returns = append(sf.Returns, ret)
		}
	}

	return sf, nil
}

func getIdent(field *ast.Field) (identDesc, bool) {
	switch tpe := field.Type.(type) {
	case *ast.Ident:
		return identDesc{
			ident: tpe,
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
			ident: tpeExpr,
		}, true
	case *ast.SelectorExpr:
		return identDesc{
			ident: tpeExpr.Sel,
		}, true
	case *ast.StarExpr:
		identTpe, ok := tpeExpr.X.(*ast.Ident)
		if ok {
			return identDesc{
				ident:     identTpe,
				isPointer: true,
			}, true
		}
	case *ast.ArrayType:
		id, ok := convertExpr(tpeExpr.Elt)
		if ok {
			return identDesc{
				ident:     id.ident,
				isPointer: id.isPointer,
				isArray:   true,
			}, true
		}
	}

	return identDesc{}, false
}

func processParameter(pkg *packages.Package, ident identDesc, iTracker *importTracker) (specParam, error) {
	var specParamIdent specParam

	switch tpe := pkg.TypesInfo.TypeOf(ident.ident).(type) {
	case *types.Named:
		specParamIdent = specParam{
			Type: ident.ident.Name,
		}

		if tpe.Obj().Pkg() != nil {
			specParamIdent.Import = tpe.Obj().Pkg().Path()
			specParamIdent.ImportName = tpe.Obj().Pkg().Name()
		}

	case *types.Basic:
		specParamIdent = specParam{
			Type:   ident.ident.Name,
			Import: "",
		}
	default:
		return specParam{}, fmt.Errorf("unexpected field type: %s", pkg.TypesInfo.TypeOf(ident.ident))
	}

	specParamIdent.ImportName = iTracker.getImportAlias(specParamIdent.ImportName, specParamIdent.Import)
	if specParamIdent.Import == pkg.ID {
		specParamIdent.Import = ""
	}

	// If parameter is pointer type, add star
	if ident.isPointer {
		specParamIdent.Type = fmt.Sprintf("*%s", specParamIdent.Type)
	}

	// Add brackets if array
	if ident.isArray {
		specParamIdent.Type = fmt.Sprintf("[]%s", specParamIdent.Type)
	}

	return specParamIdent, nil
}

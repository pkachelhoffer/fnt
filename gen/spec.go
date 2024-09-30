package gen

import (
	"fmt"
	"go/ast"
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

	sp, err := ident.GenerateSpecParam(pkg, iTracker)
	if err != nil {
		return specParam{}, err
	}

	return sp, nil
}

func getIdent(field *ast.Field) (SpecParamGenerator, bool) {
	switch tpe := field.Type.(type) {
	case *ast.Ident:
		return newSimpleIdent(tpe), true
	case ast.Expr:
		return convertExpr(tpe)
	}
	return nil, false
}

func convertExpr(expr ast.Expr) (SpecParamGenerator, bool) {
	switch tpeExpr := expr.(type) {
	case *ast.Ident:
		return newSimpleIdent(tpeExpr), true
	case *ast.SelectorExpr:
		return newSimpleIdent(tpeExpr.Sel), true
	case *ast.StarExpr:
		identTpe, ok := tpeExpr.X.(*ast.Ident)
		if ok {
			return newPointerIdent(newSimpleIdent(identTpe)), true
		}
	case *ast.ArrayType:
		id, ok := convertExpr(tpeExpr.Elt)
		if ok {
			return newArrayIdent(id), true
		}
	case *ast.MapType:
		key, ok := convertExpr(tpeExpr.Key)
		if ok {
			value, ok := convertExpr(tpeExpr.Value)
			if ok {
				return newMapIdent(key, value), true
			}
		}
	}

	return nil, false
}

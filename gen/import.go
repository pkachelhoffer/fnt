package gen

import "fmt"

type importTracker struct {
	Imports     []importAlias
	AliasMap    map[string]int
	MainPackage string
}

type importAlias struct {
	Import     string
	ImportName string
	Alias      string
}

func newImportTracker(mainPackage string) *importTracker {
	return &importTracker{
		AliasMap:    make(map[string]int),
		MainPackage: mainPackage,
	}
}

func (it *importTracker) getImportAlias(importName string, imp string) string {
	// If the import path is the target package, return blank
	if imp == it.MainPackage {
		return ""
	}

	var generatedAlias string

	for _, is := range it.Imports {
		if imp == is.Import {
			return is.Alias
		}
	}

	for _, is := range it.Imports {
		if importName == is.ImportName {
			if imp == is.Import {
				return is.Alias
			} else {
				it.AliasMap[importName]++
				generatedAlias = fmt.Sprintf("%s_%d", importName, it.AliasMap[importName])
				break
			}
		}
	}

	if generatedAlias == "" {
		generatedAlias = importName
	}

	it.Imports = append(it.Imports, importAlias{
		Import:     imp,
		ImportName: importName,
		Alias:      generatedAlias,
	})

	return generatedAlias
}

func (it *importTracker) GetImportList() []importAlias {
	var ias []importAlias
	for _, ia := range it.Imports {
		ias = append(ias, ia)
	}

	return ias
}

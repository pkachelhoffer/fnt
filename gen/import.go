package gen

import "fmt"

type ImportTracker struct {
	Imports     []ImportAlias
	AliasMap    map[string]int
	MainPackage string
}

type ImportAlias struct {
	Import     string
	ImportName string
	Alias      string
}

func newImportTracker(mainPackage string) *ImportTracker {
	return &ImportTracker{
		AliasMap:    make(map[string]int),
		MainPackage: mainPackage,
	}
}

func (it *ImportTracker) GetImportAlias(importName string, imp string) string {
	// If the import path is the target package, return blank
	if imp == it.MainPackage {
		return ""
	}

	var generatedAlias string

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

	it.Imports = append(it.Imports, ImportAlias{
		Import:     imp,
		ImportName: importName,
		Alias:      generatedAlias,
	})

	return generatedAlias
}

func (it *ImportTracker) GetImportList() []ImportAlias {
	var ias []ImportAlias
	for _, ia := range it.Imports {
		ias = append(ias, ia)
	}

	return ias
}

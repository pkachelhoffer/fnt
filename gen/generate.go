package gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"golang.org/x/tools/imports"
	"log"
	"os"
	"text/template"
)

//go:embed templates/gen.tmpl
var genTmplText string

var (
	weldTpl = template.Must(template.New("").Parse(genTmplText))
)

type tmplData struct {
	Spec    Spec
	Imports []ImportAlias
}

func GenerateFile(spec Spec, it ImportTracker, target string) error {
	data := tmplData{
		Spec:    spec,
		Imports: it.GetImportList(),
	}

	var buf bytes.Buffer
	err := weldTpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error executing template: %s", err.Error())
	}

	imports.LocalPrefix = spec.Package
	src, err := imports.Process(target, buf.Bytes(), nil)
	if err != nil {
		log.Print(fmt.Errorf("failed processing imports"))
		src = buf.Bytes()
	}

	err = os.WriteFile(target, src, 0o644)
	if err != nil {
		return fmt.Errorf("error writing file: %s", err.Error())
	}

	return nil
}

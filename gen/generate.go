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

var (
	//go:embed templates/gen.tmpl
	genTmplText string

	weldTpl = template.Must(template.New("").Parse(genTmplText))
)

type tmplData struct {
	Spec    spec
	Imports []importAlias
}

func generateFile(spec spec, it importTracker, target string) error {
	data := tmplData{
		Spec:    spec,
		Imports: it.GetImportList(),
	}

	var buf bytes.Buffer
	err := weldTpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("executing template: %s", err.Error())
	}

	imports.LocalPrefix = spec.Package
	src, err := imports.Process(target, buf.Bytes(), nil)
	if err != nil {
		log.Print(fmt.Errorf("processing imports"))
		src = buf.Bytes()
	}

	err = os.WriteFile(target, src, 0o644)
	if err != nil {
		return fmt.Errorf("writing file: %s", err.Error())
	}

	return nil
}

func PerformTypeGeneration(path string, interfaceName string, targetPackageName string, outputFile string) error {
	spec, iTracker, err := getInterfaceSpec(path, interfaceName, targetPackageName)
	if err != nil {
		return fmt.Errorf("getting interface spec: %s", err)
	}

	err = generateFile(spec, iTracker, outputFile)
	if err != nil {
		return fmt.Errorf("generating file: %s", err)
	}

	return nil
}

package gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"golang.org/x/tools/imports"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	if target == "" {
		target = getFileOutputName(spec.FileName)
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

func getFileOutputName(sourceFilePath string) string {
	folder := filepath.Dir(sourceFilePath)
	fileNameWithExt := filepath.Base(sourceFilePath)
	ext := filepath.Ext(fileNameWithExt)
	fileNameWithoutExt := strings.TrimSuffix(fileNameWithExt, ext)

	return fmt.Sprintf("%s/%s_gen%s", folder, fileNameWithoutExt, ext)
}

func PerformTypeGeneration(inputPath string, interfaceName string, targetPackageName string, outputFile string) error {
	var err error
	if inputPath == "" {
		inputPath, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	s, iTracker, err := getInterfaceSpec(inputPath, interfaceName, targetPackageName)
	if err != nil {
		return fmt.Errorf("getting interface spec: %s", err)
	}

	err = generateFile(s, iTracker, outputFile)
	if err != nil {
		return fmt.Errorf("generating file: %s", err)
	}

	return nil
}

package main

import (
	"flag"
	"fnt/gen"
	"log"
)

var (
	interfaceName = flag.String("interface", "", "Name of interface to generate")
	inputPath     = flag.String("inputPath", "", "Folder of input path")
	outputFile    = flag.String("outputFile", "", "Output file name")
	outputPackage = flag.String("outputPackage", "", "Output package name")
)

func main() {
	flag.Parse()

	if *interfaceName == "" {
		log.Fatal("interface name not specified")
	}

	err := gen.PerformTypeGeneration(*inputPath, *interfaceName, *outputPackage, *outputFile)
	if err != nil {
		log.Fatal("error generating file", err)
	}
}

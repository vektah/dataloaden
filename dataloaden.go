package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/imports"
)

var (
	keyType = flag.String("keys", "int", "what type should the keys be")
	slice   = flag.Bool("slice", false, "this dataloader will return slices")
	out     = flag.String("out", "", "the output filename (optional)")
	private = flag.Bool("private", false, "the generated dataloader will be private")
)

type templateData struct {
	LoaderName string
	BatchName  string
	Package    string
	Name       string
	KeyType    string
	ValType    string
	Import     string
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	data, err := getData(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	var filename string
	if *out != "" {
		filename = *out
	} else if *slice {
		filename = data.Name + "_sliceloader_gen.go"
	} else {
		filename = data.Name + "_loader_gen.go"
	}

	writeTemplate(filename, data)
}

func getData(typeName string) (templateData, error) {
	var data templateData
	parts := strings.Split(typeName, ".")
	if len(parts) < 2 {
		return templateData{}, fmt.Errorf("type must be in the form package.Name")
	}

	wd, err := os.Getwd()
	if err != nil {
		return templateData{}, fmt.Errorf("cant determine working dir: %s", err.Error())
	}

	pkgData := getPackage(wd)
	name := parts[len(parts)-1]
	var loaderPrefix string
	if *private {
		loaderPrefix = lcFirst(name)
	} else {
		loaderPrefix = name
	}
	data.Package = pkgData
	data.LoaderName = loaderPrefix + "Loader"
	data.BatchName = lcFirst(name) + "Batch"
	data.Name = lcFirst(name)
	data.KeyType = *keyType

	prefix := "*"
	if *slice {
		prefix = "[]"
		data.LoaderName = loaderPrefix + "SliceLoader"
		data.BatchName = lcFirst(name) + "SliceBatch"
	}

	// if we are inside the same package as the type we don't need an import and can refer directly to the type
	pkgName := strings.Join(parts[:len(parts)-1], ".")
	if strings.HasSuffix(wd, pkgName) {
		data.ValType = prefix + name
	} else {
		data.Import = pkgName
		data.ValType = prefix + filepath.Base(data.Import) + "." + name
	}

	return data, nil
}

func getPackage(wd string) string {
	fset := token.NewFileSet()
	results, err := parser.ParseDir(fset, wd, func(info os.FileInfo) bool { return true }, parser.PackageClauseOnly)
	if err != nil {
		return filepath.Base(wd)
	}
	if len(results) > 0 {
		for pkgName := range results {
			return pkgName
		}
	}
	return filepath.Base(wd)
}

func writeTemplate(filename string, data templateData) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		log.Fatalf("generating code: %v", err)
	}

	src, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		log.Printf("unable to gofmt: %s", err.Error())
		src = buf.Bytes()
	}

	if err := ioutil.WriteFile(filename, src, 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func lcFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

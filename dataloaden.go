package main

import (
	"bytes"
	"flag"
	"fmt"
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

	filename := data.Name + "loader_gen.go"
	if *slice {
		filename = data.Name + "sliceloader_gen.go"
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

	pkgName := strings.Join(parts[:len(parts)-1], ".")
	name := parts[len(parts)-1]
	data.Package = filepath.Base(wd)
	data.LoaderName = name + "Loader"
	data.BatchName = lcFirst(name) + "Batch"
	data.Name = lcFirst(name)
	data.KeyType = *keyType

	prefix := "*"
	if *slice {
		prefix = "[]"
		data.LoaderName = name + "SliceLoader"
		data.BatchName = lcFirst(name) + "SliceBatch"
	}

	// if we are inside the same package as the type we don't need an import and can refer directly to the type
	if strings.HasSuffix(wd, pkgName) {
		data.ValType = prefix + name
	} else {
		data.Import = pkgName
		data.ValType = prefix + filepath.Base(data.Import) + "." + name
	}

	return data, nil
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

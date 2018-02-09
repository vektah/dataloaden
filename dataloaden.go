package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/imports"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dataloaden <type name>")
		fmt.Fprintln(os.Stderr, "	type name should be fully qualified, eg github.com/vektah/dataloaden/example.User")
		os.Exit(1)
	}

	data, err := getData(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	filename := data["name"] + "_dlgen.go"

	writeTemplate(filename, data)
}

func getData(typeName string) (map[string]string, error) {
	data := map[string]string{}
	parts := strings.Split(typeName, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("type must be in the form package.Name")
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cant determine working dir: %s", err.Error())
	}

	pkgName := strings.Join(parts[:len(parts)-1], ".")
	data["package"] = filepath.Base(wd)
	data["Name"] = parts[len(parts)-1]
	data["name"] = lcFirst(data["Name"])

	// if we are inside the same package as the type we don't need an import and can refer directly to the type
	fmt.Println(wd, pkgName)
	if strings.HasSuffix(wd, pkgName) {
		data["type"] = data["Name"]
	} else {
		data["import"] = pkgName
		data["type"] = filepath.Base(data["import"]) + "." + data["Name"]
	}

	return data, nil
}

func writeTemplate(filename string, data map[string]string) {
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

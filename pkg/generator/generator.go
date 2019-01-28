package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type templateData struct {
	LoaderName string
	BatchName  string
	Package    string
	Name       string
	KeyType    string
	ValType    string
	Import     string
	Slice      bool
}

func Generate(typename string, keyType string, slice bool, pointer bool, wd string) error {
	data, err := getData(typename, keyType, slice, pointer, wd)
	if err != nil {
		return err
	}

	filename := data.Name + "loader_gen.go"
	if data.Slice {
		filename = data.Name + "sliceloader_gen.go"
	}

	if err := writeTemplate(filepath.Join(wd, filename), data); err != nil {
		return err
	}

	return nil
}

func getData(typeName string, keyType string, slice bool, pointer bool, wd string) (templateData, error) {
	var data templateData
	parts := strings.Split(typeName, ".")
	if len(parts) < 2 {
		return templateData{}, fmt.Errorf("type must be in the form package.Name")
	}

	name := parts[len(parts)-1]
	importPath := strings.Join(parts[:len(parts)-1], ".")

	genPkg := getPackage(wd)
	if genPkg == nil {
		return templateData{}, fmt.Errorf("unable to find package info for " + wd)
	}

	data.Package = genPkg.Name
	data.LoaderName = name + "Loader"
	data.BatchName = lcFirst(name) + "Batch"
	data.Name = lcFirst(name)
	data.KeyType = keyType
	data.Slice = slice

	prefix := ""
	if slice {
		prefix = "[]"
		data.LoaderName = name + "SliceLoader"
		data.BatchName = lcFirst(name) + "SliceBatch"
	}

	if pointer {
		prefix = prefix + "*"
	}

	// if we are inside the same package as the type we don't need an import and can refer directly to the type
	if genPkg.PkgPath == importPath {
		data.ValType = prefix + name
	} else {
		data.Import = importPath
		data.ValType = prefix + filepath.Base(data.Import) + "." + name
	}

	return data, nil
}

func getPackage(dir string) *packages.Package {
	p, _ := packages.Load(&packages.Config{
		Dir: dir,
	}, ".")

	if len(p) != 1 {
		return nil
	}

	return p[0]
}

func writeTemplate(filepath string, data templateData) error {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return errors.Wrap(err, "generating code")
	}

	src, err := imports.Process(filepath, buf.Bytes(), nil)
	if err != nil {
		return errors.Wrap(err, "unable to gofmt")
	}

	if err := ioutil.WriteFile(filepath, src, 0644); err != nil {
		return errors.Wrap(err, "writing output")
	}

	return nil
}

func lcFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

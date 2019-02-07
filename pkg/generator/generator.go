package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
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

type ValueType struct {
	Name       string
	ImportPath string
	IsSlice    bool
	IsPointer  bool
}

func NewValueTypeFromString(typeName string) (*ValueType, error) {
	valueTypeRegexp := regexp.MustCompile(`^(\[\])?(\*)?(.+)+\.(\w+)+$`)
	matches := valueTypeRegexp.FindStringSubmatch(typeName)

	if matches == nil {
		return nil, errors.New("Invalid value type format. Expected: []*github.com/dataloaden/example.User")
	}

	return &ValueType{
		Name:       matches[4],
		ImportPath: matches[3],
		IsSlice:    matches[1] != "",
		IsPointer:  matches[2] != "",
	}, nil
}

func Generate(valueType *ValueType, keyType string, wd string) error {
	data, err := getData(valueType, keyType, wd)
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

func getData(valueType *ValueType, keyType string, wd string) (templateData, error) {
	var data templateData

	genPkg := getPackage(wd)
	if genPkg == nil {
		return templateData{}, fmt.Errorf("unable to find package info for " + wd)
	}

	name := valueType.Name

	data.Package = genPkg.Name
	data.LoaderName = name + "Loader"
	data.BatchName = lcFirst(name) + "Batch"
	data.Name = lcFirst(name)
	data.KeyType = keyType
	data.Slice = valueType.IsSlice

	prefix := ""
	if valueType.IsSlice {
		prefix = "[]"
		data.LoaderName = name + "SliceLoader"
		data.BatchName = lcFirst(name) + "SliceBatch"
	}

	if valueType.IsPointer {
		prefix = prefix + "*"
	}

	// if we are inside the same package as the type we don't need an import and can refer directly to the type
	if genPkg.PkgPath == valueType.ImportPath {
		data.ValType = prefix + name
	} else {
		data.Import = valueType.ImportPath
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

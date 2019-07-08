package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vektah/dataloaden/pkg/generator"
)

func main() {
	output := flag.String("output", "", "output filename")

	flag.Parse()

	args := flag.Args()
	if len(args) != 3 {
		fmt.Println("usage: name keyType valueType")
		fmt.Println(" example:")
		fmt.Println(" dataloaden 'UserLoader int []*github.com/my/package.User'")
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	if err := generator.Generate(&generator.GenerateInput{
		Name:       args[0],
		KeyType:    args[1],
		ValueType:  args[2],
		WorkingDir: wd,
		Output:     *output,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vektah/dataloaden/pkg/generator"
)

func main() {
	keyType := flag.String("keys", "int", "what type should the keys be")

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	wd, err := os.Getwd()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	valueType, err := generator.NewValueTypeFromString(flag.Arg(0))

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	if err := generator.Generate(valueType, *keyType, wd); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}

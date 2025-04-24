package main

import (
	"os"

	"golang.org/x/mod/modfile"
)

func parse() *modfile.File {
	buf, err := os.ReadFile("go.mod")
	if err != nil {
		out.Fatal("error reading go.mod", ERR_READ)
	}

	mod, err := modfile.Parse("go.mod", buf, nil)
	if err != nil {
		out.Fatal("error parsing go.mod", ERR_PARSE)
	}

	return mod
}

func save(mod *modfile.File) {
	buf, err := mod.Format()
	if err != nil {
		out.Fatal("error formatting go.mod", ERR_PARSE)
	}

	err = os.WriteFile("go.mod", buf, 0644)
	if err != nil {
		out.Fatal("error writing go.mod", ERR_WRITE)
	}
}

package main

import (
	"os"

	"golang.org/x/mod/modfile"
)

func parse(file string) *modfile.File {
	buf, err := os.ReadFile(file)
	if err != nil {
		out.Fatal("error reading go.mod", ERR_READ)
	}

	mod, err := modfile.Parse(file, buf, nil)
	if err != nil {
		out.Fatal("error parsing go.mod", ERR_PARSE)
	}

	return mod
}

func save(file string, mod *modfile.File) {
	buf, err := mod.Format()
	if err != nil {
		out.Fatal("error formatting go.mod", ERR_PARSE)
	}

	err = os.WriteFile(file, buf, 0644)
	if err != nil {
		out.Fatal("error writing go.mod", ERR_WRITE)
	}
}

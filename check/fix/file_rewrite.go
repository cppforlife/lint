package fix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io/ioutil"
)

type fileRewrite struct {
	Diff

	Func func() error

	File *ast.File
	Fset *token.FileSet
}

func (f fileRewrite) Fix() error {
	err := f.Func()
	if err != nil {
		return err
	}

	// gofmt defaults
	config := &printer.Config{
		Mode:     printer.Mode(printer.UseSpaces | printer.TabIndent),
		Tabwidth: 8,
	}

	var buf bytes.Buffer

	err = config.Fprint(&buf, f.Fset, f.File)
	if err != nil {
		return err
	}

	pos := f.Fset.Position(f.File.Package)
	if !pos.IsValid() {
		return fmt.Errorf("Invalid position: %v", pos.String())
	}

	return ioutil.WriteFile(pos.Filename, buf.Bytes(), 0)
}

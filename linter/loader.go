package linter

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"
	gotypes "code.google.com/p/go.tools/go/types"
)

type Loader interface {
	Programs() (<-chan *goloader.Program, <-chan error, error)
}

type LoadError struct {
	packageName     string
	loadErr         error
	typeCheckerErrs []error
}

func (e LoadError) Error() string {
	return fmt.Sprintf("Failed to load %s: %s", e.packageName, e.loadErr.Error())
}

func (e LoadError) UnderlyingErrs() []error {
	return e.typeCheckerErrs
}

type loader struct {
	goSrc  string
	args   []string
	logger *log.Logger
}

type dirContents struct {
	Path  string
	Paths []string // *.go
}

func (dc *dirContents) AddPath(path string) {
	if strings.HasSuffix(path, ".go") {
		dc.Paths = append(dc.Paths, path)
	}
}

func (dc *dirContents) PackageName(goSrc string) string {
	return strings.TrimPrefix(dc.Path, goSrc+string(filepath.Separator))
}

func NewLoaderFromArgs(goPath string, args []string, logger *log.Logger) (loader, error) {
	absGoSrc, err := filepath.Abs(filepath.Join(goPath, "src"))
	if err != nil {
		return loader{}, fmt.Errorf("gosrc cannot be determined %#v", err)
	}

	if len(args) != 1 {
		return loader{}, fmt.Errorf("Must provide exactly one package to load")
	}

	return loader{
		goSrc:  absGoSrc,
		args:   args,
		logger: logger,
	}, nil
}

func (l loader) Programs() (<-chan *goloader.Program, <-chan error, error) {
	if l.goSrc == "" {
		return nil, nil, fmt.Errorf("gopath is missing")
	}

	dir, err := filepath.Abs(filepath.Join(l.goSrc, l.args[0]))
	if err != nil {
		return nil, nil, fmt.Errorf("Building path %#v", err)
	}

	pathsByDir, err := l.groupPathsByDir(dir)
	if err != nil {
		return nil, nil, err
	}

	maxResults := len(pathsByDir)

	// Keeps all loaded programs
	programsCh := make(chan *goloader.Program, maxResults)

	// Keeps errors from loading programs
	errsCh := make(chan error, maxResults)

	// Populated by loading program goroutines
	endCh := make(chan struct{})

	// Load all packages in all non-empty directories
	for _, dc := range pathsByDir {
		go func(dc *dirContents) {
			if len(dc.Paths) > 0 {
				l.logger.Printf("Loading directory %s with %d file(s)\n", dc.Path, len(dc.Paths))

				program, err := l.loadProgram(dc.PackageName(l.goSrc))
				if err != nil {
					errsCh <- err
				} else {
					programsCh <- program
				}
			} else {
				l.logger.Printf("Skipping %s with 0 files\n", dc.Path)
			}

			endCh <- struct{}{}
		}(dc)
	}

	// Wait for all programs to be loaded
	go func() {
		for i := 0; i < maxResults; i++ {
			<-endCh
		}
		close(programsCh)
		close(errsCh)
		close(endCh)
	}()

	return programsCh, errsCh, nil
}

func (l loader) groupPathsByDir(dir string) (map[string]*dirContents, error) {
	pathsByDir := map[string]*dirContents{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip root walk directory
		if path == dir {
			return nil
		}

		d := filepath.Dir(path)
		dc, ok := pathsByDir[d]
		if !ok {
			dc = &dirContents{Path: d}
			pathsByDir[d] = dc
		}

		dc.AddPath(path)

		return nil
	}

	l.logger.Printf("Walking directory %s\n", dir)

	err := filepath.Walk(dir, walkFunc)
	if err != nil {
		return pathsByDir, fmt.Errorf("Failed walking %#v", err)
	}

	return pathsByDir, nil
}

func (l loader) loadProgram(packageName string) (*goloader.Program, error) {
	conf := goloader.Config{
		Fset:          token.NewFileSet(),
		TypeChecker:   gotypes.Config{},
		ParserMode:    parser.ParseComments,
		SourceImports: true,
	}

	// ImportWithTests includes both internal/external _test.go files
	err := conf.ImportWithTests(packageName)
	if err != nil {
		return nil, fmt.Errorf("Importing %s %#v", packageName, err)
	}

	var typeCheckerErrs []error

	// By default this outputs to Stderr
	conf.TypeChecker.Error = func(err error) {
		typeCheckerErrs = append(typeCheckerErrs, err)
	}

	program, err := conf.Load()
	if err != nil {
		return nil, LoadError{packageName, err, typeCheckerErrs}
	}

	return program, nil
}

package fix

import (
	"os"
	"path/filepath"
)

type FileRename struct {
	Diff

	DirPath string
}

func (f FileRename) Fix() error {
	beforePath := filepath.Join(f.DirPath, f.CurrentStr())
	afterPath := filepath.Join(f.DirPath, f.DesiredStr())
	return os.Rename(beforePath, afterPath)
}

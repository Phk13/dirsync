package filesync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	bar := progressbar.NewOptions(int(sourceFileStat.Size()),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("Copying..."),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	_, err = io.Copy(io.MultiWriter(destination, bar), source)
	return err
}

/* SyncFiles takes a slice with two-string-slices. Strings are full paths of origin and destination file to copy.  */
func SyncFiles(actions [][]string) error {
	for i, m := range actions {
		fmt.Printf("\n(%d/%d) \t- Copying %s \t\t-> \t%s\n", i+1, len(actions), m[0], m[1])
		err := copy(m[0], m[1])
		if err != nil {
			return err
		}
	}
	return nil
}
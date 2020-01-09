package tftest

import (
	"io"
	"os"
	"path/filepath"
)

func copyFile(src string, dest string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err == nil {
		srcInfo, err := os.Stat(src)
		if err != nil {
			err = os.Chmod(dest, srcInfo.Mode())
		}

	}

	return
}

// CopyDir is a simplistic function for recursively copying a directory to a new path.
// It is intended only for limited internal use and does not cover all edge cases.
func copyDir(srcDir string, destDir string) (err error) {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(srcDir)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		srcPath := filepath.Join(srcDir, obj.Name())
		destPath := filepath.Join(destDir, obj.Name())

		if obj.IsDir() {
			err = copyDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}

	}
	return
}

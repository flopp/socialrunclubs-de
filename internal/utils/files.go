package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flopp/go-filehash"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func MakeDir(dir string) error {
	// create directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// CopyFile copies a file from src to dst. If dst exists, it will be overwritten.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src file: %w", err)
	}
	defer in.Close()

	// Create the destination directory if it doesn't exist
	if err := MakeDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("create dst dir: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src file: %w", err)
	}
	if err := os.Chmod(dst, info.Mode()); err != nil {
		return fmt.Errorf("chmod dst file: %w", err)
	}

	return nil
}

func CopyHash(src, dst string) (string, error) {
	return filehash.Copy(src, dst, "HASH")
}

func MustCopyHash(src, dst string) string {
	res, err := CopyHash(src, dst)
	if err != nil {
		panic(err)
	}
	return res
}

package app

import (
	"fmt"
	"path/filepath"

	"github.com/flopp/socialrunclubs-de/internal/utils"
)

func CopyAssets(config Config) error {
	// copy static files to output directory
	staticFiles := []string{"static/style.css"}
	for _, file := range staticFiles {
		dest := filepath.Join(config.OutputDir, "static", filepath.Base(file))
		if err := utils.CopyFile(file, dest); err != nil {
			return fmt.Errorf("copy static file %s to %s: %w", file, dest, err)
		}
	}

	return nil
}

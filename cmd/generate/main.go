package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	OutputDir string
}

// copyFile copies a file from src to dst. If dst exists, it will be overwritten.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

// loadConfig loads configuration from a JSON file into the given config struct.
func loadConfig(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return err
	}
	return nil
}

func main() {
	// read config file from command line (e.g., config.json)
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	// load config from file
	config := Config{}
	if err := loadConfig(*configFile, &config); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// use config
	fmt.Printf("Output directory: %s\n", config.OutputDir)

	// create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// copy static files to output directory
	staticFiles := []string{"static/index.html", "static/style.css"}
	for _, file := range staticFiles {
		dest := filepath.Join(config.OutputDir, filepath.Base(file))
		if err := copyFile(file, dest); err != nil {
			log.Fatalf("Error copying static file %s to %s: %v", file, dest, err)
		}
	}
}

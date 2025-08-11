package app

import (
	"fmt"
	"path/filepath"

	"github.com/flopp/socialrunclubs-de/internal/utils"
)

func trimPath(path string, prefix string) (string, error) {
	relPath, err := filepath.Rel(prefix, path)
	if err != nil {
		return "", err
	}
	return "/" + relPath, nil
}

func download(url, target string, config Config) (string, error) {
	f, err := utils.DownloadHash(url, filepath.Join(config.OutputDir, target))
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}

	t, err := trimPath(f, config.OutputDir)
	if err != nil {
		return "", fmt.Errorf("trim path %s: %w", f, err)
	}

	return t, nil
}

func CopyAssets(config Config) ([]string, []string, error) {
	cssFiles := make([]string, 0)
	jsFiles := make([]string, 0)

	// fetch additional assets from remote server
	picocss, err := download("https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css", "static/pico.HASH.css", config)
	if err != nil {
		return nil, nil, fmt.Errorf("download pico.min.css: %w", err)
	}
	cssFiles = append(cssFiles, picocss)

	// leaflet
	leafletCSS, err := download("https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/leaflet.min.css", "static/leaflet.HASH.css", config)
	if err != nil {
		return nil, nil, fmt.Errorf("download leaflet.css: %w", err)
	}
	cssFiles = append(cssFiles, leafletCSS)

	leafletJS, err := download("https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/leaflet.min.js", "static/leaflet.HASH.js", config)
	if err != nil {
		return nil, nil, fmt.Errorf("download leaflet.js: %w", err)
	}
	jsFiles = append(jsFiles, leafletJS)

	if _, err := download("https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-icon.png", "static/images/marker-icon.png", config); err != nil {
		return nil, nil, fmt.Errorf("download marker-icon.png: %w", err)
	}

	if _, err := download("https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-icon-2x.png", "static/images/marker-icon-2x.png", config); err != nil {
		return nil, nil, fmt.Errorf("download marker-icon-2x.png: %w", err)
	}

	if _, err := download("https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-shadow.png", "static/images/marker-shadow.png", config); err != nil {
		return nil, nil, fmt.Errorf("download marker-shadow.png: %w", err)
	}

	// umami
	umamiJS, err := download("https://cloud.umami.is/script.js", "static/umami.HASH.js", config)
	if err != nil {
		return nil, nil, fmt.Errorf("download umami.js: %w", err)
	}
	jsFiles = append(jsFiles, umamiJS)

	// copy static files to output directory
	styleCSS, err := utils.CopyHash("static/style.css", filepath.Join(config.OutputDir, "static", "style.HASH.css"))
	if err != nil {
		return nil, nil, fmt.Errorf("copy static file %s: %w", "static/style.css", err)
	}
	styleCSS, err = trimPath(styleCSS, config.OutputDir)
	if err != nil {
		return nil, nil, fmt.Errorf("trim path %s: %w", styleCSS, err)
	}
	cssFiles = append(cssFiles, styleCSS)

	scriptJS, err := utils.CopyHash("static/script.js", filepath.Join(config.OutputDir, "static", "script.HASH.js"))
	if err != nil {
		return nil, nil, fmt.Errorf("copy static file %s: %w", "static/script.js", err)
	}
	scriptJS, err = trimPath(scriptJS, config.OutputDir)
	if err != nil {
		return nil, nil, fmt.Errorf("trim path %s: %w", scriptJS, err)
	}
	jsFiles = append(jsFiles, scriptJS)

	return cssFiles, jsFiles, nil
}

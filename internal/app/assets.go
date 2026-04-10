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

func fetchAsset(url string, targetFile string, cssFiles *[]string, jsFiles *[]string, config Config) error {
	assetPath, err := download(url, targetFile, config)
	if err != nil {
		return err
	}

	if filepath.Ext(assetPath) == ".css" {
		*cssFiles = append(*cssFiles, assetPath)
	} else if filepath.Ext(assetPath) == ".js" {
		*jsFiles = append(*jsFiles, assetPath)
	}

	return nil
}

func CopyAssets(config Config) ([]string, []string, error) {
	cssFiles := make([]string, 0)
	jsFiles := make([]string, 0)

	jsDelivr := "https://cdn.jsdelivr.net/npm"
	picoCssUrl := jsDelivr + "/@picocss/pico@2"
	leafletUrl := jsDelivr + "/leaflet@1.9.4/dist"
	markerClusterUrl := jsDelivr + "/leaflet.markercluster@1.5.3/dist"
	gestureHandlingUrl := jsDelivr + "/@gstat/leaflet-gesture-handling@1.2.8/dist"

	type Asset struct {
		URL    string
		Target string
	}

	assets := []Asset{
		{picoCssUrl + "/css/pico.min.css", "static/pico.HASH.css"},
		{leafletUrl + "/leaflet.min.css", "static/leaflet.HASH.css"},
		{leafletUrl + "/leaflet.min.js", "static/leaflet.HASH.js"},
		{leafletUrl + "/images/marker-icon.png", "static/images/marker-icon.png"},
		{leafletUrl + "/images/marker-icon-2x.png", "static/images/marker-icon-2x.png"},
		{leafletUrl + "/images/marker-shadow.png", "static/images/marker-shadow.png"},
		{markerClusterUrl + "/MarkerCluster.Default.css", "static/marker-cluster.HASH.css"},
		{markerClusterUrl + "/leaflet.markercluster.min.js", "static/marker-cluster.HASH.js"},
		{gestureHandlingUrl + "/leaflet-gesture-handling.min.js", "static/leaflet-gesture-handling.HASH.js"},
		{gestureHandlingUrl + "/leaflet-gesture-handling.min.css", "static/leaflet-gesture-handling.HASH.css"},
		{"https://cloud.umami.is/script.js", "static/umami.HASH.js"},
	}

	// fetch additional assets from remote server
	for _, asset := range assets {
		if err := fetchAsset(asset.URL, asset.Target, &cssFiles, &jsFiles, config); err != nil {
			return nil, nil, fmt.Errorf("fetch %s: %w", asset.Target, err)
		}
	}

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

	icons := []string{
		"apple-touch-icon.png",
		"favicon-96x96.png",
		"favicon.ico",
		"favicon.svg",
		"logo.svg",
	}
	for _, icon := range icons {
		if err := utils.CopyFile("static/"+icon, filepath.Join(config.OutputDir, icon)); err != nil {
			return nil, nil, fmt.Errorf("copy static file %s: %w", "static/"+icon, err)
		}
	}

	// copy llms.txt to output directory root
	if err := utils.CopyFile("static/llms.txt", filepath.Join(config.OutputDir, "llms.txt")); err != nil {
		return nil, nil, fmt.Errorf("copy static file %s: %w", "static/llms.txt", err)
	}

	return cssFiles, jsFiles, nil
}

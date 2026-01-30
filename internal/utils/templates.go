package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type TemplateData interface {
	IsRemoteTarget() bool
	BasePath() string
}

var templates = make(map[string]*template.Template)

func loadTemplate(name string, data TemplateData) (*template.Template, error) {
	name2 := filepath.Base(name)

	if t, ok := templates[name2]; ok {
		return t, nil
	}

	// collect all *.html files in templates/parts folder
	parts, err := filepath.Glob("templates/parts/*.html")
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, 1+len(parts))
	files = append(files, fmt.Sprintf("templates/%s", name))
	files = append(files, parts...)
	t, err := template.New(name2).Funcs(template.FuncMap{
		"BasePath": func(p string) string {
			ifFile := strings.Contains(filepath.Base(p), ".")
			if data.IsRemoteTarget() {
				if !ifFile && !strings.HasSuffix(p, "/") {
					return p + "/"
				}
				return p
			}

			res := data.BasePath()
			if !strings.HasPrefix(p, "/") {
				res += "/"
			}
			res += p

			if strings.HasSuffix(p, "/") {
				res += "index.html"
			} else if !strings.Contains(filepath.Base(p), ".") {
				res += "/index.html"
			}
			return res
		},
	}).ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	templates[name] = t
	return t, nil
}

func executeTemplateToBuffer(templateName string, data TemplateData) (*bytes.Buffer, error) {
	// load template
	templ, err := loadTemplate(templateName, data)
	if err != nil {
		return nil, err
	}

	// render to buffer
	var buffer bytes.Buffer
	err = templ.Execute(&buffer, data)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func prepareOutputFile(fileName string) (*os.File, error) {
	// create output folder + file
	outDir := filepath.Dir(fileName)
	err := MakeDir(outDir)
	if err != nil {
		return nil, err
	}

	// create output file
	out, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ExecuteTemplate(templateName string, fileName string, data TemplateData) error {
	buffer, err := executeTemplateToBuffer(templateName, data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	out, err := prepareOutputFile(fileName)
	if err != nil {
		return fmt.Errorf("prepare output file: %w", err)
	}
	defer out.Close()

	// write buffer to output file
	_, err = out.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("write buffer to output file: %w", err)
	}

	return nil
}

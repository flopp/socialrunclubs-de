package utils

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Row []string
type Sheet struct {
	Name string
	Rows []Row
}

func GetSheets(api_key string, sheet_id string) ([]Sheet, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithAPIKey(api_key))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	response, err := srv.Spreadsheets.Get(sheet_id).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		return nil, fmt.Errorf("getting sheets: %w", err)
	}
	if response.HTTPStatusCode != 200 {
		return nil, fmt.Errorf("getting sheets: http status %v", response.HTTPStatusCode)
	}

	names := make([]string, 0)
	for _, v := range response.Sheets {
		prop := v.Properties
		names = append(names, prop.Title)
	}

	sheets := make([]Sheet, 0)
	for _, name := range names {
		resp, err := srv.Spreadsheets.Values.Get(sheet_id, name).Do()
		if err != nil {
			return nil, fmt.Errorf("getting values for sheet %s: %w", name, err)
		}
		if resp.HTTPStatusCode != 200 {
			return nil, fmt.Errorf("getting values for sheet %s: http status %v", name, resp.HTTPStatusCode)
		}

		sheet := Sheet{
			Name: name,
			Rows: make([]Row, 0),
		}
		for _, r := range resp.Values {
			row := make(Row, len(r))
			for i, v := range r {
				row[i] = fmt.Sprintf("%v", v)
			}
			sheet.Rows = append(sheet.Rows, row)
		}

		sheets = append(sheets, sheet)
	}

	return sheets, nil
}

// ValidateColumns checks that the header row matches exactly the required columns (no extra, no missing, no duplicates)
func ValidateColumns(header []string, required []string) (map[string]int, error) {
	if len(header) != len(required) {
		return nil, fmt.Errorf("expected %d columns, got %d", len(required), len(header))
	}
	colIdx := map[string]int{}
	seen := map[string]bool{}
	for i, col := range header {
		if seen[col] {
			return nil, fmt.Errorf("duplicate column: %s", col)
		}
		seen[col] = true
		colIdx[col] = i
	}
	for _, col := range required {
		if _, ok := colIdx[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}
	for col := range colIdx {
		found := false
		for _, req := range required {
			if col == req {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unexpected column: %s", col)
		}
	}
	return colIdx, nil
}

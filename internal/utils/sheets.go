package utils

import (
	"context"
	"fmt"

	googlesheetswrapper "github.com/flopp/go-googlesheetswrapper"
)

type Row []string
type Sheet struct {
	Name string
	Rows []Row
}

func GetSheets(api_key string, sheet_id string) ([]Sheet, error) {
	ctx := context.Background()
	client, err := googlesheetswrapper.New(api_key, sheet_id)
	if err != nil {
		return nil, fmt.Errorf("creating sheets client: %w", err)
	}

	all, err := client.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading all sheets: %w", err)
	}

	sheets := make([]Sheet, 0, len(all))
	for name, rows := range all {
		sheet := Sheet{
			Name: name,
			Rows: make([]Row, len(rows)),
		}
		for i, r := range rows {
			row := make(Row, len(r))
			copy(row, r)
			sheet.Rows[i] = row
		}
		sheets = append(sheets, sheet)
	}

	return sheets, nil
}

// ValidateColumns is a wrapper for ExtractHeader from go-googlesheetswrapper.
func ValidateColumns(header []string, required []string) (map[string]int, error) {
	return googlesheetswrapper.ExtractHeader([][]string{header}, required, false)
}

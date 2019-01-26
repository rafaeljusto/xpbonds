package xpbonds

import (
	"io"
	"regexp"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/pkg/errors"
)

const invalidCell = "#VALUE!"

var reFocusedSheet = regexp.MustCompile("(?i)^focused [a-z]+$")

func parseExcel(excel io.Reader, dateFormat DateFormat, focusedOnly bool) (Bonds, error) {
	xlsx, err := excelize.OpenReader(excel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open excel")
	}

	sheets := xlsx.GetSheetMap()
	var bonds Bonds
	for _, sheet := range sheets {
		if focusedOnly && !reFocusedSheet.MatchString(sheet) {
			continue
		}

		rows := xlsx.GetRows(sheet)
		sheetBonds, err := parseSheet(rows, dateFormat)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse sheet '%s'", sheet)
		}
		bonds = append(bonds, sheetBonds...)
	}

	return bonds, nil
}

func parseSheet(rows [][]string, dateFormat DateFormat) (Bonds, error) {
	normalized := make(Bonds, 0, len(rows))

	for _, r := range rows {
		row := row(r)
		if row.ignore() {
			continue
		}

		bond, err := parseBond(row, dateFormat)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse bond")
		}

		normalized = append(normalized, bond)
	}

	return normalized, nil
}

type row []string

func (r row) ignore() bool {
	if len(r) != 16 {
		return true
	}

	cell := strings.TrimSpace(r[0])
	if cell == "" || cell == "Name" {
		return true
	}

	// detect and remove lines without value
	empty := true
	for _, cell := range r[1:] {
		if cell != "" {
			empty = false
			break
		}
	}
	return empty
}

func (r row) get(i int) string {
	if i < 0 || i >= len(r) || r[i] == invalidCell {
		return ""
	}

	v := r[i]
	v = strings.ToLower(v)
	v = strings.TrimSpace(v)

	if strings.Contains(v, "n.a.") || strings.Contains(v, "n/a") {
		return ""
	}

	return r[i]
}

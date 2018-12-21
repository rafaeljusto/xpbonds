package xpbonds

import (
	"log"
	"regexp"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/pkg/errors"
)

var reNumber = regexp.MustCompile(`^[[:digit:]]+$`)

// ignoreCells are the cell contents that will identify rows that should be
// ignored when parsing the converted excel.
var ignoreCells = map[string]bool{
	"Name": true,
	"BRAZIL BONDS (USD) - DAILY INDICATIVE RUN": true,
	"Disclaimers": true,
}

func parseExcel(excel string) (Bonds, error) {
	xlsx, err := excelize.OpenFile(excel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open excel")
	}

	rows := xlsx.GetRows("Sheet1")
	normalized := make(Bonds, 0, len(rows))

	for i, row := range rows {
		if ignoreRow(row) {
			continue
		}

		log.Printf("Row %d: %#v", i, row)

		bond, err := parseBond(row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse bond")
		}

		normalized = append(normalized, bond)
	}

	return normalized, nil
}

func ignoreRow(row []string) bool {
	if len(row) != 16 {
		return true
	}

	cell := strings.TrimSpace(row[0])
	if cell == "" || ignoreCells[cell] || reNumber.MatchString(cell) {
		return true
	}

	// detect and remove lines without value
	empty := true
	for _, cell := range row[1:] {
		if cell != "" {
			empty = false
			break
		}
	}
	return empty
}

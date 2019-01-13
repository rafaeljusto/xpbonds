package xpbonds

import (
	"bytes"
	"context"
	"encoding/base64"
	"sort"

	"github.com/pkg/errors"
)

// FindBestBonds determinates the best bonds from the bond report. It expects
// the report to contain a xlsx content (Excel) encoded in base64. After parsing
// the xlxs it performs some filtering and sorting actions to determinate the
// best bond.
func FindBestBonds(ctx context.Context, report BondReport) (Bonds, error) {
	excel, err := base64.StdEncoding.DecodeString(report.XLXSReport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode excel report")
	}

	bonds, err := parseExcel(bytes.NewReader(excel), report.DateFormat)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse excel")
	}

	bonds = bonds.Filter(report.Filter)
	bonds.FillCurrentPrice()
	sort.Sort(bonds)
	return bonds, nil
}

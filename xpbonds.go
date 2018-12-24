package xpbonds

import (
	"context"
	"net/http"
	"sort"

	"github.com/pkg/errors"
)

// FindBestBonds determinates the best bonds from the bond report. It downloads
// the bond report from the given location, converts it from PDF to Excel format
// and perform some sorting actions to determinate the best bond.
func FindBestBonds(ctx context.Context, report BondReport) (Bonds, error) {
	response, err := http.Get(report.Location)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get report")
	}
	defer response.Body.Close()

	excel, err := pdfToExcel(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert from PDF to excel")
	}

	bonds, err := parseExcel(excel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse excel")
	}

	bonds = bonds.Filter()
	bonds.FillCurrentPrice()
	sort.Sort(bonds)
	return bonds, nil
}

package main

import (
	"context"
	"net/http"
	"sort"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

// HandleRequest determinates the best bonds from the given bond rates. It
// downloads the bond rates from the given location, converts it from PDF to
// Excel format and perform some sorting actions to determinate the best bond.
func HandleRequest(ctx context.Context, rates BondRates) (Bonds, error) {
	response, err := http.Get(rates.Location)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rates")
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
	sort.Sort(bonds)
	return bonds, nil
}

func main() {
	lambda.Start(HandleRequest)
}

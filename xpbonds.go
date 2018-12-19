package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

var reNumber = regexp.MustCompile(`^[[:digit:]]+$`)

// BondRates contains the location of the bond rates.
type BondRates struct {
	Location string `json:"location"`
}

// Bond contains the bond descriptions.
type Bond struct {
	Name            string     `json:"name"`
	Security        string     `json:"security"`
	Coupon          float64    `json:"coupon"`
	Yield           float64    `json:"yield"`
	Maturity        *time.Time `json:"maturity"`
	LastPrice       float64    `json:"lastPrice"`
	Duration        float64    `json:"duration"`
	YearsToMaturity float64    `json:"yearsToMaturity"`
	MinimumPiece    float64    `json:"minimumPiece"`
	Country         string     `json:"country"`
	Risk            BondRisk   `json:"risk"`
	Code            string     `json:"code"`
}

// Bonds is a collection of Bond.
type Bonds []Bond

func (b Bonds) Len() int {
	return len(b)
}

func (b Bonds) Less(i, j int) bool {
	return b[i].Coupon > b[j].Coupon
}

func (b Bonds) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// BondRisk determinates the risk according to different entities.
type BondRisk struct {
	StandardPoor string `json:"standardPoor"`
	Moody        string `json:"moody"`
	Fitch        string `json:"fitch"`
}

// HandleRequest determinates the best bonds from the given bond rates. It
// downloads the bond rates from the given location, converts it from PDF to
// Excel format and perform some sorting actions to determinate the best bond.
func HandleRequest(ctx context.Context, rates BondRates) (Bonds, error) {
	log.Printf("Retrieving rates from location %s", rates.Location)

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

	log.Printf("Bonds before filtering: %#v", bonds)
	bonds = filterBonds(bonds)
	log.Printf("Bonds after filtering: %#v", bonds)
	sort.Sort(bonds)
	return bonds, nil
}

func parseExcel(excel string) (Bonds, error) {
	xlsx, err := excelize.OpenFile(excel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open excel")
	}

	rows := xlsx.GetRows("Sheet1")
	normalized := make(Bonds, 0, len(rows))
	log.Printf("Parsing %d rows", len(rows))

	for i, row := range rows {
		if len(row) != 16 {
			log.Printf("Warning: Row with only %d columns", len(row))
			continue
		}

		ignoreCells := map[string]bool{
			"Name": true,
			"BRAZIL BONDS (USD) - DAILY INDICATIVE RUN": true,
			"Disclaimers": true,
		}

		cell := strings.TrimSpace(row[0])
		if cell == "" || ignoreCells[cell] || reNumber.MatchString(cell) {
			continue
		}

		// detect and remove lines without value
		empty := true
		for _, cell := range row[1:] {
			if cell != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}

		log.Printf("Row %d: %#v", i, row)

		bond := Bond{
			Name:     row[0],
			Security: row[1],
			Risk: BondRisk{
				StandardPoor: row[9],
				Moody:        row[10],
				Fitch:        row[11],
			},
			Country: row[14],
			Code:    row[15],
		}

		if bond.Coupon, err = strconv.ParseFloat(row[2], 64); err != nil {
			return nil, errors.Wrapf(err, "failed to parse coupon on line %d", i)
		}

		if row[3] != "n.a." {
			maturity, err := time.Parse("1/2/2006", row[3])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse maturity on line %d", i)
			}
			bond.Maturity = &maturity
		}

		if bond.LastPrice, err = strconv.ParseFloat(row[5], 64); err != nil {
			return nil, errors.Wrapf(err, "failed to parse last price on line %d", i)
		}

		if bond.Yield, err = strconv.ParseFloat(row[6], 64); err != nil {
			return nil, errors.Wrapf(err, "failed to parse yield on line %d", i)
		}

		if bond.Duration, err = strconv.ParseFloat(row[7], 64); err != nil {
			return nil, errors.Wrapf(err, "failed to parse duration on line %d", i)
		}

		if row[8] != "n.a." {
			if bond.YearsToMaturity, err = strconv.ParseFloat(row[8], 64); err != nil {
				return nil, errors.Wrapf(err, "failed to parse years to maturity on line %d", i)
			}
		}

		if bond.MinimumPiece, err = strconv.ParseFloat(row[13], 64); err != nil {
			return nil, errors.Wrapf(err, "failed to parse minimum piece on line %d", i)
		}

		normalized = append(normalized, bond)
	}

	return normalized, nil
}

func filterBonds(bonds Bonds) []Bond {
	filtered := make(Bonds, 0, len(bonds))

	for _, bond := range bonds {
		// remove bonds with low coupon
		if bond.Coupon < 5 {
			continue
		}

		// remove bonds with no date or more than 6 years of maturity
		if bond.Maturity == nil || bond.Maturity.After(time.Now().Add(time.Hour*24*365*6)) {
			continue
		}

		// remove bonds with low price or too expensive
		if bond.LastPrice < 95 || bond.LastPrice > 101 {
			continue
		}

		filtered = append(filtered, bond)
	}

	return filtered
}

func main() {
	lambda.Start(HandleRequest)
}

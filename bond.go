package xpbonds

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// BondReport contains the location of the bond rates.
type BondReport struct {
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

// Interesting returns if the bond is interesting according to some predefined
// rules.
func (b Bond) Interesting() bool {
	// remove bonds with low coupon
	if b.Coupon < 5 {
		return false
	}

	// remove bonds with no date or more than 6 years of maturity
	if b.Maturity == nil || b.Maturity.After(time.Now().Add(time.Hour*24*365*6)) {
		return false
	}

	// remove bonds with low price or too expensive
	if b.LastPrice < 95 || b.LastPrice > 101 {
		return false
	}

	return true
}

func parseBond(row []string) (Bond, error) {
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

	var err error

	if bond.Coupon, err = strconv.ParseFloat(row[2], 64); err != nil {
		return bond, errors.Wrap(err, "failed to parse coupon")
	}

	if row[3] != "n.a." {
		maturity, err := time.Parse("1/2/2006", row[3])
		if err != nil {
			return bond, errors.Wrap(err, "failed to parse maturity")
		}
		bond.Maturity = &maturity
	}

	if bond.LastPrice, err = strconv.ParseFloat(row[5], 64); err != nil {
		return bond, errors.Wrap(err, "failed to parse last price")
	}

	if bond.Yield, err = strconv.ParseFloat(row[6], 64); err != nil {
		return bond, errors.Wrap(err, "failed to parse yield")
	}

	if bond.Duration, err = strconv.ParseFloat(row[7], 64); err != nil {
		return bond, errors.Wrap(err, "failed to parse duration")
	}

	if row[8] != "n.a." {
		if bond.YearsToMaturity, err = strconv.ParseFloat(row[8], 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse years to maturity")
		}
	}

	if bond.MinimumPiece, err = strconv.ParseFloat(row[13], 64); err != nil {
		return bond, errors.Wrap(err, "failed to parse minimum piece")
	}

	return bond, nil
}

// BondRisk determinates the risk according to different entities.
type BondRisk struct {
	StandardPoor string `json:"standardPoor"`
	Moody        string `json:"moody"`
	Fitch        string `json:"fitch"`
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

// Filter detect the most interesting bonds according to some predefined rules.
func (b Bonds) Filter() Bonds {
	filtered := make(Bonds, 0, len(b))
	for _, bond := range b {
		if bond.Interesting() {
			filtered = append(filtered, bond)
		}
	}
	return filtered
}

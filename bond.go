package xpbonds

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// DateFormat defines all acceptable date formats used when parsing the report.
type DateFormat string

// List of available date formats.
const (
	DateFormatDDMMYYYY = "dd/mm/yyyy"
	DateFormatMMDDYYYY = "mm/dd/yyyy"
)

// UnmarshalJSON parse the date format input value. It will return an error if
// the date format isn't acceptable.
func (d *DateFormat) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	dataStr = strings.TrimSpace(dataStr)
	dataStr = strings.ToLower(dataStr)

	switch dataStr {
	case string(DateFormatDDMMYYYY):
		*d = DateFormatDDMMYYYY
	case string(DateFormatMMDDYYYY):
		*d = DateFormatMMDDYYYY
	}

	return errors.Errorf("invalid date format '%s'", dataStr)
}

// BondReport contains all bonds data to be analyzed.
type BondReport struct {
	XLXSReport string     `json:"xlsxReport"`
	DateFormat DateFormat `json:"dateFormat"`
}

// Bond contains the bond descriptions.
type Bond struct {
	Name            string     `json:"name"`
	Security        string     `json:"security"`
	Coupon          float64    `json:"coupon"`
	Yield           float64    `json:"yield"`
	Maturity        *time.Time `json:"maturity"`
	LastPrice       float64    `json:"lastPrice"`
	CurrentPrice    *float64   `json:"currentPrice"`
	CurrentPriceURL *string    `json:"currentPriceURL"`
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

func parseBond(row row, dateFormat DateFormat) (Bond, error) {
	bond := Bond{
		Name:     row.get(0),
		Security: row.get(1),
		Risk: BondRisk{
			StandardPoor: row.get(9),
			Moody:        row.get(10),
			Fitch:        row.get(11),
		},
		Country: row.get(14),
		Code:    row.get(15),
	}

	var err error

	if coupon := row.get(2); coupon != "" {
		if bond.Coupon, err = strconv.ParseFloat(coupon, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse coupon")
		}
	}

	if maturity := row.get(3); maturity != "" {
		m, err := parseTime(maturity, dateFormat)
		if err != nil {
			return bond, errors.Wrap(err, "failed to parse maturity")
		}
		bond.Maturity = &m
	}

	if lastPrice := row.get(5); lastPrice != "" {
		if bond.LastPrice, err = strconv.ParseFloat(lastPrice, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse last price")
		}
	}

	if yield := row.get(6); yield != "" {
		if bond.Yield, err = strconv.ParseFloat(yield, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse yield")
		}
	}

	if duration := row.get(7); duration != "" {
		if bond.Duration, err = strconv.ParseFloat(duration, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse duration")
		}
	}

	if yearsToMaturity := row.get(8); yearsToMaturity != "" {
		if bond.YearsToMaturity, err = strconv.ParseFloat(yearsToMaturity, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse years to maturity")
		}
	}

	if minimumPiece := row.get(13); minimumPiece != "" {
		if bond.MinimumPiece, err = strconv.ParseFloat(minimumPiece, 64); err != nil {
			return bond, errors.Wrap(err, "failed to parse minimum piece")
		}
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

// FillCurrentPrice looks for the current price from the given bonds.
func (b Bonds) FillCurrentPrice() {
	var wg sync.WaitGroup
	for i := range b {
		wg.Add(1)
		go func(bond *Bond) {
			if err := fillCurrentBondPrice(bond); err != nil {
				log.Printf("failed to retrieve the last bond price: %s", err)
			}
			wg.Done()
		}(&b[i])
	}
	wg.Wait()
}

func parseTime(value string, dateFormat DateFormat) (time.Time, error) {
	var format string
	switch dateFormat {
	case DateFormatDDMMYYYY:
		format = "2/1/2006"
	default: // DateFormatMMDDYYYY
		format = "1/2/2006"
	}

	t, err := time.Parse(format, value)
	if err != nil {
		return t, errors.Wrap(err, "failed to parse time")
	}
	return t, nil
}

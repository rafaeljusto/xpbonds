package xpbonds

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var (
	reBusinessInsiderBondID    = regexp.MustCompile(`^.*new Array\(new Array\(.*"(.*?)\|.*$`)
	reBusinessInsiderBondPrice = regexp.MustCompile(`.*>\s+(\d+\.\d{2}).*`)
)

// fillCurrentBondPrice tries to retrieve the bond price from the Business
// Insider webpage.
func fillCurrentBondPrice(bond *Bond) error {
	url, err := findBusinessInsiderURL(bond.Code)
	if err != nil {
		return errors.Wrapf(err, "failed to get bond %s url", bond.Code)
	}
	bond.CurrentPriceURL = &url

	response, err := client.Get(url)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve bond %s details", bond.Code)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to read response body for bond %s", bond.Code)
	}

	ioutil.WriteFile("/tmp/"+bond.Code+".tmp", body, os.ModePerm)
	matches := reBusinessInsiderBondPrice.FindStringSubmatch(string(body))
	if len(matches) <= 1 {
		return errors.Errorf("no bond %s price found", bond.Code)
	}

	price, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return errors.Wrapf(err, "failed to parse bond %s price", bond.Code)
	}
	bond.CurrentPrice = &price

	return nil
}

func findBusinessInsiderURL(code string) (string, error) {
	url := fmt.Sprintf("https://markets.businessinsider.com/ajax/SearchController_Suggest"+
		"?max_results=25"+
		"&Keywords_mode=APPROX"+
		"&Keywords=%[1]s"+
		"&query=%[1]s"+
		"&bias=100"+
		"&target_id=0", code)

	response, err := client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve market URL")
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	matches := reBusinessInsiderBondID.FindStringSubmatch(string(body))
	if len(matches) <= 1 {
		return "", errors.Errorf("no identifier found for bond")
	}

	return fmt.Sprintf("https://markets.businessinsider.com/bonds/%s", matches[1]), nil
}

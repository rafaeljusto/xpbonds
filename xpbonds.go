package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

// HandleRequest determinates the best bonds from the given bond rates. It
// downloads the bond rates from the given location, converts it from PDF to
// Excel format and perform some sorting actions to determinate the best bond.
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Enable CORS
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS,POST",
		"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
	}
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
		}, nil

	} else if request.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Headers:    headers,
		}, nil
	}

	var rates BondRates
	if err := json.Unmarshal([]byte(request.Body), &rates); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
		}, errors.Wrap(err, "failed to parse body")
	}

	log.Println("Report:", rates.Location)

	response, err := http.Get(rates.Location)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
		}, errors.Wrap(err, "failed to get rates")
	}
	defer response.Body.Close()

	excel, err := pdfToExcel(response.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
		}, errors.Wrap(err, "failed to convert from PDF to excel")
	}

	bonds, err := parseExcel(excel)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
		}, errors.Wrap(err, "failed to parse excel")
	}

	bonds = bonds.Filter()
	sort.Sort(bonds)

	bondsRaw, err := json.Marshal(bonds)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
		}, errors.Wrap(err, "failed to encode bonds")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       string(bondsRaw),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

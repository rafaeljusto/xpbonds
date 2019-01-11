package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/rafaeljusto/xpbonds"
)

// HandleRequest adds serveless approach for determinating the best bonds.
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// enable CORS
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

	// limit upload file to 500kb
	if err := r.ParseMultipartForm(32 << 14); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
		}, errors.Wrap(err, "failed to setup upload")
	}

	report, _, err := r.FormFile("report")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
		}, errors.Wrap(err, "failed to read upload")
	}
	defer report.Close()

	bonds, err := xpbonds.FindBestBonds(ctx, rates)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
		}, errors.Wrap(err, "failed to find the best bonds")
	}

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

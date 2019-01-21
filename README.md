XP Bonds
========

[![GoDoc](https://godoc.org/github.com/rafaeljusto/xpbonds?status.png)](https://godoc.org/github.com/rafaeljusto/xpbonds)
[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/rafaeljusto/xpbonds/master/LICENSE)

[Serverless solution](https://github.com/rafaeljusto/xpbonds/blob/master/cmd/xpbonds-serveless/main.go) to parse and analyze bond reports from [XP Investments](https://www.xpi.us). It was built to run in [AWS Lambda service](https://aws.amazon.com/lambda/). There's also a [common HTTP server solution](https://github.com/rafaeljusto/xpbonds/blob/master/cmd/xpbonds/main.go) that can be used in other environments.

How does it works?
------------------

The service receives the Focused Lists report from XP Investments, in Excel
format (xlsx), like the bellow.

![XP Investments Report Example](https://github.com/rafaeljusto/xpbonds/raw/master/xpbonds.png "XP Investments Report Example")

The Excel report will be analyzed filtering undesired bonds with the following
rules:

* Minimum coupon rate
* Maturity
* Price

The resulted bonds will be sorted by coupon. Current market price will also be
retrieved to compare with the report (when available).

Serveless Setup
---------------

```shell
% go get -u github.com/rafaeljusto/xpbonds/...
% cd $GOPATH/github.com/rafaeljusto/xpbonds/cmd/xpbonds-serveless
% GOOS=linux go build -o xpbonds *.go && zip xpbonds.zip ./xpbonds
% aws lambda create-function \
  --region <region> \
  --function-name xpbonds \
  --memory 512 \
  --role <arn:aws:iam::account-id:role/execution_role> \
  --runtime go1.x \
  --zip-file fileb://$GOPATH/github.com/rafaeljusto/xpbonds/xpbonds.zip \
  --handler xpbonds
```

Serveless Protocol
------------------

The JSON that the service is expecting a [events.APIGatewayProxyRequest](https://godoc.org/github.com/aws/aws-lambda-go/events#APIGatewayProxyRequest), where the method should be `POST` and the body should be:

```json
{
  "xlsxReport": "EsDBBQABgAIAAAAIQBG8ICPdQEAAD...BQYAAAAADwAPAN4DAABChAEAAAA=",
  "dateFormat": "MM/DD/YYYY",
  "minCoupon": 5,
  "maxMaturity": 6,
  "minPrice": 95,
  "maxPrice": 101,
  "minPiece": 0,
  "maxPiece": 400000
}
```

Where:
* **xlsxReport** is the XLSX report encoded in base64;
* **dateFormat** defines the date format used in the report, possible values are
  `MM/DD/YYYY` or `DD/MM/YYYY`;
* **minCoupon** will filter bonds that have a coupon rate (%) less than the
  provided;
* **maxMaturity** defined in years also will filter bonds with an expiration
  date too far away;
* **minPrice** and **maxPrice** defines a range of acceptable market prices for
  the bonds;
* **minPiece** and **maxPiece** defines a range of acceptable minimum piece of a
  bond. The informed values will match the minimum piece in the current
  currency.

CORS is enable to make it easy for cross-domain requests. The response will be a [events.APIGatewayProxyResponse](https://godoc.org/github.com/aws/aws-lambda-go/events#APIGatewayProxyResponse), where the body will contain a list of bonds in the following format:

```json
[
  {
    "name": "Gol Linhas Aereas Inteligentes",
    "security": "GOLLBZ 8 7/8 01/24/22",
    "coupon": 8.875,
    "yield": 8.6,
    "maturity": "2022-01-24T00:00:00Z",
    "lastPrice": 100.4,
    "duration": 1.8,
    "yearsToMaturity": 3.1,
    "minimumPiece": 200000,
    "country": "BR",
    "risk": {
      "standardPoor": "B-",
      "moody": "n.a.",
      "fitch": "B"
    },
    "code": "USL4441PAA86"
  }
]
```

**PS:** `maturity` could be null when the bond has no end date.

User Interface
--------------

There's a simple user web interface to easy retrieve the best bonds. If you decide using it, please remember to replace an [internal URL for your production server](https://github.com/rafaeljusto/xpbonds/blob/a0b793e022fc16926240c9f8eaa4ae531d41c553/ui/index.html#L177).

![User Interface](https://github.com/rafaeljusto/xpbonds/raw/master/xpbonds-ui.png "User Interface")
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/pkg/errors"
)

var client http.Client

// uploadResponse is the response returned after uploading a new file. This is
// specific for the pdftoexcel service.
type uploadResponse struct {
	JobID string `json:"jobId"`
	Error string `json:"error"`
}

// checkResponse is the response returned while checking if the PDF was already
// converted.
type checkResponse struct {
	Status      string `json:"status"`
	DownloadURL string `json:"download_url"`
}

func pdfToExcel(pdf io.Reader) (string, error) {
	jobID, err := uploadPDF(pdf)
	if err != nil {
		return "", errors.Wrap(err, "failed to upload PDF")
	}
	log.Printf("Job ID %s retrieved successfully", jobID)

	var downloadURL string
	for {
		time.Sleep(time.Second)

		if downloadURL, err = checkConversion(jobID); err != nil {
			return "", errors.Wrap(err, "failed to check conversion")
		} else if downloadURL != "" {
			break
		}

		log.Print("Conversion not done yet, waiting...")
	}

	log.Printf("Downloading excel file at %s", downloadURL)
	excel, err := downloadExcel(downloadURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to download excel")
	}

	return excel, nil
}

func uploadPDF(pdf io.Reader) (string, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="Filedata"; filename="XPBonds.pdf"`)
	h.Set("Content-Type", "application/pdf")

	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)

	part, err := w.CreatePart(h)
	if err != nil {
		return "", errors.Wrap(err, "failed to create form field")
	}

	if _, err := io.Copy(part, pdf); err != nil {
		return "", errors.Wrap(err, "failed to copy pdf content to part")
	}

	if err := w.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close multipart")
	}

	request, err := http.NewRequest("POST", "https://www.pdftoexcel.com/upload.instant.php", body)
	if err != nil {
		return "", errors.Wrap(err, "failed to build request")
	}
	request.Header.Set("Content-Type", w.FormDataContentType())
	request.Header.Set("Host", "www.pdftoexcel.com")
	request.Header.Set("Origin", "https://www.pdftoexcel.com")
	request.Header.Set("Referer", "https://www.pdftoexcel.com/")

	response, err := client.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("unexpected error while uploading PDF")
	}

	var uploadResponse uploadResponse
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&uploadResponse); err != nil {
		return "", errors.Wrap(err, "failed to parse uploaded response")
	}

	if uploadResponse.Error != "" {
		return "", errors.New(uploadResponse.Error)
	}

	return uploadResponse.JobID, nil
}

func checkConversion(jobID string) (string, error) {
	url := fmt.Sprintf("https://www.pdftoexcel.com/getIsConverted.php?jobId=%s&rand=%d", jobID, rand.Intn(10))
	response, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("unexpected error while uploading PDF")
	}

	var checkResponse checkResponse
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&checkResponse); err != nil {
		return "", errors.Wrap(err, "failed to parse check response")
	}

	return checkResponse.DownloadURL, nil
}

func downloadExcel(downloadURL string) (string, error) {
	url := fmt.Sprintf("https://www.pdftoexcel.com%s", downloadURL)
	response, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}
	defer response.Body.Close()

	f, err := ioutil.TempFile("", "xpbonds-")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temporary file")
	}
	defer f.Close()

	if _, err := io.Copy(f, response.Body); err != nil {
		return "", errors.Wrap(err, "failed to copy response content")
	}

	return f.Name(), nil
}

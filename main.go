package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	client "github.com/influxdata/influxdb1-client"
	"github.com/tkanos/gonfig"
)

// Configuration - config.json
type Configuration struct {
	InfluxUsername    string
	InfluxPassword    string
	InfluxHost        string
	InfluxDatabase    string
	InfluxMeasurement string
}

//Open311Response - JSON response
type Open311Response []struct {
	ServiceRequestID  string  `json:"service_request_id"`
	Status            string  `json:"status"`
	ServiceName       string  `json:"service_name"`
	ServiceCode       string  `json:"service_code"`
	AgencyResponsible string  `json:"agency_responsible"`
	Description       string  `json:"description,omitempty"`
	RequestedDatetime string  `json:"requested_datetime"`
	UpdatedDatetime   string  `json:"updated_datetime"`
	Address           string  `json:"address"`
	Lat               float64 `json:"lat"`
	Long              float64 `json:"long"`
	StatusNotes       string  `json:"status_notes,omitempty"`
}

const configFile = "config.json"

const endPoint = "http://311.austintexas.gov/open311/v2/requests.json"

func getRequests() Open311Response {

	resp, err := http.Get(endPoint)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		panic(resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var data Open311Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}

	return data
}

func readConfig(file string) Configuration {
	configuration := Configuration{}
	err := gonfig.GetConf(file, &configuration)
	if err != nil {
		panic(err)
	}
	return configuration
}

func establishInfluxConnection(username string, password string, databaseURL string) *client.Client {
	host, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	conf := client.Config{
		URL:      *host,
		Username: username,
		Password: password,
	}
	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	return con
}

func readWriteInflux() {
	configuration := readConfig(configFile)
	con := establishInfluxConnection(configuration.InfluxUsername, configuration.InfluxPassword, configuration.InfluxHost)
	requests := getRequests()
	pts := make([]client.Point, len(requests))

	for i, request := range requests {
		println(request.ServiceRequestID)

		requestedTime, err := time.Parse(time.RFC3339, request.RequestedDatetime)
		if err != nil {
			fmt.Println(err)
		}

		pts[i] = client.Point{
			Measurement: configuration.InfluxMeasurement,
			Tags: map[string]string{
				"service_request_id": request.ServiceRequestID,
				"service_code":       request.ServiceCode,
				"service_name":       request.ServiceName,
				"agency_responsible": request.AgencyResponsible,
				"status":             request.Status,
			},
			Fields: map[string]interface{}{
				"status_notes": request.StatusNotes,
				"description":  request.Description,
				"updated":      request.UpdatedDatetime,
				"address":      request.Address,
				"lat":          request.Lat,
				"long":         request.Long,
			},
			Time:      requestedTime,
			Precision: "s",
		}

	}

	bps := client.BatchPoints{
		Points:   pts,
		Database: configuration.InfluxDatabase,
	}

	_, err := con.Write(bps)
	if err != nil {
		log.Fatal(err)
	}
}

//Handler - AWS Lambda function
func Handler(ctx context.Context, evt json.RawMessage) (events.APIGatewayProxyResponse, error) {
	readWriteInflux()
	returnVal := events.APIGatewayProxyResponse{
		Body:       "Completed request successfully",
		StatusCode: 200,
	}
	return returnVal, nil
}

func main() {
	lambda.Start(Handler)
}

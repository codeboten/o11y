package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	owm "github.com/briandowns/openweathermap"
	"github.com/honeycombio/beeline-go"
)

type weatherRequestEvent struct {
	City   string `json:"city"`
	Planet string `json:"planet"`
}

type planetaryAPIBody struct {
	Body weatheraryResponse `json:"body"`
}

type weatheraryResponse struct {
	Planet  string `json:"planet"`
	Weather string `json:"weather"`
}

func getPlanetaryWeather(ctx context.Context, event weatherRequestEvent) string {
	ctx, span := beeline.StartSpan(ctx, "getPlanetaryWeather")
	//apiKey := os.Getenv("WEATHERARY_API_KEY")
	apiURL := os.Getenv("PLANETARY_API_URL")
	span.AddField("url", apiURL)
	url := fmt.Sprintf("%s/%s", apiURL, event.Planet)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return getErrorMessage(ctx, err)
	}

	req.Header.Set("X-Honeycomb-Trace", span.SerializeHeaders())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		span.AddField("error", err)
		return getErrorMessage(ctx, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		span.AddField("upstreamError", res.StatusCode)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		message := string(bodyBytes)
		span.AddField("upstreamBody", message)
		return getErrorMessage(ctx, fmt.Errorf("%d - %s", res.StatusCode, message))
	}

	r := new(weatheraryResponse)
	err = json.NewDecoder(res.Body).Decode(r)
	span.AddField("planetary", res)

	if err != nil {
		return getErrorMessage(ctx, err)
	}

	defer span.Send()
	return r.Weather
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event weatherRequestEvent) (Response, error) {
	ctx, span := beeline.StartSpan(ctx, "Handler")
	defer span.Send()
	span.AddField("city", event.City)

	var result string
	if len(event.Planet) > 0 {
		result = getPlanetaryWeather(ctx, event)
	} else {
		result = getWeather(ctx, event.City)
	}

	var buf bytes.Buffer
	body, err := json.Marshal(map[string]interface{}{
		"city":    event.City,
		"weather": result,
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func getErrorMessage(ctx context.Context, err error) string {
	_, span := beeline.StartSpan(ctx, "getError")
	span.AddField("error", err.Error())
	defer span.Send()
	return err.Error()
}

func getWeather(ctx context.Context, city string) string {
	var apiKey = os.Getenv("OWM_API_KEY")
	ctx, span := beeline.StartSpan(ctx, "getWeather")
	defer span.Send()

	w, err := owm.NewCurrent("C", "en", apiKey)
	if err != nil {
		return getErrorMessage(ctx, err)
	}

	w.CurrentByName(city)
	if len(w.Weather) == 0 {
		return getErrorMessage(ctx, errors.New("City not found"))
	}

	result := w.Weather[0].Description
	span.AddField("weather", result)
	return result
}

func main() {
	beeline.Init(beeline.Config{
		WriteKey: os.Getenv("HONEYCOMB_KEY"),
		Dataset:  os.Getenv("HONEYCOMB_DATASET"),
	})
	lambda.Start(HoneycombMiddleware(Handler))
}

func addRequestProperties(ctx context.Context) {
	// Add a variety of details about the lambda request
	ctx, span := beeline.StartSpan(ctx, "addRequestProperties")
	defer span.Send()
	span.AddField("function_name", lambdacontext.FunctionName)
	span.AddField("function_version", lambdacontext.FunctionVersion)
}

// HoneycombMiddleware will wrap our lambda handle funcs to create
// trace for events
func HoneycombMiddleware(fn func(ctx context.Context, event weatherRequestEvent) (Response, error)) func(ctx context.Context, event weatherRequestEvent) (Response, error) {
	return func(ctx context.Context, event weatherRequestEvent) (Response, error) {
		startHandler := time.Now()

		ctx, span := beeline.StartSpan(ctx, "HoneycombMiddleware")
		span.AddTraceField("application", "intergalactic-weatherary")
		span.AddTraceField("platform", "aws")
		defer span.Send()

		addRequestProperties(ctx)

		// don't forget to send the events
		defer beeline.Flush(ctx)

		resp, err := fn(ctx, event)
		if err != nil {
			span.AddField("lambda.error", err)
		}

		span.AddField("response.status_code", resp.StatusCode)
		handlerDuration := time.Since(startHandler)
		span.AddField("timers.total_time_ms", handlerDuration/time.Millisecond)
		return resp, err
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	owm "github.com/briandowns/openweathermap"
	"github.com/honeycombio/beeline-go"
	libhoney "github.com/honeycombio/libhoney-go"
)

var (
	honeycombKey = "honeycombEvent"
)

type weatherRequestEvent struct {
	City string `json:"city"`
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event weatherRequestEvent) (Response, error) {
	// Ensure events are sent before returning
	ev := ctx.Value(honeycombKey).(*libhoney.Event)

	// Measure execution time
	// startTime := time.Now()

	ev.AddField("city", event.City)

	result := getWeather(ev, event.City)

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
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func getWeather(ev *libhoney.Event, city string) string {
	var apiKey = os.Getenv("OWM_API_KEY")

	w, err := owm.NewCurrent("C", "en", apiKey)
	if err != nil {
		ev.AddField("error", err)
		return "unavailable"
	}

	w.CurrentByName(city)
	result := w.Weather[0].Description
	ev.AddField("weather", result)

	return result
}

func main() {
	beeline.Init(beeline.Config{
		WriteKey: os.Getenv("HONEYCOMB_KEY"),
		Dataset:  os.Getenv("HONEYCOMB_DATASET"),
	})
	lambda.Start(HoneycombMiddleware(Handler))
}

func addRequestProps(ctx context.Context, ev *libhoney.Event) {
	// Add a variety of details about the HTTP request, such as user agent
	// and method, to any created libhoney event.
	ev.AddField("function_name", lambdacontext.FunctionName)
	ev.AddField("function_version", lambdacontext.FunctionVersion)
}

// HoneycombMiddleware will wrap our HTTP handle funcs to automatically
// generate an event-per-request and set properties on them.
func HoneycombMiddleware(fn func(ctx context.Context, event weatherRequestEvent) (Response, error)) func(ctx context.Context, event weatherRequestEvent) (Response, error) {
	return func(ctx context.Context, event weatherRequestEvent) (Response, error) {
		// We'll time each HTTP request and add that as a property to
		// the sent Honeycomb event, so start the timer for that.
		startHandler := time.Now()
		ev := libhoney.NewEvent()

		defer func() {
			if err := ev.Send(); err != nil {
				log.Print("Error sending libhoney event: ", err)
			}
		}()

		addRequestProps(ctx, ev)

		// Create a context where we will store the libhoney event. We
		// will add default values to this event for every HTTP
		// request, and the user can access it to add their own
		// (powerful, custom) fields.
		newContext := context.WithValue(ctx, honeycombKey, ev)

		resp, err := fn(newContext, event)
		if err != nil {
			ev.AddField("lambda.error", err)
		}

		ev.AddField("response.status_code", resp.StatusCode)
		handlerDuration := time.Since(startHandler)
		ev.AddField("timers.total_time_ms", handlerDuration/time.Millisecond)
		return resp, err
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	beeline "github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	_, span := beeline.StartSpan(ctx, "Handler")
	span.AddTraceField("application", "planetary-api")
	defer span.Send()

	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"planet":  "myplanet",
		"weather": "fine",
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

func main() {
	beeline.Init(beeline.Config{
		WriteKey: os.Getenv("HONEYCOMB_KEY"),
		Dataset:  os.Getenv("HONEYCOMB_DATASET"),
	})
	lambda.Start(HoneycombMiddleware("planetary-api", Handler))
}

// HoneycombMiddleware will wrap our lambda handle funcs to create
// trace for events
func HoneycombMiddleware(appName string, fn func(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error)) func(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
		startHandler := time.Now()
		ctx, span := beeline.StartSpan(ctx, "HoneycombMiddleware")
		if len(request.Headers["x-honeycomb-trace"]) > 0 {
			ct, tr := trace.NewTrace(ctx, request.Headers["x-honeycomb-trace"])
			span = tr.GetRootSpan()
			span.AddField("name", "HoneycombMiddleware")
			ctx = ct
		}
		span.AddTraceField("application", appName)
		span.AddTraceField("platform", "aws")
		defer span.Send()

		time.Sleep(500 * time.Millisecond)
		// addRequestProperties(ctx)

		// don't forget to send the events
		defer beeline.Flush(ctx)

		resp, err := fn(ctx, request)
		if err != nil {
			span.AddField("lambda.error", err)
		}

		span.AddField("response.status_code", resp.StatusCode)
		handlerDuration := time.Since(startHandler)
		span.AddField("timers.total_time_ms", handlerDuration/time.Millisecond)
		return resp, err
	}
}

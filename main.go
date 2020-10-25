package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"time"

	"encoding/json"
	"fmt"
	"os"
)

type getItemsRequest struct {
	SortBy     string
	SortOrder  string
	ItemsToGet int
}

type getItemsResponseError struct {
	Message string `json:"message"`
}

type getItemsResponseData struct {
	Item string `json:"item"`
}

type getItemsResponseBody struct {
	Result string                 `json:"result"`
	Data   []getItemsResponseData `json:"data"`
	Error  getItemsResponseError  `json:"error"`
}

type getItemsResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type getItemsResponse struct {
	StatusCode int                     `json:"statusCode"`
	Headers    getItemsResponseHeaders `json:"headers"`
	Body       getItemsResponseBody    `json:"body"`
}

func main() {
	// Create Lambda service client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String("us-west-2")})

	// Create a sample event - looks like https://github.com/aws/aws-lambda-go/blob/master/events/testdata/ses-lambda-event.json

	record := events.SimpleEmailRecord{
		EventVersion: "",
		EventSource:  "",
		SES: events.SimpleEmailService{
			Mail: events.SimpleEmailMessage{
				CommonHeaders:    events.SimpleEmailCommonHeaders{},
				Source:           "",
				Timestamp:        time.Time{},
				Destination:      nil,
				Headers:          nil,
				HeadersTruncated: false,
				MessageID:        "b2t449562tmdi2429vuj52k78qjus09rr8gr5301",
			},
			Receipt: events.SimpleEmailReceipt{},
		},
	}
	request := events.SimpleEmailEvent{
		Records: []events.SimpleEmailRecord{record},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling SES request")
		os.Exit(0)
	}

	_, err = client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("ynab-email-parser"), Payload: payload})
	if err != nil {
		fmt.Printf("Error calling YNAB email parser: %v", err)
		os.Exit(0)
	}

}

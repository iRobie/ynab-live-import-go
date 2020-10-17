package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"go.bmvs.io/ynab"
	"go.bmvs.io/ynab/api"
	ynabaccount "go.bmvs.io/ynab/api/account"
	ynabbudget "go.bmvs.io/ynab/api/budget"
	ynabtransaction "go.bmvs.io/ynab/api/transaction"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type credentials struct {
	accessToken string
	bucketName  string
	tableName   string
	slackURL    string
}

type budgetAccount struct {
	budgetID string
	account  *ynabaccount.Account
}

type Transaction struct {
	MessageID  string `json:"messageID"`
	LastDigits int
	Date       string
	Amount     float32
	Merchant   string
}

const region = "us-west-2"

func getCredentials() (creds credentials, err error) {
	// check if variable is already setup
	accessToken := os.Getenv("ACCESS_TOKEN")

	if accessToken == "" {
		// Probably running locally. Load variables from .env file instead.
		err = godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file")
			return
		}
		accessToken = os.Getenv("ACCESS_TOKEN")
	}
	creds = credentials{
		accessToken: accessToken,
		bucketName:  os.Getenv("BUCKET_NAME"),
		tableName:   os.Getenv("TABLE_NAME"),
		slackURL:    os.Getenv("SLACK_URL"),
	}
	return

}

func main() {
	lambda.Start(HandleLambdaEvent)
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

func notifyError(message string, err error) {
	creds, _ := getCredentials()
	errorString := fmt.Sprintf("Error from YNAB Importer - %s: %s", message, err.Error())
	log.Print(errorString)
	serr := SendSlackNotification(creds.slackURL, errorString)
	if serr != nil {
		log.Fatal(serr)
	}
}

func HandleLambdaEvent(event events.DynamoDBEvent) error {

	credentials, err := getCredentials()
	if err != nil {
		notifyError("Missing ynab access token", err)
		return err
	}
	client := ynab.NewClient(credentials.accessToken)

	budgets, err := client.Budget().GetBudgets()
	if err != nil {
		notifyError("Could not retreive list of budgets", err)
		return err
	}

	accounts, err := getAccounts(client, budgets)
	if err != nil {
		notifyError("Could not retreive list of accounts", err)
		return err
	}

	dynamoclient := getDynamoClient(region)

	for _, record := range event.Records {
		if record.EventName == string(events.DynamoDBOperationTypeInsert) {
			image := record.Change.NewImage
			dynamoTransaction, err := unmarshallDynamoRecord(image)
			if err != nil {
				notifyError("Failed unmashalling DynamoDB record", err)
				return err
			}

			budgetAccount, err := getAccountID(accounts, dynamoTransaction.LastDigits)
			if err != nil {
				notifyError("Could not find correct account", err)
				return err
			}

			payload, err := getPayload(dynamoTransaction, budgetAccount)
			if err != nil {
				notifyError("Error getting payload", err)
				return err
			}
			err = postTransactionToAccount(client, budgetAccount, payload)
			deleteS3 := true
			if err != nil {
				notifyError("Error posting payload", err)
				deleteS3 = false
			}
			// Still delete in case of failure. Dynamo reprocesses until it succeeds.
			err = deleteDynamoRecord(dynamoclient, dynamoTransaction.MessageID)
			if err != nil {
				notifyError("Could not delete record", err)
				return err
			}
			// Only delete from S3 if successfully posted transaction. I want to see failed messages.
			if deleteS3 {
				s3Client, err := createS3Client(region)
				if err != nil {
					notifyError("Could not create s3 client", err)
					return err
				}
				err = deleteS3Object(s3Client, dynamoTransaction.MessageID)
				if err != nil {
					notifyError("Could not delete s3 bucket object", err)
					return err
				}
			}
			return nil
		}

	}
	return nil
}

func unmarshallDynamoRecord(record map[string]events.DynamoDBAttributeValue) (recordTransaction Transaction, err error) {

	lastDigits, err := record["LastDigits"].Integer()
	if err != nil {
		log.Printf("Error getting LastDigits from DynamoDB record: %v", err)
		return
	}
	dynamoAmount, err := record["Amount"].Float()
	if err != nil {
		log.Printf("Error getting amount from DynamoDB record: %v", err)
		return
	}

	recordTransaction = Transaction{
		MessageID:  record["messageID"].String(),
		LastDigits: int(lastDigits),
		Date:       record["Date"].String(),
		Amount:     float32(dynamoAmount),
		Merchant:   record["Merchant"].String(),
	}
	return
}

func getAccounts(client ynab.ClientServicer, budgets []*ynabbudget.Summary) ([]budgetAccount, error) {
	var budgetAccounts []budgetAccount

	for _, budget := range budgets {
		accounts, err := client.Account().GetAccounts(budget.ID)
		if err != nil {
			log.Printf("Could not retreive list of accounts for budget: " + err.Error())
		}
		for _, account := range accounts {
			account := budgetAccount{account: account, budgetID: budget.ID}
			budgetAccounts = append(budgetAccounts, account)
		}
	}

	if len(budgetAccounts) == 0 {
		return budgetAccounts, errors.New("could not find any accounts")
	}
	return budgetAccounts, nil
}

func postTransactionToAccount(client ynab.ClientServicer, account budgetAccount, payloadTransaction ynabtransaction.PayloadTransaction) error {
	_, err := client.Transaction().CreateTransaction(account.budgetID, payloadTransaction)
	if err != nil {
		if strings.Contains(err.Error(), "date must not be in the future or over 5 years ago") {
			originalDate := payloadTransaction.Date
			newDate := originalDate.Add(time.Hour * -12)
			payloadTransaction.Date = api.Date{newDate}
			time.Sleep(time.Second * 10)
			_, retryerr := client.Transaction().CreateTransaction(account.budgetID, payloadTransaction)
			if retryerr != nil {
				notifyError("Failed to post ynab transaction with 12 hour difference", retryerr)
				return retryerr
			}
			return nil
		}
		return err
	}
	return nil
}

func getPayload(record Transaction, account budgetAccount) (payloadTransaction ynabtransaction.PayloadTransaction, err error) {

	date, err := api.DateFromString(record.Date)

	today := time.Now()
	if date.Before(today.AddDate(0, 0, -3)) || date.After(today) {
		log.Printf("Could not parse date. Using today's date.")
		date = api.Date{Time: today.Add(time.Hour * -6)}
	}

	if err != nil {
		return
	}

	amount := int64(record.Amount * -1000)
	payee := record.Merchant
	memo := "Imported via email"

	payloadTransaction = ynabtransaction.PayloadTransaction{
		AccountID: account.account.ID,
		Date:      date,
		Amount:    amount,
		Cleared:   "uncleared",
		PayeeName: &payee,
		Memo:      &memo,
	}

	return
}

// Checks accounts for the ID given. This is the last 4 digits of the CC passed in email.
// The 4 digits must be added to the notes section of YNAB.
func getAccountID(accounts []budgetAccount, digits int) (budgetAccount, error) {

	check := strconv.Itoa(digits)
	for _, ynabAccount := range accounts {
		if ynabAccount.account.Note != nil {
			if *ynabAccount.account.Note == check {
				return ynabAccount, nil
			}
		}

	}
	return budgetAccount{}, errors.New("could not find account matching the given digits")
}

func getDynamoClient(region string) *dynamodb.DynamoDB {
	config := &aws.Config{
		Region: aws.String(region),
	}

	sess := session.Must(session.NewSession(config))

	client := dynamodb.New(sess)
	return client
}

func createS3Client(region string) (*s3.S3, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	s3Client := s3.New(sess, aws.NewConfig().WithRegion(region))
	return s3Client, nil
}

func getDynamoKey(messageID string) (key map[string]*dynamodb.AttributeValue, err error) {
	type lookup struct {
		MessageID string `json:"messageID"`
	}

	messageKey := lookup{
		MessageID: messageID,
	}

	key, err = dynamodbattribute.MarshalMap(messageKey)
	return

}

// Check if this is a duplicate entry
func getDuplicate(client *dynamodb.DynamoDB, messageID string) error {

	creds, _ := getCredentials()
	key, err := getDynamoKey(messageID)
	if err != nil {
		log.Printf("Error getting key from messageid: %v", err)
		return err
	}

	consistentread := true

	input := &dynamodb.GetItemInput{
		Key:            key,
		TableName:      aws.String(creds.tableName),
		ConsistentRead: &consistentread,
	}

	result, err := client.GetItem(input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if len(result.Item) == 0 {
		// This record was not found which is expected result
		return nil
	}
	return fmt.Errorf("Duplicate record found")
}

// Delete record from DynamoDB
func deleteDynamoRecord(client *dynamodb.DynamoDB, messageID string) error {

	creds, _ := getCredentials()
	key, err := getDynamoKey(messageID)
	if err != nil {
		log.Printf("Error getting key from messageid: %v", err)
		return err
	}

	input := &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: aws.String(creds.tableName),
	}

	_, err = client.DeleteItem(input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Delete object from S3. Run here because S3 objects should remain unless transaction was posted successfully.
func deleteS3Object(s3Client *s3.S3, key string) error {

	creds, _ := getCredentials()
	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(creds.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("S3 DeleteObject failed: %s", err)
		return err
	}
	return nil
}

// SendSlackNotification will post to an 'Incoming Webook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func SendSlackNotification(webhookUrl string, msg string) error {

	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}

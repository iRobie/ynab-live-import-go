package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	chaseString    = "Your Single Transaction Alert from Chase"
	bofaString     = "Credit card transaction exceeds alert limit you set"
	fourDigitRegex string
	amountRegex    string
	merchantRegex  string
	dateRegex      string
	dateLayout     string
	setupDone      = false
)

type Transaction struct {
	MessageID  string `json:"messageID"`
	LastDigits int
	Date       string
	Amount     float32
	Merchant   string
}

func main() {
	// check if variable is already setup
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		// Probably running locally. Load variables from .env file instead.
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatalf("Error loading .env file")
		}
		bucket = os.Getenv("BUCKET_NAME")
	}

	table := os.Getenv("TABLE_NAME")

	if bucket == "" {
		panic("Missing s3 bucket name")
	}

	if table == "" {
		panic("Missing DynamoDB table name")
	}

	contents, err := retreiveMail()
	check(err)

	if strings.Contains(contents, chaseString) {
		setupChase()
	} else if strings.Contains(contents, bofaString) {
		setupBofA()
	}

	if setupDone == false {
		panic("Email does not match a parser")
	}

	transaction, err := parseEmail(contents)
	check(err)
	transaction.MessageID = "test"
	err = saveToDynamoDB(transaction, table)
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func retreiveMail() (string, error) {
	dat, err := ioutil.ReadFile("testemails/chaseEmail.txt")
	check(err)
	emailBody := string(dat)
	return emailBody, nil
}

func setupChase() {
	fourDigitRegex = "ending in (\\d+)"
	amountRegex = "A charge of \\(\\$USD\\) (\\d+\\.\\d+) at .* has been authorized on .* at"
	merchantRegex = "A charge of \\(\\$USD\\) \\d+\\.\\d+ at (.*) has been authorized on .* at"
	dateRegex = "A charge of \\(\\$USD\\) \\d+\\.\\d+ at .* has been authorized on (.*) at"
	dateLayout = "Jan 02, 2006"
	setupDone = true
}

func setupBofA() {
	fourDigitRegex = "ending in (\\d+)"
	amountRegex = "Amount: \\$(\\d+\\.\\d+)"
	merchantRegex = "Where: (.*)\\n"
	dateRegex = "Date: (.*)\\n"
	dateLayout = "January 02, 2006"
	setupDone = true
}

func parseEmail(contents string) (Transaction, error) {
	lastDigitsString, err := getLastDigits(contents)
	check(err)
	lastDigits, err := strconv.Atoi(lastDigitsString)
	check(err)
	date, err := getDate(contents)
	check(err)
	amountString, err := getSpendAmount(contents)
	check(err)
	amount, err := strconv.ParseFloat(amountString, 32)
	check(err)
	payee, err := getMerchant(contents)
	check(err)
	transaction := Transaction{
		LastDigits: lastDigits,
		Date:       date,
		Amount:     float32(amount),
		Merchant:   payee,
	}
	return transaction, nil
}

func extractInformation(contents, title, regex string, num int) (string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", errors.New("error compiling " + title + " regex")
	}
	match := re.FindStringSubmatch(contents)
	if match != nil {
		return match[num], nil
	}
	return "", errors.New("could not parse " + title + " regex")
}

func getLastDigits(contents string) (string, error) {
	return extractInformation(contents, "last four digits", fourDigitRegex, 1)
}

func getSpendAmount(contents string) (string, error) {
	return extractInformation(contents, "amount", amountRegex, 1)
}

func getMerchant(contents string) (string, error) {
	return extractInformation(contents, "merchant", merchantRegex, 1)
}

func getDate(contents string) (string, error) {
	date, _ := extractInformation(contents, "date", dateRegex, 1)
	return parseDate(date)
}

func parseDate(date string) (string, error) {
	newDateFormat := "2006-01-02"
	t, _ := time.Parse(dateLayout, date)
	return t.Format(newDateFormat), nil
}

func saveToDynamoDB(transaction Transaction, tableName string) error {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(transaction)
	if err != nil {
		fmt.Println("Got error marshalling new transaction:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
		//ConditionExpression:      aws.String("attribute_not_exists(messageID)"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully added transaction to table " + tableName)
	return nil
}

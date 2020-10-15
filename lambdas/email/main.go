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
	parsers        []Parser
	selectedParser Parser
	bucket         string
	table          string
)

type Transaction struct {
	MessageID  string `json:"messageID"`
	LastDigits int
	Date       string
	Amount     float32
	Merchant   string
}

type Parser struct {
	name             string
	validationString string
	fourDigitRegex   string
	amountRegex      string
	merchantRegex    string
	dateRegex        string
	dateLayout       string
}

func init() {
	// check if variable is already setup
	bucket = os.Getenv("BUCKET_NAME")
	if bucket == "" {
		// Probably running locally. Load variables from .env file instead.
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file")
		}
		bucket = os.Getenv("BUCKET_NAME")
	}

	table = os.Getenv("TABLE_NAME")

}

func main() {

	if bucket == "" {
		log.Fatal("Missing bucket name")
	}

	if table == "" {
		log.Fatal("Missing dynamoDB table name")
	}

	contents, err := retreiveMail()
	if err != nil {
		log.Fatalf("Could not retrieve mail: " + err.Error())
	}

	for _, parser := range parsers {
		if strings.Contains(contents, parser.validationString) {
			selectedParser = parser
		}
	}

	if selectedParser == (Parser{}) {
		log.Fatal("Email does not match a parser")
	}

	transaction, err := parseEmail(contents)
	if err != nil {
		log.Fatalf("Could not parse mail: " + err.Error())
	}

	transaction.MessageID = "test"
	err = saveToDynamoDB(transaction, table)
	if err != nil {
		log.Fatalf("Could not save record to DynamoDB: " + err.Error())
	}
}

func retreiveMail() (string, error) {
	dat, err := ioutil.ReadFile("testemails/chaseEmail.txt")
	if err != nil {
		return "", err
	}
	emailBody := string(dat)
	return emailBody, nil
}

func parseEmail(contents string) (Transaction, error) {
	lastDigitsString, err := getLastDigits(contents)
	if err != nil {
		return Transaction{}, err
	}

	lastDigits, err := strconv.Atoi(lastDigitsString)
	if err != nil {
		return Transaction{}, err
	}

	date, err := getDate(contents)
	if err != nil {
		return Transaction{}, err
	}

	amountString, err := getSpendAmount(contents)
	if err != nil {
		return Transaction{}, err
	}

	amount, err := strconv.ParseFloat(amountString, 32)
	if err != nil {
		return Transaction{}, err
	}

	payee, err := getMerchant(contents)
	if err != nil {
		return Transaction{}, err
	}

	transaction := Transaction{
		LastDigits: lastDigits,
		Date:       date,
		Amount:     float32(amount),
		Merchant:   payee,
	}
	return transaction, nil
}

func extractInformation(contents, title, regex string) (string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", errors.New("error compiling " + title + " regex")
	}
	match := re.FindStringSubmatch(contents)
	if match != nil {
		return match[1], nil
	}
	return "", errors.New("could not parse " + title + " regex")
}

func getLastDigits(contents string) (string, error) {
	return extractInformation(contents, "last four digits", selectedParser.fourDigitRegex)
}

func getSpendAmount(contents string) (string, error) {
	return extractInformation(contents, "amount", selectedParser.amountRegex)
}

func getMerchant(contents string) (string, error) {
	return extractInformation(contents, "merchant", selectedParser.merchantRegex)
}

func getDate(contents string) (string, error) {
	date, _ := extractInformation(contents, "date", selectedParser.dateRegex)
	return parseDate(date)
}

func parseDate(date string) (string, error) {
	newDateFormat := "2006-01-02"
	t, _ := time.Parse(selectedParser.dateLayout, date)
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

	os.Exit(0)

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

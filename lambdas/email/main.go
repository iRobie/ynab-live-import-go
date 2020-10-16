package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	parsers        []Parser
	selectedParser Parser
	bucket         string
	table          string
	s3Client       *s3.S3
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

	err := createS3Client("us-west-2")
	if err != nil {
		log.Fatalf("Error creating S3 client: " + err.Error())
	}

}

func main() {

	if bucket == "" {
		log.Fatal("Missing bucket name")
	}

	if table == "" {
		log.Fatal("Missing dynamoDB table name")
	}

	lambda.Start(HandleLambdaEvent)
}

func HandleLambdaEvent(event events.SimpleEmailEvent) error {

	for _, sesMail := range event.Records {

		//Retrieve message from S3
		mailbody, err := retreiveMail(sesMail.SES.Mail.MessageID)

		if err != nil {
			log.Print("Error retrieving mail")
			return err
		}

		for _, parser := range parsers {
			if strings.Contains(mailbody, parser.validationString) {
				selectedParser = parser
			}
		}

		if selectedParser == (Parser{}) {
			log.Print("Email does not match a parser")
			return fmt.Errorf("email does not match a parser")
		}

		transaction, err := parseEmail(mailbody)
		if err != nil {
			log.Printf("Could not parse mail: %s", err)
			return err
		}

		transaction.MessageID = sesMail.SES.Mail.MessageID
		err = saveToDynamoDB(transaction, table)
		if err != nil {
			log.Printf("Could not save record to DynamoDB: %s", err)
			return err
		}

	}
	return nil
}

func createS3Client(region string) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	s3Client = s3.New(sess, aws.NewConfig().WithRegion(region))
	return nil
}

func retreiveMail(messageid string) (string, error) {
	//Retrieve message from S3
	s3Mail, err := getFromS3(messageid)
	if err != nil {
		log.Printf("get message from S3 failed: %s", err)
		return "", err
	}

	defer s3Mail.Close()

	// parse the original message
	parsedMail, err := mail.ReadMessage(s3Mail)
	if err != nil {
		log.Printf("ReadMessage failed: %s", err)
		return "", err
	}

	// message is quoted printable
	body := parsedMail.Body
	b, err := ioutil.ReadAll(quotedprintable.NewReader(body))

	if err != nil {
		log.Printf("error decoding quoted printable mail: %s", err)
		return "", err
	}

	return string(b), err

}

func getFromS3(key string) (io.ReadCloser, error) {

	obj, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("S3 GetObject failed: %s", err)
		return nil, err
	}
	return obj.Body, nil
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
		log.Printf("error compiling " + title + " regex")
		return "", fmt.Errorf("error compiling " + title + " regex")
	}
	match := re.FindStringSubmatch(contents)
	if match != nil {
		return match[1], nil
	}
	log.Printf("could not parse " + title + " regex")
	return "", fmt.Errorf("could not parse " + title + " regex")
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
		log.Printf("Got error marshalling new transaction: %s", err.Error())
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
		//ConditionExpression:      aws.String("attribute_not_exists(messageID)"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %s", err.Error())
		return err
	}

	log.Println("Successfully added transaction to table " + tableName)
	return nil
}

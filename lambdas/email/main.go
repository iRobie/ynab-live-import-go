package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	chaseString    = "Your Single Transaction Alert from Chase"
	bofaString     = "Credit card transaction exceeds alert limit you set"
	fourDigitRegex string
	amountRegex    string
	merchantRegex  string
	dateRegex      string
	dateLayout     string
)

func main() {
	fmt.Println("hello world!")
}

func validateEmail(contents, validation string) error {
	if strings.Contains(contents, validation) {
		return nil
	}
	return errors.New("email does not contain check string")
}

func setupChase() {
	fourDigitRegex = "ending in (\\d+)"
	amountRegex = "A charge of \\(\\$USD\\) (\\d+\\.\\d+) at .* has been authorized on .* at"
	merchantRegex = "A charge of \\(\\$USD\\) \\d+\\.\\d+ at (.*) has been authorized on .* at"
	dateRegex = "A charge of \\(\\$USD\\) \\d+\\.\\d+ at .* has been authorized on (.*) at"
	dateLayout = "Jan 02, 2006"
}

func setupBofA() {
	fourDigitRegex = "ending in (\\d+)"
	amountRegex = "Amount: \\$(\\d+\\.\\d+)"
	merchantRegex = "Where: (.*)\\n"
	dateRegex = "Date: (.*)\\n"
	dateLayout = "January 02, 2006"
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
	return extractInformation(contents, "date", dateRegex, 1)
}

func parseDate(date string) (string, error) {
	newDateFormat := "2006-01-02"
	t, _ := time.Parse(dateLayout, date)
	return t.Format(newDateFormat), nil
}

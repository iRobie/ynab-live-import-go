package main

import (
	"github.com/aws/aws-lambda-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"time"
)

var _ = Describe("Parse Citi emails", func() {

	var (
		emailbody             string
		expectedTransaction   Transaction
		s3messageid           string
		s3expectedTransaction Transaction
		request               events.SimpleEmailEvent
	)

	BeforeEach(func() {
		dat, _ := ioutil.ReadFile("testemails/citiEmail.html")
		emailbody = string(dat)
		//selectedParser = citiParser()
		expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 2345,
			Date:       "2020-10-14",
			Amount:     12.34,
			Merchant:   "I AM A LARGE #MERCHANT",
		}
		s3messageid = "b2t449562tmdi2429vuj52k78qjus09rr8gr5301"

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
					MessageID:        s3messageid,
				},
				Receipt: events.SimpleEmailReceipt{},
			},
		}
		request = events.SimpleEmailEvent{
			Records: []events.SimpleEmailRecord{record},
		}

		s3expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 1234,
			Date:       "2020-10-13",
			Amount:     109.00,
			Merchant:   "Test Mer\\chant.com",
		}
		_ = s3messageid
		_ = s3expectedTransaction
		_ = emailbody
		_ = expectedTransaction
	})

	Context("When given an email body to parse, ", func() {

		It("parses the function correctly", func() {
			transaction, err := parseEmail(emailbody)
			Expect(err).To(BeNil())
			Expect(transaction).To(Equal(expectedTransaction))
		})

	})

	Context("When downloading a mail from S3, ", func() {

		It("parses the function correctly", func() {
			err := HandleLambdaEvent(request)
			Expect(err).To(BeNil())
		})
	})

})

package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("Parse Citi emails", func() {

	var (
		emailbody             string
		expectedTransaction   Transaction
		s3messageid           string
		s3expectedTransaction Transaction
	)

	BeforeEach(func() {
		dat, _ := ioutil.ReadFile("testemails/citiEmail.html")
		emailbody = string(dat)
		selectedParser = citiParser()
		expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 2345,
			Date:       "2020-10-14",
			Amount:     12.34,
			Merchant:   "I AM A LARGE #MERCHANT",
		}
		s3messageid = "3871qtv1f3f4r911mm68nrodkp3gtcmvkn25eq81"
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

	//Context("When downloading a mail from S3, ", func() {
	//
	//	It("parses the function correctly", func() {
	//		mailbody, err := retrieveMail(s3messageid)
	//		Expect(err).To(BeNil())
	//		transaction, err := parseEmail(mailbody)
	//		Expect(transaction).To(Equal(s3expectedTransaction))
	//	})
	//})

})

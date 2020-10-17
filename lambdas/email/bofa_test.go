package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("Parse BofA emails", func() {

	var (
		emailbody             string
		expectedTransaction   Transaction
		s3messageid           string
		s3expectedTransaction Transaction
	)

	BeforeEach(func() {
		dat, _ := ioutil.ReadFile("testemails/bofAEmail.txt")
		emailbody = string(dat)
		selectedParser = bofAParser()
		expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 5678,
			Date:       "2020-10-15",
			Amount:     9.99,
			Merchant:   "LARGE.COM PROVIDER",
		}
		s3messageid = "hloop9dhu7j61bpd7m01ogjlponmdhslic8o5f01"
		s3expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 1234,
			Date:       "2020-10-13",
			Amount:     109.00,
			Merchant:   "Test Mer\\chant.com",
		}
		_ = s3messageid
		_ = s3expectedTransaction
	})

	Context("When given an email body to parse, ", func() {

		It("parses the function correctly", func() {
			transaction, err := parseEmail(emailbody)
			Expect(err).To(BeNil())
			Expect(transaction).To(Equal(expectedTransaction))
		})

	})
	//
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

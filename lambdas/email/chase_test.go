package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("Parse Chase emails", func() {

	var (
		emailbody           string
		expectedTransaction Transaction
	)

	BeforeEach(func() {
		dat, _ := ioutil.ReadFile("testemails/chaseEmail.txt")
		emailbody = string(dat)
		selectedParser = chaseParser()
		expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 1234,
			Date:       "2020-10-13",
			Amount:     109.00,
			Merchant:   "Test Mer\\chant.com",
		}
	})

	Context("When given an email body to parse, ", func() {

		It("parses the function correctly", func() {
			transaction, err := parseEmail(emailbody)
			Expect(err).To(BeNil())
			Expect(transaction).To(Equal(expectedTransaction))
		})

	})

})

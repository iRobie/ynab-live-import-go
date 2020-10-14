package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("Parse BofA emails", func() {

	var (
		emailbody           string
		expectedTransaction Transaction
	)

	BeforeEach(func() {
		dat, _ := ioutil.ReadFile("testemails/bofAEmail.txt")
		emailbody = string(dat)
		selectedParser = bofAParser()
		expectedTransaction = Transaction{
			MessageID:  "",
			LastDigits: 5678,
			Date:       "2020-10-13",
			Amount:     9.99,
			Merchant:   "LARGE.COM PROVIDER",
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

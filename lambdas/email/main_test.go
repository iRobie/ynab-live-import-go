package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Email suite")
}

var _ = Describe("Parse Chase emails", func() {

	var (
		emailbody           string
		expectedTransaction Transaction
		lastDigits          = "1234"
		amountString        = "109.00"
		merchant            = "Test Mer\\chant.com"
		date                = "2020-10-13"
	)

	BeforeEach(func() {
		dat, err := ioutil.ReadFile("testemails/chaseEmail.txt")
		check(err)
		emailbody = string(dat)
		setupChase()
		digits, _ := strconv.Atoi(lastDigits)
		amount, _ := strconv.ParseFloat(amountString, 32)
		expectedTransaction = Transaction{
			LastDigits: digits,
			Date:       date,
			Amount:     float32(amount),
			Merchant:   merchant,
		}
	})

	Context("When given an email body to parse, ", func() {

		It("parses the last 4 digits", func() {
			test, err := getLastDigits(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal(lastDigits))
		})

		It("grabs the spend amount", func() {
			test, err := getSpendAmount(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal(amountString))
		})

		It("grabs the Merchant", func() {
			test, err := getMerchant(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal(merchant))
		})

		It("grabs the date", func() {
			test, err := getDate(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal(date))
		})

		It("parses everything correctly", func() {
			transaction, err := parseEmail(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(transaction).To(Equal(expectedTransaction))
		})

	})

})

var _ = Describe("Parse BofA emails", func() {

	var (
		emailbody string
	)

	BeforeEach(func() {
		dat, err := ioutil.ReadFile("testemails/bofAEmail.txt")
		check(err)
		emailbody = string(dat)
		setupBofA()
	})

	Context("When given an email body to parse, ", func() {

		It("parses the last 4 digits", func() {
			test, err := getLastDigits(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("5678"))
		})

		It("grabs the spend amount", func() {
			test, err := getSpendAmount(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("9.99"))
		})

		It("grabs the Merchant", func() {
			test, err := getMerchant(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("LARGE.COM PROVIDER"))
		})

		It("grabs the date", func() {
			test, err := getDate(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("2020-10-13"))
		})

	})

})

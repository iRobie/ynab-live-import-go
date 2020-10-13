package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Email suite")
}

var _ = Describe("Parse Chase emails", func() {

	var (
		emailbody string
		spambody  string
	)

	BeforeEach(func() {
		dat, err := ioutil.ReadFile("testemails/chaseEmail.txt")
		check(err)
		emailbody = string(dat)
		dat, err = ioutil.ReadFile("testemails/testspam.txt")
		check(err)
		spambody = string(dat)
		setupChase()
	})

	Context("When given an email body to parse, ", func() {
		It("checks for validation string", func() {
			err := validateEmail(emailbody, chaseString)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("fails the validation string for spam", func() {
			err := validateEmail(spambody, chaseString)
			Expect(err).Should(HaveOccurred())
		})

		It("parses the last 4 digits", func() {
			test, err := getLastDigits(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("1234"))
		})

		It("grabs the spend amount", func() {
			test, err := getSpendAmount(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("109.00"))
		})

		It("grabs the Merchant", func() {
			test, err := getMerchant(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("Test Mer\\chant.com"))
		})

		It("grabs the date", func() {
			test, err := getDate(emailbody)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(test).To(Equal("Oct 13, 2020"))
		})

		It("parses the date", func() {
			test, _ := getDate(emailbody)
			date, err := parseDate(test)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(date).To(Equal("2020-10-13"))
		})
	})

})

var _ = Describe("Parse BofA emails", func() {

	var (
		emailbody string
		spambody  string
	)

	BeforeEach(func() {
		dat, err := ioutil.ReadFile("testemails/bofAEmail.txt")
		check(err)
		emailbody = string(dat)
		dat, err = ioutil.ReadFile("testemails/testspam.txt")
		check(err)
		spambody = string(dat)
		setupBofA()
	})

	Context("When given an email body to parse, ", func() {
		It("checks for validation string", func() {
			err := validateEmail(emailbody, bofaString)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("fails the validation string for spam", func() {
			err := validateEmail(spambody, bofaString)
			Expect(err).Should(HaveOccurred())
		})

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
			Expect(test).To(Equal("October 13, 2020"))
		})

		It("parses the date", func() {
			test, _ := getDate(emailbody)
			date, err := parseDate(test)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(date).To(Equal("2020-10-13"))
		})
	})

})

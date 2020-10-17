package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go.bmvs.io/ynab/api"
	"go.bmvs.io/ynab/api/account"
	ynabtransaction "go.bmvs.io/ynab/api/transaction"
	"io/ioutil"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handle DynamoDBEvent", func() {

	var (
		inputEvent                 events.DynamoDBEvent
		image                      map[string]events.DynamoDBAttributeValue
		expectedTransaction        Transaction
		fakeBudgetAccount          budgetAccount
		expectedPayloadTransaction ynabtransaction.PayloadTransaction
		dynamoclient               *dynamodb.DynamoDB
	)

	BeforeEach(func() {
		jsonBody, _ := ioutil.ReadFile("testdata.json")

		if err := json.Unmarshal(jsonBody, &inputEvent); err != nil {
			log.Fatalf("could not unmarshal event. details: %v", err)
		}

		for _, record := range inputEvent.Records {
			image = record.Change.NewImage
		}

		expectedTransaction = Transaction{
			MessageID:  "asdfasdfasdfasdf",
			LastDigits: 1234,
			Date:       "2020-10-13",
			Amount:     56.78,
			Merchant:   "Github",
		}

		fakeBudgetAccount = budgetAccount{
			budgetID: "fakebudgetid",
			account:  &account.Account{ID: "fakeaccountid"},
		}

		dynamoclient = getDynamoClient("us-west-2")
		_ = dynamoclient

		date, _ := api.DateFromString("2020-10-13")
		memo := "Imported via email"
		expectedPayloadTransaction = ynabtransaction.PayloadTransaction{
			AccountID:  "fakeaccountid",
			Date:       date,
			Amount:     -56780,
			Cleared:    "uncleared",
			Approved:   false,
			PayeeID:    nil,
			PayeeName:  &expectedTransaction.Merchant,
			CategoryID: nil,
			Memo:       &memo,
			FlagColor:  nil,
			ImportID:   nil,
		}

	})

	Context("When given a Dynamo Record, ", func() {

		It("sends slack correctly", func() {
			err := fmt.Errorf("Testing from test suite")
			notifyError("Testing notify", err)
		})

		It("unmarshalls record correctly", func() {
			dynamoTransaction, err := unmarshallDynamoRecord(image)
			Expect(err).To(BeNil())
			Expect(dynamoTransaction).To(Equal(expectedTransaction))
		})

		It("get payload correctly", func() {
			dynamoTransaction, err := unmarshallDynamoRecord(image)
			payload, err := getPayload(dynamoTransaction, fakeBudgetAccount)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(expectedPayloadTransaction))
		})

		//It("gets duplicate correctly", func() {
		//	err := getDuplicate(dynamoclient, "aau75dockiceclf9olvjrgumgli2r40nn1gf7j81")
		//	Expect(err).To(BeNil())
		//})
		//
		//It("gets deletes correctly", func() {
		//	err := deleteDynamoRecord(dynamoclient, "aau75dockiceclf9olvjrgumgli2r40nn1gf7j81")
		//	Expect(err).To(BeNil())
		//})

	})

})

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ynab suite")
}

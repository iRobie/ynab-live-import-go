package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"go.bmvs.io/ynab"
	"go.bmvs.io/ynab/api"
	"go.bmvs.io/ynab/api/account"
	"go.bmvs.io/ynab/api/budget"
	"go.bmvs.io/ynab/api/transaction"
	"log"
	"os"
)

var (
	accessToken string
)

type budgetAccount struct {
	budgetID string
	account  *account.Account
}

func init() {
	// check if variable is already setup
	accessToken = os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		// Probably running locally. Load variables from .env file instead.
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file")
		}
		accessToken = os.Getenv("ACCESS_TOKEN")
	}
}

func main() {
	if accessToken == "" {
		log.Fatal("Missing ynab access token")
	}

	client := ynab.NewClient(accessToken)
	budgets, err := client.Budget().GetBudgets()
	if err != nil {
		log.Fatalf("Could not retreive list of budgets: " + err.Error())
	}

	accounts, err := getAccounts(client, budgets)
	if err != nil {
		log.Fatalf("Could not retreive list of accounts: " + err.Error())
	}

	account, err := getAccountID(accounts, "2645")
	if err != nil {
		log.Fatalf("Could not find account for digit: " + err.Error())
	}

	fmt.Println("Posting transaction to account " + account.account.Name)

	transaction := getTransaction(account)
	err = postTransactionToAccount(client, account, transaction)
	if err != nil {
		log.Fatalf("Error posting the transaction to ynab: " + err.Error())
	}

}

func getAccounts(client ynab.ClientServicer, budgets []*budget.Summary) ([]budgetAccount, error) {
	var accounts []budgetAccount

	for _, ynabBudget := range budgets {
		ynabAccounts, err := client.Account().GetAccounts(ynabBudget.ID)
		if err != nil {
			log.Printf("Could not retreive list of accounts for budget: " + err.Error())
		}
		for _, ynabAccount := range ynabAccounts {
			account := budgetAccount{account: ynabAccount, budgetID: ynabBudget.ID}
			accounts = append(accounts, account)
		}
	}

	if len(accounts) == 0 {
		return accounts, errors.New("could not find any accounts")
	}
	return accounts, nil
}

// Checks accounts for the ID given. This is the last 4 digits of the CC passed in email.
// The 4 digits must be added to the notes section of YNAB.
func getAccountID(accounts []budgetAccount, digits string) (budgetAccount, error) {

	for _, ynabAccount := range accounts {
		if ynabAccount.account.Note != nil {
			if *ynabAccount.account.Note == digits {
				return ynabAccount, nil
			}
		}

	}
	return budgetAccount{}, errors.New("Could not find account matching the given digits")
}

func postTransactionToAccount(client ynab.ClientServicer, account budgetAccount, payloadTransaction transaction.PayloadTransaction) error {
	_, err := client.Transaction().CreateTransaction(account.budgetID, payloadTransaction)
	if err != nil {
		return err
	}
	return nil

}

func getTransaction(account budgetAccount) transaction.PayloadTransaction {

	date, err := api.DateFromString("2020-10-13")
	if err != nil {
		log.Printf("WARNING - could not parse date correctly")
	}

	amount := int64(9.99 * 1000)
	payee := "My test"
	memo := "Imported via email"

	payloadTransaction := transaction.PayloadTransaction{
		AccountID: account.account.ID,
		Date:      date,
		Amount:    amount,
		Cleared:   "uncleared",
		PayeeName: &payee,
		Memo:      &memo,
	}

	return payloadTransaction
}

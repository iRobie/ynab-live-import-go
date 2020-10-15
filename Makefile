lambda: 
	cd lambdas/email; GOOS=linux GOARCH=amd64 go build -o ../../bin/email
	cd lambdas/ynab; GOOS=linux GOARCH=amd64 go build -o ../../bin/ynab

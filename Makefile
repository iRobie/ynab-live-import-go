lambda:
	cd lambdas/email; GOOS=linux GOARCH=amd64 go build -o ../../bin/email
	cd lambdas/ynab; GOOS=linux GOARCH=amd64 go build -o ../../bin/ynab

zip: lambda
	cd bin; zip email.zip email
	cd bin; zip ynab.zip ynab
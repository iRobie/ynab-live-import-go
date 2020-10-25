.PHONY: zip

terraform: zip
	cd terraform; terraform apply

zip: clean lambda
	cd bin; zip email.zip email
	cd bin; zip ynab.zip ynab

clean:
	rm -f ../../bin/email
	rm -f ../../bin/ynab
	rm -f ../../bin/email.zip
	rm -f ../../bin/ynab.zip

lambda:
	cd lambdas/email; GOOS=linux GOARCH=amd64 go build -o ../../bin/email
	cd lambdas/ynab; GOOS=linux GOARCH=amd64 go build -o ../../bin/ynab

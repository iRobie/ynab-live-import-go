// Create zip file
data "archive_file" "email" {
  type        = "zip"
  source_file = "../bin/email"
  output_path = "../bin/email.zip"
  depends_on = [
    null_resource.makefile,
  ]
}

// Create lambda function
resource "aws_lambda_function" "email" {
  function_name    = "ynab-email-parser"
  filename         = data.archive_file.email.output_path
  handler          = "email" // For go, this is the name of the file.
  source_code_hash = data.archive_file.email.output_base64sha256
  role             = aws_iam_role.email_parser.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 10
  environment {
    variables = {
      BUCKET_NAME = aws_s3_bucket.bucket.bucket
      TABLE_NAME = aws_dynamodb_table.dynamodb-table.name
    }
  }
}

// Give lambda function necessary permissions
resource "aws_iam_role" "email_parser" {
  name = "ynab_email_parser_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "gets3objects" {
  name        = "get-ynab-s3-objects"
  description = "Policy to access S3 bucket with YNAB emails"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Resource": [
                "${aws_s3_bucket.bucket.arn}/*"
            ]
    }
  ]
}
EOF
}

resource "aws_iam_policy" "putDynamo" {
  name        = "put-ynab-dynamo-table"
  description = "Policy to put objects to DynamoDB table"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:PutItem"
      ],
      "Effect": "Allow",
      "Resource": "${aws_dynamodb_table.dynamodb-table.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "email-policy-attachment-s3" {
  role       = aws_iam_role.email_parser.name
  policy_arn = aws_iam_policy.gets3objects.arn
}

resource "aws_iam_role_policy_attachment" "email-policy-attachment-dynamo" {
  role       = aws_iam_role.email_parser.name
  policy_arn = aws_iam_policy.putDynamo.arn
}

resource "aws_iam_role_policy_attachment" "email-policy-executionrole" {
  role       = aws_iam_role.email_parser.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

// Allow SES to call lambda function
resource "aws_lambda_permission" "email-parser" {
  statement_id  = "AllowExecutionFromSES"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.email.function_name
  principal     = "ses.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
}
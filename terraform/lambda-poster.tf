data "archive_file" "ynab" {
  type        = "zip"
  source_file = "../bin/ynab"
  output_path = "../bin/ynab.zip"
}

resource "aws_lambda_function" "poster" {
  function_name    = "ynab-poster"
  filename         = data.archive_file.ynab.output_path
  handler          = "ynab" // For go, this is the name of the file.
  source_code_hash = data.archive_file.ynab.output_base64sha256
  role             = aws_iam_role.ynab_poster.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 10
  environment {
    variables = {
      BUCKET_NAME = aws_s3_bucket.bucket.bucket
      TABLE_NAME = aws_dynamodb_table.dynamodb-table.name
      ACCESS_TOKEN = var.ynab_access_token
      SLACK_URL = var.slack_url
    }
  }
}

resource "aws_iam_role" "ynab_poster" {
  name = "ynab_poster_role"

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

resource "aws_iam_policy" "rms3objects" {
  name        = "rm-ynab-s3-objects"
  description = "Policy to rm objects with YNAB emails"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:DeleteObject"
      ],
      "Effect": "Allow",
      "Resource": "${aws_s3_bucket.bucket.arn}/*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "updateDynamo" {
  name        = "update-ynab-dynamo-table"
  description = "Policy to put objects to DynamoDB table"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:DeleteItem",
        "dynamodb:GetItem"
      ],
      "Effect": "Allow",
      "Resource": "${aws_dynamodb_table.dynamodb-table.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "poster-policy-attachment-s3" {
  role       = aws_iam_role.ynab_poster.name
  policy_arn = aws_iam_policy.rms3objects.arn
}

resource "aws_iam_role_policy_attachment" "poster-policy-attachment-dynamo" {
  role       = aws_iam_role.ynab_poster.name
  policy_arn = aws_iam_policy.updateDynamo.arn
}

resource "aws_iam_role_policy_attachment" "poster-policy-basicexecutionrole" {
  role       = aws_iam_role.ynab_poster.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "poster-policy-dynamoexecutionrole" {
  role       = aws_iam_role.ynab_poster.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole"
}

resource "aws_lambda_event_source_mapping" "ynab-poster-stream" {
  event_source_arn  = aws_dynamodb_table.dynamodb-table.stream_arn
  function_name     = aws_lambda_function.poster.arn
  starting_position = "LATEST"
  batch_size = 1
}
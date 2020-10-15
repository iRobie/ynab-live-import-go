provider "aws" {
  profile = "default"
  region  = "us-west-2"
  version = "~> 2.7.0"
}

resource "aws_iam_policy" "policy" {
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
      "Resource": "${aws_s3_bucket.bucket.arn}"
    }
  ]
}
EOF
}
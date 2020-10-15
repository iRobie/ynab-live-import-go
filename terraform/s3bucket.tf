resource "aws_s3_bucket" "bucket" {
  bucket = var.s3_bucket_name
  acl    = "private"

  lifecycle_rule {
    enabled = true

    expiration {
      days = 2
    }
  }
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "AllowSESPuts",
  "Statement": [
    {
      "Sid": "AllowSESPuts",
      "Effect": "Allow",
      "Principal": {
                "Service": [
                    "ses.amazonaws.com"
                ]
            },
      "Action": "s3:PutObject",
      "Resource": "${aws_s3_bucket.bucket.arn}/*",
      "Condition": {
         "StringEquals": {"aws:Referer": "${data.aws_caller_identity.current.account_id}"}
      }
    }
  ]
}
POLICY
}
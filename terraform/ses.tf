# Store email in S3
resource "aws_ses_receipt_rule_set" "main" {
  rule_set_name = "ynab-live-import-rule-set"
}

resource "aws_ses_receipt_rule" "ynabimportruleset" {
name          = "process-email"
rule_set_name = "ynab-live-import-rule-set"
recipients    = toset([var.domain_name])
enabled       = true

  s3_action {
  bucket_name = aws_s3_bucket.bucket.bucket
  position    = 1
  }

  lambda_action {
    function_arn = aws_lambda_function.email.arn
    position = 2
  }
}
resource "aws_dynamodb_table" "dynamodb-table" {
  name           = "ynab_transactions"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "messageID"

  stream_enabled = true
  stream_view_type = "NEW_IMAGE"

  attribute {
    name = "messageID"
    type = "S"
  }

}
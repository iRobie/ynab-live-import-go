variable s3_bucket_name {
  type = string
}

variable ynab_access_token {
  type = string
}

variable domain_name {
  type = string
  description = "Domain name - example.com"
}

variable slack_url {
  type = string
  description = "Slack webhook url for notifications"
}
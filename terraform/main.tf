provider "aws" {
  profile = "default"
  region  = "us-west-2"
  version = "~> 2.7.0"
}

// Get account ID
data "aws_caller_identity" "current" {}


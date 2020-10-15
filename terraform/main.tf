provider "aws" {
  profile = "default"
  region  = "us-west-2"
  version = "~> 2.7.0"
}

//Make the bin files for upload
resource "null_resource" "makefile" {
  provisioner "local-exec" {
    command = "make"
    working_dir = "../"
  }
}

// Get account ID
data "aws_caller_identity" "current" {}


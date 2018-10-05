resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = "StationCatalog"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "Planet"

  attribute {
    name = "Planet"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled = false
  }

  tags {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}
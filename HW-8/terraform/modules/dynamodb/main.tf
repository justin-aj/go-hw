# DynamoDB module — single-table design for shopping carts.
#
# Partition key: cartId (String)
# Uses PAY_PER_REQUEST billing — no capacity planning needed for a lab.

resource "aws_dynamodb_table" "carts" {
  name         = "${var.service_name}-carts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cartId"

  attribute {
    name = "cartId"
    type = "S"
  }

  tags = { Name = "${var.service_name}-carts" }
}

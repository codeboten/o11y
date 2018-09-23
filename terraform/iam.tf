data "aws_caller_identity" "current" {}

variable "region" {
    type = "string"
}

provider "aws" {
      region     = "${var.region}"
}

data "aws_region" "current" {
}

data "aws_iam_policy_document" "serverless" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }
  statement {
      sid = "2"
      actions = [
        "cloudformation:DescribeStackResource",
        "cloudformation:DescribeStacks",
        "cloudformation:DescribeStackEvents",
        "cloudformation:CreateStack",
        "cloudformation:DeleteStack",
        "cloudformation:UpdateStack"
      ]
      resources = [
        "arn:aws:cloudformation:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:stack/helloworld-dev/*"
      ]
  }
  statement {
      sid = "3"
      actions = [
        "apigateway:DELETE",
        "apigateway:PATCH",
        "apigateway:POST",
        "apigateway:PUT",
        "apigateway:GET",
      ]
      resources = [
        "arn:aws:apigateway:${data.aws_region.current.name}::/restapis",
        "arn:aws:apigateway:${data.aws_region.current.name}::/restapis/*"
      ]
  }
  statement {
      sid = "4"
      actions = [
        "iam:GetRole",
        "iam:GetUser",
      ]
      resources = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:user/developer",
        "arn:aws:iam:::role/helloworld-dev-${data.aws_region.current.name}-lambdaRole",
      ]
  }
  statement {
      sid = "5"
      actions = [
        "logs:CreateLogGroup",
        "cloudformation:ValidateTemplate"
      ]
      resources = [
        "*",
      ]
  }
  statement {
      sid = "6"
      actions = [
        "lambda:AddPermission",
        "lambda:DeleteFunction",
        "lambda:GetFunction",
        "lambda:GetFunctionConfiguration",
        "lambda:InvokeFunction",
        "lambda:ListVersionsByFunction",
        "lambda:PublishVersion",
        "lambda:RemovePermission",
        "lambda:UpdateFunctionCode",
      ]
      resources = [
        "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:helloworld-dev-hello",
        "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:helloworld-dev-world",
      ]
  }
  statement {
      sid = "61"
      actions = [
        "lambda:CreateFunction",
      ]
      resources = [
        "*"
      ]
  }
  statement {
      sid = "7"
      actions = [
        "logs:DescribeLogGroups",
        "logs:DeleteLogGroup",
      ]
      resources = [
        "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group::log-stream:",
        "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/helloworld-dev-hello:log-stream:",
        "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/helloworld-dev-world:log-stream:",
      ]
  }
}

resource "aws_iam_user" "developer" {
    name = "developer"
}

resource "aws_iam_policy" "serverless_policy" {
  name        = "serverless_policy"
  path        = "/"
  description = "Policy needed for serverless deployment"
  policy = "${data.aws_iam_policy_document.serverless.json}"
}
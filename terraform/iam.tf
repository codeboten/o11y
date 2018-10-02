data "aws_caller_identity" "current" {}

variable "region" {
    type = "string"
}

variable "weatherary-app" {
    type = "string"
}

variable "weatherary-fn" {
    type = "string"
}

variable "planetary-api" {
  type = "string"
}

variable "pa-fn" {
  type = "string"
}

resource "aws_iam_user" "developer" {
    name = "developer"
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
        "arn:aws:cloudformation:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:stack/${var.weatherary-app}/*",
        "arn:aws:cloudformation:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:stack/${var.planetary-api}/*"
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
        "${aws_iam_user.developer.arn}",
        "arn:aws:iam:::role/${var.weatherary-app}-${data.aws_region.current.name}-lambdaRole",
        "arn:aws:iam:::role/${var.planetary-api}-${data.aws_region.current.name}-lambdaRole",
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
        "lambda:UpdateFunctionConfiguration",
      ]
      resources = [
        "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:${var.weatherary-app}-${var.weatherary-fn}",
        "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:${var.planetary-api}-${var.pa-fn}",
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
        "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.weatherary-app}-${var.weatherary-fn}:log-stream:",
        "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.planetary-api}-${var.pa-fn}:log-stream:",
      ]
  }
}


resource "aws_iam_policy" "serverless_policy" {
  name        = "serverless_policy"
  path        = "/"
  description = "Policy needed for serverless deployment"
  policy = "${data.aws_iam_policy_document.serverless.json}"
}
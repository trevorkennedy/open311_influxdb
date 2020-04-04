# Golang Open311 Client for InfluxDB

A Golang client that writes events from the Austin 311 API to InfluxCloud. Packaged as a AWS Lambda function that executes every hour using CloudWatch.

### Requirements
  - Go
  - AWS Command Line Interface

### Build instructions

Create a config.json from the config_template.json file with the appropriate InfluxDB credentials

> GOOS=linux go build

> zip -r main.zip .

Create the lambda function (delete the previous version of the function if necessary)

> aws lambda create-function \
>  --region us-east-1 \
>  --function-name open311-go \
>  --memory 128 \
>  --role arn:aws:iam::312318060469:role/service-role/hello-world \
>  --runtime go1.x \
>  --zip-file fileb://main.zip \
>  --handler open311-go

Create a CloudWatch trigger in the AWS Console to execute the function on an hourly basis.

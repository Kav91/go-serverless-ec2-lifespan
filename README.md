# Serverless Golang EC2 Lifespan

Provide an EC2 Instance a Lifespan tag with integer value representing remaining minutes left to live or time eg. 14:22 to stop by.


- example
- - `"10"` (10 minutes left to live)
- - `"15:30"` (will stop at specificed time)
- Optionally supply -terminate at the end of time period, to terminate the instance
- - eg. `"10-terminate"` (will terminate instance in 10m)

```sh
Ensure you have serverless framework installed & relevant imports added, see within .go files
https://serverless.com/framework/docs/providers/aws/guide/installation/

aws sdk see: https://docs.aws.amazon.com/sdk-for-go/api/
Other packages:
go get -u github.com/sirupsen/logrus
go get -u github.com/korovkin/limiter
go get -u github.com/aws/aws-lambda-go/lambda
```



### Config
```sh
serverless.yml file
See region, memorySize & timeout options to adjust if required.

main.go file
There are several options to adjust as below:

//If AWS_REGION environment variable is supplied the below default will be overwritten
var region = "us-east-1" 

//Lambda timezone is UTC by default, this will correct it if using the timestamp option
var timezone = "Australia/Sydney" 

//If an invalid lifespan tag is supplied it will be given 86400 minutes to like == 1 day
var defaultLifespan = 86400 

//limit number of concurrent executions, useful if you are making a large amount of calls to the EC2 API service
var concurrency = 5 

```

### Compiling & Deploying

As AWS Lambda runs on linux you will need to compile for linux
For further info: https://github.com/aws/aws-lambda-go

```sh
GOOS=linux GOARCH=amd64 go build -o bin/main main.go ec2.go
sls deploy -v
```
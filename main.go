package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/korovkin/limiter"
	log "github.com/sirupsen/logrus"
)

var region = "us-east-1"
var timezone = "Australia/Sydney"
var defaultLifespan = 86400
var concurrency = 5
var jobQueue = limiter.NewConcurrencyLimiter(concurrency)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	if os.Getenv("AWS_REGION") != "" {
		region = os.Getenv("AWS_REGION")
	}
}

func handler() {
	loc, _ := time.LoadLocation(timezone)
	currentTime := time.Now().In(loc)
	jobQueue = limiter.NewConcurrencyLimiter(concurrency) //we need to set the concurrency limiter again else subsequent lambda executions will fail
	log.WithFields(log.Fields{"timezone": currentTime}).Info("Beginning EC2 Lifespan Execution")

	lifespanInstances := getInstances()
	for _, instance := range lifespanInstances.([]instanceInfo) {
		confirmLifespan(instance, currentTime)
	}

	jobQueue.Wait()

	log.WithFields(log.Fields{
		"instancesProcessed": len(lifespanInstances.([]instanceInfo)),
	}).Info("Lifespan Execution Complete")

}

func main() {
	lambda.Start(handler)
}

func confirmLifespan(instance instanceInfo, currentTime time.Time) {
	jobQueue.Execute(func() {
		action := "stop"
		tag := strings.Split(instance.lifespan, "-")

		if len(tag) == 2 && tag[1] == "terminate" {
			action = "terminate"
		}

		if strings.Contains(tag[0], ":") {
			//Check Timestamp
			time := strings.Split(tag[0], ":")
			hour := strconv.FormatInt(int64(currentTime.Hour()), 10)
			minute := strconv.FormatInt(int64(currentTime.Minute()), 10)

			if hour == time[0] && minute == time[1] {
				if action == "terminate" {
					terminateInstance(instance)
				} else {
					stopInstance(instance)
				}
			} else {
				log.WithFields(log.Fields{
					"instance": instance.id,
				}).Info("Not yet time to stop")
			}

		} else {
			//Check Minutes
			minutes, err := strconv.ParseInt(tag[0], 10, 0)
			if err != nil {
				updateInstanceTag(instance, int64(defaultLifespan), action)
			} else {
				if minutes > 0 {
					updateInstanceTag(instance, minutes, action)
				} else if minutes == 0 {
					if action == "terminate" {
						terminateInstance(instance)
					} else {
						stopInstance(instance)
					}
				}
			}

		}
	})
}

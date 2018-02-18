package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

var sess = session.Must(session.NewSession())
var svc = ec2.New(sess, aws.NewConfig().WithRegion(region))

type instanceInfo struct {
	id       string
	state    int64
	tags     []*ec2.Tag
	lifespan string
}

func getInstances() interface{} {
	input := &ec2.DescribeInstancesInput{}
	result, err := svc.DescribeInstances(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		os.Exit(0)
	}

	instanceSummary := []instanceInfo{}
	//Exxtract only running instances with lifespan tag
	for _, instance := range result.Reservations {
		if *instance.Instances[0].State.Code == 16 { //16 == running: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_InstanceState.html
			for _, tag := range instance.Instances[0].Tags {
				if *tag.Key == "Lifespan" {
					instanceSummary = append(
						instanceSummary,
						instanceInfo{
							id:       *instance.Instances[0].InstanceId,
							state:    *instance.Instances[0].State.Code,
							tags:     instance.Instances[0].Tags,
							lifespan: *tag.Value,
						})
				}
			}
		}
	}

	log.WithFields(log.Fields{
		"instances": len(instanceSummary),
	}).Info("Fetch instances complete")

	return instanceSummary
}

func updateInstanceTag(instance instanceInfo, minutes int64, action string) {

	updatedMinutes := minutes - 1 //subtract 1 minute
	updatedTag := strconv.FormatInt(updatedMinutes, 10) + "-" + action
	input := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(instance.id),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Lifespan"),
				Value: aws.String(updatedTag),
			},
		},
	}

	result, err := svc.CreateTags(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			log.WithFields(log.Fields{
				"instance": instance.id,
				"err":      err.Error(),
			}).Info("Error Updating Lifespan Tag")
		}
	} else {
		log.WithFields(log.Fields{
			"instance":       instance.id,
			"minutesLeftWas": minutes,
			"minutesLeftNow": updatedMinutes,
			"result":         result,
		}).Info("Updated Lifespan Tag")
	}

	time.Sleep(time.Second * 3)

}

func stopInstance(instance instanceInfo) {
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.id),
		},
		DryRun: aws.Bool(false),
	}

	result, err := svc.StopInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			log.WithFields(log.Fields{
				"instance": instance.id,
				"err":      err.Error(),
			}).Info("Error Stopping Instance")
		}
	} else {
		log.WithFields(log.Fields{
			"instance": instance.id,
			"result":   result,
		}).Info("Stopping Instance")
	}

}

func terminateInstance(instance instanceInfo) {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.id),
		},
		DryRun: aws.Bool(false),
	}

	result, err := svc.TerminateInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			log.WithFields(log.Fields{
				"instance": instance.id,
				"err":      err.Error(),
			}).Info("Error Terminating Instance")
		}
	} else {
		log.WithFields(log.Fields{
			"instance": instance.id,
			"result":   result,
		}).Info("Terminating Instance")
	}

}

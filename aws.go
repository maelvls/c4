package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func nukeAWSInstances(accessKey, secretKey, region string, regex *regexp.Regexp, dryRun bool, olderThan time.Duration) ([]ec2.Instance, error) {
	sconf := aws.NewConfig().WithCredentials(
		credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
	)
	sconf = sconf.WithRegion(region)
	sess, err := session.NewSession(sconf)
	if err != nil {
		return nil, err
	}
	ec2client := ec2.New(sess)
	res, err := ec2client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	var toBeDeleted []ec2.Instance
	var toBeDeletedIds []*string // needed for TerminateInstances
	for _, reserv := range res.Reservations {
		for _, instance := range reserv.Instances {
			name := awsName(*instance)
			if !regex.MatchString(name) {
				continue
			}

			age := awsAge(*instance)
			if age >= olderThan {
				toBeDeleted = append(toBeDeleted, *instance)
				toBeDeletedIds = append(toBeDeletedIds, instance.InstanceId)
				fmt.Printf("aws: %s (%s), age: %s\n", yel(name), *instance.InstanceId, red(age.Truncate(time.Second).String()))
			} else {
				fmt.Printf("aws: %s (%s), age: %s\n", yel(name), *instance.InstanceId, green(age.Truncate(time.Second).String()))
			}
		}
	}

	if dryRun || len(toBeDeletedIds) == 0 {
		return toBeDeleted, nil
	}
	_, err = ec2client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: toBeDeletedIds,
	})
	if err != nil {
		return nil, err
	}

	return toBeDeleted, nil
}

func awsName(instance ec2.Instance) string {
	for _, t := range instance.Tags {
		if *t.Key == "Name" {
			return *t.Value
		}
	}

	return ""
}

func awsAge(instance ec2.Instance) time.Duration {
	return time.Now().Sub(*instance.LaunchTime)
}

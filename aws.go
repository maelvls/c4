package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func nukeAWSInstances(accessKey, secretKey, region, nameContains string, dryRun bool, olderThan time.Duration) error {
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
		return err
	}
	ec2client := ec2.New(sess)
	res, err := ec2client.DescribeInstances(&ec2.DescribeInstancesInput{Filters: []*ec2.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: aws.StringSlice([]string{"*" + nameContains + "*"}),
		},
	}})
	if err != nil {
		return err
	}
	var ids []*string
	for _, reserv := range res.Reservations {
		for _, instance := range reserv.Instances {
			age := time.Now().Sub(*instance.LaunchTime)
			name := findTagOrEmpty(instance.Tags, "Name")
			if age >= olderThan {
				ids = append(ids, instance.InstanceId)
				fmt.Printf("found aws instance %s (%s), removing since age is %s\n", yel(name), *instance.InstanceId, red(age.String()))
			} else {
				fmt.Printf("found aws instance %s (%s), keeping it since age is %s\n", yel(name), *instance.InstanceId, green(age.String()))
			}
		}
	}

	if dryRun || len(ids) == 0 {
		return nil
	}
	_, err = ec2client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: ids,
	})
	if err != nil {
		return err
	}

	return nil
}

func findTagOrEmpty(tags []*ec2.Tag, tag string) string {
	for _, t := range tags {
		if *t.Key == tag {
			return *t.Value
		}
	}

	return ""
}

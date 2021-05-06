package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	ctx           = context.TODO()
	envInstanceID = os.Getenv("AWS_DEV_INSTANCE_ID")
	envUserID     = os.Getenv("AWS_DEV_USER_ID")
	envAccountID  = os.Getenv("AWS_DEV_ACCOUNT_ID")
	envSTSProfile = os.Getenv("AWS_DEV_STS_PROFILE")

	accountID  = flag.String("a", envAccountID, "Account hosting the dev VM")
	instanceID = flag.String("i", envInstanceID, "The ID of the EC2 instance")
	userID     = flag.String("u", envUserID, "User ID for STS login")
	stsProfile = flag.String("p", envSTSProfile, "profile to use with STS")
	tokenCode  = flag.String("t", "", "MFA token code")
)

func main() {
	flag.Parse()
	if len(*tokenCode) == 0 {
		log.Fatal("No token code provided (use -t)")
	}
	if len(*accountID) == 0 {
		log.Fatal("No account ID provided (use -a or AWS_DEV_ACCOUNT_ID)")
	}
	if len(*userID) == 0 {
		log.Fatal("No user ID provided (use -u or AWS_DEV_USER_ID)")
	}
	if len(*instanceID) == 0 {
		log.Fatal("No EC2 instance ID provided (use -i or AWS_DEV_INSTANCE_ID)")
	}
	if len(*stsProfile) == 0 {
		log.Fatal("No STS profile specified (use -p or AWS_DEV_STS_PROFILE)")
	}
	if len(flag.Args()) < 1 {
		log.Fatal("No actions specified, supported: [start, stop, describe]")
	}
	switch flag.Arg(0) {
	case "start":
		cfg := logIn()
		ec2Client := ec2.NewFromConfig(cfg)
		startInstance(*instanceID, ec2Client)
		time.Sleep(10 * time.Second)
		describeInstance(*instanceID, ec2Client)
	case "stop":
		cfg := logIn()
		ec2Client := ec2.NewFromConfig(cfg)
		stopInstance(*instanceID, ec2Client)
	case "describe":
		cfg := logIn()
		ec2Client := ec2.NewFromConfig(cfg)
		describeInstance(*instanceID, ec2Client)
	default:
		log.Fatalf("Unknown action, %s, supported: [start, stop, describe]", flag.Arg(0))
	}
}

func logIn() aws.Config {
	stsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(*stsProfile),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	cfg, err := getSessionToken(stsCfg)
	if err != nil {
		log.Fatalf("unable to get session config, %v", err)
	}
	return *cfg
}

func getSessionToken(cfg aws.Config) (*aws.Config, error) {
	svc := sts.NewFromConfig(cfg)
	serialNumber := fmt.Sprintf("arn:aws:iam::%s:mfa/%s", *accountID, *userID)
	input := sts.GetSessionTokenInput{
		SerialNumber: &serialNumber,
		TokenCode:    tokenCode,
	}
	out, err := svc.GetSessionToken(ctx, &input)
	if err != nil {
		return nil, err
	}

	newCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     *out.Credentials.AccessKeyId,
				SecretAccessKey: *out.Credentials.SecretAccessKey,
				SessionToken:    *out.Credentials.SessionToken,
			},
		}))
	if err != nil {
		return nil, err
	}

	return &newCfg, nil
}

func startInstance(id string, client *ec2.Client) {
	input := ec2.StartInstancesInput{
		InstanceIds: []string{
			*instanceID,
		},
	}
	_, err := client.StartInstances(ctx, &input)
	if err != nil {
		log.Fatalf("Unable to start instance, %v", err)
	}
}

func stopInstance(id string, client *ec2.Client) {
	input := ec2.StopInstancesInput{
		InstanceIds: []string{
			*instanceID,
		},
	}
	_, err := client.StopInstances(ctx, &input)
	if err != nil {
		log.Fatalf("Unable to stop instance, %v", err)
	}
}

func describeInstance(id string, client *ec2.Client) {
	input := ec2.DescribeInstancesInput{
		InstanceIds: []string{
			*instanceID,
		},
	}
	out, err := client.DescribeInstances(ctx, &input)
	if err != nil {
		log.Fatalf("Unable to describe instance, %v", err)
	}

	fmt.Printf("Instance ID: %s\n", *out.Reservations[0].Instances[0].InstanceId)
	fmt.Printf("Public IP: %s\n", *out.Reservations[0].Instances[0].PublicIpAddress)
	fmt.Printf("Public DNS: %s\n", *out.Reservations[0].Instances[0].PublicDnsName)
}

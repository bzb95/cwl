package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchClient struct {
	client    *cloudwatchlogs.Client
	logGroup  string
	logStream string
}

func NewCloudWatchClient(profile, logGroup string) (*CloudWatchClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := cloudwatchlogs.NewFromConfig(cfg)

	// Create log group if it doesn't exist
	if err := createLogGroupIfNotExists(client, logGroup); err != nil {
		return nil, fmt.Errorf("failed to create log group: %w", err)
	}

	// Create a log stream with timestamp
	logStream := fmt.Sprintf("cwlogs-%d", time.Now().Unix())

	if err := createLogStream(client, logGroup, logStream); err != nil {
		return nil, fmt.Errorf("failed to create log stream: %w", err)
	}

	return &CloudWatchClient{
		client:    client,
		logGroup:  logGroup,
		logStream: logStream,
	}, nil
}

func createLogGroupIfNotExists(client *cloudwatchlogs.Client, logGroup string) error {
	_, err := client.CreateLogGroup(context.TODO(), &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroup),
	})
	if err != nil {
		var alreadyExists *types.ResourceAlreadyExistsException
		if !errorsAs(err, &alreadyExists) {
			return err
		}
		// Log group already exists, which is fine
	}
	return nil
}

func createLogStream(client *cloudwatchlogs.Client, logGroup, logStream string) error {
	_, err := client.CreateLogStream(context.TODO(), &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
	})
	if err != nil {
		var alreadyExists *types.ResourceAlreadyExistsException
		if !errorsAs(err, &alreadyExists) {
			return err
		}
		// Log stream already exists, which is fine
	}
	return nil
}

func (c *CloudWatchClient) SendLogs(messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	var logEvents []types.InputLogEvent
	timestamp := time.Now().UnixMilli()

	for _, message := range messages {
		logEvents = append(logEvents, types.InputLogEvent{
			Message:   aws.String(message),
			Timestamp: aws.Int64(timestamp),
		})
	}

	_, err := c.client.PutLogEvents(context.TODO(), &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(c.logGroup),
		LogStreamName: aws.String(c.logStream),
		LogEvents:     logEvents,
	})
	if err != nil {
		return fmt.Errorf("failed to send logs: %w", err)
	}

	return nil
}

// Helper function for error type assertion
func errorsAs(err error, target interface{}) bool {
	// This is a simplified version of errors.As for AWS SDK v2
	switch t := target.(type) {
	case **types.ResourceAlreadyExistsException:
		if e, ok := err.(*types.ResourceAlreadyExistsException); ok {
			*t = e
			return true
		}
	}
	return false
}

# cwlogs - CloudWatch Log Forwarder

A simple Go tool that reads from stdin and forwards logs to AWS CloudWatch Logs.

## Features

- **First-time setup**: Interactive configuration on first run
- **Configuration persistence**: Saves settings to `~/.config/cwlogs/config.json`
- **Session overrides**: Override log group and AWS profile via command-line flags
- **Batch processing**: Efficiently batches logs before sending to CloudWatch
- **Auto-creation**: Automatically creates log groups and streams as needed
- **AWS profile support**: Uses AWS credentials from specified profiles

## Installation

```bash
go build -o cwlogs .
```

## Usage

### First Run (Setup)
On first run, the tool will prompt you for configuration:
```bash
./cwlogs
```

You'll be asked to provide:
- CloudWatch log group name
- AWS profile name (defaults to "default")

### Regular Usage
```bash
# Use configured settings
./cwlogs

# Override log group for this session
./cwlogs -log my-app-logs

# Override AWS profile for this session
./cwlogs -profile dev

# Override both
./cwlogs -log my-app-logs -profile dev
```

### Configuration Management
```bash
# Re-run setup to update configuration
./cwlogs -setup
```

## Examples

### Basic usage
```bash
# Forward application logs
./my-app | ./cwlogs

# Forward system logs
tail -f /var/log/app.log | ./cwlogs

# Forward with custom log group
echo "Starting deployment" | ./cwlogs -log deployment-logs

# Use specific AWS profile
cat /var/log/nginx/access.log | ./cwlogs -profile dev -log nginx-logs
```

### Using environment variables
```bash
# Set AWS profile via environment
./my-app | ./cwlogs

# Set AWS region
./my-app | ./cwlogs -log my-app-logs
```

## Configuration

Configuration is stored in `~/.config/cwlogs/config.json`:

```json
{
  "log_group": "my-application-logs",
  "profile": "default"
}
```

## AWS Permissions

Ensure your AWS profile has the following permissions:
- `logs:CreateLogGroup`
- `logs:CreateLogStream`
- `logs:PutLogEvents`

Example IAM policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
```

## Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd cwlogs

# Download dependencies
go mod tidy

# Build the binary
go build -o cwlogs .
```

## Environment Variables

The tool respects standard AWS environment variables:
- `AWS_PROFILE`: Default AWS profile to use
- `AWS_REGION`: AWS region for CloudWatch Logs
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key

## Troubleshooting

### Common Issues

1. **"NoCredentialProviders" error**
   - Ensure AWS credentials are configured
   - Check `~/.aws/credentials` or environment variables

2. **"AccessDenied" error**
   - Verify IAM permissions for CloudWatch Logs
   - Check if the IAM user/role has the required permissions

3. **Log group not found**
   - The tool will automatically create log groups
   - Ensure you have `logs:CreateLogGroup` permission

4. **Logs not appearing in CloudWatch**
   - Check AWS region configuration
   - Verify log group name is correct
   - Check CloudWatch Logs console for the log stream

## Development

### Project Structure
```
cwlogs/
├── main.go          # Entry point and CLI handling
├── config.go        # Configuration management
├── cloudwatch.go    # AWS CloudWatch client
├── forwarder.go     # Log forwarding logic
├── go.mod          # Go module definition
└── README.md       # This file
```

### Adding Features
- Modify `config.go` for configuration changes
- Update `cloudwatch.go` for AWS SDK modifications
- Enhance `forwarder.go` for log processing improvements

## License

MIT License - see LICENSE file for details.

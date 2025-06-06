# ghsecrets

A CLI tool to manage GitHub Secrets with automatic cloud backup to AWS Secrets Manager or GCP Secret Manager.

## Features

- Push secrets to GitHub repository secrets
- Automatically backup secrets to AWS Secrets Manager or GCP Secret Manager
- Configuration file support for default settings
- Secure encryption using GitHub's public key

## Installation

```bash
go install github.com/tom-023/ghsecrets@latest
```

Or build from source:

```bash
git clone https://github.com/tom-023/ghsecrets.git
cd ghsecrets
go build -o ghsecrets main.go
```

## Configuration

Create a configuration file `ghsecrets.yaml` in your current directory:

```yaml
# GitHub configuration
github:
  # GitHub personal access token (can also use GITHUB_TOKEN env var)
  # token: your-github-token

  # Default repository owner
  owner: your-username

  # Default repository name
  repo: your-repo

# AWS configuration
aws:
  # AWS region for Secrets Manager
  region: us-east-1
  
  # Secret name in AWS Secrets Manager
  # All GitHub secrets will be stored in this single secret as JSON
  # NOTE: This secret must be created in AWS first (can be empty JSON: {})
  secret_name: github-secrets-backup

# GCP configuration
gcp:
  # GCP project ID
  project: your-project-id

  # Path to service account credentials JSON file (optional)
  # credentials_path: /path/to/service-account.json
```

## Authentication

### GitHub
You can authenticate using one of these methods (in order of priority):

1. **Configuration file** - Set token in `ghsecrets.yaml`
2. **Environment variable** - Set `GITHUB_TOKEN`
3. **GitHub CLI** - Login with `gh auth login` (recommended)

```bash
# Option 1: Use gh CLI (recommended)
gh auth login

# Option 2: Use environment variable
export GITHUB_TOKEN=your-github-token

# Option 3: Add to config file
# Copy ghsecrets.yaml.example to ghsecrets.yaml and edit
cp ghsecrets.yaml.example ghsecrets.yaml
```

### AWS
Configure AWS credentials using standard AWS credential chain:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file (`~/.aws/credentials`)
- AWS SSO profiles (`~/.aws/config`)
- IAM role (when running on EC2/ECS/Lambda)

You can specify which AWS profile to use:
```bash
# Use a specific profile from ~/.aws/credentials or ~/.aws/config
ghsecrets push -k KEY -v VALUE -b aws --aws-profile production

# Use AWS SSO profile (login first with: aws sso login --profile blog-profile)
ghsecrets push -k KEY -v VALUE -b aws --aws-profile blog-profile

# Or set in config file
# aws:
#   profile: production
```

#### AWS SSO Configuration Example
If you have SSO configured in `~/.aws/config`:
```ini
[profile blog-profile]
sso_start_url = https://d-0000000000.awsapps.com/start
sso_region = ap-northeast-1
sso_account_id = 000000000000
sso_role_name = AWSAdministratorAccess
region = ap-northeast-1
```

Make sure to login first:
```bash
aws sso login --profile blog-profile
ghsecrets push -k SECRET -v "value" -b aws --aws-profile blog-profile
```

### GCP
Configure GCP credentials:
- Service account key file (specified in config)
- Application Default Credentials (`gcloud auth application-default login`)
- GCE/GKE metadata service

## Usage

### Push a secret to GitHub only

```bash
ghsecrets push -k API_KEY -v "secret-value"
```

### Push a secret with AWS backup

First, create the secret in AWS Secrets Manager:
```bash
# Create an empty secret in AWS
aws secretsmanager create-secret --name github-secrets-backup --secret-string '{}'
```

Then push secrets:
```bash
ghsecrets push -k DATABASE_URL -v "postgres://..." -b aws
```

This will (in order):
1. Store the key-value pair in AWS Secrets Manager under the configured `secret_name` as JSON
2. Create/update the GitHub secret `DATABASE_URL` (only if backup succeeds)

Example AWS Secrets Manager content:
```json
{
  "DATABASE_URL": "postgres://...",
  "API_KEY": "sk-...",
  "OTHER_SECRET": "value"
}
```

**Note**: The AWS secret must exist before using ghsecrets. If it doesn't exist, you'll get an error message.

### Push a secret with GCP backup

```bash
ghsecrets push -k TOKEN -v "token123" -b gcp
```

### Override repository settings

```bash
ghsecrets push -k SECRET_KEY -v "value" -o owner -r repo -b aws
```

## Command Reference

### `ghsecrets push`

Push a secret to GitHub and optionally backup to cloud.

**Flags:**
- `-k, --key`: Secret key name (required)
- `-v, --value`: Secret value (required)
- `-b, --backup`: Backup destination: `aws`, `gcp`, or `none`
- `-o, --owner`: GitHub repository owner
- `-r, --repo`: GitHub repository name
- `--aws-region`: AWS region for Secrets Manager (default: us-east-1)
- `--aws-profile`: AWS profile to use from ~/.aws/credentials
- `--gcp-project`: GCP project ID

### `ghsecrets restore`

Restore all GitHub Secrets from backup (AWS Secrets Manager or GCP Secret Manager).

**Usage:**
```bash
# Restore all secrets from AWS to GitHub
ghsecrets restore -b aws

# Restore to a specific repository
ghsecrets restore -b aws -o owner -r repo

# Use a specific AWS profile
ghsecrets restore -b aws --aws-profile production

# Future: Restore from GCP (not yet implemented)
# ghsecrets restore -b gcp
```

**Flags:**
- `-b, --backup`: Backup source to restore from: `aws` or `gcp` (required)
- `-o, --owner`: GitHub repository owner
- `-r, --repo`: GitHub repository name
- `--aws-region`: AWS region for Secrets Manager (default: us-east-1)
- `--aws-profile`: AWS profile to use from ~/.aws/credentials
- `--gcp-project`: GCP project ID (for future GCP support)

This command will:
1. Read all key-value pairs from the specified backup source
2. Create or update each secret in the specified GitHub repository
3. Report the number of successfully restored secrets

## Security

- Secrets are encrypted using GitHub's repository public key before transmission
- Cloud backups use the respective service's encryption at rest
- Never commit secrets directly to your repository

## License

MIT License - see LICENSE file for details
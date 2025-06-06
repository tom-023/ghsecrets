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
- IAM role (when running on EC2/ECS/Lambda)

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

```bash
ghsecrets push -k DATABASE_URL -v "postgres://..." -b aws
```

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
- `--gcp-project`: GCP project ID

## Security

- Secrets are encrypted using GitHub's repository public key before transmission
- Cloud backups use the respective service's encryption at rest
- Never commit secrets directly to your repository

## License

MIT License - see LICENSE file for details
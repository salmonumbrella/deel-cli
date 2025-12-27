# 👥 Deel CLI — Workforce management in your terminal.

Manage contracts, people, payroll, time off, and more from the command line.

## Features

- **Authentication** - authenticate once, tokens refresh indefinitely
- **Benefits** - view employee benefits by country
- **Background checks** - manage screening and verification
- **Calculators** - estimate employer costs and take-home pay
- **Compliance** - access documents, templates, and validations
- **Contracts** - list, view, and manage contract details
- **Immigration** - track visa cases and requirements
- **Invoices** - manage invoices and adjustments
- **IT assets** - track hardware and orders
- **Multiple accounts** - switch between Deel accounts
- **Onboarding** - monitor employee onboarding status
- **Organization** - view org structure and legal entities
- **Payroll** - access payslips, payments, and receipts
- **People** - list workers and custom fields
- **Shifts** - view shift schedules and rates
- **Teams** - manage team structures
- **Time off** - create and track PTO requests

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/deel-cli
```

### Go Install

```bash
go install github.com/salmonumbrella/deel-cli/cmd/deel@latest
```

### From Source

```bash
git clone https://github.com/salmonumbrella/deel-cli.git
cd deel-cli
go build -o deel ./cmd/deel
sudo mv deel /usr/local/bin/
```

### Pre-built Binaries

Download from the [Releases](https://github.com/salmonumbrella/deel-cli/releases) page.

## Quick Start

### 1. Get a Deel API Token

1. Go to [Deel API Settings](https://app.deel.com/settings/api)
2. Create a new Personal Access Token
3. Copy the token

### 2. Authenticate

Choose one of two methods:

**Browser:**
```bash
deel auth login
```

**Terminal:**
```bash
deel auth add my-account
# You'll be prompted securely for your token
```

### 3. Test Authentication

```bash
deel auth test --account my-account
```

### 4. Start Using

```bash
deel people list --account my-account
deel contracts list --account my-account
```

## Configuration

### Account Selection

Specify the account using either a flag or environment variable:

```bash
# Via flag
deel people list --account my-account

# Via environment
export DEEL_ACCOUNT=my-account
deel people list
```

### Environment Variables

- `DEEL_TOKEN` - Direct API token (bypasses keychain)
- `DEEL_ACCOUNT` - Default account name to use
- `DEEL_OUTPUT` - Output format: `text` (default) or `json`
- `DEEL_COLOR` - Color mode: `auto` (default), `always`, or `never`
- `NO_COLOR` - Set to any value to disable colors (standard convention)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Commands

### Authentication

```bash
deel auth login                      # Authenticate via browser (recommended)
deel auth add <name>                 # Add credentials manually (prompts securely)
deel auth list                       # List configured accounts
deel auth remove <name>              # Remove account
deel auth test [--account <name>]    # Test credentials
```

### People

```bash
deel people list [--limit <n>] [--cursor <token>]    # List all people
deel people get <hris-profile-id>                    # Get person details
deel people search --email <email>                   # Find person by email
deel people custom-fields list                       # List custom fields
deel people custom-fields get <field-id>             # Get custom field details
```

### Contracts

```bash
deel contracts list [--limit <n>]            # List all contracts
deel contracts get <contract-id>             # Get contract details
deel contracts amendments <contract-id>      # List contract amendments
deel contracts payment-dates <contract-id>   # Get payment schedule
```

### Time Off

```bash
deel time-off list [--contract-id <id>] [--status <status>]    # List requests
deel time-off policies [--contract-id <id>]                    # List policies
deel time-off create --contract-id <id> --type <type> \
    --start-date <date> --end-date <date> [--reason <text>]    # Create request
deel time-off cancel <request-id> [--reason <text>]            # Cancel request
```

Aliases: `timeoff`, `pto`

### Payroll

```bash
deel payroll payslips eor --contract-id <id> [--year <yyyy>] [--month <mm>]   # EOR payslips
deel payroll payslips gp --contract-id <id> [--year <yyyy>] [--month <mm>]    # GP payslips
deel payroll payments --year <yyyy> --month <mm>                               # Payment breakdown
deel payroll receipts [--year <yyyy>] [--month <mm>]                           # Payment receipts
```

### Invoices

```bash
deel invoices list [--limit <n>]                           # List invoices
deel invoices get <invoice-id>                             # Get invoice details
deel invoices adjustments list <contract-id>               # List adjustments
deel invoices adjustments create <contract-id> \
    --amount <n> --reason <text> [--type <type>]           # Create adjustment
```

### Payments

```bash
deel payments off-cycle list                               # List off-cycle payments
deel payments off-cycle create --contract-id <id> \
    --amount <n> --reason <text>                           # Create off-cycle payment
```

### Benefits

```bash
deel benefits list --country-code <xx>       # List by country
deel benefits employee --contract-id <id>    # Get employee benefits
```

### IT Assets

```bash
deel it assets [--status <status>]    # List IT assets
deel it orders                        # List IT orders
deel it policies                      # List hardware policies
```

### Teams

```bash
deel teams list              # List all teams
deel teams get <team-id>     # Get team details
```

### Organization

```bash
deel org get                      # Get organization info
deel org structures               # Get org structures
deel org entities [--limit <n>]   # List legal entities
```

### Onboarding

```bash
deel onboarding list [--status <status>]    # List onboarding employees
deel onboarding get <contract-id>           # Get onboarding details
```

### Compliance

```bash
deel compliance docs [--contract-id <id>] [--status <status>]    # List documents
deel compliance templates [--country <xx>]                        # List templates
deel compliance validations --contract-id <id>                    # Get validations
```

### Background Checks

```bash
deel background-checks options --contract-id <id>    # List options
deel background-checks list --contract-id <id>       # List by contract
```

Aliases: `bgcheck`, `checks`

### Immigration

```bash
deel immigration cases <contract-id>                              # Get case details
deel immigration docs <case-id>                                   # List documents
deel immigration visa-types --country <xx>                        # List visa types
deel immigration check --country <xx> --nationality <xx>          # Check requirements
```

Alias: `visa`

### ATS (Applicant Tracking)

```bash
deel ats offers [--status <status>] [--limit <n>]    # List offers
```

### Shifts

```bash
deel shifts list --contract-id <id> [--from <date>] [--to <date>]    # List shifts
deel shifts rates --contract-id <id>                                  # List shift rates
```

### Calculators

```bash
deel calc cost --country <xx> --salary <n> [--currency <xxx>]          # Calculate employer cost
deel calc take-home --country <xx> --salary <n> [--currency <xxx>]     # Calculate take-home
deel calc salary-histogram --country <xx> --job-title <text>           # Get salary data
```

Alias: `calculator`

### Tokens

```bash
deel tokens create --contract-id <id> [--expires-in <duration>]    # Create worker token
```

## Output Formats

### Text

Human-readable tables with colors and formatting:

```bash
$ deel people list
NAME              EMAIL                    STATUS    CONTRACT TYPE
John Doe          john@example.com         active    employee
Jane Smith        jane@example.com         active    contractor
```

### JSON

Machine-readable output:

```bash
$ deel people list --output json
{
  "data": [
    {"name": "John Doe", "email": "john@example.com", "status": "active"}
  ]
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### List active workers in JSON

```bash
deel people list --output json | jq '.data[] | select(.status == "active")'
```

### Export contracts to CSV

```bash
deel contracts list --output json | \
  jq -r '.data[] | [.id, .worker_name, .type, .status] | @csv'
```

### Check time-off balance

```bash
CONTRACT_ID="abc123"
deel time-off policies --contract-id $CONTRACT_ID
```

### Calculate hiring costs

```bash
deel calc cost --country US --salary 80000 --currency USD
```

### Switch between accounts

```bash
# Check production account
deel people list --account prod

# Check sandbox account
deel people list --account sandbox

# Or set default
export DEEL_ACCOUNT=prod
deel people list
```

### CI/CD Usage

For CI/CD pipelines, use the `DEEL_TOKEN` environment variable:

```bash
export DEEL_TOKEN="your-api-token"
deel people list
```

### JQ Filtering

Filter JSON output with JQ expressions:

```bash
# Get only active contracts
deel contracts list --output json | jq '.data[] | select(.status == "active")'

# Extract contract IDs
deel contracts list --output json | jq '[.data[].id]'
```

## Global Flags

All commands support these flags:

- `--account <name>` - Account to use (overrides DEEL_ACCOUNT)
- `--output <format>` - Output format: `text` or `json` (default: text)
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--debug` - Enable debug output (shows API requests/responses)
- `--help` - Show help for any command
- `--version` - Show version information

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
deel completion bash > $(brew --prefix)/etc/bash_completion.d/deel

# Linux:
deel completion bash > /etc/bash_completion.d/deel

# Or source directly:
source <(deel completion bash)
```

### Zsh

```zsh
deel completion zsh > "${fpath[1]}/_deel"
```

### Fish

```fish
deel completion fish > ~/.config/fish/completions/deel.fish
```

### PowerShell

```powershell
deel completion powershell | Out-String | Invoke-Expression
```

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

## License

MIT

## Links

- [Deel API Documentation](https://developer.deel.com/)

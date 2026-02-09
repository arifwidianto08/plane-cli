# Plane CLI

A Go-based CLI tool for automating Plane.so work item management with fuzzy title matching and template-based descriptions.

## Features

- **Work Item CRUD**: Create, read, and update work items across multiple projects
- **Fuzzy Title Matching**: Find and update work items by approximate title matching
- **JSON Templates**: Standardized description templates with variable substitution
- **Multi-Project Support**: Work across different Plane projects
- **Interactive Mode**: Select from multiple matches or projects
- **Batch Operations**: Update multiple work items at once

## Installation

```bash
# Clone or download the repository
git clone <repo-url>
cd plane-automation

# Build the binary
go build -o plane-cli ./cmd/plane-cli/

# Or install globally
go install ./cmd/plane-cli/
```

## Quick Start

### 1. Initialize Configuration

```bash
./plane-cli init
```

This will:
- Create `.env` file for API credentials
- Create `config.yaml` for defaults
- Set up templates directory with examples

### 2. Configure Credentials

Edit `.env`:
```bash
PLANE_BASE_URL=https://plane.your-domain.com
PLANE_API_TOKEN=your_personal_access_token_here
```

### 3. List Your Projects

```bash
./plane-cli project list
```

### 4. Create a Work Item

```bash
./plane-cli create \
  --project my-project \
  --title "Admin tenant UI" \
  --template feature \
  --vars feature_name="Authentication Module"
```

### 5. Update by Fuzzy Title

```bash
./plane-cli update \
  --title-fuzzy "admin tenant" \
  --template feature \
  --interactive
```

## Commands

### Work Items

```bash
# Create new work item
plane-cli create --project <project> --title <title> [options]

# Update work items
plane-cli update --id <id> [options]                    # By ID
plane-cli update --title-fuzzy <pattern> [options]      # Fuzzy search

# List work items
plane-cli list --project <project> [filters]
```

### Projects

```bash
# List all projects
plane-cli project list

# Select project interactively
plane-cli project select
```

### Templates

```bash
# List templates
plane-cli template list

# Show template
plane-cli template show feature

# Create template
plane-cli template create my-template

# Delete template
plane-cli template delete my-template
```

## Examples

### Create Work Item with Template

```bash
plane-cli create \
  --project admin-panel \
  --title "User authentication module" \
  --template feature \
  --vars feature_name="Authentication" \
  --vars checklist_items='["Login form","Session management","Password reset"]' \
  --priority high \
  --assignee user@example.com
```

### Update Empty Descriptions

```bash
# Find all work items matching pattern and add descriptions
plane-cli update \
  --title-fuzzy "tenant admin" \
  --template feature \
  --auto \
  --min-score 70
```

### Interactive Mode

```bash
plane-cli update --title-fuzzy "api" --interactive

Found 4 matching work items:
  1. [92%] API integration module (ADM-123)
  2. [85%] Update API documentation (ADM-124)
  3. [78%] Fix API authentication bug (ADM-125)
  4. [65%] Create API client library (ADM-126)

Select items (comma-separated, 'all', or 'cancel'): 1,3
```

### Dry Run

```bash
plane-cli update --title-fuzzy "dashboard" --template feature --dry-run

DRY RUN - No changes will be made

Found 3 matching work items:
  [88%] Admin dashboard redesign (WEB-001)
    → Would update description using template "feature"
  
  [76%] Dashboard analytics widget (WEB-002)
    → Would update description using template "feature"
```

## Templates

Templates are stored in JSON format in the `templates/` directory.

### Template Format

```json
{
  "name": "feature",
  "description": "Feature development template",
  "content": "## Definition Of Done\n\n* [ ] {{feature_name}}\n{{#checklist_items}}  * [ ] {{item}}\n{{/checklist_items}}",
  "variables": ["feature_name", "checklist_items"]
}
```

### Variable Syntax

- Simple variables: `{{variable_name}}`
- Lists: `{{#list}} {{item}} {{/list}}`
- Conditionals: `{{#if condition}} ... {{/if}}`

### Default Templates

- **feature** - Definition of Done checklist for features
- **bug** - Bug report with reproduction steps
- **task** - Simple task with checklist

## Configuration

### Environment Variables (.env)

```bash
PLANE_BASE_URL=https://plane.your-domain.com
PLANE_API_TOKEN=your_token_here
```

### Config File (config.yaml)

```yaml
defaults:
  project: "my-project"
  state: "Backlog"
  priority: 2

templates:
  directory: "./templates"
  default: "feature"

fuzzy:
  min_score: 60
  max_results: 10
```

## API Requirements

This tool requires:
- Self-hosted Plane.so instance
- Personal Access Token (from Plane settings)
- API access enabled on your instance

## Development

```bash
# Run tests
go test ./...

# Build locally
go build -o plane-cli ./cmd/plane-cli/

# Install dependencies
go mod tidy
```

## License

MIT

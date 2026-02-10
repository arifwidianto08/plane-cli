# Plane CLI

A powerful Go-based CLI tool for managing Plane.so projects with interactive workflows, bulk operations, and comprehensive project management features.

## Features

- **Interactive Mode**: User-friendly menu-driven interface with arrow key navigation
- **Work Item Management**: Create, read, update, and delete work items
- **Bulk Operations**: Update multiple work items simultaneously
- **Fuzzy Title Matching**: Find work items by approximate title matching
- **Project Management**: Manage modules, labels, and pages
- **JSON Templates**: Standardized description templates with variable substitution
- **Configuration Wizard**: Interactive setup and configuration management
- **Multi-Project Support**: Work across different Plane projects

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd plane-cli

# Build the binary
go build -o plane-cli ./cmd/plane-cli/

# Or install globally
go install ./cmd/plane-cli/
```

## Quick Start

### 1. Configure the CLI

```bash
# Interactive configuration (recommended)
./plane-cli configure

# Or view current configuration
./plane-cli configure --show
```

The configuration wizard will prompt you for:
- **Plane Base URL**: Your Plane instance URL (e.g., `https://project.your-domain.com`)
- **API Token**: Your personal access token from Plane settings
- **Workspace**: Your workspace slug

Configuration is saved to `.env` file in the current directory.

### 2. Launch Interactive Mode

```bash
./plane-cli interactive
```

This opens the main menu with all features:
- Work Items (single update)
- Bulk Update (multiple work items)
- Modules
- Labels  
- Pages

### 3. List Your Projects

```bash
./plane-cli project list
```

## Commands Reference

### Interactive Mode

```bash
# Main interactive menu - access all features
plane-cli interactive

# Interactive work item update
plane-cli interactive-update

# Interactive module management
plane-cli module interactive

# Interactive label management
plane-cli label interactive

# Interactive page management
plane-cli page interactive
```

### Configuration

```bash
# Interactive configuration setup
plane-cli configure

# View current configuration (token is masked)
plane-cli configure --show
```

### Work Items

```bash
# Create new work item
plane-cli create \
  --project <project-id> \
  --title "Work item title" \
  [--description "Description text"] \
  [--description-file path/to/file.md] \
  [--state "Backlog"] \
  [--priority high] \
  [--assignees user-id-1,user-id-2]

# Update work item by ID
plane-cli update \
  --id <work-item-id> \
  --project <project-id> \
  [--title "New title"] \
  [--description-file update.md] \
  [--state "In Progress"] \
  [--assignees user-id-1]

# Update by fuzzy title matching
plane-cli update \
  --title-fuzzy "search pattern" \
  --project <project-id> \
  [--template feature] \
  [--interactive] \
  [--auto]

# List work items
plane-cli list --project <project-id> [options]
  [--state "In Progress"]
  [--priority high]
  [--limit 50]
```

### Bulk Update

```bash
# Interactive bulk update (select multiple items with SPACE)
plane-cli bulk-update --project <project-id>

# Bulk update by search pattern
plane-cli bulk-update \
  --project <project-id> \
  --search "BE" \
  --state "In Progress" \
  --assignees user-id-1,user-id-2

# Bulk set estimate points
plane-cli bulk-update \
  --project <project-id> \
  --search "SaaS" \
  --estimate 5

# Bulk add labels (merges with existing)
plane-cli bulk-update \
  --project <project-id> \
  --labels label-1,label-2

# Dry run to preview changes
plane-cli bulk-update \
  --project <project-id> \
  --search "API" \
  --state "Done" \
  --dry-run
```

### Modules

```bash
# List all modules
plane-cli module list --project <project-id>

# Create module
plane-cli module create \
  --project <project-id> \
  --name "Frontend" \
  [--description "Module description"] \
  [--status backlog]

# Update module
plane-cli module update \
  --project <project-id> \
  --id <module-id> \
  [--name "New Name"] \
  [--status started]

# Delete module
plane-cli module delete \
  --project <project-id> \
  --id <module-id>

# Interactive module management
plane-cli module interactive
```

### Labels

```bash
# List all labels
plane-cli label list --project <project-id>

# Create label
plane-cli label create \
  --project <project-id> \
  --name "bug" \
  [--color "#ff0000"]

# Update label
plane-cli label update \
  --project <project-id> \
  --id <label-id> \
  [--name "Bug"] \
  [--color "#ff0000"]

# Delete label
plane-cli label delete \
  --project <project-id> \
  --id <label-id>

# Interactive label management
plane-cli label interactive
```

### Pages

```bash
# List all pages
plane-cli page list --project <project-id>

# Create page from file
plane-cli page create \
  --project <project-id> \
  --name "Documentation" \
  --description-file docs.md \
  [--access public]

# Update page
plane-cli page update \
  --project <project-id> \
  --id <page-id> \
  [--description-file updated.md]

# Delete page
plane-cli page delete \
  --project <project-id> \
  --id <page-id>

# Interactive page management
plane-cli page interactive
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

## Interactive Mode Examples

### Single Work Item Update

```bash
plane-cli interactive-update

📁 Step 1: Select a Project
> GloryX (ERPGLX)

🔍 Step 2: Find Work Item
? Enter search term: API

Found 12 match(es):
? Select work item:
> [61] [BE] SaaS - Auth (Score: 95%)
  [32] [BE] API Integration (Score: 90%)

✏️  Step 3: What would you like to update?
? Select an option:
> Description
  Title
  State
  Priority
  Assignees
  Estimate Points
  Module
```

### Bulk Update Multiple Items

```bash
plane-cli bulk-update --project c20fcc54-c675-47c4-85db-a4acdde3c9e1

📥 Fetching work items from project 'GloryX'...
Found 45 work items. Select which ones to update:

? Select work items (SPACE to select, ENTER to confirm):
> [ ] [61] [BE] SaaS - Auth
  [ ] [32] [BE] Setup Project
  [x] [62] [BE] Tenant - Auth
  [x] [56] [BE] SaaS - Add Ons
  [x] [57] [BE] SaaS - Invoice

✓ Selected 3 work items

What would you like to update?
> Assignees
  Estimate Points
  Labels
  Module
  State
  Priority

👥 Update Assignees
? What would you like to do?
> Add assignees to existing ones
  Replace all assignees
  Clear all assignees

? Select assignees to add:
> [x] John Doe (john@example.com)
  [ ] Jane Smith (jane@example.com)
  [x] Bob Wilson (bob@example.com)

📋 Bulk Update Preview:
----------------------------------------------------------------------
Project: GloryX
Work items to update: 3

Selected work items:
  • [62] [BE] Tenant - Auth
  • [56] [BE] SaaS - Add Ons
  • [57] [BE] SaaS - Invoice

Updates to apply:
   → Assignees: 2 selected

? Apply these updates to all selected work items? Yes

🔄 Updating 3 work items...

  ✅ Updated: [62] [BE] Tenant - Auth
  ✅ Updated: [56] [BE] SaaS - Add Ons
  ✅ Updated: [57] [BE] SaaS - Invoice

----------------------------------------------------------------------
✅ Completed: 3/3 work items updated successfully
```

## Configuration

### Environment Variables (.env)

```bash
PLANE_BASE_URL=https://project.your-domain.com
PLANE_API_TOKEN=plane_api_your_token_here
PLANE_WORKSPACE=your-workspace-slug
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

## Features in Detail

### Fuzzy Title Matching

The CLI uses fuzzy matching to find work items by approximate titles:
- Tolerates typos and partial matches
- Configurable minimum score (0-100)
- Falls back to substring matching for short patterns

```bash
# Will find "[BE] API Integration" when searching "api"
plane-cli update --title-fuzzy "api" --project my-project
```

### Bulk Operations

Update multiple work items simultaneously:
- Select items with SPACE key in interactive mode
- Update assignees, estimates, labels, module, state, priority
- Preview changes before applying
- Shows success/failure for each item

### Arrow Key Navigation

All interactive prompts support:
- **↑/↓ arrows**: Navigate through options
- **SPACE**: Select/deselect (for multi-select)
- **ENTER**: Confirm selection
- **Ctrl+C**: Cancel operation

## API Requirements

This tool requires:
- Plane.so instance (self-hosted or cloud)
- Personal Access Token (from Plane workspace settings → API)
- API access enabled on your instance

To get your API token:
1. Go to your Plane workspace
2. Navigate to Settings → API
3. Generate a new personal access token

## Development

```bash
# Run tests
go test ./...

# Build locally
go build -o plane-cli ./cmd/plane-cli/

# Install dependencies
go mod tidy

# Run with debug output
./plane-cli --debug <command>
```

## License

MIT

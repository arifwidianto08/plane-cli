# Plane.so CLI Automation Tool - Implementation Plan

## Overview

A Go-based CLI tool to automate Plane.so work item management for self-hosted instances. Features fuzzy title-based search, JSON templates for standardized descriptions (Definition of Done format), and comprehensive CRUD operations.

---

## Features

### Core Features
- **Personal Access Token (PAT) Authentication** - Secure API access
- **Work Item CRUD** - Create, read, update work items across multiple projects
- **Fuzzy Title Matching** - Find work items by approximate title matching
- **JSON Template System** - Standardized description templates with variable substitution
- **Multi-Project Support** - Work across different Plane projects
- **Batch Operations** - Update multiple work items at once
- **Interactive Mode** - Select from multiple fuzzy matches

### CLI Commands

```bash
# Initialize configuration
plane-cli init

# Create new work item
plane-cli create --project "my-project" \
                 --title "Admin tenant UI" \
                 --template feature \
                 --vars feature_area="auth"

# Update by fuzzy title matching
plane-cli update --title-fuzzy "adm tennt ui" \
                 --template feature \
                 --auto

# Interactive fuzzy search
plane-cli update --title-fuzzy "api" --interactive

# List work items
plane-cli list --project "my-project" --state "In Progress"

# Template management
plane-cli template list
plane-cli template show feature
plane-cli template create bug

# Dry-run to preview changes
plane-cli update --title-fuzzy "dashboard" --dry-run
```

### Template System

**JSON Template Format:**

```json
{
  "name": "feature",
  "description": "Feature development with DOD checklist",
  "content": "## Definition Of Done\n\n* [ ] {{feature_area}}\n{{#checklist_items}}  * [ ] {{item}}\n{{/checklist_items}}\n\n## Notes\n{{additional_notes}}",
  "variables": ["feature_area", "checklist_items", "additional_notes"]
}
```

**Example Usage:**

```bash
plane-cli create --project "my-project" \
  --title "Admin tenant UI" \
  --template feature \
  --vars feature_area="Authentication Module" \
  --vars additional_notes="Priority: High"
```

---

## Technical Architecture

### Project Structure

```
plane-automation/
├── cmd/
│   └── plane-cli/
│       └── main.go                    # Application entry point
├── internal/
│   ├── plane/
│   │   ├── client.go                  # HTTP client with PAT auth
│   │   ├── workitems.go               # Work item API operations
│   │   ├── projects.go                # Project API operations
│   │   └── types.go                   # Data structures
│   ├── fuzzy/
│   │   └── matcher.go                 # Fuzzy string matching logic
│   ├── templates/
│   │   ├── loader.go                  # JSON template loader
│   │   ├── parser.go                  # Variable substitution
│   │   └── manager.go                 # Template CRUD operations
│   ├── config/
│   │   └── config.go                  # Configuration management
│   └── commands/
│       ├── root.go                    # Root command
│       ├── create.go                  # Create command
│       ├── update.go                  # Update command
│       ├── list.go                    # List command
│       ├── template.go                # Template commands
│       └── init.go                    # Init command
├── templates/                         # Default templates
│   ├── feature.json
│   ├── bug.json
│   └── custom/                        # User custom templates
├── .env.example                       # Environment template
├── config.yaml.example               # Config template
├── go.mod
├── go.sum
└── PLAN.md                           # This file
```

### Dependencies

**Core:**
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/sahilm/fuzzy` - Fuzzy string matching
- `github.com/joho/godotenv` - Environment variable loading

**Standard Library:**
- `net/http` - HTTP client
- `encoding/json` - JSON handling
- `text/template` - Template variable substitution

---

## Configuration

### Environment Variables (.env)

```bash
# Required
PLANE_BASE_URL=https://plane.your-domain.com
PLANE_API_TOKEN=your_personal_access_token

# Optional
PLANE_DEFAULT_PROJECT=my-project
PLANE_REQUEST_TIMEOUT=30
```

### Config File (config.yaml)

```yaml
# Default settings
defaults:
  project: "main-project"
  state: "Backlog"
  assignee: "user@example.com"
  priority: 1  # 0=Urgent, 1=High, 2=Medium, 3=Low

# Project mappings (shortcuts)
projects:
  admin: "admin-panel-project"
  api: "backend-api-project"
  web: "frontend-web-project"

# Template paths
templates:
  directory: "./templates"
  default: "feature"

# Fuzzy matching settings
fuzzy:
  min_score: 60  # Minimum match percentage (0-100)
  max_results: 10  # Max results to show
  case_sensitive: false
```

---

## API Integration

### Plane.so REST API Endpoints

**Authentication:**
- Header: `X-API-Key: <token>`

**Work Items:**
- `GET /api/v1/workspaces/{workspace}/projects/{project}/issues/` - List issues
- `POST /api/v1/workspaces/{workspace}/projects/{project}/issues/` - Create issue
- `PATCH /api/v1/workspaces/{workspace}/projects/{project}/issues/{issue}/` - Update issue
- `GET /api/v1/workspaces/{workspace}/projects/{project}/issues/{issue}/` - Get issue

**Supporting APIs:**
- Projects, States, Labels, Modules, Cycles, Assignees

### Work Item Fields Supported

- `name` (title)
- `description` (rich text/html)
- `state` (workflow state)
- `priority` (0-3)
- `assignees` (array of user IDs)
- `labels` (array of label IDs)
- `start_date`, `target_date`
- `estimate_point`
- `module` (custom field)
- `cycle` (cycle ID)
- `parent` (parent issue ID)

---

## Implementation Phases

### Phase 1: Foundation
- [x] Project initialization (go mod init)
- [x] Basic CLI structure with Cobra
- [x] Configuration management (Viper + godotenv)
- [x] Plane API client with authentication
- [ ] Basic work item CRUD operations

### Phase 2: Fuzzy Matching
- [ ] Fuzzy string matcher implementation
- [ ] Title-based search functionality
- [ ] Multiple match handling (interactive/auto)
- [ ] Score threshold configuration

### Phase 3: Templates
- [ ] JSON template loader
- [ ] Variable substitution engine
- [ ] Template management commands
- [ ] Default templates (feature, bug)

### Phase 4: Advanced Features
- [ ] Batch operations
- [ ] Dry-run mode
- [ ] Better error handling and validation
- [ ] Progress indicators
- [ ] Export/import functionality

### Phase 5: Polish
- [ ] Comprehensive error messages
- [ ] CLI help documentation
- [ ] Configuration validation
- [ ] Testing

---

## Usage Examples

### Example 1: Create Work Item with Template

```bash
plane-cli create \
  --project "admin-panel" \
  --title "User authentication module" \
  --template feature \
  --vars feature_name="Authentication" \
  --vars checklist_items='["Login form","Session management","Password reset","OAuth integration"]' \
  --assignee "john@example.com" \
  --priority high
```

### Example 2: Update Empty Descriptions

```bash
# Find all work items with "tenant" in title and add description
plane-cli update \
  --title-fuzzy "tenant admin" \
  --template feature \
  --auto \
  --min-score 70
```

### Example 3: Interactive Mode

```bash
$ plane-cli update --title-fuzzy "api" --interactive

Found 4 matching work items:
  1. [92%] API integration module (ADM-123)
  2. [85%] Update API documentation (ADM-124)
  3. [78%] Fix API authentication bug (ADM-125)
  4. [65%] Create API client library (ADM-126)

Select items (comma-separated, 'all', 'none', or numbers): 1,3

Applying template "feature" to 2 items...
✓ Updated ADM-123: API integration module
✓ Updated ADM-125: Fix API authentication bug
```

### Example 4: Dry Run

```bash
$ plane-cli update --title-fuzzy "dashboard" --template feature --dry-run

DRY RUN - No changes will be made

Found 3 matching work items:
  [88%] Admin dashboard redesign (WEB-001)
    → Would update description using template "feature"
  
  [76%] Dashboard analytics widget (WEB-002)
    → Would update description using template "feature"
  
  [72%] Fix dashboard loading (WEB-003)
    → Already has description, would skip

Run without --dry-run to apply changes.
```

---

## Notes

### Plane.so API Considerations

1. **Self-hosted Instance**: Tool assumes self-hosted Plane at custom domain
2. **API Version**: Using v1 API endpoints
3. **Rate Limiting**: May need to implement rate limiting for batch operations
4. **Authentication**: PAT (Personal Access Token) required from Plane settings

### Template Variables

Support for various variable types:
- Simple strings: `{{variable_name}}`
- Arrays/lists: `{{#items}} {{name}} {{/items}}`
- Conditionals: `{{#if condition}} ... {{/if}}`
- Default values: `{{variable|default:"N/A"}}`

### Error Handling

- Clear error messages for API failures
- Validation of required fields
- Handling of fuzzy match edge cases
- Network timeout handling

---

## Future Enhancements

- [ ] Webhook integration for real-time updates
- [ ] Sync with external tools (GitHub, GitLab)
- [ ] Import/Export work items
- [ ] GUI/web interface option
- [ ] CI/CD integration
- [ ] AI-powered template suggestions

---

## License

MIT License - Open source

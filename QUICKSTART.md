# Plane CLI - Quick Start Guide

## 🚀 What Was Built

A complete Go CLI tool for automating Plane.so work item management with these features:

### ✅ Core Features Implemented

1. **Work Item Management**
   - Create work items with templates
   - Update by ID or fuzzy title matching
   - List with filtering
   - All fields supported (state, priority, assignees, labels, dates, estimates, modules, cycles, parent)

2. **Fuzzy Title Matching**
   - Find work items by approximate title
   - Configurable match threshold (0-100)
   - Interactive selection from multiple matches
   - Batch operations

3. **Template System**
   - JSON-based templates
   - Variable substitution (e.g., `{{feature_name}}`)
   - Default templates: feature, bug, task
   - Create custom templates

4. **Project Management**
   - List all projects
   - Interactive project selection
   - Multi-project support

5. **Configuration**
   - `.env` for API credentials
   - `config.yaml` for defaults
   - Environment-based setup

## 📁 Project Structure

```
plane-automation/
├── plane-cli              # Compiled binary
├── cmd/plane-cli/
│   └── main.go           # Entry point
├── internal/
│   ├── commands/         # CLI commands
│   ├── plane/            # API client
│   ├── templates/        # Template engine
│   ├── fuzzy/            # Fuzzy matching
│   └── config/           # Configuration
├── templates/            # Template files
│   ├── feature.json
│   ├── bug.json
│   └── task.json
├── .env.example
├── config.yaml.example
├── PLAN.md              # Implementation plan
├── README.md            # Full documentation
└── QUICKSTART.md        # This file
```

## 🏃 Quick Start

### 1. Configure Your Plane Instance

Copy example files and configure:

```bash
# Copy examples
cp .env.example .env
cp config.yaml.example config.yaml
```

Edit `.env`:
```bash
PLANE_BASE_URL=https://plane.your-domain.com
PLANE_API_TOKEN=your_personal_access_token_here
```

Get your API token from Plane:
- Go to your Plane instance settings
- Navigate to "Developer" or "API" section
- Generate a Personal Access Token

### 2. Initialize (Optional)

```bash
./plane-cli init
```

This interactively sets up your configuration.

### 3. Test Connection

List your projects:

```bash
./plane-cli project list
```

If this shows your projects, you're connected!

### 4. Create Your First Work Item

```bash
./plane-cli create \
  --project your-project-id \
  --title "Test Work Item" \
  --template feature \
  --vars feature_name="Test Feature"
```

## 📖 Common Use Cases

### Use Case 1: Create Work Items with Templates

```bash
# Create a feature with full DOD checklist
./plane-cli create \
  --project admin-panel \
  --title "User authentication module" \
  --template feature \
  --vars feature_name="Authentication System" \
  --vars modules='["Login form","Session management","Password reset","OAuth"]' \
  --vars acceptance_criteria='["User can login","Session persists","Password can be reset"]' \
  --vars notes="High priority - blocking other features" \
  --priority high \
  --state "Backlog"
```

### Use Case 2: Update Empty Descriptions (Your Use Case!)

You mentioned cards with only titles and no descriptions:

```bash
# Find all work items with "tenant" in title and add descriptions
./plane-cli update \
  --title-fuzzy "adm tenant" \
  --project admin-panel \
  --template feature \
  --auto \
  --min-score 70
```

With interactive selection:

```bash
./plane-cli update \
  --title-fuzzy "tenant" \
  --project admin-panel \
  --template feature \
  --interactive

# Output:
# Found 3 matching work items:
#   1. [85%] Admin tenant UI
#   2. [72%] Tenant management
#   3. [68%] Multi-tenant support
#
# Select items (comma-separated, 'all', or 'cancel'): 1,2
```

### Use Case 3: Batch Update Multiple Cards

```bash
# Update all matching work items
./plane-cli update \
  --title-fuzzy "API integration" \
  --project backend \
  --template feature \
  --auto \
  --vars feature_name="API Integration" \
  --state "In Progress"
```

### Use Case 4: Select Project Interactively

```bash
./plane-cli project select

# Shows list of projects, you pick one
# Then use it in commands:
./plane-cli list --project <selected-project>
```

## 🎨 Template Examples

### Your Template (Definition of Done Format)

Create `templates/admin-tenant.json`:

```json
{
  "name": "admin-tenant",
  "description": "Admin tenant UI development template",
  "content": "#### Definition Of Done\n\n* [ ] {{feature_name}}\n{{#modules}}  * [ ] {{.}}\n{{/modules}}\n\n#### Notes\n{{notes}}",
  "variables": ["feature_name", "modules", "notes"]
}
```

Use it:

```bash
./plane-cli create \
  --project admin-panel \
  --title "membuat tampilan di admin tenant" \
  --template admin-tenant \
  --vars feature_name="Auth" \
  --vars modules='["auth","onboarding","dashboard","list merchant","detail merchant","add merchant","edit merchant","list user management","subs and planning","activity logs admin tenant"]'
```

## 🔧 Advanced Usage

### Dry Run (Preview Changes)

```bash
./plane-cli update \
  --title-fuzzy "dashboard" \
  --template feature \
  --dry-run
```

### Custom Match Threshold

```bash
# Stricter matching (higher score required)
./plane-cli update \
  --title-fuzzy "auth" \
  --min-score 85 \
  --template feature
```

### Filter by State

```bash
./plane-cli list \
  --project my-project \
  --state "In Progress" \
  --priority high
```

## 🛠️ Development

### Build

```bash
go build -o plane-cli ./cmd/plane-cli/
```

### Run Tests

```bash
go test ./...
```

### Add Custom Templates

```bash
# Create new template
./plane-cli template create my-custom-template

# Or manually create JSON file in templates/
```

## 🔐 Security Notes

- Never commit `.env` file (contains API token)
- API token should have appropriate permissions
- Use `.env.example` for documentation
- Keep API tokens secure and rotate regularly

## 🐛 Troubleshooting

### "PLANE_BASE_URL is required"
- Create `.env` file with your Plane instance URL
- Or run `./plane-cli init`

### "API error 401"
- Check your API token is correct
- Token may be expired - generate a new one

### "No projects found"
- Verify workspace is correct
- Check API token has project access

### "Template not found"
- Check templates are in `./templates/` directory
- Use `./plane-cli template list` to verify

## 📚 Next Steps

1. **Create your custom templates** in `templates/` directory
2. **Set up config.yaml** with your common project mappings
3. **Test with dry-run** before batch operations
4. **Explore all commands** with `./plane-cli --help`

## 💡 Tips

- Use `--dry-run` to preview changes before applying
- Set `fuzzy.min_score` in config.yaml for your preferred match strictness
- Create project shortcuts in config.yaml for easier access
- Templates support Go's text/template syntax for advanced formatting

## 🎉 Ready to Use!

Your Plane CLI is ready! Start with:

```bash
./plane-cli init          # Setup configuration
./plane-cli project list  # See your projects
./plane-cli --help        # See all commands
```

Happy automating! 🚀

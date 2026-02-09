package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Template represents a work item description template
type Template struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Variables   []string `json:"variables"`
}

// Manager handles template loading and processing
type Manager struct {
	templatesDir string
	templates    map[string]*Template
}

// NewManager creates a new template manager
func NewManager(templatesDir string) (*Manager, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	mgr := &Manager{
		templatesDir: templatesDir,
		templates:    make(map[string]*Template),
	}

	// Load existing templates
	if err := mgr.LoadAll(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return mgr, nil
}

// LoadAll loads all templates from the templates directory
func (m *Manager) LoadAll() error {
	entries, err := os.ReadDir(m.templatesDir)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		// Remove .json extension
		templateName := strings.TrimSuffix(name, ".json")

		if err := m.Load(templateName); err != nil {
			// Log error but continue loading other templates
			fmt.Fprintf(os.Stderr, "Warning: failed to load template %s: %v\n", templateName, err)
		}
	}

	return nil
}

// Load loads a single template by name
func (m *Manager) Load(name string) error {
	filename := filepath.Join(m.templatesDir, name+".json")

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	var tmpl Template
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("failed to parse template JSON: %w", err)
	}

	// Validate
	if tmpl.Name == "" {
		tmpl.Name = name
	}
	if tmpl.Content == "" {
		return fmt.Errorf("template content cannot be empty")
	}

	m.templates[name] = &tmpl
	return nil
}

// Save saves a template to disk
func (m *Manager) Save(tmpl *Template) error {
	if tmpl.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if tmpl.Content == "" {
		return fmt.Errorf("template content is required")
	}

	filename := filepath.Join(m.templatesDir, tmpl.Name+".json")

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	m.templates[tmpl.Name] = tmpl
	return nil
}

// Get retrieves a template by name
func (m *Manager) Get(name string) (*Template, error) {
	tmpl, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return tmpl, nil
}

// List returns all available template names
func (m *Manager) List() []string {
	names := make([]string, 0, len(m.templates))
	for name := range m.templates {
		names = append(names, name)
	}
	return names
}

// Delete removes a template
func (m *Manager) Delete(name string) error {
	filename := filepath.Join(m.templatesDir, name+".json")

	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	delete(m.templates, name)
	return nil
}

// Render renders a template with the given variables
func (m *Manager) Render(name string, variables map[string]string) (string, error) {
	tmpl, err := m.Get(name)
	if err != nil {
		return "", err
	}

	return RenderTemplate(tmpl, variables)
}

// RenderTemplate renders a template with variables using Go's text/template
func RenderTemplate(tmpl *Template, variables map[string]string) (string, error) {
	// Create Go template
	t, err := template.New(tmpl.Name).Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := t.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// ValidateVariables checks if all required variables are provided
func (t *Template) ValidateVariables(variables map[string]string) []string {
	var missing []string
	for _, v := range t.Variables {
		if _, ok := variables[v]; !ok {
			missing = append(missing, v)
		}
	}
	return missing
}

// CreateDefaultTemplates creates default templates if they don't exist
func CreateDefaultTemplates(templatesDir string) error {
	templates := []Template{
		{
			Name:        "feature",
			Description: "Feature development with Definition of Done",
			Content: `## Definition Of Done

* [ ] {{feature_name}}
{{#modules}}  * [ ] {{name}}
{{/modules}}

## Acceptance Criteria
{{#acceptance_criteria}}* [ ] {{.}}
{{/acceptance_criteria}}

## Notes
{{notes}}`,
			Variables: []string{"feature_name", "modules", "acceptance_criteria", "notes"},
		},
		{
			Name:        "bug",
			Description: "Bug report template",
			Content: `## Bug Description
{{description}}

## Steps to Reproduce
{{#steps}}1. {{.}}
{{/steps}}

## Expected Behavior
{{expected}}

## Actual Behavior
{{actual}}

## Environment
- Version: {{version}}
- Browser: {{browser}}
- OS: {{os}}

## Notes
{{notes}}`,
			Variables: []string{"description", "steps", "expected", "actual", "version", "browser", "os", "notes"},
		},
		{
			Name:        "task",
			Description: "Simple task template",
			Content: `## Task Description
{{description}}

## Checklist
{{#checklist}}* [ ] {{.}}
{{/checklist}}

## Notes
{{notes}}`,
			Variables: []string{"description", "checklist", "notes"},
		},
	}

	mgr, err := NewManager(templatesDir)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		// Check if template already exists
		if _, err := mgr.Get(tmpl.Name); err == nil {
			continue // Skip if exists
		}

		if err := mgr.Save(&tmpl); err != nil {
			return fmt.Errorf("failed to create default template %s: %w", tmpl.Name, err)
		}
	}

	return nil
}

package fuzzy

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// Match represents a fuzzy match result
type Match struct {
	Item    interface{}
	Score   int
	Title   string
	ID      string
	Matched bool
}

// Matcher handles fuzzy string matching
type Matcher struct {
	minScore int
}

// NewMatcher creates a new fuzzy matcher with minimum score threshold
func NewMatcher(minScore int) *Matcher {
	if minScore < 0 {
		minScore = 0
	}
	if minScore > 100 {
		minScore = 100
	}
	return &Matcher{
		minScore: minScore,
	}
}

// MatchResult represents a single match with its score
type MatchResult struct {
	Index int
	Score int
	Item  interface{}
}

// FindMatches finds fuzzy matches in a list of items
// items should be a slice of structs with a Title or Name field
func (m *Matcher) FindMatches(pattern string, items []string) []MatchResult {
	if pattern == "" {
		return nil
	}

	// Normalize pattern
	pattern = strings.ToLower(strings.TrimSpace(pattern))

	// Find matches using sahilm/fuzzy
	matches := fuzzy.Find(pattern, items)

	// Convert to our format
	results := make([]MatchResult, 0, len(matches))
	for _, match := range matches {
		score := match.Score
		// Convert to 0-100 scale (fuzzy.Score returns raw score)
		// We'll normalize based on pattern length
		normalizedScore := normalizeScore(score, len(pattern))

		if normalizedScore >= m.minScore {
			results = append(results, MatchResult{
				Index: match.Index,
				Score: normalizedScore,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// FindBestMatch finds the best matching item
func (m *Matcher) FindBestMatch(pattern string, items []string) *MatchResult {
	matches := m.FindMatches(pattern, items)
	if len(matches) == 0 {
		return nil
	}
	return &matches[0]
}

// IsMatch checks if a pattern matches a string with minimum score
func (m *Matcher) IsMatch(pattern, text string) bool {
	matches := m.FindMatches(pattern, []string{text})
	return len(matches) > 0 && matches[0].Score >= m.minScore
}

// normalizeScore converts raw fuzzy score to 0-100 scale
// Adjusted to be more lenient with short patterns
func normalizeScore(rawScore, patternLength int) int {
	if rawScore <= 0 || patternLength <= 0 {
		return 0
	}

	// For short patterns (2-3 chars), be more lenient
	// Maximum possible score varies by pattern length
	var maxScore int
	switch {
	case patternLength <= 2:
		maxScore = patternLength * 2 // More lenient for very short patterns
	case patternLength <= 4:
		maxScore = patternLength * 3
	default:
		maxScore = patternLength * 4
	}

	percentage := (rawScore * 100) / maxScore

	// Boost score for short exact matches
	if patternLength <= 3 && rawScore >= patternLength {
		percentage += 30 // Boost short matches
	}

	if percentage > 100 {
		return 100
	}
	if percentage < 0 {
		return 0
	}
	return percentage
}

// SetMinScore updates the minimum score threshold
func (m *Matcher) SetMinScore(minScore int) {
	if minScore < 0 {
		minScore = 0
	}
	if minScore > 100 {
		minScore = 100
	}
	m.minScore = minScore
}

// GetMinScore returns the current minimum score threshold
func (m *Matcher) GetMinScore() int {
	return m.minScore
}

// FilterByScore filters matches by minimum score
func FilterByScore(matches []MatchResult, minScore int) []MatchResult {
	var filtered []MatchResult
	for _, match := range matches {
		if match.Score >= minScore {
			filtered = append(filtered, match)
		}
	}
	return filtered
}

// LimitResults limits the number of results
func LimitResults(matches []MatchResult, limit int) []MatchResult {
	if limit <= 0 || len(matches) <= limit {
		return matches
	}
	return matches[:limit]
}

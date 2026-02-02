package help

import (
	"sort"
	"strings"
	"unicode"
)

// SearchResult represents a search match.
type SearchResult struct {
	Topic   *Topic
	Section *Section // nil if topic-level match
	Score   float64
	Snippet string // Context around match
}

// searchIndex provides full-text search.
type searchIndex struct {
	topics map[string]*Topic          // topicID -> Topic
	index  map[string]map[string]bool // word -> set of topicIDs
}

// newSearchIndex creates a new empty search index.
func newSearchIndex() *searchIndex {
	return &searchIndex{
		topics: make(map[string]*Topic),
		index:  make(map[string]map[string]bool),
	}
}

// Add indexes a topic for searching.
func (i *searchIndex) Add(topic *Topic) {
	i.topics[topic.ID] = topic

	// Index title words with boost
	for _, word := range tokenize(topic.Title) {
		i.addToIndex(word, topic.ID)
	}

	// Index content words
	for _, word := range tokenize(topic.Content) {
		i.addToIndex(word, topic.ID)
	}

	// Index section titles and content
	for _, section := range topic.Sections {
		for _, word := range tokenize(section.Title) {
			i.addToIndex(word, topic.ID)
		}
		for _, word := range tokenize(section.Content) {
			i.addToIndex(word, topic.ID)
		}
	}

	// Index tags
	for _, tag := range topic.Tags {
		for _, word := range tokenize(tag) {
			i.addToIndex(word, topic.ID)
		}
	}
}

// addToIndex adds a word-to-topic mapping.
func (i *searchIndex) addToIndex(word, topicID string) {
	if i.index[word] == nil {
		i.index[word] = make(map[string]bool)
	}
	i.index[word][topicID] = true
}

// Search finds topics matching the query.
func (i *searchIndex) Search(query string) []*SearchResult {
	queryWords := tokenize(query)
	if len(queryWords) == 0 {
		return nil
	}

	// Track scores per topic
	scores := make(map[string]float64)

	for _, word := range queryWords {
		// Exact matches
		if topicIDs, ok := i.index[word]; ok {
			for topicID := range topicIDs {
				scores[topicID] += 1.0
			}
		}

		// Prefix matches (partial word matching)
		for indexWord, topicIDs := range i.index {
			if strings.HasPrefix(indexWord, word) && indexWord != word {
				for topicID := range topicIDs {
					scores[topicID] += 0.5 // Lower score for partial matches
				}
			}
		}
	}

	// Build results with title boost and snippet extraction
	var results []*SearchResult
	for topicID, score := range scores {
		topic := i.topics[topicID]
		if topic == nil {
			continue
		}

		// Title boost: if query words appear in title
		titleLower := strings.ToLower(topic.Title)
		for _, word := range queryWords {
			if strings.Contains(titleLower, word) {
				score += 2.0 // Title matches are worth more
			}
		}

		// Find matching section and extract snippet
		section, snippet := i.findBestMatch(topic, queryWords)

		results = append(results, &SearchResult{
			Topic:   topic,
			Section: section,
			Score:   score,
			Snippet: snippet,
		})
	}

	// Sort by score (highest first)
	sort.Slice(results, func(a, b int) bool {
		return results[a].Score > results[b].Score
	})

	return results
}

// findBestMatch finds the section with the best match and extracts a snippet.
func (i *searchIndex) findBestMatch(topic *Topic, queryWords []string) (*Section, string) {
	var bestSection *Section
	var bestSnippet string
	bestScore := 0

	// Check topic title
	titleScore := countMatches(topic.Title, queryWords)
	if titleScore > 0 {
		bestSnippet = extractSnippet(topic.Content, queryWords)
	}

	// Check sections
	for idx := range topic.Sections {
		section := &topic.Sections[idx]
		sectionScore := countMatches(section.Title, queryWords)
		contentScore := countMatches(section.Content, queryWords)
		totalScore := sectionScore*2 + contentScore // Title matches worth more

		if totalScore > bestScore {
			bestScore = totalScore
			bestSection = section
			if contentScore > 0 {
				bestSnippet = extractSnippet(section.Content, queryWords)
			} else {
				bestSnippet = extractSnippet(section.Content, nil)
			}
		}
	}

	// If no section matched, use topic content
	if bestSnippet == "" && topic.Content != "" {
		bestSnippet = extractSnippet(topic.Content, queryWords)
	}

	return bestSection, bestSnippet
}

// tokenize splits text into lowercase words for indexing/searching.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var words []string
	var word strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word.WriteRune(r)
		} else if word.Len() > 0 {
			w := word.String()
			if len(w) >= 2 { // Skip single-character words
				words = append(words, w)
			}
			word.Reset()
		}
	}

	// Don't forget the last word
	if word.Len() >= 2 {
		words = append(words, word.String())
	}

	return words
}

// countMatches counts how many query words appear in the text.
func countMatches(text string, queryWords []string) int {
	textLower := strings.ToLower(text)
	count := 0
	for _, word := range queryWords {
		if strings.Contains(textLower, word) {
			count++
		}
	}
	return count
}

// extractSnippet extracts a short snippet around the first match.
// Uses rune-based indexing to properly handle multi-byte UTF-8 characters.
func extractSnippet(content string, queryWords []string) string {
	if content == "" {
		return ""
	}

	const snippetLen = 150

	// If no query words, return start of content
	if len(queryWords) == 0 {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				runes := []rune(line)
				if len(runes) > snippetLen {
					return string(runes[:snippetLen]) + "..."
				}
				return line
			}
		}
		return ""
	}

	// Find first match position (byte-based for strings.Index)
	contentLower := strings.ToLower(content)
	matchPos := -1
	for _, word := range queryWords {
		pos := strings.Index(contentLower, word)
		if pos != -1 && (matchPos == -1 || pos < matchPos) {
			matchPos = pos
		}
	}

	// Convert to runes for safe slicing
	runes := []rune(content)
	runeLen := len(runes)

	if matchPos == -1 {
		// No match found, return start of content
		if runeLen > snippetLen {
			return string(runes[:snippetLen]) + "..."
		}
		return content
	}

	// Convert byte position to rune position (use same string as Index)
	matchRunePos := len([]rune(contentLower[:matchPos]))

	// Extract snippet around match (rune-based)
	start := matchRunePos - 50
	if start < 0 {
		start = 0
	}

	end := start + snippetLen
	if end > runeLen {
		end = runeLen
	}

	snippet := string(runes[start:end])

	// Trim to word boundaries
	if start > 0 {
		if idx := strings.Index(snippet, " "); idx != -1 {
			snippet = "..." + snippet[idx+1:]
		}
	}
	if end < runeLen {
		if idx := strings.LastIndex(snippet, " "); idx != -1 {
			snippet = snippet[:idx] + "..."
		}
	}

	return strings.TrimSpace(snippet)
}

package help

import (
	"regexp"
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
	topics map[string]*Topic   // topicID -> Topic
	index  map[string][]string // word -> []topicID
}

// newSearchIndex creates a new empty search index.
func newSearchIndex() *searchIndex {
	return &searchIndex{
		topics: make(map[string]*Topic),
		index:  make(map[string][]string),
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
	// Avoid duplicates
	for _, id := range i.index[word] {
		if id == topicID {
			return
		}
	}
	i.index[word] = append(i.index[word], topicID)
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
			for _, topicID := range topicIDs {
				scores[topicID] += 1.0
			}
		}

		// Prefix matches (partial word matching)
		for indexWord, topicIDs := range i.index {
			if strings.HasPrefix(indexWord, word) && indexWord != word {
				for _, topicID := range topicIDs {
					scores[topicID] += 0.5 // Lower score for partial matches
				}
			}
		}
	}

	// Pre-compile regexes for snippets
	var res []*regexp.Regexp
	for _, word := range queryWords {
		if len(word) >= 2 {
			if re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(word)); err == nil {
				res = append(res, re)
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
		hasTitleMatch := false
		for _, word := range queryWords {
			if strings.Contains(titleLower, word) {
				hasTitleMatch = true
				break
			}
		}
		if hasTitleMatch {
			score += 10.0
		}

		// Find matching section and extract snippet
		section, snippet := i.findBestMatch(topic, queryWords, res)

		// Section title boost
		if section != nil {
			sectionTitleLower := strings.ToLower(section.Title)
			hasSectionTitleMatch := false
			for _, word := range queryWords {
				if strings.Contains(sectionTitleLower, word) {
					hasSectionTitleMatch = true
					break
				}
			}
			if hasSectionTitleMatch {
				score += 5.0
			}
		}

		results = append(results, &SearchResult{
			Topic:   topic,
			Section: section,
			Score:   score,
			Snippet: snippet,
		})
	}

	// Sort by score (highest first)
	sort.Slice(results, func(a, b int) bool {
		if results[a].Score != results[b].Score {
			return results[a].Score > results[b].Score
		}
		return results[a].Topic.Title < results[b].Topic.Title
	})

	return results
}

// findBestMatch finds the section with the best match and extracts a snippet.
func (i *searchIndex) findBestMatch(topic *Topic, queryWords []string, res []*regexp.Regexp) (*Section, string) {
	var bestSection *Section
	var bestSnippet string
	bestScore := 0

	// Check topic title
	titleScore := countMatches(topic.Title, queryWords)
	if titleScore > 0 {
		bestSnippet = extractSnippet(topic.Content, res)
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
				bestSnippet = extractSnippet(section.Content, res)
			} else {
				bestSnippet = extractSnippet(section.Content, nil)
			}
		}
	}

	// If no section matched, use topic content
	if bestSnippet == "" && topic.Content != "" {
		bestSnippet = extractSnippet(topic.Content, res)
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

// extractSnippet extracts a short snippet around the first match and highlights matches.
func extractSnippet(content string, res []*regexp.Regexp) string {
	if content == "" {
		return ""
	}

	const snippetLen = 150

	// If no regexes, return start of content without highlighting
	if len(res) == 0 {
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

	// Find first match position (byte-based)
	matchPos := -1
	for _, re := range res {
		loc := re.FindStringIndex(content)
		if loc != nil && (matchPos == -1 || loc[0] < matchPos) {
			matchPos = loc[0]
		}
	}

	// Convert to runes for safe slicing
	runes := []rune(content)
	runeLen := len(runes)

	var start, end int
	if matchPos == -1 {
		// No match found, use start of content
		start = 0
		end = snippetLen
		if end > runeLen {
			end = runeLen
		}
	} else {
		// Convert byte position to rune position
		matchRunePos := len([]rune(content[:matchPos]))

		// Extract snippet around match (rune-based)
		start = matchRunePos - 50
		if start < 0 {
			start = 0
		}

		end = start + snippetLen
		if end > runeLen {
			end = runeLen
		}
	}

	snippet := string(runes[start:end])

	// Trim to word boundaries
	prefix := ""
	suffix := ""
	if start > 0 {
		if idx := strings.Index(snippet, " "); idx != -1 {
			snippet = snippet[idx+1:]
			prefix = "..."
		}
	}
	if end < runeLen {
		if idx := strings.LastIndex(snippet, " "); idx != -1 {
			snippet = snippet[:idx]
			suffix = "..."
		}
	}

	snippet = strings.TrimSpace(snippet)
	if snippet == "" {
		return ""
	}

	// Apply highlighting
	highlighted := highlight(snippet, res)

	return prefix + highlighted + suffix
}

// highlight wraps matches in **bold**.
func highlight(text string, res []*regexp.Regexp) string {
	if len(res) == 0 {
		return text
	}

	type match struct {
		start, end int
	}
	var matches []match

	for _, re := range res {
		indices := re.FindAllStringIndex(text, -1)
		for _, idx := range indices {
			matches = append(matches, match{idx[0], idx[1]})
		}
	}

	if len(matches) == 0 {
		return text
	}

	// Sort matches by start position
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].start != matches[j].start {
			return matches[i].start < matches[j].start
		}
		return matches[i].end > matches[j].end
	})

	// Merge overlapping or adjacent matches
	var merged []match
	if len(matches) > 0 {
		curr := matches[0]
		for i := 1; i < len(matches); i++ {
			if matches[i].start <= curr.end {
				if matches[i].end > curr.end {
					curr.end = matches[i].end
				}
			} else {
				merged = append(merged, curr)
				curr = matches[i]
			}
		}
		merged = append(merged, curr)
	}

	// Build highlighted string from back to front to avoid position shifts
	result := text
	for i := len(merged) - 1; i >= 0; i-- {
		m := merged[i]
		result = result[:m.end] + "**" + result[m.end:]
		result = result[:m.start] + "**" + result[m.start:]
	}

	return result
}

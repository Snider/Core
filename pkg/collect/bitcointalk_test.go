package collect

import (
	"context"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestBitcoinTalkCollector_Name_Good(t *testing.T) {
	b := &BitcoinTalkCollector{TopicID: "12345"}
	assert.Equal(t, "bitcointalk:12345", b.Name())
}

func TestBitcoinTalkCollector_Name_Good_URL(t *testing.T) {
	b := &BitcoinTalkCollector{URL: "https://bitcointalk.org/index.php?topic=12345.0"}
	assert.Equal(t, "bitcointalk:url", b.Name())
}

func TestBitcoinTalkCollector_Collect_Bad_NoTopicID(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	b := &BitcoinTalkCollector{}
	_, err := b.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestBitcoinTalkCollector_Collect_Good_DryRun(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestParsePostsFromHTML_Good(t *testing.T) {
	sampleHTML := `
	<html><body>
		<div class="post">
			<div class="poster_info">satoshi</div>
			<div class="headerandpost">
				<div class="smalltext">January 03, 2009</div>
			</div>
			<div class="inner">This is the first post content.</div>
		</div>
		<div class="post">
			<div class="poster_info">hal</div>
			<div class="headerandpost">
				<div class="smalltext">January 10, 2009</div>
			</div>
			<div class="inner">Running bitcoin!</div>
		</div>
	</body></html>`

	posts, err := ParsePostsFromHTML(sampleHTML)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)

	assert.Contains(t, posts[0].Author, "satoshi")
	assert.Contains(t, posts[0].Content, "This is the first post content.")
	assert.Contains(t, posts[0].Date, "January 03, 2009")

	assert.Contains(t, posts[1].Author, "hal")
	assert.Contains(t, posts[1].Content, "Running bitcoin!")
}

func TestParsePostsFromHTML_Good_Empty(t *testing.T) {
	posts, err := ParsePostsFromHTML("<html><body></body></html>")
	assert.NoError(t, err)
	assert.Empty(t, posts)
}

func TestFormatPostMarkdown_Good(t *testing.T) {
	md := FormatPostMarkdown(1, "satoshi", "January 03, 2009", "Hello, world!")

	assert.Contains(t, md, "# Post 1 by satoshi")
	assert.Contains(t, md, "**Date:** January 03, 2009")
	assert.Contains(t, md, "Hello, world!")
}

func TestFormatPostMarkdown_Good_NoDate(t *testing.T) {
	md := FormatPostMarkdown(5, "user", "", "Content here")

	assert.Contains(t, md, "# Post 5 by user")
	assert.NotContains(t, md, "**Date:**")
	assert.Contains(t, md, "Content here")
}

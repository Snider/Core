package collect

import (
	"context"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestPapersCollector_Name_Good(t *testing.T) {
	p := &PapersCollector{Source: PaperSourceIACR}
	assert.Equal(t, "papers:iacr", p.Name())
}

func TestPapersCollector_Name_Good_ArXiv(t *testing.T) {
	p := &PapersCollector{Source: PaperSourceArXiv}
	assert.Equal(t, "papers:arxiv", p.Name())
}

func TestPapersCollector_Name_Good_All(t *testing.T) {
	p := &PapersCollector{Source: PaperSourceAll}
	assert.Equal(t, "papers:all", p.Name())
}

func TestPapersCollector_Collect_Bad_NoQuery(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	p := &PapersCollector{Source: PaperSourceIACR}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_Collect_Bad_UnknownSource(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	p := &PapersCollector{Source: "unknown", Query: "test"}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_Collect_Good_DryRun(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	p := &PapersCollector{Source: PaperSourceAll, Query: "cryptography"}
	result, err := p.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestFormatPaperMarkdown_Good(t *testing.T) {
	md := FormatPaperMarkdown(
		"Zero-Knowledge Proofs Revisited",
		[]string{"Alice", "Bob"},
		"2025-01-15",
		"https://eprint.iacr.org/2025/001",
		"iacr",
		"We present a new construction for zero-knowledge proofs.",
	)

	assert.Contains(t, md, "# Zero-Knowledge Proofs Revisited")
	assert.Contains(t, md, "**Authors:** Alice, Bob")
	assert.Contains(t, md, "**Published:** 2025-01-15")
	assert.Contains(t, md, "**URL:** https://eprint.iacr.org/2025/001")
	assert.Contains(t, md, "**Source:** iacr")
	assert.Contains(t, md, "## Abstract")
	assert.Contains(t, md, "zero-knowledge proofs")
}

func TestFormatPaperMarkdown_Good_Minimal(t *testing.T) {
	md := FormatPaperMarkdown("Title Only", nil, "", "", "", "")

	assert.Contains(t, md, "# Title Only")
	assert.NotContains(t, md, "**Authors:**")
	assert.NotContains(t, md, "## Abstract")
}

func TestArxivEntryToPaper_Good(t *testing.T) {
	entry := arxivEntry{
		ID:        "http://arxiv.org/abs/2501.12345v1",
		Title:     "  A Great Paper  ",
		Summary:   "  This paper presents...  ",
		Published: "2025-01-15T00:00:00Z",
		Authors: []arxivAuthor{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		Links: []arxivLink{
			{Href: "http://arxiv.org/abs/2501.12345v1", Rel: "alternate"},
			{Href: "http://arxiv.org/pdf/2501.12345v1", Rel: "related", Type: "application/pdf"},
		},
	}

	ppr := arxivEntryToPaper(entry)

	assert.Equal(t, "2501.12345v1", ppr.ID)
	assert.Equal(t, "A Great Paper", ppr.Title)
	assert.Equal(t, "This paper presents...", ppr.Abstract)
	assert.Equal(t, "2025-01-15T00:00:00Z", ppr.Date)
	assert.Equal(t, []string{"Alice", "Bob"}, ppr.Authors)
	assert.Equal(t, "http://arxiv.org/abs/2501.12345v1", ppr.URL)
	assert.Equal(t, "arxiv", ppr.Source)
}

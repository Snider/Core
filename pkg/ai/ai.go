// Package ai provides the unified AI package for the core CLI.
//
// It composes functionality from pkg/rag (vector search) and pkg/agentic
// (task management) into a single public API surface. New AI features
// should be added here; existing packages remain importable but pkg/ai
// is the canonical entry point.
//
// Sub-packages composed:
//   - pkg/rag: Qdrant vector database + Ollama embeddings
//   - pkg/agentic: Task queue client and context building
package ai

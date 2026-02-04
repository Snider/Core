# RAG Pipeline for Host UK Documentation

Store documentation in a vector database so Claude (and local LLMs) can retrieve relevant context without being reminded every conversation.

## The Problem This Solves

> "The amount of times I've had to re-tell you how to make a Flux button is crazy"

Instead of wasting context window on "remember, Flux buttons work like this...", the RAG system:
1. Stores all documentation in Qdrant
2. Claude queries before answering
3. Relevant docs injected automatically
4. No more re-teaching

## Prerequisites

**Already running on your lab:**
- Qdrant: `linux.snider.dev:6333`
- Ollama: `linux.snider.dev:11434` (or local)

**Install Python deps:**
```bash
pip install -r requirements.txt
```

**Ensure embedding model is available:**
```bash
ollama pull nomic-embed-text
```

## Quick Start

### 1. Ingest Documentation

```bash
# Ingest recovered Host UK docs
python ingest.py /Users/snider/Code/host-uk/core/tasks/recovered-hostuk \
    --collection hostuk-docs \
    --recreate

# Ingest Flux UI docs separately (higher priority)
python ingest.py /path/to/flux-ui-docs \
    --collection flux-ui-docs \
    --recreate
```

### 2. Query the Database

```bash
# Search for Flux button docs
python query.py "how to create a Flux button component"

# Filter by category
python query.py "path sandboxing" --category architecture

# Get more results
python query.py "Vi personality" --top 10

# Output as JSON
python query.py "brand voice" --format json

# Output for LLM context injection
python query.py "Flux modal component" --format context
```

### 3. List Collections

```bash
python query.py --list-collections
python query.py --stats --collection flux-ui-docs
```

## Collections Strategy

| Collection | Content | Priority |
|------------|---------|----------|
| `flux-ui-docs` | Flux Pro component docs | High (UI questions) |
| `hostuk-docs` | Recovered implementation docs | Medium |
| `brand-docs` | Vi, brand voice, visual identity | For content generation |
| `lethean-docs` | SASE/dVPN technical docs | Product-specific |

## Integration with Claude Code

### Option 1: MCP Server (Best)

Create an MCP server that Claude can query:

```go
// In core CLI
func (s *RagServer) Query(query string) ([]Document, error) {
    // Query Qdrant
    // Return relevant docs
}
```

Then Claude can call `rag.query("Flux button")` and get docs automatically.

### Option 2: CLAUDE.md Instruction

Add to project CLAUDE.md:

```markdown
## Before Answering UI Questions

When asked about Flux UI components, query the RAG database first:
```bash
python /path/to/query.py "your question" --collection flux-ui-docs --format context
```

Include the retrieved context in your response.
```

### Option 3: Claude Code Hook

Create a hook that auto-injects context for certain queries.

## Category Taxonomy

The ingestion automatically categorizes files:

| Category | Matches |
|----------|---------|
| `ui-component` | flux, ui/component |
| `brand` | brand, mascot |
| `product-brief` | brief |
| `help-doc` | help, draft |
| `task` | task, plan |
| `architecture` | architecture, migration |
| `documentation` | default |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `QDRANT_HOST` | linux.snider.dev | Qdrant server |
| `QDRANT_PORT` | 6333 | Qdrant port |
| `EMBEDDING_MODEL` | nomic-embed-text | Ollama model |
| `CHUNK_SIZE` | 500 | Characters per chunk |
| `CHUNK_OVERLAP` | 50 | Overlap between chunks |

## Training a Model vs RAG

**RAG** (what this does):
- Model weights unchanged
- Documents retrieved at query time
- Knowledge updates instantly (re-ingest)
- Good for: facts, API docs, current information

**Fine-tuning** (separate process):
- Model weights updated
- Knowledge baked into model
- Requires retraining to update
- Good for: style, patterns, conventions

**For Flux UI**: RAG is perfect. The docs change, API changes, you want current info.

**For Vi's voice**: Fine-tuning is better. Style doesn't change often, should be "baked in".

## Vector Math (For Understanding)

```text
"How do I make a Flux button?"
    ↓ Embedding
[0.12, -0.45, 0.78, ...768 floats...]
    ↓ Cosine similarity search
Find chunks with similar vectors
    ↓ Results
1. doc/ui/flux/components/button.md (score: 0.89)
2. doc/ui/flux/forms.md (score: 0.76)
3. doc/ui/flux/components/input.md (score: 0.71)
```

The embedding model converts text to "meaning vectors". Similar meanings = similar vectors = found by search.

## Troubleshooting

**"No results found"**
- Lower threshold: `--threshold 0.3`
- Check collection has data: `--stats`
- Verify Ollama is running: `ollama list`

**"Connection refused"**
- Check Qdrant is running: `curl http://linux.snider.dev:6333/collections`
- Check firewall/network

**"Embedding model not available"**
```bash
ollama pull nomic-embed-text
```

---

*Part of the Host UK Core CLI tooling*

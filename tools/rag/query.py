#!/usr/bin/env python3
"""
RAG Query Tool for Host UK Documentation

Query the vector database and retrieve relevant documentation chunks.

Usage:
    python query.py "how do I create a Flux button"
    python query.py "what is Vi's personality" --collection hostuk-docs
    python query.py "path sandboxing" --top 10 --category architecture

Requirements:
    pip install qdrant-client ollama
"""

import argparse
import html
import json
import os
import sys
from typing import Optional

try:
    from qdrant_client import QdrantClient
    from qdrant_client.models import Filter, FieldCondition, MatchValue
    import ollama
except ImportError:
    print("Install dependencies: pip install qdrant-client ollama")
    sys.exit(1)


# Configuration
QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", "6333"))
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "nomic-embed-text")


def generate_embedding(text: str) -> list[float]:
    """Generate embedding using Ollama."""
    response = ollama.embeddings(model=EMBEDDING_MODEL, prompt=text)
    return response["embedding"]


def query_rag(
    query: str,
    client: QdrantClient,
    collection: str,
    top_k: int = 5,
    category: Optional[str] = None,
    score_threshold: float = 0.5,
) -> list[dict]:
    """Query the RAG database and return relevant chunks."""

    # Generate query embedding
    query_embedding = generate_embedding(query)

    # Build filter if category specified
    query_filter = None
    if category:
        query_filter = Filter(
            must=[
                FieldCondition(key="category", match=MatchValue(value=category))
            ]
        )

    # Search
    results = client.query_points(
        collection_name=collection,
        query=query_embedding,
        query_filter=query_filter,
        limit=top_k,
        score_threshold=score_threshold,
    ).points

    return [
        {
            "score": hit.score,
            "text": hit.payload["text"],
            "source": hit.payload["source"],
            "section": hit.payload.get("section", ""),
            "category": hit.payload.get("category", ""),
        }
        for hit in results
    ]


def format_results(results: list[dict], query: str, format: str = "text") -> str:
    """Format results for display."""

    if format == "json":
        return json.dumps(results, indent=2)

    if not results:
        return f"No results found for: {query}"

    output = []
    output.append(f"Query: {query}")
    output.append(f"Results: {len(results)}")
    output.append("=" * 60)

    for i, r in enumerate(results, 1):
        output.append(f"\n[{i}] {r['source']} (score: {r['score']:.3f})")
        if r['section']:
            output.append(f"    Section: {r['section']}")
        output.append(f"    Category: {r['category']}")
        output.append("-" * 40)
        # Truncate long text for display
        text = r['text']
        if len(text) > 500:
            text = text[:500] + "..."
        output.append(text)
        output.append("")

    return "\n".join(output)


def format_for_context(results: list[dict], query: str) -> str:
    """Format results as context for LLM injection."""

    if not results:
        return ""

    output = []
    output.append(f'<retrieved_context query="{html.escape(query)}">')

    for r in results:
        output.append(f'\n<document source="{html.escape(r["source"])}" category="{html.escape(r["category"])}">')
        output.append(html.escape(r['text']))
        output.append("</document>")

    output.append("\n</retrieved_context>")

    return "\n".join(output)

def main():
    parser = argparse.ArgumentParser(description="Query RAG documentation")
    parser.add_argument("query", nargs="?", help="Search query")
    parser.add_argument("--collection", default="hostuk-docs", help="Qdrant collection name")
    parser.add_argument("--top", "-k", type=int, default=5, help="Number of results")
    parser.add_argument("--category", "-c", help="Filter by category")
    parser.add_argument("--threshold", "-t", type=float, default=0.5, help="Score threshold")
    parser.add_argument("--format", "-f", choices=["text", "json", "context"], default="text")
    parser.add_argument("--qdrant-host", default=QDRANT_HOST)
    parser.add_argument("--qdrant-port", type=int, default=QDRANT_PORT)
    parser.add_argument("--list-collections", action="store_true", help="List available collections")
    parser.add_argument("--stats", action="store_true", help="Show collection stats")

    args = parser.parse_args()

    # Connect to Qdrant
    client = QdrantClient(host=args.qdrant_host, port=args.qdrant_port)

    # List collections
    if args.list_collections:
        collections = client.get_collections().collections
        print("Available collections:")
        for c in collections:
            info = client.get_collection(c.name)
            print(f"  - {c.name}: {info.points_count} vectors")
        return

    # Show stats
    if args.stats:
        try:
            info = client.get_collection(args.collection)
            print(f"Collection: {args.collection}")
            print(f"  Vectors: {info.points_count}")
            print(f"  Status: {info.status}")
        except Exception as e:
            print(f"Collection not found: {args.collection}")
        return

    # Query required
    if not args.query:
        parser.print_help()
        return

    # Execute query
    results = query_rag(
        query=args.query,
        client=client,
        collection=args.collection,
        top_k=args.top,
        category=args.category,
        score_threshold=args.threshold,
    )

    # Format output
    if args.format == "context":
        print(format_for_context(results, args.query))
    else:
        print(format_results(results, args.query, args.format))


if __name__ == "__main__":
    main()
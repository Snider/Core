#!/usr/bin/env python3
"""
RAG Ingestion Pipeline for Host UK Documentation

Chunks markdown files, generates embeddings via Ollama, stores in Qdrant.

Usage:
    python ingest.py /path/to/docs --collection hostuk-docs
    python ingest.py /path/to/flux-ui --collection flux-ui-docs

Requirements:
    pip install qdrant-client ollama markdown
"""

import argparse
import hashlib
import json
import os
import re
import sys
from pathlib import Path
from typing import Generator

try:
    from qdrant_client import QdrantClient
    from qdrant_client.models import Distance, VectorParams, PointStruct
    import ollama
except ImportError:
    print("Install dependencies: pip install qdrant-client ollama")
    sys.exit(1)


# Configuration
QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", "6333"))
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "nomic-embed-text")
CHUNK_SIZE = int(os.getenv("CHUNK_SIZE", "500"))  # chars
CHUNK_OVERLAP = int(os.getenv("CHUNK_OVERLAP", "50"))  # chars
VECTOR_DIM = 768  # nomic-embed-text dimension


def chunk_markdown(text: str, chunk_size: int = CHUNK_SIZE, overlap: int = CHUNK_OVERLAP) -> Generator[dict, None, None]:
    """
    Chunk markdown by sections (## headers), then by paragraphs if too long.
    Preserves context with overlap.
    """
    # Split by ## headers first
    sections = re.split(r'\n(?=## )', text)

    for section in sections:
        if not section.strip():
            continue

        # Extract section title
        lines = section.strip().split('\n')
        title = lines[0].lstrip('#').strip() if lines[0].startswith('#') else ""

        # If section is small enough, yield as-is
        if len(section) <= chunk_size:
            yield {
                "text": section.strip(),
                "section": title,
            }
            continue

        # Otherwise, chunk by paragraphs
        paragraphs = re.split(r'\n\n+', section)
        current_chunk = ""

        for para in paragraphs:
            if len(current_chunk) + len(para) <= chunk_size:
                current_chunk += "\n\n" + para if current_chunk else para
            else:
                if current_chunk:
                    yield {
                        "text": current_chunk.strip(),
                        "section": title,
                    }
                # Start new chunk with overlap from previous
                if overlap and current_chunk:
                    overlap_text = current_chunk[-overlap:]
                    current_chunk = overlap_text + "\n\n" + para
                else:
                    current_chunk = para

        # Don't forget the last chunk
        if current_chunk.strip():
            yield {
                "text": current_chunk.strip(),
                "section": title,
            }


def generate_embedding(text: str) -> list[float]:
    """Generate embedding using Ollama."""
    response = ollama.embeddings(model=EMBEDDING_MODEL, prompt=text)
    return response["embedding"]


def get_file_category(path: str) -> str:
    """Determine category from file path."""
    path_lower = path.lower()

    if "flux" in path_lower or "ui/component" in path_lower:
        return "ui-component"
    elif "brand" in path_lower or "mascot" in path_lower:
        return "brand"
    elif "brief" in path_lower:
        return "product-brief"
    elif "help" in path_lower or "draft" in path_lower:
        return "help-doc"
    elif "task" in path_lower or "plan" in path_lower:
        return "task"
    elif "architecture" in path_lower or "migration" in path_lower:
        return "architecture"
    else:
        return "documentation"


def ingest_directory(
    directory: Path,
    client: QdrantClient,
    collection: str,
    verbose: bool = False
) -> dict:
    """Ingest all markdown files from directory into Qdrant."""

    stats = {"files": 0, "chunks": 0, "errors": 0}
    points = []

    # Find all markdown files
    md_files = list(directory.rglob("*.md"))
    print(f"Found {len(md_files)} markdown files")

    for file_path in md_files:
        try:
            rel_path = str(file_path.relative_to(directory))

            with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                content = f.read()

            if not content.strip():
                continue

            # Extract metadata
            category = get_file_category(rel_path)

            # Chunk the content
            for i, chunk in enumerate(chunk_markdown(content)):
                chunk_id = hashlib.md5(
                    f"{rel_path}:{i}:{chunk['text'][:100]}".encode()
                ).hexdigest()

                # Generate embedding
                embedding = generate_embedding(chunk["text"])

                # Create point
                point = PointStruct(
                    id=chunk_id,
                    vector=embedding,
                    payload={
                        "text": chunk["text"],
                        "source": rel_path,
                        "section": chunk["section"],
                        "category": category,
                        "chunk_index": i,
                    }
                )
                points.append(point)
                stats["chunks"] += 1

                if verbose:
                    print(f"  [{category}] {rel_path} chunk {i}: {len(chunk['text'])} chars")

            stats["files"] += 1
            if not verbose:
                print(f"  Processed: {rel_path} ({stats['chunks']} chunks total)")

        except Exception as e:
            print(f"  Error processing {file_path}: {e}")
            stats["errors"] += 1

    # Batch upsert to Qdrant
    if points:
        print(f"\nUpserting {len(points)} vectors to Qdrant...")

        # Upsert in batches of 100
        batch_size = 100
        for i in range(0, len(points), batch_size):
            batch = points[i:i + batch_size]
            client.upsert(collection_name=collection, points=batch)
            print(f"  Uploaded batch {i // batch_size + 1}/{(len(points) - 1) // batch_size + 1}")

    return stats


def main():
    parser = argparse.ArgumentParser(description="Ingest markdown docs into Qdrant")
    parser.add_argument("directory", type=Path, help="Directory containing markdown files")
    parser.add_argument("--collection", default="hostuk-docs", help="Qdrant collection name")
    parser.add_argument("--recreate", action="store_true", help="Delete and recreate collection")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")
    parser.add_argument("--qdrant-host", default=QDRANT_HOST, help="Qdrant host")
    parser.add_argument("--qdrant-port", type=int, default=QDRANT_PORT, help="Qdrant port")

    args = parser.parse_args()

    if not args.directory.exists():
        print(f"Error: Directory not found: {args.directory}")
        sys.exit(1)

    # Connect to Qdrant
    print(f"Connecting to Qdrant at {args.qdrant_host}:{args.qdrant_port}...")
    client = QdrantClient(host=args.qdrant_host, port=args.qdrant_port)

    # Create or recreate collection
    collections = [c.name for c in client.get_collections().collections]

    if args.recreate and args.collection in collections:
        print(f"Deleting existing collection: {args.collection}")
        client.delete_collection(args.collection)
        collections.remove(args.collection)

    if args.collection not in collections:
        print(f"Creating collection: {args.collection}")
        client.create_collection(
            collection_name=args.collection,
            vectors_config=VectorParams(size=VECTOR_DIM, distance=Distance.COSINE)
        )

    # Verify Ollama model is available
    print(f"Using embedding model: {EMBEDDING_MODEL}")
    try:
        ollama.embeddings(model=EMBEDDING_MODEL, prompt="test")
    except Exception as e:
        print(f"Error: Embedding model not available. Run: ollama pull {EMBEDDING_MODEL}")
        sys.exit(1)

    # Ingest files
    print(f"\nIngesting from: {args.directory}")
    stats = ingest_directory(args.directory, client, args.collection, args.verbose)

    # Summary
    print(f"\n{'=' * 50}")
    print(f"Ingestion complete!")
    print(f"  Files processed: {stats['files']}")
    print(f"  Chunks created:  {stats['chunks']}")
    print(f"  Errors:          {stats['errors']}")
    print(f"  Collection:      {args.collection}")
    print(f"{'=' * 50}")


if __name__ == "__main__":
    main()

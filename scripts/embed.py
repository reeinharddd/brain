#!/usr/bin/env python3
"""
embed.py - Semantic embedding backend for brain repo vector sync.

Strategy:
  1. Try OpenAI text-embedding-3-small (best quality, requires OPENAI_API_KEY)
  2. Try local sentence-transformers (medium quality, no API needed)
  3. Fall back to deterministic TF-IDF-style hash projection (no deps, low quality but consistent)

Usage:
  echo "some text" | python3 embed.py
  python3 embed.py --text "some text" --dim 1536
  python3 embed.py --batch input.jsonl --output embeddings.jsonl

Returns: JSON list of floats
"""

import argparse
import hashlib
import json
import math
import os
import re
import sys
import urllib.error
import urllib.request


# ── Backend: OpenAI ──────────────────────────────────────────────────────────

def embed_openai(texts: list[str], dim: int = 1536) -> list[list[float]]:
    api_key = os.environ.get("OPENAI_API_KEY", "")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY not set")

    payload = json.dumps({
        "model": "text-embedding-3-small",
        "input": texts,
        "dimensions": dim,
    }).encode("utf-8")

    req = urllib.request.Request(
        "https://api.openai.com/v1/embeddings",
        data=payload,
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        data = json.loads(resp.read().decode("utf-8"))

    data["data"].sort(key=lambda x: x["index"])
    return [item["embedding"] for item in data["data"]]


# ── Backend: sentence-transformers (local, optional) ─────────────────────────

def embed_local(texts: list[str], dim: int = 384) -> list[list[float]]:
    try:
        from sentence_transformers import SentenceTransformer  # type: ignore
    except ImportError:
        raise RuntimeError("sentence-transformers not installed. Run: pip install sentence-transformers")

    model_name = os.environ.get("BRAIN_EMBED_MODEL", "all-MiniLM-L6-v2")
    model = SentenceTransformer(model_name)
    vecs = model.encode(texts, normalize_embeddings=True).tolist()
    # Optionally truncate/pad to requested dim
    if dim and dim != len(vecs[0]):
        vecs = [v[:dim] if len(v) >= dim else v + [0.0] * (dim - len(v)) for v in vecs]
    return vecs


# ── Backend: Deterministic hash projection (zero-dep fallback) ───────────────

def embed_hash(texts: list[str], dim: int = 256) -> list[list[float]]:
    """
    Deterministic TF-IDF-style bag-of-words hash embedding.
    Uses SHA-256 to project tokens into dim-dimensional space.
    Quality: poor for semantic similarity, but consistent and reproducible.
    Used only when no real embedding backend is available.
    """
    results = []
    for text in texts:
        vector = [0.0] * dim
        tokens = re.findall(r"[A-Za-z0-9_]{2,}", text.lower())
        if not tokens:
            results.append(vector)
            continue
        # TF weighting: count occurrences
        tf: dict[str, int] = {}
        for token in tokens:
            tf[token] = tf.get(token, 0) + 1
        for token, count in tf.items():
            digest = hashlib.sha256(token.encode("utf-8")).hexdigest()
            bucket = int(digest[:8], 16) % dim
            sign = -1.0 if int(digest[8:10], 16) % 2 else 1.0
            # Weight by log(1 + tf) for smoother distribution
            vector[bucket] += sign * math.log1p(count)
        norm = math.sqrt(sum(v * v for v in vector)) or 1.0
        results.append([v / norm for v in vector])
    return results


# ── Auto-select backend ───────────────────────────────────────────────────────

def embed(texts: list[str], dim: int | None = None, backend: str = "auto") -> tuple[list[list[float]], str]:
    """
    Returns (embeddings, backend_used).
    backend: "auto" | "openai" | "local" | "hash"
    """
    if not texts:
        return [], "none"

    if backend == "openai" or (backend == "auto" and os.environ.get("OPENAI_API_KEY")):
        try:
            vecs = embed_openai(texts, dim=dim or 1536)
            return vecs, "openai"
        except Exception as e:
            if backend == "openai":
                raise
            sys.stderr.write(f"[embed] OpenAI failed: {e}. Trying local.\n")

    if backend == "local" or backend == "auto":
        try:
            vecs = embed_local(texts, dim=dim or 384)
            return vecs, "local"
        except Exception as e:
            if backend == "local":
                raise
            sys.stderr.write(f"[embed] Local model failed: {e}. Falling back to hash.\n")

    # Always available
    vecs = embed_hash(texts, dim=dim or 256)
    return vecs, "hash"


# ── CLI ───────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Embed text using best available backend")
    parser.add_argument("--text", help="Single text to embed")
    parser.add_argument("--batch", help="NDJSON file with {id, content} per line")
    parser.add_argument("--output", help="Output NDJSON file (for --batch)")
    parser.add_argument("--dim", type=int, default=None, help="Embedding dimensions")
    parser.add_argument("--backend", default="auto", choices=["auto", "openai", "local", "hash"])
    parser.add_argument("--info", action="store_true", help="Print backend info and exit")
    args = parser.parse_args()

    if args.info:
        has_openai = bool(os.environ.get("OPENAI_API_KEY"))
        try:
            import sentence_transformers  # noqa: F401
            has_local = True
        except ImportError:
            has_local = False
        print(json.dumps({
            "backends_available": {
                "openai": has_openai,
                "local": has_local,
                "hash": True,
            },
            "active_backend": "openai" if has_openai else ("local" if has_local else "hash"),
        }, indent=2))
        return

    if args.text:
        vecs, backend_used = embed([args.text], dim=args.dim, backend=args.backend)
        print(json.dumps({"embedding": vecs[0], "backend": backend_used, "dim": len(vecs[0])}))
        return

    if args.batch:
        items = []
        with open(args.batch, encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line:
                    items.append(json.loads(line))

        texts = [item.get("content", "") for item in items]
        vecs, backend_used = embed(texts, dim=args.dim, backend=args.backend)
        sys.stderr.write(f"[embed] Backend: {backend_used}, dim: {len(vecs[0]) if vecs else 0}, items: {len(vecs)}\n")

        out_lines = []
        for item, vec in zip(items, vecs):
            out_lines.append(json.dumps({**item, "embedding": vec, "embedding_backend": backend_used}))

        output = "\n".join(out_lines) + "\n"
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(output)
        else:
            sys.stdout.write(output)
        return

    # stdin mode
    text = sys.stdin.read().strip()
    if text:
        vecs, backend_used = embed([text], dim=args.dim, backend=args.backend)
        print(json.dumps({"embedding": vecs[0], "backend": backend_used, "dim": len(vecs[0])}))


if __name__ == "__main__":
    main()

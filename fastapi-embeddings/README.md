# FastAPI Embedding Service

Lightweight HTTP service for generating text embeddings via sentence-transformers. Used by the Ragtime backend for document ingestion and semantic search.

Default model: `sentence-transformers/all-MiniLM-L6-v2` — 384 dimensions, fast inference. Override with the `EMBED_MODEL` env var.

## Quickstart

```bash
python3.13 -m venv .venv && .venv/bin/pip install -r requirements.txt
.venv/bin/uvicorn server:app --host 0.0.0.0 --port 8000
curl http://localhost:8000/health
```

## Endpoints

- `GET /health` — health check, returns active model name
- `POST /embed` — embed one or more strings; optional `normalize: true` for L2-normalized vectors (cosine similarity)
- `GET /docs` — Swagger UI

## Deploy on Railway

Railway auto-detects the `Dockerfile`. Set `EMBED_MODEL` in Railway env vars to change the model.

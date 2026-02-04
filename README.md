# FastAPI Embedding Service

A lightweight FastAPI microservice for generating text embeddings using sentence-transformers models.

## Overview

This service provides a simple HTTP API for converting text into vector embeddings. It's designed to be used by the Ragtime knowledge base system for document chunking and retrieval operations.

## Features

- Fast embedding generation using sentence-transformers
- Model loaded once at startup for quick inference
- Optional L2 normalization for cosine similarity
- Batch processing support
- Health check endpoint

## Prerequisites

- Python 3.8+
- pip or uv for package management

## Installation

1. Create and activate a virtual environment:

```bash
python -m venv .venv
source .venv/bin/activate  # On macOS/Linux
# or
.venv\Scripts\activate  # On Windows
```

2. Install dependencies:

```bash
pip install -r requirements.txt
```

## Running the Service

### Default Configuration

Start the service with the default model (`sentence-transformers/all-MiniLM-L6-v2`):

```bash
uvicorn server:app --host 0.0.0.0 --port 8000
```

### Custom Model

Specify a different embedding model via environment variable:

```bash
EMBED_MODEL=sentence-transformers/all-mpnet-base-v2 uvicorn server:app --host 0.0.0.0 --port 8000
```

### Development Mode

Run with auto-reload for development:

```bash
uvicorn server:app --reload --host 0.0.0.0 --port 8000
```

## API Documentation

### Health Check

Check if the service is running and see which model is loaded.

**Endpoint:** `GET /health`

**Example:**

```bash
curl http://localhost:8000/health
```

**Response:**

```json
{
  "ok": true,
  "model": "sentence-transformers/all-MiniLM-L6-v2"
}
```

### Generate Embeddings

Generate vector embeddings for one or more text strings.

**Endpoint:** `POST /embed`

**Request Body:**

```json
{
  "texts": ["string1", "string2", ...],
  "normalize": false  // optional, default: false
}
```

**Parameters:**

- `texts` (required): Array of strings to embed (minimum 1 item)
- `normalize` (optional): If true, L2-normalize the output vectors for cosine similarity (default: false)

**Response:**

```json
{
  "model": "sentence-transformers/all-MiniLM-L6-v2",
  "dim": 384,
  "embeddings": [
    [0.123, -0.456, 0.789, ...],
    [0.321, -0.654, 0.987, ...]
  ]
}
```

**Example - Single Text:**

```bash
curl -X POST http://localhost:8000/embed \
  -H "Content-Type: application/json" \
  -d '{
    "texts": ["This is a sample sentence to embed."]
  }'
```

**Example - Multiple Texts:**

```bash
curl -X POST http://localhost:8000/embed \
  -H "Content-Type: application/json" \
  -d '{
    "texts": [
      "First document chunk",
      "Second document chunk",
      "Third document chunk"
    ]
  }'
```

**Example - With Normalization:**

```bash
curl -X POST http://localhost:8000/embed \
  -H "Content-Type: application/json" \
  -d '{
    "texts": ["Normalize this text for cosine similarity"],
    "normalize": true
  }'
```

## Interactive API Documentation

Once the service is running, you can access the auto-generated interactive API documentation:

- **Swagger UI:** http://localhost:8000/docs
- **ReDoc:** http://localhost:8000/redoc

## Configuration

### Environment Variables

- `EMBED_MODEL`: The sentence-transformers model to use (default: `sentence-transformers/all-MiniLM-L6-v2`)

### Supported Models

Any model from the [sentence-transformers library](https://www.sbert.net/docs/pretrained_models.html) can be used. Popular choices:

- `sentence-transformers/all-MiniLM-L6-v2` (default) - Fast, 384 dimensions
- `sentence-transformers/all-mpnet-base-v2` - High quality, 768 dimensions
- `sentence-transformers/all-MiniLM-L12-v2` - Balanced, 384 dimensions

## Notes

- Empty or whitespace-only strings are automatically filtered out
- The model is loaded once at startup and cached for all requests
- Batch size is set to 64 for optimal performance
- The service returns embeddings as plain Python lists (JSON-serializable)

## Integration with Ragtime

This service is called by the Go backend's `EmbeddingService` client during:

1. **Document Ingestion**: Embedding chunks after structural chunking
2. **Retrieval**: Embedding user queries for semantic search

The backend expects the response format defined in `EmbedResponse` and uses the `dim` field to route embeddings to the correct dimension-specific table (e.g., `embeddings_384`, `embeddings_768`).

## Deployment

### Railway

This service is ready to be deployed on [Railway](https://railway.app/).

1. Fork or push this repository to GitHub.
2. In Railway, create a new project and select "Deploy from GitHub repo".
3. Select this repository.
4. Railway will automatically detect the `Dockerfile` and build the service.
5. (Optional) Set the `EMBED_MODEL` environment variable in Railway settings to use a different model.


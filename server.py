from __future__ import annotations

from typing import List, Optional
from fastapi import FastAPI
from pydantic import BaseModel, Field
from sentence_transformers import SentenceTransformer
import os

MODEL_NAME = os.getenv("EMBED_MODEL", "sentence-transformers/all-MiniLM-L6-v2")

app = FastAPI(title="Embedding Service", version="0.1.0")

# Load once at startup (so requests are fast)
model: Optional[SentenceTransformer] = None


class EmbedRequest(BaseModel):
    texts: List[str] = Field(..., min_items=1, description="List of strings to embed")
    normalize: bool = Field(
        False,
        description="If true, L2-normalize vectors (useful for cosine similarity).",
    )


class EmbedResponse(BaseModel):
    model: str
    dim: int
    embeddings: List[List[float]]


@app.on_event("startup")
def startup():
    global model
    model = SentenceTransformer(MODEL_NAME)


@app.get("/health")
def health():
    return {"ok": True, "model": MODEL_NAME}


@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    assert model is not None, "Model not loaded"

    # Strip and drop empty strings
    cleaned = [t.strip() for t in req.texts if t and t.strip()]
    if not cleaned:
        return {"model": MODEL_NAME, "dim": 0, "embeddings": []}

    # sentence-transformers returns np.ndarray; convert to plain python lists
    vectors = model.encode(
        cleaned,
        batch_size=64,
        normalize_embeddings=req.normalize,
        show_progress_bar=False,
    )

    embeddings = vectors.tolist()
    dim = len(embeddings[0]) if embeddings else 0

    return {"model": MODEL_NAME, "dim": dim, "embeddings": embeddings}

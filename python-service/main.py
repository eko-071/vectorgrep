import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

from parser import parse_manpage
from vector_store import VectorStore

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

INDEX_PATH = "../data/index.faiss"
METADATA_PATH = "../data/metadata.json"
SCORE_THRESHOLD = 0.5

model = None
store = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global model, store
    logger.info("Loading BGE model...")
    model = SentenceTransformer("BAAI/bge-small-en-v1.5")
    model.encode(["warmup"])
    logger.info("Model ready")

    store = VectorStore(INDEX_PATH, METADATA_PATH)
    logger.info("Vector store ready, total vectors: %d", store.index.ntotal)
    
    yield

app = FastAPI(lifespan=lifespan)

class EmbedRequest(BaseModel):
    text: str

class IngestRequest(BaseModel):
    command: str

@app.get("/health")
def health():
    return {
        "status": "ok",
        "model_loaded": model is not None,
        "vectors_indexed": store.index.ntotal if store else 0
    }

@app.post("/embed")
def embed(req: EmbedRequest):
    vector = model.encode([req.text], normalize_embeddings=True)[0]
    return {"embedding": vector.tolist()}

@app.post("/ingest")
def ingest(req: IngestRequest):
    try:
        chunks = parse_manpage(req.command)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    
    if not chunks:
        raise HTTPException(status_code=422, detail=f"no parsable content for '{req.command}'")

    texts = [chunk["text"] for chunk in chunks]
    embeddings = model.encode(texts, normalize_embeddings=True).tolist()

    store.replace_command(req.command, chunks, embeddings)
    logger.info("ingested '%s': %d chunks", req.command, len(chunks))

    return {"command": req.command, "chunks_indexed": len(chunks)}

@app.get("/search")
def search(q: str, top_k: int = 5):
    query_vector = model.encode([q], normalize_embeddings=True)[0].tolist()
    results = store.search(query_vector, top_k=top_k)
    filtered = [r for r in results if r["score"] >= SCORE_THRESHOLD]

    return {"query": q, "results": filtered}
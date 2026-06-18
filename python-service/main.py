import os
import logging
import sys
from contextlib import asynccontextmanager
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

model = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global model
    logger.info("Loading BGE model...")
    model = SentenceTransformer("BAAI/bge-small-en-v1.5")
    model.encode(["warmup"])
    logger.info("Model ready")
    yield

app = FastAPI(lifespan=lifespan)

class EmbedRequest(BaseModel):
    text: str

@app.get("/health")
def health():
    return {"status": "ok", "model_loaded": model is not None}

@app.post("/embed")
def embed(req: EmbedRequest):
    vector = model.encode([req.text], normalize_embeddings=True)[0]
    return {"embedding": vector.tolist()}
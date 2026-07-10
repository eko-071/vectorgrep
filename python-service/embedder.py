import logging

from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

class Embedder:
    def __init__(self, model_name: str = "BAAI/bge-small-en-v1.5"):
        logger.info("Loading BGE model: %s", model_name)
        self.model = SentenceTransformer(model_name)
        self.model.encode(["warmup"])
        logger.info("Model ready")

    def embed(self, text: str) -> list[float]:
        return self.model.encode([text], normalize_embeddings=True)[0].tolist()

    def embed_batch(self, texts: list[str]) -> list[list[float]]:
        return self.model.encode(texts, normalize_embeddings=True).tolist()

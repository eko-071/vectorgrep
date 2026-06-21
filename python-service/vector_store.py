import json
import os
import numpy as np
import faiss

EMBEDDING_DIM = 384

class VectorStore:
    """Wraps a FAISS flat index with a parallel metadata store, keyed by command"""

    def __init__(self, index_path: str, metadata_path: str):
        self.index_path = index_path
        self.metadata_path = metadata_path
        self.index = None
        self.metadata: list[dict] = [] # ith element would correspond to ith vector in index
        self._load_or_create()
    
    def _load_or_create(self):
        if os.path.exists(self.index_path) and os.path.exists(self.metadata_path):
            self.index = faiss.read_index(self.index_path)
            with open(self.metadata_path) as f:
                self.metadata = json.load(f)
        else:
            self.index = faiss.IndexFlatIP(EMBEDDING_DIM)
            self.metadata = []
    
    def _persist(self):
        faiss.write_index(self.index, self.index_path)
        with open(self.metadata_path, "w") as f:
            json.dump(self.metadata, f)

    def _all_vectors(self) -> np.ndarray:
        """Reconstruct every stored vector."""
        if self.index.ntotal == 0:
            return np.empty((0, EMBEDDING_DIM), dtype=np.float32)
        return np.array([self.index.reconstruct(i) for i in range(self.index.ntotal)])
    
    def replace_command(self, command: str, chunks: list[dict], embeddings: list[list[float]]):
        """Drop all existing entries for this command, then add the new chunks + embeddings."""
        surviving = [
            (meta, vec)
            for meta, vec in zip(self.metadata, self._all_vectors())
            if meta["metadata"]["command"] != command
        ]
        new_index = faiss.IndexFlatIP(EMBEDDING_DIM)
        new_metadata = []

        if surviving:
            surviving_metas, surviving_vecs = zip(*surviving)
            new_index.add(np.array(surviving_vecs, dtype=np.float32))
            new_metadata.extend(surviving_metas)

        vectors = np.array(embeddings, dtype=np.float32)
        new_index.add(vectors)
        new_metadata.extend(chunks)

        self.index = new_index
        self.metadata = new_metadata
        self._persist()

    def search(self, query_vector: list[float], top_k: int = 5) -> list[dict]:
        if self.index.ntotal == 0:
            return []

        query = np.array([query_vector], dtype=np.float32)
        scores, indices = self.index.search(query, min(top_k, self.index.ntotal))

        results = []
        for score, idx in zip(scores[0], indices[0]):
            if idx == -1:
                continue
            entry = self.metadata[idx]
            results.append({
                "id": entry["id"],
                "text": entry["text"],
                "metadata": entry["metadata"],
                "score": float(score),
            })
        return results
import json
import os
import numpy as np
import faiss

EMBEDDING_DIM = 384

class VectorStore:
    """FAISS IndexIDMap wrapper. Each vector gets a unique int64 ID at insert
    time so we can remove a specific command's vectors without a full rebuild."""

    def __init__(self, index_path: str, metadata_path: str):
        self.index_path = index_path
        self.metadata_path = metadata_path
        self.index = None
        self.next_id = 0
        self.command_ids: dict[str, list[int]] = {}
        self.metadata: dict[str, dict] = {}
        self._load_or_create()

    def _load_or_create(self):
        if os.path.exists(self.index_path) and os.path.exists(self.metadata_path):
            self.index = faiss.read_index(self.index_path)
            with open(self.metadata_path) as f:
                data = json.load(f)
            self.next_id = data["next_id"]
            self.command_ids = data["command_ids"]
            self.metadata = data["metadata"]
        else:
            self.index = faiss.IndexIDMap(faiss.IndexFlatIP(EMBEDDING_DIM))
            self.next_id = 0
            self.command_ids = {}
            self.metadata = {}

    def _persist(self):
        faiss.write_index(self.index, self.index_path)
        data = {
            "next_id": self.next_id,
            "command_ids": self.command_ids,
            "metadata": self.metadata,
        }
        with open(self.metadata_path, "w") as f:
            json.dump(data, f)

    def replace_command(self, command: str, chunks: list[dict], embeddings: list[list[float]]):
        old_ids = self.command_ids.pop(command, [])
        if old_ids:
            selector = faiss.IDSelectorArray(np.array(old_ids, dtype=np.int64))
            self.index.remove_ids(selector)

        vectors = np.array(embeddings, dtype=np.float32)
        ids = np.arange(self.next_id, self.next_id + len(vectors), dtype=np.int64)
        self.index.add_with_ids(vectors, ids)

        for chunk, faiss_id in zip(chunks, ids):
            self.metadata[str(faiss_id)] = chunk

        self.command_ids[command] = ids.tolist()
        self.next_id += len(vectors)
        self._persist()

    def search(self, query_vector: list[float], top_k: int = 5) -> list[dict]:
        if self.index.ntotal == 0:
            return []

        query = np.array([query_vector], dtype=np.float32)
        scores, indices = self.index.search(query, min(top_k, self.index.ntotal))

        results = []
        for score, faiss_id in zip(scores[0], indices[0]):
            if faiss_id == -1:
                continue
            entry = self.metadata[str(faiss_id)]
            results.append({
                "id": entry["id"],
                "text": entry["text"],
                "metadata": entry["metadata"],
                "score": float(score),
            })
        return results

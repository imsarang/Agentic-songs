import json
import logging
import sys
import types

import langchain_community.chat_models

vertex = types.ModuleType("langchain_community.chat_models.vertexai")
vertex.ChatVertexAI = type("ChatVertexAI", (), {})
sys.modules["langchain_community.chat_models.vertexai"] = vertex

from datasets import Dataset
from ragas import evaluate
from ragas.metrics import (
    Faithfulness,
    LLMContextPrecisionWithReference,
    LLMContextRecall,
    ResponseRelevancy,
)
from ragas.testset import TestsetGenerator
from ragas.run_config import RunConfig
from ragas.testset.graph import KnowledgeGraph, Node,NodeType
from ragas.testset.transforms.default import default_transforms
from ragas.testset.transforms.engine import Parallel, apply_transforms
from agent import ask, retrieve
from config import GRAPH, OUT, model
from embed import EmbedServer
from store import pages

log = logging.getLogger(__name__)
RUN = RunConfig(max_workers=1)
SIZE=10
EVALUATE_RUN = RunConfig(max_workers=1)

def save(rows: list[dict]) -> None:
    OUT.write_text(json.dumps(rows, indent=2))

def save_graph(kg: KnowledgeGraph, done: int) -> None:
    tmp = GRAPH.with_suffix(".tmp")
    kg.save(tmp)
    data = json.loads(tmp.read_text())
    data["step"] = done
    tmp.write_text(json.dumps(data))

    tmp.replace(GRAPH)

def steps(transforms):
    if isinstance(transforms,  Parallel):
        yield from steps(transforms.transformations)
    elif isinstance(transforms,  list):
        for item in transforms:
            yield from steps(item)
    else:
        yield transforms

def generate(embeddings: EmbedServer, rows: list[dict]) -> list[dict]:
    generator = TestsetGenerator.from_langchain(model(), embeddings)
    docs = pages()
    transforms = list(steps(default_transforms(
        docs, generator.llm, generator.embedding_model
    )))

    if GRAPH.exists():
        kg = KnowledgeGraph.load(GRAPH)
        done = json.loads(GRAPH.read_text()).get("step", 0)
        log.info("resume graph nodes=%s step=%s/%s", len(kg.nodes), done, len(transforms))

    else:
        kg = KnowledgeGraph(nodes=[
            Node(
                type=NodeType.DOCUMENT,
                properties={
                    "page_content": doc.page_content,
                    "document_metadata": doc.metadata,
                },
            )
            for doc in docs
        ])

        done = 0
    
    for index, transform in enumerate(transforms):
        if index<done:
            continue
        
        log.info("step %s/%s %s", index + 1, len(transforms), type(transform).__name__)
        apply_transforms(kg, transform, run_config=RUN)
        save_graph(kg, index + 1)

    generator.knowledge_graph = kg

    while len(rows) < SIZE:
        made = generator.generate(testset_size=1, run_config=RUN).to_list()
        rows.extend(made)
        save(rows)
        log.info("questions %s/%s", len(rows), SIZE)

    return rows



def main() -> None:
    logging.basicConfig(level=logging.INFO)
    embeddings = EmbedServer()
    rows = json.loads(OUT.read_text()) if OUT.exists() else []

    if len(rows) < SIZE:
        rows = generate(embeddings, rows)
        
    for row in rows:
        if "response" in row:
            continue
        question = row["user_input"]

        if "retrieved_contexts" not in row:
            row["retrieved_contexts"] = retrieve(question)
            save(rows)
        row["response"] = ask(question)
        save(rows)

    result = evaluate(
        Dataset.from_list(rows),
        metrics=[
            Faithfulness(),
            ResponseRelevancy(),
            LLMContextPrecisionWithReference(),
            LLMContextRecall(),
        ],
        llm=model(),
        embeddings=embeddings,
        run_config=EVALUATE_RUN
    )
    print(result)


if __name__ == "__main__":
    main()

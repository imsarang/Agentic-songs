from __future__ import annotations

import logging
from collections.abc import AsyncIterator
from typing import Any

from langchain_core.messages import HumanMessage
from langsmith import trace

import shanks_agent.memory as memory
from shanks_agent.graph.domains import build_all
from shanks_agent.graph.nodes.agent import seed_messages
from shanks_agent.graph.nodes.writer import draft_from
from shanks_agent.graph.orchestrator import merge, plan
from shanks_agent.middleware.policy.domain_defaults import letter_to_send
from shanks_agent.llm.failover import FailoverLLM, build_failover_llm
from shanks_agent.observability import new_run_id

logger = logging.getLogger(__name__)

_GRAPHS: dict | None = None
_LLM: FailoverLLM | None = None


def get_graphs():
    global _GRAPHS, _LLM
    if _GRAPHS is None:
        _LLM = build_failover_llm()
        _GRAPHS = build_all(_LLM)
        logger.info("compiled domain graphs: %s", list(_GRAPHS))
    return _GRAPHS


def get_graph():
    return get_graphs()["spotify"]


async def run_turn(
    user_text: str,
    *,
    thread_id: str = "default",
    run_id: str | None = None,
    domain: str | None = None,
    graph=None,
) -> AsyncIterator[dict[str, Any]]:
    graphs = get_graphs()
    llm = _LLM or build_failover_llm()
    run_uuid = new_run_id(run_id)
    run_key = str(run_uuid)

    hint = await memory.index(thread_id)
    follow = await memory.last_domain(thread_id) if memory.continues(user_text) else None

    async with trace(
        "shanks-turn",
        run_type="chain",
        run_id=run_uuid,
        inputs={"text": user_text, "thread_id": thread_id},
        tags=["shanks"],
        metadata={"thread_id": thread_id, "run_id": run_key},
    ) as ls_run:
        yield {"type": "run", "run_id": run_key, "thread_id": thread_id}

        async with trace(
            "orchestrator-plan",
            run_type="chain",
            inputs={"text": user_text},
            tags=["shanks", "orchestrator"],
            metadata={"run_id": run_key},
        ) as plan_run:
            domains = await plan(llm, user_text, hint, forced=domain or follow)
            if letter_to_send(user_text):
                rest = [d for d in domains if d not in ("creative", "google")]
                domains = ["creative", "google", *rest]
            plan_run.end(outputs={"domains": domains})

        ls_run.name = "shanks-" + "+".join(domains)
        ls_run.add_metadata({"domains": domains, "subagents": domains})
        ls_run.add_tags(list(domains))
        logger.info(
            "turn start run_id=%s thread_id=%s domains=%s text=%s",
            run_key,
            thread_id,
            domains,
            user_text[:200],
        )
        yield {"type": "domain", "domain": domains[0]}

        stubs: list[str] = []
        carry = ""
        saved: list[tuple[str, str]] = []
        n = 0

        for d in domains:
            g = graph if (graph and len(domains) == 1) else graphs.get(d) or graphs["spotify"]
            past = memory.history(await memory.load(thread_id, d))
            messages = seed_messages(user_text)
            if past:
                messages = [HumanMessage(content=f"Recent {d} turns:\n{past}"), *messages]
            letter = carry
            if d == "google" and not letter:
                letter = memory.latest_note(await memory.load(thread_id, "creative"))
            if letter and d != "creative":
                messages = [
                    HumanMessage(
                        content=(
                            "Creative draft — copy subject and body into the mail. Do not rewrite.\n"
                            + letter
                        )
                    ),
                    *messages,
                ]
            if d == "google" and letter_to_send(user_text):
                messages = [
                    HumanMessage(
                        content=(
                            "Letter flow: create a draft from the Creative draft, then send it "
                            "when the user named an email or a person."
                        )
                    ),
                    *messages,
                ]
            domain_values: dict[str, Any] | None = None
            n += 1
            yield {"type": "subagent", "name": d}

            async with trace(
                f"subagent-{d}",
                run_type="chain",
                inputs={"text": user_text, "domain": d},
                tags=["shanks", "subagent", d],
                metadata={
                    "subagent": d,
                    "domain": d,
                    "thread_id": thread_id,
                    "run_id": run_key,
                },
            ) as child:
                saw = False
                async for mode, data in g.astream(
                    {
                        "messages": messages,
                        "domain": d,
                        "next_tool": None,
                        "tool_args": {},
                        "present": "raw",
                    },
                    config={
                        "configurable": {"thread_id": thread_id},
                        "tags": ["shanks", "subagent", d],
                        "metadata": {
                            "subagent": d,
                            "domain": d,
                            "thread_id": thread_id,
                            "run_id": run_key,
                        },
                    },
                    stream_mode=["custom", "values"],
                ):
                    if mode == "values" and isinstance(data, dict):
                        domain_values = data
                        continue
                    if mode != "custom" or not isinstance(data, dict):
                        continue
                    if data.get("type") == "assistant_message":
                        saw = True
                        stubs.append(str(data.get("text") or "")[:800])
                    n += 1
                    yield data

                if not saw and domain_values is not None:
                    for msg in reversed(domain_values.get("messages") or []):
                        if getattr(msg, "type", None) == "ai":
                            t = str(msg.content)
                            stubs.append(t[:800])
                            n += 1
                            yield {"type": "assistant_message", "text": t}
                            break
                child.end(outputs={"domain": d, "events": n})
            if d == "creative":
                carry = draft_from(domain_values)
            saved.append((d, memory.note_from((domain_values or {}).get("messages"))))

        if len(domains) > 1 and stubs and "creative" not in domains:
            async with trace(
                "orchestrator-merge",
                run_type="chain",
                inputs={"stubs": len(stubs)},
                tags=["shanks", "orchestrator", "merge"],
                metadata={"run_id": run_key, "subagents": domains},
            ) as merge_run:
                text = await merge(llm, user_text, stubs)
                merge_run.end(outputs={"text": text[:500]})
            n += 1
            yield {"type": "assistant_message", "text": text}

        for d, note in saved:
            await memory.save(thread_id, d, user_text, note)
        n += 1
        yield {"type": "turn_done", "run_id": run_key}
        ls_run.end(outputs={"events": n, "domains": domains})

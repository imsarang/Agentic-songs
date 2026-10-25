from __future__ import annotations

import asyncio
import uuid

from rich.console import Console
from rich.markup import escape
from rich.panel import Panel

from shanks_ui.config import get_settings
from shanks_ui.transport import stream_turn
from shanks_ui.voice import PushToTalk

console = Console()


def _print_event(event: dict) -> None:
    kind = event.get("type")
    if kind == "tool_start":
        name = event.get("name", "?")
        args = event.get("args") or {}
        console.print(f"[cyan]→ {name}[/] {escape(str(args))}")
    elif kind == "tool_end":
        name = event.get("name", "?")
        ok = event.get("ok")
        mark = "[green]ok[/]" if ok else "[red]fail[/]"
        preview = escape(str(event.get("preview") or "")[:400])
        console.print(f"[cyan]← {name}[/] {mark}\n{preview}")
    elif kind == "assistant_message":
        text = escape(str(event.get("text") or ""))
        console.print(Panel(text, title="assistant", border_style="green"))
    elif kind == "error":
        console.print(f"[red]error:[/] {escape(str(event.get('message') or ''))}")
    elif kind == "run":
        console.print(
            f"[dim]run {event.get('run_id')}  thread {event.get('thread_id')}[/]"
        )
    elif kind == "turn_done":
        console.print("[dim]— turn done —[/]\n")
    else:
        console.print(f"[dim]{escape(str(event))}[/]")
    console.file.flush()


async def run_repl() -> None:
    settings = get_settings()
    thread_id = settings.thread_id or str(uuid.uuid4())
    console.print(
        Panel(
            f"agent: {settings.agent_turn_url}\n"
            f"thread: {thread_id}\n"
            "hold Left Ctrl to talk, or type a line\n"
            "quit with :q / exit / Ctrl-C",
            title="shanks-ui",
            border_style="blue",
        )
    )
    loop = asyncio.get_running_loop()
    queue: asyncio.Queue[str | None] = asyncio.Queue()
    talk = PushToTalk(
        loop,
        queue,
        lambda msg: console.print(f"[yellow]{escape(msg)}[/]") if msg else None,
    )
    if err := talk.start():
        console.print(f"[red]{escape(err)}[/]")
    stdin_task = asyncio.create_task(_stdin(queue))
    try:
        while True:
            item = await queue.get()
            if item is None:
                break
            text = item.strip()
            if not text:
                continue
            if text in {":q", "quit", "exit"}:
                break
            console.print(f"[bold]> [/]{escape(text)}")
            await _one_turn(settings.agent_turn_url, text, thread_id)
    finally:
        talk.stop()
        stdin_task.cancel()


async def _stdin(queue: asyncio.Queue[str | None]) -> None:
    while True:
        try:
            line = await asyncio.to_thread(console.input, "[bold]> [/]")
        except (EOFError, KeyboardInterrupt):
            queue.put_nowait(None)
            return
        queue.put_nowait(line)


async def _one_turn(url: str, text: str, thread_id: str) -> None:
    try:
        async for event in stream_turn(url, text, thread_id=thread_id, run_id=str(uuid.uuid4())):
            _print_event(event)
            if event.get("type") == "turn_done":
                break
    except Exception as exc:  # noqa: BLE001
        console.print(f"[red]transport:[/] {escape(str(exc))}")

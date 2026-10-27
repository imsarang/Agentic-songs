from __future__ import annotations

import asyncio
import uuid

from rich.console import Console
from rich.markup import escape
from rich.panel import Panel

from shanks_ui.config import get_settings
from shanks_ui.speak import load, speak
from shanks_ui.transport import stream_turn
from shanks_ui.voice import PushToTalk

console = Console()


def _print_event(event: dict, subagent: str = "") -> None:
    kind = event.get("type")
    who = str(event.get("subagent") or event.get("name") or subagent or "").strip()
    if kind == "subagent":
        console.print(f"[bold cyan]{escape(who)}[/]")
    elif kind in ("tool_start", "tool_end"):
        return
    elif kind == "assistant_message":
        text = escape(str(event.get("text") or ""))
        console.print(Panel(text, title=who or "assistant", border_style="green"))
    elif kind == "error":
        console.print(f"[red]error:[/] {escape(str(event.get('message') or ''))}")
    elif kind == "domain":
        return
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
            "hold Num Lock to talk, or type a line\n"
            "quit with :q / exit / Ctrl-C",
            title="shanks-ui",
            border_style="blue",
        )
    )
    if settings.speak_replies:
        console.print("[yellow]loading voice…[/]")
        await asyncio.to_thread(load)
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
            await _one_turn(settings.agent_turn_url, text, thread_id, settings.speak_replies)
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


async def _one_turn(url: str, text: str, thread_id: str, speak_replies: bool) -> None:
    domain = ""
    try:
        async for event in stream_turn(url, text, thread_id=thread_id, run_id=str(uuid.uuid4())):
            kind = event.get("type")
            if kind in ("domain", "subagent"):
                domain = str(event.get("name") or event.get("domain") or domain)
            _print_event(event, domain)
            if speak_replies and event.get("type") == "assistant_message" and domain == "machine":
                console.print("[yellow]speaking…[/]")
                try:
                    await asyncio.to_thread(speak, str(event.get("text") or ""))
                except Exception as exc:  # noqa: BLE001
                    console.print(f"[red]voice:[/] {escape(str(exc))}")
            if event.get("type") == "turn_done":
                break
    except Exception as exc:  # noqa: BLE001
        console.print(f"[red]transport:[/] {escape(str(exc))}")

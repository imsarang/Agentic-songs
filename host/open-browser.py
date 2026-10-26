#!/usr/bin/env python3
"""Open a Google login URL in the host browser.

The Google MCP runs inside Docker Desktop, which cannot launch this desktop's
browser. Login POSTs the URL here.
"""

from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import subprocess

HOST = "0.0.0.0"
PORT = 4390
ALLOWED = "https://accounts.google.com/"


class Handler(BaseHTTPRequestHandler):
    def do_POST(self) -> None:
        if self.path != "/open":
            self.send_response(404)
            self.end_headers()
            return
        length = int(self.headers.get("Content-Length", "0"))
        url = self.rfile.read(length).decode().strip()
        if not url.startswith(ALLOWED):
            self.send_response(400)
            self.end_headers()
            return
        subprocess.Popen(["xdg-open", url])
        self.send_response(204)
        self.end_headers()

    def log_message(self, fmt: str, *args) -> None:
        print(f"open-browser {self.address_string()} {fmt % args}")


if __name__ == "__main__":
    print(f"open-browser listening on {HOST}:{PORT}")
    ThreadingHTTPServer((HOST, PORT), Handler).serve_forever()

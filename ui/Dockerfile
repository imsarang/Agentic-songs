FROM python:3.12-slim
WORKDIR /app
COPY pyproject.toml ./
RUN --mount=type=cache,target=/root/.cache/pip \
    python -c "import subprocess,tomllib; deps=tomllib.load(open('pyproject.toml','rb'))['project']['dependencies']; subprocess.check_call(['pip','install','--timeout','300','--retries','10',*deps])"
COPY src ./src
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --timeout 300 --retries 10 --no-deps .
ENV AGENT_TURN_URL=http://agent:8001/turn
ENTRYPOINT ["shanks-ui"]

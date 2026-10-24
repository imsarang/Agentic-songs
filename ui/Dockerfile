FROM python:3.12-slim
WORKDIR /app
COPY pyproject.toml ./
COPY src ./src
RUN pip install --no-cache-dir .
ENV AGENT_TURN_URL=http://agent:8001/turn
ENTRYPOINT ["shanks-ui"]
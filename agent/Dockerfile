FROM python:3.12-slim

WORKDIR /app
COPY pyproject.toml ./
COPY src ./src
RUN pip install --no-cache-dir .

ENV AGENT_HOST=0.0.0.0
ENV AGENT_PORT=8001
EXPOSE 8001
CMD ["shanks-agent"]

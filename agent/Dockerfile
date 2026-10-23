FROM python:3.12-slim

WORKDIR /app
COPY pyproject.toml ./
COPY src ./src
ENV PYTHONPATH=/app/src
RUN pip install --no-cache-dir "fastapi>=0.115" "uvicorn[standard]>=0.32" "pydantic-settings>=2.6" "python-dotenv>=1.0"

ENV AGENT_HOST=0.0.0.0
ENV AGENT_PORT=8000
EXPOSE 8000
CMD ["python", "-m", "shanks_agent"]
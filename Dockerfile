FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY pyproject.toml .
RUN pip install --no-cache-dir .

# Copy application code
COPY playlist_agent/ playlist_agent/

# Create non-root user
RUN useradd -r -u 1000 playlist-agent && \
    chown -R playlist-agent:playlist-agent /app

USER playlist-agent

# Default command
CMD ["python", "-m", "playlist_agent.cli", "--help"]

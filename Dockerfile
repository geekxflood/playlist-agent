FROM python:3.12-slim

WORKDIR /app

# Copy all project files (pyproject.toml references README.md)
COPY pyproject.toml README.md LICENSE ./
COPY program_director/ program_director/

# Install dependencies
RUN pip install --no-cache-dir .

# Create non-root user
RUN useradd -r -u 1000 program-director && \
    chown -R program-director:program-director /app

USER program-director

# Default command
CMD ["python", "-m", "program_director.cli", "--help"]

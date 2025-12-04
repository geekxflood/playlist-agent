# Playlist Agent

AI-powered playlist generator for ErsatzTV using LangChain and Ollama.

## Overview

Playlist Agent generates themed Smart Collections in ErsatzTV by:
1. Querying Radarr and Sonarr for available media with full metadata
2. Using an LLM (Ollama) to intelligently select content matching a theme
3. Creating Smart Collections in ErsatzTV for the curated content

## Features

- **AI-Powered Selection**: Uses LangChain with Ollama for intelligent media curation
- **Theme-Based Playlists**: Configure themes with keywords for matching
- **Radarr/Sonarr Integration**: Fetches rich metadata (genres, ratings, runtime, overviews)
- **ErsatzTV Smart Collections**: Automatically creates and updates Smart Collections
- **CLI Interface**: Easy command-line usage for manual or scheduled execution

## Installation

### Using pip

```bash
pip install .
```

### Using Docker

```bash
docker pull ghcr.io/geekxflood/playlist-agent:latest
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `OLLAMA_URL` | Ollama API URL | Yes |
| `OLLAMA_MODEL` | Ollama model name (default: llama3:8b) | No |
| `ERSATZTV_URL` | ErsatzTV API URL | Yes |
| `RADARR_URL` | Radarr API URL | Yes |
| `RADARR_API_KEY` | Radarr API key | Yes |
| `SONARR_URL` | Sonarr API URL | Yes |
| `SONARR_API_KEY` | Sonarr API key | Yes |

### Config File

Create a `config.yaml`:

```yaml
ollama:
  url: "http://localhost:11434"
  model: "llama3:8b"

ersatztv:
  url: "http://localhost:8409"

radarr:
  url: "http://localhost:7878"
  # api_key from environment variable

sonarr:
  url: "http://localhost:8989"
  # api_key from environment variable

themes:
  - name: "sci-fi-night"
    description: "Science fiction themed evening"
    duration: 180
    keywords:
      - "Science Fiction"
      - "space"
      - "alien"

  - name: "horror-marathon"
    description: "Horror movie marathon"
    duration: 240
    keywords:
      - "Horror"
      - "Thriller"
```

## Usage

### CLI Commands

```bash
# Generate playlist for a specific theme
playlist-agent generate --theme sci-fi-night

# Generate playlists for all themes
playlist-agent generate --all-themes

# Dry run (preview without applying)
playlist-agent generate --theme sci-fi-night --dry-run

# Scan media library
playlist-agent scan

# List configured themes
playlist-agent themes
```

### Docker Usage

```bash
docker run --rm \
  -e RADARR_API_KEY=your-key \
  -e SONARR_API_KEY=your-key \
  -e OLLAMA_URL=http://ollama:11434 \
  -e ERSATZTV_URL=http://ersatztv:8409 \
  -e RADARR_URL=http://radarr:7878 \
  -e SONARR_URL=http://sonarr:8989 \
  -v /path/to/config.yaml:/app/config/config.yaml \
  ghcr.io/geekxflood/playlist-agent:latest \
  python -m playlist_agent.cli generate --all-themes
```

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Radarr    │────▶│  Playlist   │────▶│  ErsatzTV   │
│   (Movies)  │     │    Agent    │     │  (Smart     │
└─────────────┘     │             │     │ Collections)│
                    │  LangChain  │     └─────────────┘
┌─────────────┐     │      +      │
│   Sonarr    │────▶│   Ollama    │
│ (TV/Anime)  │     │             │
└─────────────┘     └─────────────┘
```

## Development

### Setup

```bash
# Create virtual environment
python -m venv .venv
source .venv/bin/activate

# Install with dev dependencies
pip install -e ".[dev]"
```

### Linting

```bash
ruff check .
ruff format .
mypy playlist_agent/
```

### Testing

```bash
pytest
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [ErsatzTV](https://github.com/ErsatzTV/ErsatzTV) - IPTV server for custom channels
- [Radarr](https://github.com/Radarr/Radarr) - Movie management
- [Sonarr](https://github.com/Sonarr/Sonarr) - TV show management
- [Ollama](https://github.com/ollama/ollama) - Local LLM runtime
- [LangChain](https://github.com/langchain-ai/langchain) - LLM framework

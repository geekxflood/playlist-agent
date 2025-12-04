"""Media scanner using Radarr and Sonarr APIs for metadata."""

from dataclasses import dataclass, field

import httpx


@dataclass
class MovieItem:
    """Represents a movie with metadata from Radarr."""

    id: int
    title: str
    year: int
    genres: list[str] = field(default_factory=list)
    runtime: int = 0  # minutes
    overview: str = ""
    imdb_rating: float = 0.0
    tmdb_rating: float = 0.0
    has_file: bool = False

    def __str__(self) -> str:
        return f"{self.title} ({self.year})"


@dataclass
class SeriesItem:
    """Represents a TV series with metadata from Sonarr."""

    id: int
    title: str
    year: int
    genres: list[str] = field(default_factory=list)
    runtime: int = 0  # minutes per episode
    overview: str = ""
    rating: float = 0.0
    episode_count: int = 0
    series_type: str = "standard"  # standard, anime, daily

    def __str__(self) -> str:
        return f"{self.title} ({self.year})"


class ArrAPIClient:
    """Base client for *arr APIs (Radarr, Sonarr)."""

    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.client = httpx.Client(timeout=30.0)

    def _get(self, endpoint: str) -> list | dict:
        """Make a GET request to the API."""
        url = f"{self.base_url}/api/v3/{endpoint}"
        headers = {"X-Api-Key": self.api_key}
        response = self.client.get(url, headers=headers)
        response.raise_for_status()
        return response.json()

    def close(self) -> None:
        """Close the HTTP client."""
        self.client.close()


class RadarrClient(ArrAPIClient):
    """Client for Radarr API."""

    def get_movies(self) -> list[MovieItem]:
        """Get all movies from Radarr."""
        data = self._get("movie")
        movies = []

        for item in data:
            if not item.get("hasFile", False):
                continue  # Skip movies without files

            ratings = item.get("ratings", {})
            imdb = ratings.get("imdb", {}).get("value", 0.0)
            tmdb = ratings.get("tmdb", {}).get("value", 0.0)

            movies.append(
                MovieItem(
                    id=item.get("id", 0),
                    title=item.get("title", "Unknown"),
                    year=item.get("year", 0),
                    genres=item.get("genres", []),
                    runtime=item.get("runtime", 0),
                    overview=item.get("overview", ""),
                    imdb_rating=imdb,
                    tmdb_rating=tmdb,
                    has_file=True,
                )
            )

        return movies

    def get_movies_by_genre(self, genre: str) -> list[MovieItem]:
        """Get movies filtered by genre."""
        movies = self.get_movies()
        genre_lower = genre.lower()
        return [m for m in movies if any(g.lower() == genre_lower for g in m.genres)]


class SonarrClient(ArrAPIClient):
    """Client for Sonarr API."""

    def get_series(self) -> list[SeriesItem]:
        """Get all series from Sonarr."""
        data = self._get("series")
        series_list = []

        for item in data:
            # Skip series with no episodes
            stats = item.get("statistics", {})
            episode_count = stats.get("episodeFileCount", 0)
            if episode_count == 0:
                continue

            ratings = item.get("ratings", {})

            series_list.append(
                SeriesItem(
                    id=item.get("id", 0),
                    title=item.get("title", "Unknown"),
                    year=item.get("year", 0),
                    genres=item.get("genres", []),
                    runtime=item.get("runtime", 0),
                    overview=item.get("overview", ""),
                    rating=ratings.get("value", 0.0),
                    episode_count=episode_count,
                    series_type=item.get("seriesType", "standard"),
                )
            )

        return series_list

    def get_anime(self) -> list[SeriesItem]:
        """Get anime series (series with 'Anime' genre or anime series type)."""
        series = self.get_series()
        return [
            s
            for s in series
            if s.series_type == "anime" or "anime" in [g.lower() for g in s.genres]
        ]

    def get_series_by_genre(self, genre: str) -> list[SeriesItem]:
        """Get series filtered by genre."""
        series = self.get_series()
        genre_lower = genre.lower()
        return [s for s in series if any(g.lower() == genre_lower for g in s.genres)]


class MediaLibrary:
    """Combined media library using Radarr and Sonarr APIs."""

    def __init__(
        self,
        radarr_url: str,
        radarr_api_key: str,
        sonarr_url: str,
        sonarr_api_key: str,
    ):
        self.radarr = RadarrClient(radarr_url, radarr_api_key)
        self.sonarr = SonarrClient(sonarr_url, sonarr_api_key)

        # Cache for media items
        self._movies: list[MovieItem] | None = None
        self._series: list[SeriesItem] | None = None

    @property
    def movies(self) -> list[MovieItem]:
        """Get all movies (cached)."""
        if self._movies is None:
            self._movies = self.radarr.get_movies()
        return self._movies

    @property
    def series(self) -> list[SeriesItem]:
        """Get all series (cached)."""
        if self._series is None:
            self._series = self.sonarr.get_series()
        return self._series

    @property
    def anime(self) -> list[SeriesItem]:
        """Get anime series."""
        return [
            s
            for s in self.series
            if s.series_type == "anime" or "anime" in [g.lower() for g in s.genres]
        ]

    @property
    def tv_shows(self) -> list[SeriesItem]:
        """Get non-anime TV series."""
        anime_ids = {a.id for a in self.anime}
        return [s for s in self.series if s.id not in anime_ids]

    def get_media_summary(self) -> str:
        """Get a summary of available media for LLM context."""
        lines = ["# Available Media Library\n"]

        # Movies section
        lines.append("## Movies")
        lines.append("| Title | Year | Genres | Rating | Runtime |")
        lines.append("|-------|------|--------|--------|---------|")

        # Sort by rating, take top 100
        sorted_movies = sorted(self.movies, key=lambda m: m.imdb_rating or 0, reverse=True)
        for movie in sorted_movies[:100]:
            genres = ", ".join(movie.genres[:3]) if movie.genres else "-"
            rating = f"{movie.imdb_rating:.1f}" if movie.imdb_rating else "-"
            lines.append(f"| {movie.title} | {movie.year} | {genres} | {rating} | {movie.runtime}m |")

        lines.append(f"\n*Total movies available: {len(self.movies)}*\n")

        # TV Shows section
        lines.append("## TV Shows")
        lines.append("| Title | Year | Genres | Rating | Episodes |")
        lines.append("|-------|------|--------|--------|----------|")

        sorted_shows = sorted(self.tv_shows, key=lambda s: s.rating or 0, reverse=True)
        for show in sorted_shows[:50]:
            genres = ", ".join(show.genres[:3]) if show.genres else "-"
            rating = f"{show.rating:.1f}" if show.rating else "-"
            lines.append(f"| {show.title} | {show.year} | {genres} | {rating} | {show.episode_count} |")

        lines.append(f"\n*Total TV shows available: {len(self.tv_shows)}*\n")

        # Anime section
        lines.append("## Anime")
        lines.append("| Title | Year | Genres | Rating | Episodes |")
        lines.append("|-------|------|--------|--------|----------|")

        sorted_anime = sorted(self.anime, key=lambda a: a.rating or 0, reverse=True)
        for anime in sorted_anime[:50]:
            genres = ", ".join([g for g in anime.genres[:3] if g.lower() != "anime"]) or "-"
            rating = f"{anime.rating:.1f}" if anime.rating else "-"
            lines.append(f"| {anime.title} | {anime.year} | {genres} | {rating} | {anime.episode_count} |")

        lines.append(f"\n*Total anime available: {len(self.anime)}*\n")

        return "\n".join(lines)

    def get_genre_stats(self) -> dict[str, int]:
        """Get count of media items per genre."""
        genre_counts: dict[str, int] = {}

        for movie in self.movies:
            for genre in movie.genres:
                genre_counts[genre] = genre_counts.get(genre, 0) + 1

        for series in self.series:
            for genre in series.genres:
                genre_counts[genre] = genre_counts.get(genre, 0) + 1

        return dict(sorted(genre_counts.items(), key=lambda x: x[1], reverse=True))

    def search_by_theme(self, theme_keywords: list[str]) -> dict[str, list]:
        """Search media matching theme keywords in genres or overview."""
        matching_movies = []
        matching_shows = []
        matching_anime = []

        keywords_lower = [k.lower() for k in theme_keywords]

        for movie in self.movies:
            if self._matches_theme(movie.genres, movie.overview, keywords_lower):
                matching_movies.append(movie)

        for show in self.tv_shows:
            if self._matches_theme(show.genres, show.overview, keywords_lower):
                matching_shows.append(show)

        for anime in self.anime:
            if self._matches_theme(anime.genres, anime.overview, keywords_lower):
                matching_anime.append(anime)

        return {
            "movies": matching_movies,
            "shows": matching_shows,
            "anime": matching_anime,
        }

    def _matches_theme(
        self, genres: list[str], overview: str, keywords: list[str]
    ) -> bool:
        """Check if media matches theme keywords."""
        genres_lower = [g.lower() for g in genres]
        overview_lower = overview.lower()

        for keyword in keywords:
            if keyword in genres_lower or keyword in overview_lower:
                return True
        return False

    def close(self) -> None:
        """Close API clients."""
        self.radarr.close()
        self.sonarr.close()

    def __enter__(self) -> "MediaLibrary":
        return self

    def __exit__(self, *args) -> None:
        self.close()

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/geekxflood/program-director/internal/clients/trakt"
)

var (
	traktQuery      string
	traktLimit      int
	traktShowMovies bool
	traktShowShows  bool
)

// traktCmd represents the trakt command
var traktCmd = &cobra.Command{
	Use:   "trakt",
	Short: "Query Trakt.tv for media information",
	Long: `Query Trakt.tv for media information, trending content, and search.

This command allows you to explore media metadata from Trakt.tv,
including trending movies and shows, popular content, and search results.

Examples:
  # Get trending movies
  program-director trakt trending --movies

  # Get trending shows
  program-director trakt trending --shows

  # Search for a movie or show
  program-director trakt search --query "Inception"

  # Get popular movies
  program-director trakt popular --movies`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := cmd.Help(); err != nil {
			return fmt.Errorf("failed to show help: %w", err)
		}
		return nil
	},
}

// traktTrendingCmd shows trending media
var traktTrendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Show trending movies and shows",
	Long:  `Display currently trending movies and TV shows from Trakt.tv`,
	RunE:  runTraktTrending,
}

// traktPopularCmd shows popular media
var traktPopularCmd = &cobra.Command{
	Use:   "popular",
	Short: "Show popular movies and shows",
	Long:  `Display popular movies and TV shows from Trakt.tv`,
	RunE:  runTraktPopular,
}

// traktSearchCmd searches for media
var traktSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for movies and shows",
	Long:  `Search for movies and TV shows on Trakt.tv by title or keyword`,
	RunE:  runTraktSearch,
}

func init() {
	// Add subcommands
	traktCmd.AddCommand(traktTrendingCmd)
	traktCmd.AddCommand(traktPopularCmd)
	traktCmd.AddCommand(traktSearchCmd)

	// Trending flags
	traktTrendingCmd.Flags().BoolVar(&traktShowMovies, "movies", false, "show trending movies")
	traktTrendingCmd.Flags().BoolVar(&traktShowShows, "shows", false, "show trending shows")
	traktTrendingCmd.Flags().IntVarP(&traktLimit, "limit", "l", 10, "number of results to show")

	// Popular flags
	traktPopularCmd.Flags().BoolVar(&traktShowMovies, "movies", false, "show popular movies")
	traktPopularCmd.Flags().BoolVar(&traktShowShows, "shows", false, "show popular shows")
	traktPopularCmd.Flags().IntVarP(&traktLimit, "limit", "l", 10, "number of results to show")

	// Search flags
	traktSearchCmd.Flags().StringVarP(&traktQuery, "query", "q", "", "search query (required)")
	traktSearchCmd.Flags().IntVarP(&traktLimit, "limit", "l", 10, "number of results to show")
	if err := traktSearchCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Sprintf("failed to mark query flag as required: %v", err))
	}
}

func runTraktTrending(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	if cfg.Trakt.ClientID == "" {
		return errors.New("trakt client ID not configured - set TRAKT_CLIENT_ID environment variable")
	}

	client := trakt.New(&cfg.Trakt)

	// Default to both if neither specified
	if !traktShowMovies && !traktShowShows {
		traktShowMovies = true
		traktShowShows = true
	}

	if traktShowMovies {
		logger.Info("fetching trending movies", "limit", traktLimit)
		movies, err := client.GetTrendingMovies(ctx, traktLimit)
		if err != nil {
			logger.Error("failed to fetch trending movies", "error", err)
			return fmt.Errorf("failed to fetch trending movies: %w", err)
		}

		fmt.Println()
		fmt.Println("Trending Movies")
		fmt.Println("===============")
		for i, tm := range movies {
			fmt.Printf("%2d. %s (%d) - Watchers: %d, Rating: %.1f/10\n",
				i+1, tm.Movie.Title, tm.Movie.Year, tm.Watchers, tm.Movie.Rating)
			if tm.Movie.Overview != "" {
				overview := tm.Movie.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
			fmt.Println()
		}
	}

	if traktShowShows {
		logger.Info("fetching trending shows", "limit", traktLimit)
		shows, err := client.GetTrendingShows(ctx, traktLimit)
		if err != nil {
			logger.Error("failed to fetch trending shows", "error", err)
			return fmt.Errorf("failed to fetch trending shows: %w", err)
		}

		fmt.Println()
		fmt.Println("Trending TV Shows")
		fmt.Println("=================")
		for i, ts := range shows {
			fmt.Printf("%2d. %s (%d) - Watchers: %d, Rating: %.1f/10\n",
				i+1, ts.Show.Title, ts.Show.Year, ts.Watchers, ts.Show.Rating)
			if ts.Show.Overview != "" {
				overview := ts.Show.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
			fmt.Println()
		}
	}

	return nil
}

func runTraktPopular(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	if cfg.Trakt.ClientID == "" {
		return errors.New("trakt client ID not configured - set TRAKT_CLIENT_ID environment variable")
	}

	client := trakt.New(&cfg.Trakt)

	// Default to both if neither specified
	if !traktShowMovies && !traktShowShows {
		traktShowMovies = true
		traktShowShows = true
	}

	if traktShowMovies {
		logger.Info("fetching popular movies", "limit", traktLimit)
		movies, err := client.GetPopularMovies(ctx, traktLimit)
		if err != nil {
			logger.Error("failed to fetch popular movies", "error", err)
			return fmt.Errorf("failed to fetch popular movies: %w", err)
		}

		fmt.Println()
		fmt.Println("Popular Movies")
		fmt.Println("==============")
		for i, movie := range movies {
			fmt.Printf("%2d. %s (%d) - Rating: %.1f/10\n",
				i+1, movie.Title, movie.Year, movie.Rating)
			if movie.Overview != "" {
				overview := movie.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
			fmt.Println()
		}
	}

	if traktShowShows {
		logger.Info("fetching popular shows", "limit", traktLimit)
		shows, err := client.GetPopularShows(ctx, traktLimit)
		if err != nil {
			logger.Error("failed to fetch popular shows", "error", err)
			return fmt.Errorf("failed to fetch popular shows: %w", err)
		}

		fmt.Println()
		fmt.Println("Popular TV Shows")
		fmt.Println("================")
		for i, show := range shows {
			fmt.Printf("%2d. %s (%d) - Rating: %.1f/10\n",
				i+1, show.Title, show.Year, show.Rating)
			if show.Overview != "" {
				overview := show.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
			fmt.Println()
		}
	}

	return nil
}

func runTraktSearch(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	if cfg.Trakt.ClientID == "" {
		return errors.New("trakt client ID not configured - set TRAKT_CLIENT_ID environment variable")
	}

	client := trakt.New(&cfg.Trakt)

	logger.Info("searching Trakt", "query", traktQuery, "limit", traktLimit)

	results, err := client.Search(ctx, traktQuery, traktLimit)
	if err != nil {
		logger.Error("search failed", "error", err)
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("Search Results for '%s'\n", traktQuery)
	fmt.Println("=======================")

	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	for i, result := range results {
		if result.Movie != nil {
			fmt.Printf("%2d. [Movie] %s (%d) - Rating: %.1f/10, Score: %.2f\n",
				i+1, result.Movie.Title, result.Movie.Year, result.Movie.Rating, result.Score)
			if result.Movie.Overview != "" {
				overview := result.Movie.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
		} else if result.Show != nil {
			fmt.Printf("%2d. [Show] %s (%d) - Rating: %.1f/10, Score: %.2f\n",
				i+1, result.Show.Title, result.Show.Year, result.Show.Rating, result.Score)
			if result.Show.Overview != "" {
				overview := result.Show.Overview
				if len(overview) > 100 {
					overview = overview[:100]
				}
				fmt.Printf("    %s\n", overview+"...")
			}
		}
		fmt.Println()
	}

	return nil
}

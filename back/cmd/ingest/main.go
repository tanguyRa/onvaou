package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/ingestion"
)

func main() {
	var source string
	var all bool
	var replayArtifact string
	var listArtifacts bool
	var limit int

	flag.StringVar(&source, "source", "", "ingestion source to run: openagenda|datagouv|vacuum")
	flag.BoolVar(&all, "all", false, "run all configured sources once")
	flag.StringVar(&replayArtifact, "replay-artifact", "", "replay a saved ingestion artifact")
	flag.BoolVar(&listArtifacts, "list-artifacts", false, "list recent ingestion artifacts")
	flag.IntVar(&limit, "limit", 20, "artifact list limit")
	flag.Parse()

	if err := ingestion.ValidateMode(source, all, replayArtifact, listArtifacts); err != nil {
		fail(err)
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.Database.ConnectionString)
	if err != nil {
		fail(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fail(err)
	}

	runner := ingestion.NewRunner(*cfg, logger, pool)

	switch {
	case replayArtifact != "":
		result, err := runner.ReplayArtifact(ctx, replayArtifact)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	case listArtifacts:
		artifacts, err := runner.ListArtifacts(limit)
		if err != nil {
			fail(err)
		}
		printJSON(artifacts)
	case all || source == "":
		results, err := runner.RunAll(ctx)
		if err != nil {
			fail(err)
		}
		printJSON(results)
	default:
		result, err := runner.RunSource(ctx, source)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	}
}

func printJSON(value interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/ingestion/artifacts"
	"github.com/tanguyRa/onvaou/internal/ingestion/geocoder"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
	"github.com/tanguyRa/onvaou/internal/ingestion/replay"
	"github.com/tanguyRa/onvaou/internal/ingestion/sources/datagouv"
	"github.com/tanguyRa/onvaou/internal/ingestion/sources/openagenda"
	"github.com/tanguyRa/onvaou/internal/ingestion/store"
)

type Runner struct {
	logger      *slog.Logger
	openAgenda  *openagenda.Service
	dataGouv    *datagouv.Service
	replay      *replay.Service
	store       *store.Store
	artifactSet *artifacts.Store
}

func NewRunner(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *Runner {
	timeout := time.Duration(cfg.Ingestion.HTTPTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}
	artifactStore := artifacts.New(cfg.Ingestion.DebugDir)
	dbStore := store.New(pool, artifactStore)
	ban := geocoder.NewBANClient(cfg, httpClient)

	openAgendaService := openagenda.NewService(cfg, logger, httpClient, ban, dbStore, artifactStore)
	dataGouvService := datagouv.NewService(cfg, logger, httpClient, ban, dbStore, artifactStore)

	return &Runner{
		logger:      logger,
		openAgenda:  openAgendaService,
		dataGouv:    dataGouvService,
		replay:      replay.NewService(artifactStore, dbStore, openAgendaService, dataGouvService),
		store:       dbStore,
		artifactSet: artifactStore,
	}
}

func (r *Runner) RunSource(ctx context.Context, source string) (model.IngestionResult, error) {
	switch source {
	case "openagenda":
		return r.openAgenda.Run(ctx)
	case "datagouv":
		return r.dataGouv.Run(ctx)
	case "vacuum":
		if err := r.store.VacuumEvents(ctx); err != nil {
			return model.IngestionResult{}, err
		}
		result := model.NewResult("vacuum")
		result.Details = append(result.Details, "vacuum completed")
		return result, nil
	default:
		return model.IngestionResult{}, fmt.Errorf("unsupported source %q", source)
	}
}

func (r *Runner) RunAll(ctx context.Context) ([]model.IngestionResult, error) {
	sources := []string{"openagenda", "datagouv"}
	results := make([]model.IngestionResult, 0, len(sources))
	for _, source := range sources {
		result, err := r.RunSource(ctx, source)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (r *Runner) ReplayArtifact(ctx context.Context, artifactID string) (model.IngestionResult, error) {
	return r.replay.ReplayArtifact(ctx, artifactID)
}

func (r *Runner) ListArtifacts(limit int) ([]model.ReplayArtifact, error) {
	return r.artifactSet.List(limit)
}

func ValidateMode(source string, all bool, replayArtifact string, listArtifacts bool) error {
	modeCount := 0
	if source != "" {
		modeCount++
	}
	if all {
		modeCount++
	}
	if replayArtifact != "" {
		modeCount++
	}
	if listArtifacts {
		modeCount++
	}
	if modeCount == 0 {
		return nil
	}
	if modeCount > 1 {
		return errors.New("choose only one of --source, --all, --replay-artifact, or --list-artifacts")
	}
	return nil
}

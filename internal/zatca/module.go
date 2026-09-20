package zatca

import (
	"context"
	"net/http"
	"strings"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/zatca/application/onboarding"
	"erp/internal/zatca/application/services"
	"erp/internal/zatca/application/submission"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/infrastructure/client"
	"erp/internal/zatca/infrastructure/postgres"
	"erp/internal/zatca/jobs"
	"erp/internal/zatca/ports"
	"erp/internal/zatca/signing"
	zhttp "erp/internal/zatca/transport/http"
	"erp/internal/zatca/validation"
	"erp/internal/zatca/xml"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool              *pgxpool.Pool
	Service           ports.Service
	OnboardingService ports.OnboardingService
	SubmissionService ports.SubmissionService
	Repo              ports.Repository
	SubRepo           ports.SubmissionRepository
	ZATCAClient       ports.ZATCAClient
	OnboardingClient  ports.ZATCAOnboardingClient
	RetryWorker       *jobs.RetryWorker
	Handler           *zhttp.ZATCAHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	subRepo := postgres.NewSubmissionRepository(db)
	xmlGen := xml.NewUBLGenerator()
	validator := validation.NewInvoiceValidator()
	signer := signing.NewInvoiceSigner()

	var zatcaClient ports.ZATCAClient
	var onboardingClient ports.ZATCAOnboardingClient
	environment := domain.EnvironmentSimulation
	if strings.EqualFold(cfg.ZATCA.Environment, "simulation") {
		simClient := client.NewSimulatorClient()
		zatcaClient = simClient
		onboardingClient = simClient
	} else {
		environment = domain.EnvironmentSandbox
		if strings.EqualFold(cfg.ZATCA.Environment, "production") {
			environment = domain.EnvironmentProduction
		}
		httpConfig := client.DefaultConfig(environment)
		if cfg.ZATCA.BaseURL != "" {
			httpConfig.BaseURL = strings.TrimRight(cfg.ZATCA.BaseURL, "/")
		}
		httpConfig.Timeout = cfg.ZATCA.Timeout
		httpClient := client.NewHTTPClient(httpConfig)
		zatcaClient = httpClient
		onboardingClient = httpClient
	}

	allowSimulatedCredentials := strings.EqualFold(cfg.ZATCA.Environment, "simulation")
	svc := services.NewZATCAService(repo, xmlGen, validator, signer, db, allowSimulatedCredentials)
	onboardingSvc := onboarding.NewOnboardingService(repo, onboardingClient, xmlGen, signer, environment)
	submissionSvc := submission.NewSubmissionService(repo, subRepo, zatcaClient, environment)
	retryWorker := jobs.NewRetryWorker(jobs.DefaultWorkerConfig(), subRepo, repo, submissionSvc)

	handler := zhttp.NewZATCAHandler(svc, onboardingSvc, submissionSvc)

	return &Module{
		pool:              db.Pool,
		Service:           svc,
		OnboardingService: onboardingSvc,
		SubmissionService: submissionSvc,
		Repo:              repo,
		SubRepo:           subRepo,
		ZATCAClient:       zatcaClient,
		OnboardingClient:  onboardingClient,
		RetryWorker:       retryWorker,
		Handler:           handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/zatca", m.Handler.Routes(authMiddleware))
}

// SeedDefaultZATCAData initializes baseline EGS units, configuration versions, and compliance CSID credentials.
func (m *Module) SeedDefaultZATCAData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL zatca.seed_default_zatca_data($1)", legalEntityID)
	return err
}

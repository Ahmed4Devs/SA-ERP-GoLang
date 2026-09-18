package iam

import (
	"net/http"

	"erp/internal/iam/application/services"
	"erp/internal/iam/infrastructure/crypto"
	"erp/internal/iam/infrastructure/postgres"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	AuthService *services.AuthService
	MFAService  *services.MFAService
	Handler     *iamhttp.AuthHandler
}

func NewModule(cfg *config.Config, db *database.DB, txManager transaction.Manager) *Module {
	userRepo := postgres.NewUserRepository(db)
	passwordRepo := postgres.NewPasswordRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	attemptRepo := postgres.NewLoginAttemptRepository(db)
	mfaRepo := postgres.NewMFARepository(db)

	hasher := crypto.NewArgon2idHasher()
	tokenGen := crypto.NewSessionTokenGenerator()

	authService := services.NewAuthService(
		userRepo,
		passwordRepo,
		sessionRepo,
		attemptRepo,
		hasher,
		tokenGen,
		txManager,
	)
	var secretBox *crypto.SecretBox
	if cfg.Security.MFAKey != "" {
		secretBox, _ = crypto.NewSecretBox(cfg.Security.MFAKey)
	}
	mfaService := services.NewMFAService(mfaRepo, tokenGen, secretBox, txManager, cfg.AppName)
	authService.SetMFAService(mfaService)

	isProd := cfg.Environment == config.EnvProduction
	handler := iamhttp.NewAuthHandler(authService, mfaService, isProd)

	return &Module{
		AuthService: authService,
		MFAService:  mfaService,
		Handler:     handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Mount("/api/v1/auth", m.Handler.Routes())
}

func (m *Module) AuthMiddleware() func(http.Handler) http.Handler {
	return iamhttp.RequireAuth(m.AuthService)
}

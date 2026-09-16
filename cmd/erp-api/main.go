package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"erp/internal/accounting"
	"erp/internal/audit"
	"erp/internal/authorization"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	platformhttp "erp/internal/platform/http"
	"erp/internal/platform/observability"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	"erp/internal/tax"
	"erp/internal/workflow"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Structured Logger (slog)
	logger := observability.NewLogger(cfg.AppName, string(cfg.Environment), cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("Starting Saudi ERP Platform API",
		"service", cfg.AppName,
		"version", cfg.AppVersion,
		"env", string(cfg.Environment),
		"port", cfg.HTTP.Port,
	)

	// 3. Initialize OpenTelemetry Tracing
	tracerProvider, err := observability.InitTracer(cfg.AppName)
	if err != nil {
		logger.Warn("Failed to initialize OpenTelemetry tracer", "error", err)
	} else {
		defer func() {
			_ = tracerProvider.Shutdown(context.Background())
		}()
	}

	// 4. Setup Root Context with Signal Cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 5. Connect to PostgreSQL Pool (Optional/soft-fail on bootstrap if offline)
	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Warn("Database pool initialization deferred or unavailable", "error", err)
	} else {
		defer db.Close()
		logger.Info("Database pool connected successfully")
	}

	// 6. Setup Production HTTP Server
	server := platformhttp.NewServer(cfg, logger, db)

	// 7. Mount Domain Modules
	if db != nil {
		txManager := transaction.NewManager(db.Pool)

		// A. IAM Module
		iamModule := iam.NewModule(cfg, db, txManager)
		iamModule.RegisterRoutes(server.Router)

		// Seed initial system administrator if missing
		_, err := iamModule.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "مدير النظام الشامل", "Admin@ERP123!")
		if err != nil {
			logger.Debug("Initial admin user already exists or seeding skipped", "reason", err.Error())
		} else {
			logger.Info("Default admin user seeded successfully", "username", "admin", "email", "admin@erp.sa")
		}

		// B. Organization Module
		orgModule := organization.NewModule(cfg, db, txManager)
		orgModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())

		// Seed initial Saudi enterprise organization & legal entity
		seededLE, err := orgModule.SeedDefaultStructure(ctx)
		if err != nil {
			logger.Warn("Failed to seed default organization structure", "error", err)
		} else if seededLE != nil {
			logger.Info("Default Saudi organization hierarchy verified",
				"legal_entity_id", seededLE.ID.String(),
				"legal_entity_code", seededLE.Code,
				"cr_number", seededLE.CRNumber,
			)
		}

		// C. Authorization Module
		authzModule := authorization.NewModule(cfg, db)
		authzModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())

		// Seed initial SUPER_ADMIN assignment for admin user
		adminUser, _ := iamModule.AuthService.GetUserByIdentifier(ctx, "admin")
		if adminUser != nil {
			if err := authzModule.SeedAdminRole(ctx, adminUser.ID); err != nil {
				logger.Debug("Admin role already assigned or skipped", "reason", err.Error())
			} else {
				logger.Info("SUPER_ADMIN role assigned to default administrator", "user_id", adminUser.ID.String())
			}
		}

		// D. Audit Module (Append-Only Cryptographic Audit Trail)
		auditModule := audit.NewModule(cfg, db)
		auditModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())
		logger.Info("Audit module and access trail registered")

		// E. Workflow Module (Versioned Approval Workflow Engine)
		workflowModule := workflow.NewModule(cfg, db)
		workflowModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())
		logger.Info("Workflow module and approval engine registered")

		// F. Accounting Module (General Ledger Foundation)
		accountingModule := accounting.NewModule(cfg, db)
		accountingModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())
		if seededLE != nil {
			if err := accountingModule.SeedDefaultCOA(ctx, seededLE.ID); err != nil {
				logger.Warn("Failed to seed default Chart of Accounts", "error", err)
			} else {
				logger.Info("Default Saudi Chart of Accounts and Fiscal Calendar verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Accounting module and general ledger registered")

		// G. Tax Module (Saudi Effective-Dated VAT Engine & ZATCA Returns)
		taxModule := tax.NewModule(cfg, db)
		taxModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())
		if seededLE != nil {
			if err := taxModule.SeedDefaultTaxData(ctx, seededLE.ID); err != nil {
				logger.Warn("Failed to seed default Tax Structure", "error", err)
			} else {
				logger.Info("Default Saudi VAT configuration and 2026 periods verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Tax module and ZATCA VAT engine registered")

		// H. Products Module (ERP Product Master, UOMs, Variants, SKUs, Barcodes)
		productsModule := products.NewModule(cfg, db)
		productsModule.RegisterRoutes(server.Router, iamModule.AuthMiddleware())
		if seededLE != nil {
			if err := productsModule.SeedDefaultProductsData(ctx, seededLE.ID); err != nil {
				logger.Warn("Failed to seed default Products Data", "error", err)
			} else {
				logger.Info("Default Saudi Categories, UOMs, and Product Master catalog verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Products module and catalog engine registered")
	}

	// 7. Start HTTP Server with Graceful Shutdown
	if err := server.Start(ctx); err != nil {
		logger.Error("HTTP server exited with error", "error", err)
		os.Exit(1)
	}
}

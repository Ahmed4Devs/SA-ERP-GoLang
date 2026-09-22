package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"erp/internal/accounting"
	"erp/internal/attendance"
	"erp/internal/audit"
	"erp/internal/authorization"
	"erp/internal/banking"
	"erp/internal/budgeting"
	"erp/internal/dashboard"
	"erp/internal/documents"
	"erp/internal/fixedassets"
	"erp/internal/gosi"
	"erp/internal/hr"
	"erp/internal/iam"
	"erp/internal/inventory"
	"erp/internal/leave"
	"erp/internal/notifications"
	"erp/internal/organization"
	organizationdomain "erp/internal/organization/domain"
	"erp/internal/payables"
	"erp/internal/payroll"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	platformhttp "erp/internal/platform/http"
	"erp/internal/platform/mail"
	"erp/internal/platform/observability"
	"erp/internal/platform/storage"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	"erp/internal/purchasing"
	"erp/internal/receivables"
	"erp/internal/reporting"
	"erp/internal/sales"
	"erp/internal/tax"
	"erp/internal/treasury"
	"erp/internal/workflow"
	"erp/internal/zatca"
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
	tracerProvider, err := observability.InitTracer(context.Background(), cfg.AppName, cfg.AppVersion, cfg.Telemetry)
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

	// 5. Connect to PostgreSQL. Serving a partial ERP without its database is
	// unsafe, so startup fails closed when the dependency is unavailable.
	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("Database pool initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database pool connected successfully")
	if err := db.Check(); err != nil {
		logger.Error("Database schema readiness check failed", "error", err)
		os.Exit(1)
	}

	// 6. Persistent document storage is a required runtime dependency. Fail
	// closed if its mounted volume cannot be initialized or written.
	blobStore, err := storage.NewLocalFilesystemBlobStore(cfg.Storage.LocalDir)
	if err != nil {
		logger.Error("Failed to initialize local filesystem blob store", "error", err)
		os.Exit(1)
	}
	if err := blobStore.Check(); err != nil {
		logger.Error("Document storage readiness check failed", "error", err)
		os.Exit(1)
	}

	// 7. Setup Production HTTP Server
	server := platformhttp.NewServer(cfg, logger, db, blobStore)

	// 8. Mount Domain Modules
	if db != nil {
		txManager := transaction.NewManager(db.Pool)
		auditModule := audit.NewModule(cfg, db)
		auditRequests := auditModule.AccessMiddleware()

		// A. IAM Module
		iamModule := iam.NewModule(cfg, db, txManager)
		iamModule.RegisterRoutes(server.Router, auditRequests)

		// Bootstrap is explicit in production and uses operator-provided values.
		if cfg.Bootstrap.AdminEnabled {
			_, err := iamModule.AuthService.SeedAdmin(ctx, cfg.Bootstrap.AdminUsername, cfg.Bootstrap.AdminEmail, cfg.Bootstrap.AdminName, cfg.Bootstrap.AdminPassword)
			if err != nil {
				logger.Debug("Initial admin user already exists or seeding skipped", "reason", err.Error())
			} else {
				logger.Info("Bootstrap administrator created", "username", cfg.Bootstrap.AdminUsername, "email", cfg.Bootstrap.AdminEmail)
			}
		}

		// B. Organization Module
		orgModule := organization.NewModule(cfg, db, txManager)

		// Demo business records are opt-in and prohibited by production config.
		var seededLE *organizationdomain.LegalEntity
		if cfg.Bootstrap.DemoDataEnabled {
			seededLE, err = orgModule.SeedDefaultStructure(ctx)
			if err != nil {
				logger.Error("Failed to seed demo organization structure", "error", err)
				os.Exit(1)
			} else if seededLE != nil {
				logger.Info("Demo organization hierarchy verified",
					"legal_entity_id", seededLE.ID.String(),
					"legal_entity_code", seededLE.Code,
					"cr_number", seededLE.CRNumber,
				)
			}
		}

		// C. Authorization Module
		authzModule := authorization.NewModule(cfg, db)
		authMiddleware := iamModule.AuthMiddleware()
		moduleAccess := func(viewPermission, managePermission string) func(http.Handler) http.Handler {
			authorize := authzModule.RequireModuleAccess(viewPermission, managePermission)
			return func(next http.Handler) http.Handler {
				return authMiddleware(authorize(auditRequests(next)))
			}
		}
		orgModule.RegisterRoutes(server.Router, moduleAccess("org.view", "org.manage"))
		authzModule.RegisterRoutes(server.Router, func(next http.Handler) http.Handler {
			return authMiddleware(auditRequests(next))
		})

		// Seed initial SUPER_ADMIN assignment for admin user
		adminUser, _ := iamModule.AuthService.GetUserByIdentifier(ctx, cfg.Bootstrap.AdminUsername)
		if cfg.Bootstrap.AdminEnabled && adminUser != nil {
			if err := authzModule.SeedAdminRole(ctx, adminUser.ID); err != nil {
				logger.Debug("Admin role already assigned or skipped", "reason", err.Error())
			} else {
				logger.Info("SUPER_ADMIN role assigned to default administrator", "user_id", adminUser.ID.String())
			}
		}

		// D. Audit Module (Append-Only Cryptographic Audit Trail)
		auditModule.RegisterRoutes(server.Router, moduleAccess("audit.view", "audit.manage"))
		logger.Info("Audit module and access trail registered")

		// E. Workflow Module (Versioned Approval Workflow Engine)
		workflowModule := workflow.NewModule(cfg, db)
		workflowModule.RegisterRoutes(server.Router, moduleAccess("workflow.view", "workflow.manage"))
		logger.Info("Workflow module and approval engine registered")

		// F. Accounting Module (General Ledger Foundation)
		accountingModule := accounting.NewModule(cfg, db)
		accountingModule.RegisterRoutes(server.Router, moduleAccess("accounting.view", "accounting.manage"))
		if seededLE != nil {
			if err := accountingModule.SeedDefaultCOA(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Chart of Accounts", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Chart of Accounts and Fiscal Calendar verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Accounting module and general ledger registered")

		// G. Tax Module (Saudi Effective-Dated VAT Engine & ZATCA Returns)
		taxModule := tax.NewModule(cfg, db)
		taxModule.RegisterRoutes(server.Router, moduleAccess("tax.view", "tax.manage"))
		if seededLE != nil {
			if err := taxModule.SeedDefaultTaxData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Tax Structure", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi VAT configuration and 2026 periods verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Tax module and ZATCA VAT engine registered")

		// H. Products Module (ERP Product Master, UOMs, Variants, SKUs, Barcodes)
		productsModule := products.NewModule(cfg, db)
		productsModule.RegisterRoutes(server.Router, moduleAccess("products.view", "products.manage"))
		if seededLE != nil {
			if err := productsModule.SeedDefaultProductsData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Products Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Categories, UOMs, and Product Master catalog verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Products module and catalog engine registered")

		// I. Inventory Module (Warehouses, Bins, Movements Ledger, Projected Balances, Transfers)
		inventoryModule := inventory.NewModule(cfg, db)
		inventoryModule.RegisterRoutes(server.Router, moduleAccess("inventory.view", "inventory.manage"))
		if seededLE != nil {
			if err := inventoryModule.SeedDefaultInventoryData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Inventory Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Central Warehouse, Bins, and Initial Stock Balances verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Inventory module and immutable movement ledger registered")

		// J. Receivables Module (Customers, Credit Limits, Payment Terms, Open Items, Receipts)
		receivablesModule := receivables.NewModule(cfg, db)
		receivablesModule.RegisterRoutes(server.Router, moduleAccess("receivables.view", "receivables.manage"))
		if seededLE != nil {
			if err := receivablesModule.SeedDefaultReceivablesData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Receivables Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Payment Terms, Credit Limits, and Customer Master verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Receivables module and accounts receivable engine registered")

		// K. Sales Module (Quotations, Orders, Deliveries, ZATCA Invoices, Credit Notes)
		salesModule := sales.NewModule(cfg, db)
		salesModule.RegisterRoutes(server.Router, moduleAccess("sales.view", "sales.manage"))
		if seededLE != nil {
			if err := salesModule.SeedDefaultSalesData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Sales Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Sales quotations, orders, and ZATCA invoices verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Sales module and order-to-cash workflow registered")

		// L. Payables Module (Suppliers, Payment Terms, Supplier Invoices, Payments, Aging)
		payablesModule := payables.NewModule(cfg, db)
		payablesModule.RegisterRoutes(server.Router, moduleAccess("payables.view", "payables.manage"))
		if seededLE != nil {
			if err := payablesModule.SeedDefaultPayablesData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Payables Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Suppliers, Payment Terms, and AP verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Payables module and supplier management registered")

		// M. Purchasing Module (Purchase Requests, RFQs, Quotes, Purchase Orders, Goods Receipts, 3-Way Match)
		purchasingModule := purchasing.NewModule(cfg, db)
		purchasingModule.RegisterRoutes(server.Router, moduleAccess("purchasing.view", "purchasing.manage"))
		if seededLE != nil {
			if err := purchasingModule.SeedDefaultPurchasingData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Purchasing Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Purchasing PRs, POs, and Receipts verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Purchasing module and procure-to-pay workflow registered")

		// N. Treasury Module (Cash Accounts, Petty Cash, Receipts, Payments, Transfers)
		treasuryModule := treasury.NewModule(cfg, db)
		treasuryModule.RegisterRoutes(server.Router, moduleAccess("treasury.view", "treasury.manage"))
		if seededLE != nil {
			if err := treasuryModule.SeedDefaultTreasuryData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Treasury Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Treasury Cash Accounts and Bank Master verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Treasury module and cash liquidity engine registered")

		// O. Banking Module (Bank Accounts, Statements Import, Reconciliation Engine)
		bankingModule := banking.NewModule(cfg, db)
		bankingModule.RegisterRoutes(server.Router, moduleAccess("banking.view", "banking.manage"))
		logger.Info("Banking module and reconciliation engine registered")

		// P. HR Module (Employee Master, Contracts, Jobs, Positions, Saudization Engine)
		hrModule := hr.NewModule(cfg, db)
		hrModule.RegisterRoutes(server.Router, moduleAccess("hr.view", "hr.manage"))
		if seededLE != nil {
			if err := hrModule.SeedDefaultHRData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default HR Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Jobs, Positions, Employees, and Qiwa Contracts verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("HR module and employee/contract management registered")

		// Q. Attendance Module (Shifts, Rosters, Daily Timesheets, Overtime Engine)
		attendanceModule := attendance.NewModule(cfg, db)
		attendanceModule.RegisterRoutes(server.Router, moduleAccess("attendance.view", "attendance.manage"))
		if seededLE != nil {
			if err := attendanceModule.SeedDefaultAttendanceData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Attendance Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Shifts, Working Hours, and Rosters verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Attendance module and time-tracking registered")

		// R. Leave Module (Statutory Saudi Leaves, Policies, Accruals, Requests Ledger)
		leaveModule := leave.NewModule(cfg, db)
		leaveModule.RegisterRoutes(server.Router, moduleAccess("leave.view", "leave.manage"))
		if seededLE != nil {
			if err := leaveModule.SeedDefaultLeaveData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Leave Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Statutory Leave Types and Accrual Policies verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Leave module and reconcilable balance ledger registered")

		// S. Payroll Module (Deterministic Engine, WPS/SIF File Generator, Payslips, Loans, EOS)
		payrollModule := payroll.NewModule(cfg, db)
		payrollModule.RegisterRoutes(server.Router, moduleAccess("payroll.view", "payroll.manage"))
		if seededLE != nil {
			if err := payrollModule.SeedDefaultPayrollData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Payroll Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Payroll Components, Structures, and Sample Periods verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Payroll module and deterministic calculation engine registered")

		// T. GOSI Module (Versioned Social Insurance Engine, Annuity, OHRSI, SANED, Export Batches)
		gosiModule := gosi.NewModule(cfg, db)
		gosiModule.RegisterRoutes(server.Router, moduleAccess("gosi.view", "gosi.manage"))
		if seededLE != nil {
			if err := gosiModule.SeedDefaultGOSIData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default GOSI Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi GOSI Rules (Annuity, OHRSI, SANED) and Rule Versions verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("GOSI module and versioned social insurance engine registered")

		// U. Fixed Assets Module (Asset Lifecycle, Locations, Assignments, Straight-Line Depreciation, Transfers, Disposals)
		fixedAssetsModule := fixedassets.NewModule(cfg, db)
		fixedAssetsModule.RegisterRoutes(server.Router, moduleAccess("fixedassets.view", "fixedassets.manage"))
		if seededLE != nil {
			if err := fixedAssetsModule.SeedDefaultFixedAssetsData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Fixed Assets Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Saudi Fixed Asset Categories, Locations, Depreciation Books, and Sample Assets verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Fixed assets module and straight-line depreciation engine registered")

		// V. Budgeting Module (Budgets, Versions, Multi-Dimensional Lines, Transfers, Budget vs Actual)
		budgetingModule := budgeting.NewModule(cfg, db)
		budgetingModule.RegisterRoutes(server.Router, moduleAccess("budgeting.view", "budgeting.manage"))
		if seededLE != nil {
			if err := budgetingModule.SeedDefaultBudgetingData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Budgeting Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Annual Operating and CAPEX Budgets verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Budgeting module and multi-dimensional variance engine registered")

		// W. ZATCA Module (EGS Units, CSIDs, UBL 2.1 XML Generator, BR-KSA Validator, ECDSA Signer, TLV QR)
		if cfg.ZATCA.Enabled {
			zatcaModule := zatca.NewModule(cfg, db)
			zatcaModule.RegisterRoutes(server.Router, moduleAccess("zatca.view", "zatca.manage"))
			if seededLE != nil {
				if err := zatcaModule.SeedDefaultZATCAData(ctx, seededLE.ID); err != nil {
					logger.Error("Failed to seed default ZATCA Data", "error", err)
					os.Exit(1)
				} else {
					logger.Info("Default ZATCA EGS Units and configuration verified", "legal_entity_id", seededLE.ID.String())
				}
			}
			logger.Info("ZATCA electronic invoicing integration registered", "environment", cfg.ZATCA.Environment)
		} else {
			logger.Warn("ZATCA integration is disabled; enable only after formal implementation review and operator acknowledgement")
		}

		// X. Documents Module (Secure Blob Storage, Versions & Entity Attachments)
		documentsModule := documents.NewModule(cfg, db, blobStore)
		documentsModule.RegisterRoutes(server.Router, moduleAccess("documents.view", "documents.manage"))
		if seededLE != nil {
			if err := documentsModule.SeedDefaultDocumentsData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Documents Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Company Documents and Legal Certificates verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Documents module and secure blob storage registered")

		// Y. Notifications Module (In-App Alerts, Multi-Channel Dispatch & Mail Platform)
		var mailer mail.Mailer = mail.NewMockMailer()
		if cfg.Mail.Driver == "smtp" {
			mailer = mail.NewSMTPMailer(mail.SMTPConfig{
				Host: cfg.Mail.Host, Port: cfg.Mail.Port, Username: cfg.Mail.Username,
				Password: cfg.Mail.Password, From: cfg.Mail.From,
			})
		}
		notificationsModule := notifications.NewModule(cfg, db, mailer)
		notificationsModule.RegisterRoutes(server.Router, moduleAccess("notifications.view", "notifications.manage"))
		if seededLE != nil {
			if err := notificationsModule.SeedDefaultNotificationsData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Notifications Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Notification Templates and Baseline System Alerts verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Notifications module and platform mailer registered")

		// Z. Reporting Module (Platform Financial, Commercial, Inventory & Statutory BI)
		reportingModule := reporting.NewModule(cfg, db)
		reportingModule.RegisterRoutes(server.Router, moduleAccess("reporting.view", "reporting.manage"))
		if seededLE != nil {
			if err := reportingModule.SeedDefaultReportingData(ctx, seededLE.ID); err != nil {
				logger.Error("Failed to seed default Reporting Data", "error", err)
				os.Exit(1)
			} else {
				logger.Info("Default Reporting Views and Analytic Configurations verified", "legal_entity_id", seededLE.ID.String())
			}
		}
		logger.Info("Reporting module and business intelligence engine registered")

		// AA. Executive Dashboard Module (Cross-Module Aggregation & Command Center)
		dashboardModule := dashboard.NewModule(cfg, db)
		dashboardModule.RegisterRoutes(server.Router, moduleAccess("dashboard.view", "dashboard.manage"))
		logger.Info("Executive Dashboard module and aggregation engine registered")
	}

	// 9. Start HTTP Server with Graceful Shutdown
	if err := server.Start(ctx); err != nil {
		logger.Error("HTTP server exited with error", "error", err)
		os.Exit(1)
	}
}

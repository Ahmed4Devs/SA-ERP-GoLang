package organization

import (
	"context"
	"fmt"
	"net/http"

	"erp/internal/organization/application/services"
	"erp/internal/organization/domain"
	"erp/internal/organization/infrastructure/postgres"
	"erp/internal/organization/ports"
	orghttp "erp/internal/organization/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Module struct {
	OrgService ports.OrganizationService
	SeqService ports.NumberSequenceService
	Handler    *orghttp.OrganizationHandler
}

func NewModule(cfg *config.Config, db *database.DB, txManager transaction.Manager) *Module {
	orgRepo := postgres.NewOrganizationRepository(db)
	legalEntityRepo := postgres.NewLegalEntityRepository(db)
	branchRepo := postgres.NewBranchRepository(db)
	deptRepo := postgres.NewDepartmentRepository(db)
	costCenterRepo := postgres.NewCostCenterRepository(db)
	addrRepo := postgres.NewAddressRepository(db)
	seqRepo := postgres.NewNumberSequenceRepository(db)

	orgService := services.NewOrganizationService(
		orgRepo,
		legalEntityRepo,
		branchRepo,
		deptRepo,
		costCenterRepo,
		addrRepo,
	)

	seqService := services.NewNumberSequenceService(seqRepo, txManager)
	handler := orghttp.NewOrganizationHandler(orgService, seqService)

	return &Module{
		OrgService: orgService,
		SeqService: seqService,
		Handler:    handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/org", m.Handler.Routes(authMiddleware))
}

// SeedDefaultStructure ensures baseline enterprise organization, legal entity, branch, and sequences exist.
func (m *Module) SeedDefaultStructure(ctx context.Context) (*domain.LegalEntity, error) {
	orgs, err := m.OrgService.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing organizations: %w", err)
	}

	var org *domain.Organization
	if len(orgs) == 0 {
		org, err = m.OrgService.CreateOrganization(
			ctx,
			"SAUDI-HOLDING",
			"مجموعة الأفق السعودية القابضة",
			"Saudi Horizon Holding Group",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to seed organization: %w", err)
		}
	} else {
		org = orgs[0]
	}

	entities, err := m.OrgService.ListLegalEntities(ctx, org.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing legal entities: %w", err)
	}

	var legalEntity *domain.LegalEntity
	if len(entities) == 0 {
		legalEntity, err = m.OrgService.CreateLegalEntity(
			ctx,
			org.ID,
			"LE-01",
			"شركة الأفق المتقدمة للتجارة والتقنية",
			"Horizon Advanced Trading & Technology Co.",
			"1010789456", // 10-digit KSA CR
			"SAR",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to seed legal entity: %w", err)
		}

		// Seed primary branch
		branch, err := m.OrgService.CreateBranch(
			ctx,
			legalEntity.ID,
			"BR-RUH",
			"الفرع الرئيسي - الرياض",
			"Riyadh Main Branch",
		)
		if err == nil {
			// Seed National Address for Riyadh branch
			_, _ = m.OrgService.CreateNationalAddress(
				ctx,
				&legalEntity.ID,
				&branch.ID,
				"7534",
				"طريق الملك فهد",
				"King Fahd Road",
				"العليا",
				"Al Olaya",
				"الرياض",
				"Riyadh",
				"12214",
				"2345",
			)
		}

		// Seed departments
		deptManagement, _ := m.OrgService.CreateDepartment(ctx, legalEntity.ID, nil, "MGMT", "الإدارة العامة", "General Management")
		var parentMgmtID *uuid.UUID
		if deptManagement != nil {
			parentMgmtID = &deptManagement.ID
		}
		_, _ = m.OrgService.CreateDepartment(ctx, legalEntity.ID, parentMgmtID, "FIN", "الإدارة المالية والمحاسبة", "Finance & Accounting")
		_, _ = m.OrgService.CreateDepartment(ctx, legalEntity.ID, parentMgmtID, "HR", "الموارد البشرية", "Human Resources")
		_, _ = m.OrgService.CreateDepartment(ctx, legalEntity.ID, parentMgmtID, "IT", "تقنية المعلومات", "Information Technology")

		// Seed Cost Centers
		ccCompany, _ := m.OrgService.CreateCostCenter(ctx, legalEntity.ID, nil, "CC-ALL", "مركز التكلفة العام", "General Cost Center")
		var parentCCID *uuid.UUID
		if ccCompany != nil {
			parentCCID = &ccCompany.ID
		}
		_, _ = m.OrgService.CreateCostCenter(ctx, legalEntity.ID, parentCCID, "CC-ADM", "المصاريف الإدارية", "Administrative Expenses")
		_, _ = m.OrgService.CreateCostCenter(ctx, legalEntity.ID, parentCCID, "CC-OPS", "العمليات والتشغيل", "Operations & Services")

		// Seed Document Number Sequences
		_, _ = m.SeqService.DefineSequence(ctx, legalEntity.ID, "INV-SALES", "INV", "", 1, 1, 6, domain.ResetYearly)
		_, _ = m.SeqService.DefineSequence(ctx, legalEntity.ID, "PO-PURCHASE", "PO", "", 1, 1, 6, domain.ResetYearly)
		_, _ = m.SeqService.DefineSequence(ctx, legalEntity.ID, "JOURNAL-ENTRY", "JV", "", 1, 1, 6, domain.ResetYearly)
		_, _ = m.SeqService.DefineSequence(ctx, legalEntity.ID, "EMPLOYEE", "EMP", "", 1001, 1, 5, domain.ResetNever)
	} else {
		legalEntity = entities[0]
	}

	return legalEntity, nil
}

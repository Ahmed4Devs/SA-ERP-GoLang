package organization_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"erp/internal/organization"
	"erp/internal/organization/domain"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

func setupTestModule(t *testing.T) (*organization.Module, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL not accessible, skipping integration test: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	mod := organization.NewModule(cfg, db, txManager)
	return mod, ctx
}

func TestOrganizationHierarchy_FullLifecycle(t *testing.T) {
	mod, ctx := setupTestModule(t)

	// 1. Create Organization
	orgCode := fmt.Sprintf("ORG-%d", time.Now().UnixNano()%100000)
	org, err := mod.OrgService.CreateOrganization(ctx, orgCode, "شركة التجربة السعودية", "Saudi Testing Co")
	if err != nil {
		t.Fatalf("failed to create organization: %v", err)
	}
	if org.ID == uuid.Nil {
		t.Fatalf("expected valid organization UUID, got nil")
	}

	// 2. Reject Invalid CR Number
	_, err = mod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-TEST", "الكيان التجريبي", "Testing Entity", "12345", "SAR")
	if err == nil {
		t.Fatalf("expected error for invalid 5-digit CR number, got nil")
	}

	// 3. Create Legal Entity with Valid 10-Digit KSA CR
	crNumber := "1010998877"
	le, err := mod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-01", "الكيان التجريبي المعتمد", "Certified Test Legal Entity", crNumber, "SAR")
	if err != nil {
		t.Fatalf("failed to create legal entity: %v", err)
	}
	if le.CRNumber != crNumber {
		t.Errorf("expected CR %s, got %s", crNumber, le.CRNumber)
	}

	// 4. Create Branch
	branch, err := mod.OrgService.CreateBranch(ctx, le.ID, "BR-JED", "فرع جدة", "Jeddah Branch")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if branch.Code != "BR-JED" {
		t.Errorf("expected code BR-JED, got %s", branch.Code)
	}

	// 5. Create Hierarchical Departments
	deptParent, err := mod.OrgService.CreateDepartment(ctx, le.ID, nil, "DEPT-OPS", "إدارة العمليات", "Operations Department")
	if err != nil {
		t.Fatalf("failed to create parent department: %v", err)
	}
	deptChild, err := mod.OrgService.CreateDepartment(ctx, le.ID, &deptParent.ID, "DEPT-LOG", "قسم اللوجستيات", "Logistics Unit")
	if err != nil {
		t.Fatalf("failed to create child department: %v", err)
	}
	if deptChild.ParentID == nil || *deptChild.ParentID != deptParent.ID {
		t.Errorf("expected child parent ID %v, got %v", deptParent.ID, deptChild.ParentID)
	}

	// 6. Create Hierarchical Cost Centers
	ccParent, err := mod.OrgService.CreateCostCenter(ctx, le.ID, nil, "CC-HEAD", "مركز التكلفة الرئيسي", "Main Cost Center")
	if err != nil {
		t.Fatalf("failed to create parent cost center: %v", err)
	}
	ccChild, err := mod.OrgService.CreateCostCenter(ctx, le.ID, &ccParent.ID, "CC-FLEET", "أسطول النقل", "Transport Fleet")
	if err != nil {
		t.Fatalf("failed to create child cost center: %v", err)
	}
	if ccChild.ParentID == nil || *ccChild.ParentID != ccParent.ID {
		t.Errorf("expected child cost center parent ID %v, got %v", ccParent.ID, ccChild.ParentID)
	}

	// 7. Create KSA National Address
	addr, err := mod.OrgService.CreateNationalAddress(
		ctx,
		&le.ID,
		&branch.ID,
		"8241",
		"طريق الملك عبد العزيز",
		"King Abdulaziz Road",
		"الزهراء",
		"Al Zahra",
		"جدة",
		"Jeddah",
		"23522",
		"4123",
	)
	if err != nil {
		t.Fatalf("failed to create KSA national address: %v", err)
	}
	if addr.BuildingNumber != "8241" || addr.PostalCode != "23522" {
		t.Errorf("unexpected address numbers: %s, %s", addr.BuildingNumber, addr.PostalCode)
	}
}

func TestNumberSequence_ConcurrencyPessimisticLock(t *testing.T) {
	mod, ctx := setupTestModule(t)

	// Setup an Organization and Legal Entity
	orgCode := fmt.Sprintf("ORG-SEQ-%d", time.Now().UnixNano()%100000)
	org, err := mod.OrgService.CreateOrganization(ctx, orgCode, "مؤسسة الترقيم", "Sequence Org")
	if err != nil {
		t.Fatalf("failed to create organization: %v", err)
	}
	le, err := mod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-SEQ", "كيان الترقيم", "Sequence Entity", "1010443322", "SAR")
	if err != nil {
		t.Fatalf("failed to create legal entity: %v", err)
	}

	// Define Sequence: INV-SALES, Prefix: INV, Step: 1, Padding: 6, ResetYearly
	seqCode := "INV-SALES"
	_, err = mod.SeqService.DefineSequence(ctx, le.ID, seqCode, "INV", "", 1, 1, 6, domain.ResetYearly)
	if err != nil {
		t.Fatalf("failed to define number sequence: %v", err)
	}

	// Concurrency test: 20 concurrent goroutines requesting sequence generation
	concurrency := 20
	results := make([]string, concurrency)
	errorsList := make([]error, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			num, genErr := mod.SeqService.GenerateNext(ctx, le.ID, seqCode)
			results[idx] = num
			errorsList[idx] = genErr
		}(i)
	}

	wg.Wait()

	// Verify all succeeded
	for i, genErr := range errorsList {
		if genErr != nil {
			t.Fatalf("goroutine %d failed to generate sequence: %v", i, genErr)
		}
	}

	// Verify all generated numbers are completely unique (no duplicates!)
	seen := make(map[string]bool)
	currentYear := time.Now().Year()
	for _, num := range results {
		if seen[num] {
			t.Fatalf("duplicate sequence number detected under concurrent load: %s", num)
		}
		seen[num] = true

		// Check format e.g. "INV-2026-000001"
		expectedPrefix := fmt.Sprintf("INV-%d-", currentYear)
		if len(num) != len(expectedPrefix)+6 {
			t.Errorf("unexpected formatted sequence length for %s", num)
		}
	}

	if len(seen) != concurrency {
		t.Fatalf("expected %d unique sequence numbers, got %d", concurrency, len(seen))
	}
}

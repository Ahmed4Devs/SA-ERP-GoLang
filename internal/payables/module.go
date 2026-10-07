package payables

import (
	"context"
	"erp/internal/accounting"
	postingapp "erp/internal/accounting/application/posting"
	"erp/internal/audit"
	"net/http"

	"erp/internal/payables/application/services"
	"erp/internal/payables/infrastructure/postgres"
	"erp/internal/payables/ports"
	payableshttp "erp/internal/payables/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool            *pgxpool.Pool
	SupplierService ports.SupplierService
	PayableService  ports.PayableService
	Handler         *payableshttp.PayableHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	supplierRepo := postgres.NewSupplierRepository(db)
	payableRepo := postgres.NewPayableRepository(db)

	supplierSvc := services.NewSupplierService(supplierRepo)
	payableSvc := services.NewPayableService(supplierRepo, payableRepo, db, postingapp.NewSourceService(db, accounting.NewModule(cfg, db).PostingEngine, audit.NewModule(cfg, db).Service))

	handler := payableshttp.NewPayableHandler(supplierSvc, payableSvc)

	return &Module{
		pool:            db.Pool,
		SupplierService: supplierSvc,
		PayableService:  payableSvc,
		Handler:         handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/payables", m.Handler.Routes(authMiddleware))
}

const ensureSeedProcedureSQL = `
CREATE OR REPLACE FUNCTION payables.seed_default_payables_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_net30_id UUID;
    v_net15_id UUID;
    v_immediate_id UUID;
    v_ap_account_id UUID;
    v_exp_account_id UUID;
BEGIN
    -- 1. Payment Terms
    INSERT INTO payables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES 
        (p_legal_entity_id, 'NET30', 'Net 30 Days', 'صافي 30 يوماً', 30, 0, 0.00),
        (p_legal_entity_id, 'NET15', 'Net 15 Days', 'صافي 15 يوماً', 15, 0, 0.00),
        (p_legal_entity_id, 'IMMEDIATE', 'Cash on Delivery', 'سداد فوري عند الاستلام', 0, 0, 0.00)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    SELECT id INTO v_net30_id FROM payables.payment_terms WHERE legal_entity_id = p_legal_entity_id AND code = 'NET30';
    SELECT id INTO v_immediate_id FROM payables.payment_terms WHERE legal_entity_id = p_legal_entity_id AND code = 'IMMEDIATE';

	-- Find AP GL account (211000) and a posting-level operating expense (523000)
    SELECT id INTO v_ap_account_id FROM accounting.accounts 
    WHERE legal_entity_id = p_legal_entity_id AND code = '211000' LIMIT 1;

    SELECT id INTO v_exp_account_id FROM accounting.accounts 
	WHERE legal_entity_id = p_legal_entity_id AND code = '523000' AND is_active AND allow_posting LIMIT 1;

    -- 2. Default Saudi Suppliers
    IF v_ap_account_id IS NOT NULL THEN
        INSERT INTO payables.suppliers (
            legal_entity_id, supplier_number, supplier_type, name_en, name_ar,
            tax_id, cr_number, iban, bank_name,
            building_no, street_name, district, city, postal_code, additional_no, country_code,
            email, phone, contact_person, payment_terms_id, currency_code,
            default_liability_account_id, default_expense_account_id, status
        )
        VALUES 
            (
                p_legal_entity_id, 'SUP-0001', 'LOCAL',
                'Al-Fozan Building Materials Co.', 'شركة الفوزان لمواد البناء',
                '310123456789013', '1010345678', 'SA0380000000123456789012', 'Al Rajhi Bank',
                '7890', 'King Fahd Road', 'Al Olaya', 'Riyadh', '12214', '3456', 'SA',
                'procurement@alfozan.example.com', '+966114567890', 'Eng. Mohammed Al-Fozan',
                v_net30_id, 'SAR', v_ap_account_id, v_exp_account_id, 'ACTIVE'
            ),
            (
                p_legal_entity_id, 'SUP-0002', 'LOCAL',
                'Saudi Electricity Company (SEC)', 'الشركة السعودية للكهرباء',
                '300012345678903', '1010000012', 'SA4410000000987654321098', 'National Commercial Bank (SNB)',
                '1234', 'Olaya Main Street', 'Al Malaz', 'Riyadh', '11411', '5678', 'SA',
                'billing@se.com.sa', '+966118001234', 'Billing Services',
                v_immediate_id, 'SAR', v_ap_account_id, v_exp_account_id, 'ACTIVE'
            )
        ON CONFLICT (legal_entity_id, supplier_number) DO NOTHING;
    END IF;
END;
$$ LANGUAGE plpgsql;
`

// SeedDefaultPayablesData initializes default payment terms and sample suppliers for a legal entity.
func (m *Module) SeedDefaultPayablesData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, _ = m.pool.Exec(ctx, ensureSeedProcedureSQL)
	_, err := m.pool.Exec(ctx, "SELECT payables.seed_default_payables_data($1)", legalEntityID)
	return err
}

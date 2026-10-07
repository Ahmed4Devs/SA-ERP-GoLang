package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/payables/domain"
	taxdomain "erp/internal/tax/domain"
	taxpostgres "erp/internal/tax/infrastructure/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *payableService) resolveInputTax(ctx context.Context, entity uuid.UUID, selected *uuid.UUID, date time.Time, legacyRate decimal.Decimal) (uuid.UUID, decimal.Decimal, error) {
	var id uuid.UUID
	var rate decimal.Decimal
	// The legacy rate field selects a configured standard/out-of-scope code;
	// an explicit code always uses its effective-dated server rate.
	err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT c.id,r.rate_percent
 FROM tax.tax_codes c JOIN LATERAL (
 SELECT rate_percent FROM tax.tax_rate_versions r WHERE r.tax_code_id=c.id
 AND r.effective_from <= $2 AND (r.effective_to IS NULL OR r.effective_to >= $2)
 ORDER BY effective_from DESC LIMIT 1) r ON true
 WHERE c.legal_entity_id=$1 AND c.is_active AND c.tax_type IN ('INPUT','BOTH')
 AND (($3::uuid IS NOT NULL AND c.id=$3) OR ($3::uuid IS NULL AND r.rate_percent=$4 AND c.tax_category_code=CASE WHEN $4=0 THEN 'O' ELSE 'S' END))
 ORDER BY c.code LIMIT 1 FOR SHARE OF c`, entity, date, selected, legacyRate).Scan(&id, &rate)
	if err != nil {
		return id, rate, fmt.Errorf("an active input tax code and rate must be configured for the invoice date: %w", err)
	}
	return id, rate, nil
}

func (s *payableService) postInputTax(ctx context.Context, inv *domain.SupplierInvoice, journalID, actor uuid.UUID) error {
	repo := taxpostgres.NewTaxRepository(s.db.Pool)
	period, err := repo.GetOpenPeriodForDate(ctx, inv.LegalEntityID, inv.InvoiceDate)
	if err != nil {
		return err
	}
	if period == nil {
		return fmt.Errorf("an open input tax period is required")
	}
	var status string
	if err = s.db.DBTX(ctx).QueryRow(ctx, "SELECT status FROM tax.tax_periods WHERE id=$1 FOR SHARE", period.ID).Scan(&status); err != nil {
		return err
	}
	if status != "OPEN" {
		return fmt.Errorf("input tax period is closed")
	}
	record := &taxdomain.TaxTransaction{LegalEntityID: inv.LegalEntityID, TaxPeriodID: &period.ID, SourceModule: "PAYABLES", SourceType: "SUPPLIER_INVOICE", SourceID: inv.ID, TransactionNumber: "TAX-" + inv.InvoiceNumber, TransactionDate: inv.InvoiceDate, TransactionType: "INPUT", PartnerID: &inv.SupplierID, PartnerName: &inv.SupplierNameEn, PartnerTaxNumber: inv.SupplierTaxID, JournalEntryID: &journalID, CreatedBy: &actor, Status: "POSTED"}
	var taxable, tax, previousTaxable, previousTax decimal.Decimal
	baseTaxableTotal := inv.TaxableAmount.Mul(inv.ExchangeRate).Round(4)
	for i, line := range inv.Lines {
		if line.TaxCodeID == nil {
			return fmt.Errorf("legacy draft has no tax-code snapshot; recreate it with a configured input tax code")
		}
		code, err := repo.GetTaxCodeByID(ctx, *line.TaxCodeID)
		if err != nil {
			return err
		}
		if code == nil || code.LegalEntityID != inv.LegalEntityID || (code.TaxType != "INPUT" && code.TaxType != "BOTH") {
			return fmt.Errorf("input tax snapshot scope mismatch")
		}
		taxable = taxable.Add(line.Subtotal)
		tax = tax.Add(line.TaxAmount)
		nextTaxable := taxable.Mul(inv.ExchangeRate).Round(4)
		// VAT is the final debit after expenses; use the same cumulative FX
		// allocation as the posting engine so both ledgers agree exactly.
		nextTax := inv.TaxableAmount.Add(tax).Mul(inv.ExchangeRate).Round(4).Sub(baseTaxableTotal)
		item := taxdomain.TaxTransactionLine{LineNumber: i + 1, TaxCodeID: code.ID, TaxCategoryCode: code.TaxCategoryCode, TaxRatePercent: line.TaxRate, TaxableAmount: nextTaxable.Sub(previousTaxable), TaxAmount: nextTax.Sub(previousTax), IsRecoverable: code.IsRecoverable}
		item.GrossAmount = item.TaxableAmount.Add(item.TaxAmount)
		record.TotalTaxableAmount = record.TotalTaxableAmount.Add(item.TaxableAmount)
		record.TotalTaxAmount = record.TotalTaxAmount.Add(item.TaxAmount)
		record.TotalGrossAmount = record.TotalGrossAmount.Add(item.GrossAmount)
		record.Lines = append(record.Lines, item)
		previousTaxable, previousTax = nextTaxable, nextTax
	}
	_, err = repo.CreateTaxTransaction(ctx, record)
	return err
}

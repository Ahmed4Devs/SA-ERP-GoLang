package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp/internal/sales/domain"
	taxdomain "erp/internal/tax/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func validateSalesLine(quantity, price, discount decimal.Decimal) error {
	if !quantity.IsPositive() || price.IsNegative() || discount.IsNegative() || discount.GreaterThan(decimal.NewFromInt(100)) {
		return errors.New("sales line requires positive quantity, nonnegative price and a discount between zero and 100")
	}
	return nil
}

func (s *salesService) resolveSalesTax(ctx context.Context, entityID uuid.UUID, code string, date time.Time) (string, decimal.Decimal, error) {
	if date.IsZero() {
		return "", decimal.Zero, errors.New("sales document date is required")
	}
	taxCode, err := s.tax.GetTaxCodeByCode(ctx, entityID, code)
	if err != nil {
		return "", decimal.Zero, err
	}
	if taxCode == nil || !taxCode.IsActive || taxCode.TaxType == "INPUT" {
		return "", decimal.Zero, errors.New("active output tax code is required")
	}
	rate, err := s.tax.GetActiveRateVersion(ctx, taxCode.ID, date)
	if err != nil {
		return "", decimal.Zero, err
	}
	if rate == nil {
		return "", decimal.Zero, errors.New("no tax rate is configured for the document date")
	}
	return taxCode.Code, rate.RatePercent, nil
}

type salesTaxSnapshot struct {
	taxableAmount decimal.Decimal
	taxAmount     decimal.Decimal
	lines         []taxdomain.TaxTransactionLine
}

func invoiceTaxSnapshot(inv *domain.SalesInvoice) salesTaxSnapshot {
	lines := make([]taxdomain.TaxTransactionLine, 0, len(inv.Lines))
	for _, line := range inv.Lines {
		lines = append(lines, taxdomain.TaxTransactionLine{
			LineNumber: line.LineNumber, TaxCode: line.TaxCode, TaxRatePercent: line.TaxRatePercent,
			TaxableAmount: line.TaxableAmount,
			TaxAmount:     line.TaxAmount,
		})
	}
	return convertSalesTaxSnapshot(lines, inv.ExchangeRate)
}

func creditNoteTaxSnapshot(cn *domain.CreditNote) salesTaxSnapshot {
	lines := make([]taxdomain.TaxTransactionLine, 0, len(cn.Lines))
	for _, line := range cn.Lines {
		lines = append(lines, taxdomain.TaxTransactionLine{
			LineNumber: line.LineNumber, TaxCode: line.TaxCode, TaxRatePercent: line.TaxRatePercent,
			TaxableAmount: line.TaxableAmount,
			TaxAmount:     line.TaxAmount,
		})
	}
	return convertSalesTaxSnapshot(lines, cn.ExchangeRate)
}

// Allocate rounding through cumulative converted totals. The difference between
// consecutive rounded totals belongs to that line, so line sums exactly equal
// the converted document totals used by GL. A zero-tax line receives no tax.
func convertSalesTaxSnapshot(lines []taxdomain.TaxTransactionLine, rate decimal.Decimal) salesTaxSnapshot {
	result := salesTaxSnapshot{lines: lines}
	taxable, tax := decimal.Zero, decimal.Zero
	for i := range result.lines {
		line := &result.lines[i]
		taxable = taxable.Add(line.TaxableAmount)
		tax = tax.Add(line.TaxAmount)
		baseTaxable, baseTax := taxable.Mul(rate).Round(6), tax.Mul(rate).Round(6)
		line.TaxableAmount = baseTaxable.Sub(result.taxableAmount)
		line.TaxAmount = baseTax.Sub(result.taxAmount)
		result.taxableAmount, result.taxAmount = baseTaxable, baseTax
	}
	return result
}

func (s *salesService) postInvoiceTax(ctx context.Context, inv *domain.SalesInvoice, journalID, actorID uuid.UUID, snapshot salesTaxSnapshot) error {
	return s.postSalesTax(ctx, inv.LegalEntityID, inv.ID, inv.CustomerID, journalID,
		"INVOICE", inv.InvoiceNumber, inv.IssueDate, &actorID, snapshot.lines)
}

func (s *salesService) postCreditNoteTax(ctx context.Context, cn *domain.CreditNote, journalID, actorID uuid.UUID, snapshot salesTaxSnapshot) error {
	lines := append([]taxdomain.TaxTransactionLine(nil), snapshot.lines...)
	for i := range lines {
		lines[i].TaxableAmount = lines[i].TaxableAmount.Neg()
		lines[i].TaxAmount = lines[i].TaxAmount.Neg()
	}
	return s.postSalesTax(ctx, cn.LegalEntityID, cn.ID, cn.CustomerID, journalID,
		"CREDIT_NOTE", cn.CreditNoteNumber, cn.IssueDate, &actorID, lines)
}

// The tax ledger copies the invoice's immutable calculation snapshot. Resolving a
// new effective rate here would change historical invoices after configuration changes.
func (s *salesService) postSalesTax(ctx context.Context, entityID, sourceID, customerID, journalID uuid.UUID,
	sourceType, number string, date time.Time, actorID *uuid.UUID, lines []taxdomain.TaxTransactionLine) error {
	period, err := s.tax.GetOpenPeriodForDate(ctx, entityID, date)
	if err != nil {
		return err
	}
	if period == nil || period.Status != taxdomain.TaxPeriodStatusOpen {
		return errors.New("an open tax period is required for the sales document date")
	}
	// Prevent a concurrent VAT close while creating its final transaction.
	var status string
	if err := s.db.DBTX(ctx).QueryRow(ctx, "SELECT status FROM tax.tax_periods WHERE id = $1 FOR SHARE", period.ID).Scan(&status); err != nil {
		return err
	}
	if status != string(taxdomain.TaxPeriodStatusOpen) {
		return errors.New("tax period is not open")
	}
	record := &taxdomain.TaxTransaction{
		LegalEntityID: entityID, TaxPeriodID: &period.ID, SourceModule: "SALES", SourceType: sourceType,
		SourceID: sourceID, TransactionNumber: "TAX-" + number, TransactionDate: date,
		TransactionType: "OUTPUT", PartnerID: &customerID, JournalEntryID: &journalID,
		Status: "POSTED", CreatedBy: actorID,
	}
	for _, line := range lines {
		code, err := s.tax.GetTaxCodeByCode(ctx, entityID, line.TaxCode)
		if err != nil {
			return err
		}
		if code == nil {
			return fmt.Errorf("sales snapshot tax code %q was not found", line.TaxCode)
		}
		line.TaxCodeID, line.TaxCategoryCode = code.ID, code.TaxCategoryCode
		line.GrossAmount = line.TaxableAmount.Add(line.TaxAmount)
		line.IsRecoverable = false
		record.TotalTaxableAmount = record.TotalTaxableAmount.Add(line.TaxableAmount)
		record.TotalTaxAmount = record.TotalTaxAmount.Add(line.TaxAmount)
		record.TotalGrossAmount = record.TotalGrossAmount.Add(line.GrossAmount)
		record.Lines = append(record.Lines, line)
	}
	_, err = s.tax.CreateTaxTransaction(ctx, record)
	return err
}

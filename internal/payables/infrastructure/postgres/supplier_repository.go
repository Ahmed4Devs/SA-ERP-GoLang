package postgres

import (
	"context"
	"errors"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SupplierRepository struct {
	db *database.DB
}

func NewSupplierRepository(db *database.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

func (r *SupplierRepository) CreateSupplier(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error) {
	q := r.queries(ctx)

	var pTermID, expAccID, liabAccID pgtype.UUID
	if s.PaymentTermsID != nil {
		pTermID = toPgUUID(*s.PaymentTermsID)
	}
	if s.DefaultExpenseAccountID != nil {
		expAccID = toPgUUID(*s.DefaultExpenseAccountID)
	}
	if s.DefaultLiabilityAccountID != nil {
		liabAccID = toPgUUID(*s.DefaultLiabilityAccountID)
	}

	country := s.Address.CountryCode
	if country == "" {
		country = "SA"
	}
	curr := s.CurrencyCode
	if curr == "" {
		curr = "SAR"
	}

	row, err := q.CreateSupplier(ctx, sqlc.CreateSupplierParams{
		LegalEntityID:             toPgUUID(s.LegalEntityID),
		SupplierNumber:            s.SupplierNumber,
		SupplierType:              string(s.SupplierType),
		NameEn:                    s.NameEn,
		NameAr:                    s.NameAr,
		TaxID:                     toPgTextPtr(s.TaxID),
		CrNumber:                  toPgTextPtr(s.CRNumber),
		NationalID:                toPgTextPtr(s.NationalID),
		Iban:                      toPgTextPtr(s.IBAN),
		BankName:                  toPgTextPtr(s.BankName),
		SwiftCode:                 toPgTextPtr(s.SwiftCode),
		BuildingNo:                toPgText(s.Address.BuildingNo),
		StreetName:                toPgText(s.Address.StreetName),
		District:                  toPgText(s.Address.District),
		City:                      s.Address.City,
		PostalCode:                toPgText(s.Address.PostalCode),
		AdditionalNo:              toPgText(s.Address.AdditionalNo),
		CountryCode:               country,
		Email:                     toPgTextPtr(s.Email),
		Phone:                     toPgTextPtr(s.Phone),
		ContactPerson:             toPgTextPtr(s.ContactPerson),
		Website:                   toPgTextPtr(s.Website),
		PaymentTermsID:            pTermID,
		CurrencyCode:              curr,
		DefaultExpenseAccountID:   expAccID,
		DefaultLiabilityAccountID: liabAccID,
		Status:                    string(s.Status),
	})
	if err != nil {
		return nil, err
	}

	return r.GetSupplierByID(ctx, toUUID(row.ID))
}

func (r *SupplierRepository) GetSupplierByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}

	s := &domain.Supplier{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		SupplierNumber: row.SupplierNumber,
		SupplierType:   domain.SupplierType(row.SupplierType),
		NameEn:         row.NameEn,
		NameAr:         row.NameAr,
		TaxID:          fromPgTextPtr(row.TaxID),
		CRNumber:       fromPgTextPtr(row.CrNumber),
		NationalID:     fromPgTextPtr(row.NationalID),
		IBAN:           fromPgTextPtr(row.Iban),
		BankName:       fromPgTextPtr(row.BankName),
		SwiftCode:      fromPgTextPtr(row.SwiftCode),
		Address: domain.SaudiAddress{
			BuildingNo:   fromPgText(row.BuildingNo),
			StreetName:   fromPgText(row.StreetName),
			District:     fromPgText(row.District),
			City:         row.City,
			PostalCode:   fromPgText(row.PostalCode),
			AdditionalNo: fromPgText(row.AdditionalNo),
			CountryCode:  row.CountryCode,
		},
		Email:                       fromPgTextPtr(row.Email),
		Phone:                       fromPgTextPtr(row.Phone),
		ContactPerson:               fromPgTextPtr(row.ContactPerson),
		Website:                     fromPgTextPtr(row.Website),
		PaymentTermsID:              toUUIDPtr(row.PaymentTermsID),
		PaymentTermsCode:            fromPgText(row.PaymentTermsCode),
		PaymentTermsNameAr:          fromPgText(row.PaymentTermsNameAr),
		PaymentTermsDueDays:         row.PaymentTermsDueDays.Int32,
		CurrencyCode:                row.CurrencyCode,
		DefaultExpenseAccountID:     toUUIDPtr(row.DefaultExpenseAccountID),
		DefaultLiabilityAccountID:   toUUIDPtr(row.DefaultLiabilityAccountID),
		DefaultExpenseAccountCode:   fromPgText(row.DefaultExpenseAccountNumber),
		DefaultLiabilityAccountCode: fromPgText(row.DefaultLiabilityAccountNumber),
		Status:                      domain.SupplierStatus(row.Status),
		CreatedAt:                   fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                   fromPgTimestamptz(row.UpdatedAt),
	}
	return s, nil
}

func (r *SupplierRepository) GetSupplierByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Supplier, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierByNumber(ctx, sqlc.GetSupplierByNumberParams{
		LegalEntityID:  toPgUUID(legalEntityID),
		SupplierNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return r.GetSupplierByID(ctx, toUUID(row.ID))
}

func (r *SupplierRepository) ListSuppliers(ctx context.Context, filter ports.SupplierFilter) ([]*domain.Supplier, int64, error) {
	q := r.queries(ctx)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	sType := filter.SupplierType
	if sType == "" {
		sType = "ALL"
	}
	status := filter.Status
	if status == "" {
		status = "ALL"
	}

	rows, err := q.ListSuppliers(ctx, sqlc.ListSuppliersParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       sType,
		Column4:       status,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := q.CountSuppliers(ctx, sqlc.CountSuppliersParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       sType,
		Column4:       status,
	})
	if err != nil {
		return nil, 0, err
	}

	var suppliers []*domain.Supplier
	for _, row := range rows {
		suppliers = append(suppliers, &domain.Supplier{
			ID:             toUUID(row.ID),
			LegalEntityID:  toUUID(row.LegalEntityID),
			SupplierNumber: row.SupplierNumber,
			SupplierType:   domain.SupplierType(row.SupplierType),
			NameEn:         row.NameEn,
			NameAr:         row.NameAr,
			TaxID:          fromPgTextPtr(row.TaxID),
			CRNumber:       fromPgTextPtr(row.CrNumber),
			NationalID:     fromPgTextPtr(row.NationalID),
			IBAN:           fromPgTextPtr(row.Iban),
			BankName:       fromPgTextPtr(row.BankName),
			SwiftCode:      fromPgTextPtr(row.SwiftCode),
			Address: domain.SaudiAddress{
				BuildingNo:   fromPgText(row.BuildingNo),
				StreetName:   fromPgText(row.StreetName),
				District:     fromPgText(row.District),
				City:         row.City,
				PostalCode:   fromPgText(row.PostalCode),
				AdditionalNo: fromPgText(row.AdditionalNo),
				CountryCode:  row.CountryCode,
			},
			Email:                       fromPgTextPtr(row.Email),
			Phone:                       fromPgTextPtr(row.Phone),
			ContactPerson:               fromPgTextPtr(row.ContactPerson),
			Website:                     fromPgTextPtr(row.Website),
			PaymentTermsID:              toUUIDPtr(row.PaymentTermsID),
			PaymentTermsCode:            fromPgText(row.PaymentTermsCode),
			PaymentTermsNameAr:          fromPgText(row.PaymentTermsNameAr),
			PaymentTermsDueDays:         row.PaymentTermsDueDays.Int32,
			CurrencyCode:                row.CurrencyCode,
			DefaultExpenseAccountID:     toUUIDPtr(row.DefaultExpenseAccountID),
			DefaultLiabilityAccountID:   toUUIDPtr(row.DefaultLiabilityAccountID),
			DefaultExpenseAccountCode:   fromPgText(row.DefaultExpenseAccountNumber),
			DefaultLiabilityAccountCode: fromPgText(row.DefaultLiabilityAccountNumber),
			Status:                      domain.SupplierStatus(row.Status),
			CreatedAt:                   fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                   fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return suppliers, total, nil
}

func (r *SupplierRepository) UpdateSupplier(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error) {
	q := r.queries(ctx)

	country := s.Address.CountryCode
	if country == "" {
		country = "SA"
	}

	_, err := q.UpdateSupplier(ctx, sqlc.UpdateSupplierParams{
		ID:                        toPgUUID(s.ID),
		LegalEntityID:             toPgUUID(s.LegalEntityID),
		NameEn:                    s.NameEn,
		NameAr:                    s.NameAr,
		TaxID:                     toPgTextPtr(s.TaxID),
		CrNumber:                  toPgTextPtr(s.CRNumber),
		NationalID:                toPgTextPtr(s.NationalID),
		Iban:                      toPgTextPtr(s.IBAN),
		BankName:                  toPgTextPtr(s.BankName),
		SwiftCode:                 toPgTextPtr(s.SwiftCode),
		BuildingNo:                toPgText(s.Address.BuildingNo),
		StreetName:                toPgText(s.Address.StreetName),
		District:                  toPgText(s.Address.District),
		City:                      s.Address.City,
		PostalCode:                toPgText(s.Address.PostalCode),
		AdditionalNo:              toPgText(s.Address.AdditionalNo),
		CountryCode:               country,
		Email:                     toPgTextPtr(s.Email),
		Phone:                     toPgTextPtr(s.Phone),
		ContactPerson:             toPgTextPtr(s.ContactPerson),
		Website:                   toPgTextPtr(s.Website),
		PaymentTermsID:            toPgUUIDPtr(s.PaymentTermsID),
		DefaultExpenseAccountID:   toPgUUIDPtr(s.DefaultExpenseAccountID),
		DefaultLiabilityAccountID: toPgUUIDPtr(s.DefaultLiabilityAccountID),
	})
	if err != nil {
		return nil, err
	}
	return r.GetSupplierByID(ctx, s.ID)
}

func (r *SupplierRepository) SetSupplierStatus(ctx context.Context, id uuid.UUID, status domain.SupplierStatus) error {
	q := r.queries(ctx)
	return q.SetSupplierStatus(ctx, sqlc.SetSupplierStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
}

func (r *SupplierRepository) CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error) {
	q := r.queries(ctx)
	row, err := q.CreatePayablesPaymentTerm(ctx, sqlc.CreatePayablesPaymentTermParams{
		LegalEntityID:   toPgUUID(term.LegalEntityID),
		Code:            term.Code,
		NameEn:          term.NameEn,
		NameAr:          term.NameAr,
		DueDays:         term.DueDays,
		DiscountDays:    term.DiscountDays,
		DiscountPercent: toPgNumeric(term.DiscountPercent),
		IsActive:        term.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &domain.PaymentTerm{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		Code:            row.Code,
		NameEn:          row.NameEn,
		NameAr:          row.NameAr,
		DueDays:         row.DueDays,
		DiscountDays:    row.DiscountDays,
		DiscountPercent: fromPgNumeric(row.DiscountPercent),
		IsActive:        row.IsActive,
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *SupplierRepository) GetPaymentTermByID(ctx context.Context, id uuid.UUID) (*domain.PaymentTerm, error) {
	q := r.queries(ctx)
	row, err := q.GetPayablesPaymentTermByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentTermNotFound
		}
		return nil, err
	}
	return &domain.PaymentTerm{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		Code:            row.Code,
		NameEn:          row.NameEn,
		NameAr:          row.NameAr,
		DueDays:         row.DueDays,
		DiscountDays:    row.DiscountDays,
		DiscountPercent: fromPgNumeric(row.DiscountPercent),
		IsActive:        row.IsActive,
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *SupplierRepository) ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID, includeInactive bool) ([]*domain.PaymentTerm, error) {
	q := r.queries(ctx)
	rows, err := q.ListPayablesPaymentTerms(ctx, sqlc.ListPayablesPaymentTermsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       includeInactive,
	})
	if err != nil {
		return nil, err
	}
	var terms []*domain.PaymentTerm
	for _, row := range rows {
		terms = append(terms, &domain.PaymentTerm{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			Code:            row.Code,
			NameEn:          row.NameEn,
			NameAr:          row.NameAr,
			DueDays:         row.DueDays,
			DiscountDays:    row.DiscountDays,
			DiscountPercent: fromPgNumeric(row.DiscountPercent),
			IsActive:        row.IsActive,
			CreatedAt:       fromPgTimestamptz(row.CreatedAt),
		})
	}
	return terms, nil
}

func (r *SupplierRepository) CreateSupplierAccount(ctx context.Context, supplierID, glAccountID uuid.UUID, currencyCode string, isPrimary bool) error {
	q := r.queries(ctx)
	_, err := q.CreateSupplierAccount(ctx, sqlc.CreateSupplierAccountParams{
		SupplierID:   toPgUUID(supplierID),
		GlAccountID:  toPgUUID(glAccountID),
		CurrencyCode: currencyCode,
		IsPrimary:    isPrimary,
	})
	return err
}

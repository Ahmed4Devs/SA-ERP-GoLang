package postgres

import (
	"context"
	"errors"

	"erp/internal/platform/database/sqlc"
	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type customerRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewCustomerRepository(pool *pgxpool.Pool) ports.CustomerRepository {
	return &customerRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *customerRepository) CreateCustomer(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	row, err := r.queries.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		LegalEntityID:    toPgUUID(c.LegalEntityID),
		CustomerNumber:   c.CustomerNumber,
		CustomerType:     string(c.CustomerType),
		NameEn:           c.NameEn,
		NameAr:           c.NameAr,
		TaxID:            toPgTextPtr(c.TaxID),
		CrNumber:         toPgTextPtr(c.CRNumber),
		NationalID:       toPgTextPtr(c.NationalID),
		BuildingNo:       toPgText(c.Address.BuildingNo),
		StreetName:       toPgText(c.Address.StreetName),
		District:         toPgText(c.Address.District),
		City:             c.Address.City,
		PostalCode:       toPgText(c.Address.PostalCode),
		AdditionalNo:     toPgText(c.Address.AdditionalNo),
		CountryCode:      c.Address.CountryCode,
		Email:            toPgTextPtr(c.Email),
		Phone:            toPgTextPtr(c.Phone),
		ContactPerson:    toPgTextPtr(c.ContactPerson),
		PaymentTermsID:   toPgUUIDPtr(c.PaymentTermsID),
		CurrencyCode:     c.CurrencyCode,
		DefaultAccountID: toPgUUIDPtr(c.DefaultAccountID),
		Status:           string(c.Status),
	})
	if err != nil {
		return nil, err
	}
	c.ID = toUUID(row.ID)
	c.CreatedAt = row.CreatedAt.Time
	c.UpdatedAt = row.UpdatedAt.Time
	return c, nil
}

func (r *customerRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}
	return &domain.Customer{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		CustomerNumber: row.CustomerNumber,
		CustomerType:   domain.CustomerType(row.CustomerType),
		NameEn:         row.NameEn,
		NameAr:         row.NameAr,
		TaxID:          fromPgTextPtr(row.TaxID),
		CRNumber:       fromPgTextPtr(row.CrNumber),
		NationalID:     fromPgTextPtr(row.NationalID),
		Address: domain.SaudiAddress{
			BuildingNo:   fromPgText(row.BuildingNo),
			StreetName:   fromPgText(row.StreetName),
			District:     fromPgText(row.District),
			City:         row.City,
			PostalCode:   fromPgText(row.PostalCode),
			AdditionalNo: fromPgText(row.AdditionalNo),
			CountryCode:  row.CountryCode,
		},
		Email:              fromPgTextPtr(row.Email),
		Phone:              fromPgTextPtr(row.Phone),
		ContactPerson:      fromPgTextPtr(row.ContactPerson),
		PaymentTermsID:     toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:       row.CurrencyCode,
		DefaultAccountID:   toUUIDPtr(row.DefaultAccountID),
		Status:             domain.CustomerStatus(row.Status),
		PaymentTermsCode:   fromPgTextPtr(row.PaymentTermsCode),
		PaymentTermsNameAr: fromPgTextPtr(row.PaymentTermsNameAr),
		CreditLimit:        fromPgNumeric(row.CreditLimit),
		CurrentBalance:     fromPgNumeric(row.CurrentBalance),
		IsCreditHold:       row.IsCreditHold.Bool,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}, nil
}

func (r *customerRepository) GetCustomerByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Customer, error) {
	row, err := r.queries.GetCustomerByNumber(ctx, sqlc.GetCustomerByNumberParams{
		LegalEntityID:  toPgUUID(legalEntityID),
		CustomerNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}
	return &domain.Customer{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		CustomerNumber: row.CustomerNumber,
		CustomerType:   domain.CustomerType(row.CustomerType),
		NameEn:         row.NameEn,
		NameAr:         row.NameAr,
		TaxID:          fromPgTextPtr(row.TaxID),
		CRNumber:       fromPgTextPtr(row.CrNumber),
		NationalID:     fromPgTextPtr(row.NationalID),
		Address: domain.SaudiAddress{
			BuildingNo:   fromPgText(row.BuildingNo),
			StreetName:   fromPgText(row.StreetName),
			District:     fromPgText(row.District),
			City:         row.City,
			PostalCode:   fromPgText(row.PostalCode),
			AdditionalNo: fromPgText(row.AdditionalNo),
			CountryCode:  row.CountryCode,
		},
		Email:            fromPgTextPtr(row.Email),
		Phone:            fromPgTextPtr(row.Phone),
		ContactPerson:    fromPgTextPtr(row.ContactPerson),
		PaymentTermsID:   toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:     row.CurrencyCode,
		DefaultAccountID: toUUIDPtr(row.DefaultAccountID),
		Status:           domain.CustomerStatus(row.Status),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}, nil
}

func (r *customerRepository) ListCustomers(ctx context.Context, filter ports.CustomerFilter) ([]*domain.Customer, int64, error) {
	total, err := r.queries.CountCustomers(ctx, sqlc.CountCustomersParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       filter.CustomerType,
		Column4:       filter.Status,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListCustomers(ctx, sqlc.ListCustomersParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       filter.CustomerType,
		Column4:       filter.Status,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	customers := make([]*domain.Customer, len(rows))
	for i, row := range rows {
		customers[i] = &domain.Customer{
			ID:             toUUID(row.ID),
			LegalEntityID:  toUUID(row.LegalEntityID),
			CustomerNumber: row.CustomerNumber,
			CustomerType:   domain.CustomerType(row.CustomerType),
			NameEn:         row.NameEn,
			NameAr:         row.NameAr,
			TaxID:          fromPgTextPtr(row.TaxID),
			CRNumber:       fromPgTextPtr(row.CrNumber),
			NationalID:     fromPgTextPtr(row.NationalID),
			Address: domain.SaudiAddress{
				BuildingNo:   fromPgText(row.BuildingNo),
				StreetName:   fromPgText(row.StreetName),
				District:     fromPgText(row.District),
				City:         row.City,
				PostalCode:   fromPgText(row.PostalCode),
				AdditionalNo: fromPgText(row.AdditionalNo),
				CountryCode:  row.CountryCode,
			},
			Email:              fromPgTextPtr(row.Email),
			Phone:              fromPgTextPtr(row.Phone),
			ContactPerson:      fromPgTextPtr(row.ContactPerson),
			PaymentTermsID:     toUUIDPtr(row.PaymentTermsID),
			CurrencyCode:       row.CurrencyCode,
			DefaultAccountID:   toUUIDPtr(row.DefaultAccountID),
			Status:             domain.CustomerStatus(row.Status),
			PaymentTermsCode:   fromPgTextPtr(row.PaymentTermsCode),
			PaymentTermsNameAr: fromPgTextPtr(row.PaymentTermsNameAr),
			CreditLimit:        fromPgNumeric(row.CreditLimit),
			CurrentBalance:     fromPgNumeric(row.CurrentBalance),
			IsCreditHold:       row.IsCreditHold,
			CreatedAt:          row.CreatedAt.Time,
			UpdatedAt:          row.UpdatedAt.Time,
		}
	}
	return customers, total, nil
}

func (r *customerRepository) UpdateCustomer(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	row, err := r.queries.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
		ID:               toPgUUID(c.ID),
		NameEn:           c.NameEn,
		NameAr:           c.NameAr,
		TaxID:            toPgTextPtr(c.TaxID),
		CrNumber:         toPgTextPtr(c.CRNumber),
		NationalID:       toPgTextPtr(c.NationalID),
		BuildingNo:       toPgText(c.Address.BuildingNo),
		StreetName:       toPgText(c.Address.StreetName),
		District:         toPgText(c.Address.District),
		City:             c.Address.City,
		PostalCode:       toPgText(c.Address.PostalCode),
		AdditionalNo:     toPgText(c.Address.AdditionalNo),
		CountryCode:      c.Address.CountryCode,
		Email:            toPgTextPtr(c.Email),
		Phone:            toPgTextPtr(c.Phone),
		ContactPerson:    toPgTextPtr(c.ContactPerson),
		PaymentTermsID:   toPgUUIDPtr(c.PaymentTermsID),
		CurrencyCode:     c.CurrencyCode,
		DefaultAccountID: toPgUUIDPtr(c.DefaultAccountID),
	})
	if err != nil {
		return nil, err
	}
	c.UpdatedAt = row.UpdatedAt.Time
	return c, nil
}

func (r *customerRepository) UpdateCustomerStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error {
	_, err := r.queries.UpdateCustomerStatus(ctx, sqlc.UpdateCustomerStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	return err
}

func (r *customerRepository) CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error) {
	row, err := r.queries.CreatePaymentTerm(ctx, sqlc.CreatePaymentTermParams{
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
	term.ID = toUUID(row.ID)
	term.CreatedAt = row.CreatedAt.Time
	return term, nil
}

func (r *customerRepository) GetPaymentTermByID(ctx context.Context, id uuid.UUID) (*domain.PaymentTerm, error) {
	row, err := r.queries.GetPaymentTermByID(ctx, toPgUUID(id))
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
		CreatedAt:       row.CreatedAt.Time,
	}, nil
}

func (r *customerRepository) ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID, includeInactive bool) ([]*domain.PaymentTerm, error) {
	rows, err := r.queries.ListPaymentTerms(ctx, sqlc.ListPaymentTermsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       includeInactive,
	})
	if err != nil {
		return nil, err
	}
	terms := make([]*domain.PaymentTerm, len(rows))
	for i, row := range rows {
		terms[i] = &domain.PaymentTerm{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			Code:            row.Code,
			NameEn:          row.NameEn,
			NameAr:          row.NameAr,
			DueDays:         row.DueDays,
			DiscountDays:    row.DiscountDays,
			DiscountPercent: fromPgNumeric(row.DiscountPercent),
			IsActive:        row.IsActive,
			CreatedAt:       row.CreatedAt.Time,
		}
	}
	return terms, nil
}

func (r *customerRepository) UpsertCustomerAccount(ctx context.Context, account *domain.CustomerAccount) (*domain.CustomerAccount, error) {
	row, err := r.queries.UpsertCustomerAccount(ctx, sqlc.UpsertCustomerAccountParams{
		CustomerID:   toPgUUID(account.CustomerID),
		GlAccountID:  toPgUUID(account.GLAccountID),
		CurrencyCode: account.CurrencyCode,
		IsPrimary:    account.IsPrimary,
	})
	if err != nil {
		return nil, err
	}
	account.ID = toUUID(row.ID)
	account.CreatedAt = row.CreatedAt.Time
	return account, nil
}

func (r *customerRepository) GetCustomerAccounts(ctx context.Context, customerID uuid.UUID) ([]*domain.CustomerAccount, error) {
	rows, err := r.queries.GetCustomerAccounts(ctx, toPgUUID(customerID))
	if err != nil {
		return nil, err
	}
	accounts := make([]*domain.CustomerAccount, len(rows))
	for i, row := range rows {
		accounts[i] = &domain.CustomerAccount{
			ID:            toUUID(row.ID),
			CustomerID:    toUUID(row.CustomerID),
			GLAccountID:   toUUID(row.GlAccountID),
			CurrencyCode:  row.CurrencyCode,
			IsPrimary:     row.IsPrimary,
			AccountCode:   row.AccountCode,
			AccountNameAr: row.AccountNameAr,
			AccountNameEn: row.AccountNameEn,
			CreatedAt:     row.CreatedAt.Time,
		}
	}
	return accounts, nil
}

func (r *customerRepository) UpsertCreditLimit(ctx context.Context, limit *domain.CreditLimit) (*domain.CreditLimit, error) {
	row, err := r.queries.UpsertCreditLimit(ctx, sqlc.UpsertCreditLimitParams{
		CustomerID:     toPgUUID(limit.CustomerID),
		CreditLimit:    toPgNumeric(limit.CreditLimit),
		CurrentBalance: toPgNumeric(limit.CurrentBalance),
		MaxOverdueDays: limit.MaxOverdueDays,
		BlockAction:    string(limit.BlockAction),
		IsCreditHold:   limit.IsCreditHold,
		HoldReason:     toPgTextPtr(limit.HoldReason),
	})
	if err != nil {
		return nil, err
	}
	limit.ID = toUUID(row.ID)
	limit.UpdatedAt = row.UpdatedAt.Time
	return limit, nil
}

func (r *customerRepository) GetCreditLimit(ctx context.Context, customerID uuid.UUID) (*domain.CreditLimit, error) {
	row, err := r.queries.GetCreditLimitByCustomerID(ctx, toPgUUID(customerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.CreditLimit{
		ID:             toUUID(row.ID),
		CustomerID:     toUUID(row.CustomerID),
		CreditLimit:    fromPgNumeric(row.CreditLimit),
		CurrentBalance: fromPgNumeric(row.CurrentBalance),
		MaxOverdueDays: row.MaxOverdueDays,
		BlockAction:    domain.BlockAction(row.BlockAction),
		IsCreditHold:   row.IsCreditHold,
		HoldReason:     fromPgTextPtr(row.HoldReason),
		UpdatedAt:      row.UpdatedAt.Time,
	}, nil
}

func (r *customerRepository) UpdateCustomerBalance(ctx context.Context, customerID uuid.UUID, delta decimal.Decimal) (*domain.CreditLimit, error) {
	row, err := r.queries.UpdateCustomerBalance(ctx, sqlc.UpdateCustomerBalanceParams{
		CustomerID:     toPgUUID(customerID),
		CurrentBalance: toPgNumeric(delta),
	})
	if err != nil {
		return nil, err
	}
	return &domain.CreditLimit{
		ID:             toUUID(row.ID),
		CustomerID:     toUUID(row.CustomerID),
		CreditLimit:    fromPgNumeric(row.CreditLimit),
		CurrentBalance: fromPgNumeric(row.CurrentBalance),
		MaxOverdueDays: row.MaxOverdueDays,
		BlockAction:    domain.BlockAction(row.BlockAction),
		IsCreditHold:   row.IsCreditHold,
		HoldReason:     fromPgTextPtr(row.HoldReason),
		UpdatedAt:      row.UpdatedAt.Time,
	}, nil
}

func (r *customerRepository) SetCreditHold(ctx context.Context, customerID uuid.UUID, isHold bool, reason *string) error {
	_, err := r.queries.SetCreditHold(ctx, sqlc.SetCreditHoldParams{
		CustomerID:   toPgUUID(customerID),
		IsCreditHold: isHold,
		HoldReason:   toPgTextPtr(reason),
	})
	return err
}

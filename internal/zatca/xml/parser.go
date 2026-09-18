package xml

import (
	stdxml "encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"erp/internal/zatca/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type xmlAmount struct {
	Value string `xml:",chardata"`
}

type xmlParty struct {
	Party struct {
		PostalAddress struct {
			Street         string `xml:"StreetName"`
			BuildingNumber string `xml:"BuildingNumber"`
			District       string `xml:"CitySubdivisionName"`
			City           string `xml:"CityName"`
			PostalCode     string `xml:"PostalZone"`
			Country        struct {
				Code string `xml:"IdentificationCode"`
			} `xml:"Country"`
		} `xml:"PostalAddress"`
		TaxScheme struct {
			VATNumber string `xml:"CompanyID"`
		} `xml:"PartyTaxScheme"`
		LegalEntity struct {
			Name string `xml:"RegistrationName"`
		} `xml:"PartyLegalEntity"`
	} `xml:"Party"`
}

type xmlInvoice struct {
	XMLName      stdxml.Name `xml:"Invoice"`
	ID           string      `xml:"ID"`
	UUID         string      `xml:"UUID"`
	IssueDate    string      `xml:"IssueDate"`
	IssueTime    string      `xml:"IssueTime"`
	CurrencyCode string      `xml:"DocumentCurrencyCode"`
	TypeCode     struct {
		Name  string `xml:"name,attr"`
		Value string `xml:",chardata"`
	} `xml:"InvoiceTypeCode"`
	References []struct {
		ID         string `xml:"ID"`
		UUID       string `xml:"UUID"`
		Attachment struct {
			Value string `xml:"EmbeddedDocumentBinaryObject"`
		} `xml:"Attachment"`
	} `xml:"AdditionalDocumentReference"`
	Seller xmlParty `xml:"AccountingSupplierParty"`
	Buyer  xmlParty `xml:"AccountingCustomerParty"`
	Taxes  []struct {
		Amount xmlAmount `xml:"TaxAmount"`
	} `xml:"TaxTotal"`
	MonetaryTotal struct {
		LineExtension xmlAmount `xml:"LineExtensionAmount"`
		TaxExclusive  xmlAmount `xml:"TaxExclusiveAmount"`
		TaxInclusive  xmlAmount `xml:"TaxInclusiveAmount"`
		Allowance     xmlAmount `xml:"AllowanceTotalAmount"`
		Payable       xmlAmount `xml:"PayableAmount"`
	} `xml:"LegalMonetaryTotal"`
	Lines []struct {
		ID       string `xml:"ID"`
		Quantity struct {
			Value string `xml:",chardata"`
		} `xml:"InvoicedQuantity"`
		Extension xmlAmount `xml:"LineExtensionAmount"`
		Tax       struct {
			Amount  xmlAmount `xml:"TaxAmount"`
			Rounded xmlAmount `xml:"RoundingAmount"`
		} `xml:"TaxTotal"`
		Item struct {
			Name        string `xml:"Name"`
			TaxCategory struct {
				Percent string `xml:"Percent"`
			} `xml:"ClassifiedTaxCategory"`
		} `xml:"Item"`
		Price struct {
			Amount xmlAmount `xml:"PriceAmount"`
		} `xml:"Price"`
	} `xml:"InvoiceLine"`
}

// ParseInvoiceData reconstructs the validation and QR inputs from the stored
// UBL document. This keeps validation and signing bound to the exact document
// being processed instead of caller-supplied or placeholder values.
func ParseInvoiceData(xmlBytes []byte) (domain.InvoiceData, error) {
	var raw xmlInvoice
	if err := stdxml.Unmarshal(xmlBytes, &raw); err != nil {
		return domain.InvoiceData{}, fmt.Errorf("invalid UBL XML: %w", err)
	}
	if raw.XMLName.Local != "Invoice" {
		return domain.InvoiceData{}, fmt.Errorf("invalid UBL root element %q", raw.XMLName.Local)
	}

	parsedUUID, err := uuid.Parse(strings.TrimSpace(raw.UUID))
	if err != nil {
		parsedUUID = uuid.Nil
	}

	data := domain.InvoiceData{
		InvoiceNumber: strings.TrimSpace(raw.ID),
		UUID:          parsedUUID,
		IssueDate:     strings.TrimSpace(raw.IssueDate),
		IssueTime:     strings.TrimSpace(raw.IssueTime),
		CurrencyCode:  strings.TrimSpace(raw.CurrencyCode),
		InvoiceType:   domain.InvoiceTypeStandard,
		Seller:        mapXMLParty(raw.Seller),
		Buyer:         mapXMLParty(raw.Buyer),
	}
	if strings.HasPrefix(strings.TrimSpace(raw.TypeCode.Name), "02") {
		data.InvoiceType = domain.InvoiceTypeSimplified
	}
	switch strings.TrimSpace(raw.TypeCode.Value) {
	case "381":
		data.InvoiceSubtype = domain.InvoiceSubtypeCreditNote
	case "383":
		data.InvoiceSubtype = domain.InvoiceSubtypeDebitNote
	default:
		data.InvoiceSubtype = domain.InvoiceSubtypeTaxInvoice
	}

	for _, ref := range raw.References {
		switch strings.TrimSpace(ref.ID) {
		case "ICV":
			data.ICV, _ = strconv.ParseInt(strings.TrimSpace(ref.UUID), 10, 64)
		case "PIH":
			data.PIH = strings.TrimSpace(ref.Attachment.Value)
		}
	}

	if data.LineExtensionAmount, err = parseDecimal("line extension amount", raw.MonetaryTotal.LineExtension.Value); err != nil {
		return domain.InvoiceData{}, err
	}
	if data.TaxExclusiveAmount, err = parseDecimal("tax exclusive amount", raw.MonetaryTotal.TaxExclusive.Value); err != nil {
		return domain.InvoiceData{}, err
	}
	if data.TaxInclusiveAmount, err = parseDecimal("tax inclusive amount", raw.MonetaryTotal.TaxInclusive.Value); err != nil {
		return domain.InvoiceData{}, err
	}
	if data.AllowanceTotal, err = parseDecimal("allowance total", raw.MonetaryTotal.Allowance.Value); err != nil {
		return domain.InvoiceData{}, err
	}
	if data.PayableAmount, err = parseDecimal("payable amount", raw.MonetaryTotal.Payable.Value); err != nil {
		return domain.InvoiceData{}, err
	}
	if len(raw.Taxes) > 0 {
		if data.TaxTotal, err = parseDecimal("tax total", raw.Taxes[0].Amount.Value); err != nil {
			return domain.InvoiceData{}, err
		}
	}

	data.Lines = make([]domain.InvoiceLineData, 0, len(raw.Lines))
	for index, rawLine := range raw.Lines {
		quantity, parseErr := parseDecimal(fmt.Sprintf("line %d quantity", index+1), rawLine.Quantity.Value)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		unitPrice, parseErr := parseDecimal(fmt.Sprintf("line %d unit price", index+1), rawLine.Price.Amount.Value)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		vatRate, parseErr := parseDecimal(fmt.Sprintf("line %d VAT rate", index+1), rawLine.Item.TaxCategory.Percent)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		vatAmount, parseErr := parseDecimal(fmt.Sprintf("line %d VAT amount", index+1), rawLine.Tax.Amount.Value)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		lineTotal, parseErr := parseDecimal(fmt.Sprintf("line %d total", index+1), rawLine.Tax.Rounded.Value)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		extension, parseErr := parseDecimal(fmt.Sprintf("line %d extension", index+1), rawLine.Extension.Value)
		if parseErr != nil {
			return domain.InvoiceData{}, parseErr
		}
		discount := quantity.Mul(unitPrice).Sub(extension)
		if discount.IsNegative() {
			discount = decimal.Zero
		}
		data.Lines = append(data.Lines, domain.InvoiceLineData{
			ID:        strings.TrimSpace(rawLine.ID),
			Name:      strings.TrimSpace(rawLine.Item.Name),
			Quantity:  quantity,
			UnitPrice: unitPrice,
			Discount:  discount,
			VATRate:   vatRate,
			VATAmount: vatAmount,
			LineTotal: lineTotal,
		})
	}

	return data, nil
}

func mapXMLParty(raw xmlParty) domain.PartyInfo {
	return domain.PartyInfo{
		Name:           strings.TrimSpace(raw.Party.LegalEntity.Name),
		VATNumber:      strings.TrimSpace(raw.Party.TaxScheme.VATNumber),
		Street:         strings.TrimSpace(raw.Party.PostalAddress.Street),
		BuildingNumber: strings.TrimSpace(raw.Party.PostalAddress.BuildingNumber),
		District:       strings.TrimSpace(raw.Party.PostalAddress.District),
		City:           strings.TrimSpace(raw.Party.PostalAddress.City),
		PostalCode:     strings.TrimSpace(raw.Party.PostalAddress.PostalCode),
		CountryCode:    strings.TrimSpace(raw.Party.PostalAddress.Country.Code),
	}
}

func parseDecimal(field, value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, nil
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid %s %q: %w", field, value, err)
	}
	return parsed, nil
}

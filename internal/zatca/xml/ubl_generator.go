package xml

import (
	"bytes"
	"context"
	"fmt"
	"html"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
)

type UBLGenerator struct{}

func NewUBLGenerator() *UBLGenerator {
	return &UBLGenerator{}
}

func (g *UBLGenerator) GenerateInvoice(ctx context.Context, input domain.InvoiceData) ([]byte, error) {
	var buf bytes.Buffer

	// Invoice Type Code: 388 (Invoice), 381 (Credit Note), 383 (Debit Note)
	typeCode := "388"
	if input.InvoiceSubtype == domain.InvoiceSubtypeCreditNote {
		typeCode = "381"
	} else if input.InvoiceSubtype == domain.InvoiceSubtypeDebitNote {
		typeCode = "383"
	}

	// Subtype name mask: 0100000 for Standard, 0200000 for Simplified
	subTypeName := "0100000"
	profileID := "clearance:1.0"
	if input.InvoiceType == domain.InvoiceTypeSimplified {
		subTypeName = "0200000"
		profileID = "reporting:1.0"
	}

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2" `)
	buf.WriteString(`xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" `)
	buf.WriteString(`xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" `)
	buf.WriteString(`xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2">` + "\n")

	// UBL Extension Placeholder (populated during digital signing)
	buf.WriteString("  <ext:UBLExtensions>\n")
	buf.WriteString("    <ext:UBLExtension>\n")
	buf.WriteString("      <ext:ExtensionURI>urn:oasis:names:specification:ubl:dsig:enveloped:xades</ext:ExtensionURI>\n")
	buf.WriteString("      <ext:ExtensionContent>\n")
	buf.WriteString("        <sig:UBLDocumentSignatures xmlns:sig=\"urn:oasis:names:specification:ubl:schema:xsd:CommonSignatureComponents-2\"/>\n")
	buf.WriteString("      </ext:ExtensionContent>\n")
	buf.WriteString("    </ext:UBLExtension>\n")
	buf.WriteString("  </ext:UBLExtensions>\n")

	// Header
	buf.WriteString(fmt.Sprintf("  <cbc:ProfileID>%s</cbc:ProfileID>\n", profileID))
	buf.WriteString(fmt.Sprintf("  <cbc:ID>%s</cbc:ID>\n", html.EscapeString(input.InvoiceNumber)))
	buf.WriteString(fmt.Sprintf("  <cbc:UUID>%s</cbc:UUID>\n", input.UUID.String()))
	buf.WriteString(fmt.Sprintf("  <cbc:IssueDate>%s</cbc:IssueDate>\n", input.IssueDate))
	buf.WriteString(fmt.Sprintf("  <cbc:IssueTime>%s</cbc:IssueTime>\n", input.IssueTime))
	buf.WriteString(fmt.Sprintf("  <cbc:InvoiceTypeCode name=\"%s\">%s</cbc:InvoiceTypeCode>\n", subTypeName, typeCode))
	buf.WriteString(fmt.Sprintf("  <cbc:DocumentCurrencyCode>%s</cbc:DocumentCurrencyCode>\n", input.CurrencyCode))
	buf.WriteString(fmt.Sprintf("  <cbc:TaxCurrencyCode>%s</cbc:TaxCurrencyCode>\n", input.CurrencyCode))

	// Additional Document References: ICV & PIH
	buf.WriteString("  <cac:AdditionalDocumentReference>\n")
	buf.WriteString("    <cbc:ID>ICV</cbc:ID>\n")
	buf.WriteString(fmt.Sprintf("    <cbc:UUID>%d</cbc:UUID>\n", input.ICV))
	buf.WriteString("  </cac:AdditionalDocumentReference>\n")

	buf.WriteString("  <cac:AdditionalDocumentReference>\n")
	buf.WriteString("    <cbc:ID>PIH</cbc:ID>\n")
	buf.WriteString("    <cac:Attachment>\n")
	buf.WriteString(fmt.Sprintf("      <cbc:EmbeddedDocumentBinaryObject mimeCode=\"text/plain\">%s</cbc:EmbeddedDocumentBinaryObject>\n", input.PIH))
	buf.WriteString("    </cac:Attachment>\n")
	buf.WriteString("  </cac:AdditionalDocumentReference>\n")

	// Accounting Supplier Party (Seller)
	buf.WriteString("  <cac:AccountingSupplierParty>\n")
	buf.WriteString("    <cac:Party>\n")
	buf.WriteString("      <cac:PostalAddress>\n")
	buf.WriteString(fmt.Sprintf("        <cbc:StreetName>%s</cbc:StreetName>\n", html.EscapeString(input.Seller.Street)))
	buf.WriteString(fmt.Sprintf("        <cbc:BuildingNumber>%s</cbc:BuildingNumber>\n", html.EscapeString(input.Seller.BuildingNumber)))
	buf.WriteString(fmt.Sprintf("        <cbc:CitySubdivisionName>%s</cbc:CitySubdivisionName>\n", html.EscapeString(input.Seller.District)))
	buf.WriteString(fmt.Sprintf("        <cbc:CityName>%s</cbc:CityName>\n", html.EscapeString(input.Seller.City)))
	buf.WriteString(fmt.Sprintf("        <cbc:PostalZone>%s</cbc:PostalZone>\n", html.EscapeString(input.Seller.PostalCode)))
	buf.WriteString("        <cac:Country>\n")
	buf.WriteString(fmt.Sprintf("          <cbc:IdentificationCode>%s</cbc:IdentificationCode>\n", input.Seller.CountryCode))
	buf.WriteString("        </cac:Country>\n")
	buf.WriteString("      </cac:PostalAddress>\n")
	buf.WriteString("      <cac:PartyTaxScheme>\n")
	buf.WriteString(fmt.Sprintf("        <cbc:CompanyID>%s</cbc:CompanyID>\n", input.Seller.VATNumber))
	buf.WriteString("        <cac:TaxScheme>\n")
	buf.WriteString("          <cbc:ID>VAT</cbc:ID>\n")
	buf.WriteString("        </cac:TaxScheme>\n")
	buf.WriteString("      </cac:PartyTaxScheme>\n")
	buf.WriteString("      <cac:PartyLegalEntity>\n")
	buf.WriteString(fmt.Sprintf("        <cbc:RegistrationName>%s</cbc:RegistrationName>\n", html.EscapeString(input.Seller.Name)))
	buf.WriteString("      </cac:PartyLegalEntity>\n")
	buf.WriteString("    </cac:Party>\n")
	buf.WriteString("  </cac:AccountingSupplierParty>\n")

	// Accounting Customer Party (Buyer)
	buf.WriteString("  <cac:AccountingCustomerParty>\n")
	buf.WriteString("    <cac:Party>\n")
	buf.WriteString("      <cac:PostalAddress>\n")
	buf.WriteString(fmt.Sprintf("        <cbc:StreetName>%s</cbc:StreetName>\n", html.EscapeString(input.Buyer.Street)))
	buf.WriteString(fmt.Sprintf("        <cbc:BuildingNumber>%s</cbc:BuildingNumber>\n", html.EscapeString(input.Buyer.BuildingNumber)))
	buf.WriteString(fmt.Sprintf("        <cbc:CitySubdivisionName>%s</cbc:CitySubdivisionName>\n", html.EscapeString(input.Buyer.District)))
	buf.WriteString(fmt.Sprintf("        <cbc:CityName>%s</cbc:CityName>\n", html.EscapeString(input.Buyer.City)))
	buf.WriteString(fmt.Sprintf("        <cbc:PostalZone>%s</cbc:PostalZone>\n", html.EscapeString(input.Buyer.PostalCode)))
	buf.WriteString("        <cac:Country>\n")
	buf.WriteString(fmt.Sprintf("          <cbc:IdentificationCode>%s</cbc:IdentificationCode>\n", input.Buyer.CountryCode))
	buf.WriteString("        </cac:Country>\n")
	buf.WriteString("      </cac:PostalAddress>\n")
	if input.Buyer.VATNumber != "" {
		buf.WriteString("      <cac:PartyTaxScheme>\n")
		buf.WriteString(fmt.Sprintf("        <cbc:CompanyID>%s</cbc:CompanyID>\n", input.Buyer.VATNumber))
		buf.WriteString("        <cac:TaxScheme>\n")
		buf.WriteString("          <cbc:ID>VAT</cbc:ID>\n")
		buf.WriteString("        </cac:TaxScheme>\n")
		buf.WriteString("      </cac:PartyTaxScheme>\n")
	}
	buf.WriteString("      <cac:PartyLegalEntity>\n")
	buf.WriteString(fmt.Sprintf("        <cbc:RegistrationName>%s</cbc:RegistrationName>\n", html.EscapeString(input.Buyer.Name)))
	buf.WriteString("      </cac:PartyLegalEntity>\n")
	buf.WriteString("    </cac:Party>\n")
	buf.WriteString("  </cac:AccountingCustomerParty>\n")

	// Delivery
	buf.WriteString("  <cac:Delivery>\n")
	buf.WriteString(fmt.Sprintf("    <cbc:ActualDeliveryDate>%s</cbc:ActualDeliveryDate>\n", input.IssueDate))
	buf.WriteString("  </cac:Delivery>\n")

	// Tax Total
	buf.WriteString("  <cac:TaxTotal>\n")
	buf.WriteString(fmt.Sprintf("    <cbc:TaxAmount currencyID=\"%s\">%s</cbc:TaxAmount>\n", input.CurrencyCode, input.TaxTotal.StringFixed(2)))
	buf.WriteString("    <cac:TaxSubtotal>\n")
	buf.WriteString(fmt.Sprintf("      <cbc:TaxableAmount currencyID=\"%s\">%s</cbc:TaxableAmount>\n", input.CurrencyCode, input.TaxExclusiveAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("      <cbc:TaxAmount currencyID=\"%s\">%s</cbc:TaxAmount>\n", input.CurrencyCode, input.TaxTotal.StringFixed(2)))
	buf.WriteString("      <cac:TaxCategory>\n")
	buf.WriteString("        <cbc:ID>S</cbc:ID>\n")
	buf.WriteString("        <cbc:Percent>15.00</cbc:Percent>\n")
	buf.WriteString("        <cac:TaxScheme>\n")
	buf.WriteString("          <cbc:ID>VAT</cbc:ID>\n")
	buf.WriteString("        </cac:TaxScheme>\n")
	buf.WriteString("      </cac:TaxCategory>\n")
	buf.WriteString("    </cac:TaxSubtotal>\n")
	buf.WriteString("  </cac:TaxTotal>\n")

	// Legal Monetary Total
	buf.WriteString("  <cac:LegalMonetaryTotal>\n")
	buf.WriteString(fmt.Sprintf("    <cbc:LineExtensionAmount currencyID=\"%s\">%s</cbc:LineExtensionAmount>\n", input.CurrencyCode, input.LineExtensionAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("    <cbc:TaxExclusiveAmount currencyID=\"%s\">%s</cbc:TaxExclusiveAmount>\n", input.CurrencyCode, input.TaxExclusiveAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("    <cbc:TaxInclusiveAmount currencyID=\"%s\">%s</cbc:TaxInclusiveAmount>\n", input.CurrencyCode, input.TaxInclusiveAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("    <cbc:AllowanceTotalAmount currencyID=\"%s\">%s</cbc:AllowanceTotalAmount>\n", input.CurrencyCode, input.AllowanceTotal.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("    <cbc:PayableAmount currencyID=\"%s\">%s</cbc:PayableAmount>\n", input.CurrencyCode, input.PayableAmount.StringFixed(2)))
	buf.WriteString("  </cac:LegalMonetaryTotal>\n")

	// Invoice Lines
	for i, line := range input.Lines {
		buf.WriteString("  <cac:InvoiceLine>\n")
		buf.WriteString(fmt.Sprintf("    <cbc:ID>%d</cbc:ID>\n", i+1))
		buf.WriteString(fmt.Sprintf("    <cbc:InvoicedQuantity unitCode=\"PCE\">%s</cbc:InvoicedQuantity>\n", line.Quantity.StringFixed(2)))
		sub := line.Quantity.Mul(line.UnitPrice).Sub(line.Discount)
		buf.WriteString(fmt.Sprintf("    <cbc:LineExtensionAmount currencyID=\"%s\">%s</cbc:LineExtensionAmount>\n", input.CurrencyCode, sub.StringFixed(2)))
		buf.WriteString("    <cac:TaxTotal>\n")
		buf.WriteString(fmt.Sprintf("      <cbc:TaxAmount currencyID=\"%s\">%s</cbc:TaxAmount>\n", input.CurrencyCode, line.VATAmount.StringFixed(2)))
		buf.WriteString(fmt.Sprintf("      <cbc:RoundingAmount currencyID=\"%s\">%s</cbc:RoundingAmount>\n", input.CurrencyCode, line.LineTotal.StringFixed(2)))
		buf.WriteString("    </cac:TaxTotal>\n")
		buf.WriteString("    <cac:Item>\n")
		buf.WriteString(fmt.Sprintf("      <cbc:Name>%s</cbc:Name>\n", html.EscapeString(line.Name)))
		buf.WriteString("      <cac:ClassifiedTaxCategory>\n")
		buf.WriteString("        <cbc:ID>S</cbc:ID>\n")
		buf.WriteString(fmt.Sprintf("        <cbc:Percent>%s</cbc:Percent>\n", line.VATRate.StringFixed(2)))
		buf.WriteString("        <cac:TaxScheme>\n")
		buf.WriteString("          <cbc:ID>VAT</cbc:ID>\n")
		buf.WriteString("        </cac:TaxScheme>\n")
		buf.WriteString("      </cac:ClassifiedTaxCategory>\n")
		buf.WriteString("    </cac:Item>\n")
		buf.WriteString("    <cac:Price>\n")
		buf.WriteString(fmt.Sprintf("      <cbc:PriceAmount currencyID=\"%s\">%s</cbc:PriceAmount>\n", input.CurrencyCode, line.UnitPrice.StringFixed(2)))
		buf.WriteString("    </cac:Price>\n")
		buf.WriteString("  </cac:InvoiceLine>\n")
	}

	buf.WriteString("</Invoice>\n")
	return buf.Bytes(), nil
}

var _ ports.XMLGenerator = (*UBLGenerator)(nil)

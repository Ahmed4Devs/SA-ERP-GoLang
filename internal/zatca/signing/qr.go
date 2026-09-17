package signing

import (
	"bytes"
	"encoding/base64"

	"github.com/shopspring/decimal"
)

// TLVTag represents a single Tag-Length-Value entry.
type TLVTag struct {
	Tag   byte
	Value []byte
}

func encodeTLV(tag byte, val string) []byte {
	valueBytes := []byte(val)
	length := len(valueBytes)

	var buf bytes.Buffer
	buf.WriteByte(tag)
	if length < 128 {
		buf.WriteByte(byte(length))
	} else {
		// Multi-byte length encoding for long signatures / certificates
		lenBytes := []byte{byte(length >> 8), byte(length & 0xFF)}
		if lenBytes[0] > 0 {
			buf.WriteByte(0x82)
			buf.Write(lenBytes)
		} else {
			buf.WriteByte(0x81)
			buf.WriteByte(byte(length))
		}
	}
	buf.Write(valueBytes)
	return buf.Bytes()
}

// GenerateZATCAQRCode builds the canonical 8-tag TLV base64 QR string.
func GenerateZATCAQRCode(
	sellerName string,
	vatNumber string,
	timestamp string,
	totalWithVAT decimal.Decimal,
	vatTotal decimal.Decimal,
	invoiceHash string,
	digitalSignature string,
	publicKey string,
) string {
	var tlvBuf bytes.Buffer

	// Tag 1: Seller's Name
	tlvBuf.Write(encodeTLV(1, sellerName))
	// Tag 2: Seller's VAT Number
	tlvBuf.Write(encodeTLV(2, vatNumber))
	// Tag 3: Invoice Timestamp
	tlvBuf.Write(encodeTLV(3, timestamp))
	// Tag 4: Invoice Total (with VAT)
	tlvBuf.Write(encodeTLV(4, totalWithVAT.StringFixed(2)))
	// Tag 5: VAT Total
	tlvBuf.Write(encodeTLV(5, vatTotal.StringFixed(2)))

	// Tag 6: Invoice Hash (for Phase 2 compliance)
	if invoiceHash != "" {
		tlvBuf.Write(encodeTLV(6, invoiceHash))
	}

	// Tag 7: ECDSA Digital Signature (for Phase 2 compliance)
	if digitalSignature != "" {
		tlvBuf.Write(encodeTLV(7, digitalSignature))
	}

	// Tag 8: ECDSA Public Key (for Phase 2 compliance)
	if publicKey != "" {
		tlvBuf.Write(encodeTLV(8, publicKey))
	}

	return base64.StdEncoding.EncodeToString(tlvBuf.Bytes())
}

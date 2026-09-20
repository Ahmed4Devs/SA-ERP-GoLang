package signing

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/shopspring/decimal"
)

type InvoiceSigner struct{}

func NewInvoiceSigner() *InvoiceSigner {
	return &InvoiceSigner{}
}

// Hash computes the canonical SHA-256 digest of the XML document.
func (s *InvoiceSigner) Hash(xmlBytes []byte) (string, error) {
	document := etree.NewDocument()
	if err := document.ReadFromBytes(xmlBytes); err != nil {
		return "", fmt.Errorf("invalid XML for canonical hashing: %w", err)
	}
	root := document.Root()
	if root == nil {
		return "", fmt.Errorf("XML document has no root element")
	}

	// ZATCA's invoice digest excludes the UBL extensions (which hold the
	// signature), the QR reference, and the invoice signature node before
	// applying inclusive C14N 1.1 and SHA-256.
	removeElementsByNamespace(root, "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2", "UBLExtensions")
	removeQRDocumentReferences(root)
	removeElementsByNamespace(root, "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2", "Signature")

	canonical, err := dsig.MakeC14N11Canonicalizer().Canonicalize(root)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize XML: %w", err)
	}
	hash := sha256.Sum256(canonical)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

func removeElementsByNamespace(root *etree.Element, namespaceURI, localName string) {
	for _, element := range root.FindElements("//*") {
		if element.Tag == localName && element.NamespaceURI() == namespaceURI && element.Parent() != nil {
			element.Parent().RemoveChild(element)
		}
	}
}

func removeQRDocumentReferences(root *etree.Element) {
	for _, reference := range root.FindElements("//*") {
		if reference.Tag != "AdditionalDocumentReference" || reference.Parent() == nil {
			continue
		}
		for _, child := range reference.ChildElements() {
			if child.Tag == "ID" && strings.TrimSpace(child.Text()) == "QR" {
				reference.Parent().RemoveChild(reference)
				break
			}
		}
	}
}

// GenerateECDSAKeyPair produces the NIST P-256 key pair required for CSID generation.
func GenerateECDSAKeyPair() (privateKeyPEM string, publicKeyPEM string, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(privPEM), string(pubPEM), nil
}

// Sign hashes and signs the document, injecting the XAdES signature block.
func (s *InvoiceSigner) Sign(ctx context.Context, xmlBytes []byte, csid *domain.CSID, invoiceHash string) (signedXML string, digitalSig string, pubKeyHash string, err error) {
	if csid == nil {
		return "", "", "", fmt.Errorf("CSID is required for signing")
	}

	// 1. Decode and validate the ECDSA private key. Signing must fail closed:
	// fabricated signatures are never acceptable for tax invoices.
	block, _ := pem.Decode([]byte(csid.PrivateKeyPEM))
	if block == nil {
		return "", "", "", fmt.Errorf("invalid CSID private key PEM")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid CSID EC private key: %w", err)
	}
	digest, err := base64.StdEncoding.DecodeString(invoiceHash)
	if err != nil || len(digest) != sha256.Size {
		return "", "", "", fmt.Errorf("invalid invoice hash")
	}
	sigBytes, err := ecdsa.SignASN1(rand.Reader, priv, digest)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign invoice hash: %w", err)
	}
	digitalSig = base64.StdEncoding.EncodeToString(sigBytes)

	pubKeyHash = csid.BinarySecurityToken
	if pubKeyHash == "" {
		return "", "", "", fmt.Errorf("CSID certificate token is required for signing")
	}

	// 2. Inject XAdES signature block into XML
	xmlStr := string(xmlBytes)
	signatureBlock := fmt.Sprintf(`  <ext:UBLExtensions>
    <ext:UBLExtension>
      <ext:ExtensionURI>urn:oasis:names:specification:ubl:dsig:enveloped:xades</ext:ExtensionURI>
      <ext:ExtensionContent>
        <sig:UBLDocumentSignatures xmlns:sig="urn:oasis:names:specification:ubl:schema:xsd:CommonSignatureComponents-2">
          <sac:SignatureInformation xmlns:sac="urn:oasis:names:specification:ubl:schema:xsd:SignatureAggregateComponents-2">
            <cbc:ID>urn:oasis:names:specification:ubl:signature:1</cbc:ID>
            <sbc:ReferencedSignatureID xmlns:sbc="urn:oasis:names:specification:ubl:schema:xsd:SignatureBasicComponents-2">urn:oasis:names:specification:ubl:signature:Invoice</sbc:ReferencedSignatureID>
            <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#" Id="signature">
              <ds:SignedInfo>
                <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2006/12/xml-c14n11"/>
                <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256"/>
                <ds:Reference Id="invoiceSignedData" URI="">
                  <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
                  <ds:DigestValue>%s</ds:DigestValue>
                </ds:Reference>
              </ds:SignedInfo>
              <ds:SignatureValue>%s</ds:SignatureValue>
              <ds:KeyInfo>
                <ds:X509Data>
                  <ds:X509Certificate>%s</ds:X509Certificate>
                </ds:X509Data>
              </ds:KeyInfo>
            </ds:Signature>
          </sac:SignatureInformation>
        </sig:UBLDocumentSignatures>
      </ext:ExtensionContent>
    </ext:UBLExtension>
  </ext:UBLExtensions>`, invoiceHash, digitalSig, pubKeyHash)

	// Replace extension placeholder with signature
	if strings.Contains(xmlStr, "<ext:UBLExtensions>") {
		start := strings.Index(xmlStr, "<ext:UBLExtensions>")
		end := strings.Index(xmlStr, "</ext:UBLExtensions>") + len("</ext:UBLExtensions>")
		signedXML = xmlStr[:start] + signatureBlock + xmlStr[end:]
	} else {
		signedXML = xmlStr
	}

	return signedXML, digitalSig, pubKeyHash, nil
}

// GenerateQR wraps the TLV QR code encoder.
func (s *InvoiceSigner) GenerateQR(sellerName, vatNumber, timestamp string, totalAmount, vatAmount decimal.Decimal, invoiceHash, signature, publicKey string) (string, error) {
	qr := GenerateZATCAQRCode(
		sellerName,
		vatNumber,
		timestamp,
		totalAmount,
		vatAmount,
		invoiceHash,
		signature,
		publicKey,
	)
	return qr, nil
}

var _ ports.InvoiceSigner = (*InvoiceSigner)(nil)

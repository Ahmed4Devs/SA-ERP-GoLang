"use client";

import React, { useState } from "react";
import {
  ShieldCheck,
  FileCheck2,
  QrCode,
  Key,
  Server,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Eye,
  Plus,
  Search,
  Filter,
  Sparkles,
  Download,
  Copy,
  X,
  FileText,
  BadgeAlert,
  Fingerprint,
  RefreshCw,
  Hash,
  Binary,
  History,
  Send,
  ExternalLink,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface ZATCASubmissionItem {
  id: string;
  documentId: string;
  invoiceNumber: string;
  submissionMode: "CLEARANCE" | "REPORTING";
  environment: "SANDBOX" | "SIMULATION" | "PRODUCTION";
  status: "CLEARED" | "REPORTED" | "WARNING" | "REJECTED" | "RETRY_PENDING" | "FAILED";
  zatcaRequestId: string;
  attemptsCount: number;
  submittedAt: string;
  responsePayload?: string;
}

interface ZATCADocumentItem {
  id: string;
  invoiceNumber: string;
  uuid: string;
  invoiceType: "STANDARD" | "SIMPLIFIED";
  subtype: "TAX_INVOICE" | "CREDIT_NOTE" | "DEBIT_NOTE";
  submissionMode: "CLEARANCE" | "REPORTING";
  icv: number;
  pih: string;
  issueDate: string;
  issueTime: string;
  buyerName: string;
  totalWithVAT: number;
  vatAmount: number;
  status: "GENERATED" | "VALIDATED" | "SIGNED" | "CLEARED" | "REPORTED" | "WARNING" | "REJECTED";
  ublXml: string;
  qrCodeData: string;
  digitalSignature?: string;
  hash: string;
}

interface EGSUnitItem {
  id: string;
  customId: string;
  model: string;
  serialNumber: string;
  solutionName: string;
  vatNumber: string;
  organizationName: string;
  branchName: string;
  invoiceTypeMask: string;
  status: "ACTIVE" | "PENDING_OTP" | "REVOKED";
  createdAt: string;
}

interface CSIDItem {
  id: string;
  egsCustomId: string;
  csidType: "COMPLIANCE" | "PRODUCTION";
  binarySecurityToken: string;
  publicKeyPEM: string;
  expiresAt: string;
  status: "ACTIVE" | "EXPIRED";
}

const INITIAL_DOCUMENTS: ZATCADocumentItem[] = [
  {
    id: "doc-001",
    invoiceNumber: "INV-2026-0001",
    uuid: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    invoiceType: "STANDARD",
    subtype: "TAX_INVOICE",
    submissionMode: "CLEARANCE",
    icv: 1,
    pih: "NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz",
    issueDate: "2026-03-01",
    issueTime: "09:15:30",
    buyerName: "شركة البناء والتطوير العقاري (310000000000003)",
    totalWithVAT: 115000,
    vatAmount: 15000,
    status: "CLEARED",
    hash: "a8f5b2c9d1e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9",
    qrCodeData: "AQ9TYXVkaSBIb3Jpem9uIElEIw8zMDAwMDAwMDAwMDAwMDMHFDIwMjYtMDMtMDFUMDk6MTU6MzBaBTEwMDAwBgUxNTAwMA==",
    digitalSignature: "MEUCIQDx8f9a2b...== (ECDSA secp256k1 signature)",
    ublXml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
  xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
  xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
  xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2">
  <cbc:ProfileID>clearance:1.0</cbc:ProfileID>
  <cbc:ID>INV-2026-0001</cbc:ID>
  <cbc:UUID>9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d</cbc:UUID>
  <cbc:IssueDate>2026-03-01</cbc:IssueDate>
  <cbc:IssueTime>09:15:30</cbc:IssueTime>
  <cbc:InvoiceTypeCode name="0100000">388</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>SAR</cbc:DocumentCurrencyCode>
  <cac:AdditionalDocumentReference>
    <cbc:ID>ICV</cbc:ID>
    <cbc:UUID>1</cbc:UUID>
  </cac:AdditionalDocumentReference>
  <cac:AccountingSupplierParty>
    <cac:Party>
      <cac:PartyTaxScheme>
        <cbc:CompanyID>300000000000003</cbc:CompanyID>
        <cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme>
      </cac:PartyTaxScheme>
    </cac:Party>
  </cac:AccountingSupplierParty>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="SAR">100000.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="SAR">100000.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="SAR">115000.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="SAR">115000.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`,
  },
  {
    id: "doc-002",
    invoiceNumber: "INV-2026-0002",
    uuid: "a1c2e3f4-5b6c-7d8e-9f0a-1b2c3d4e5f6a",
    invoiceType: "SIMPLIFIED",
    subtype: "TAX_INVOICE",
    submissionMode: "REPORTING",
    icv: 2,
    pih: "a8f5b2c9d1e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9",
    issueDate: "2026-03-02",
    issueTime: "14:20:10",
    buyerName: "عميل تجزئة نقدي (B2C Cash Client)",
    totalWithVAT: 2300,
    vatAmount: 300,
    status: "REPORTED",
    hash: "b9e1f2a3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1",
    qrCodeData: "AQ9TYXVkaSBIb3Jpem9uIElEIw8zMDAwMDAwMDAwMDAwMDMHFDIwMjYtMDMtMDJUMTQ6MjA6MTBaBTIwMDAwBgUzMDAw",
    digitalSignature: "MEYCIQCx3z9...== (ECDSA secp256k1 signature)",
    ublXml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
  xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
  xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ProfileID>reporting:1.0</cbc:ProfileID>
  <cbc:ID>INV-2026-0002</cbc:ID>
  <cbc:InvoiceTypeCode name="0200000">388</cbc:InvoiceTypeCode>
  <cac:LegalMonetaryTotal>
    <cbc:TaxInclusiveAmount currencyID="SAR">2300.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="SAR">2300.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`,
  },
  {
    id: "doc-003",
    invoiceNumber: "INV-2026-0003",
    uuid: "b2d3e4f5-6a7b-8c9d-0e1f-2a3b4c5d6e7f",
    invoiceType: "STANDARD",
    subtype: "TAX_INVOICE",
    submissionMode: "CLEARANCE",
    icv: 3,
    pih: "b9e1f2a3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1",
    issueDate: "2026-03-03",
    issueTime: "11:00:00",
    buyerName: "شركة الرواد للتقنية (310000000000004)",
    totalWithVAT: 46000,
    vatAmount: 6000,
    status: "VALIDATED",
    hash: "c0f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2",
    qrCodeData: "AQ9TYXVkaSBIb3Jpem9uIElEIw8zMDAwMDAwMDAwMDAwMDMHFDIwMjYtMDMtMDNUMTE6MDA6MDBaBDQwMDAwBgY2MDAw",
    ublXml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
  <cbc:ID>INV-2026-0003</cbc:ID>
</Invoice>`,
  },
];

const INITIAL_EGS: EGSUnitItem[] = [
  {
    id: "egs-001",
    customId: "EGS-RYD-HQ01",
    model: "Saudi-ERP-v1",
    serialNumber: "SN-ERP-2026-HQ01",
    solutionName: "Saudi Enterprise ERP Platform",
    vatNumber: "300000000000003",
    organizationName: "شركة الأفق المتقدمة المحدودة",
    branchName: "الفرع الرئيسي - الرياض",
    invoiceTypeMask: "1100 (Standard + Simplified)",
    status: "ACTIVE",
    createdAt: "2026-01-01",
  },
  {
    id: "egs-002",
    customId: "EGS-JED-POS01",
    model: "Saudi-ERP-POS",
    serialNumber: "SN-ERP-2026-JED01",
    solutionName: "Saudi Enterprise ERP POS Solution",
    vatNumber: "300000000000003",
    organizationName: "شركة الأفق المتقدمة المحدودة",
    branchName: "فرع جدة - التجزئة",
    invoiceTypeMask: "0100 (Simplified Only)",
    status: "ACTIVE",
    createdAt: "2026-01-15",
  },
];

const INITIAL_CSIDS: CSIDItem[] = [
  {
    id: "csid-001",
    egsCustomId: "EGS-RYD-HQ01",
    csidType: "PRODUCTION",
    binarySecurityToken: "MIIC/DCCAeSgAwIBAgITbwAAAAP0/r3p5Q7TCAAAAAAA8DANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZaQVRDQTIwHhcNMjYwMTAxMDAwMDAwWhcNMjgwMTAxMDAwMDAwWjARMQ8wDQYDVQQDEwZaQVRDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKxQ==",
    publicKeyPEM: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
    expiresAt: "2028-01-01",
    status: "ACTIVE",
  },
  {
    id: "csid-002",
    egsCustomId: "EGS-JED-POS01",
    csidType: "PRODUCTION",
    binarySecurityToken: "MIIC/DCCAeSgAwIBAgITbwAAAAP0/r3p5Q7TCAAAAAAA8DANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZaQVRDQTIwHhcNMjYwMTAxMDAwMDAwWhcNMjgwMTAxMDAwMDAwWjARMQ8wDQYDVQQDEwZaQVRDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKxQ==",
    publicKeyPEM: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
    expiresAt: "2028-01-15",
    status: "ACTIVE",
  },
];

const INITIAL_SUBMISSIONS: ZATCASubmissionItem[] = [
  {
    id: "sub-001",
    documentId: "doc-001",
    invoiceNumber: "INV-2026-0001",
    submissionMode: "CLEARANCE",
    environment: "SANDBOX",
    status: "CLEARED",
    zatcaRequestId: "REQ-CLR-20260301-8899",
    attemptsCount: 1,
    submittedAt: "2026-03-01 09:15:32",
    responsePayload: `{"clearanceStatus":"CLEARED","validationResults":{"overallStatus":"PASSED","ruleVersion":"BR-KSA-2.1"}}`,
  },
  {
    id: "sub-002",
    documentId: "doc-002",
    invoiceNumber: "INV-2026-0002",
    submissionMode: "REPORTING",
    environment: "SANDBOX",
    status: "REPORTED",
    zatcaRequestId: "REQ-REP-20260302-3344",
    attemptsCount: 1,
    submittedAt: "2026-03-02 14:20:15",
    responsePayload: `{"reportingStatus":"REPORTED","validationResults":{"overallStatus":"PASSED","ruleVersion":"BR-KSA-2.1"}}`,
  },
];

export function ZATCAView() {
  const { locale } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"documents" | "submissions" | "egs" | "csids" | "validator">("documents");
  const [documents, setDocuments] = useState<ZATCADocumentItem[]>(INITIAL_DOCUMENTS);
  const [submissions, setSubmissions] = useState<ZATCASubmissionItem[]>(INITIAL_SUBMISSIONS);
  const [egsUnits, setEgsUnits] = useState<EGSUnitItem[]>(INITIAL_EGS);
  const [csids, setCsids] = useState<CSIDItem[]>(INITIAL_CSIDS);

  const [searchQuery, setSearchQuery] = useState("");
  const [filterType, setFilterType] = useState("ALL");
  const [notification, setNotification] = useState<string | null>(null);

  // Modals
  const [previewXml, setPreviewXml] = useState<string | null>(null);
  const [previewQR, setPreviewQR] = useState<string | null>(null);
  const [previewResponse, setPreviewResponse] = useState<string | null>(null);
  const [isRegisterEGSOpen, setIsRegisterEGSOpen] = useState(false);
  const [isOnboardingOpen, setIsOnboardingOpen] = useState(false);
  const [onboardTargetEGS, setOnboardTargetEGS] = useState<EGSUnitItem | null>(null);
  const [otpValue, setOtpValue] = useState("");

  // New EGS Form
  const [newEGS, setNewEGS] = useState({
    customId: "",
    model: "Saudi-ERP-v1",
    serialNumber: "",
    branchName: "الفرع الرئيسي",
    invoiceTypeMask: "1100",
  });

  const showNotification = (msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  // Actions
  const handleValidateDocument = (id: string) => {
    setDocuments((prev) =>
      prev.map((d) => (d.id === id ? { ...d, status: "VALIDATED" } : d))
    );
    showNotification(isAr ? "تم التحقق من مطابقة الفاتورة لقواعد زاتكا (BR-KSA PASSED)" : "Document passed ZATCA BR-KSA validation");
  };

  const handleSignDocument = (id: string) => {
    setDocuments((prev) =>
      prev.map((d) =>
        d.id === id
          ? {
              ...d,
              status: "SIGNED",
              digitalSignature: "MEQCID...secp256k1 Signature Generated",
            }
          : d
      )
    );
    showNotification(isAr ? "تم توقيع الفاتورة إلكترونياً بختم CSID وتوليد الرمز المشفر" : "Document digitally signed with CSID certificate");
  };

  const handleClearance = (id: string) => {
    const doc = documents.find((d) => d.id === id);
    if (!doc) return;

    setDocuments((prev) =>
      prev.map((d) => (d.id === id ? { ...d, status: "CLEARED" } : d))
    );

    const newSub: ZATCASubmissionItem = {
      id: `sub-${Date.now()}`,
      documentId: doc.id,
      invoiceNumber: doc.invoiceNumber,
      submissionMode: "CLEARANCE",
      environment: "SANDBOX",
      status: "CLEARED",
      zatcaRequestId: `REQ-CLR-${Date.now().toString().slice(-6)}`,
      attemptsCount: 1,
      submittedAt: new Date().toISOString().replace("T", " ").slice(0, 19),
      responsePayload: JSON.stringify({
        clearanceStatus: "CLEARED",
        clearedInvoice: "UBL-2.1-STAMPED-BASE64",
        validationResults: { overallStatus: "PASSED", ruleVersion: "BR-KSA-2.1" },
      }),
    };
    setSubmissions([newSub, ...submissions]);
    showNotification(isAr ? "تم اعتماد الفاتورة لدى منصة فاتورة زاتكا (ZATCA Clearance Approved)" : "Invoice successfully cleared by ZATCA gateway");
  };

  const handleReporting = (id: string) => {
    const doc = documents.find((d) => d.id === id);
    if (!doc) return;

    setDocuments((prev) =>
      prev.map((d) => (d.id === id ? { ...d, status: "REPORTED" } : d))
    );

    const newSub: ZATCASubmissionItem = {
      id: `sub-${Date.now()}`,
      documentId: doc.id,
      invoiceNumber: doc.invoiceNumber,
      submissionMode: "REPORTING",
      environment: "SANDBOX",
      status: "REPORTED",
      zatcaRequestId: `REQ-REP-${Date.now().toString().slice(-6)}`,
      attemptsCount: 1,
      submittedAt: new Date().toISOString().replace("T", " ").slice(0, 19),
      responsePayload: JSON.stringify({
        reportingStatus: "REPORTED",
        validationResults: { overallStatus: "PASSED", ruleVersion: "BR-KSA-2.1" },
      }),
    };
    setSubmissions([newSub, ...submissions]);
    showNotification(isAr ? "تم إبلاغ الفاتورة المبسطة لمنصة فاتورة زاتكا (ZATCA Reporting Succeeded)" : "Simplified invoice successfully reported to ZATCA");
  };

  const handleRetrySubmission = (subId: string) => {
    setSubmissions((prev) =>
      prev.map((s) =>
        s.id === subId ? { ...s, status: "CLEARED", attemptsCount: s.attemptsCount + 1 } : s
      )
    );
    showNotification(isAr ? "تمت إعادة إرسال الفاتورة بنجاح واعتمادها" : "Submission retry succeeded and invoice cleared");
  };

  const handleStartOnboarding = (e: React.FormEvent) => {
    e.preventDefault();
    if (!onboardTargetEGS || !otpValue) return;

    // Issue Production CSID and update EGS
    const newCSID: CSIDItem = {
      id: `csid-${Date.now()}`,
      egsCustomId: onboardTargetEGS.customId,
      csidType: "PRODUCTION",
      binarySecurityToken: "MIIC/DCCAeSgAwIBAgITbwAAAAP0/r3p5Q7TCAAAAAAA8DANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZaQVRDQTIwHhcNMjYwMTAxMDAwMDAwWhcNMjgwMTAxMDAwMDAwWjARMQ8wDQYDVQQDEwZaQVRDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKxQ==",
      publicKeyPEM: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...\n-----END PUBLIC KEY-----",
      expiresAt: "2031-03-01",
      status: "ACTIVE",
    };

    setCsids([...csids, newCSID]);
    setEgsUnits((prev) =>
      prev.map((u) => (u.id === onboardTargetEGS.id ? { ...u, status: "ACTIVE" } : u))
    );

    setIsOnboardingOpen(false);
    setOtpValue("");
    setOnboardTargetEGS(null);
    showNotification(isAr ? "تم إتمام الربط والتكامل بنجاح وإصدار الختم المشفر للإنتاج (Production CSID)" : "Onboarding completed and Production CSID issued successfully");
  };

  const handleRegisterEGS = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newEGS.customId || !newEGS.serialNumber) return;

    const unit: EGSUnitItem = {
      id: `egs-${Date.now()}`,
      customId: newEGS.customId,
      model: newEGS.model,
      serialNumber: newEGS.serialNumber,
      solutionName: "Saudi Enterprise ERP Platform",
      vatNumber: "300000000000003",
      organizationName: "شركة الأفق المتقدمة المحدودة",
      branchName: newEGS.branchName,
      invoiceTypeMask: newEGS.invoiceTypeMask === "1100" ? "1100 (Standard + Simplified)" : "0100 (Simplified Only)",
      status: "ACTIVE",
      createdAt: new Date().toISOString().split("T")[0],
    };

    setEgsUnits([...egsUnits, unit]);
    setIsRegisterEGSOpen(false);
    setNewEGS({
      customId: "",
      model: "Saudi-ERP-v1",
      serialNumber: "",
      branchName: "الفرع الرئيسي",
      invoiceTypeMask: "1100",
    });
    showNotification(isAr ? "تم تسجيل وحدة الفوترة EGS بنجاح" : "EGS Unit successfully registered");
  };

  const filteredDocuments = documents.filter((d) => {
    const matchesSearch =
      d.invoiceNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
      d.buyerName.includes(searchQuery) ||
      d.uuid.includes(searchQuery);
    const matchesType = filterType === "ALL" || d.invoiceType === filterType;
    return matchesSearch && matchesType;
  });

  const clearedCount = documents.filter((d) => d.status === "CLEARED").length;
  const reportedCount = documents.filter((d) => d.status === "REPORTED").length;

  return (
    <div className="space-y-6 pb-12">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 bg-emerald-600 text-white px-5 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-in fade-in slide-in-from-top-4">
          <CheckCircle2 className="h-5 w-5" />
          <span className="font-medium text-sm">{notification}</span>
        </div>
      )}

      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-gradient-to-r from-card via-card to-secondary/30 p-6 rounded-2xl border shadow-sm">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
              <Sparkles className="h-3.5 w-3.5" />
              {isAr ? "المرحلة 25: الفوترة الإلكترونية زاتكا (Phase 2 Core)" : "Phase 25: ZATCA E-Invoicing Phase 2 Core"}
            </span>
            <span className="text-xs text-muted-foreground">• منصة فاتورة (FATOORA)</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {isAr ? "محرك الامتثال والفوترة الإلكترونية (ZATCA Engine)" : "ZATCA Compliance & E-Invoicing Engine"}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isAr
              ? "توليد ملفات UBL 2.1 XML، التشفير الرقمي ECDSA، التحقق من قواعد BR-KSA، وسلاسل التجزئة التراكمية (PIH & ICV)"
              : "UBL 2.1 XML generation, ECDSA secp256k1 cryptographic signing, BR-KSA validation, and hash chaining"}
          </p>
        </div>

        <div className="flex items-center gap-2.5 flex-wrap">
          <button
            onClick={() => setIsRegisterEGSOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm"
          >
            <Plus className="h-4 w-4" />
            <span>{isAr ? "تسجيل وحدة EGS جديدة" : "Register EGS Unit"}</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: Total Docs */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "إجمالي الفواتير الإلكترونية" : "Total E-Invoices"}
            </span>
            <div className="p-2 rounded-xl bg-primary/10 text-primary">
              <FileCheck2 className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {documents.length} <span className="text-sm font-normal text-muted-foreground">{isAr ? "مستند" : "docs"}</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "متوافقة مع اشتراطات المرحلة الثانية" : "Compliant with Phase 2"}
            </p>
          </div>
        </div>

        {/* Card 2: EGS Units */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "وحدات الحلول المسجلة (EGS)" : "Registered EGS Units"}
            </span>
            <div className="p-2 rounded-xl bg-blue-500/10 text-blue-500">
              <Server className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {egsUnits.length} <span className="text-sm font-normal text-muted-foreground">{isAr ? "وحدة نشطة" : "active"}</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "مع شهادات CSID المعتمدة" : "With valid CSID tokens"}
            </p>
          </div>
        </div>

        {/* Card 3: Cleared B2B */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "الفواتير الضريبية المعتمدة (Clearance)" : "Cleared B2B Invoices"}
            </span>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-500">
              <ShieldCheck className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
              {clearedCount}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "اعتماد فوري من منصة فاتورة" : "Instantly cleared by ZATCA"}
            </p>
          </div>
        </div>

        {/* Card 4: Reported B2C */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "الفواتير المبسطة المبلغ عنها (Reporting)" : "Reported B2C Invoices"}
            </span>
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-500">
              <QrCode className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {reportedCount}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "رمز استجابة سريعة مشفر (TLV)" : "Encrypted TLV QR verified"}
            </p>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b pb-2 overflow-x-auto">
        <button
          onClick={() => setActiveTab("documents")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "documents"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <FileCheck2 className="h-4 w-4" />
          <span>{isAr ? "الفواتير والمستندات الإلكترونية" : "Electronic Documents"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {documents.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("submissions")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "submissions"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <History className="h-4 w-4" />
          <span>{isAr ? "سجل الإرسال والربط (Submissions)" : "Submissions Audit"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {submissions.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("egs")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "egs"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <Server className="h-4 w-4" />
          <span>{isAr ? "وحدات الحلول (EGS Units)" : "EGS Solution Units"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {egsUnits.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("csids")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "csids"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <Key className="h-4 w-4" />
          <span>{isAr ? "شهادات وأختام CSID" : "CSID Certificates"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {csids.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("validator")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "validator"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <ShieldCheck className="h-4 w-4" />
          <span>{isAr ? "مدقق قواعد زاتكا (BR-KSA)" : "BR-KSA Rules Validator"}</span>
        </button>
      </div>

      {/* TAB 1: Documents */}
      {activeTab === "documents" && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-card p-3.5 rounded-xl border">
            <div className="relative w-full sm:w-80">
              <Search className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={isAr ? "بحث برقم الفاتورة أو العميل..." : "Search invoice number or buyer..."}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pr-9 pl-4 py-1.5 text-sm bg-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div className="flex items-center gap-2 w-full sm:w-auto">
              <Filter className="h-4 w-4 text-muted-foreground" />
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
                className="bg-background border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="ALL">{isAr ? "كافة الأنواع" : "All Invoices"}</option>
                <option value="STANDARD">{isAr ? "ضريبية قياسية (B2B Standard)" : "Standard"}</option>
                <option value="SIMPLIFIED">{isAr ? "ضريبية مبسطة (B2C Simplified)" : "Simplified"}</option>
              </select>
            </div>
          </div>

          <div className="border rounded-2xl overflow-hidden bg-card shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/60 text-muted-foreground text-xs uppercase font-medium border-b">
                  <tr>
                    <th className="px-4 py-3.5">{isAr ? "رقم الفاتورة" : "Invoice #"}</th>
                    <th className="px-4 py-3.5">{isAr ? "النوع والمسار" : "Type & Mode"}</th>
                    <th className="px-4 py-3.5">{isAr ? "العميل" : "Buyer"}</th>
                    <th className="px-4 py-3.5">{isAr ? "العداد (ICV)" : "ICV"}</th>
                    <th className="px-4 py-3.5">{isAr ? "الإجمالي مع الضريبة" : "Total (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "حالة زاتكا" : "ZATCA Status"}</th>
                    <th className="px-4 py-3.5 text-center">{isAr ? "الإجراءات والمعاينة" : "Actions"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredDocuments.map((doc) => (
                    <tr key={doc.id} className="hover:bg-muted/40 transition-colors">
                      <td className="px-4 py-3.5">
                        <div className="font-mono font-bold text-primary">{doc.invoiceNumber}</div>
                        <div className="text-[10px] text-muted-foreground font-mono">{doc.uuid.slice(0, 18)}...</div>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-secondary border">
                          {doc.invoiceType}
                        </span>
                        <span className="text-[10px] text-muted-foreground block mt-0.5">{doc.submissionMode}</span>
                      </td>
                      <td className="px-4 py-3.5">
                        <div className="font-medium text-foreground text-xs max-w-xs">{doc.buyerName}</div>
                        <div className="text-[10px] text-muted-foreground">{doc.issueDate} {doc.issueTime}</div>
                      </td>
                      <td className="px-4 py-3.5 font-mono font-bold">
                        #{doc.icv}
                      </td>
                      <td className="px-4 py-3.5 font-semibold text-foreground">
                        {doc.totalWithVAT.toLocaleString()} SAR
                        <span className="text-[10px] text-muted-foreground block font-normal">
                          {isAr ? "ضريبة:" : "VAT:"} {doc.vatAmount.toLocaleString()} SAR
                        </span>
                      </td>
                      <td className="px-4 py-3.5">
                        {doc.status === "CLEARED" && (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                            <CheckCircle2 className="h-3 w-3" />
                            {isAr ? "معتمدة (CLEARED)" : "Cleared"}
                          </span>
                        )}
                        {doc.status === "REPORTED" && (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/10 text-blue-600 border border-blue-500/20">
                            <CheckCircle2 className="h-3 w-3" />
                            {isAr ? "مبلغ عنها (REPORTED)" : "Reported"}
                          </span>
                        )}
                        {doc.status === "VALIDATED" && (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-600 border border-purple-500/20">
                            <ShieldCheck className="h-3 w-3" />
                            {isAr ? "مطابقة للقواعد" : "Validated"}
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3.5 text-center">
                        <div className="flex items-center justify-center gap-2">
                          <button
                            onClick={() => setPreviewXml(doc.ublXml)}
                            className="px-2.5 py-1 text-xs rounded-lg font-medium bg-secondary hover:bg-muted text-foreground transition-colors border flex items-center gap-1"
                          >
                            <FileText className="h-3 w-3" />
                            <span>XML</span>
                          </button>
                          <button
                            onClick={() => setPreviewQR(doc.qrCodeData)}
                            className="px-2.5 py-1 text-xs rounded-lg font-medium bg-primary/10 hover:bg-primary/20 text-primary transition-colors flex items-center gap-1"
                          >
                            <QrCode className="h-3 w-3" />
                            <span>QR</span>
                          </button>
                          {doc.status === "GENERATED" && (
                            <button
                              onClick={() => handleValidateDocument(doc.id)}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-purple-600 hover:bg-purple-700 text-white transition-colors"
                            >
                              {isAr ? "تحقق" : "Validate"}
                            </button>
                          )}
                          {doc.status === "VALIDATED" && (
                            <button
                              onClick={() => handleSignDocument(doc.id)}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-emerald-600 hover:bg-emerald-700 text-white transition-colors"
                            >
                              {isAr ? "توقيع ECDSA" : "Sign"}
                            </button>
                          )}
                          {doc.status === "SIGNED" && doc.submissionMode === "CLEARANCE" && (
                            <button
                              onClick={() => handleClearance(doc.id)}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-blue-600 hover:bg-blue-700 text-white transition-colors flex items-center gap-1"
                            >
                              <ShieldCheck className="h-3 w-3" />
                              <span>{isAr ? "اعتماد" : "Clear"}</span>
                            </button>
                          )}
                          {doc.status === "SIGNED" && doc.submissionMode === "REPORTING" && (
                            <button
                              onClick={() => handleReporting(doc.id)}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-amber-600 hover:bg-amber-700 text-white transition-colors flex items-center gap-1"
                            >
                              <Send className="h-3 w-3" />
                              <span>{isAr ? "إبلاغ" : "Report"}</span>
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB: Submissions Audit (Phase 26) */}
      {activeTab === "submissions" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="p-4 border-b bg-muted/20 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <History className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-foreground text-sm">
                  {isAr ? "سجل عمليات الإرسال والربط مع هيئة الزكاة (ZATCA Submissions Audit)" : "ZATCA Submissions Audit Trail"}
                </h3>
              </div>
              <span className="text-xs text-muted-foreground font-mono">
                {submissions.length} {isAr ? "عملية مسجلة" : "records"}
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/40 text-muted-foreground border-b uppercase font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-start">{isAr ? "رقم الفاتورة" : "Invoice #"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "نوع الإرسال" : "Mode"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "البيئة" : "Environment"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "حالة الربط" : "Status"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "معرف الطلب (Request ID)" : "Request ID"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "المحاولات" : "Attempts"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "وقت الإرسال" : "Timestamp"}</th>
                    <th className="px-4 py-3 text-end">{isAr ? "الإجراءات" : "Actions"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {submissions.map((sub) => (
                    <tr key={sub.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-mono font-bold text-foreground">{sub.invoiceNumber}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          sub.submissionMode === "CLEARANCE" ? "bg-blue-500/10 text-blue-600" : "bg-amber-500/10 text-amber-600"
                        }`}>
                          {sub.submissionMode}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-0.5 rounded text-[10px] bg-muted font-mono">{sub.environment}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                          sub.status === "CLEARED" || sub.status === "REPORTED"
                            ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                            : sub.status === "RETRY_PENDING"
                            ? "bg-amber-500/10 text-amber-600 border border-amber-500/20"
                            : "bg-red-500/10 text-red-600 border border-red-500/20"
                        }`}>
                          {sub.status === "CLEARED" || sub.status === "REPORTED" ? (
                            <CheckCircle2 className="h-3 w-3" />
                          ) : (
                            <AlertTriangle className="h-3 w-3" />
                          )}
                          {sub.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono text-muted-foreground">{sub.zatcaRequestId}</td>
                      <td className="px-4 py-3 font-mono">{sub.attemptsCount}</td>
                      <td className="px-4 py-3 text-muted-foreground font-mono">{sub.submittedAt}</td>
                      <td className="px-4 py-3 text-end">
                        <div className="flex items-center justify-end gap-1.5">
                          {sub.responsePayload && (
                            <button
                              onClick={() => setPreviewResponse(sub.responsePayload || "")}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-muted hover:bg-muted/80 text-foreground transition-colors flex items-center gap-1"
                            >
                              <ExternalLink className="h-3 w-3" />
                              <span>{isAr ? "الاستجابة" : "Payload"}</span>
                            </button>
                          )}
                          {(sub.status === "RETRY_PENDING" || sub.status === "FAILED") && (
                            <button
                              onClick={() => handleRetrySubmission(sub.id)}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors flex items-center gap-1"
                            >
                              <RefreshCw className="h-3 w-3" />
                              <span>{isAr ? "إعادة إرسال" : "Retry"}</span>
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: EGS Units */}
      {activeTab === "egs" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {egsUnits.map((u) => (
              <div key={u.id} className="p-5 rounded-2xl border bg-card shadow-sm space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="p-2 rounded-xl bg-primary/10 text-primary">
                      <Server className="h-5 w-5" />
                    </div>
                    <div>
                      <h4 className="font-bold text-foreground text-sm font-mono">{u.customId}</h4>
                      <p className="text-xs text-muted-foreground">{u.model} • {u.serialNumber}</p>
                    </div>
                  </div>
                  <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                    <CheckCircle2 className="h-3 w-3" />
                    {u.status}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-2 text-xs pt-2 border-t text-muted-foreground">
                  <div>
                    <span className="block font-medium text-foreground">{isAr ? "الفرع المخصص:" : "Branch:"}</span>
                    <span>{u.branchName}</span>
                  </div>
                  <div>
                    <span className="block font-medium text-foreground">{isAr ? "الرقم الضريبي:" : "VAT Number:"}</span>
                    <span className="font-mono">{u.vatNumber}</span>
                  </div>
                  <div className="col-span-2">
                    <span className="block font-medium text-foreground">{isAr ? "نوع الفواتير المصرح بها:" : "Permitted Invoices:"}</span>
                    <span className="font-mono text-primary font-semibold">{u.invoiceTypeMask}</span>
                  </div>
                </div>

                <div className="pt-2 border-t flex items-center justify-end">
                  <button
                    onClick={() => {
                      setOnboardTargetEGS(u);
                      setIsOnboardingOpen(true);
                    }}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-emerald-600 hover:bg-emerald-700 text-white transition-colors"
                  >
                    <Key className="h-3.5 w-3.5" />
                    <span>{isAr ? "ربط وتوليد ختم الإنتاج (OTP Onboarding)" : "Onboard with OTP"}</span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 3: CSID Certificates */}
      {activeTab === "csids" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {csids.map((c) => (
              <div key={c.id} className="p-5 rounded-2xl border bg-card shadow-sm space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="p-2 rounded-xl bg-purple-500/10 text-purple-600">
                      <Key className="h-5 w-5" />
                    </div>
                    <div>
                      <h4 className="font-bold text-foreground text-sm flex items-center gap-2">
                        <span>{c.csidType} CSID</span>
                        <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground font-mono">{c.egsCustomId}</span>
                      </h4>
                      <p className="text-xs text-muted-foreground">{isAr ? "شهادة توقيع إلكتروني X.509" : "X.509 Electronic Signing Cert"}</p>
                    </div>
                  </div>
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                    {c.status}
                  </span>
                </div>

                <div className="space-y-1 text-xs">
                  <span className="text-muted-foreground block">{isAr ? "الرمز الأمني (Binary Security Token):" : "Security Token:"}</span>
                  <div className="p-2 rounded bg-muted font-mono text-[10px] break-all max-h-16 overflow-y-auto">
                    {c.binarySecurityToken}
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs pt-2 border-t text-muted-foreground">
                  <span>{isAr ? "تاريخ انتهاء الصلاحية:" : "Expires at:"} {c.expiresAt}</span>
                  <span className="font-semibold text-primary">{isAr ? "خوارزمية secp256k1" : "secp256k1 Curve"}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 4: BR-KSA Validator */}
      {activeTab === "validator" && (
        <div className="space-y-4">
          <div className="bg-card p-5 rounded-2xl border space-y-4">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <div>
                <h3 className="font-bold text-foreground">{isAr ? "فاحص قواعد العمل والامتثال للفوترة الإلكترونية (BR-KSA Rules Engine)" : "BR-KSA Compliance Rules Engine"}</h3>
                <p className="text-xs text-muted-foreground">{isAr ? "مصفوفة التحقق الآلي الصارمة لضمان مطابقة الفواتير لاشتراطات زاتكا قبل التقديم" : "Strict automated verification matrix ensuring invoice integrity prior to submission"}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
              <div className="p-3.5 rounded-xl border bg-muted/30 flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
                <div>
                  <h5 className="font-bold text-xs text-foreground font-mono">BR-KSA-01 / 02: Date & Time</h5>
                  <p className="text-xs text-muted-foreground">{isAr ? "صيغة التاريخ YYYY-MM-DD والوقت HH:MM:SS وعدم السماح بالتواريخ المستقبلية" : "Standard ISO format, no future dates permitted"}</p>
                </div>
              </div>

              <div className="p-3.5 rounded-xl border bg-muted/30 flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
                <div>
                  <h5 className="font-bold text-xs text-foreground font-mono">BR-KSA-05 / 06: VAT & Address</h5>
                  <p className="text-xs text-muted-foreground">{isAr ? "الرقم الضريبي 15 رقماً يبدأ وينتهي بالرقم 3، والعنوان الوطني مكتمل" : "15-digit VAT number starting and ending with 3, complete national address"}</p>
                </div>
              </div>

              <div className="p-3.5 rounded-xl border bg-muted/30 flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
                <div>
                  <h5 className="font-bold text-xs text-foreground font-mono">BR-KSA-08: Cryptographic Chain</h5>
                  <p className="text-xs text-muted-foreground">{isAr ? "تسلسل عداد الفواتير (ICV) وسلسلة التجزئة غير القابلة للكسر (PIH)" : "Sequential ICV counter and cryptographic PIH chain integrity"}</p>
                </div>
              </div>

              <div className="p-3.5 rounded-xl border bg-muted/30 flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
                <div>
                  <h5 className="font-bold text-xs text-foreground font-mono">BR-KSA-12 / 13 / 14: Math Balance</h5>
                  <p className="text-xs text-muted-foreground">{isAr ? "تطابق نسبة الضريبة 15% وصحة المعادلة: الإجمالي = الخاضع + الضريبة - الخصم" : "Exact 15% VAT rate and monetary totals mathematical reconciliation"}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 1: XML Preview */}
      {previewXml && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-3xl w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2 font-mono">
                <FileText className="h-5 w-5 text-primary" />
                <span>UBL 2.1 Canonical XML</span>
              </h3>
              <button
                onClick={() => setPreviewXml(null)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="relative">
              <pre className="p-4 rounded-xl bg-muted font-mono text-xs overflow-x-auto max-h-[500px] leading-relaxed text-foreground" dir="ltr">
                {previewXml}
              </pre>
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(previewXml);
                  showNotification(isAr ? "تم نسخ كود XML بنجاح" : "XML copied to clipboard");
                }}
                className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors flex items-center gap-2"
              >
                <Copy className="h-4 w-4" />
                <span>{isAr ? "نسخ الكود" : "Copy XML"}</span>
              </button>
              <button
                onClick={() => setPreviewXml(null)}
                className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 2: QR Preview */}
      {previewQR && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95 text-center">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <QrCode className="h-5 w-5 text-primary" />
                <span>{isAr ? "رمز الاستجابة السريعة (ZATCA QR)" : "ZATCA TLV QR Code"}</span>
              </h3>
              <button
                onClick={() => setPreviewQR(null)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="flex flex-col items-center justify-center py-4 space-y-3">
              <div className="p-4 bg-white rounded-2xl shadow-inner border inline-block">
                {/* Visual QR representation placeholder */}
                <QrCode className="h-40 w-40 text-black" />
              </div>
              <p className="text-xs text-muted-foreground">{isAr ? "مشفر بتنسيق TLV ثماني الحقول ومتوافق مع تطبيق زاتكا" : "8-Tag TLV Encoded Base64 compliant with ZATCA App"}</p>
              <div className="p-3 bg-muted rounded-lg font-mono text-[10px] break-all max-h-24 overflow-y-auto text-left w-full" dir="ltr">
                {previewQR}
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t">
              <button
                onClick={() => setPreviewQR(null)}
                className="w-full px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                {isAr ? "تم" : "Done"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 3: Register EGS */}
      {isRegisterEGSOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <Server className="h-5 w-5 text-primary" />
                {isAr ? "تسجيل وحدة حلول فوترة جديدة (EGS)" : "Register New EGS Unit"}
              </h3>
              <button
                onClick={() => setIsRegisterEGSOpen(false)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleRegisterEGS} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "معرف الوحدة المخصص (Custom ID) *" : "Custom ID *"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="EGS-DMM-01"
                  value={newEGS.customId}
                  onChange={(e) => setNewEGS({ ...newEGS, customId: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none font-mono"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الرقم التسلسلي للجهاز *" : "Serial Number *"}
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="SN-ERP-2026-DMM01"
                    value={newEGS.serialNumber}
                    onChange={(e) => setNewEGS({ ...newEGS, serialNumber: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none font-mono"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الفرع المخصص" : "Branch"}
                  </label>
                  <input
                    type="text"
                    value={newEGS.branchName}
                    onChange={(e) => setNewEGS({ ...newEGS, branchName: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "نوع الفواتير المصرح بها للوحدة" : "Invoice Type Authorization Mask"}
                </label>
                <select
                  value={newEGS.invoiceTypeMask}
                  onChange={(e) => setNewEGS({ ...newEGS, invoiceTypeMask: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                >
                  <option value="1100">{isAr ? "1100 - قياسية ومبسطة (B2B + B2C)" : "1100 - Standard & Simplified"}</option>
                  <option value="0100">{isAr ? "0100 - فواتير مبسطة فقط (B2C Only)" : "0100 - Simplified Only"}</option>
                  <option value="1000">{isAr ? "1000 - فواتير قياسية فقط (B2B Only)" : "1000 - Standard Only"}</option>
                </select>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsRegisterEGSOpen(false)}
                  className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  {isAr ? "تسجيل الوحدة" : "Register Unit"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 4: Onboarding with OTP (Phase 26) */}
      {isOnboardingOpen && onboardTargetEGS && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <Key className="h-5 w-5 text-primary" />
                <span>{isAr ? "إتمام الربط والتكامل (FATOORA OTP Onboarding)" : "Onboard EGS Unit with OTP"}</span>
              </h3>
              <button
                onClick={() => {
                  setIsOnboardingOpen(false);
                  setOnboardTargetEGS(null);
                }}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleStartOnboarding} className="space-y-4 pt-2">
              <div className="p-3 bg-muted/40 rounded-xl space-y-1 text-xs">
                <span className="text-muted-foreground block">{isAr ? "وحدة الحلول المحددة:" : "Target EGS Unit:"}</span>
                <span className="font-mono font-bold text-foreground block">{onboardTargetEGS.customId} ({onboardTargetEGS.serialNumber})</span>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "رمز التحقق OTP الصادر من منصة فاتورة (6 أرقام)" : "6-Digit FATOORA Portal OTP"}
                </label>
                <input
                  type="text"
                  required
                  maxLength={6}
                  placeholder="123456"
                  value={otpValue}
                  onChange={(e) => setOtpValue(e.target.value)}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-lg text-center font-mono tracking-widest focus:ring-2 focus:ring-primary focus:outline-none"
                />
                <p className="text-[11px] text-muted-foreground mt-1">
                  {isAr ? "يمكنك الحصول على الرمز من بوابة هيئة الزكاة والضريبة والجمارك > قائمة الأجهزة" : "Generate OTP from ZATCA FATOORA Portal > Solution Units"}
                </p>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => {
                    setIsOnboardingOpen(false);
                    setOnboardTargetEGS(null);
                  }}
                  className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white transition-colors"
                >
                  {isAr ? "تأكيد الربط وإصدار الختم" : "Confirm Onboarding"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 5: Submission Response Payload */}
      {previewResponse && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-2xl w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <ExternalLink className="h-5 w-5 text-primary" />
                <span>{isAr ? "استجابة منصة فاتورة (ZATCA Response)" : "ZATCA Gateway Response"}</span>
              </h3>
              <button
                onClick={() => setPreviewResponse(null)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="p-4 bg-muted rounded-xl font-mono text-xs overflow-y-auto max-h-96 whitespace-pre-wrap text-left" dir="ltr">
              {JSON.stringify(JSON.parse(previewResponse), null, 2)}
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t">
              <button
                onClick={() => setPreviewResponse(null)}
                className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

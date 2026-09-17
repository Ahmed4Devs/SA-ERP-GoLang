"use client";

import React, { useState } from "react";
import {
  FolderArchive,
  FileText,
  UploadCloud,
  Search,
  Filter,
  Download,
  Trash2,
  Paperclip,
  History,
  Layers,
  HardDrive,
  ShieldCheck,
  CheckCircle2,
  FileCode,
  FileCheck,
  X,
  Plus,
  ExternalLink,
  Copy,
  Tag,
  Building,
  RefreshCw,
  Eye,
  AlertCircle,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

export interface DocumentItem {
  id: string;
  title: string;
  description: string;
  category: "GENERAL" | "CONTRACT" | "INVOICE" | "RECEIPT" | "LEGAL" | "IDENTITY" | "TAX" | "ATTACHMENT" | "OTHER";
  folderPath: string;
  currentVersionId: string;
  isArchived: boolean;
  versionNumber: number;
  originalFilename: string;
  mimeType: string;
  byteSize: number;
  sha256Hash: string;
  createdAt: string;
  updatedAt: string;
  attachmentsCount?: number;
}

export interface DocumentVersion {
  id: string;
  fileId: string;
  versionNumber: number;
  originalFilename: string;
  mimeType: string;
  byteSize: number;
  sha256Hash: string;
  storageProvider: string;
  createdAt: string;
}

export interface EntityAttachmentItem {
  id: string;
  fileId: string;
  fileTitle: string;
  category: string;
  entityType: string;
  entityId: string;
  filename: string;
  byteSize: number;
  createdAt: string;
}

// Baseline mock data for interactive UI demonstration
const initialDocuments: DocumentItem[] = [
  {
    id: "doc-001",
    title: "السجل التجاري الموحد 2026 (Commercial Registration)",
    description: "نسخة السجل التجاري المعتمدة من وزارة التجارة لشركة الأفق المتقدمة",
    category: "LEGAL",
    folderPath: "/legal",
    currentVersionId: "ver-001-v2",
    isArchived: false,
    versionNumber: 2,
    originalFilename: "cr_certificate_2026_renewed.pdf",
    mimeType: "application/pdf",
    byteSize: 1048576,
    sha256Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    createdAt: "2026-01-01T10:00:00Z",
    updatedAt: "2026-02-15T14:30:00Z",
    attachmentsCount: 2,
  },
  {
    id: "doc-002",
    title: "شهادة تسجيل ضريبة القيمة المضافة (ZATCA VAT Certificate)",
    description: "شهادة الامتثال الضريبي الصادرة من هيئة الزكاة والضريبة والجمارك رقم 310123456700003",
    category: "TAX",
    folderPath: "/tax",
    currentVersionId: "ver-002-v1",
    isArchived: false,
    versionNumber: 1,
    originalFilename: "zatca_vat_cert_2026.pdf",
    mimeType: "application/pdf",
    byteSize: 524288,
    sha256Hash: "1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1ddb12879b",
    createdAt: "2026-01-02T11:15:00Z",
    updatedAt: "2026-01-02T11:15:00Z",
    attachmentsCount: 1,
  },
  {
    id: "doc-003",
    title: "عقد تأسيس الشركة ونظام الحوكمة (Articles of Association)",
    description: "عقد التأسيس الموثق والنظام الأساسي لشركة المساهمة المقفلة",
    category: "CONTRACT",
    folderPath: "/corporate",
    currentVersionId: "ver-003-v1",
    isArchived: false,
    versionNumber: 1,
    originalFilename: "articles_of_association.pdf",
    mimeType: "application/pdf",
    byteSize: 2097152,
    sha256Hash: "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
    createdAt: "2026-01-03T09:00:00Z",
    updatedAt: "2026-01-03T09:00:00Z",
    attachmentsCount: 1,
  },
  {
    id: "doc-004",
    title: "عقد توريد خوادم البنية التحتية السحابية",
    description: "عقد توريد وتجهيز سيرفرات قواعد البيانات مع شركة التقنية المتقدمة",
    category: "CONTRACT",
    folderPath: "/contracts/procurement",
    currentVersionId: "ver-004-v1",
    isArchived: false,
    versionNumber: 1,
    originalFilename: "server_procurement_contract_signed.pdf",
    mimeType: "application/pdf",
    byteSize: 1843200,
    sha256Hash: "8f481c4e72b9807785afee48bbca978112ca1bbdcafac231b39a23dc4da786ef",
    createdAt: "2026-02-10T15:20:00Z",
    updatedAt: "2026-02-10T15:20:00Z",
    attachmentsCount: 1,
  },
  {
    id: "doc-005",
    title: "سند قبض دفعة إيجار المقر الإداري",
    description: "سند صرف وإشعار تحويل بنكي للمالك عن الربع الأول 2026",
    category: "RECEIPT",
    folderPath: "/finance/receipts",
    currentVersionId: "ver-005-v1",
    isArchived: false,
    versionNumber: 1,
    originalFilename: "lease_payment_receipt_q1_2026.pdf",
    mimeType: "application/pdf",
    byteSize: 312500,
    sha256Hash: "4ca95991b7852b855e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934",
    createdAt: "2026-03-01T12:00:00Z",
    updatedAt: "2026-03-01T12:00:00Z",
    attachmentsCount: 1,
  },
];

const initialEntityAttachments: EntityAttachmentItem[] = [
  {
    id: "att-001",
    fileId: "doc-004",
    fileTitle: "عقد توريد خوادم البنية التحتية السحابية",
    category: "CONTRACT",
    entityType: "purchase_order",
    entityId: "PO-2026-00042",
    filename: "server_procurement_contract_signed.pdf",
    byteSize: 1843200,
    createdAt: "2026-02-10T15:22:00Z",
  },
  {
    id: "att-002",
    fileId: "doc-002",
    fileTitle: "شهادة تسجيل ضريبة القيمة المضافة (ZATCA VAT Certificate)",
    category: "TAX",
    entityType: "sales_invoice",
    entityId: "INV-2026-00188",
    filename: "zatca_vat_cert_2026.pdf",
    byteSize: 524288,
    createdAt: "2026-01-05T08:30:00Z",
  },
  {
    id: "att-003",
    fileId: "doc-001",
    fileTitle: "السجل التجاري الموحد 2026 (Commercial Registration)",
    category: "LEGAL",
    entityType: "supplier",
    entityId: "SUP-0001",
    filename: "cr_certificate_2026_renewed.pdf",
    byteSize: 1048576,
    createdAt: "2026-01-02T10:00:00Z",
  },
];

export function DocumentsView() {
  const { locale } = useLocale();
  const isAr = locale === "ar";

  const [documentsList, setDocumentsList] = useState<DocumentItem[]>(initialDocuments);
  const [attachmentsList, setAttachmentsList] = useState<EntityAttachmentItem[]>(initialEntityAttachments);
  const [activeTab, setActiveTab] = useState<"library" | "attachments" | "archived">("library");
  const [selectedCategory, setSelectedCategory] = useState<string>("ALL");
  const [searchQuery, setSearchQuery] = useState<string>("");

  // Modals state
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [isVersionsOpen, setIsVersionsOpen] = useState(false);
  const [isAttachOpen, setIsAttachOpen] = useState(false);
  const [selectedDoc, setSelectedDoc] = useState<DocumentItem | null>(null);
  const [docVersions, setDocVersions] = useState<DocumentVersion[]>([]);
  const [copiedHash, setCopiedHash] = useState<string | null>(null);

  // Form states
  const [uploadTitle, setUploadTitle] = useState("");
  const [uploadCategory, setUploadCategory] = useState<DocumentItem["category"]>("GENERAL");
  const [uploadFolderPath, setUploadFolderPath] = useState("/");
  const [uploadDescription, setUploadDescription] = useState("");
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadEntityType, setUploadEntityType] = useState("");
  const [uploadEntityId, setUploadEntityId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Attach modal form
  const [attachEntityType, setAttachEntityType] = useState("sales_invoice");
  const [attachEntityId, setAttachEntityId] = useState("");

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const getCategoryBadgeClass = (category: string) => {
    switch (category) {
      case "LEGAL":
        return "bg-purple-100 text-purple-800 dark:bg-purple-950/60 dark:text-purple-300 border-purple-200 dark:border-purple-800";
      case "TAX":
        return "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800";
      case "CONTRACT":
        return "bg-blue-100 text-blue-800 dark:bg-blue-950/60 dark:text-blue-300 border-blue-200 dark:border-blue-800";
      case "INVOICE":
        return "bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 border-amber-200 dark:border-amber-800";
      case "RECEIPT":
        return "bg-teal-100 text-teal-800 dark:bg-teal-950/60 dark:text-teal-300 border-teal-200 dark:border-teal-800";
      case "IDENTITY":
        return "bg-indigo-100 text-indigo-800 dark:bg-indigo-950/60 dark:text-indigo-300 border-indigo-200 dark:border-indigo-800";
      default:
        return "bg-muted text-muted-foreground border-border";
    }
  };

  const getCategoryLabel = (cat: string) => {
    switch (cat) {
      case "LEGAL":
        return isAr ? "وثائق قانونية" : "Legal";
      case "TAX":
        return isAr ? "ضرائب وزكاة" : "Tax & ZATCA";
      case "CONTRACT":
        return isAr ? "عقود واتفاقيات" : "Contracts";
      case "INVOICE":
        return isAr ? "فواتير" : "Invoices";
      case "RECEIPT":
        return isAr ? "سندات وإيصالات" : "Receipts";
      case "IDENTITY":
        return isAr ? "هويات وثبوتيات" : "Identities";
      case "ATTACHMENT":
        return isAr ? "مرفقات عامة" : "Attachments";
      default:
        return isAr ? "عام" : "General";
    }
  };

  // KPI calculations
  const totalFiles = documentsList.filter((d) => !d.isArchived).length;
  const totalVersions = documentsList.reduce((acc, d) => acc + (d.isArchived ? 0 : d.versionNumber), 0);
  const totalBytes = documentsList.reduce((acc, d) => acc + (d.isArchived ? 0 : d.byteSize), 0);
  const totalAttachments = attachmentsList.length;

  // Filtered documents
  const filteredDocs = documentsList.filter((doc) => {
    if (activeTab === "archived") {
      return doc.isArchived;
    }
    if (doc.isArchived) return false;

    const matchesCategory = selectedCategory === "ALL" || doc.category === selectedCategory;
    const matchesSearch =
      searchQuery === "" ||
      doc.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      doc.originalFilename.toLowerCase().includes(searchQuery.toLowerCase()) ||
      doc.description.toLowerCase().includes(searchQuery.toLowerCase());

    return matchesCategory && matchesSearch;
  });

  const handleCopyHash = (hash: string) => {
    navigator.clipboard.writeText(hash);
    setCopiedHash(hash);
    setTimeout(() => setCopiedHash(null), 2000);
  };

  const handleOpenVersions = (doc: DocumentItem) => {
    setSelectedDoc(doc);
    // Generate sample version history
    const versions: DocumentVersion[] = [
      {
        id: doc.currentVersionId,
        fileId: doc.id,
        versionNumber: doc.versionNumber,
        originalFilename: doc.originalFilename,
        mimeType: doc.mimeType,
        byteSize: doc.byteSize,
        sha256Hash: doc.sha256Hash,
        storageProvider: "local",
        createdAt: doc.updatedAt,
      },
    ];
    if (doc.versionNumber > 1) {
      versions.push({
        id: "ver-legacy-v1",
        fileId: doc.id,
        versionNumber: 1,
        originalFilename: doc.originalFilename.replace("_renewed", ""),
        mimeType: doc.mimeType,
        byteSize: Math.round(doc.byteSize * 0.9),
        sha256Hash: "0a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef",
        storageProvider: "local",
        createdAt: doc.createdAt,
      });
    }
    setDocVersions(versions);
    setIsVersionsOpen(true);
  };

  const handleOpenAttach = (doc: DocumentItem) => {
    setSelectedDoc(doc);
    setAttachEntityType("sales_invoice");
    setAttachEntityId("");
    setIsAttachOpen(true);
  };

  const handleArchiveDoc = (docId: string) => {
    if (!confirm(isAr ? "هل أنت متأكد من رغبتك في أرشفة هذا المستند؟" : "Are you sure you want to archive this document?")) {
      return;
    }
    setDocumentsList((prev) => prev.map((d) => (d.id === docId ? { ...d, isArchived: true } : d)));
  };

  const handleRestoreDoc = (docId: string) => {
    setDocumentsList((prev) => prev.map((d) => (d.id === docId ? { ...d, isArchived: false } : d)));
  };

  const handleUploadSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!uploadTitle.trim()) {
      alert(isAr ? "يرجى كتابة عنوان المستند" : "Please enter document title");
      return;
    }
    setIsSubmitting(true);

    setTimeout(() => {
      const newDocId = `doc-${Date.now()}`;
      const filename = uploadFile ? uploadFile.name : `document_${Date.now()}.pdf`;
      const size = uploadFile ? uploadFile.size : 786432;

      const newDoc: DocumentItem = {
        id: newDocId,
        title: uploadTitle,
        description: uploadDescription || (isAr ? "مستند مدخل حديثاً في النظام" : "Newly uploaded document"),
        category: uploadCategory,
        folderPath: uploadFolderPath || "/",
        currentVersionId: `ver-${Date.now()}-v1`,
        isArchived: false,
        versionNumber: 1,
        originalFilename: filename,
        mimeType: uploadFile ? uploadFile.type || "application/pdf" : "application/pdf",
        byteSize: size,
        sha256Hash: "d41d8cd98f00b204e9800998ecf8427e" + Math.random().toString(16).substring(2, 10),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        attachmentsCount: uploadEntityType ? 1 : 0,
      };

      setDocumentsList([newDoc, ...documentsList]);

      if (uploadEntityType && uploadEntityId) {
        const newAtt: EntityAttachmentItem = {
          id: `att-${Date.now()}`,
          fileId: newDocId,
          fileTitle: newDoc.title,
          category: newDoc.category,
          entityType: uploadEntityType,
          entityId: uploadEntityId,
          filename: filename,
          byteSize: size,
          createdAt: new Date().toISOString(),
        };
        setAttachmentsList([newAtt, ...attachmentsList]);
      }

      setIsSubmitting(false);
      setIsUploadOpen(false);
      // Reset form
      setUploadTitle("");
      setUploadDescription("");
      setUploadFolderPath("/");
      setUploadFile(null);
      setUploadEntityType("");
      setUploadEntityId("");
    }, 500);
  };

  const handleAttachSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDoc || !attachEntityId.trim()) return;

    const newAtt: EntityAttachmentItem = {
      id: `att-${Date.now()}`,
      fileId: selectedDoc.id,
      fileTitle: selectedDoc.title,
      category: selectedDoc.category,
      entityType: attachEntityType,
      entityId: attachEntityId,
      filename: selectedDoc.originalFilename,
      byteSize: selectedDoc.byteSize,
      createdAt: new Date().toISOString(),
    };

    setAttachmentsList([newAtt, ...attachmentsList]);
    setDocumentsList((prev) =>
      prev.map((d) => (d.id === selectedDoc.id ? { ...d, attachmentsCount: (d.attachmentsCount || 0) + 1 } : d))
    );
    setIsAttachOpen(false);
  };

  const handleDetach = (attId: string, fileId: string) => {
    if (!confirm(isAr ? "هل تريد إلغاء ربط المستند بهذه العملية؟" : "Detach document from entity?")) return;
    setAttachmentsList((prev) => prev.filter((a) => a.id !== attId));
    setDocumentsList((prev) =>
      prev.map((d) => (d.id === fileId ? { ...d, attachmentsCount: Math.max(0, (d.attachmentsCount || 1) - 1) } : d))
    );
  };

  return (
    <div className="space-y-6">
      {/* 1. Header & Primary Action */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight">
              {isAr ? "إدارة وتخزين الوثائق والمستندات" : "Secure Document Storage & Files"}
            </h1>
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              Phase 27
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            {isAr
              ? "منصة مركزية لإدارة وثائق المنشأة، أرشفة الإصدارات بالبصمة الرقمية (SHA-256)، والربط الشامل بكيانات الـ ERP"
              : "Enterprise document platform with SHA-256 integrity verification, versioning, and polymorphic entity linking"}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsUploadOpen(true)}
            className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition shadow-sm"
          >
            <UploadCloud className="h-4 w-4" />
            {isAr ? "رفع مستند جديد" : "Upload Document"}
          </button>
        </div>
      </div>

      {/* 2. KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "إجمالي المستندات النشطة" : "Active Documents"}
            </span>
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <FolderArchive className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{totalFiles}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">{isAr ? "ملف محفوظ" : "files"}</span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "إصدارات الملفات المحفوظة" : "Document Versions"}
            </span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <History className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{totalVersions}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">{isAr ? "إصدار تعاقبي" : "versions"}</span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "حجم التخزين الفعلي" : "Storage Volume"}
            </span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <HardDrive className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{formatBytes(totalBytes)}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">
              {isAr ? "مساحة مشغولة" : "used"}
            </span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "مرفقات العمليات المرتبطة" : "Linked Attachments"}
            </span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <Paperclip className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{totalAttachments}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">
              {isAr ? "عملية مرتبطة" : "attachments"}
            </span>
          </div>
        </div>
      </div>

      {/* 3. Navigation Tabs */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b">
        <div className="flex items-center gap-2 overflow-x-auto pb-px">
          <button
            onClick={() => setActiveTab("library")}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition whitespace-nowrap flex items-center gap-2 ${
              activeTab === "library"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <FolderArchive className="h-4 w-4" />
            {isAr ? "مكتبة المستندات" : "Document Library"}
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-xs bg-muted">
              {documentsList.filter((d) => !d.isArchived).length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("attachments")}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition whitespace-nowrap flex items-center gap-2 ${
              activeTab === "attachments"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <Paperclip className="h-4 w-4" />
            {isAr ? "مرفقات العمليات والكيانات" : "Entity Attachments"}
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-xs bg-muted">
              {attachmentsList.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("archived")}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition whitespace-nowrap flex items-center gap-2 ${
              activeTab === "archived"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <Trash2 className="h-4 w-4" />
            {isAr ? "أرشيف المستندات" : "Archived Documents"}
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-xs bg-muted">
              {documentsList.filter((d) => d.isArchived).length}
            </span>
          </button>
        </div>
      </div>

      {/* 4. Tab 1: Document Library */}
      {activeTab === "library" && (
        <div className="space-y-4">
          {/* Category Filter Pills & Search */}
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 bg-card p-3 rounded-lg border">
            <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0">
              {[
                { key: "ALL", labelAr: "الكل", labelEn: "All" },
                { key: "LEGAL", labelAr: "وثائق رسمية", labelEn: "Legal" },
                { key: "TAX", labelAr: "ضرائب وزكاة", labelEn: "Tax" },
                { key: "CONTRACT", labelAr: "عقود", labelEn: "Contracts" },
                { key: "INVOICE", labelAr: "فواتير", labelEn: "Invoices" },
                { key: "RECEIPT", labelAr: "سندات", labelEn: "Receipts" },
                { key: "IDENTITY", labelAr: "هويات", labelEn: "Identities" },
              ].map((c) => (
                <button
                  key={c.key}
                  onClick={() => setSelectedCategory(c.key)}
                  className={`px-3 py-1.5 rounded-md text-xs font-medium transition whitespace-nowrap ${
                    selectedCategory === c.key
                      ? "bg-primary text-primary-foreground shadow-sm"
                      : "bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground"
                  }`}
                >
                  {isAr ? c.labelAr : c.labelEn}
                </button>
              ))}
            </div>

            <div className="relative w-full md:w-72">
              <Search className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground pointer-events-none" />
              <input
                type="text"
                placeholder={isAr ? "بحث في العناوين والملفات..." : "Search documents..."}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pr-9 pl-3 py-1.5 text-xs rounded-md border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              />
            </div>
          </div>

          {/* Documents Table */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-start">
                <thead className="bg-muted/50 border-b text-xs text-muted-foreground font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-start">{isAr ? "عنوان المستند والملف" : "Document & File"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "التصنيف" : "Category"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "المجلد" : "Folder"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "الإصدار والحجم" : "Version & Size"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "البصمة الرقمية (SHA-256)" : "SHA-256 Integrity"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "المرفقات" : "Attached"}</th>
                    <th className="px-4 py-3 text-end">{isAr ? "الإجراءات" : "Actions"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredDocs.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground text-sm">
                        {isAr ? "لا توجد مستندات مطابقة لمعايير البحث" : "No documents match the search criteria"}
                      </td>
                    </tr>
                  ) : (
                    filteredDocs.map((doc) => (
                      <tr key={doc.id} className="hover:bg-muted/30 transition">
                        <td className="px-4 py-3.5">
                          <div className="flex items-start gap-3">
                            <div className="p-2 rounded-lg bg-primary/10 text-primary mt-0.5">
                              <FileText className="h-4 w-4" />
                            </div>
                            <div>
                              <span className="font-semibold text-foreground block">{doc.title}</span>
                              <span className="text-xs text-muted-foreground block font-mono">
                                {doc.originalFilename}
                              </span>
                              {doc.description && (
                                <span className="text-[11px] text-muted-foreground line-clamp-1 mt-0.5">
                                  {doc.description}
                                </span>
                              )}
                            </div>
                          </div>
                        </td>

                        <td className="px-4 py-3.5 whitespace-nowrap">
                          <span
                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${getCategoryBadgeClass(
                              doc.category
                            )}`}
                          >
                            {getCategoryLabel(doc.category)}
                          </span>
                        </td>

                        <td className="px-4 py-3.5 whitespace-nowrap">
                          <span className="text-xs font-mono text-muted-foreground bg-muted/60 px-2 py-0.5 rounded">
                            {doc.folderPath}
                          </span>
                        </td>

                        <td className="px-4 py-3.5 whitespace-nowrap">
                          <button
                            onClick={() => handleOpenVersions(doc)}
                            className="inline-flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded bg-blue-50 text-blue-700 dark:bg-blue-950/70 dark:text-blue-300 hover:underline"
                            title={isAr ? "استعراض سجل الإصدارات" : "View version history"}
                          >
                            <History className="h-3 w-3" />v{doc.versionNumber}
                          </button>
                          <span className="text-xs text-muted-foreground block mt-1">
                            {formatBytes(doc.byteSize)}
                          </span>
                        </td>

                        <td className="px-4 py-3.5">
                          <div className="flex items-center gap-1.5">
                            <span className="text-xs font-mono text-muted-foreground bg-muted px-2 py-0.5 rounded max-w-[120px] truncate block">
                              {doc.sha256Hash}
                            </span>
                            <button
                              onClick={() => handleCopyHash(doc.sha256Hash)}
                              className="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted"
                              title={isAr ? "نسخ بصمة التجزئة SHA-256" : "Copy SHA-256 hash"}
                            >
                              {copiedHash === doc.sha256Hash ? (
                                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                              ) : (
                                <Copy className="h-3.5 w-3.5" />
                              )}
                            </button>
                          </div>
                        </td>

                        <td className="px-4 py-3.5 whitespace-nowrap">
                          {doc.attachmentsCount && doc.attachmentsCount > 0 ? (
                            <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded bg-emerald-50 text-emerald-700 dark:bg-emerald-950/70 dark:text-emerald-300">
                              <Paperclip className="h-3 w-3" />
                              {doc.attachmentsCount} {isAr ? "عملية" : "linked"}
                            </span>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </td>

                        <td className="px-4 py-3.5 text-end whitespace-nowrap">
                          <div className="flex items-center justify-end gap-1.5">
                            <a
                              href={`/api/v1/documents/${doc.id}/download?legal_entity_id=00000000-0000-0000-0000-000000000000`}
                              download={doc.originalFilename}
                              className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground transition"
                              title={isAr ? "تحميل الملف" : "Download File"}
                            >
                              <Download className="h-4 w-4" />
                            </a>

                            <button
                              onClick={() => handleOpenAttach(doc)}
                              className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground transition"
                              title={isAr ? "ربط بعملية في الـ ERP" : "Link to ERP entity"}
                            >
                              <Paperclip className="h-4 w-4" />
                            </button>

                            <button
                              onClick={() => handleArchiveDoc(doc.id)}
                              className="p-1.5 rounded-md hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition"
                              title={isAr ? "أرشفة المستند" : "Archive Document"}
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* 5. Tab 2: Entity Attachments */}
      {activeTab === "attachments" && (
        <div className="space-y-4">
          <div className="rounded-xl border bg-card p-4 shadow-sm">
            <h3 className="font-semibold text-sm mb-1">
              {isAr ? "مصفوفة ربط المستندات بالعمليات (Entity Attachments Ledger)" : "Entity Attachments Ledger"}
            </h3>
            <p className="text-xs text-muted-foreground">
              {isAr
                ? "استعراض كافة الوثائق المرفقة والمربوطة بالعمليات المحاسبية، فواتير المبيعات، أوامر الشراء، وسجلات الموظفين"
                : "All documents linked across business transactions, sales invoices, purchase orders, and HR records"}
            </p>
          </div>

          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-start">
                <thead className="bg-muted/50 border-b text-xs text-muted-foreground font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-start">{isAr ? "الكيان المرتبط" : "Entity Type & Ref"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "المستند المرفق" : "Attached Document"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "التصنيف" : "Category"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "اسم الملف والحجم" : "Filename & Size"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "تاريخ الربط" : "Attached At"}</th>
                    <th className="px-4 py-3 text-end">{isAr ? "الإجراء" : "Action"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {attachmentsList.map((att) => (
                    <tr key={att.id} className="hover:bg-muted/30 transition">
                      <td className="px-4 py-3.5 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <span className="px-2 py-0.5 rounded text-xs font-semibold bg-primary/10 text-primary border border-primary/20 uppercase font-mono">
                            {att.entityType}
                          </span>
                          <span className="text-xs font-mono font-bold text-foreground">
                            {att.entityId}
                          </span>
                        </div>
                      </td>

                      <td className="px-4 py-3.5">
                        <span className="font-semibold text-foreground block">{att.fileTitle}</span>
                      </td>

                      <td className="px-4 py-3.5 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${getCategoryBadgeClass(
                            att.category
                          )}`}
                        >
                          {getCategoryLabel(att.category)}
                        </span>
                      </td>

                      <td className="px-4 py-3.5 whitespace-nowrap">
                        <span className="text-xs font-mono text-muted-foreground block">{att.filename}</span>
                        <span className="text-[11px] text-muted-foreground">{formatBytes(att.byteSize)}</span>
                      </td>

                      <td className="px-4 py-3.5 whitespace-nowrap text-xs text-muted-foreground">
                        {new Date(att.createdAt).toLocaleDateString(isAr ? "ar-SA" : "en-US")}
                      </td>

                      <td className="px-4 py-3.5 text-end whitespace-nowrap">
                        <button
                          onClick={() => handleDetach(att.id, att.fileId)}
                          className="px-2.5 py-1 text-xs font-medium rounded text-destructive hover:bg-destructive/10 transition"
                        >
                          {isAr ? "إلغاء الربط" : "Detach"}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* 6. Tab 3: Archived Documents */}
      {activeTab === "archived" && (
        <div className="space-y-4">
          <div className="rounded-xl border bg-card p-4 shadow-sm">
            <h3 className="font-semibold text-sm mb-1">
              {isAr ? "أرشيف المستندات المستبعدة (Archived Documents)" : "Archived Documents Archive"}
            </h3>
            <p className="text-xs text-muted-foreground">
              {isAr
                ? "المستندات المؤرشفة لا تحذف نهائياً لضمان النزاهة والمطابقة النظامية، ويمكن استعادتها بأي وقت"
                : "Archived documents are preserved for regulatory compliance and can be restored at any time"}
            </p>
          </div>

          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-start">
                <thead className="bg-muted/50 border-b text-xs text-muted-foreground font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-start">{isAr ? "عنوان المستند" : "Document"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "التصنيف" : "Category"}</th>
                    <th className="px-4 py-3 text-start">{isAr ? "الحجم" : "Size"}</th>
                    <th className="px-4 py-3 text-end">{isAr ? "الإجراء" : "Action"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredDocs.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground text-sm">
                        {isAr ? "لا توجد مستندات في الأرشيف" : "No archived documents"}
                      </td>
                    </tr>
                  ) : (
                    filteredDocs.map((doc) => (
                      <tr key={doc.id} className="hover:bg-muted/30 transition opacity-80">
                        <td className="px-4 py-3.5">
                          <span className="font-semibold block line-through text-muted-foreground">{doc.title}</span>
                          <span className="text-xs font-mono text-muted-foreground">{doc.originalFilename}</span>
                        </td>
                        <td className="px-4 py-3.5">
                          <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">
                            {getCategoryLabel(doc.category)}
                          </span>
                        </td>
                        <td className="px-4 py-3.5 text-xs text-muted-foreground">{formatBytes(doc.byteSize)}</td>
                        <td className="px-4 py-3.5 text-end">
                          <button
                            onClick={() => handleRestoreDoc(doc.id)}
                            className="inline-flex items-center gap-1 px-3 py-1 rounded bg-primary/10 text-primary text-xs font-medium hover:bg-primary/20 transition"
                          >
                            <RefreshCw className="h-3 w-3" />
                            {isAr ? "استعادة المستند" : "Restore"}
                          </button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* 7. Modal: Upload Document */}
      {isUploadOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-lg shadow-xl overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div className="flex items-center gap-2">
                <UploadCloud className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-base">
                  {isAr ? "رفع مستند جديد وتخزينه آمناً" : "Upload Secure Document"}
                </h3>
              </div>
              <button
                onClick={() => setIsUploadOpen(false)}
                className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleUploadSubmit} className="p-6 space-y-4">
              <div>
                <label className="text-xs font-semibold text-foreground block mb-1.5">
                  {isAr ? "عنوان المستند *" : "Document Title *"}
                </label>
                <input
                  type="text"
                  required
                  placeholder={isAr ? "مثال: عقد توريد برمجيات 2026" : "e.g. Software License Agreement 2026"}
                  value={uploadTitle}
                  onChange={(e) => setUploadTitle(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1.5">
                    {isAr ? "التصنيف *" : "Category *"}
                  </label>
                  <select
                    value={uploadCategory}
                    onChange={(e) => setUploadCategory(e.target.value as any)}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="GENERAL">{isAr ? "عام" : "General"}</option>
                    <option value="LEGAL">{isAr ? "وثائق رسمية وقانونية" : "Legal"}</option>
                    <option value="TAX">{isAr ? "ضرائب وزكاة" : "Tax & ZATCA"}</option>
                    <option value="CONTRACT">{isAr ? "عقود واتفاقيات" : "Contracts"}</option>
                    <option value="INVOICE">{isAr ? "فواتير" : "Invoices"}</option>
                    <option value="RECEIPT">{isAr ? "سندات وإيصالات" : "Receipts"}</option>
                    <option value="IDENTITY">{isAr ? "هويات وثبوتيات" : "Identity"}</option>
                    <option value="OTHER">{isAr ? "أخرى" : "Other"}</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1.5">
                    {isAr ? "المسار / المجلد" : "Folder Path"}
                  </label>
                  <input
                    type="text"
                    placeholder="/contracts"
                    value={uploadFolderPath}
                    onChange={(e) => setUploadFolderPath(e.target.value)}
                    className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary font-mono text-xs"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1.5">
                  {isAr ? "الوصف والملاحظات" : "Description & Notes"}
                </label>
                <textarea
                  rows={2}
                  placeholder={isAr ? "وصف مختصر لمحتوى الوثيقة..." : "Brief description..."}
                  value={uploadDescription}
                  onChange={(e) => setUploadDescription(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>

              {/* File Dropzone */}
              <div>
                <label className="text-xs font-semibold text-foreground block mb-1.5">
                  {isAr ? "الملف الرقمي *" : "Digital File *"}
                </label>
                <div className="border-2 border-dashed border-border rounded-xl p-4 text-center hover:border-primary/50 transition cursor-pointer bg-muted/20">
                  <input
                    type="file"
                    id="file-input"
                    className="hidden"
                    onChange={(e) => {
                      if (e.target.files && e.target.files[0]) {
                        setUploadFile(e.target.files[0]);
                        if (!uploadTitle) {
                          setUploadTitle(e.target.files[0].name.replace(/\.[^/.]+$/, ""));
                        }
                      }
                    }}
                  />
                  <label htmlFor="file-input" className="cursor-pointer block">
                    <UploadCloud className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                    {uploadFile ? (
                      <div>
                        <span className="text-xs font-semibold text-foreground block">{uploadFile.name}</span>
                        <span className="text-[11px] text-muted-foreground">{formatBytes(uploadFile.size)}</span>
                      </div>
                    ) : (
                      <div>
                        <span className="text-xs font-medium text-primary block">
                          {isAr ? "انقر لاختيار ملف أو اسحبه إلى هنا" : "Click to select a file or drag & drop"}
                        </span>
                        <span className="text-[11px] text-muted-foreground block mt-1">
                          PDF, Word, Excel, PNG, JPEG (Up to 50MB)
                        </span>
                      </div>
                    )}
                  </label>
                </div>
              </div>

              {/* Optional Entity Linking */}
              <div className="pt-2 border-t">
                <span className="text-xs font-semibold text-muted-foreground block mb-2">
                  {isAr ? "ربط اختياري بكيان في الـ ERP" : "Optional ERP Entity Link"}
                </span>
                <div className="grid grid-cols-2 gap-3">
                  <select
                    value={uploadEntityType}
                    onChange={(e) => setUploadEntityType(e.target.value)}
                    className="w-full px-3 py-1.5 text-xs rounded-lg border bg-background text-foreground"
                  >
                    <option value="">{isAr ? "— بدون ربط مباشر —" : "— No direct link —"}</option>
                    <option value="sales_invoice">{isAr ? "فاتورة مبيعات (Sales)" : "Sales Invoice"}</option>
                    <option value="purchase_order">{isAr ? "أمر شراء (Purchasing)" : "Purchase Order"}</option>
                    <option value="employee">{isAr ? "سجل موظف (HR)" : "Employee Record"}</option>
                    <option value="supplier">{isAr ? "مورد (Payables)" : "Supplier Record"}</option>
                    <option value="fixed_asset">{isAr ? "أصل ثابت (Fixed Asset)" : "Fixed Asset"}</option>
                  </select>

                  <input
                    type="text"
                    placeholder={isAr ? "رقم أو معرف الكيان" : "Entity Ref / ID"}
                    value={uploadEntityId}
                    onChange={(e) => setUploadEntityId(e.target.value)}
                    disabled={!uploadEntityType}
                    className="w-full px-3 py-1.5 text-xs rounded-lg border bg-background text-foreground disabled:opacity-50 font-mono"
                  />
                </div>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsUploadOpen(false)}
                  className="px-4 py-2 text-xs font-medium rounded-lg border hover:bg-muted transition"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-5 py-2 text-xs font-semibold rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm disabled:opacity-50 flex items-center gap-2"
                >
                  {isSubmitting ? (
                    <>
                      <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                      {isAr ? "جاري الحفظ والتشفير..." : "Encrypting & Storing..."}
                    </>
                  ) : (
                    <>
                      <ShieldCheck className="h-4 w-4" />
                      {isAr ? "حفظ وتخزين آمن" : "Save & Secure"}
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 8. Modal: Version History */}
      {isVersionsOpen && selectedDoc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-2xl shadow-xl overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div className="flex items-center gap-2">
                <History className="h-5 w-5 text-blue-600" />
                <div>
                  <h3 className="font-bold text-base">
                    {isAr ? "سجل إصدارات المستند" : "Document Version History"}
                  </h3>
                  <span className="text-xs text-muted-foreground block">{selectedDoc.title}</span>
                </div>
              </div>
              <button
                onClick={() => setIsVersionsOpen(false)}
                className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="p-6 space-y-4">
              <div className="space-y-3">
                {docVersions.map((v) => (
                  <div
                    key={v.id}
                    className="p-3.5 rounded-xl border bg-muted/20 hover:bg-muted/40 transition flex items-center justify-between gap-4"
                  >
                    <div className="flex items-start gap-3">
                      <div className="px-2.5 py-1 rounded bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300 font-bold text-xs">
                        v{v.versionNumber}
                      </div>
                      <div>
                        <span className="text-xs font-semibold text-foreground block">{v.originalFilename}</span>
                        <div className="flex items-center gap-3 text-[11px] text-muted-foreground mt-0.5">
                          <span>{formatBytes(v.byteSize)}</span>
                          <span>•</span>
                          <span>{new Date(v.createdAt).toLocaleString(isAr ? "ar-SA" : "en-US")}</span>
                          <span>•</span>
                          <span className="font-mono">{v.storageProvider}</span>
                        </div>
                        <div className="flex items-center gap-1 text-[11px] text-muted-foreground mt-1">
                          <span className="font-mono truncate max-w-[240px]">{v.sha256Hash}</span>
                        </div>
                      </div>
                    </div>

                    <a
                      href={`/api/v1/documents/${v.fileId}/download?legal_entity_id=00000000-0000-0000-0000-000000000000&version_id=${v.id}`}
                      download={v.originalFilename}
                      className="px-3 py-1.5 rounded-lg bg-background border text-xs font-medium hover:bg-muted transition flex items-center gap-1.5"
                    >
                      <Download className="h-3.5 w-3.5" />
                      {isAr ? "تحميل" : "Download"}
                    </a>
                  </div>
                ))}
              </div>
            </div>

            <div className="px-6 py-4 bg-muted/40 border-t flex justify-end">
              <button
                onClick={() => setIsVersionsOpen(false)}
                className="px-4 py-2 text-xs font-medium rounded-lg border hover:bg-muted transition"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 9. Modal: Attach to Entity */}
      {isAttachOpen && selectedDoc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-md shadow-xl overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div className="flex items-center gap-2">
                <Paperclip className="h-5 w-5 text-emerald-600" />
                <h3 className="font-bold text-base">
                  {isAr ? "ربط المستند بعملية في الـ ERP" : "Link Document to Entity"}
                </h3>
              </div>
              <button
                onClick={() => setIsAttachOpen(false)}
                className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleAttachSubmit} className="p-6 space-y-4">
              <div className="p-3 rounded-lg bg-muted/30 border text-xs">
                <span className="font-semibold block text-foreground">{selectedDoc.title}</span>
                <span className="text-muted-foreground font-mono">{selectedDoc.originalFilename}</span>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1.5">
                  {isAr ? "نوع الكيان المستهدف *" : "Target Entity Type *"}
                </label>
                <select
                  value={attachEntityType}
                  onChange={(e) => setAttachEntityType(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="sales_invoice">{isAr ? "فاتورة مبيعات (Sales Invoice)" : "Sales Invoice"}</option>
                  <option value="purchase_order">{isAr ? "أمر شراء (Purchase Order)" : "Purchase Order"}</option>
                  <option value="employee">{isAr ? "ملف موظف (HR Employee)" : "HR Employee"}</option>
                  <option value="supplier">{isAr ? "مورد (AP Supplier)" : "AP Supplier"}</option>
                  <option value="customer">{isAr ? "عميل (AR Customer)" : "AR Customer"}</option>
                  <option value="fixed_asset">{isAr ? "أصل ثابت (Fixed Asset)" : "Fixed Asset"}</option>
                  <option value="journal_entry">{isAr ? "قيد يومية (Journal Entry)" : "Journal Entry"}</option>
                </select>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1.5">
                  {isAr ? "رقم أو معرف الكيان *" : "Entity Reference / ID *"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. INV-2026-0045"
                  value={attachEntityId}
                  onChange={(e) => setAttachEntityId(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary font-mono"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsAttachOpen(false)}
                  className="px-4 py-2 text-xs font-medium rounded-lg border hover:bg-muted transition"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 text-xs font-semibold rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 transition shadow-sm"
                >
                  {isAr ? "تأكيد الربط" : "Attach"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

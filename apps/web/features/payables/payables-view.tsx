"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Receipt,
  FileText,
  CreditCard,
  Clock,
  Coins,
  CheckCircle2,
  AlertCircle,
  Plus,
  Search,
  Landmark,
  X,
  TrendingDown,
  Building2,
  Calendar,
  Layers,
  ArrowDownLeft,
  Check,
} from "lucide-react";

interface SupplierInvoiceItem {
  id: string;
  invoiceNumber: string;
  supplierInvoiceNumber: string;
  supplierId: string;
  supplierNameAr: string;
  supplierNameEn: string;
  invoiceDate: string;
  dueDate: string;
  currencyCode: string;
  subtotal: number;
  taxAmount: number;
  totalAmount: number;
  status: "DRAFT" | "POSTED" | "PAID" | "CANCELLED";
  paymentStatus: "UNPAID" | "PARTIALLY_PAID" | "PAID";
}

interface SupplierPaymentItem {
  id: string;
  paymentNumber: string;
  supplierId: string;
  supplierNameAr: string;
  supplierNameEn: string;
  paymentDate: string;
  paymentMethod: "BANK_TRANSFER" | "CASH" | "CHEQUE" | "MADA";
  disbursementAccount: string;
  amount: number;
  allocatedAmount: number;
  unallocatedAmount: number;
  referenceNumber?: string;
  status: "POSTED" | "DRAFT";
}

interface OpenItemEntity {
  id: string;
  supplierNumber: string;
  supplierNameAr: string;
  documentType: "INVOICE" | "CREDIT_NOTE";
  documentNumber: string;
  documentDate: string;
  dueDate: string;
  originalAmount: number;
  remainingAmount: number;
  daysOverdue: number;
  status: "OPEN" | "PARTIALLY_PAID" | "PAID";
}

interface CreditNoteEntity {
  id: string;
  creditNoteNumber: string;
  supplierNameAr: string;
  date: string;
  subtotal: number;
  taxAmount: number;
  totalAmount: number;
  reason: string;
  status: "POSTED";
}

const INITIAL_INVOICES: SupplierInvoiceItem[] = [
  {
    id: "inv-1",
    invoiceNumber: "PINV-2026-001",
    supplierInvoiceNumber: "ALF-99882",
    supplierId: "sup-1",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    supplierNameEn: "Al-Fozan Building Materials Co.",
    invoiceDate: "2026-09-01",
    dueDate: "2026-10-01",
    currencyCode: "SAR",
    subtotal: 20000.0,
    taxAmount: 3000.0,
    totalAmount: 23000.0,
    status: "POSTED",
    paymentStatus: "PARTIALLY_PAID",
  },
  {
    id: "inv-2",
    invoiceNumber: "PINV-2026-002",
    supplierInvoiceNumber: "SEC-BILL-881",
    supplierId: "sup-2",
    supplierNameAr: "الشركة السعودية للكهرباء (SEC)",
    supplierNameEn: "Saudi Electricity Company",
    invoiceDate: "2026-09-10",
    dueDate: "2026-09-25",
    currencyCode: "SAR",
    subtotal: 11174.35,
    taxAmount: 1676.15,
    totalAmount: 12850.5,
    status: "POSTED",
    paymentStatus: "UNPAID",
  },
  {
    id: "inv-3",
    invoiceNumber: "PINV-2026-003",
    supplierInvoiceNumber: "DELL-INV-5510",
    supplierId: "sup-3",
    supplierNameAr: "شركة ديل العالمية للتقنية",
    supplierNameEn: "Dell Technologies EMEA",
    invoiceDate: "2026-08-15",
    dueDate: "2026-09-30",
    currencyCode: "SAR",
    subtotal: 160869.57,
    taxAmount: 24130.43,
    totalAmount: 185000.0,
    status: "POSTED",
    paymentStatus: "UNPAID",
  },
  {
    id: "inv-4",
    invoiceNumber: "PINV-2026-004",
    supplierInvoiceNumber: "RAW-4401",
    supplierId: "sup-1",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    supplierNameEn: "Al-Fozan Building Materials Co.",
    invoiceDate: "2026-09-15",
    dueDate: "2026-10-15",
    currencyCode: "SAR",
    subtotal: 5000.0,
    taxAmount: 750.0,
    totalAmount: 5750.0,
    status: "DRAFT",
    paymentStatus: "UNPAID",
  },
];

const INITIAL_PAYMENTS: SupplierPaymentItem[] = [
  {
    id: "pay-1",
    paymentNumber: "PAY-2026-0001",
    supplierId: "sup-1",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    supplierNameEn: "Al-Fozan Building Materials Co.",
    paymentDate: "2026-09-08",
    paymentMethod: "BANK_TRANSFER",
    disbursementAccount: "111002 - مصرف الراجحي (SAR)",
    amount: 10000.0,
    allocatedAmount: 10000.0,
    unallocatedAmount: 0.0,
    referenceNumber: "TX-SNB-998811",
    status: "POSTED",
  },
];

const INITIAL_OPEN_ITEMS: OpenItemEntity[] = [
  {
    id: "oi-1",
    supplierNumber: "SUP-0001",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    documentType: "INVOICE",
    documentNumber: "PINV-2026-001",
    documentDate: "2026-09-01",
    dueDate: "2026-10-01",
    originalAmount: 23000.0,
    remainingAmount: 13000.0,
    daysOverdue: 0,
    status: "PARTIALLY_PAID",
  },
  {
    id: "oi-2",
    supplierNumber: "SUP-0002",
    supplierNameAr: "الشركة السعودية للكهرباء (SEC)",
    documentType: "INVOICE",
    documentNumber: "PINV-2026-002",
    documentDate: "2026-09-10",
    dueDate: "2026-09-25",
    originalAmount: 12850.5,
    remainingAmount: 12850.5,
    daysOverdue: 0,
    status: "OPEN",
  },
  {
    id: "oi-3",
    supplierNumber: "SUP-0003",
    supplierNameAr: "شركة ديل العالمية للتقنية",
    documentType: "INVOICE",
    documentNumber: "PINV-2026-003",
    documentDate: "2026-08-15",
    dueDate: "2026-09-30",
    originalAmount: 185000.0,
    remainingAmount: 185000.0,
    daysOverdue: 0,
    status: "OPEN",
  },
];

const INITIAL_CREDIT_NOTES: CreditNoteEntity[] = [
  {
    id: "cn-1",
    creditNoteNumber: "PCN-2026-001",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    date: "2026-09-12",
    subtotal: 1000.0,
    taxAmount: 150.0,
    totalAmount: 1150.0,
    reason: "مردودات مواد بناء تالفة بموجب إرسالية فحص المخزن",
    status: "POSTED",
  },
];

export function PayablesView() {
  const { locale, t } = useLocale();
  const isAr = locale === "ar";
  const pT = (t as any).payables || {};

  const [activeTab, setActiveTab] = useState<"invoices" | "payments" | "open_items" | "aging" | "credit_notes">("invoices");
  const [invoices, setInvoices] = useState<SupplierInvoiceItem[]>(INITIAL_INVOICES);
  const [payments, setPayments] = useState<SupplierPaymentItem[]>(INITIAL_PAYMENTS);
  const [openItems, setOpenItems] = useState<OpenItemEntity[]>(INITIAL_OPEN_ITEMS);
  const [creditNotes, setCreditNotes] = useState<CreditNoteEntity[]>(INITIAL_CREDIT_NOTES);

  const [search, setSearch] = useState("");
  const [isInvoiceModalOpen, setIsInvoiceModalOpen] = useState(false);
  const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);

  // New Invoice Form
  const [newInv, setNewInv] = useState({
    supplierNameAr: "شركة الفوزان لمواد البناء",
    supplierInvoiceNumber: "",
    invoiceDate: new Date().toISOString().slice(0, 10),
    dueDate: new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10),
    lineDescription: "مشتريات مواد ومهمات تشغيلية",
    quantity: 1,
    unitPrice: 1000,
    taxRate: 15,
  });

  // New Payment Form
  const [newPay, setNewPay] = useState({
    supplierNameAr: "شركة الفوزان لمواد البناء",
    paymentMethod: "BANK_TRANSFER" as "BANK_TRANSFER" | "CASH" | "CHEQUE" | "MADA",
    disbursementAccount: "111002 - مصرف الراجحي (SAR)",
    amount: 5000,
    referenceNumber: "",
  });

  const totalInvoicesCount = invoices.length;
  const totalOutstanding = openItems.reduce((acc, oi) => acc + oi.remainingAmount, 0);
  const totalPaymentsAmount = payments.reduce((acc, p) => acc + p.amount, 0);
  const overduePayables = openItems.filter((oi) => oi.daysOverdue > 0).reduce((acc, oi) => acc + oi.remainingAmount, 0);

  const handlePostInvoice = (id: string) => {
    setInvoices(
      invoices.map((inv) => (inv.id === id ? { ...inv, status: "POSTED" } : inv))
    );
  };

  const handleCreateInvoice = (e: React.FormEvent) => {
    e.preventDefault();
    const subtotal = newInv.quantity * newInv.unitPrice;
    const taxAmount = (subtotal * newInv.taxRate) / 100;
    const totalAmount = subtotal + taxAmount;

    const inv: SupplierInvoiceItem = {
      id: `inv-${Date.now()}`,
      invoiceNumber: `PINV-2026-${String(invoices.length + 1).padStart(3, "0")}`,
      supplierInvoiceNumber: newInv.supplierInvoiceNumber || `BILL-${Date.now().toString().slice(-4)}`,
      supplierId: "sup-1",
      supplierNameAr: newInv.supplierNameAr,
      supplierNameEn: "Al-Fozan Building Materials Co.",
      invoiceDate: newInv.invoiceDate,
      dueDate: newInv.dueDate,
      currencyCode: "SAR",
      subtotal,
      taxAmount,
      totalAmount,
      status: "POSTED",
      paymentStatus: "UNPAID",
    };

    setInvoices([inv, ...invoices]);
    setOpenItems([
      {
        id: `oi-${Date.now()}`,
        supplierNumber: "SUP-0001",
        supplierNameAr: newInv.supplierNameAr,
        documentType: "INVOICE",
        documentNumber: inv.invoiceNumber,
        documentDate: inv.invoiceDate,
        dueDate: inv.dueDate,
        originalAmount: totalAmount,
        remainingAmount: totalAmount,
        daysOverdue: 0,
        status: "OPEN",
      },
      ...openItems,
    ]);

    setIsInvoiceModalOpen(false);
  };

  const handleCreatePayment = (e: React.FormEvent) => {
    e.preventDefault();
    const pay: SupplierPaymentItem = {
      id: `pay-${Date.now()}`,
      paymentNumber: `PAY-2026-${String(payments.length + 1).padStart(4, "0")}`,
      supplierId: "sup-1",
      supplierNameAr: newPay.supplierNameAr,
      supplierNameEn: "Al-Fozan Building Materials Co.",
      paymentDate: new Date().toISOString().slice(0, 10),
      paymentMethod: newPay.paymentMethod,
      disbursementAccount: newPay.disbursementAccount,
      amount: Number(newPay.amount),
      allocatedAmount: Number(newPay.amount),
      unallocatedAmount: 0,
      referenceNumber: newPay.referenceNumber,
      status: "POSTED",
    };

    setPayments([pay, ...payments]);
    setIsPaymentModalOpen(false);
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <Receipt className="h-7 w-7 text-primary" />
            {pT.title || (isAr ? "حسابات الدائنين وفواتير المشتريات وسندات الصرف" : "Accounts Payable & Disbursements")}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {pT.subtitle ||
              (isAr
                ? "إدارة فواتير الموردين، القيود اليومية للأستاذ العام، سندات الصرف، ومحرك أعمار الديون"
                : "Manage vendor invoices, double-entry GL postings, payment allocations, and AP aging")}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsInvoiceModalOpen(true)}
            className="inline-flex items-center gap-2 px-3 py-2 rounded-lg border bg-card text-foreground font-medium text-xs hover:bg-muted transition shadow-sm"
          >
            <Plus className="h-4 w-4" />
            {pT.new_invoice || (isAr ? "تسجيل فاتورة مورد" : "New Supplier Invoice")}
          </button>
          <button
            onClick={() => setIsPaymentModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground font-medium text-xs hover:bg-primary/90 transition shadow-sm"
          >
            <ArrowDownLeft className="h-4 w-4" />
            {pT.new_payment || (isAr ? "إصدار سند صرف" : "New Disbursement")}
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {pT.kpi_invoices || (isAr ? "إجمالي فواتير الموردين" : "Total Invoices")}
            </span>
            <FileText className="h-4 w-4 text-primary" />
          </div>
          <div className="text-2xl font-bold mt-2">{totalInvoicesCount}</div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "مسجلة في دفتر المشتريات" : "Recorded in AP ledger"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {pT.kpi_outstanding || (isAr ? "الرصيد المستحق للدائنين" : "Outstanding AP")}
            </span>
            <Coins className="h-4 w-4 text-amber-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-amber-600 dark:text-amber-400">
            {totalOutstanding.toLocaleString(isAr ? "ar-SA" : "en-US", {
              style: "currency",
              currency: "SAR",
            })}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "مستحقة السداد للذمم الدائنة" : "Due to suppliers"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {pT.kpi_payments || (isAr ? "سندات الصرف المعتمدة" : "Processed Payments")}
            </span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-emerald-600 dark:text-emerald-400">
            {totalPaymentsAmount.toLocaleString(isAr ? "ar-SA" : "en-US", {
              style: "currency",
              currency: "SAR",
            })}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "مدفوعة عبر الحسابات البنكية" : "Disbursed via bank"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {pT.kpi_overdue || (isAr ? "المستحقات المتأخرة" : "Overdue Payables")}
            </span>
            <Clock className="h-4 w-4 text-rose-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-rose-600 dark:text-rose-400">
            {overduePayables.toLocaleString(isAr ? "ar-SA" : "en-US", {
              style: "currency",
              currency: "SAR",
            })}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "تجاوزت تاريخ الاستحقاق" : "Past due invoices"}
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b flex items-center gap-1 overflow-x-auto text-xs">
        <button
          onClick={() => setActiveTab("invoices")}
          className={`px-4 py-2.5 font-medium border-b-2 transition whitespace-nowrap ${
            activeTab === "invoices"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          {pT.tab_invoices || (isAr ? "فواتير الموردين (Invoices)" : "Invoices")}
        </button>
        <button
          onClick={() => setActiveTab("payments")}
          className={`px-4 py-2.5 font-medium border-b-2 transition whitespace-nowrap ${
            activeTab === "payments"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          {pT.tab_payments || (isAr ? "سندات الصرف (Payments)" : "Payments")}
        </button>
        <button
          onClick={() => setActiveTab("open_items")}
          className={`px-4 py-2.5 font-medium border-b-2 transition whitespace-nowrap ${
            activeTab === "open_items"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          {pT.tab_open_items || (isAr ? "الفواتير المفتوحة (Open Items)" : "Open Items")}
        </button>
        <button
          onClick={() => setActiveTab("aging")}
          className={`px-4 py-2.5 font-medium border-b-2 transition whitespace-nowrap ${
            activeTab === "aging"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          {pT.tab_aging || (isAr ? "تحليل أعمار الديون (AP Aging)" : "AP Aging")}
        </button>
        <button
          onClick={() => setActiveTab("credit_notes")}
          className={`px-4 py-2.5 font-medium border-b-2 transition whitespace-nowrap ${
            activeTab === "credit_notes"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          {pT.tab_credit_notes || (isAr ? "إشعارات الخصم (Credit Notes)" : "Credit Notes")}
        </button>
      </div>

      {/* Tab 1: Supplier Invoices */}
      {activeTab === "invoices" && (
        <div className="space-y-4">
          <div className="border rounded-xl bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground">
                  <tr>
                    <th className="py-3 px-4 text-start font-medium">{pT.invoice_number || (isAr ? "رقم الفاتورة" : "Invoice #")}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.supplier_invoice_number || (isAr ? "فاتورة المورد" : "Vendor Bill #")}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.supplier || (isAr ? "المورد" : "Supplier")}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.invoice_date || (isAr ? "تاريخ الفاتورة" : "Date")}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.due_date || (isAr ? "الاستحقاق" : "Due Date")}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.subtotal || (isAr ? "المبلغ الخاضع" : "Subtotal")}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.tax_amount || (isAr ? "الضريبة 15%" : "VAT 15%")}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.total_amount || (isAr ? "الإجمالي" : "Total")}</th>
                    <th className="py-3 px-4 text-center font-medium">{isAr ? "الحالة المحاسبية" : "GL Status"}</th>
                    <th className="py-3 px-4 text-center font-medium">{pT.payment_status || (isAr ? "حالة السداد" : "Payment")}</th>
                    <th className="py-3 px-4 text-center font-medium">{isAr ? "الإجراء" : "Action"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {invoices.map((inv) => (
                    <tr key={inv.id} className="hover:bg-muted/30 transition">
                      <td className="py-3 px-4 font-mono font-medium text-foreground">{inv.invoiceNumber}</td>
                      <td className="py-3 px-4 font-mono text-muted-foreground">{inv.supplierInvoiceNumber}</td>
                      <td className="py-3 px-4 font-semibold text-foreground">{isAr ? inv.supplierNameAr : inv.supplierNameEn}</td>
                      <td className="py-3 px-4 text-muted-foreground">{inv.invoiceDate}</td>
                      <td className="py-3 px-4 text-muted-foreground">{inv.dueDate}</td>
                      <td className="py-3 px-4 text-end font-mono">
                        {inv.subtotal.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-3 px-4 text-end font-mono text-primary font-medium">
                        {inv.taxAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-foreground">
                        {inv.totalAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} SAR
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium ${
                            inv.status === "POSTED"
                              ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
                              : "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300"
                          }`}
                        >
                          {inv.status === "POSTED" ? (isAr ? "مرحل للأستاذ العام" : "Posted to GL") : isAr ? "مسودة" : "Draft"}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium ${
                            inv.paymentStatus === "PAID"
                              ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
                              : inv.paymentStatus === "PARTIALLY_PAID"
                              ? "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300"
                              : "bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
                          }`}
                        >
                          {inv.paymentStatus === "PAID"
                            ? isAr ? "مسدد بالكامل" : "Paid"
                            : inv.paymentStatus === "PARTIALLY_PAID"
                            ? isAr ? "مسدد جزئياً" : "Partial"
                            : isAr ? "غير مسدد" : "Unpaid"}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-center">
                        {inv.status === "DRAFT" ? (
                          <button
                            onClick={() => handlePostInvoice(inv.id)}
                            className="px-2.5 py-1 rounded bg-primary text-primary-foreground text-[10px] hover:bg-primary/90 font-medium"
                          >
                            {pT.post_invoice || (isAr ? "ترحيل" : "Post")}
                          </button>
                        ) : (
                          <span className="text-[10px] text-muted-foreground">{isAr ? "معتمد" : "Locked"}</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Supplier Payments */}
      {activeTab === "payments" && (
        <div className="space-y-4">
          <div className="border rounded-xl bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground">
                  <tr>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "رقم السند" : "Payment #"}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.supplier || (isAr ? "المورد" : "Supplier")}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "تاريخ الصرف" : "Date"}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.payment_method || (isAr ? "وسيلة الدفع" : "Method")}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.disbursement_account || (isAr ? "حساب الصرف" : "Account")}</th>
                    <th className="py-3 px-4 text-end font-medium">{isAr ? "المبلغ المصروف" : "Amount"}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.allocated_amount || (isAr ? "الموزع على الفواتير" : "Allocated")}</th>
                    <th className="py-3 px-4 text-center font-medium">{isAr ? "الحالة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {payments.map((p) => (
                    <tr key={p.id} className="hover:bg-muted/30 transition">
                      <td className="py-3 px-4 font-mono font-medium text-foreground">{p.paymentNumber}</td>
                      <td className="py-3 px-4 font-semibold text-foreground">{isAr ? p.supplierNameAr : p.supplierNameEn}</td>
                      <td className="py-3 px-4 text-muted-foreground">{p.paymentDate}</td>
                      <td className="py-3 px-4">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-secondary border">
                          {p.paymentMethod}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-muted-foreground font-mono">{p.disbursementAccount}</td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-foreground">
                        {p.amount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} SAR
                      </td>
                      <td className="py-3 px-4 text-end font-mono text-emerald-600 dark:text-emerald-400 font-medium">
                        {p.allocatedAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} SAR
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-medium">
                          {isAr ? "مرحل محاسبياً" : "Posted"}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: Open Items */}
      {activeTab === "open_items" && (
        <div className="space-y-4">
          <div className="border rounded-xl bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground">
                  <tr>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "رقم المورد" : "Supplier #"}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.supplier || (isAr ? "المورد" : "Supplier")}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "نوع الحركة" : "Type"}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "رقم المستند" : "Doc #"}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "تاريخ الاستحقاق" : "Due Date"}</th>
                    <th className="py-3 px-4 text-end font-medium">{isAr ? "المبلغ الأصلي" : "Original"}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.remaining_amount || (isAr ? "الرصيد المتبقي" : "Remaining")}</th>
                    <th className="py-3 px-4 text-center font-medium">{isAr ? "حالة الحركة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {openItems.map((oi) => (
                    <tr key={oi.id} className="hover:bg-muted/30 transition">
                      <td className="py-3 px-4 font-mono">{oi.supplierNumber}</td>
                      <td className="py-3 px-4 font-semibold text-foreground">{oi.supplierNameAr}</td>
                      <td className="py-3 px-4">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-secondary border">
                          {oi.documentType}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono font-medium text-foreground">{oi.documentNumber}</td>
                      <td className="py-3 px-4 text-muted-foreground">{oi.dueDate}</td>
                      <td className="py-3 px-4 text-end font-mono">
                        {oi.originalAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-amber-600 dark:text-amber-400">
                        {oi.remainingAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} SAR
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium ${
                            oi.status === "OPEN"
                              ? "bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
                              : "bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300"
                          }`}
                        >
                          {oi.status === "OPEN" ? (isAr ? "مفتوح بالكامل" : "Open") : isAr ? "مسدد جزئياً" : "Partial"}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: AP Aging */}
      {activeTab === "aging" && (
        <div className="space-y-4">
          <div className="p-4 rounded-xl border bg-card space-y-4 shadow-sm">
            <h3 className="text-sm font-bold text-foreground">
              {isAr ? "توزيع أعمار الديون المستحقة للموردين (AP Aging Buckets)" : "AP Aging Buckets Distribution"}
            </h3>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
              <div className="p-3 rounded-lg border bg-background text-center">
                <span className="text-[11px] text-muted-foreground block">{isAr ? "ديون جارية (غير مستحقة)" : "Current (Not Due)"}</span>
                <span className="text-lg font-bold text-emerald-600 dark:text-emerald-400 mt-1 block">
                  {totalOutstanding.toLocaleString(isAr ? "ar-SA" : "en-US", { style: "currency", currency: "SAR" })}
                </span>
              </div>
              <div className="p-3 rounded-lg border bg-background text-center">
                <span className="text-[11px] text-muted-foreground block">{isAr ? "متأخرات 1-30 يوماً" : "1-30 Days Overdue"}</span>
                <span className="text-lg font-bold text-sky-600 dark:text-sky-400 mt-1 block">0.00 SAR</span>
              </div>
              <div className="p-3 rounded-lg border bg-background text-center">
                <span className="text-[11px] text-muted-foreground block">{isAr ? "متأخرات 31-60 يوماً" : "31-60 Days Overdue"}</span>
                <span className="text-lg font-bold text-amber-600 dark:text-amber-400 mt-1 block">0.00 SAR</span>
              </div>
              <div className="p-3 rounded-lg border bg-background text-center">
                <span className="text-[11px] text-muted-foreground block">{isAr ? "متأخرات 61-90 يوماً" : "61-90 Days Overdue"}</span>
                <span className="text-lg font-bold text-orange-600 dark:text-orange-400 mt-1 block">0.00 SAR</span>
              </div>
              <div className="p-3 rounded-lg border bg-background text-center">
                <span className="text-[11px] text-muted-foreground block">{isAr ? "متأخرات > 90 يوماً" : "> 90 Days Overdue"}</span>
                <span className="text-lg font-bold text-rose-600 dark:text-rose-400 mt-1 block">0.00 SAR</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 5: Credit Notes */}
      {activeTab === "credit_notes" && (
        <div className="space-y-4">
          <div className="border rounded-xl bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground">
                  <tr>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "رقم الإشعار" : "Credit Note #"}</th>
                    <th className="py-3 px-4 text-start font-medium">{pT.supplier || (isAr ? "المورد" : "Supplier")}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "التاريخ" : "Date"}</th>
                    <th className="py-3 px-4 text-start font-medium">{isAr ? "سبب الإشعار ومردود المشتريات" : "Reason"}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.subtotal || (isAr ? "المبلغ" : "Subtotal")}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.tax_amount || (isAr ? "الضريبة 15%" : "VAT")}</th>
                    <th className="py-3 px-4 text-end font-medium">{pT.total_amount || (isAr ? "الإجمالي" : "Total")}</th>
                    <th className="py-3 px-4 text-center font-medium">{isAr ? "الحالة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {creditNotes.map((cn) => (
                    <tr key={cn.id} className="hover:bg-muted/30 transition">
                      <td className="py-3 px-4 font-mono font-medium text-foreground">{cn.creditNoteNumber}</td>
                      <td className="py-3 px-4 font-semibold text-foreground">{cn.supplierNameAr}</td>
                      <td className="py-3 px-4 text-muted-foreground">{cn.date}</td>
                      <td className="py-3 px-4 text-muted-foreground">{cn.reason}</td>
                      <td className="py-3 px-4 text-end font-mono">
                        {cn.subtotal.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-3 px-4 text-end font-mono text-primary font-medium">
                        {cn.taxAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-foreground">
                        {cn.totalAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} SAR
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-medium">
                          {isAr ? "مرحل للأستاذ العام" : "Posted"}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Modal: New Supplier Invoice */}
      {isInvoiceModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-lg shadow-2xl">
            <div className="flex items-center justify-between p-5 border-b sticky top-0 bg-card z-10">
              <h2 className="text-base font-bold text-foreground flex items-center gap-2">
                <FileText className="h-5 w-5 text-primary" />
                {pT.new_invoice || (isAr ? "تسجيل فاتورة مورد جديدة" : "New Supplier Invoice")}
              </h2>
              <button
                onClick={() => setIsInvoiceModalOpen(false)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreateInvoice} className="p-5 space-y-3">
              <div>
                <label className="block text-xs font-medium mb-1">{pT.supplier || (isAr ? "المورد" : "Supplier")}</label>
                <select
                  value={newInv.supplierNameAr}
                  onChange={(e) => setNewInv({ ...newInv, supplierNameAr: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="شركة الفوزان لمواد البناء">شركة الفوزان لمواد البناء (SUP-0001)</option>
                  <option value="الشركة السعودية للكهرباء (SEC)">الشركة السعودية للكهرباء (SEC) (SUP-0002)</option>
                  <option value="شركة ديل العالمية للتقنية">شركة ديل العالمية للتقنية (SUP-0003)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">
                  {pT.supplier_invoice_number || (isAr ? "رقم فاتورة المورد الأصلية *" : "Vendor Bill # *")}
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. ALF-99120"
                  value={newInv.supplierInvoiceNumber}
                  onChange={(e) => setNewInv({ ...newInv, supplierInvoiceNumber: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{pT.invoice_date || (isAr ? "تاريخ الفاتورة" : "Date")}</label>
                  <input
                    type="date"
                    required
                    value={newInv.invoiceDate}
                    onChange={(e) => setNewInv({ ...newInv, invoiceDate: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{pT.due_date || (isAr ? "تاريخ الاستحقاق" : "Due Date")}</label>
                  <input
                    type="date"
                    required
                    value={newInv.dueDate}
                    onChange={(e) => setNewInv({ ...newInv, dueDate: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">{isAr ? "بيان البنود والمصروف" : "Line Description"}</label>
                <input
                  type="text"
                  required
                  value={newInv.lineDescription}
                  onChange={(e) => setNewInv({ ...newInv, lineDescription: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="grid grid-cols-3 gap-2">
                <div>
                  <label className="block text-xs font-medium mb-1">{isAr ? "الكمية" : "Qty"}</label>
                  <input
                    type="number"
                    min="1"
                    required
                    value={newInv.quantity}
                    onChange={(e) => setNewInv({ ...newInv, quantity: Number(e.target.value) })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{isAr ? "سعر الوحدة" : "Unit Price"}</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    required
                    value={newInv.unitPrice}
                    onChange={(e) => setNewInv({ ...newInv, unitPrice: Number(e.target.value) })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{isAr ? "الضريبة %" : "VAT %"}</label>
                  <input
                    type="number"
                    readOnly
                    value={newInv.taxRate}
                    className="w-full bg-muted border rounded-lg px-3 py-2 text-xs font-mono text-muted-foreground"
                  />
                </div>
              </div>

              <div className="p-3 rounded-lg bg-muted/40 text-xs space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "المبلغ الخاضع:" : "Subtotal:"}</span>
                  <span className="font-mono font-medium">{(newInv.quantity * newInv.unitPrice).toFixed(2)} SAR</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "ضريبة المدخلات (15%):" : "VAT (15%):"}</span>
                  <span className="font-mono font-medium">{((newInv.quantity * newInv.unitPrice * 15) / 100).toFixed(2)} SAR</span>
                </div>
                <div className="flex justify-between border-t pt-1 font-bold text-foreground">
                  <span>{isAr ? "الإجمالي المستحق:" : "Total Amount:"}</span>
                  <span className="font-mono text-primary">
                    {(newInv.quantity * newInv.unitPrice * 1.15).toFixed(2)} SAR
                  </span>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsInvoiceModalOpen(false)}
                  className="px-4 py-2 rounded-lg border text-xs font-medium hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90"
                >
                  {isAr ? "حفظ واعتماد الفاتورة" : "Save & Post Invoice"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: New Supplier Payment */}
      {isPaymentModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-lg shadow-2xl">
            <div className="flex items-center justify-between p-5 border-b sticky top-0 bg-card z-10">
              <h2 className="text-base font-bold text-foreground flex items-center gap-2">
                <ArrowDownLeft className="h-5 w-5 text-primary" />
                {pT.new_payment || (isAr ? "إصدار سند صرف للمورد" : "New Supplier Disbursement")}
              </h2>
              <button
                onClick={() => setIsPaymentModalOpen(false)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreatePayment} className="p-5 space-y-3">
              <div>
                <label className="block text-xs font-medium mb-1">{pT.supplier || (isAr ? "المورد" : "Supplier")}</label>
                <select
                  value={newPay.supplierNameAr}
                  onChange={(e) => setNewPay({ ...newPay, supplierNameAr: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="شركة الفوزان لمواد البناء">شركة الفوزان لمواد البناء (SUP-0001)</option>
                  <option value="الشركة السعودية للكهرباء (SEC)">الشركة السعودية للكهرباء (SEC) (SUP-0002)</option>
                  <option value="شركة ديل العالمية للتقنية">شركة ديل العالمية للتقنية (SUP-0003)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">{pT.disbursement_account || (isAr ? "حساب الصرف / البنك" : "Bank Account")}</label>
                <select
                  value={newPay.disbursementAccount}
                  onChange={(e) => setNewPay({ ...newPay, disbursementAccount: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="111002 - مصرف الراجحي (SAR)">111002 - مصرف الراجحي - الحساب الجاري (SAR)</option>
                  <option value="111003 - البنك الأهلي السعودي (SAR)">111003 - البنك الأهلي السعودي (SAR)</option>
                  <option value="111001 - العهدة النقدية والصندوق">111001 - العهدة النقدية والصندوق الرئيسي</option>
                </select>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1">{pT.payment_method || (isAr ? "وسيلة الدفع" : "Method")}</label>
                  <select
                    value={newPay.paymentMethod}
                    onChange={(e) => setNewPay({ ...newPay, paymentMethod: e.target.value as any })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="BANK_TRANSFER">{isAr ? "تحويل بنكي (سريع)" : "Bank Transfer"}</option>
                    <option value="CHEQUE">{isAr ? "شيك بنكي" : "Cheque"}</option>
                    <option value="MADA">{isAr ? "بطاقة مدى" : "Mada"}</option>
                    <option value="CASH">{isAr ? "نقداً من الصندوق" : "Cash"}</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">{isAr ? "مبلغ السداد (SAR) *" : "Payment Amount *"}</label>
                  <input
                    type="number"
                    min="1"
                    step="0.01"
                    required
                    value={newPay.amount}
                    onChange={(e) => setNewPay({ ...newPay, amount: Number(e.target.value) })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono font-bold focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">{isAr ? "الرقم المرجعي للحوالة / الشيك" : "Transfer / Reference #"}</label>
                <input
                  type="text"
                  placeholder="e.g. TX-SAR-110022"
                  value={newPay.referenceNumber}
                  onChange={(e) => setNewPay({ ...newPay, referenceNumber: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsPaymentModalOpen(false)}
                  className="px-4 py-2 rounded-lg border text-xs font-medium hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90"
                >
                  {pT.post_payment || (isAr ? "ترحيل السند وتوزيعه" : "Post & Allocate Payment")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

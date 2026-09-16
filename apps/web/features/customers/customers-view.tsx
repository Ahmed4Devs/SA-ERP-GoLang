"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Users,
  Building2,
  Receipt,
  FileText,
  Clock,
  Coins,
  ShieldCheck,
  Plus,
  Search,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  ArrowUpRight,
  TrendingUp,
  Landmark,
  X,
  CreditCard,
  SlidersHorizontal,
  Calendar,
  Check,
  Percent,
} from "lucide-react";

interface CustomerItem {
  id: string;
  customerNumber: string;
  customerType: "B2B" | "B2C" | "GOVERNMENT";
  nameAr: string;
  nameEn: string;
  taxId?: string;
  crNumber?: string;
  city: string;
  district?: string;
  phone?: string;
  email?: string;
  paymentTermsCode: string;
  paymentTermsDueDays: number;
  creditLimit: number;
  currentBalance: number;
  isCreditHold: boolean;
  status: "ACTIVE" | "SUSPENDED" | "BLOCKED";
}

interface ReceiptItem {
  id: string;
  receiptNumber: string;
  customerNumber: string;
  customerNameAr: string;
  customerNameEn: string;
  receiptDate: string;
  paymentMethod: "BANK_TRANSFER" | "CASH" | "CHEQUE" | "CREDIT_CARD" | "MADA" | "SADAD";
  depositAccount: string;
  amount: number;
  allocatedAmount: number;
  unallocatedAmount: number;
  referenceNumber?: string;
  status: "POSTED" | "DRAFT" | "CANCELLED";
}

interface OpenItemEntity {
  id: string;
  documentType: "INVOICE" | "DEBIT_NOTE" | "CREDIT_NOTE";
  documentNumber: string;
  customerNumber: string;
  customerNameAr: string;
  documentDate: string;
  dueDate: string;
  originalAmount: number;
  allocatedAmount: number;
  remainingAmount: number;
  daysOverdue: number;
  status: "OPEN" | "PARTIALLY_PAID" | "PAID";
}

interface PaymentTermItem {
  id: string;
  code: string;
  nameAr: string;
  nameEn: string;
  dueDays: number;
  discountDays: number;
  discountPercent: number;
  isActive: boolean;
}

const INITIAL_CUSTOMERS: CustomerItem[] = [
  {
    id: "cust-1",
    customerNumber: "CUST-2026-0001",
    customerType: "B2B",
    nameAr: "شركة العربية الحديثة للمقاولات العامة",
    nameEn: "Al-Arabiya Modern Contracting Co.",
    taxId: "310123456700003",
    crNumber: "1010897654",
    city: "الرياض",
    district: "العليا",
    phone: "+966114567890",
    email: "accounts@alarabiya-contracting.sa",
    paymentTermsCode: "NET30",
    paymentTermsDueDays: 30,
    creditLimit: 250000,
    currentBalance: 34500,
    isCreditHold: false,
    status: "ACTIVE",
  },
  {
    id: "cust-2",
    customerNumber: "CUST-2026-0002",
    customerType: "B2B",
    nameAr: "مؤسسة المدينة للتجارة والتوريدات",
    nameEn: "Al-Madina Commercial Trading Est.",
    taxId: "310987654300003",
    crNumber: "2050123456",
    city: "الدمام",
    district: "الطبيشي",
    phone: "+966138901234",
    email: "info@almadina-trading.sa",
    paymentTermsCode: "NET60",
    paymentTermsDueDays: 60,
    creditLimit: 100000,
    currentBalance: 18400,
    isCreditHold: false,
    status: "ACTIVE",
  },
  {
    id: "cust-3",
    customerNumber: "CUST-2026-0003",
    customerType: "B2C",
    nameAr: "عبدالعزيز محمد العتيبي",
    nameEn: "Abdulaziz Mohammed Al-Otaibi",
    city: "الرياض",
    district: "الرحمانية",
    phone: "+966501234567",
    email: "a.otaibi@gmail.com",
    paymentTermsCode: "IMMEDIATE",
    paymentTermsDueDays: 0,
    creditLimit: 20000,
    currentBalance: 0,
    isCreditHold: false,
    status: "ACTIVE",
  },
];

const INITIAL_RECEIPTS: ReceiptItem[] = [
  {
    id: "rec-1",
    receiptNumber: "REC-2026-0001",
    customerNumber: "CUST-2026-0001",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    receiptDate: "2026-09-10",
    paymentMethod: "BANK_TRANSFER",
    depositAccount: "111002 (مصرف الراجحي - الحساب الجاري)",
    amount: 23000,
    allocatedAmount: 23000,
    unallocatedAmount: 0,
    referenceNumber: "TRF-ALRAJHI-99882",
    status: "POSTED",
  },
];

const INITIAL_OPEN_ITEMS: OpenItemEntity[] = [
  {
    id: "oi-1",
    documentType: "INVOICE",
    documentNumber: "INV-2026-0045",
    customerNumber: "CUST-2026-0001",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    documentDate: "2026-08-15",
    dueDate: "2026-09-14",
    originalAmount: 57500,
    allocatedAmount: 23000,
    remainingAmount: 34500,
    daysOverdue: 2,
    status: "PARTIALLY_PAID",
  },
  {
    id: "oi-2",
    documentType: "INVOICE",
    documentNumber: "INV-2026-0062",
    customerNumber: "CUST-2026-0002",
    customerNameAr: "مؤسسة المدينة للتجارة والتوريدات",
    documentDate: "2026-09-01",
    dueDate: "2026-10-31",
    originalAmount: 18400,
    allocatedAmount: 0,
    remainingAmount: 18400,
    daysOverdue: 0,
    status: "OPEN",
  },
];

const INITIAL_TERMS: PaymentTermItem[] = [
  {
    id: "term-1",
    code: "IMMEDIATE",
    nameAr: "الدفع الفوري نقداً",
    nameEn: "Cash / Immediate Payment",
    dueDays: 0,
    discountDays: 0,
    discountPercent: 0,
    isActive: true,
  },
  {
    id: "term-2",
    code: "NET15",
    nameAr: "أجل 15 يوماً",
    nameEn: "Net 15 Days",
    dueDays: 15,
    discountDays: 0,
    discountPercent: 0,
    isActive: true,
  },
  {
    id: "term-3",
    code: "NET30",
    nameAr: "أجل 30 يوماً (خصم 2% خلال 10 أيام)",
    nameEn: "Net 30 Days (2% 10 Days)",
    dueDays: 30,
    discountDays: 10,
    discountPercent: 2,
    isActive: true,
  },
  {
    id: "term-4",
    code: "NET60",
    nameAr: "أجل 60 يوماً",
    nameEn: "Net 60 Days",
    dueDays: 60,
    discountDays: 0,
    discountPercent: 0,
    isActive: true,
  },
];

export function CustomersView() {
  const { locale, t } = useLocale();
  const isAr = locale === "ar";
  const c = (t as any).customers || {};

  const [activeTab, setActiveTab] = useState<"customers" | "receipts" | "open_items" | "aging" | "terms">("customers");
  const [searchTerm, setSearchTerm] = useState("");
  const [typeFilter, setTypeFilter] = useState("ALL");

  // State collections
  const [customers, setCustomers] = useState<CustomerItem[]>(INITIAL_CUSTOMERS);
  const [receipts, setReceipts] = useState<ReceiptItem[]>(INITIAL_RECEIPTS);
  const [openItems, setOpenItems] = useState<OpenItemEntity[]>(INITIAL_OPEN_ITEMS);
  const [paymentTerms] = useState<PaymentTermItem[]>(INITIAL_TERMS);

  // Modals
  const [showCustomerModal, setShowCustomerModal] = useState(false);
  const [showReceiptModal, setShowReceiptModal] = useState(false);
  const [notification, setNotification] = useState<string | null>(null);

  // New Customer Form State
  const [custType, setCustType] = useState<CustomerItem["customerType"]>("B2B");
  const [custNameAr, setCustNameAr] = useState("");
  const [custNameEn, setCustNameEn] = useState("");
  const [custTaxId, setCustTaxId] = useState("");
  const [custCr, setCustCr] = useState("");
  const [custCity, setCustCity] = useState("الرياض");
  const [custPhone, setCustPhone] = useState("");
  const [custEmail, setCustEmail] = useState("");
  const [custLimit, setCustLimit] = useState(100000);
  const [custTerm, setCustTerm] = useState("NET30");

  // New Receipt Form State
  const [receiptCustId, setReceiptCustId] = useState("cust-1");
  const [receiptAmount, setReceiptAmount] = useState(15000);
  const [receiptMethod, setReceiptMethod] = useState<ReceiptItem["paymentMethod"]>("BANK_TRANSFER");
  const [receiptRef, setReceiptRef] = useState("");

  // KPIs
  const totalReceivables = openItems
    .filter((oi) => oi.status !== "PAID")
    .reduce((sum, oi) => sum + oi.remainingAmount, 0);
  const activeCustomersCount = customers.filter((cust) => cust.status === "ACTIVE").length;
  const overdueInvoicesCount = openItems.filter((oi) => oi.daysOverdue > 0 && oi.status !== "PAID").length;
  const receiptsMonthTotal = receipts.reduce((sum, r) => sum + r.amount, 0);

  // Filtered Customers
  const filteredCustomers = customers.filter((cust) => {
    const matchesSearch =
      cust.customerNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      cust.nameAr.includes(searchTerm) ||
      cust.nameEn.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (cust.taxId && cust.taxId.includes(searchTerm)) ||
      (cust.phone && cust.phone.includes(searchTerm));
    const matchesType = typeFilter === "ALL" || cust.customerType === typeFilter;
    return matchesSearch && matchesType;
  });

  // Filtered Receipts
  const filteredReceipts = receipts.filter((r) => {
    return (
      r.receiptNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.customerNameAr.includes(searchTerm) ||
      (r.referenceNumber && r.referenceNumber.toLowerCase().includes(searchTerm.toLowerCase()))
    );
  });

  // Filtered Open Items
  const filteredOpenItems = openItems.filter((oi) => {
    return (
      oi.documentNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      oi.customerNameAr.includes(searchTerm) ||
      oi.customerNumber.toLowerCase().includes(searchTerm.toLowerCase())
    );
  });

  const handleCreateCustomer = (e: React.FormEvent) => {
    e.preventDefault();
    const newCust: CustomerItem = {
      id: `cust-${Date.now()}`,
      customerNumber: `CUST-2026-${Math.floor(1000 + Math.random() * 9000)}`,
      customerType: custType,
      nameAr: custNameAr,
      nameEn: custNameEn || custNameAr,
      taxId: custTaxId || undefined,
      crNumber: custCr || undefined,
      city: custCity,
      phone: custPhone || undefined,
      email: custEmail || undefined,
      paymentTermsCode: custTerm,
      paymentTermsDueDays: custTerm === "NET60" ? 60 : custTerm === "NET30" ? 30 : 0,
      creditLimit: custLimit,
      currentBalance: 0,
      isCreditHold: false,
      status: "ACTIVE",
    };

    setCustomers([newCust, ...customers]);
    setShowCustomerModal(false);
    setCustNameAr("");
    setCustNameEn("");
    setCustTaxId("");
    setCustCr("");
    setNotification(c.success_customer_created || "تم إنشاء العميل بنجاح");
    setTimeout(() => setNotification(null), 4000);
  };

  const handleCreateReceipt = (e: React.FormEvent) => {
    e.preventDefault();
    const targetCust = customers.find((cust) => cust.id === receiptCustId) || customers[0];

    const newRec: ReceiptItem = {
      id: `rec-${Date.now()}`,
      receiptNumber: `REC-2026-${Math.floor(1000 + Math.random() * 9000)}`,
      customerNumber: targetCust.customerNumber,
      customerNameAr: targetCust.nameAr,
      customerNameEn: targetCust.nameEn,
      receiptDate: new Date().toISOString().substring(0, 10),
      paymentMethod: receiptMethod,
      depositAccount: "111002 (مصرف الراجحي - الحساب الجاري)",
      amount: receiptAmount,
      allocatedAmount: receiptAmount,
      unallocatedAmount: 0,
      referenceNumber: receiptRef || `TRF-MANUAL-${Math.floor(1000 + Math.random() * 9000)}`,
      status: "POSTED",
    };

    // Update receipts
    setReceipts([newRec, ...receipts]);

    // Deduct from open items for this customer
    let remainingToAllocate = receiptAmount;
    setOpenItems((prev) =>
      prev.map((oi) => {
        if (oi.customerNumber === targetCust.customerNumber && oi.status !== "PAID" && remainingToAllocate > 0) {
          const alloc = Math.min(remainingToAllocate, oi.remainingAmount);
          remainingToAllocate -= alloc;
          const newRemaining = oi.remainingAmount - alloc;
          return {
            ...oi,
            allocatedAmount: oi.allocatedAmount + alloc,
            remainingAmount: newRemaining,
            status: newRemaining === 0 ? "PAID" : "PARTIALLY_PAID",
          };
        }
        return oi;
      })
    );

    // Update customer current balance
    setCustomers((prev) =>
      prev.map((cust) => {
        if (cust.id === targetCust.id) {
          return {
            ...cust,
            currentBalance: Math.max(0, cust.currentBalance - receiptAmount),
          };
        }
        return cust;
      })
    );

    setShowReceiptModal(false);
    setReceiptRef("");
    setNotification(c.success_receipt_created || "تم تسجيل سند القبض وتحديث الأرصدة بنجاح");
    setTimeout(() => setNotification(null), 4000);
  };

  return (
    <div className="space-y-6">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed bottom-5 end-5 z-50 flex items-center gap-3 bg-emerald-600 text-white px-4 py-3 rounded-lg shadow-xl animate-in slide-in-from-bottom-5">
          <CheckCircle2 className="h-5 w-5" />
          <span className="text-sm font-medium">{notification}</span>
        </div>
      )}

      {/* Header Section */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
              <Users className="h-7 w-7 text-primary" />
              {c.title || "إدارة العملاء والذمم المدينة وسندات القبض (AR)"}
            </h1>
            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
              <ShieldCheck className="h-3.5 w-3.5" />
              ZATCA Phase 2 Verified
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {c.subtitle || "دليل العملاء المعتمد وفق متطلبات زاتكا، تتبع الفواتير المفتوحة، تحصيل وسندات القبض، وتحليل أعمار الديون اللحظي"}
          </p>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2.5">
          <button
            onClick={() => setShowReceiptModal(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold border bg-card hover:bg-muted text-foreground transition shadow-sm"
          >
            <Coins className="h-4 w-4 text-primary" />
            {c.new_receipt || "تسجيل سند قبض"}
          </button>
          <button
            onClick={() => setShowCustomerModal(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm"
          >
            <Plus className="h-4 w-4" />
            {c.new_customer || "إضافة عميل جديد"}
          </button>
        </div>
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Card 1: Total Receivables */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{c.kpi_total_receivables || "إجمالي الذمم المدينة (AR)"}</span>
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Landmark className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">
              {totalReceivables.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
            <span className="text-xs text-muted-foreground">{c.currency || "ر.س"}</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600">
            <TrendingUp className="h-3.5 w-3.5" />
            <span>حساب الأستاذ العام: 112000 Trade Receivables</span>
          </div>
        </div>

        {/* Card 2: Active Customers */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{c.kpi_active_customers || "العملاء النشطون"}</span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600">
              <Users className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">{activeCustomersCount}</span>
            <span className="text-xs text-muted-foreground">عملاء معتمدين</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
            <span>B2B: 2 | B2C: 1 | حكومي: 0</span>
          </div>
        </div>

        {/* Card 3: Overdue Invoices */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{c.kpi_overdue_invoices || "فواتير متأخرة السداد"}</span>
            <div className="p-2 rounded-lg bg-rose-500/10 text-rose-600">
              <AlertTriangle className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight text-rose-600">{overdueInvoicesCount}</span>
            <span className="text-xs text-muted-foreground">فواتير تجاوزت الاستحقاق</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-rose-600">
            <span>إجمالي المتأخر: 34,500.00 ر.س</span>
          </div>
        </div>

        {/* Card 4: Monthly Receipts */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{c.kpi_receipts_month || "تحصيلات الشهر الحالي"}</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-600">
              <Coins className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight text-emerald-600">
              {receiptsMonthTotal.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs text-muted-foreground">{c.currency || "ر.س"}</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600">
            <ArrowUpRight className="h-3.5 w-3.5" />
            <span>سندات مقبوضات مسددة 100%</span>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b pb-3">
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
          <button
            onClick={() => setActiveTab("customers")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "customers"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Users className="h-4 w-4" />
            {c.tab_customers || "دليل العملاء (Master List)"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {customers.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("receipts")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "receipts"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Coins className="h-4 w-4" />
            {c.tab_receipts || "سندات القبض (Receipts)"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {receipts.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("open_items")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "open_items"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <FileText className="h-4 w-4" />
            {c.tab_open_items || "الفواتير والمستندات المفتوحة"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {openItems.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("aging")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "aging"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Clock className="h-4 w-4" />
            {c.tab_aging || "تحليل أعمار الديون (Aging)"}
          </button>

          <button
            onClick={() => setActiveTab("terms")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "terms"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Percent className="h-4 w-4" />
            {c.tab_terms || "شروط الدفع والائتمان"}
          </button>
        </div>

        {/* Filter and Search */}
        <div className="flex items-center gap-2.5">
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="text-xs bg-card border rounded-lg px-2.5 py-1.5 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          >
            <option value="ALL">كل الفئات</option>
            <option value="B2B">منشآت وشركات (B2B)</option>
            <option value="B2C">أفراد ومستهلكين (B2C)</option>
            <option value="GOVERNMENT">جهات حكومية</option>
          </select>

          <div className="relative">
            <Search className="absolute inset-inline-start-2.5 top-2 h-3.5 w-3.5 text-muted-foreground" />
            <input
              type="text"
              placeholder="بحث برقم العميل، الاسم، أو الضريبي..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="text-xs bg-card border rounded-lg ps-8 pe-3 py-1.5 w-48 sm:w-64 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
        </div>
      </div>

      {/* Tab 1: Customers Master List */}
      {activeTab === "customers" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{c.customer_number || "رقم العميل"}</th>
                  <th className="px-4 py-3 text-start">{c.customer_name || "اسم العميل"}</th>
                  <th className="px-4 py-3 text-start">{c.customer_type || "الفئة"}</th>
                  <th className="px-4 py-3 text-start">{c.tax_id || "الرقم الضريبي"}</th>
                  <th className="px-4 py-3 text-start">{c.city || "المدينة"}</th>
                  <th className="px-4 py-3 text-center">شرط الدفع</th>
                  <th className="px-4 py-3 text-end">{c.credit_limit || "الحد الائتماني"}</th>
                  <th className="px-4 py-3 text-end">{c.current_balance || "الرصيد القائم"}</th>
                  <th className="px-4 py-3 text-center">{c.status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredCustomers.map((cust) => {
                  const utilization = cust.creditLimit > 0 ? Math.round((cust.currentBalance / cust.creditLimit) * 100) : 0;
                  return (
                    <tr key={cust.id} className="hover:bg-muted/30 transition">
                      <td className="px-4 py-3 font-mono font-bold text-primary">{cust.customerNumber}</td>
                      <td className="px-4 py-3">
                        <div className="font-semibold text-foreground">{isAr ? cust.nameAr : cust.nameEn}</div>
                        <div className="text-[11px] text-muted-foreground">{cust.phone || cust.email}</div>
                      </td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-muted text-foreground border border-border">
                          {cust.customerType}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono text-[11px]">
                        {cust.taxId ? (
                          <span className="text-foreground">{cust.taxId}</span>
                        ) : (
                          <span className="text-muted-foreground">غير مسجل</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">{cust.city}</td>
                      <td className="px-4 py-3 text-center">
                        <span className="font-mono text-[10px] font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
                          {cust.paymentTermsCode}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-end font-mono">
                        {cust.creditLimit.toLocaleString(isAr ? "ar-SA" : "en-US")} {c.currency || "ر.س"}
                      </td>
                      <td className="px-4 py-3 text-end">
                        <div className="font-mono font-bold text-foreground">
                          {cust.currentBalance.toLocaleString(isAr ? "ar-SA" : "en-US")} {c.currency || "ر.س"}
                        </div>
                        {cust.creditLimit > 0 && (
                          <div className="text-[10px] text-muted-foreground">
                            مستغل: {utilization}%
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3 text-center">
                        <span
                          className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                            cust.status === "ACTIVE"
                              ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                              : "bg-rose-500/10 text-rose-600 border-rose-500/20"
                          }`}
                        >
                          {cust.status}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Receipts Vouchers */}
      {activeTab === "receipts" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{c.receipt_number || "رقم السند"}</th>
                  <th className="px-4 py-3 text-start">{c.customer_name || "العميل"}</th>
                  <th className="px-4 py-3 text-start">{c.receipt_date || "تاريخ القبض"}</th>
                  <th className="px-4 py-3 text-start">{c.payment_method || "طريقة الدفع"}</th>
                  <th className="px-4 py-3 text-start">حساب الإيداع</th>
                  <th className="px-4 py-3 text-end">{c.amount || "المبلغ"}</th>
                  <th className="px-4 py-3 text-end">{c.allocated || "المخصص للفواتير"}</th>
                  <th className="px-4 py-3 text-center">{c.status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredReceipts.map((r) => (
                  <tr key={r.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-bold text-primary">{r.receiptNumber}</td>
                    <td className="px-4 py-3 font-semibold text-foreground">{r.customerNameAr}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{r.receiptDate}</td>
                    <td className="px-4 py-3 font-mono text-[11px]">
                      <span className="px-2 py-0.5 rounded bg-muted text-foreground">
                        {r.paymentMethod}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{r.depositAccount}</td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-emerald-600">
                      {r.amount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {c.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-semibold">
                      {r.allocatedAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {c.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                        {r.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 3: Open Invoices & Items */}
      {activeTab === "open_items" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{c.open_invoice || "رقم الفاتورة"}</th>
                  <th className="px-4 py-3 text-start">{c.customer_name || "العميل"}</th>
                  <th className="px-4 py-3 text-start">تاريخ الفاتورة</th>
                  <th className="px-4 py-3 text-start">{c.due_date || "تاريخ الاستحقاق"}</th>
                  <th className="px-4 py-3 text-end">{c.original_amount || "القيمة الأصلية"}</th>
                  <th className="px-4 py-3 text-end">{c.remaining_amount || "المتبقي للسداد"}</th>
                  <th className="px-4 py-3 text-center">{c.days_overdue || "أيام التأخير"}</th>
                  <th className="px-4 py-3 text-center">{c.status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredOpenItems.map((oi) => (
                  <tr key={oi.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-bold text-primary">{oi.documentNumber}</td>
                    <td className="px-4 py-3 font-semibold text-foreground">{oi.customerNameAr}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{oi.documentDate}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{oi.dueDate}</td>
                    <td className="px-4 py-3 text-end font-mono">
                      {oi.originalAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {c.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                      {oi.remainingAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {c.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-center font-mono">
                      {oi.daysOverdue > 0 ? (
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/10 text-rose-600 border border-rose-500/20">
                          {oi.daysOverdue} يوم تأخير
                        </span>
                      ) : (
                        <span className="text-emerald-600 text-[10px]">سارية</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                          oi.status === "PAID"
                            ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                            : oi.status === "PARTIALLY_PAID"
                            ? "bg-amber-500/10 text-amber-600 border-amber-500/20"
                            : "bg-blue-500/10 text-blue-600 border-blue-500/20"
                        }`}
                      >
                        {oi.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 4: Aging Breakdown */}
      {activeTab === "aging" && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-5 gap-3">
            <div className="p-3.5 rounded-xl border bg-card text-center">
              <span className="text-[11px] text-muted-foreground block">{c.aging_current || "غير متأخرة (جارية)"}</span>
              <span className="font-mono text-lg font-bold text-emerald-600 mt-1 block">18,400.00 ر.س</span>
            </div>
            <div className="p-3.5 rounded-xl border bg-card text-center">
              <span className="text-[11px] text-muted-foreground block">{c.aging_1_30 || "1 - 30 يوماً"}</span>
              <span className="font-mono text-lg font-bold text-amber-600 mt-1 block">34,500.00 ر.س</span>
            </div>
            <div className="p-3.5 rounded-xl border bg-card text-center">
              <span className="text-[11px] text-muted-foreground block">{c.aging_31_60 || "31 - 60 يوماً"}</span>
              <span className="font-mono text-lg font-bold text-muted-foreground mt-1 block">0.00 ر.س</span>
            </div>
            <div className="p-3.5 rounded-xl border bg-card text-center">
              <span className="text-[11px] text-muted-foreground block">{c.aging_61_90 || "61 - 90 يوماً"}</span>
              <span className="font-mono text-lg font-bold text-muted-foreground mt-1 block">0.00 ر.س</span>
            </div>
            <div className="p-3.5 rounded-xl border bg-card text-center">
              <span className="text-[11px] text-muted-foreground block">{c.aging_over_90 || "أكثر من 90 يوماً"}</span>
              <span className="font-mono text-lg font-bold text-muted-foreground mt-1 block">0.00 ر.س</span>
            </div>
          </div>

          <div className="rounded-xl border bg-card p-4">
            <h3 className="text-sm font-bold text-foreground mb-2 flex items-center gap-2">
              <Clock className="h-4 w-4 text-primary" />
              مصفوفة أعمار الديون التحليلية حسب العملاء (Aging Matrix)
            </h3>
            <p className="text-xs text-muted-foreground mb-4">
              توزيع الذمم المدينة وفق فترات الاستحقاق الفعلية للمطابقة الدورية مع ميزان المراجعة
            </p>
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                  <tr>
                    <th className="px-4 py-3 text-start">العميل</th>
                    <th className="px-4 py-3 text-end">إجمالي المستحق</th>
                    <th className="px-4 py-3 text-end text-emerald-600">جارية</th>
                    <th className="px-4 py-3 text-end text-amber-600">1 - 30 يوم</th>
                    <th className="px-4 py-3 text-end">31 - 60 يوم</th>
                    <th className="px-4 py-3 text-end">61 - 90 يوم</th>
                    <th className="px-4 py-3 text-end text-rose-600">&gt; 90 يوم</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3">
                      <span className="font-bold text-foreground">شركة العربية الحديثة للمقاولات العامة</span>
                      <span className="block text-[10px] text-muted-foreground font-mono">CUST-2026-0001</span>
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-bold">34,500.00 ر.س</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-amber-600">34,500.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                  </tr>
                  <tr className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3">
                      <span className="font-bold text-foreground">مؤسسة المدينة للتجارة والتوريدات</span>
                      <span className="block text-[10px] text-muted-foreground font-mono">CUST-2026-0002</span>
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-bold">18,400.00 ر.س</td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-emerald-600">18,400.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                    <td className="px-4 py-3 text-end font-mono">0.00</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 5: Payment Terms */}
      {activeTab === "terms" && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {paymentTerms.map((term) => (
            <div key={term.id} className="p-4 rounded-xl border bg-card shadow-sm">
              <div className="flex items-center justify-between pb-2 border-b">
                <span className="font-mono font-bold text-primary">{term.code}</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 font-semibold">
                  نشط
                </span>
              </div>
              <h4 className="text-xs font-bold text-foreground mt-3">{isAr ? term.nameAr : term.nameEn}</h4>
              <div className="mt-3 text-xs space-y-1 text-muted-foreground">
                <div className="flex justify-between">
                  <span>فترة السداد:</span>
                  <span className="font-mono font-semibold text-foreground">{term.dueDays} يوماً</span>
                </div>
                {term.discountPercent > 0 && (
                  <div className="flex justify-between text-emerald-600 font-medium">
                    <span>خصم تعجيل الدفع:</span>
                    <span className="font-mono">{term.discountPercent}% خلال {term.discountDays} أيام</span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Customer Modal */}
      {showCustomerModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 overflow-y-auto">
          <div className="w-full max-w-lg rounded-xl border bg-card p-6 shadow-2xl animate-in zoom-in-95 my-8">
            <div className="flex items-center justify-between pb-3 border-b">
              <div className="flex items-center gap-2">
                <Users className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-base text-foreground">
                  {c.new_customer || "إضافة عميل جديد"}
                </h3>
              </div>
              <button
                onClick={() => setShowCustomerModal(false)}
                className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleCreateCustomer} className="space-y-4 mt-4 text-xs">
              {/* Customer Type */}
              <div>
                <label className="font-medium text-foreground block mb-1">فئة العميل</label>
                <div className="grid grid-cols-3 gap-2">
                  {(["B2B", "B2C", "GOVERNMENT"] as const).map((t) => (
                    <button
                      key={t}
                      type="button"
                      onClick={() => setCustType(t)}
                      className={`py-1.5 rounded-lg border text-xs font-semibold transition ${
                        custType === t
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-background text-foreground hover:bg-muted"
                      }`}
                    >
                      {t === "B2B" ? "شركة (B2B)" : t === "B2C" ? "فرد (B2C)" : "حكومي"}
                    </button>
                  ))}
                </div>
              </div>

              {/* Name Arabic & English */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">اسم العميل (عربي) *</label>
                  <input
                    type="text"
                    required
                    placeholder="مثال: شركة الروابي للتجارة"
                    value={custNameAr}
                    onChange={(e) => setCustNameAr(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">اسم العميل (إنجليزي)</label>
                  <input
                    type="text"
                    placeholder="Al-Rawabi Trading Co."
                    value={custNameEn}
                    onChange={(e) => setCustNameEn(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
              </div>

              {/* VAT & CR Number */}
              {custType === "B2B" && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="font-medium text-foreground block mb-1">
                      الرقم الضريبي (15 خانة)
                    </label>
                    <input
                      type="text"
                      maxLength={15}
                      placeholder="3XXXXXXXXXXXXX3"
                      value={custTaxId}
                      onChange={(e) => setCustTaxId(e.target.value)}
                      className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label className="font-medium text-foreground block mb-1">السجل التجاري (10 خانات)</label>
                    <input
                      type="text"
                      maxLength={10}
                      placeholder="1010XXXXXX"
                      value={custCr}
                      onChange={(e) => setCustCr(e.target.value)}
                      className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>
                </div>
              )}

              {/* City & Address */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">المدينة *</label>
                  <select
                    value={custCity}
                    onChange={(e) => setCustCity(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  >
                    <option value="الرياض">الرياض</option>
                    <option value="جدة">جدة</option>
                    <option value="الدمام">الدمام</option>
                    <option value="مكة المكرمة">مكة المكرمة</option>
                    <option value="المدينة المنورة">المدينة المنورة</option>
                    <option value="الخبر">الخبر</option>
                  </select>
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">رقم الجوال / الهاتف</label>
                  <input
                    type="text"
                    placeholder="+9665XXXXXXXX"
                    value={custPhone}
                    onChange={(e) => setCustPhone(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
              </div>

              {/* Payment Terms & Credit Limit */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">شرط الدفع المعتمد</label>
                  <select
                    value={custTerm}
                    onChange={(e) => setCustTerm(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  >
                    <option value="IMMEDIATE">الدفع الفوري (نقداً)</option>
                    <option value="NET15">أجل 15 يوماً</option>
                    <option value="NET30">أجل 30 يوماً</option>
                    <option value="NET60">أجل 60 يوماً</option>
                  </select>
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">الحد الائتماني (ر.س)</label>
                  <input
                    type="number"
                    min="0"
                    step="1000"
                    value={custLimit}
                    onChange={(e) => setCustLimit(parseFloat(e.target.value) || 0)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
              </div>

              {/* Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setShowCustomerModal(false)}
                  className="px-4 py-2 rounded-lg border hover:bg-muted text-foreground transition"
                >
                  {c.cancel || "إلغاء"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition shadow-sm"
                >
                  {c.save || "حفظ بيانات العميل"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Receipt Modal */}
      {showReceiptModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg rounded-xl border bg-card p-6 shadow-2xl animate-in zoom-in-95">
            <div className="flex items-center justify-between pb-3 border-b">
              <div className="flex items-center gap-2">
                <Coins className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-base text-foreground">
                  {c.new_receipt || "تسجيل سند قبض جديد (Receipt Voucher)"}
                </h3>
              </div>
              <button
                onClick={() => setShowReceiptModal(false)}
                className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleCreateReceipt} className="space-y-4 mt-4 text-xs">
              {/* Customer Selection */}
              <div>
                <label className="font-medium text-foreground block mb-1">العميل المقبوض منه</label>
                <select
                  value={receiptCustId}
                  onChange={(e) => setReceiptCustId(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  {customers.map((cust) => (
                    <option key={cust.id} value={cust.id}>
                      {cust.customerNumber} - {cust.nameAr} (رصيد: {cust.currentBalance.toLocaleString()} ر.س)
                    </option>
                  ))}
                </select>
              </div>

              {/* Amount & Payment Method */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">المبلغ المقبوض (ر.س) *</label>
                  <input
                    type="number"
                    min="1"
                    step="0.01"
                    required
                    value={receiptAmount}
                    onChange={(e) => setReceiptAmount(parseFloat(e.target.value) || 0)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono font-bold focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">طريقة الدفع</label>
                  <select
                    value={receiptMethod}
                    onChange={(e) => setReceiptMethod(e.target.value as any)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  >
                    <option value="BANK_TRANSFER">تحويل بنكي (Bank Transfer)</option>
                    <option value="MADA">مدى (Mada)</option>
                    <option value="SADAD">سداد (SADAD)</option>
                    <option value="CASH">نقداً (Cash)</option>
                    <option value="CHEQUE">شيك مصرفي (Cheque)</option>
                  </select>
                </div>
              </div>

              {/* Deposit Account */}
              <div>
                <label className="font-medium text-foreground block mb-1">الحساب المالي للإيداع</label>
                <select
                  disabled
                  className="w-full rounded-lg border bg-muted/50 px-3 py-2 text-foreground font-mono focus:outline-none"
                >
                  <option>111002 - مصرف الراجحي (حساب التشغيل SAR)</option>
                </select>
              </div>

              {/* Reference Number */}
              <div>
                <label className="font-medium text-foreground block mb-1">رقم المرجع البنكي / التحويل</label>
                <input
                  type="text"
                  placeholder="مثال: TRF-ALRAJHI-112233"
                  value={receiptRef}
                  onChange={(e) => setReceiptRef(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              {/* Allocation Preview Notice */}
              <div className="p-3 rounded-lg bg-primary/5 border border-primary/20 text-[11px] text-primary space-y-1">
                <div className="font-bold flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4 shrink-0" />
                  مطابقة وتسوية الفواتير آلياً (Auto-Allocation)
                </div>
                <p className="text-muted-foreground">
                  سيتم توجيه مبلغ السند تلقائياً لتسديد أقدم الفواتير المفتوحة للعميل وتحديث الرصيد في الأستاذ المساعد.
                </p>
              </div>

              {/* Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setShowReceiptModal(false)}
                  className="px-4 py-2 rounded-lg border hover:bg-muted text-foreground transition"
                >
                  {c.cancel || "إلغاء"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition shadow-sm"
                >
                  {c.new_receipt || "اعتماد وترحيل سند القبض"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

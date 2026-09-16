"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  ShoppingBag,
  FileText,
  Truck,
  RotateCcw,
  CheckCircle2,
  Clock,
  QrCode,
  Plus,
  Search,
  SlidersHorizontal,
  TrendingUp,
  Boxes,
  ShieldCheck,
  AlertTriangle,
  X,
  CreditCard,
  Building2,
  Calendar,
  Layers,
  ArrowUpRight,
  Send,
  Eye,
  Percent,
} from "lucide-react";

interface InvoiceItem {
  id: string;
  invoiceNumber: string;
  invoiceType: "STANDARD" | "SIMPLIFIED";
  customerNameAr: string;
  customerNameEn: string;
  customerTaxId?: string;
  customerCr?: string;
  issueDate: string;
  dueDate: string;
  subtotal: number;
  taxAmount: number;
  totalAmount: number;
  status: "DRAFT" | "APPROVED" | "POSTED" | "CANCELLED";
  orderNumber?: string;
  qrCodeData?: string;
  lines: Array<{
    description: string;
    quantity: number;
    unitPrice: number;
    taxableAmount: number;
    taxAmount: number;
    totalAmount: number;
  }>;
}

interface OrderItem {
  id: string;
  orderNumber: string;
  customerNameAr: string;
  customerNameEn: string;
  orderDate: string;
  totalAmount: number;
  status: "DRAFT" | "CONFIRMED" | "PARTIALLY_DELIVERED" | "DELIVERED" | "INVOICED";
  deliveredPercent: number;
  invoicedPercent: number;
  quoteNumber?: string;
}

interface QuotationItem {
  id: string;
  quotationNumber: string;
  customerNameAr: string;
  customerNameEn: string;
  quoteDate: string;
  validUntil: string;
  totalAmount: number;
  status: "DRAFT" | "SENT" | "ACCEPTED" | "REJECTED" | "CONVERTED";
}

interface DeliveryItem {
  id: string;
  deliveryNumber: string;
  orderNumber: string;
  customerNameAr: string;
  customerNameEn: string;
  warehouseName: string;
  deliveryDate: string;
  status: "DRAFT" | "SHIPPED" | "DELIVERED";
  itemsCount: number;
}

interface CreditNoteItem {
  id: string;
  creditNoteNumber: string;
  originalInvoiceNumber: string;
  customerNameAr: string;
  customerNameEn: string;
  issueDate: string;
  reason: "RETURN" | "DISCOUNT" | "CORRECTION";
  totalAmount: number;
  status: "DRAFT" | "POSTED";
}

const INITIAL_INVOICES: InvoiceItem[] = [
  {
    id: "inv-1",
    invoiceNumber: "INV-2026-0041",
    invoiceType: "STANDARD",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    customerTaxId: "310123456700003",
    customerCr: "1010897654",
    issueDate: "2026-09-15",
    dueDate: "2026-10-15",
    subtotal: 30000,
    taxAmount: 4500,
    totalAmount: 34500,
    status: "POSTED",
    orderNumber: "SO-2026-0082",
    qrCodeData: "ZATCA|Seller:Al-Arabiya|VAT:310123456700003|Total:34500.00|Tax:4500.00",
    lines: [
      { description: "توريد معدات ومضخات غاطسة", quantity: 2, unitPrice: 15000, taxableAmount: 30000, taxAmount: 4500, totalAmount: 34500 },
    ],
  },
  {
    id: "inv-2",
    invoiceNumber: "INV-2026-0042",
    invoiceType: "STANDARD",
    customerNameAr: "مؤسسة المدينة للتجارة والتوريدات",
    customerNameEn: "Al-Madina Commercial Trading Est.",
    customerTaxId: "310987654300003",
    customerCr: "2050123456",
    issueDate: "2026-09-14",
    dueDate: "2026-11-14",
    subtotal: 16000,
    taxAmount: 2400,
    totalAmount: 18400,
    status: "POSTED",
    orderNumber: "SO-2026-0083",
    qrCodeData: "ZATCA|Seller:Al-Madina|VAT:310987654300003|Total:18400.00|Tax:2400.00",
    lines: [
      { description: "قواطع كهربائية صناعية 400 أمبير", quantity: 8, unitPrice: 2000, taxableAmount: 16000, taxAmount: 2400, totalAmount: 18400 },
    ],
  },
  {
    id: "inv-3",
    invoiceNumber: "INV-2026-0043",
    invoiceType: "SIMPLIFIED",
    customerNameAr: "عبدالعزيز محمد العتيبي",
    customerNameEn: "Abdulaziz Mohammed Al-Otaibi",
    issueDate: "2026-09-16",
    dueDate: "2026-09-16",
    subtotal: 2500,
    taxAmount: 375,
    totalAmount: 2875,
    status: "POSTED",
    qrCodeData: "ZATCA|B2C|Total:2875.00|Tax:375.00",
    lines: [
      { description: "قطع غيار وصيانة فورية", quantity: 1, unitPrice: 2500, taxableAmount: 2500, taxAmount: 375, totalAmount: 2875 },
    ],
  },
  {
    id: "inv-4",
    invoiceNumber: "INV-2026-0044",
    invoiceType: "STANDARD",
    customerNameAr: "شركة اليمامة للاستشارات الهندسية",
    customerNameEn: "Al-Yamama Engineering Consultancies",
    customerTaxId: "310555666700003",
    customerCr: "1010776655",
    issueDate: "2026-09-16",
    dueDate: "2026-10-16",
    subtotal: 50000,
    taxAmount: 7500,
    totalAmount: 57500,
    status: "DRAFT",
    lines: [
      { description: "برمجيات وحلول تحكم آلي", quantity: 1, unitPrice: 50000, taxableAmount: 50000, taxAmount: 7500, totalAmount: 57500 },
    ],
  },
];

const INITIAL_ORDERS: OrderItem[] = [
  {
    id: "ord-1",
    orderNumber: "SO-2026-0082",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    orderDate: "2026-09-10",
    totalAmount: 34500,
    status: "INVOICED",
    deliveredPercent: 100,
    invoicedPercent: 100,
    quoteNumber: "QT-2026-0012",
  },
  {
    id: "ord-2",
    orderNumber: "SO-2026-0083",
    customerNameAr: "مؤسسة المدينة للتجارة والتوريدات",
    customerNameEn: "Al-Madina Commercial Trading Est.",
    orderDate: "2026-09-12",
    totalAmount: 18400,
    status: "DELIVERED",
    deliveredPercent: 100,
    invoicedPercent: 100,
  },
  {
    id: "ord-3",
    orderNumber: "SO-2026-0084",
    customerNameAr: "شركة الخليج الكهروميكانيكية",
    customerNameEn: "Gulf Electro-Mechanical Co.",
    orderDate: "2026-09-14",
    totalAmount: 86250,
    status: "CONFIRMED",
    deliveredPercent: 60,
    invoicedPercent: 0,
    quoteNumber: "QT-2026-0014",
  },
  {
    id: "ord-4",
    orderNumber: "SO-2026-0085",
    customerNameAr: "شركة آفاق التوزيع المحدودة",
    customerNameEn: "Horizon Distribution Co.",
    orderDate: "2026-09-15",
    totalAmount: 43125,
    status: "DRAFT",
    deliveredPercent: 0,
    invoicedPercent: 0,
  },
];

const INITIAL_QUOTATIONS: QuotationItem[] = [
  {
    id: "qt-1",
    quotationNumber: "QT-2026-0012",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    quoteDate: "2026-09-05",
    validUntil: "2026-10-05",
    totalAmount: 34500,
    status: "CONVERTED",
  },
  {
    id: "qt-2",
    quotationNumber: "QT-2026-0014",
    customerNameAr: "شركة الخليج الكهروميكانيكية",
    customerNameEn: "Gulf Electro-Mechanical Co.",
    quoteDate: "2026-09-11",
    validUntil: "2026-10-11",
    totalAmount: 86250,
    status: "CONVERTED",
  },
  {
    id: "qt-3",
    quotationNumber: "QT-2026-0015",
    customerNameAr: "شركة نجد للحلول اللوجستية",
    customerNameEn: "Najd Logistics Solutions",
    quoteDate: "2026-09-15",
    validUntil: "2026-10-15",
    totalAmount: 28750,
    status: "SENT",
  },
];

const INITIAL_DELIVERIES: DeliveryItem[] = [
  {
    id: "del-1",
    deliveryNumber: "DN-2026-0031",
    orderNumber: "SO-2026-0082",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    warehouseName: "المستودع المركزي - الرياض",
    deliveryDate: "2026-09-13",
    status: "DELIVERED",
    itemsCount: 2,
  },
  {
    id: "del-2",
    deliveryNumber: "DN-2026-0032",
    orderNumber: "SO-2026-0083",
    customerNameAr: "مؤسسة المدينة للتجارة والتوريدات",
    customerNameEn: "Al-Madina Commercial Trading Est.",
    warehouseName: "مستودع فرع الدمام",
    deliveryDate: "2026-09-14",
    status: "DELIVERED",
    itemsCount: 8,
  },
  {
    id: "del-3",
    deliveryNumber: "DN-2026-0033",
    orderNumber: "SO-2026-0084",
    customerNameAr: "شركة الخليج الكهروميكانيكية",
    customerNameEn: "Gulf Electro-Mechanical Co.",
    warehouseName: "المستودع المركزي - الرياض",
    deliveryDate: "2026-09-16",
    status: "SHIPPED",
    itemsCount: 5,
  },
];

const INITIAL_CREDIT_NOTES: CreditNoteItem[] = [
  {
    id: "cn-1",
    creditNoteNumber: "CN-2026-0005",
    originalInvoiceNumber: "INV-2026-0041",
    customerNameAr: "شركة العربية الحديثة للمقاولات العامة",
    customerNameEn: "Al-Arabiya Modern Contracting Co.",
    issueDate: "2026-09-16",
    reason: "RETURN",
    totalAmount: 2300,
    status: "POSTED",
  },
];

export function SalesView() {
  const { locale, t } = useLocale();
  const isAr = locale === "ar";
  const s = (t as any).sales || {};

  const [activeTab, setActiveTab] = useState<"invoices" | "orders" | "quotations" | "deliveries" | "credit_notes">("invoices");
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("ALL");

  // State
  const [invoices, setInvoices] = useState<InvoiceItem[]>(INITIAL_INVOICES);
  const [orders, setOrders] = useState<OrderItem[]>(INITIAL_ORDERS);
  const [quotations, setQuotations] = useState<QuotationItem[]>(INITIAL_QUOTATIONS);
  const [deliveries, setDeliveries] = useState<DeliveryItem[]>(INITIAL_DELIVERIES);
  const [creditNotes, setCreditNotes] = useState<CreditNoteItem[]>(INITIAL_CREDIT_NOTES);

  // Modals
  const [showInvoiceModal, setShowInvoiceModal] = useState(false);
  const [showOrderModal, setShowOrderModal] = useState(false);
  const [showQuotationModal, setShowQuotationModal] = useState(false);
  const [selectedQR, setSelectedQR] = useState<{ number: string; qr: string } | null>(null);
  const [notification, setNotification] = useState<string | null>(null);

  // Form State - Invoice
  const [invType, setInvType] = useState<"STANDARD" | "SIMPLIFIED">("STANDARD");
  const [invCustomer, setInvCustomer] = useState("شركة العربية الحديثة للمقاولات العامة");
  const [invTaxId, setInvTaxId] = useState("310123456700003");
  const [invItem, setInvItem] = useState("توريد صمامات ومحابس صناعية");
  const [invQty, setInvQty] = useState(5);
  const [invPrice, setInvPrice] = useState(2000);

  // Form State - Order
  const [ordCustomer, setOrdCustomer] = useState("شركة الخليج الكهروميكانيكية");
  const [ordAmount, setOrdAmount] = useState(25000);

  // Form State - Quote
  const [quoteCustomer, setQuoteCustomer] = useState("شركة الأمل للتجارة");
  const [quoteAmount, setQuoteAmount] = useState(40000);

  // KPIs
  const totalInvoiced = invoices.filter((i) => i.status === "POSTED").reduce((sum, i) => sum + i.totalAmount, 0);
  const activeOrdersCount = orders.filter((o) => o.status === "CONFIRMED" || o.status === "PARTIALLY_DELIVERED").length;
  const completedDeliveriesCount = deliveries.filter((d) => d.status === "DELIVERED").length;
  const creditNotesTotal = creditNotes.reduce((sum, c) => sum + c.totalAmount, 0);

  const notify = (msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  // Handlers
  const handlePostInvoice = (id: string) => {
    setInvoices((prev) =>
      prev.map((inv) =>
        inv.id === id
          ? {
              ...inv,
              status: "POSTED",
              qrCodeData: `ZATCA|${inv.invoiceNumber}|Total:${inv.totalAmount}|Tax:${inv.taxAmount}`,
            }
          : inv
      )
    );
    notify(isAr ? "تم اعتماد وترحيل الفاتورة إلى الأستاذ العام وتوليد رمز ZATCA" : "Invoice posted to General Ledger and ZATCA QR generated");
  };

  const handleConfirmOrder = (id: string) => {
    setOrders((prev) =>
      prev.map((o) => (o.id === id ? { ...o, status: "CONFIRMED" } : o))
    );
    notify(isAr ? "تم فحص الائتمان وتأكيد أمر البيع بنجاح" : "Credit verified and sales order confirmed successfully");
  };

  const handleConvertQuote = (id: string) => {
    setQuotations((prev) =>
      prev.map((q) => (q.id === id ? { ...q, status: "CONVERTED" } : q))
    );
    notify(isAr ? "تم قبول عرض الأسعار وتحويله لأمر بيع جديد" : "Quotation accepted and converted to sales order");
  };

  const handleCreateInvoice = (e: React.FormEvent) => {
    e.preventDefault();
    const subtotal = invQty * invPrice;
    const tax = subtotal * 0.15;
    const total = subtotal + tax;
    const num = `INV-2026-00${invoices.length + 45}`;

    const newInv: InvoiceItem = {
      id: `inv-${Date.now()}`,
      invoiceNumber: num,
      invoiceType: invType,
      customerNameAr: invCustomer,
      customerNameEn: "Commercial Customer",
      customerTaxId: invType === "STANDARD" ? invTaxId : undefined,
      issueDate: new Date().toISOString().split("T")[0],
      dueDate: new Date(Date.now() + 30 * 86400000).toISOString().split("T")[0],
      subtotal,
      taxAmount: tax,
      totalAmount: total,
      status: "DRAFT",
      lines: [
        {
          description: invItem,
          quantity: invQty,
          unitPrice: invPrice,
          taxableAmount: subtotal,
          taxAmount: tax,
          totalAmount: total,
        },
      ],
    };

    setInvoices([newInv, ...invoices]);
    setShowInvoiceModal(false);
    notify(isAr ? "تم إنشاء الفاتورة الضريبية بنجاح" : "Tax invoice created successfully");
  };

  // Filtered invoices
  const filteredInvoices = invoices.filter((i) => {
    const matchesSearch =
      i.invoiceNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      i.customerNameAr.includes(searchTerm) ||
      i.customerNameEn.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === "ALL" || i.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="p-4 sm:p-6 lg:p-8 space-y-6 max-w-7xl mx-auto">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 bg-primary text-primary-foreground px-6 py-3 rounded-xl shadow-2xl flex items-center gap-3 border border-primary/20 backdrop-blur animate-in fade-in slide-in-from-top-4 duration-200">
          <CheckCircle2 className="h-5 w-5 text-green-300 shrink-0" />
          <span className="font-semibold text-sm">{notification}</span>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
              <ShoppingBag className="h-7 w-7" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
                {s.title || "إدارة المبيعات ودورة الفوترة الإلكترونية (Order-to-Cash)"}
              </h1>
              <p className="text-sm text-muted-foreground mt-0.5">
                {s.subtitle || "عروض الأسعار، أوامر البيع، مذكرات التسليم وصرف المخزون، الفواتير الضريبية والمبسطة ZATCA، وإشعارات الدائن"}
              </p>
            </div>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          <button
            onClick={() => setShowInvoiceModal(true)}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition shadow-sm hover:shadow active:scale-95"
          >
            <Plus className="h-4 w-4" />
            <span>{s.new_invoice || "إصدار فاتورة ضريبية"}</span>
          </button>
        </div>
      </div>

      {/* KPI Overview */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Invoiced */}
        <div className="p-5 rounded-2xl bg-card border shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {s.kpi_total_invoiced || "إجمالي المبيعات المفوترة"}
            </span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-600">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold text-foreground">
              {totalInvoiced.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs text-muted-foreground mr-1.5 ml-1.5">ر.س</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600 font-medium">
            <ArrowUpRight className="h-3.5 w-3.5" />
            <span>+18.4% نمو المبيعات الشهرية</span>
          </div>
        </div>

        {/* Active Orders */}
        <div className="p-5 rounded-2xl bg-card border shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {s.kpi_open_orders || "أوامر البيع الجارية"}
            </span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600">
              <ShoppingBag className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold text-foreground">{activeOrdersCount}</span>
            <span className="text-xs text-muted-foreground mr-1.5 ml-1.5">أمر نشط</span>
          </div>
          <div className="mt-2 text-xs text-muted-foreground flex items-center gap-1">
            <Clock className="h-3.5 w-3.5 text-amber-500" />
            <span>تحت التنفيذ والتسليم الجزئي</span>
          </div>
        </div>

        {/* Completed Deliveries */}
        <div className="p-5 rounded-2xl bg-card border shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {s.kpi_deliveries || "إرساليات التسليم المكتملة"}
            </span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600">
              <Truck className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold text-foreground">{completedDeliveriesCount}</span>
            <span className="text-xs text-muted-foreground mr-1.5 ml-1.5">إرسالية</span>
          </div>
          <div className="mt-2 text-xs text-muted-foreground flex items-center gap-1">
            <Boxes className="h-3.5 w-3.5 text-emerald-500" />
            <span>صُرفت من المخزون بموجب حركات محكمة</span>
          </div>
        </div>

        {/* Credit Notes Total */}
        <div className="p-5 rounded-2xl bg-card border shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {s.kpi_credit_notes || "إجمالي المردودات والدائن"}
            </span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-600">
              <RotateCcw className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold text-foreground">
              {creditNotesTotal.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs text-muted-foreground mr-1.5 ml-1.5">ر.س</span>
          </div>
          <div className="mt-2 text-xs text-muted-foreground flex items-center gap-1">
            <ShieldCheck className="h-3.5 w-3.5 text-purple-500" />
            <span>تسويات مردودات مبيعات معتمدة</span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b overflow-x-auto pb-1">
        <button
          onClick={() => setActiveTab("invoices")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold transition border-b-2 whitespace-nowrap ${
            activeTab === "invoices"
              ? "border-primary text-primary bg-primary/5"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileText className="h-4 w-4" />
          <span>{s.tab_invoices || "فواتير المبيعات"}</span>
          <span className="px-2 py-0.5 text-xs rounded-full bg-primary/10 text-primary">
            {invoices.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("orders")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold transition border-b-2 whitespace-nowrap ${
            activeTab === "orders"
              ? "border-primary text-primary bg-primary/5"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <ShoppingBag className="h-4 w-4" />
          <span>{s.tab_orders || "أوامر البيع"}</span>
          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
            {orders.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("quotations")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold transition border-b-2 whitespace-nowrap ${
            activeTab === "quotations"
              ? "border-primary text-primary bg-primary/5"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Send className="h-4 w-4" />
          <span>{s.tab_quotations || "عروض الأسعار"}</span>
          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
            {quotations.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("deliveries")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold transition border-b-2 whitespace-nowrap ${
            activeTab === "deliveries"
              ? "border-primary text-primary bg-primary/5"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Truck className="h-4 w-4" />
          <span>{s.tab_deliveries || "مذكرات التسليم"}</span>
          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
            {deliveries.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("credit_notes")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold transition border-b-2 whitespace-nowrap ${
            activeTab === "credit_notes"
              ? "border-primary text-primary bg-primary/5"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <RotateCcw className="h-4 w-4" />
          <span>{s.tab_credit_notes || "إشعارات الدائن"}</span>
          <span className="px-2 py-0.5 text-xs rounded-full bg-muted text-muted-foreground">
            {creditNotes.length}
          </span>
        </button>
      </div>

      {/* Tab Content: INVOICES */}
      {activeTab === "invoices" && (
        <div className="space-y-4">
          {/* Controls */}
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute right-3 rtl:right-3 ltr:left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={isAr ? "بحث برقم الفاتورة أو اسم العميل..." : "Search by invoice # or customer..."}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-9 py-2 rounded-xl border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div className="flex items-center gap-2">
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-2 rounded-xl border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="ALL">{isAr ? "كافة الحالات" : "All Statuses"}</option>
                <option value="POSTED">{isAr ? "مرحلة محاسبياً (POSTED)" : "Posted"}</option>
                <option value="DRAFT">{isAr ? "مسودة (DRAFT)" : "Draft"}</option>
              </select>
            </div>
          </div>

          {/* Table */}
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right rtl:text-right ltr:text-left">
                <thead className="bg-muted/50 text-muted-foreground text-xs font-semibold border-b">
                  <tr>
                    <th className="p-3.5">{s.invoice_number || "رقم الفاتورة"}</th>
                    <th className="p-3.5">{s.invoice_type || "النوع"}</th>
                    <th className="p-3.5">{s.customer || "العميل"}</th>
                    <th className="p-3.5">{s.issue_date || "تاريخ الإصدار"}</th>
                    <th className="p-3.5">{s.subtotal || "الخاضع للضريبة"}</th>
                    <th className="p-3.5">{s.tax_amount || "الضريبة (15%)"}</th>
                    <th className="p-3.5">{s.total_amount || "الإجمالي"}</th>
                    <th className="p-3.5">{s.status || "الحالة"}</th>
                    <th className="p-3.5 text-center">{s.actions || "الإجراءات"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredInvoices.map((inv) => (
                    <tr key={inv.id} className="hover:bg-muted/30 transition">
                      <td className="p-3.5 font-mono font-semibold text-foreground">
                        {inv.invoiceNumber}
                      </td>
                      <td className="p-3.5">
                        <span
                          className={`px-2 py-0.5 text-xs rounded-md font-semibold ${
                            inv.invoiceType === "STANDARD"
                              ? "bg-blue-500/10 text-blue-600 border border-blue-500/20"
                              : "bg-purple-500/10 text-purple-600 border border-purple-500/20"
                          }`}
                        >
                          {inv.invoiceType === "STANDARD"
                            ? (s.type_standard || "ضريبية B2B")
                            : (s.type_simplified || "مبسطة B2C")}
                        </span>
                      </td>
                      <td className="p-3.5">
                        <div className="font-medium text-foreground">
                          {isAr ? inv.customerNameAr : inv.customerNameEn}
                        </div>
                        {inv.customerTaxId && (
                          <div className="text-xs text-muted-foreground font-mono">
                            VAT: {inv.customerTaxId}
                          </div>
                        )}
                      </td>
                      <td className="p-3.5 text-muted-foreground">{inv.issueDate}</td>
                      <td className="p-3.5 font-mono">
                        {inv.subtotal.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3.5 font-mono text-muted-foreground">
                        {inv.taxAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3.5 font-mono font-bold text-foreground">
                        {inv.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                      </td>
                      <td className="p-3.5">
                        <span
                          className={`px-2.5 py-1 text-xs rounded-full font-semibold inline-flex items-center gap-1.5 ${
                            inv.status === "POSTED"
                              ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                              : "bg-amber-500/10 text-amber-600 border border-amber-500/20"
                          }`}
                        >
                          {inv.status === "POSTED" ? (
                            <>
                              <CheckCircle2 className="h-3 w-3" />
                              <span>{s.status_posted || "مرحلة"}</span>
                            </>
                          ) : (
                            <>
                              <Clock className="h-3 w-3" />
                              <span>{s.status_draft || "مسودة"}</span>
                            </>
                          )}
                        </span>
                      </td>
                      <td className="p-3.5 text-center">
                        <div className="flex items-center justify-center gap-1.5">
                          {inv.status === "DRAFT" ? (
                            <button
                              onClick={() => handlePostInvoice(inv.id)}
                              className="px-2.5 py-1 text-xs rounded-md bg-emerald-600 text-white font-semibold hover:bg-emerald-700 transition"
                            >
                              {s.post_invoice || "ترحيل واعتماد"}
                            </button>
                          ) : (
                            <button
                              onClick={() =>
                                setSelectedQR({
                                  number: inv.invoiceNumber,
                                  qr: inv.qrCodeData || "ZATCA-QR-SAMPLE",
                                })
                              }
                              className="p-1.5 rounded-lg border hover:bg-muted text-primary transition"
                              title={s.view_qr || "معاينة رمز الاستجابة السريع"}
                            >
                              <QrCode className="h-4 w-4" />
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

      {/* Tab Content: ORDERS */}
      {activeTab === "orders" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right rtl:text-right ltr:text-left">
                <thead className="bg-muted/50 text-muted-foreground text-xs font-semibold border-b">
                  <tr>
                    <th className="p-3.5">رقم أمر البيع</th>
                    <th className="p-3.5">العميل</th>
                    <th className="p-3.5">تاريخ الأمر</th>
                    <th className="p-3.5">إجمالي القيمة</th>
                    <th className="p-3.5">نسبة التسليم</th>
                    <th className="p-3.5">نسبة الفوترة</th>
                    <th className="p-3.5">الحالة</th>
                    <th className="p-3.5 text-center">الإجراءات</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {orders.map((ord) => (
                    <tr key={ord.id} className="hover:bg-muted/30 transition">
                      <td className="p-3.5 font-mono font-semibold text-foreground">
                        {ord.orderNumber}
                      </td>
                      <td className="p-3.5 font-medium text-foreground">
                        {isAr ? ord.customerNameAr : ord.customerNameEn}
                      </td>
                      <td className="p-3.5 text-muted-foreground">{ord.orderDate}</td>
                      <td className="p-3.5 font-mono font-bold text-foreground">
                        {ord.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                      </td>
                      <td className="p-3.5">
                        <div className="w-24 bg-muted rounded-full h-2 overflow-hidden">
                          <div
                            className="bg-emerald-500 h-2 rounded-full"
                            style={{ width: `${ord.deliveredPercent}%` }}
                          />
                        </div>
                        <span className="text-xs text-muted-foreground mt-0.5 block font-mono">
                          {ord.deliveredPercent}%
                        </span>
                      </td>
                      <td className="p-3.5">
                        <div className="w-24 bg-muted rounded-full h-2 overflow-hidden">
                          <div
                            className="bg-blue-500 h-2 rounded-full"
                            style={{ width: `${ord.invoicedPercent}%` }}
                          />
                        </div>
                        <span className="text-xs text-muted-foreground mt-0.5 block font-mono">
                          {ord.invoicedPercent}%
                        </span>
                      </td>
                      <td className="p-3.5">
                        <span className="px-2 py-0.5 text-xs rounded-full font-semibold bg-primary/10 text-primary">
                          {ord.status}
                        </span>
                      </td>
                      <td className="p-3.5 text-center">
                        {ord.status === "DRAFT" ? (
                          <button
                            onClick={() => handleConfirmOrder(ord.id)}
                            className="px-2.5 py-1 text-xs rounded-md bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition"
                          >
                            {s.confirm_order || "تأكيد وفحص الائتمان"}
                          </button>
                        ) : ord.invoicedPercent < 100 ? (
                          <button
                            onClick={() => notify(isAr ? "تم إنشاء الفاتورة من أمر البيع" : "Invoice created from order")}
                            className="px-2.5 py-1 text-xs rounded-md border hover:bg-muted text-foreground transition font-semibold"
                          >
                            {s.create_invoice || "تحويل لفاتورة"}
                          </button>
                        ) : (
                          <span className="text-xs text-emerald-600 font-medium inline-flex items-center gap-1">
                            <CheckCircle2 className="h-3.5 w-3.5" />
                            مكتمل
                          </span>
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

      {/* Tab Content: QUOTATIONS */}
      {activeTab === "quotations" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right rtl:text-right ltr:text-left">
                <thead className="bg-muted/50 text-muted-foreground text-xs font-semibold border-b">
                  <tr>
                    <th className="p-3.5">رقم عرض السعر</th>
                    <th className="p-3.5">العميل</th>
                    <th className="p-3.5">تاريخ العرض</th>
                    <th className="p-3.5">ساري حتى</th>
                    <th className="p-3.5">إجمالي القيمة</th>
                    <th className="p-3.5">الحالة</th>
                    <th className="p-3.5 text-center">الإجراءات</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {quotations.map((q) => (
                    <tr key={q.id} className="hover:bg-muted/30 transition">
                      <td className="p-3.5 font-mono font-semibold text-foreground">
                        {q.quotationNumber}
                      </td>
                      <td className="p-3.5 font-medium text-foreground">
                        {isAr ? q.customerNameAr : q.customerNameEn}
                      </td>
                      <td className="p-3.5 text-muted-foreground">{q.quoteDate}</td>
                      <td className="p-3.5 text-muted-foreground">{q.validUntil}</td>
                      <td className="p-3.5 font-mono font-bold text-foreground">
                        {q.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                      </td>
                      <td className="p-3.5">
                        <span className="px-2 py-0.5 text-xs rounded-full font-semibold bg-muted text-muted-foreground">
                          {q.status}
                        </span>
                      </td>
                      <td className="p-3.5 text-center">
                        {q.status !== "CONVERTED" ? (
                          <button
                            onClick={() => handleConvertQuote(q.id)}
                            className="px-3 py-1 text-xs rounded-md bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition"
                          >
                            تحويل لأمر بيع
                          </button>
                        ) : (
                          <span className="text-xs text-emerald-600 font-medium">مُحوّل</span>
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

      {/* Tab Content: DELIVERIES */}
      {activeTab === "deliveries" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right rtl:text-right ltr:text-left">
                <thead className="bg-muted/50 text-muted-foreground text-xs font-semibold border-b">
                  <tr>
                    <th className="p-3.5">رقم مذكرة التسليم</th>
                    <th className="p-3.5">رقم أمر البيع</th>
                    <th className="p-3.5">العميل</th>
                    <th className="p-3.5">المستودع المصدر</th>
                    <th className="p-3.5">تاريخ التسليم</th>
                    <th className="p-3.5">عدد الأصناف</th>
                    <th className="p-3.5">الحالة</th>
                    <th className="p-3.5 text-center">الإجراءات</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {deliveries.map((del) => (
                    <tr key={del.id} className="hover:bg-muted/30 transition">
                      <td className="p-3.5 font-mono font-semibold text-foreground">
                        {del.deliveryNumber}
                      </td>
                      <td className="p-3.5 font-mono text-muted-foreground">{del.orderNumber}</td>
                      <td className="p-3.5 font-medium text-foreground">
                        {isAr ? del.customerNameAr : del.customerNameEn}
                      </td>
                      <td className="p-3.5 text-muted-foreground">{del.warehouseName}</td>
                      <td className="p-3.5 text-muted-foreground">{del.deliveryDate}</td>
                      <td className="p-3.5 font-mono">{del.itemsCount} صنف</td>
                      <td className="p-3.5">
                        <span
                          className={`px-2 py-0.5 text-xs rounded-full font-semibold ${
                            del.status === "DELIVERED"
                              ? "bg-emerald-500/10 text-emerald-600"
                              : "bg-amber-500/10 text-amber-600"
                          }`}
                        >
                          {del.status}
                        </span>
                      </td>
                      <td className="p-3.5 text-center">
                        {del.status !== "DELIVERED" ? (
                          <button
                            onClick={() => notify(isAr ? "تم صرف البضاعة من المخزن وتسجيل حركة OUT_SALES" : "Stock issued with OUT_SALES movement")}
                            className="px-2.5 py-1 text-xs rounded-md bg-emerald-600 text-white font-semibold hover:bg-emerald-700 transition"
                          >
                            صرف وتسليم
                          </button>
                        ) : (
                          <span className="text-xs text-emerald-600 font-medium inline-flex items-center gap-1">
                            <CheckCircle2 className="h-3.5 w-3.5" />
                            مُسلّم ومخصوم
                          </span>
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

      {/* Tab Content: CREDIT NOTES */}
      {activeTab === "credit_notes" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right rtl:text-right ltr:text-left">
                <thead className="bg-muted/50 text-muted-foreground text-xs font-semibold border-b">
                  <tr>
                    <th className="p-3.5">رقم إشعار الدائن</th>
                    <th className="p-3.5">الفاتورة الأصلية</th>
                    <th className="p-3.5">العميل</th>
                    <th className="p-3.5">تاريخ الإصدار</th>
                    <th className="p-3.5">السبب</th>
                    <th className="p-3.5">إجمالي المبلغ</th>
                    <th className="p-3.5">الحالة</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {creditNotes.map((cn) => (
                    <tr key={cn.id} className="hover:bg-muted/30 transition">
                      <td className="p-3.5 font-mono font-semibold text-foreground">
                        {cn.creditNoteNumber}
                      </td>
                      <td className="p-3.5 font-mono text-muted-foreground">{cn.originalInvoiceNumber}</td>
                      <td className="p-3.5 font-medium text-foreground">
                        {isAr ? cn.customerNameAr : cn.customerNameEn}
                      </td>
                      <td className="p-3.5 text-muted-foreground">{cn.issueDate}</td>
                      <td className="p-3.5">
                        <span className="px-2 py-0.5 text-xs rounded bg-muted text-muted-foreground font-medium">
                          {cn.reason === "RETURN" ? "مردودات مبيعات" : "تسوية وخصم"}
                        </span>
                      </td>
                      <td className="p-3.5 font-mono font-bold text-foreground">
                        {cn.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                      </td>
                      <td className="p-3.5">
                        <span className="px-2.5 py-1 text-xs rounded-full font-semibold bg-purple-500/10 text-purple-600 border border-purple-500/20">
                          {cn.status}
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

      {/* QR Code Modal */}
      {selectedQR && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl p-6 max-w-sm w-full shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="font-bold text-base text-foreground flex items-center gap-2">
                <QrCode className="h-5 w-5 text-primary" />
                <span>رمز الاستجابة السريع (ZATCA QR)</span>
              </h3>
              <button
                onClick={() => setSelectedQR(null)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="flex flex-col items-center justify-center p-6 bg-muted/40 rounded-xl border border-dashed">
              <div className="w-48 h-48 bg-white p-3 rounded-lg shadow-inner flex flex-col items-center justify-center border">
                <div className="w-full h-full bg-slate-900 rounded flex items-center justify-center text-white text-xs font-mono p-2 text-center break-all">
                  [ZATCA Phase 2 Cryptographic TLV QR: {selectedQR.number}]
                </div>
              </div>
              <span className="mt-3 text-xs font-mono text-muted-foreground font-semibold">
                {selectedQR.number}
              </span>
            </div>

            <p className="text-xs text-muted-foreground leading-relaxed text-center">
              تم تشفير الرمز طبقاً لمواصفات هيئة الزكاة والضريبة والجمارك متضمناً اسم المورد، الرقم الضريبي، الختم الزمني، وإجمالي الضريبة.
            </p>

            <button
              onClick={() => setSelectedQR(null)}
              className="w-full py-2 rounded-xl bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition text-sm"
            >
              إغلاق
            </button>
          </div>
        </div>
      )}

      {/* New Invoice Modal */}
      {showInvoiceModal && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl p-6 max-w-lg w-full shadow-2xl space-y-5">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="font-bold text-lg text-foreground flex items-center gap-2">
                <FileText className="h-5 w-5 text-primary" />
                <span>{s.new_invoice || "إصدار فاتورة ضريبية جديدة"}</span>
              </h3>
              <button
                onClick={() => setShowInvoiceModal(false)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleCreateInvoice} className="space-y-4">
              {/* Type Switcher */}
              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                  {s.invoice_type || "نوع الفاتورة الإلكترونية"}
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setInvType("STANDARD")}
                    className={`py-2 px-3 rounded-xl text-xs font-bold border transition ${
                      invType === "STANDARD"
                        ? "bg-primary text-primary-foreground border-primary"
                        : "bg-background text-muted-foreground border-input hover:bg-muted"
                    }`}
                  >
                    فاتورة ضريبية قياسية (B2B)
                  </button>
                  <button
                    type="button"
                    onClick={() => setInvType("SIMPLIFIED")}
                    className={`py-2 px-3 rounded-xl text-xs font-bold border transition ${
                      invType === "SIMPLIFIED"
                        ? "bg-primary text-primary-foreground border-primary"
                        : "bg-background text-muted-foreground border-input hover:bg-muted"
                    }`}
                  >
                    فاتورة ضريبية مبسطة (B2C)
                  </button>
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                  اسم العميل
                </label>
                <input
                  type="text"
                  required
                  value={invCustomer}
                  onChange={(e) => setInvCustomer(e.target.value)}
                  className="w-full px-3 py-2 rounded-xl border bg-background text-foreground text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              {invType === "STANDARD" && (
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                    الرقم الضريبي للعميل (15 رقماً إلزامي لـ B2B)
                  </label>
                  <input
                    type="text"
                    required
                    pattern="^3[0-9]{13}3$"
                    placeholder="310123456700003"
                    value={invTaxId}
                    onChange={(e) => setInvTaxId(e.target.value)}
                    className="w-full px-3 py-2 rounded-xl border bg-background text-foreground text-sm font-mono focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
              )}

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div className="sm:col-span-2">
                  <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                    بيان الصنف والخدمة
                  </label>
                  <input
                    type="text"
                    required
                    value={invItem}
                    onChange={(e) => setInvItem(e.target.value)}
                    className="w-full px-3 py-2 rounded-xl border bg-background text-foreground text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                    الكمية
                  </label>
                  <input
                    type="number"
                    min="1"
                    required
                    value={invQty}
                    onChange={(e) => setInvQty(Number(e.target.value))}
                    className="w-full px-3 py-2 rounded-xl border bg-background text-foreground text-sm font-mono focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-1 block">
                  سعر الوحدة (ر.س)
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  required
                  value={invPrice}
                  onChange={(e) => setInvPrice(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-xl border bg-background text-foreground text-sm font-mono focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              {/* Total Calculation Card */}
              <div className="p-3.5 rounded-xl bg-muted/50 border space-y-1.5 text-xs">
                <div className="flex justify-between text-muted-foreground">
                  <span>المبلغ الخاضع للضريبة:</span>
                  <span className="font-mono font-semibold text-foreground">
                    {(invQty * invPrice).toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                  </span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>ضريبة القيمة المضافة (15%):</span>
                  <span className="font-mono font-semibold text-foreground">
                    {(invQty * invPrice * 0.15).toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                  </span>
                </div>
                <div className="flex justify-between border-t pt-1.5 font-bold text-sm text-foreground">
                  <span>الإجمالي النهائي المستحق:</span>
                  <span className="font-mono text-primary">
                    {(invQty * invPrice * 1.15).toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                  </span>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2.5 pt-2">
                <button
                  type="button"
                  onClick={() => setShowInvoiceModal(false)}
                  className="px-4 py-2 rounded-xl border hover:bg-muted text-foreground transition text-sm"
                >
                  {s.cancel || "إلغاء"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition text-sm shadow-sm"
                >
                  {s.save || "حفظ واعتماد الفاتورة"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

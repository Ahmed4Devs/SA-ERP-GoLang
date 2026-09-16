"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  ShoppingCart,
  FileText,
  Truck,
  CheckCircle2,
  Clock,
  AlertTriangle,
  Plus,
  Search,
  Layers,
  ArrowRight,
  ArrowLeft,
  DollarSign,
  Package,
  Calendar,
  Building2,
  Check,
  Percent,
  X,
  Scale,
} from "lucide-react";

interface PurchaseRequestItem {
  id: string;
  prNumber: string;
  requesterName: string;
  department: string;
  requestDate: string;
  requiredDate: string;
  purpose: string;
  estimatedTotal: number;
  currency: string;
  status: "DRAFT" | "SUBMITTED" | "APPROVED" | "ORDERED";
}

interface RFQItem {
  id: string;
  rfqNumber: string;
  title: string;
  rfqDate: string;
  closingDate: string;
  quotesCount: number;
  bestBidSupplier: string;
  bestBidAmount: number;
  status: "DRAFT" | "SENT" | "CLOSED";
}

interface PurchaseOrderItem {
  id: string;
  poNumber: string;
  supplierNameAr: string;
  supplierNameEn: string;
  orderDate: string;
  deliveryDate: string;
  warehouseName: string;
  subtotal: number;
  taxAmount: number;
  totalAmount: number;
  status: "DRAFT" | "APPROVED" | "ISSUED" | "RECEIVED";
}

interface GoodsReceiptItem {
  id: string;
  receiptNumber: string;
  poNumber: string;
  supplierNameAr: string;
  receiptDate: string;
  deliveryNote: string;
  receivedQuantity: number;
  acceptedQuantity: number;
  rejectedQuantity: number;
  status: "DRAFT" | "ACCEPTED" | "REJECTED";
}

interface MatchingSimulation {
  poNumber: string;
  poQty: number;
  poUnitPrice: number;
  grnNumber: string;
  grnAcceptedQty: number;
  invoiceNumber: string;
  invoiceQty: number;
  invoiceUnitPrice: number;
  qtyTolerancePercent: number;
  priceTolerancePercent: number;
}

const INITIAL_REQUESTS: PurchaseRequestItem[] = [
  {
    id: "pr-1",
    prNumber: "PR-2026-0001",
    requesterName: "م. أحمد الغامدي (مهندس الموقع)",
    department: "إدارة العمليات والمشاريع",
    requestDate: "2026-09-10",
    requiredDate: "2026-09-24",
    purpose: "مستلزمات ومواد تشغيلية دورية للربع الحالي",
    estimatedTotal: 5000.0,
    currency: "SAR",
    status: "APPROVED",
  },
  {
    id: "pr-2",
    prNumber: "PR-2026-0002",
    requesterName: "سارة العتيبي (إدارة المشتريات)",
    department: "الإدارة العامة",
    requestDate: "2026-09-14",
    requiredDate: "2026-09-28",
    purpose: "أجهزة حاسب آلي وتجهيزات مكتبية وتراخيص برمجية",
    estimatedTotal: 18500.0,
    currency: "SAR",
    status: "SUBMITTED",
  },
];

const INITIAL_RFQS: RFQItem[] = [
  {
    id: "rfq-1",
    rfqNumber: "RFQ-2026-0001",
    title: "توريد معدات وأدوات أمن وسلامة معتمدة",
    rfqDate: "2026-09-11",
    closingDate: "2026-09-18",
    quotesCount: 3,
    bestBidSupplier: "شركة الفوزان لمواد البناء",
    bestBidAmount: 22000.0,
    status: "SENT",
  },
];

const INITIAL_ORDERS: PurchaseOrderItem[] = [
  {
    id: "po-1",
    poNumber: "PO-2026-0001",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    supplierNameEn: "Al-Fozan Building Materials Co.",
    orderDate: "2026-09-12",
    deliveryDate: "2026-09-22",
    warehouseName: "مستودع الرياض المركزي (WH-CENTRAL-01)",
    subtotal: 5000.0,
    taxAmount: 750.0,
    totalAmount: 5750.0,
    status: "ISSUED",
  },
  {
    id: "po-2",
    poNumber: "PO-2026-0002",
    supplierNameAr: "الشركة السعودية للكهرباء (SEC)",
    supplierNameEn: "Saudi Electricity Company",
    orderDate: "2026-09-15",
    deliveryDate: "2026-09-25",
    warehouseName: "مستودع الرياض المركزي (WH-CENTRAL-01)",
    subtotal: 12000.0,
    taxAmount: 1800.0,
    totalAmount: 13800.0,
    status: "APPROVED",
  },
];

const INITIAL_RECEIPTS: GoodsReceiptItem[] = [
  {
    id: "grn-1",
    receiptNumber: "GRN-2026-0001",
    poNumber: "PO-2026-0001",
    supplierNameAr: "شركة الفوزان لمواد البناء",
    receiptDate: "2026-09-16",
    deliveryNote: "DN-FOZAN-88991",
    receivedQuantity: 100,
    acceptedQuantity: 100,
    rejectedQuantity: 0,
    status: "ACCEPTED",
  },
];

export function PurchasingView() {
  const { dir, t, locale } = useLocale();
  const p = (t as any).purchasing || {};

  const [activeTab, setActiveTab] = useState<
    "overview" | "requests" | "rfqs" | "orders" | "receipts" | "matching"
  >("overview");

  const [requests, setRequests] = useState<PurchaseRequestItem[]>(INITIAL_REQUESTS);
  const [orders, setOrders] = useState<PurchaseOrderItem[]>(INITIAL_ORDERS);
  const [receipts, setReceipts] = useState<GoodsReceiptItem[]>(INITIAL_RECEIPTS);

  // New PR Modal
  const [showNewPRModal, setShowNewPRModal] = useState(false);
  const [newPRPurpose, setNewPRPurpose] = useState("");
  const [newPRAmount, setNewPRAmount] = useState("5000");

  // 3-Way Match interactive tool state
  const [matchSim, setMatchSim] = useState<MatchingSimulation>({
    poNumber: "PO-2026-0001",
    poQty: 100,
    poUnitPrice: 50.0,
    grnNumber: "GRN-2026-0001",
    grnAcceptedQty: 100,
    invoiceNumber: "PINV-2026-001",
    invoiceQty: 100,
    invoiceUnitPrice: 50.0,
    qtyTolerancePercent: 0,
    priceTolerancePercent: 2,
  });

  const [matchResult, setMatchResult] = useState<{
    status: "PERFECT" | "PRICE_VAR" | "QTY_VAR" | "BOTH_VAR";
    qtyDiff: number;
    priceDiff: number;
    passed: boolean;
  } | null>(null);

  // Handlers
  const handleApprovePR = (id: string) => {
    setRequests((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: "APPROVED" } : r))
    );
  };

  const handleApprovePO = (id: string) => {
    setOrders((prev) =>
      prev.map((o) => (o.id === id ? { ...o, status: "APPROVED" } : o))
    );
  };

  const handleIssuePO = (id: string) => {
    setOrders((prev) =>
      prev.map((o) => (o.id === id ? { ...o, status: "ISSUED" } : o))
    );
  };

  const handleInspectAcceptGRN = (id: string) => {
    setReceipts((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: "ACCEPTED" } : r))
    );
  };

  const handleRunMatch = () => {
    const qtyDiff = matchSim.invoiceQty - matchSim.grnAcceptedQty;
    const priceDiff = matchSim.invoiceUnitPrice - matchSim.poUnitPrice;

    const maxPriceDiff = (matchSim.poUnitPrice * matchSim.priceTolerancePercent) / 100;
    const pricePassed = Math.abs(priceDiff) <= maxPriceDiff;
    const qtyPassed = qtyDiff === 0;

    let status: "PERFECT" | "PRICE_VAR" | "QTY_VAR" | "BOTH_VAR" = "PERFECT";
    if (qtyDiff !== 0 && priceDiff !== 0) status = "BOTH_VAR";
    else if (qtyDiff !== 0) status = "QTY_VAR";
    else if (priceDiff !== 0) status = "PRICE_VAR";

    setMatchResult({
      status,
      qtyDiff,
      priceDiff,
      passed: pricePassed && qtyPassed,
    });
  };

  const handleCreatePR = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPRPurpose) return;
    const newPR: PurchaseRequestItem = {
      id: `pr-${Date.now()}`,
      prNumber: `PR-2026-000${requests.length + 1}`,
      requesterName: "مدير المشتريات (النظام)",
      department: "الإدارة العامة",
      requestDate: new Date().toISOString().split("T")[0],
      requiredDate: new Date(Date.now() + 14 * 86400000).toISOString().split("T")[0],
      purpose: newPRPurpose,
      estimatedTotal: parseFloat(newPRAmount) || 0,
      currency: "SAR",
      status: "DRAFT",
    };
    setRequests([newPR, ...requests]);
    setShowNewPRModal(false);
    setNewPRPurpose("");
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary shadow-sm">
              <ShoppingCart className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-foreground">
                {p.title || "إدارة المشتريات والتعميد (Procure-to-Pay)"}
              </h1>
              <p className="text-sm text-muted-foreground mt-0.5">
                {p.description ||
                  "دورة المشتريات المتكاملة: طلبات الشراء، استدراج العروض، أوامر التعميد، استلام وفحص البضائع، والمطابقة الثلاثية"}
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowNewPRModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition shadow-sm"
          >
            <Plus className="h-4 w-4" />
            {p.btn_new_request || "طلب شراء جديد"}
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-medium">
            <span>{p.kpi_active_pos || "أوامر الشراء النشطة"}</span>
            <ShoppingCart className="h-4 w-4 text-blue-500" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {orders.length}
            </span>
            <span className="text-xs text-muted-foreground">أمر تعميد</span>
          </div>
          <p className="mt-1 text-[11px] text-muted-foreground">
            إجمالي التعميدات: 19,550.00 ر.س
          </p>
        </div>

        <div className="p-4 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-medium">
            <span>{p.kpi_pending_grn || "سندات قيد الفحص (GRN)"}</span>
            <Truck className="h-4 w-4 text-amber-500" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {receipts.filter((r) => r.status === "DRAFT").length || 1}
            </span>
            <span className="text-xs text-amber-600 dark:text-amber-400 font-medium">
              بانتظار قبول المستودع
            </span>
          </div>
          <p className="mt-1 text-[11px] text-muted-foreground">
            تحديث أرصدة المخزون تلقائياً
          </p>
        </div>

        <div className="p-4 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-medium">
            <span>{p.kpi_pending_requests || "طلبات شراء قيد الاعتماد"}</span>
            <Clock className="h-4 w-4 text-indigo-500" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {requests.filter((r) => r.status === "SUBMITTED").length}
            </span>
            <span className="text-xs text-muted-foreground">طلب داخلي</span>
          </div>
          <p className="mt-1 text-[11px] text-muted-foreground">
            تتطلب اعتماد مدير الإدارة
          </p>
        </div>

        <div className="p-4 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-medium">
            <span>{p.kpi_match_clearance || "نسبة المطابقة الثلاثية"}</span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-emerald-600 dark:text-emerald-400">
              100%
            </span>
            <span className="text-xs text-emerald-600/80 font-medium">مطابقة تامة</span>
          </div>
          <p className="mt-1 text-[11px] text-muted-foreground">
            أوامر الشراء vs الاستلام vs الفواتير
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-border/60 overflow-x-auto pb-px">
        {[
          { key: "overview", label: p.tab_overview || "نظرة عامة والمسار", icon: Layers },
          { key: "requests", label: p.tab_requests || "طلبات الشراء (PR)", icon: FileText },
          { key: "rfqs", label: p.tab_rfqs || "عروض الأسعار (RFQ)", icon: Scale },
          { key: "orders", label: p.tab_orders || "أوامر الشراء (PO)", icon: ShoppingCart },
          { key: "receipts", label: p.tab_receipts || "استلام البضائع (GRN)", icon: Truck },
          { key: "matching", label: p.tab_matching || "المطابقة الثلاثية", icon: CheckCircle2 },
        ].map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.key;
          return (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key as any)}
              className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition whitespace-nowrap ${
                isActive
                  ? "border-primary text-primary"
                  : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
              }`}
            >
              <Icon className="h-4 w-4" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab 1: Overview & P2P Stepper */}
      {activeTab === "overview" && (
        <div className="space-y-6">
          {/* P2P Lifecycle Stepper */}
          <div className="p-6 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm">
            <h2 className="text-base font-semibold text-foreground mb-4">
              مسار دورة الشراء والدفع المتكاملة (Procure-to-Pay Lifecycle)
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
              {[
                {
                  step: "1",
                  title: "طلب شراء (PR)",
                  desc: "تقديم طلب الاحتياج واعتماده داخلياً",
                  status: "مكتمل",
                  active: false,
                },
                {
                  step: "2",
                  title: "استدراج العروض (RFQ)",
                  desc: "دعوة الموردين والمقارنة التنافسية",
                  status: "مكتمل",
                  active: false,
                },
                {
                  step: "3",
                  title: "أمر الشراء (PO)",
                  desc: "إصدار التعميد الرسمي وتثبيت الأسعار",
                  status: "نشط",
                  active: true,
                },
                {
                  step: "4",
                  title: "استلام وفحص (GRN)",
                  desc: "الفحص الفني وزيادة المخزون الفعلي",
                  status: "قيد التنفيذ",
                  active: false,
                },
                {
                  step: "5",
                  title: "المطابقة الثلاثية & AP",
                  desc: "مطابقة PO و GRN وفاتورة المورد",
                  status: "جاهز",
                  active: false,
                },
              ].map((item, idx) => (
                <div
                  key={idx}
                  className={`p-4 rounded-lg border transition ${
                    item.active
                      ? "border-primary/50 bg-primary/5 ring-1 ring-primary/20"
                      : "border-border/40 bg-muted/20"
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span
                      className={`h-6 w-6 rounded-full flex items-center justify-center text-xs font-bold ${
                        item.active
                          ? "bg-primary text-primary-foreground"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      {item.step}
                    </span>
                    <span className="text-[10px] font-medium px-2 py-0.5 rounded bg-muted/60 text-muted-foreground">
                      {item.status}
                    </span>
                  </div>
                  <div className="font-medium text-sm text-foreground mb-1">
                    {item.title}
                  </div>
                  <div className="text-xs text-muted-foreground leading-relaxed">
                    {item.desc}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Quick Actions & Recent Summary */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="p-6 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm space-y-4">
              <h3 className="text-sm font-semibold text-foreground">
                إجراءات الشراء السريعة
              </h3>
              <div className="grid grid-cols-2 gap-3">
                <button
                  onClick={() => {
                    setActiveTab("requests");
                    setShowNewPRModal(true);
                  }}
                  className="p-3 text-start rounded-lg border border-border/60 hover:border-primary/50 hover:bg-primary/5 transition group"
                >
                  <FileText className="h-5 w-5 text-primary mb-2" />
                  <div className="text-sm font-medium text-foreground">
                    إنشاء طلب شراء
                  </div>
                  <div className="text-[11px] text-muted-foreground">
                    لأقسام الشركة والمشاريع
                  </div>
                </button>
                <button
                  onClick={() => setActiveTab("orders")}
                  className="p-3 text-start rounded-lg border border-border/60 hover:border-primary/50 hover:bg-primary/5 transition group"
                >
                  <ShoppingCart className="h-5 w-5 text-blue-500 mb-2" />
                  <div className="text-sm font-medium text-foreground">
                    إصدار أمر تعميد
                  </div>
                  <div className="text-[11px] text-muted-foreground">
                    مباشر للموردين المعتمدين
                  </div>
                </button>
                <button
                  onClick={() => setActiveTab("receipts")}
                  className="p-3 text-start rounded-lg border border-border/60 hover:border-primary/50 hover:bg-primary/5 transition group"
                >
                  <Truck className="h-5 w-5 text-amber-500 mb-2" />
                  <div className="text-sm font-medium text-foreground">
                    سند استلام بضاعة
                  </div>
                  <div className="text-[11px] text-muted-foreground">
                    فحص الجودة والمطابقة
                  </div>
                </button>
                <button
                  onClick={() => setActiveTab("matching")}
                  className="p-3 text-start rounded-lg border border-border/60 hover:border-primary/50 hover:bg-primary/5 transition group"
                >
                  <CheckCircle2 className="h-5 w-5 text-emerald-500 mb-2" />
                  <div className="text-sm font-medium text-foreground">
                    المطابقة الثلاثية
                  </div>
                  <div className="text-[11px] text-muted-foreground">
                    فحص الفروقات قبل الصرف
                  </div>
                </button>
              </div>
            </div>

            <div className="p-6 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm space-y-3">
              <h3 className="text-sm font-semibold text-foreground">
                الربط المالي والمحاسبي (GR/IR Clearing)
              </h3>
              <div className="space-y-2.5 text-xs">
                <div className="p-3 rounded-lg bg-muted/40 border border-border/40">
                  <div className="font-medium text-foreground mb-1">
                    1. عند استلام البضاعة بالمستودع (GRN)
                  </div>
                  <div className="text-muted-foreground flex justify-between">
                    <span>من حـ/ المخزون (113000)</span>
                    <span className="font-mono text-emerald-600 dark:text-emerald-400">
                      مدين (DR)
                    </span>
                  </div>
                  <div className="text-muted-foreground flex justify-between mt-1">
                    <span>إلى حـ/ وسيط مشتريات مستلمة غير مفوترة (212000)</span>
                    <span className="font-mono text-blue-600 dark:text-blue-400">
                      دائن (CR)
                    </span>
                  </div>
                </div>

                <div className="p-3 rounded-lg bg-muted/40 border border-border/40">
                  <div className="font-medium text-foreground mb-1">
                    2. عند قيد فاتورة المورد في قسم الدائنين (AP)
                  </div>
                  <div className="text-muted-foreground flex justify-between">
                    <span>من حـ/ وسيط مشتريات (212000) + ضريبة المدخلات (114000)</span>
                    <span className="font-mono text-emerald-600 dark:text-emerald-400">
                      مدين (DR)
                    </span>
                  </div>
                  <div className="text-muted-foreground flex justify-between mt-1">
                    <span>إلى حـ/ ذمم الموردين (211000)</span>
                    <span className="font-mono text-blue-600 dark:text-blue-400">
                      دائن (CR)
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Requisitions */}
      {activeTab === "requests" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-semibold text-foreground">
              طلبات الشراء الداخلية (Requisitions)
            </h2>
            <button
              onClick={() => setShowNewPRModal(true)}
              className="flex items-center gap-2 px-3 py-1.5 bg-primary text-primary-foreground text-xs font-medium rounded-lg hover:bg-primary/90 transition"
            >
              <Plus className="h-3.5 w-3.5" />
              {p.btn_new_request || "طلب شراء جديد"}
            </button>
          </div>

          <div className="rounded-xl border border-border/60 bg-card overflow-hidden shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs font-medium text-muted-foreground border-b border-border/60">
                <tr>
                  <th className="px-4 py-3 text-start">{p.col_pr_number || "رقم الطلب"}</th>
                  <th className="px-4 py-3 text-start">{p.col_requester || "مقدم الطلب"}</th>
                  <th className="px-4 py-3 text-start">القسم</th>
                  <th className="px-4 py-3 text-start">{p.col_date || "تاريخ الطلب"}</th>
                  <th className="px-4 py-3 text-start">{p.col_required_date || "تاريخ الاحتياج"}</th>
                  <th className="px-4 py-3 text-end">{p.col_amount || "المبلغ التقديري"}</th>
                  <th className="px-4 py-3 text-center">{p.col_status || "الحالة"}</th>
                  <th className="px-4 py-3 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {requests.map((r) => (
                  <tr key={r.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-medium text-primary">
                      {r.prNumber}
                    </td>
                    <td className="px-4 py-3 text-foreground font-medium">
                      {r.requesterName}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{r.department}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {r.requestDate}
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {r.requiredDate}
                    </td>
                    <td className="px-4 py-3 font-mono font-semibold text-end text-foreground">
                      {r.estimatedTotal.toLocaleString(undefined, {
                        minimumFractionDigits: 2,
                      })}{" "}
                      ر.س
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium ${
                          r.status === "APPROVED"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : r.status === "SUBMITTED"
                            ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                            : "bg-muted text-muted-foreground"
                        }`}
                      >
                        {r.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      {r.status === "SUBMITTED" && (
                        <button
                          onClick={() => handleApprovePR(r.id)}
                          className="px-2.5 py-1 bg-emerald-600 text-white rounded text-xs font-medium hover:bg-emerald-700 transition"
                        >
                          اعتماد الطلب
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 3: RFQs & Supplier Quotes */}
      {activeTab === "rfqs" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-semibold text-foreground">
              استدراج العروض وعروض أسعار الموردين (RFQs & Bids)
            </h2>
          </div>

          <div className="rounded-xl border border-border/60 bg-card overflow-hidden shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs font-medium text-muted-foreground border-b border-border/60">
                <tr>
                  <th className="px-4 py-3 text-start">رقم المنافسة (RFQ)</th>
                  <th className="px-4 py-3 text-start">العنوان / الموضوع</th>
                  <th className="px-4 py-3 text-start">تاريخ الطرح</th>
                  <th className="px-4 py-3 text-start">آخر موعد لتقديم العروض</th>
                  <th className="px-4 py-3 text-center">عدد العروض المستلمة</th>
                  <th className="px-4 py-3 text-start">أفضل عرض مالي</th>
                  <th className="px-4 py-3 text-center">الحالة</th>
                  <th className="px-4 py-3 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {INITIAL_RFQS.map((q) => (
                  <tr key={q.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-medium text-primary">
                      {q.rfqNumber}
                    </td>
                    <td className="px-4 py-3 text-foreground font-medium">{q.title}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {q.rfqDate}
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {q.closingDate}
                    </td>
                    <td className="px-4 py-3 text-center font-mono font-bold text-foreground">
                      {q.quotesCount} عروض
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-foreground">
                        {q.bestBidSupplier}
                      </div>
                      <div className="text-xs font-mono text-emerald-600 dark:text-emerald-400">
                        {q.bestBidAmount.toLocaleString()} ر.س
                      </div>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400">
                        {q.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <button className="px-2.5 py-1 bg-primary text-primary-foreground rounded text-xs font-medium hover:bg-primary/90 transition">
                        ترسية العرض
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 4: Purchase Orders */}
      {activeTab === "orders" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-semibold text-foreground">
              أوامر الشراء والتعميد (Purchase Orders)
            </h2>
          </div>

          <div className="rounded-xl border border-border/60 bg-card overflow-hidden shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs font-medium text-muted-foreground border-b border-border/60">
                <tr>
                  <th className="px-4 py-3 text-start">{p.col_po_number || "رقم التعميد"}</th>
                  <th className="px-4 py-3 text-start">{p.col_supplier || "المورد"}</th>
                  <th className="px-4 py-3 text-start">{p.col_date || "تاريخ التعميد"}</th>
                  <th className="px-4 py-3 text-start">{p.col_warehouse || "المستودع المستهدف"}</th>
                  <th className="px-4 py-3 text-end">المبلغ قبل الضريبة</th>
                  <th className="px-4 py-3 text-end">الضريبة (15%)</th>
                  <th className="px-4 py-3 text-end">{p.col_amount || "الإجمالي الشامل"}</th>
                  <th className="px-4 py-3 text-center">{p.col_status || "الحالة"}</th>
                  <th className="px-4 py-3 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {orders.map((o) => (
                  <tr key={o.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-medium text-primary">
                      {o.poNumber}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-medium text-foreground">
                        {locale === "ar" ? o.supplierNameAr : o.supplierNameEn}
                      </div>
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {o.orderDate}
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {o.warehouseName}
                    </td>
                    <td className="px-4 py-3 font-mono text-end text-muted-foreground">
                      {o.subtotal.toLocaleString(undefined, { minimumFractionDigits: 2 })}{" "}
                      ر.س
                    </td>
                    <td className="px-4 py-3 font-mono text-end text-muted-foreground">
                      {o.taxAmount.toLocaleString(undefined, { minimumFractionDigits: 2 })}{" "}
                      ر.س
                    </td>
                    <td className="px-4 py-3 font-mono font-bold text-end text-foreground">
                      {o.totalAmount.toLocaleString(undefined, {
                        minimumFractionDigits: 2,
                      })}{" "}
                      ر.س
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-medium ${
                          o.status === "ISSUED"
                            ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                            : o.status === "APPROVED"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : "bg-muted text-muted-foreground"
                        }`}
                      >
                        {o.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      {o.status === "APPROVED" && (
                        <button
                          onClick={() => handleIssuePO(o.id)}
                          className="px-2.5 py-1 bg-blue-600 text-white rounded text-xs font-medium hover:bg-blue-700 transition"
                        >
                          إصدار وتعميد المورد
                        </button>
                      )}
                      {o.status === "DRAFT" && (
                        <button
                          onClick={() => handleApprovePO(o.id)}
                          className="px-2.5 py-1 bg-emerald-600 text-white rounded text-xs font-medium hover:bg-emerald-700 transition"
                        >
                          اعتماد أمر الشراء
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 5: Goods Receipts */}
      {activeTab === "receipts" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-semibold text-foreground">
              سندات استلام وفحص البضائع (Goods Receipts Notes - GRN)
            </h2>
          </div>

          <div className="rounded-xl border border-border/60 bg-card overflow-hidden shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs font-medium text-muted-foreground border-b border-border/60">
                <tr>
                  <th className="px-4 py-3 text-start">{p.col_grn_number || "رقم السند"}</th>
                  <th className="px-4 py-3 text-start">{p.col_po_number || "أمر الشراء المرتبط"}</th>
                  <th className="px-4 py-3 text-start">{p.col_supplier || "المورد"}</th>
                  <th className="px-4 py-3 text-start">إشعار تسليم المورد (DN)</th>
                  <th className="px-4 py-3 text-start">{p.col_date || "تاريخ الاستلام"}</th>
                  <th className="px-4 py-3 text-center">{p.col_received_qty || "المستلم"}</th>
                  <th className="px-4 py-3 text-center">{p.col_accepted_qty || "المقبول بالمستودع"}</th>
                  <th className="px-4 py-3 text-center">{p.col_rejected_qty || "المرفوض"}</th>
                  <th className="px-4 py-3 text-center">{p.col_status || "الحالة"}</th>
                  <th className="px-4 py-3 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {receipts.map((r) => (
                  <tr key={r.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-medium text-primary">
                      {r.receiptNumber}
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{r.poNumber}</td>
                    <td className="px-4 py-3 text-foreground font-medium">
                      {r.supplierNameAr}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {r.deliveryNote}
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      {r.receiptDate}
                    </td>
                    <td className="px-4 py-3 font-mono text-center font-semibold text-foreground">
                      {r.receivedQuantity}
                    </td>
                    <td className="px-4 py-3 font-mono text-center font-semibold text-emerald-600 dark:text-emerald-400">
                      {r.acceptedQuantity}
                    </td>
                    <td className="px-4 py-3 font-mono text-center font-semibold text-red-600 dark:text-red-400">
                      {r.rejectedQuantity}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium ${
                          r.status === "ACCEPTED"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                        }`}
                      >
                        {r.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      {r.status === "DRAFT" && (
                        <button
                          onClick={() => handleInspectAcceptGRN(r.id)}
                          className="px-2.5 py-1 bg-emerald-600 text-white rounded text-xs font-medium hover:bg-emerald-700 transition"
                        >
                          فحص وقبول
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 6: Three-Way Matching Engine */}
      {activeTab === "matching" && (
        <div className="space-y-6">
          <div className="p-6 rounded-xl border border-border/60 bg-card/60 backdrop-blur shadow-sm space-y-6">
            <div>
              <h2 className="text-base font-semibold text-foreground">
                محرك المطابقة الثلاثية الذكي (Three-Way Matching Engine)
              </h2>
              <p className="text-xs text-muted-foreground mt-1">
                التحقق الصارم من مطابقة كميات وأسعار أمر الشراء (PO) وسند الاستلام الفعلي
                (GRN) وفاتورة المورد في قسم الدائنين (Supplier Invoice)
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Document 1: Purchase Order */}
              <div className="p-4 rounded-lg border border-border/60 bg-muted/20 space-y-3">
                <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                  <ShoppingCart className="h-4 w-4 text-primary" />
                  أمر الشراء (PO)
                </div>
                <div className="space-y-2 text-xs">
                  <div>
                    <label className="text-muted-foreground">رقم أمر الشراء:</label>
                    <input
                      type="text"
                      value={matchSim.poNumber}
                      disabled
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">الكمية المتعاقد عليها:</label>
                    <input
                      type="number"
                      value={matchSim.poQty}
                      onChange={(e) =>
                        setMatchSim({ ...matchSim, poQty: parseFloat(e.target.value) || 0 })
                      }
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">سعر الوحدة المتفق عليه:</label>
                    <input
                      type="number"
                      value={matchSim.poUnitPrice}
                      onChange={(e) =>
                        setMatchSim({
                          ...matchSim,
                          poUnitPrice: parseFloat(e.target.value) || 0,
                        })
                      }
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                </div>
              </div>

              {/* Document 2: Goods Receipt */}
              <div className="p-4 rounded-lg border border-border/60 bg-muted/20 space-y-3">
                <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                  <Truck className="h-4 w-4 text-amber-500" />
                  سند الاستلام الفعلي (GRN)
                </div>
                <div className="space-y-2 text-xs">
                  <div>
                    <label className="text-muted-foreground">رقم سند الاستلام:</label>
                    <input
                      type="text"
                      value={matchSim.grnNumber}
                      disabled
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">الكمية المقبولة فحصاً:</label>
                    <input
                      type="number"
                      value={matchSim.grnAcceptedQty}
                      onChange={(e) =>
                        setMatchSim({
                          ...matchSim,
                          grnAcceptedQty: parseFloat(e.target.value) || 0,
                        })
                      }
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">تكلفة الوحدة المحسوبة:</label>
                    <input
                      type="number"
                      value={matchSim.poUnitPrice}
                      disabled
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-muted-foreground"
                    />
                  </div>
                </div>
              </div>

              {/* Document 3: Supplier Invoice */}
              <div className="p-4 rounded-lg border border-border/60 bg-muted/20 space-y-3">
                <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                  <FileText className="h-4 w-4 text-blue-500" />
                  فاتورة المورد (AP Invoice)
                </div>
                <div className="space-y-2 text-xs">
                  <div>
                    <label className="text-muted-foreground">رقم فاتورة المورد:</label>
                    <input
                      type="text"
                      value={matchSim.invoiceNumber}
                      disabled
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">الكمية المفوترة:</label>
                    <input
                      type="number"
                      value={matchSim.invoiceQty}
                      onChange={(e) =>
                        setMatchSim({
                          ...matchSim,
                          invoiceQty: parseFloat(e.target.value) || 0,
                        })
                      }
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                  <div>
                    <label className="text-muted-foreground">سعر الوحدة المفوتر:</label>
                    <input
                      type="number"
                      value={matchSim.invoiceUnitPrice}
                      onChange={(e) =>
                        setMatchSim({
                          ...matchSim,
                          invoiceUnitPrice: parseFloat(e.target.value) || 0,
                        })
                      }
                      className="w-full mt-1 px-3 py-1.5 rounded border border-border bg-card font-mono text-foreground"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between pt-2">
              <div className="text-xs text-muted-foreground">
                نسبة السماح السعري المقبولة:{" "}
                <span className="font-bold text-foreground">
                  {matchSim.priceTolerancePercent}%
                </span>
              </div>
              <button
                onClick={handleRunMatch}
                className="flex items-center gap-2 px-5 py-2.5 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition shadow-sm"
              >
                <CheckCircle2 className="h-4 w-4" />
                {p.btn_verify_match || "إجراء فحص المطابقة الثلاثية"}
              </button>
            </div>

            {/* Results Panel */}
            {matchResult && (
              <div
                className={`p-4 rounded-xl border transition ${
                  matchResult.passed
                    ? "bg-emerald-500/10 border-emerald-500/30"
                    : "bg-red-500/10 border-red-500/30"
                }`}
              >
                <div className="flex items-center gap-3">
                  {matchResult.passed ? (
                    <CheckCircle2 className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                  ) : (
                    <AlertTriangle className="h-6 w-6 text-red-600 dark:text-red-400" />
                  )}
                  <div>
                    <div
                      className={`text-sm font-bold ${
                        matchResult.passed
                          ? "text-emerald-700 dark:text-emerald-400"
                          : "text-red-700 dark:text-red-400"
                      }`}
                    >
                      {matchResult.passed
                        ? "نتيجة المطابقة: مقبولة ومعتمدة للصرف المالي (Cleared)"
                        : "تنبيه رقابي: وجود انحرافات تتجاوز النطاق المسموح (Blocked)"}
                    </div>
                    <div className="text-xs text-muted-foreground mt-0.5">
                      فارق الكمية:{" "}
                      <span className="font-mono font-bold text-foreground">
                        {matchResult.qtyDiff}
                      </span>{" "}
                      وحدة | فارق سعر الوحدة:{" "}
                      <span className="font-mono font-bold text-foreground">
                        {matchResult.priceDiff.toFixed(2)}
                      </span>{" "}
                      ر.س
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Modal: New Purchase Request */}
      {showNewPRModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
          <div className="bg-card border border-border rounded-xl shadow-lg w-full max-w-lg p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-border/40 pb-3">
              <h3 className="font-bold text-lg text-foreground">
                {p.btn_new_request || "إنشاء طلب شراء داخلي جديد"}
              </h3>
              <button
                onClick={() => setShowNewPRModal(false)}
                className="text-muted-foreground hover:text-foreground p-1 rounded"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreatePR} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-foreground block mb-1">
                  الغرض من الطلب / بيان الاحتياج
                </label>
                <textarea
                  required
                  rows={3}
                  value={newPRPurpose}
                  onChange={(e) => setNewPRPurpose(e.target.value)}
                  placeholder="مثال: توريد مواد أولية ومستلزمات تشغيلية للصيانة الدورية..."
                  className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background focus:ring-1 focus:ring-primary text-foreground"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-foreground block mb-1">
                  المبلغ التقديري الإجمالي (ر.س)
                </label>
                <input
                  type="number"
                  required
                  value={newPRAmount}
                  onChange={(e) => setNewPRAmount(e.target.value)}
                  className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background focus:ring-1 focus:ring-primary text-foreground font-mono"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-2 border-t border-border/40">
                <button
                  type="button"
                  onClick={() => setShowNewPRModal(false)}
                  className="px-4 py-2 text-sm text-muted-foreground hover:bg-muted rounded-lg transition"
                >
                  إلغاء
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm bg-primary text-primary-foreground font-medium rounded-lg hover:bg-primary/90 transition shadow-sm"
                >
                  حفظ وتقديم الطلب
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

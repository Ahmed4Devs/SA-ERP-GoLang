"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  ReceiptText,
  Calculator,
  FileCheck,
  Building2,
  Calendar,
  AlertCircle,
  CheckCircle2,
  ArrowUpRight,
  ArrowDownLeft,
  Scale,
  Percent,
  ShieldCheck,
  Search,
  Plus,
  RefreshCw,
  Send,
  Layers,
  HelpCircle,
} from "lucide-react";

interface TaxCodeItem {
  id: string;
  code: string;
  name_ar: string;
  name_en: string;
  category: "S" | "Z" | "E" | "O";
  current_rate: string;
  type: string;
  sales_account: string;
  purchase_account: string;
  effective_date: string;
}

interface TaxTransactionItem {
  id: string;
  number: string;
  date: string;
  type: "OUTPUT" | "INPUT";
  module: string;
  source_type: string;
  partner_name: string;
  partner_vat: string;
  taxable: number;
  vat: number;
  gross: number;
  status: string;
}

export function TaxView() {
  const { locale, t, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"overview" | "codes" | "calculator" | "transactions" | "return">("overview");

  // Mock initial state representing the seeded database structure
  const taxCodes: TaxCodeItem[] = [
    {
      id: "1",
      code: "VAT-STD-15",
      name_ar: "ضريبة القيمة المضافة بالنسبة الأساسية 15%",
      name_en: "Standard Rate 15% VAT",
      category: "S",
      current_rate: "15.00%",
      type: "OUTPUT / INPUT",
      sales_account: "213000 - ضريبة المخرجات المستحقة",
      purchase_account: "114000 - ضريبة المدخلات المستردة",
      effective_date: "2020-07-01 -> مستمر",
    },
    {
      id: "2",
      code: "VAT-ZERO-0",
      name_ar: "ضريبة القيمة المضافة بنسبة الصفر 0%",
      name_en: "Zero Rated 0% VAT",
      category: "Z",
      current_rate: "0.00%",
      type: "OUTPUT / INPUT",
      sales_account: "213000 - ضريبة المخرجات المستحقة",
      purchase_account: "114000 - ضريبة المدخلات المستردة",
      effective_date: "2018-01-01 -> مستمر",
    },
    {
      id: "3",
      code: "VAT-EXEMPT-0",
      name_ar: "معفى من ضريبة القيمة المضافة",
      name_en: "Exempt from VAT",
      category: "E",
      current_rate: "0.00%",
      type: "OUTPUT / INPUT",
      sales_account: "213000 - ضريبة المخرجات المستحقة",
      purchase_account: "غير مؤهل للاسترداد",
      effective_date: "2018-01-01 -> مستمر",
    },
    {
      id: "4",
      code: "VAT-OOS-0",
      name_ar: "خارج نطاق ضريبة القيمة المضافة",
      name_en: "Out of Scope Supplies",
      category: "O",
      current_rate: "0.00%",
      type: "BOTH",
      sales_account: "غير خاضع",
      purchase_account: "غير خاضع",
      effective_date: "2018-01-01 -> مستمر",
    },
  ];

  // Interactive Live Calculator State
  const [calcQuantity, setCalcQuantity] = useState<number>(1);
  const [calcPrice, setCalcPrice] = useState<number>(1000);
  const [calcDiscount, setCalcDiscount] = useState<number>(0);
  const [calcCode, setCalcCode] = useState<string>("VAT-STD-15");
  const [calcReason, setCalcReason] = useState<string>("");

  const selectedTaxCode = taxCodes.find((c) => c.code === calcCode) || taxCodes[0];
  const ratePercent = selectedTaxCode.category === "S" ? 15 : 0;
  const calcSubtotal = Math.max(0, calcQuantity * calcPrice - calcDiscount);
  const calcTaxAmount = Math.round(calcSubtotal * (ratePercent / 100) * 100) / 100;
  const calcGross = calcSubtotal + calcTaxAmount;

  // Recent Transactions Mock
  const transactions: TaxTransactionItem[] = [
    {
      id: "1",
      number: "TX-20260710-01",
      date: "2026-07-10",
      type: "OUTPUT",
      module: "SALES",
      source_type: "INVOICE",
      partner_name: "شركة البناء الحديث للتجارة",
      partner_vat: "310123456700003",
      taxable: 10000.0,
      vat: 1500.0,
      gross: 11500.0,
      status: "POSTED",
    },
    {
      id: "2",
      number: "TX-20260712-02",
      date: "2026-07-12",
      type: "OUTPUT",
      module: "SALES",
      source_type: "INVOICE",
      partner_name: "مؤسسة دبي للحلول اللوجستية (تصدير)",
      partner_vat: "N/A (GCC Export)",
      taxable: 4000.0,
      vat: 0.0,
      gross: 4000.0,
      status: "POSTED",
    },
    {
      id: "3",
      number: "TX-20260715-03",
      date: "2026-07-15",
      type: "OUTPUT",
      module: "SALES",
      source_type: "INVOICE",
      partner_name: "شركة التطوير العقاري السكني (إيجار)",
      partner_vat: "310998877600003",
      taxable: 2000.0,
      vat: 0.0,
      gross: 2000.0,
      status: "POSTED",
    },
    {
      id: "4",
      number: "TX-20260718-04",
      date: "2026-07-18",
      type: "INPUT",
      module: "PURCHASES",
      source_type: "BILL",
      partner_name: "شركة التجهيزات الصناعية السعودية",
      partner_vat: "310554433200003",
      taxable: 6000.0,
      vat: 900.0,
      gross: 6900.0,
      status: "POSTED",
    },
  ];

  // Return Submission State
  const [returnStatus, setReturnStatus] = useState<"DRAFT" | "SUBMITTED">("DRAFT");

  return (
    <div className="space-y-6" dir={dir}>
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20 mb-2">
            <ShieldCheck className="h-3.5 w-3.5" />
            <span>ZATCA Indirect Tax Engine v2.0</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            {locale === "ar" ? "نظام ضريبة القيمة المضافة وهيئة الزكاة (VAT & ZATCA)" : "VAT & ZATCA Tax Engine"}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {locale === "ar"
              ? "إدارة الرموز الضريبية، السريان التاريخي للنسب (Effective-Dated Rates)، وسجل القيود وإقرار الـ 16 حقلاً المعتمد"
              : "Effective-dated VAT rates, exact financial arithmetic, tax transaction ledger, and official 16-box ZATCA declaration."}
          </p>
        </div>

        {/* Global Action Badges */}
        <div className="flex items-center gap-2">
          <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border bg-card text-xs font-medium text-foreground">
            <Building2 className="h-4 w-4 text-primary" />
            <span>{locale === "ar" ? "الرقم الضريبي: 310123456700003" : "VAT: 310123456700003"}</span>
          </div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border bg-card text-xs font-medium text-foreground">
            <Calendar className="h-4 w-4 text-emerald-500" />
            <span>{locale === "ar" ? "الفترة: يوليو 2026 (مفتوحة)" : "Period: July 2026 (Open)"}</span>
          </div>
        </div>
      </div>

      {/* Primary Navigation Tabs */}
      <div className="flex flex-wrap gap-2 border-b pb-2">
        <button
          onClick={() => setActiveTab("overview")}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "overview"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "hover:bg-muted text-muted-foreground"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>{locale === "ar" ? "نظرة عامة والملخص" : "Overview & Metrics"}</span>
        </button>

        <button
          onClick={() => setActiveTab("codes")}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "codes"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "hover:bg-muted text-muted-foreground"
          }`}
        >
          <Percent className="h-4 w-4" />
          <span>{locale === "ar" ? "دليل الرموز والنسب" : "Tax Codes & Rates"}</span>
        </button>

        <button
          onClick={() => setActiveTab("calculator")}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "calculator"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "hover:bg-muted text-muted-foreground"
          }`}
        >
          <Calculator className="h-4 w-4" />
          <span>{locale === "ar" ? "الحاسبة الدقيقة" : "Live Calculator"}</span>
        </button>

        <button
          onClick={() => setActiveTab("transactions")}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "transactions"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "hover:bg-muted text-muted-foreground"
          }`}
        >
          <ReceiptText className="h-4 w-4" />
          <span>{locale === "ar" ? "سجل المعاملات الضريبية" : "Tax Ledger"}</span>
        </button>

        <button
          onClick={() => setActiveTab("return")}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "return"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "hover:bg-muted text-muted-foreground"
          }`}
        >
          <FileCheck className="h-4 w-4" />
          <span>{locale === "ar" ? "إقرار ZATCA (16 حقلاً)" : "VAT Return Form"}</span>
        </button>
      </div>

      {/* TAB 1: OVERVIEW & KEY METRICS */}
      {activeTab === "overview" && (
        <div className="space-y-6">
          {/* 4 Cards Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Card 1: Output Tax */}
            <div className="rounded-xl border bg-card p-5 space-y-2">
              <div className="flex items-center justify-between text-muted-foreground">
                <span className="text-xs font-medium">{locale === "ar" ? "ضريبة المخرجات (المبيعات)" : "Output Tax (Sales)"}</span>
                <div className="h-8 w-8 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center">
                  <ArrowUpRight className="h-4 w-4" />
                </div>
              </div>
              <div className="text-2xl font-bold text-foreground">1,500.00 <span className="text-xs font-normal text-muted-foreground">SAR</span></div>
              <div className="text-[11px] text-emerald-600 dark:text-emerald-400 flex items-center gap-1 font-medium">
                <span>{locale === "ar" ? "من إجمالي مبيعات 16,000.00 ر.س" : "From 16,000.00 SAR taxable sales"}</span>
              </div>
            </div>

            {/* Card 2: Input Tax */}
            <div className="rounded-xl border bg-card p-5 space-y-2">
              <div className="flex items-center justify-between text-muted-foreground">
                <span className="text-xs font-medium">{locale === "ar" ? "ضريبة المدخلات (المشتريات)" : "Input Tax (Purchases)"}</span>
                <div className="h-8 w-8 rounded-lg bg-blue-500/10 text-blue-600 flex items-center justify-center">
                  <ArrowDownLeft className="h-4 w-4" />
                </div>
              </div>
              <div className="text-2xl font-bold text-foreground">900.00 <span className="text-xs font-normal text-muted-foreground">SAR</span></div>
              <div className="text-[11px] text-blue-600 dark:text-blue-400 flex items-center gap-1 font-medium">
                <span>{locale === "ar" ? "مشتريات خاضعة للخصم 6,000.00 ر.س" : "From 6,000.00 SAR eligible purchases"}</span>
              </div>
            </div>

            {/* Card 3: Net VAT Due */}
            <div className="rounded-xl border bg-card p-5 space-y-2">
              <div className="flex items-center justify-between text-muted-foreground">
                <span className="text-xs font-medium">{locale === "ar" ? "صافي الضريبة المستحقة للهيئة" : "Net VAT Payable to ZATCA"}</span>
                <div className="h-8 w-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                  <Scale className="h-4 w-4" />
                </div>
              </div>
              <div className="text-2xl font-bold text-primary">600.00 <span className="text-xs font-normal text-muted-foreground">SAR</span></div>
              <div className="text-[11px] text-muted-foreground flex items-center gap-1">
                <span>{locale === "ar" ? "المخرجات (1,500) - المدخلات (900)" : "Output (1,500) - Input (900)"}</span>
              </div>
            </div>

            {/* Card 4: Compliance Status */}
            <div className="rounded-xl border bg-card p-5 space-y-2">
              <div className="flex items-center justify-between text-muted-foreground">
                <span className="text-xs font-medium">{locale === "ar" ? "حالة الإقرار الضريبي الحالي" : "Return Filing Status"}</span>
                <div className="h-8 w-8 rounded-lg bg-purple-500/10 text-purple-600 flex items-center justify-center">
                  <FileCheck className="h-4 w-4" />
                </div>
              </div>
              <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
                {returnStatus === "DRAFT" ? (locale === "ar" ? "مسودة جاهزة" : "Draft Ready") : (locale === "ar" ? "تم الاعتماد" : "Submitted")}
              </div>
              <div className="text-[11px] text-muted-foreground flex items-center gap-1">
                <span>{locale === "ar" ? "تاريخ الاستحقاق: 31 أغسطس 2026" : "Due Date: Aug 31, 2026"}</span>
              </div>
            </div>
          </div>

          {/* Historical Reproducibility & ZATCA Invariant Alert Banner */}
          <div className="rounded-xl border border-primary/20 bg-primary/5 p-4 sm:p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-start gap-3">
              <ShieldCheck className="h-6 w-6 text-primary shrink-0 mt-0.5" />
              <div>
                <h4 className="font-semibold text-sm text-foreground">
                  {locale === "ar"
                    ? "قواعد الامتثال وإعادة الإنتاج التاريخي (Historical Reproducibility)"
                    : "Regulatory Invariant: Historical Reproducibility"}
                </h4>
                <p className="text-xs text-muted-foreground mt-1 max-w-3xl leading-relaxed">
                  {locale === "ar"
                    ? "وفقاً للمواصفة القياسية، تُحفظ نسب الضرائب بصيغة السريان الزمني (Effective-Dating). الفواتير التاريخية الصادرة قبل 1 يوليو 2020 تُحسب بنسبة 5% قطيعاً دون إعادة احتسابها بالنسبة الحالية 15%."
                    : "As per ERP canonical invariants, tax rates are effective-dated. Historical invoices prior to July 1, 2020 are calculated at 5% deterministically without recalculation from modern 15% configurations."}
                </p>
              </div>
            </div>
            <div className="px-3 py-1.5 rounded-md bg-background border text-xs font-mono font-medium shrink-0">
              SOCPA / ZATCA 2.0
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: TAX CODES & RATES */}
      {activeTab === "codes" && (
        <div className="space-y-6">
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="p-4 sm:p-5 border-b flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="font-semibold text-base text-foreground">
                  {locale === "ar" ? "الأكواد والنسب الضريبية المعتمدة" : "Active Tax Codes & Rate Versions"}
                </h3>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {locale === "ar"
                    ? "الأكواد المرتبطة بحسابات دفتر الأستاذ العام وتصنيفات ZATCA القياسية (S, Z, E, O)"
                    : "Configured entity tax codes linked to General Ledger output/input accounts."}
                </p>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "الكود" : "Code"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "الاسم" : "Name"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "فئة ZATCA" : "Category"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "النسبة الحالية" : "Current Rate"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "النوع" : "Scope"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "حساب ضريبة المخرجات" : "Output GL"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "حساب ضريبة المدخلات" : "Input GL"}</th>
                    <th className="py-3 px-4 text-start">{locale === "ar" ? "فترة السريان" : "Effective Window"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {taxCodes.map((code) => (
                    <tr key={code.id} className="hover:bg-muted/30 transition-colors">
                      <td className="py-3 px-4 font-mono font-bold text-primary">{code.code}</td>
                      <td className="py-3 px-4 font-medium text-foreground">
                        {locale === "ar" ? code.name_ar : code.name_en}
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-bold ${
                            code.category === "S"
                              ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                              : code.category === "Z"
                              ? "bg-blue-500/10 text-blue-600 border border-blue-500/20"
                              : code.category === "E"
                              ? "bg-amber-500/10 text-amber-600 border border-amber-500/20"
                              : "bg-purple-500/10 text-purple-600 border border-purple-500/20"
                          }`}
                        >
                          {code.category} - {code.category === "S" ? "Standard" : code.category === "Z" ? "Zero" : code.category === "E" ? "Exempt" : "Out of Scope"}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-bold text-foreground">{code.current_rate}</td>
                      <td className="py-3 px-4 text-muted-foreground">{code.type}</td>
                      <td className="py-3 px-4 text-muted-foreground">{code.sales_account}</td>
                      <td className="py-3 px-4 text-muted-foreground">{code.purchase_account}</td>
                      <td className="py-3 px-4 font-mono text-[11px] text-muted-foreground">{code.effective_date}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Reference Reasons Table */}
          <div className="rounded-xl border bg-card p-5 space-y-3">
            <h4 className="font-semibold text-sm text-foreground flex items-center gap-2">
              <HelpCircle className="h-4 w-4 text-primary" />
              <span>{locale === "ar" ? "أكواد أسباب الإعفاء ونسبة الصفر المعتمدة من هيئة الزكاة (ZATCA Reason Codes)" : "Standard ZATCA Exemption & Zero-Rate Reason Codes"}</span>
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
              <div className="p-3 rounded-lg bg-muted/40 border space-y-1">
                <span className="font-bold text-amber-600">VATEX-SA-29</span>
                <p className="text-muted-foreground">{locale === "ar" ? "الخدمات المالية المعفاة المنصوص عليها في المادة 29 من اللائحة" : "Exempt financial services under Article 29"}</p>
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border space-y-1">
                <span className="font-bold text-amber-600">VATEX-SA-31</span>
                <p className="text-muted-foreground">{locale === "ar" ? "تأجير العقارات السكنية المعفى من ضريبة القيمة المضافة" : "Lease of residential real estate exempt from VAT"}</p>
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border space-y-1">
                <span className="font-bold text-blue-600">VATEX-SA-32</span>
                <p className="text-muted-foreground">{locale === "ar" ? "صادرات السلع خارج إقليم دول مجلس التعاون الخليجي (نسبة الصفر)" : "Export of goods outside GCC territory (0%)"}</p>
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border space-y-1">
                <span className="font-bold text-blue-600">VATEX-SA-EDU / HEA</span>
                <p className="text-muted-foreground">{locale === "ar" ? "خدمات التعليم الأهلي والرعاية الصحية للمواطنين وتتحملها الدولة" : "Private education and healthcare services to Saudi citizens (0%)"}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: LIVE EXACT CALCULATOR */}
      {activeTab === "calculator" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Input Form */}
          <div className="lg:col-span-7 rounded-xl border bg-card p-5 sm:p-6 space-y-4 shadow-sm">
            <div>
              <h3 className="font-semibold text-base text-foreground">
                {locale === "ar" ? "حاسبة الضريبة التفاعلية الدقيقة" : "Live VAT Interactive Calculator"}
              </h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {locale === "ar"
                  ? "حساب دقيق متوافق مع خوارزميات التقريب المالي لهيئة الزكاة حتى أجزاء الهللة"
                  : "Exact arithmetic with halala-precision rounding and ZATCA compliance validations."}
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">{locale === "ar" ? "الرمز الضريبي" : "Tax Code"}</label>
                <select
                  value={calcCode}
                  onChange={(e) => setCalcCode(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  {taxCodes.map((c) => (
                    <option key={c.id} value={c.code}>
                      {c.code} - {locale === "ar" ? c.name_ar : c.name_en} ({c.current_rate})
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">{locale === "ar" ? "الكمية" : "Quantity"}</label>
                <input
                  type="number"
                  min="1"
                  value={calcQuantity}
                  onChange={(e) => setCalcQuantity(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">{locale === "ar" ? "سعر الوحدة (ر.س)" : "Unit Price (SAR)"}</label>
                <input
                  type="number"
                  step="0.01"
                  value={calcPrice}
                  onChange={(e) => setCalcPrice(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">{locale === "ar" ? "مبلغ الخصم (ر.س)" : "Discount Amount (SAR)"}</label>
                <input
                  type="number"
                  step="0.01"
                  value={calcDiscount}
                  onChange={(e) => setCalcDiscount(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>

            {/* Invariant Note */}
            <div className="p-3 rounded-lg bg-muted/40 border text-xs text-muted-foreground space-y-1">
              <span className="font-semibold text-foreground">{locale === "ar" ? "قاعدة الاحتساب:" : "Calculation Rule:"}</span>
              <p>
                {locale === "ar"
                  ? "المبلغ الخاضع = (الكمية × السعر) - الخصم | مبلغ الضريبة = المبلغ الخاضع × النسبة مقرباً لأقرب هللتين"
                  : "Taxable = (Qty * Price) - Discount | VAT = Taxable * Rate rounded to 2 decimals."}
              </p>
            </div>
          </div>

          {/* Results Summary Card */}
          <div className="lg:col-span-5 rounded-xl border bg-card p-5 sm:p-6 space-y-4 shadow-sm flex flex-col justify-between">
            <div>
              <h4 className="font-semibold text-sm text-foreground mb-3">
                {locale === "ar" ? "نتيجة الحساب المالي" : "Calculated Breakdown"}
              </h4>

              <div className="space-y-3 divide-y divide-border text-xs">
                <div className="flex justify-between items-center pt-2">
                  <span className="text-muted-foreground">{locale === "ar" ? "المجموع قبل الضريبة" : "Taxable Subtotal"}</span>
                  <span className="font-mono font-bold text-foreground text-sm">{calcSubtotal.toFixed(2)} SAR</span>
                </div>

                <div className="flex justify-between items-center pt-2">
                  <span className="text-muted-foreground">
                    {locale === "ar" ? `مبلغ الضريبة (${ratePercent}%)` : `VAT Amount (${ratePercent}%)`}
                  </span>
                  <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400 text-sm">
                    {calcTaxAmount.toFixed(2)} SAR
                  </span>
                </div>

                <div className="flex justify-between items-center pt-3 text-base">
                  <span className="font-bold text-foreground">{locale === "ar" ? "الإجمالي شامل الضريبة" : "Gross Total"}</span>
                  <span className="font-mono font-bold text-primary text-lg">{calcGross.toFixed(2)} SAR</span>
                </div>
              </div>
            </div>

            <div className="pt-4 border-t flex items-center gap-2 text-xs text-muted-foreground">
              <CheckCircle2 className="h-4 w-4 text-emerald-500" />
              <span>{locale === "ar" ? "مطابق لمعايير التقريب الرياضي ZATCA" : "Exact Decimal Precision Certified"}</span>
            </div>
          </div>
        </div>
      )}

      {/* TAB 4: TAX LEDGER TRANSACTIONS */}
      {activeTab === "transactions" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="p-4 sm:p-5 border-b flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h3 className="font-semibold text-base text-foreground">
                {locale === "ar" ? "سجل القيود والمعاملات الضريبية" : "Tax Ledger Transactions"}
              </h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {locale === "ar"
                  ? "معاملات المبيعات والمشتريات المرحلة آلياً والمصنفة ضريبياً"
                  : "Immutable posted transactions feeding the VAT return declaration."}
              </p>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                <tr>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "رقم الحركة" : "Tx Number"}</th>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "التاريخ" : "Date"}</th>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "النوع" : "Type"}</th>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "المصدر" : "Source"}</th>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "الطرف الآخر (العميل/المورد)" : "Partner / Customer"}</th>
                  <th className="py-3 px-4 text-start">{locale === "ar" ? "الرقم الضريبي للطرف" : "Partner VAT"}</th>
                  <th className="py-3 px-4 text-end">{locale === "ar" ? "المبلغ الخاضع" : "Taxable"}</th>
                  <th className="py-3 px-4 text-end">{locale === "ar" ? "مبلغ الضريبة" : "VAT Amount"}</th>
                  <th className="py-3 px-4 text-end">{locale === "ar" ? "الإجمالي" : "Gross"}</th>
                  <th className="py-3 px-4 text-center">{locale === "ar" ? "الحالة" : "Status"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {transactions.map((tx) => (
                  <tr key={tx.id} className="hover:bg-muted/30 transition-colors">
                    <td className="py-3 px-4 font-mono font-bold text-foreground">{tx.number}</td>
                    <td className="py-3 px-4 text-muted-foreground">{tx.date}</td>
                    <td className="py-3 px-4">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold ${
                          tx.type === "OUTPUT"
                            ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                            : "bg-blue-500/10 text-blue-600 border border-blue-500/20"
                        }`}
                      >
                        {tx.type === "OUTPUT" ? (locale === "ar" ? "مخرجات (مبيعات)" : "Output (Sales)") : (locale === "ar" ? "مدخلات (مشتريات)" : "Input (Purchases)")}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-muted-foreground">{tx.module} / {tx.source_type}</td>
                    <td className="py-3 px-4 font-medium text-foreground">{tx.partner_name}</td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{tx.partner_vat}</td>
                    <td className="py-3 px-4 text-end font-mono font-medium">{tx.taxable.toFixed(2)}</td>
                    <td className="py-3 px-4 text-end font-mono font-bold text-emerald-600 dark:text-emerald-400">
                      {tx.vat.toFixed(2)}
                    </td>
                    <td className="py-3 px-4 text-end font-mono font-bold">{tx.gross.toFixed(2)}</td>
                    <td className="py-3 px-4 text-center">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                        {tx.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 5: ZATCA 16-BOX VAT RETURN FORM */}
      {activeTab === "return" && (
        <div className="space-y-6">
          <div className="rounded-xl border bg-card p-5 sm:p-6 shadow-sm space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-4">
              <div>
                <h3 className="font-bold text-base sm:text-lg text-foreground flex items-center gap-2">
                  <ReceiptText className="h-5 w-5 text-primary" />
                  <span>{locale === "ar" ? "نموذج إقرار ضريبة القيمة المضافة (ZATCA 16-Box Form)" : "Official ZATCA 16-Box VAT Declaration"}</span>
                </h3>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {locale === "ar"
                    ? "نموذج الإقرار الضريبي الرسمي المعتمد لهيئة الزكاة والضريبة والجمارك للفترة: يوليو 2026"
                    : "Standard ZATCA electronic declaration form for period: July 2026"}
                </p>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => alert(locale === "ar" ? "تم تحديث مسودة الإقرار الضريبي بنجاح" : "Draft updated")}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium hover:bg-muted"
                >
                  <RefreshCw className="h-3.5 w-3.5" />
                  <span>{locale === "ar" ? "تحديث المسودة" : "Refresh Draft"}</span>
                </button>

                <button
                  onClick={() => {
                    setReturnStatus("SUBMITTED");
                    alert(locale === "ar" ? "تم اعتماد وتقديم الإقرار الضريبي بنجاح وقفل الفترة" : "Return submitted and period locked!");
                  }}
                  disabled={returnStatus === "SUBMITTED"}
                  className="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-primary text-primary-foreground text-xs font-semibold hover:bg-primary/90 disabled:opacity-50 shadow-sm"
                >
                  <Send className="h-3.5 w-3.5" />
                  <span>{returnStatus === "SUBMITTED" ? (locale === "ar" ? "معتمد ومقدّم" : "Submitted") : (locale === "ar" ? "اعتماد وتقديم الإقرار" : "Authorize & Submit")}</span>
                </button>
              </div>
            </div>

            {/* Declaration Boxes Table */}
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start border">
                <thead className="bg-muted/70 border-b text-foreground font-semibold">
                  <tr>
                    <th className="py-2.5 px-3 text-start w-16">{locale === "ar" ? "الحقل" : "Box"}</th>
                    <th className="py-2.5 px-3 text-start">{locale === "ar" ? "بيان التوريدات والمشتريات" : "Description of Supplies"}</th>
                    <th className="py-2.5 px-3 text-end">{locale === "ar" ? "المبلغ (ر.س)" : "Amount (SAR)"}</th>
                    <th className="py-2.5 px-3 text-end">{locale === "ar" ? "التعديل (ر.س)" : "Adjustment"}</th>
                    <th className="py-2.5 px-3 text-end">{locale === "ar" ? "مبلغ الضريبة (ر.س)" : "VAT Amount (SAR)"}</th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-border">
                  {/* Sales Section Header */}
                  <tr className="bg-primary/5 font-bold text-foreground">
                    <td colSpan={5} className="py-2 px-3 text-primary">
                      {locale === "ar" ? "أولاً: المبيعات والتوريدات (ضريبة المخرجات)" : "Part 1: Sales / Supplies (Output VAT)"}
                    </td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">1</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المبيعات الخاضعة للنسبة الأساسية (15%)" : "Standard rated sales (15%)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">10,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono font-bold text-emerald-600 dark:text-emerald-400">1,500.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">2</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المبيعات للمواطنين (الرعاية الصحية الخاصة والتعليم الأهلي)" : "Sales to citizens (Private healthcare & education)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">3</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المبيعات المحلية الخاضعة لنسبة الصفر" : "Zero rated domestic sales"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">4</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "الصادرات خارج دول المجلس (نسبة الصفر)" : "Exports outside GCC (0%)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">4,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">5</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المبيعات المعفاة من الضريبة" : "Exempt sales"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">2,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr className="bg-muted/30 font-bold border-t border-b">
                    <td className="py-2.5 px-3 font-mono text-primary">6</td>
                    <td className="py-2.5 px-3 text-foreground">{locale === "ar" ? "إجمالي المبيعات (الحقول 1 إلى 5)" : "Total Sales (Boxes 1 to 5)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">16,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-primary">1,500.00</td>
                  </tr>

                  {/* Purchases Section Header */}
                  <tr className="bg-primary/5 font-bold text-foreground">
                    <td colSpan={5} className="py-2 px-3 text-primary">
                      {locale === "ar" ? "ثانياً: المشتريات والمصروفات (ضريبة المدخلات)" : "Part 2: Purchases / Expenses (Input VAT)"}
                    </td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">7</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المشتريات الخاضعة للنسبة الأساسية (15%)" : "Standard rated purchases (15%)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">6,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono font-bold text-blue-600 dark:text-blue-400">900.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">8</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "الاستيرادات الخاضعة للضريبة والمسددة في الجمارك" : "Imports subject to VAT paid at customs"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">9</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "الاستيرادات الخاضعة لآلية الاحتساب العكسي" : "Imports subject to reverse charge mechanism"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">10</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المشتريات الخاضعة لنسبة الصفر" : "Zero rated purchases"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">11</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "المشتريات المعفاة من الضريبة" : "Exempt purchases"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr className="bg-muted/30 font-bold border-t border-b">
                    <td className="py-2.5 px-3 font-mono text-primary">12</td>
                    <td className="py-2.5 px-3 text-foreground">{locale === "ar" ? "إجمالي المشتريات (الحقول 7 إلى 11)" : "Total Purchases (Boxes 7 to 11)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono">6,000.00</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                    <td className="py-2.5 px-3 text-end font-mono text-blue-600 dark:text-blue-400">900.00</td>
                  </tr>

                  {/* Net VAT Section Header */}
                  <tr className="bg-primary/10 font-bold text-foreground">
                    <td colSpan={5} className="py-2 px-3 text-primary">
                      {locale === "ar" ? "ثالثاً: صافي ضريبة القيمة المضافة المستحقة (الحقول 13 إلى 16)" : "Part 3: Net VAT Due (Boxes 13 to 16)"}
                    </td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">13</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "إجمالي الضريبة المستحقة للفترة (حقل 6 - حقل 12)" : "Total VAT due for period (Box 6 - Box 12)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono font-bold text-foreground">600.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">14</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "تصحيحات من فترات سابقة (أقل من 5,000 ر.س)" : "Corrections from previous periods (< 5,000 SAR)"}</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr>
                    <td className="py-2.5 px-3 font-mono font-bold">15</td>
                    <td className="py-2.5 px-3 font-medium">{locale === "ar" ? "رصيد الضريبة الدائن المرحّل من فترات سابقة" : "VAT credit carried forward"}</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-2.5 px-3 text-end font-mono">0.00</td>
                  </tr>

                  <tr className="bg-primary/20 font-bold border-t text-sm">
                    <td className="py-3 px-3 font-mono text-primary font-extrabold text-base">16</td>
                    <td className="py-3 px-3 text-foreground font-extrabold">{locale === "ar" ? "صافي الضريبة الواجب سدادها (أو المستردة)" : "Net VAT Payable / (Refundable)"}</td>
                    <td className="py-3 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-3 px-3 text-end font-mono text-muted-foreground">-</td>
                    <td className="py-3 px-3 text-end font-mono text-lg text-primary font-black">600.00 SAR</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

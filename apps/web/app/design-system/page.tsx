"use client";

import React, { useState } from "react";
import {
  Sparkles,
  ShieldCheck,
  CheckCircle2,
  FileText,
  DollarSign,
  History,
  Scale,
  Users,
  Building2,
  Sun,
  Moon,
  Globe,
  Layers,
  Search,
  Plus,
  RefreshCw,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { DocumentHeader } from "@/components/erp/document-header";
import { DocumentStatus } from "@/components/erp/document-status";
import { DocumentActions } from "@/components/erp/document-actions";
import { ApprovalTimeline, ApprovalStep } from "@/components/erp/approval-timeline";
import { AuditDrawer, AuditRecord } from "@/components/erp/audit-drawer";
import { TotalsPanel } from "@/components/erp/totals-panel";
import { ZATCAStatusBadge } from "@/components/erp/zatca-status-badge";
import { MoneyInput } from "@/components/erp/money-input";
import { DataTable, ColumnDef } from "@/components/data-table/data-table";
import { KPIMetricCard } from "@/components/charts/kpi-metric-card";

interface SampleRow {
  id: string;
  code: string;
  name: string;
  category: string;
  amount: number;
  status: string;
}

export default function DesignSystemPage() {
  const { locale, setLocale, t, dir } = useLocale();
  const isAr = locale === "ar";
  const { theme, setTheme } = useTheme();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [auditOpen, setAuditOpen] = useState(false);
  const [moneyVal, setMoneyVal] = useState(12500.5);
  const [btnLoading, setBtnLoading] = useState(false);

  // Sample data for Approval Timeline
  const sampleApprovalSteps: ApprovalStep[] = [
    {
      id: "step-1",
      stepNumber: 1,
      stepNameAr: "مراجعة المحاسب",
      stepNameEn: "Accountant Review",
      approverName: isAr ? "سارة العتيبي" : "Sarah Al-Otaibi",
      status: "APPROVED",
      decidedAt: "2026-09-18 09:30",
      notes: isAr ? "تم التحقق من مطابقة القيود ومرفقات الفاتورة" : "Entries and invoice attachments verified",
    },
    {
      id: "step-2",
      stepNumber: 2,
      stepNameAr: "اعتماد المدير المالي",
      stepNameEn: "Finance Director Approval",
      approverName: isAr ? "خالد المطيري" : "Khaled Al-Mutairi",
      status: "APPROVED",
      decidedAt: "2026-09-18 10:15",
      notes: isAr ? "معتمد ضمن ميزانية الربع الثالث" : "Approved within Q3 budget allocation",
    },
    {
      id: "step-3",
      stepNumber: 3,
      stepNameAr: "الموافقة التنفيذية النهائية",
      stepNameEn: "Executive Sign-off",
      approverName: isAr ? "أحمد الغامدي" : "Ahmed Al-Ghamdi",
      status: "PENDING",
    },
  ];

  // Sample data for Audit Drawer
  const sampleAuditRecords: AuditRecord[] = [
    {
      id: "rec-1",
      action: "DOCUMENT_CREATED",
      actorName: isAr ? "سارة العتيبي (محاسب)" : "Sarah Al-Otaibi (Accountant)",
      occurredAt: "2026-09-18 09:12:04",
      recordHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    },
    {
      id: "rec-2",
      action: "SUBMITTED_FOR_APPROVAL",
      actorName: isAr ? "سارة العتيبي" : "Sarah Al-Otaibi",
      occurredAt: "2026-09-18 09:30:18",
      recordHash: "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4",
    },
    {
      id: "rec-3",
      action: "DIRECTOR_APPROVED",
      actorName: isAr ? "خالد المطيري (المدير المالي)" : "Khaled Al-Mutairi (CFO)",
      occurredAt: "2026-09-18 10:15:45",
      recordHash: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
    },
  ];

  // Sample data for DataTable
  const sampleTableData: SampleRow[] = [
    { id: "1", code: "INV-2026-0101", name: isAr ? "شركة اليمامة للمقاولات" : "Al-Yamama Contracting Co.", category: "B2B", amount: 48500.0, status: "POSTED" },
    { id: "2", code: "INV-2026-0102", name: isAr ? "مؤسسة الرياض للتقنية" : "Riyadh Tech Est.", category: "SME", amount: 12400.0, status: "CLEARED" },
    { id: "3", code: "INV-2026-0103", name: isAr ? "مستشفى الجزيرة التخصصي" : "Al-Jazeera Hospital", category: "HEALTH", amount: 89000.0, status: "APPROVED" },
    { id: "4", code: "INV-2026-0104", name: isAr ? "مركز نجد التجاري" : "Najd Commercial Hub", category: "RETAIL", amount: 6200.0, status: "PENDING_APPROVAL" },
    { id: "5", code: "INV-2026-0105", name: isAr ? "شركة الشرق للخدمات اللوجستية" : "Orient Logistics Corp", category: "SUPPLY", amount: 34100.0, status: "DRAFT" },
  ];

  const tableColumns: ColumnDef<SampleRow>[] = [
    { key: "code", headerAr: "رقم المستند", headerEn: "Document #", sortable: true },
    { key: "name", headerAr: "اسم العميل / الجهة", headerEn: "Account / Entity", sortable: true },
    { key: "category", headerAr: "التصنيف", headerEn: "Category" },
    {
      key: "amount",
      headerAr: "المبلغ الإجمالي",
      headerEn: "Total Amount",
      sortable: true,
      align: "end",
      render: (row) => (
        <span className="font-mono font-semibold">
          {new Intl.NumberFormat(isAr ? "ar-SA" : "en-US", { style: "currency", currency: "SAR" }).format(row.amount)}
        </span>
      ),
    },
    {
      key: "status",
      headerAr: "الحالة",
      headerEn: "Status",
      align: "center",
      render: (row) => <DocumentStatus status={row.status} isArabic={isAr} />,
    },
  ];

  const currentModeTitle = isAr
    ? `العربية (RTL) - ${theme === "dark" ? "المظهر الداكن (Dark)" : "المظهر الفاتح (Light)"}`
    : `English (LTR) - ${theme === "dark" ? "Dark Mode" : "Light Mode"}`;

  return (
    <div className="space-y-8 pb-16">
      {/* 1. Header Banner & Live Matrix Switcher */}
      <div className="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-card via-card to-primary/10 p-6 sm:p-8 shadow-sm">
        <div className="space-y-3 max-w-3xl">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
            <Sparkles className="h-3.5 w-3.5" />
            <span>{isAr ? "المرحلة 31: تقوية وتثبيت نظام التصميم الشامل" : "Phase 31: UI Hardening & Design System"}</span>
          </div>

          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
            {isAr ? "نظام التصميم الموحد ومصفوفة التحقق البصري" : "Bilingual RTL/LTR ERP Design System Console"}
          </h1>

          <p className="text-sm text-muted-foreground leading-relaxed">
            {isAr
              ? "التحقق والاعتماد الشامل لكافة المكونات وقواعد التخطيط والألوان عبر المصفوفات الأربع: العربية (RTL) الفاتحة والداكنة، والإنجليزية (LTR) الفاتحة والداكنة."
              : "Rigorous verification of all core ERP design tokens, typography, atomic components, and domain widgets across the 4 operational matrices."}
          </p>

          {/* Current Mode Badge */}
          <div className="pt-2 flex items-center gap-2">
            <span className="text-xs text-muted-foreground font-medium">{isAr ? "الوضع النشط حالياً:" : "Current Active Mode:"}</span>
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-bold bg-primary text-primary-foreground shadow-xs">
              <CheckCircle2 className="h-3.5 w-3.5" />
              <span>{currentModeTitle}</span>
            </span>
          </div>
        </div>

        {/* 4-Matrix Quick Switch Buttons */}
        <div className="pt-6 mt-6 border-t flex flex-wrap gap-2.5">
          <Button
            variant={isAr && theme !== "dark" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setLocale("ar");
              setTheme("light");
            }}
            className="gap-1.5"
          >
            <Sun className="h-3.5 w-3.5 text-amber-500" />
            <span>1. Arabic RTL Light</span>
          </Button>

          <Button
            variant={isAr && theme === "dark" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setLocale("ar");
              setTheme("dark");
            }}
            className="gap-1.5"
          >
            <Moon className="h-3.5 w-3.5 text-blue-400" />
            <span>2. Arabic RTL Dark</span>
          </Button>

          <Button
            variant={!isAr && theme !== "dark" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setLocale("en");
              setTheme("light");
            }}
            className="gap-1.5"
          >
            <Sun className="h-3.5 w-3.5 text-amber-500" />
            <span>3. English LTR Light</span>
          </Button>

          <Button
            variant={!isAr && theme === "dark" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setLocale("en");
              setTheme("dark");
            }}
            className="gap-1.5"
          >
            <Moon className="h-3.5 w-3.5 text-blue-400" />
            <span>4. English LTR Dark</span>
          </Button>
        </div>
      </div>

      {/* 2. Canonical ERP Domain Components (Section 13) */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
            <Building2 className="h-5 w-5 text-primary" />
            <span>{isAr ? "مكونات الـ ERP المعيارية (Section 13 Reusable ERP Components)" : "Canonical ERP Reusable Components"}</span>
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {isAr
              ? "المكونات الموحدة لإدارة وترحيل واعتماد المستندات المالية والتشغيلية"
              : "Standardized building blocks for document workflows, totals, status tracking, and audit trails."}
          </p>
        </div>

        {/* Demo DocumentHeader */}
        <DocumentHeader
          documentType={isAr ? "فاتورة مبيعات ضريبية" : "Tax Sales Invoice"}
          documentNumber="INV-2026-0042"
          title={isAr ? "فاتورة توريد وتركيب خوادم البنية التحتية" : "Infrastructure Server Supply & Setup"}
          status="SUBMITTED"
          legalEntityName={isAr ? "شركة الأفق المتقدمة المحدودة" : "Al-Ufuq Advanced Co. Ltd."}
          branchName={isAr ? "فرع الرياض الرئيسي" : "Riyadh Main Branch"}
          documentDate="2026-09-18"
          dueDate="2026-10-18"
          createdByName={isAr ? "سارة العتيبي" : "Sarah Al-Otaibi"}
          isArabic={isAr}
        >
          <ZATCAStatusBadge status="CLEARED" isArabic={isAr} />
        </DocumentHeader>

        {/* Demo DocumentActions */}
        <DocumentActions
          status="SUBMITTED"
          onSaveDraft={() => alert("Draft saved")}
          onSubmit={() => alert("Submitted")}
          onApprove={() => alert("Approved successfully")}
          onReject={() => alert("Rejected")}
          onPost={() => alert("Posted to GL")}
          onPrint={() => window.print()}
          onExport={() => alert("Exporting PDF...")}
          onOpenAudit={() => setAuditOpen(true)}
          isArabic={isAr}
        />

        {/* Demo ApprovalTimeline */}
        <ApprovalTimeline steps={sampleApprovalSteps} isArabic={isAr} />

        {/* Demo Row: TotalsPanel & MoneyInput */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
          <Card>
            <CardHeader>
              <CardTitle>{isAr ? "حقل إدخال المبالغ النقدية (MoneyInput)" : "Precision MoneyInput"}</CardTitle>
              <CardDescription>
                {isAr
                  ? "إدخال دقيق للمبالغ المالية بالريال السعودي مع التنسيق المحاسبي والتحقق من الكسور"
                  : "Deterministic decimal arithmetic input with SAR currency badge and numeric validation."}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-foreground">
                  {isAr ? "المبلغ المراد سداده:" : "Amount to Disburse:"}
                </label>
                <MoneyInput
                  value={moneyVal}
                  onChange={(val) => setMoneyVal(val)}
                  currency="SAR"
                  isArabic={isAr}
                />
              </div>

              <div className="p-3 rounded-lg bg-muted/50 border text-xs space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "القيمة الرقمية المخزنة:" : "Raw Stored Value:"}</span>
                  <span className="font-mono font-bold text-foreground">{moneyVal}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "التنسيق المحاسبي:" : "Formatted Accounting:"}</span>
                  <span className="font-mono font-bold text-primary">
                    {new Intl.NumberFormat(isAr ? "ar-SA" : "en-US", { style: "currency", currency: "SAR" }).format(moneyVal)}
                  </span>
                </div>
              </div>

              <div className="pt-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setDialogOpen(true)}
                  className="w-full gap-2"
                >
                  <FileText className="h-4 w-4" />
                  <span>{isAr ? "فتح نافذة التأكيد (Modal Dialog Test)" : "Open Confirmation Dialog"}</span>
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Demo TotalsPanel */}
          <TotalsPanel
            subtotal={48500.0}
            discountTotal={1500.0}
            taxableAmount={47000.0}
            taxRate={15}
            taxAmount={7050.0}
            grandTotal={54050.0}
            currency="SAR"
            isArabic={isAr}
          />
        </div>
      </div>

      {/* 3. Status Badges Catalog */}
      <Card>
        <CardHeader>
          <CardTitle>{isAr ? "كتالوج حالات المستندات (Document Status Badges)" : "Document Status Badges Catalog"}</CardTitle>
          <CardDescription>
            {isAr
              ? "مؤشرات الحالة الرسمية المعتمدة للمنظومة لجميع دورات العمل"
              : "Standardized badge indicators for document lifecycle progression."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <DocumentStatus status="DRAFT" isArabic={isAr} />
            <DocumentStatus status="SUBMITTED" isArabic={isAr} />
            <DocumentStatus status="PENDING_APPROVAL" isArabic={isAr} />
            <DocumentStatus status="APPROVED" isArabic={isAr} />
            <DocumentStatus status="POSTED" isArabic={isAr} />
            <DocumentStatus status="CLEARED" isArabic={isAr} />
            <DocumentStatus status="REJECTED" isArabic={isAr} />
            <DocumentStatus status="CANCELLED" isArabic={isAr} />
          </div>
        </CardContent>
      </Card>

      {/* 4. DataTable Component Demonstration */}
      <Card>
        <CardHeader>
          <CardTitle>{isAr ? "جدول البيانات المعياري (Standard DataTable)" : "Standard DataTable with Search & Sorting"}</CardTitle>
          <CardDescription>
            {isAr
              ? "دعم كامل للفرز الأبجدي والرقمي، البحث اللحظي، والترقيم، والتوافق مع الـ RTL/LTR"
              : "Fully responsive, sortable, paginated data table with search filtering and logical layout."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <DataTable
            data={sampleTableData}
            columns={tableColumns}
            searchKey="name"
            pageSize={4}
            isArabic={isAr}
            actionsSlot={
              <Button size="sm" className="gap-1.5">
                <Plus className="h-3.5 w-3.5" />
                <span>{isAr ? "سند جديد" : "New Record"}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      {/* 5. Metric Cards Grid */}
      <div className="space-y-4">
        <div className="border-b pb-2">
          <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
            <Scale className="h-5 w-5 text-primary" />
            <span>{isAr ? "بطاقات المؤشرات (KPIMetricCard)" : "Executive Metric Cards"}</span>
          </h2>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <KPIMetricCard
            title={isAr ? "المبيعات الإجمالية" : "Total Revenue"}
            value="1,450,000 SAR"
            change="+14.2%"
            changeType="positive"
            changeNote={isAr ? "نمو ربع سنوي" : "QoQ growth"}
            icon={DollarSign}
            accentColor="emerald"
            progressPercent={78}
          />

          <KPIMetricCard
            title={isAr ? "المصروفات التشغيلية" : "Operating Expenses"}
            value="920,000 SAR"
            change="+6.4%"
            changeType="negative"
            changeNote={isAr ? "تحت السيطرة" : "within bounds"}
            icon={Scale}
            accentColor="rose"
            progressPercent={58}
          />

          <KPIMetricCard
            title={isAr ? "الذمم المدينة المفتوحة" : "Open Receivables"}
            value="680,000 SAR"
            change="14 فواتير"
            changeType="neutral"
            changeNote={isAr ? "متأخر 85 ألف" : "85k overdue"}
            icon={Users}
            accentColor="blue"
            progressPercent={65}
          />

          <KPIMetricCard
            title={isAr ? "تكامل زاتكا (المرحلة 2)" : "ZATCA Integration"}
            value="98.5%"
            change="142 مخلصة"
            changeType="positive"
            icon={ShieldCheck}
            accentColor="teal"
            progressPercent={98.5}
          />
        </div>
      </div>

      {/* Interactive Modal Dialog Test */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogHeader>
          <DialogTitle>{isAr ? "تأكيد ترحيل القيد المحاسبي" : "Confirm Journal Posting"}</DialogTitle>
          <DialogDescription>
            {isAr
              ? "ترحيل القيد عملية نهائية غير قابلة للإلغاء أو الحذف وتوثق مباشرة في الدفتر العام وسلسلة التدقيق."
              : "Posting a journal entry is an immutable operation that commits to the General Ledger and audit trail."}
          </DialogDescription>
        </DialogHeader>
        <div className="py-4 space-y-2 text-xs">
          <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-700 dark:text-amber-300">
            {isAr
              ? "تحذير نظامي: توازن المدين والدائن إلزامي بموجب القسم 40.01."
              : "System Invariant: Total Debit must equal Total Credit per Section 40.01."}
          </div>
          <p className="text-muted-foreground">
            {isAr ? "المبلغ الإجمالي للقيد: 54,050.00 ر.س" : "Total Entry Amount: 54,050.00 SAR"}
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" size="sm" onClick={() => setDialogOpen(false)}>
            {isAr ? "إلغاء" : "Cancel"}
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={() => {
              setDialogOpen(false);
              alert(isAr ? "تم الترحيل بنجاح" : "Successfully posted");
            }}
          >
            {isAr ? "تأكيد الترحيل" : "Confirm Post"}
          </Button>
        </DialogFooter>
      </Dialog>

      {/* Interactive Audit Drawer Slide-over Test */}
      <AuditDrawer
        open={auditOpen}
        onClose={() => setAuditOpen(false)}
        documentTitle={isAr ? "فاتورة توريد خوادم #INV-2026-0042" : "Supply Invoice #INV-2026-0042"}
        records={sampleAuditRecords}
        isArabic={isAr}
      />
    </div>
  );
}

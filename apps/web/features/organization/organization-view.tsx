"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Building2,
  Building,
  MapPin,
  FolderTree,
  Coins,
  Hash,
  Sparkles,
  CheckCircle2,
  ShieldCheck,
  Globe,
  Briefcase,
  ChevronRight,
  ChevronLeft,
} from "lucide-react";

export function OrganizationView() {
  const { t, locale, dir } = useLocale();

  const [activeTab, setActiveTab] = useState<"entities" | "branches" | "departments" | "cost_centers" | "sequences">("entities");
  const [generating, setGenerating] = useState(false);
  const [generatedSequence, setGeneratedSequence] = useState<string | null>("INV-2026-000001");
  const [sequenceCounter, setSequenceCounter] = useState(1);

  const isRTL = dir === "rtl";
  const ArrowIcon = isRTL ? ChevronLeft : ChevronRight;

  const handleGenerateSequence = () => {
    setGenerating(true);
    setTimeout(() => {
      const nextNum = sequenceCounter + 1;
      setSequenceCounter(nextNum);
      const padded = String(nextNum).padStart(6, "0");
      setGeneratedSequence(`INV-2026-${padded}`);
      setGenerating(false);
    }, 300);
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary mb-2">
            <Briefcase className="h-3.5 w-3.5" />
            <span>{t.app.phase}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.organization.title}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t.organization.subtitle}
          </p>
        </div>

        {/* Global Holding Badge */}
        <div className="flex items-center gap-3 p-3 rounded-xl border bg-card/60 shadow-sm">
          <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
            <Building2 className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-muted-foreground font-medium">{t.organization.holding_title}</div>
            <div className="text-sm font-bold text-foreground">
              {locale === "ar" ? "مجموعة الأفق السعودية القابضة" : "Saudi Horizon Holding Group"}
            </div>
          </div>
          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            {t.organization.active}
          </span>
        </div>
      </div>

      {/* Financial Isolation Callout */}
      <div className="p-4 rounded-xl bg-primary/5 border border-primary/20 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="flex items-start gap-3">
          <ShieldCheck className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
          <div>
            <h4 className="text-sm font-semibold text-foreground">
              {locale === "ar"
                ? "عزل النطاق المالي والمحاسبي (Legal Entity Financial Isolation Boundary)"
                : "Legal Entity Financial Isolation Boundary (Section 4)"}
            </h4>
            <p className="text-xs text-muted-foreground mt-0.5">
              {locale === "ar"
                ? "يتم عزل الحسابات والقيود والعمليات الضريبية لكل كيان قانوني بشكل مستقل وفق معايير الحوكمة المالية."
                : "Accounts, journal postings, and tax obligations are strictly partitioned by legal_entity_id."}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 self-stretch sm:self-auto">
          <span className="text-xs font-mono bg-background px-3 py-1.5 rounded-md border text-foreground">
            legal_entity_id: LE-01
          </span>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex overflow-x-auto border-b border-border gap-2 pb-px">
        {[
          { id: "entities", label: t.organization.legal_entities_title, icon: Building },
          { id: "branches", label: t.organization.branches_title, icon: MapPin },
          { id: "departments", label: t.organization.departments_title, icon: FolderTree },
          { id: "cost_centers", label: t.organization.cost_centers_title, icon: Coins },
          { id: "sequences", label: t.organization.number_sequences_title, icon: Hash },
        ].map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
                isActive
                  ? "border-primary text-primary bg-primary/5 rounded-t-lg"
                  : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30"
              }`}
            >
              <Icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      <div className="space-y-6">
        {/* 1. Legal Entities */}
        {activeTab === "entities" && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="rounded-xl border bg-card p-5 space-y-4 shadow-sm">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
                    <Building className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold text-foreground">
                      {locale === "ar" ? "شركة الأفق المتقدمة للتجارة والتقنية" : "Horizon Advanced Trading & Technology Co."}
                    </h3>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                      <span className="font-mono">LE-01</span>
                      <span>•</span>
                      <span>{t.organization.currency}: SAR</span>
                    </div>
                  </div>
                </div>
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  {t.organization.active}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-3 pt-3 border-t text-xs">
                <div>
                  <span className="text-muted-foreground block">{t.organization.cr_number}</span>
                  <span className="font-semibold font-mono text-foreground">1010789456 (KSA CR)</span>
                </div>
                <div>
                  <span className="text-muted-foreground block">الرقم الضريبي (VAT No)</span>
                  <span className="font-semibold font-mono text-foreground">300123456700003</span>
                </div>
              </div>

              <div className="p-3 rounded-lg bg-muted/40 text-xs flex items-center justify-between">
                <span className="text-muted-foreground">الفروع التابعة: 1 فرع رئيسي</span>
                <span className="text-primary font-medium flex items-center gap-1 cursor-pointer" onClick={() => setActiveTab("branches")}>
                  {locale === "ar" ? "عرض الفروع" : "View Branches"}
                  <ArrowIcon className="h-3.5 w-3.5" />
                </span>
              </div>
            </div>
          </div>
        )}

        {/* 2. Branches & National Address */}
        {activeTab === "branches" && (
          <div className="space-y-4">
            <div className="rounded-xl border bg-card p-5 space-y-4 shadow-sm">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
                    <MapPin className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold text-foreground">
                      {locale === "ar" ? "الفرع الرئيسي - الرياض" : "Riyadh Main Branch"}
                    </h3>
                    <span className="text-xs font-mono text-muted-foreground">BR-RUH</span>
                  </div>
                </div>
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  {t.organization.active}
                </span>
              </div>

              {/* National Address Card */}
              <div className="rounded-lg border bg-muted/30 p-4 space-y-2">
                <div className="flex items-center gap-2 text-xs font-bold text-foreground">
                  <Globe className="h-4 w-4 text-primary" />
                  <span>{locale === "ar" ? "العنوان الوطني المعتمد (KSA National Address)" : "Certified KSA National Address"}</span>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs pt-1">
                  <div>
                    <span className="text-muted-foreground block">رقم المبنى</span>
                    <span className="font-semibold font-mono text-foreground">7534</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block">اسم الشارع والحي</span>
                    <span className="font-semibold text-foreground">
                      {locale === "ar" ? "طريق الملك فهد، العليا" : "King Fahd Rd, Al Olaya"}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block">المدينة والرمز البريدي</span>
                    <span className="font-semibold text-foreground">
                      {locale === "ar" ? "الرياض 12214" : "Riyadh 12214"}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block">الرقم الإضافي</span>
                    <span className="font-semibold font-mono text-foreground">2345</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 3. Departments Hierarchy */}
        {activeTab === "departments" && (
          <div className="rounded-xl border bg-card p-5 space-y-4 shadow-sm">
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              <FolderTree className="h-4 w-4 text-primary" />
              <span>{locale === "ar" ? "الهيكل الإداري الشجري (Organizational Hierarchy)" : "Department Tree"}</span>
            </h3>

            <div className="space-y-3">
              {/* Root */}
              <div className="p-3 rounded-lg border bg-background flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div className="h-8 w-8 rounded bg-primary/10 text-primary flex items-center justify-center font-mono text-xs font-bold">
                    MGMT
                  </div>
                  <div>
                    <div className="text-xs sm:text-sm font-bold text-foreground">
                      {locale === "ar" ? "الإدارة العامة التنفيذية" : "General Executive Management"}
                    </div>
                    <div className="text-[11px] text-muted-foreground"> المستوى الأول • 3 أقسام تابعة</div>
                  </div>
                </div>
                <span className="text-xs px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 font-medium">نشط</span>
              </div>

              {/* Children */}
              <div className="ps-6 space-y-2 border-s-2 border-primary/20 ms-4">
                {[
                  { code: "FIN", ar: "الإدارة المالية والمحاسبة", en: "Finance & Accounting", desc: "الحسابات والميزانيات والضرائب" },
                  { code: "HR", ar: "إدارة الموارد البشرية", en: "Human Resources", desc: "التوظيف والرواتب والتأمينات GOSI" },
                  { code: "IT", ar: "تقنية المعلومات والتحول الرقمي", en: "Information Technology", desc: "البنية التحتية والفوترة الإلكترونية ZATCA" },
                ].map((dept) => (
                  <div key={dept.code} className="p-2.5 rounded-lg border bg-card flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs bg-muted px-2 py-0.5 rounded font-semibold">{dept.code}</span>
                      <span className="text-xs font-medium text-foreground">{locale === "ar" ? dept.ar : dept.en}</span>
                    </div>
                    <span className="text-[11px] text-muted-foreground hidden sm:inline">{dept.desc}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* 4. Cost Centers */}
        {activeTab === "cost_centers" && (
          <div className="rounded-xl border bg-card p-5 space-y-4 shadow-sm">
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              <Coins className="h-4 w-4 text-primary" />
              <span>{locale === "ar" ? "مراكز التكلفة التحليلية (Financial Dimensions)" : "Cost Center Tree"}</span>
            </h3>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {[
                { code: "CC-ALL", ar: "مركز التكلفة العام للمنشأة", en: "Company General Center", type: "المستوى الرئيسي" },
                { code: "CC-ADM", ar: "المصاريف الإدارية والعمومية", en: "Administrative Expenses", type: "مصاريف تشغيلية" },
                { code: "CC-OPS", ar: "العمليات التشغيلية والخدمات", en: "Operations & Services", type: "تكاليف مباشرة" },
              ].map((cc) => (
                <div key={cc.code} className="p-3.5 rounded-xl border bg-muted/20 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs font-bold text-primary bg-primary/10 px-2 py-0.5 rounded">
                      {cc.code}
                    </span>
                    <span className="text-[11px] text-muted-foreground">{cc.type}</span>
                  </div>
                  <div className="text-xs sm:text-sm font-bold text-foreground">
                    {locale === "ar" ? cc.ar : cc.en}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 5. Number Sequences & Atomic Generation Engine */}
        {activeTab === "sequences" && (
          <div className="space-y-5">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {[
                { code: "INV-SALES", prefix: "INV", sample: "INV-2026-000001", ar: "فواتير المبيعات الضريبية", en: "Tax Invoices", reset: "سنوي (YEARLY)" },
                { code: "PO-PURCHASE", prefix: "PO", sample: "PO-2026-000001", ar: "أوامر الشراء المعتمدة", en: "Purchase Orders", reset: "سنوي (YEARLY)" },
                { code: "JOURNAL-ENTRY", prefix: "JV", sample: "JV-2026-000001", ar: "قيود اليومية العامة", en: "Journal Vouchers", reset: "سنوي (YEARLY)" },
                { code: "EMPLOYEE", prefix: "EMP", sample: "EMP-1001", ar: "أرقام الموظفين الوظيفية", en: "Employee Numbers", reset: "بدون تصفير (NEVER)" },
              ].map((seq) => (
                <div key={seq.code} className="p-4 rounded-xl border bg-card space-y-3 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Hash className="h-4 w-4 text-primary" />
                      <span className="font-bold font-mono text-xs text-foreground">{seq.code}</span>
                    </div>
                    <span className="text-[11px] px-2 py-0.5 rounded bg-muted text-muted-foreground font-mono">
                      {seq.reset}
                    </span>
                  </div>
                  <div className="text-xs font-medium text-foreground">
                    {locale === "ar" ? seq.ar : seq.en}
                  </div>
                  <div className="flex items-center justify-between text-xs pt-2 border-t text-muted-foreground">
                    <span>صيغة الرقم:</span>
                    <span className="font-mono font-semibold text-primary">{seq.sample}</span>
                  </div>
                </div>
              ))}
            </div>

            {/* Live Sequence Generation Tester */}
            <div className="p-5 rounded-xl border bg-card/60 shadow-sm space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-sm font-bold text-foreground flex items-center gap-2">
                    <Sparkles className="h-4 w-4 text-amber-500" />
                    <span>{locale === "ar" ? "محاكي الترقيم الذري (Pessimistic Lock Demo)" : "Atomic Numbering Engine Demo"}</span>
                  </h4>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {locale === "ar"
                      ? "اضغط لتوليد الرقم التسلسلي التالي لفاتورة المبيعات مع ضمان منع التكرار والتزامن التام."
                      : "Generate the next consecutive invoice sequence guaranteed by database pessimistic locks."}
                  </p>
                </div>
                <button
                  id="generate-sequence-btn"
                  onClick={handleGenerateSequence}
                  disabled={generating}
                  className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90 disabled:opacity-50 transition-colors"
                >
                  <Hash className="h-4 w-4" />
                  <span>{generating ? t.organization.generating : t.organization.test_sequence_button}</span>
                </button>
              </div>

              {generatedSequence && (
                <div className="p-3.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-700 dark:text-emerald-300 flex items-center justify-between">
                  <div className="flex items-center gap-2 text-xs">
                    <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
                    <span>{t.organization.generated_badge}</span>
                  </div>
                  <span id="generated-sequence-val" className="font-mono text-sm font-extrabold tracking-wider bg-background/80 px-3 py-1 rounded border">
                    {generatedSequence}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

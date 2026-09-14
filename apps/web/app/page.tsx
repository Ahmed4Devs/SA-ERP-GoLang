"use client";

import React from "react";
import { useLocale } from "@/components/layout/shell";
import {
  CheckCircle2,
  ShieldCheck,
  Server,
  Database,
  Building,
  Scale,
  FileSpreadsheet,
  ArrowRight,
  Sparkles,
} from "lucide-react";

export default function HomePage() {
  const { t, locale, dir } = useLocale();

  return (
    <div className="space-y-6">
      {/* Hero Welcome Card */}
      <div className="relative overflow-hidden rounded-xl border bg-card p-6 sm:p-8 shadow-sm">
        <div className="absolute top-0 inset-inline-end-0 w-96 h-96 bg-primary/5 rounded-full blur-3xl pointer-events-none -z-0" />
        <div className="relative z-10 max-w-3xl space-y-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
            <Sparkles className="h-3.5 w-3.5" />
            <span>{t.app.phase}</span>
          </div>

          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            {t.app.title}
          </h1>

          <p className="text-sm sm:text-base text-muted-foreground leading-relaxed">
            {t.app.subtitle}
          </p>

          {/* Platform Status Indicators */}
          <div className="pt-3 flex flex-wrap gap-2 sm:gap-3">
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              <CheckCircle2 className="h-3.5 w-3.5" />
              <span>{t.status.api_connected}</span>
            </div>
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
              <Database className="h-3.5 w-3.5" />
              <span>{t.status.db_ready}</span>
            </div>
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              <ShieldCheck className="h-3.5 w-3.5" />
              <span>{t.status.zatca_compliant}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Grid of Key System Modules */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
        {/* Module 1: Organization & Multi-Entity */}
        <div className="rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow">
          <div className="h-10 w-10 rounded-lg bg-blue-500/10 text-blue-600 flex items-center justify-center">
            <Building className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-semibold text-sm text-foreground">
              {locale === "ar" ? "الهيكل المؤسسي" : "Organization Structure"}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {locale === "ar"
                ? "عزل مالي وقانوني كامل عبر legal_entity_id مع دعم متعدد الفروع ومراكز التكلفة."
                : "Full legal-entity isolation boundary with multi-branch and cost-center hierarchies."}
            </p>
          </div>
          <div className="pt-2 border-t text-[11px] font-medium text-primary flex items-center justify-between">
            <span>{locale === "ar" ? "الحد العزلي: صارم" : "Isolation: Strict"}</span>
            <span className="text-emerald-500">Active</span>
          </div>
        </div>

        {/* Module 2: Accounting & General Ledger */}
        <div className="rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow">
          <div className="h-10 w-10 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center">
            <Scale className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-semibold text-sm text-foreground">
              {locale === "ar" ? "المحاسبة ودفتر الأستاذ" : "General Ledger"}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {locale === "ar"
                ? "محرك قيود غير قابل للتعديل (Immutable Journals) مع توازن مدين/دائن إلزامي."
                : "Immutable posted journals, transactional balance enforcement, and exact arithmetic."}
            </p>
          </div>
          <div className="pt-2 border-t text-[11px] font-medium text-primary flex items-center justify-between">
            <span>{locale === "ar" ? "الحسابات: NUMERIC(24,6)" : "Precision: Exact"}</span>
            <span className="text-emerald-500">Locked</span>
          </div>
        </div>

        {/* Module 3: ZATCA E-Invoicing */}
        <div className="rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow">
          <div className="h-10 w-10 rounded-lg bg-amber-500/10 text-amber-600 flex items-center justify-center">
            <FileSpreadsheet className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-semibold text-sm text-foreground">
              {locale === "ar" ? "الفوترة الإلكترونية (زاتكا)" : "ZATCA E-Invoicing"}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {locale === "ar"
                ? "توليد XML المعتمد، التوقيع الرقمي ECDSA، الربط بالربط والتخليص للمرحلة الثانية."
                : "UBL 2.1 XML generation, cryptographic signing, clearance & reporting pipeline."}
            </p>
          </div>
          <div className="pt-2 border-t text-[11px] font-medium text-primary flex items-center justify-between">
            <span>{locale === "ar" ? "المرحلة: الثانية" : "Phase: 2 Ready"}</span>
            <span className="text-emerald-500">Ready</span>
          </div>
        </div>

        {/* Module 4: GOSI & Saudi Payroll */}
        <div className="rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow">
          <div className="h-10 w-10 rounded-lg bg-purple-500/10 text-purple-600 flex items-center justify-center">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-semibold text-sm text-foreground">
              {locale === "ar" ? "التأمينات الاجتماعية والرواتب" : "GOSI & Payroll"}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {locale === "ar"
                ? "محرك قواعد تاريخية محددة الصلاحية لنسب السعوديين وغير السعوديين دون ترميز ثابت."
                : "Effective-dated rule engine for Saudi/non-Saudi social insurance contributions."}
            </p>
          </div>
          <div className="pt-2 border-t text-[11px] font-medium text-primary flex items-center justify-between">
            <span>{locale === "ar" ? "القواعد: مؤرخة ومفصلة" : "Rules: Dynamic"}</span>
            <span className="text-emerald-500">Configured</span>
          </div>
        </div>
      </div>

      {/* Financial & Architectural Invariants Banner */}
      <div className="rounded-xl border bg-card p-6 space-y-4">
        <h2 className="text-base font-bold text-foreground">
          {locale === "ar"
            ? "الثوابت الهندسية والمالية غير القابلة للتفاوض (Non-Negotiable Invariants)"
            : "Non-Negotiable Architectural & Financial Invariants"}
        </h2>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 text-xs">
          <div className="p-3 rounded-lg bg-muted/50 border space-y-1">
            <p className="font-semibold text-foreground">
              {locale === "ar" ? "١. منع الحسابات العائمة (No Floats)" : "1. Exact Arithmetic"}
            </p>
            <p className="text-muted-foreground">
              {locale === "ar"
                ? "جميع العمليات المالية تستخدم NUMERIC(24,6) بدقة ثابتة تامة."
                : "All money and quantity operations use fixed exact decimal precision."}
            </p>
          </div>

          <div className="p-3 rounded-lg bg-muted/50 border space-y-1">
            <p className="font-semibold text-foreground">
              {locale === "ar" ? "٢. مناعة القيود المرحلة (Immutability)" : "2. Immutable Journals"}
            </p>
            <p className="text-muted-foreground">
              {locale === "ar"
                ? "لا يمكن تعديل أو حذف أي قيد محاسبي مرحل؛ التصحيح يتم فقط عبر قيود العكس."
                : "Posted entries and stock movements cannot be modified; corrections use reversal."}
            </p>
          </div>

          <div className="p-3 rounded-lg bg-muted/50 border space-y-1">
            <p className="font-semibold text-foreground">
              {locale === "ar" ? "٣. عزل الكيان القانوني (Legal Entity)" : "3. Multi-Entity Isolation"}
            </p>
            <p className="text-muted-foreground">
              {locale === "ar"
                ? "حماية القيود المشتركة بمفاتيح أجنبية مركبة تمنع التداخل بين الشركات التابعة."
                : "Composite foreign keys enforce legal entity isolation directly at the database layer."}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

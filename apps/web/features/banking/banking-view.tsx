"use client";

import React, { useState } from "react";
import {
  Building2,
  FileSpreadsheet,
  CheckCheck,
  Lock,
  Plus,
  ArrowDownLeft,
  ArrowUpRight,
  ShieldCheck,
  Sparkles,
  CheckCircle2,
  Clock,
  Link as LinkIcon,
  Search
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface BankAccountItem {
  id: string;
  code: string;
  bankNameEn: string;
  bankNameAr: string;
  branch: string;
  maskedAccount: string;
  iban: string;
  swift: string;
  currency: string;
  glCode: string;
  currentBalance: number;
  isPrimary: boolean;
  status: "ACTIVE" | "FROZEN" | "CLOSED";
}

interface StatementLineItem {
  id: string;
  lineNumber: number;
  date: string;
  description: string;
  reference: string;
  type: "CREDIT" | "DEBIT";
  amount: number;
  balanceAfter: number;
  status: "UNMATCHED" | "MATCHED";
}

interface InternalLedgerItem {
  id: string;
  date: string;
  type: string;
  reference: string;
  partyOrDescription: string;
  amount: number;
  matched: boolean;
}

export function BankingView() {
  const { t, locale, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"accounts" | "statements" | "reconciliation">("accounts");

  // Bank Accounts Data
  const [bankAccounts] = useState<BankAccountItem[]>([
    {
      id: "ba-1",
      code: "BANK-RAJHI-01",
      bankNameEn: "Al-Rajhi Bank",
      bankNameAr: "مصرف الراجحي",
      branch: "فرع العليا - الرياض",
      maskedAccount: "**** **** **** 1234",
      iban: "SA4480000204608010001234",
      swift: "RJHISARI",
      currency: "SAR",
      glCode: "111002 (الحساب الجاري)",
      currentBalance: 185200.0,
      isPrimary: true,
      status: "ACTIVE",
    },
    {
      id: "ba-2",
      code: "BANK-SNB-01",
      bankNameEn: "Saudi National Bank (SNB)",
      bankNameAr: "البنك الأهلي السعودي",
      branch: "فرع طريق الملك فهد",
      maskedAccount: "**** **** **** 4321",
      iban: "SA8810000101000987654321",
      swift: "NCBKSARI",
      currency: "SAR",
      glCode: "111003 (حساب الشركات)",
      currentBalance: 94000.0,
      isPrimary: false,
      status: "ACTIVE",
    },
  ]);

  // Statement Lines
  const [statementLines, setStatementLines] = useState<StatementLineItem[]>([
    {
      id: "stl-1",
      lineNumber: 1,
      date: "2026-09-16",
      description: "Cash Deposit Branch 01",
      reference: "DEP-RAJHI-8891",
      type: "CREDIT",
      amount: 5000.0,
      balanceAfter: 105000.0,
      status: "MATCHED",
    },
    {
      id: "stl-2",
      lineNumber: 2,
      date: "2026-09-16",
      description: "Client Wire Payment - Horizons Co",
      reference: "WIRE-9920",
      type: "CREDIT",
      amount: 15400.0,
      balanceAfter: 120400.0,
      status: "UNMATCHED",
    },
    {
      id: "stl-3",
      lineNumber: 3,
      date: "2026-09-16",
      description: "Monthly Account Maintenance Fee",
      reference: "FEE-2026-01",
      type: "DEBIT",
      amount: 15.0,
      balanceAfter: 120385.0,
      status: "UNMATCHED",
    },
  ]);

  // ERP Internal Ledger Entries for Reconciliation
  const [ledgerEntries, setLedgerEntries] = useState<InternalLedgerItem[]>([
    {
      id: "leg-1",
      date: "2026-09-16",
      type: "تحويل نقدية (TRF)",
      reference: "DEP-RAJHI-8891",
      partyOrDescription: "إيداع صندوق نقطة البيع الرئيسي",
      amount: 5000.0,
      matched: true,
    },
    {
      id: "leg-2",
      date: "2026-09-16",
      type: "سند تحصيل عميل (REC)",
      reference: "WIRE-9920",
      partyOrDescription: "سداد فاتورة شركة الآفاق للمقاولات",
      amount: 15400.0,
      matched: false,
    },
    {
      id: "leg-3",
      date: "2026-09-16",
      type: "قيد تسوية عمولات (JE)",
      reference: "FEE-2026-01",
      partyOrDescription: "مصروف رسوم بنكية شهرية",
      amount: 15.0,
      matched: false,
    },
  ]);

  const [isReconLocked, setIsReconLocked] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleAutoMatch = () => {
    setStatementLines((prev) => prev.map((l) => ({ ...l, status: "MATCHED" })));
    setLedgerEntries((prev) => prev.map((e) => ({ ...e, matched: true })));
    showToast(
      locale === "ar"
        ? "تمت المطابقة الآلية بنجاح بنسبة 100% (2 حركات مطابقة بالمبلغ والمرجع)"
        : "Auto-Match completed successfully at 100% (2 items matched by amount & ref)"
    );
  };

  const handleLockReconciliation = () => {
    setIsReconLocked(true);
    showToast(
      locale === "ar"
        ? "تم إقفال واعتماد المطابقة البنكية وتوثيق الرصيد الدفتري"
        : "Reconciliation session locked and finalized successfully"
    );
  };

  const totalBankBalance = bankAccounts.reduce((acc, curr) => acc + curr.currentBalance, 0);
  const matchedCount = statementLines.filter((l) => l.status === "MATCHED").length;
  const matchRate = Math.round((matchedCount / statementLines.length) * 100);

  return (
    <div className="space-y-6">
      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed bottom-6 end-6 z-50 bg-primary text-primary-foreground px-5 py-3 rounded-xl shadow-2xl flex items-center gap-3 animate-in fade-in slide-in-from-bottom-5">
          <CheckCircle2 className="w-5 h-5 text-emerald-400" />
          <span className="font-medium text-sm">{toastMessage}</span>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground flex items-center gap-3">
            <span className="p-2.5 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 text-white shadow-md shadow-blue-500/20">
              <Building2 className="w-6 h-6" />
            </span>
            {(t.banking as any)?.title || "إدارة البنوك والمطابقة المصرفية"}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm sm:text-base">
            {(t.banking as any)?.description || "إدارة الحسابات المصرفية الرسمية، استيراد كشوف الحسابات الإلكترونية، ومحرك المطابقة والتسوية الآلية"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {activeTab === "reconciliation" ? (
            <>
              <button
                onClick={handleAutoMatch}
                disabled={isReconLocked}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white transition shadow-sm font-medium text-sm disabled:opacity-50"
              >
                <Sparkles className="w-4 h-4" />
                {(t.banking as any)?.btn_auto_match || "تشغيل المطابقة الآلية"}
              </button>
              <button
                onClick={handleLockReconciliation}
                disabled={isReconLocked}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm font-medium text-sm disabled:opacity-50"
              >
                <Lock className="w-4 h-4" />
                {isReconLocked
                  ? locale === "ar" ? "المطابقة مقفلة وموثقة" : "Reconciliation Locked"
                  : (t.banking as any)?.btn_lock || "إقفال واعتماد المطابقة"}
              </button>
            </>
          ) : (
            <button
              onClick={() => showToast(locale === "ar" ? "الحسابات البنكية مسجلة ومحمية بالتشفير" : "Bank accounts securely registered and masked")}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm font-medium text-sm"
            >
              <Plus className="w-4 h-4" />
              {(t.banking as any)?.btn_new_account || "حساب بنكي جديد"}
            </button>
          )}
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-blue-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.banking as any)?.kpi_total_bank_balance || "إجمالي الأرصدة البنكية"}
            </span>
            <span className="p-2 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <Building2 className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {totalBankBalance.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 font-medium">
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>حسابات مصرفية سعودية معتمدة</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-emerald-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.banking as any)?.kpi_recon_rate || "نسبة المطابقة البنكية"}
            </span>
            <span className="p-2 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <CheckCheck className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {matchRate}%
            </span>
            <span className="text-xs font-semibold text-muted-foreground">
              ({matchedCount}/{statementLines.length})
            </span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>{matchRate === 100 ? "مطابقة مكتملة ومغلقة" : "جاري استكمال المطابقة"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-purple-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.banking as any)?.kpi_active_accounts || "الحسابات النشطة"}
            </span>
            <span className="p-2 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <LinkIcon className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {bankAccounts.length}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">بنوك</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-purple-600 dark:text-purple-400 font-medium">
            <span>الراجحي + الأهلي السعودي</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-amber-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.banking as any)?.kpi_unmatched_items || "حركات قيد المطابقة"}
            </span>
            <span className="p-2 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <Clock className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {statementLines.length - matchedCount}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">حركة</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-amber-600 dark:text-amber-400 font-medium">
            <span>جاهزة للمطابقة الفورية</span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b space-x-2 rtl:space-x-reverse overflow-x-auto">
        <button
          onClick={() => setActiveTab("accounts")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "accounts"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Building2 className="w-4 h-4" />
          {(t.banking as any)?.tab_accounts || "الحسابات المصرفية"}
        </button>
        <button
          onClick={() => setActiveTab("statements")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "statements"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileSpreadsheet className="w-4 h-4" />
          {(t.banking as any)?.tab_statements || "كشوف الحسابات المستوردة"}
        </button>
        <button
          onClick={() => setActiveTab("reconciliation")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "reconciliation"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Sparkles className="w-4 h-4" />
          {(t.banking as any)?.tab_reconciliation || "محرك المطابقة والتسوية"}
        </button>
      </div>

      {/* Tab: Bank Accounts */}
      {activeTab === "accounts" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {bankAccounts.map((acc) => (
            <div
              key={acc.id}
              className="rounded-2xl border bg-card p-6 shadow-sm hover:shadow-md transition space-y-4 relative overflow-hidden"
            >
              {acc.isPrimary && (
                <div className="absolute top-0 end-0 bg-primary text-primary-foreground text-[10px] font-bold px-3 py-1 rounded-bl-xl shadow">
                  الحساب الرئيسي التشغيلي
                </div>
              )}
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-bold text-lg text-foreground">
                    {locale === "ar" ? acc.bankNameAr : acc.bankNameEn}
                  </h3>
                  <p className="text-xs text-muted-foreground mt-0.5">{acc.branch}</p>
                </div>
                <span className="p-2 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
                  <Building2 className="w-5 h-5" />
                </span>
              </div>

              <div className="p-3.5 rounded-xl bg-muted/40 font-mono text-sm space-y-1.5 border">
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>{(t.banking as any)?.col_account_num || "رقم الحساب"}:</span>
                  <span className="font-bold text-foreground">{acc.maskedAccount}</span>
                </div>
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>{(t.banking as any)?.col_iban || "الآيبان"}:</span>
                  <span className="font-bold text-foreground select-all">{acc.iban}</span>
                </div>
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>{(t.banking as any)?.col_swift || "سويفت"}:</span>
                  <span className="text-foreground">{acc.swift}</span>
                </div>
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>الربط بالدفتر العام (GL):</span>
                  <span className="text-primary font-semibold">{acc.glCode}</span>
                </div>
              </div>

              <div className="flex items-baseline justify-between pt-2 border-t">
                <span className="text-xs text-muted-foreground">{(t.banking as any)?.col_balance || "الرصيد المتاح"}:</span>
                <div className="text-end">
                  <span className="text-xl font-bold text-foreground font-mono">
                    {acc.currentBalance.toLocaleString()}
                  </span>
                  <span className="text-xs font-semibold text-muted-foreground ms-1.5">SAR</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Tab: Bank Statements */}
      {activeTab === "statements" && (
        <div className="space-y-4">
          <div className="rounded-2xl border bg-card p-5 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono font-bold text-primary">STMT-2026-01</span>
                <span className="px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  {isReconLocked ? "مطابق ومعتمد بالكامل" : "قيد المطابقة"}
                </span>
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                مصرف الراجحي — كشف الحساب الإلكتروني الموحد (rajhi_statement_jan2026.csv)
              </p>
            </div>
            <div className="flex items-center gap-6 font-mono text-sm">
              <div>
                <span className="text-xs text-muted-foreground block">إجمالي الإيداعات</span>
                <span className="font-bold text-emerald-600 dark:text-emerald-400">+20,400.00 SAR</span>
              </div>
              <div>
                <span className="text-xs text-muted-foreground block">إجمالي المسحوبات</span>
                <span className="font-bold text-rose-600 dark:text-rose-400">-15.00 SAR</span>
              </div>
            </div>
          </div>

          <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-start">
                <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
                  <tr>
                    <th className="py-3 px-4 text-start">#</th>
                    <th className="py-3 px-4 text-start">تاريخ المعاملة</th>
                    <th className="py-3 px-4 text-start">الوصف / البيان البنكي</th>
                    <th className="py-3 px-4 text-start">المرجع البنكي</th>
                    <th className="py-3 px-4 text-start">النوع</th>
                    <th className="py-3 px-4 text-end">المبلغ</th>
                    <th className="py-3 px-4 text-center">حالة المطابقة</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {statementLines.map((l) => (
                    <tr key={l.id} className="hover:bg-muted/20 transition">
                      <td className="py-3 px-4 text-muted-foreground font-mono">{l.lineNumber}</td>
                      <td className="py-3 px-4 text-muted-foreground">{l.date}</td>
                      <td className="py-3 px-4 font-medium text-foreground">{l.description}</td>
                      <td className="py-3 px-4 font-mono text-xs text-muted-foreground">{l.reference}</td>
                      <td className="py-3 px-4">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold ${
                          l.type === "CREDIT"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : "bg-rose-500/10 text-rose-600 dark:text-rose-400"
                        }`}>
                          {l.type === "CREDIT" ? "إيداع (Credit)" : "خصم (Debit)"}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-foreground">
                        {l.amount.toLocaleString()} SAR
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          l.status === "MATCHED"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                        }`}>
                          {l.status === "MATCHED" ? "مطابق" : "غير مطابق"}
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

      {/* Tab: Smart Reconciliation Engine */}
      {activeTab === "reconciliation" && (
        <div className="space-y-6">
          {/* Split Screen Reconciliation View */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Left: Statement Lines */}
            <div className="rounded-2xl border bg-card p-5 shadow-sm space-y-4">
              <div className="flex items-center justify-between border-b pb-3">
                <div className="flex items-center gap-2">
                  <FileSpreadsheet className="w-5 h-5 text-blue-500" />
                  <h3 className="font-bold text-foreground">حركات كشف الحساب البنكي (Bank Statement)</h3>
                </div>
                <span className="text-xs font-mono text-muted-foreground">{statementLines.length} حركات</span>
              </div>

              <div className="space-y-3">
                {statementLines.map((line) => (
                  <div
                    key={line.id}
                    className={`p-3.5 rounded-xl border transition flex items-center justify-between ${
                      line.status === "MATCHED"
                        ? "bg-emerald-500/5 border-emerald-500/30"
                        : "bg-card border-border hover:border-primary/40"
                    }`}
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm text-foreground">{line.description}</span>
                        {line.status === "MATCHED" && (
                          <span className="p-0.5 rounded-full bg-emerald-500 text-white">
                            <CheckCheck className="w-3 h-3" />
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground font-mono flex items-center gap-3">
                        <span>{line.date}</span>
                        <span>المرجع: {line.reference}</span>
                      </div>
                    </div>
                    <div className="text-end">
                      <span className={`font-mono font-bold text-sm ${
                        line.type === "CREDIT"
                          ? "text-emerald-600 dark:text-emerald-400"
                          : "text-rose-600 dark:text-rose-400"
                      }`}>
                        {line.type === "CREDIT" ? "+" : "-"}{line.amount.toLocaleString()} SAR
                      </span>
                      <span className="block text-[11px] text-muted-foreground">
                        {line.status === "MATCHED" ? "مطابق" : "قيد الانتظار"}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Right: ERP Internal Ledger Entries */}
            <div className="rounded-2xl border bg-card p-5 shadow-sm space-y-4">
              <div className="flex items-center justify-between border-b pb-3">
                <div className="flex items-center gap-2">
                  <ShieldCheck className="w-5 h-5 text-indigo-500" />
                  <h3 className="font-bold text-foreground">الحركات الدفترية في النظام (ERP GL Ledger)</h3>
                </div>
                <span className="text-xs font-mono text-muted-foreground">{ledgerEntries.length} قيود وسندات</span>
              </div>

              <div className="space-y-3">
                {ledgerEntries.map((entry) => (
                  <div
                    key={entry.id}
                    className={`p-3.5 rounded-xl border transition flex items-center justify-between ${
                      entry.matched
                        ? "bg-emerald-500/5 border-emerald-500/30"
                        : "bg-card border-border hover:border-primary/40"
                    }`}
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm text-foreground">{entry.partyOrDescription}</span>
                        <span className="text-[10px] px-2 py-0.5 rounded bg-muted text-muted-foreground font-mono">
                          {entry.type}
                        </span>
                      </div>
                      <div className="text-xs text-muted-foreground font-mono flex items-center gap-3">
                        <span>{entry.date}</span>
                        <span>المرجع: {entry.reference}</span>
                      </div>
                    </div>
                    <div className="text-end">
                      <span className="font-mono font-bold text-sm text-foreground">
                        {entry.amount.toLocaleString()} SAR
                      </span>
                      <span className="block text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
                        {entry.matched ? "مرتبط بالبنك" : "بانتظار المطابقة"}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

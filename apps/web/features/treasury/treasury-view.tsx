"use client";

import React, { useState } from "react";
import {
  Wallet,
  ArrowDownLeft,
  ArrowUpRight,
  ArrowLeftRight,
  Plus,
  CheckCircle2,
  Clock,
  Building2,
  ShieldCheck,
  AlertCircle,
  FileText,
  DollarSign
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface CashAccountItem {
  id: string;
  code: string;
  nameEn: string;
  nameAr: string;
  type: "PETTY_CASH" | "CASH_REGISTER" | "MAIN_VAULT";
  branch: string;
  balance: number;
  limit: number;
  status: "ACTIVE" | "SUSPENDED" | "CLOSED";
}

interface ReceiptItem {
  id: string;
  number: string;
  date: string;
  party: string;
  account: string;
  amount: number;
  method: "CASH" | "CHEQUE";
  status: "DRAFT" | "POSTED" | "VOIDED";
}

interface PaymentItem {
  id: string;
  number: string;
  date: string;
  party: string;
  account: string;
  amount: number;
  type: string;
  status: "DRAFT" | "APPROVED" | "PAID" | "VOIDED";
}

interface TransferItem {
  id: string;
  number: string;
  date: string;
  source: string;
  destination: string;
  amount: number;
  fee: number;
  status: "DRAFT" | "IN_TRANSIT" | "COMPLETED" | "CANCELLED";
}

export function TreasuryView() {
  const { t, locale, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"accounts" | "receipts" | "payments" | "transfers">("accounts");

  // Initial State Data matching Seed Data
  const [accounts, setAccounts] = useState<CashAccountItem[]>([
    {
      id: "ca-1",
      code: "CASH-HO-01",
      nameEn: "Head Office Petty Cash",
      nameAr: "صندوق المصروفات النثرية - الإدارة العامة",
      type: "PETTY_CASH",
      branch: "الفرع الرئيسي - الرياض",
      balance: 3500.0,
      limit: 5000.0,
      status: "ACTIVE",
    },
    {
      id: "ca-2",
      code: "CASH-REG-01",
      nameEn: "Main Branch POS Register",
      nameAr: "صندوق نقطة بيع الفرع الرئيسي",
      type: "CASH_REGISTER",
      branch: "الفرع الرئيسي - الرياض",
      balance: 12450.0,
      limit: 20000.0,
      status: "ACTIVE",
    },
  ]);

  const [receipts, setReceipts] = useState<ReceiptItem[]>([
    {
      id: "rec-1",
      number: "REC-2026-0001",
      date: "2026-09-16",
      party: "مبيعات نقدية مباشرة للفرع",
      account: "CASH-REG-01",
      amount: 2500.0,
      method: "CASH",
      status: "POSTED",
    },
  ]);

  const [payments, setPayments] = useState<PaymentItem[]>([
    {
      id: "pmt-1",
      number: "PMT-2026-0001",
      date: "2026-09-16",
      party: "شركة جرير للمستلزمات المكتبية",
      account: "CASH-HO-01",
      amount: 350.0,
      type: "PETTY_CASH_EXPENSE",
      status: "PAID",
    },
  ]);

  const [transfers, setTransfers] = useState<TransferItem[]>([
    {
      id: "trf-1",
      number: "TRF-2026-0001",
      date: "2026-09-16",
      source: "CASH-REG-01 (صندوق الفرع)",
      destination: "مصرف الراجحي (حساب جاري)",
      amount: 5000.0,
      fee: 0.0,
      status: "COMPLETED",
    },
  ]);

  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handlePostReceipt = (id: string) => {
    setReceipts((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: "POSTED" } : r))
    );
    showToast(locale === "ar" ? "تم ترحيل سند القبض وتحديث رصيد الصندوق بنجاح" : "Receipt posted and cash updated successfully");
  };

  const handlePayPayment = (id: string) => {
    setPayments((prev) =>
      prev.map((p) => (p.id === id ? { ...p, status: "PAID" } : p))
    );
    showToast(locale === "ar" ? "تم صرف المبلغ وخصمه من رصيد العهدة بنجاح" : "Payment disbursed from cash drawer successfully");
  };

  const handleCompleteTransfer = (id: string) => {
    setTransfers((prev) =>
      prev.map((t) => (t.id === id ? { ...t, status: "COMPLETED" } : t))
    );
    showToast(locale === "ar" ? "تم تأكيد التحويل المالي وإغلاق العملية" : "Transfer completed successfully");
  };

  const totalCash = accounts.reduce((acc, curr) => acc + curr.balance, 0);
  const pettyCash = accounts.filter((a) => a.type === "PETTY_CASH").reduce((acc, curr) => acc + curr.balance, 0);
  const totalReceipts = receipts.reduce((acc, curr) => acc + curr.amount, 0);

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
            <span className="p-2.5 rounded-xl bg-gradient-to-tr from-emerald-600 to-teal-500 text-white shadow-md shadow-emerald-500/20">
              <Wallet className="w-6 h-6" />
            </span>
            {(t.treasury as any)?.title || "الخزينة والسيولة النقدية"}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm sm:text-base">
            {(t.treasury as any)?.description || "إدارة صناديق النقدية، العهد النثرية، سندات القبض والصرف، والتحويلات المالية الداخلية"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => {
              if (activeTab === "receipts") {
                const newRec: ReceiptItem = {
                  id: "rec-" + (receipts.length + 1),
                  number: `REC-2026-${String(receipts.length + 1).padStart(4, "0")}`,
                  date: new Date().toISOString().split("T")[0],
                  party: "عميل تجزئة / نقدي",
                  account: "CASH-REG-01",
                  amount: 1200.0,
                  method: "CASH",
                  status: "DRAFT",
                };
                setReceipts([newRec, ...receipts]);
                showToast(locale === "ar" ? "تم إنشاء سند قبض مسودة جديد" : "New draft receipt voucher created");
              } else if (activeTab === "payments") {
                const newPmt: PaymentItem = {
                  id: "pmt-" + (payments.length + 1),
                  number: `PMT-2026-${String(payments.length + 1).padStart(4, "0")}`,
                  date: new Date().toISOString().split("T")[0],
                  party: "مصروفات ضيافة الفرع",
                  account: "CASH-HO-01",
                  amount: 150.0,
                  type: "PETTY_CASH_EXPENSE",
                  status: "DRAFT",
                };
                setPayments([newPmt, ...payments]);
                showToast(locale === "ar" ? "تم إنشاء سند صرف عهدة جديد" : "New draft payment voucher created");
              } else if (activeTab === "transfers") {
                const newTrf: TransferItem = {
                  id: "trf-" + (transfers.length + 1),
                  number: `TRF-2026-${String(transfers.length + 1).padStart(4, "0")}`,
                  date: new Date().toISOString().split("T")[0],
                  source: "CASH-HO-01",
                  destination: "CASH-REG-01",
                  amount: 1000.0,
                  fee: 0.0,
                  status: "IN_TRANSIT",
                };
                setTransfers([newTrf, ...transfers]);
                showToast(locale === "ar" ? "تم تسجيل تحويل مالي جديد قيد النقل" : "New transfer initiated in transit");
              } else {
                showToast(locale === "ar" ? "الصناديق الرئيسية محددة ومربوطة بشجرة الحسابات" : "Cash boxes configured and linked to GL");
              }
            }}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm font-medium text-sm"
          >
            <Plus className="w-4 h-4" />
            {activeTab === "receipts"
              ? (t.treasury as any)?.btn_new_receipt || "سند قبض جديد"
              : activeTab === "payments"
              ? (t.treasury as any)?.btn_new_payment || "سند صرف جديد"
              : activeTab === "transfers"
              ? (t.treasury as any)?.btn_new_transfer || "تحويل نقدية جديد"
              : (t.treasury as any)?.btn_new_account || "صندوق / عهدة جديدة"}
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-emerald-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.treasury as any)?.kpi_total_cash || "إجمالي السيولة النقدية"}
            </span>
            <span className="p-2 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <DollarSign className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {totalCash.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>موزعة عبر {accounts.length} صناديق فرعية</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-blue-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.treasury as any)?.kpi_petty_cash || "أرصدة العهد النثرية"}
            </span>
            <span className="p-2 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <Wallet className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {pettyCash.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 font-medium">
            <Clock className="w-3.5 h-3.5" />
            <span>سقف الصرف المسموح 5,000 ريال</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-purple-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.treasury as any)?.kpi_posted_receipts || "مقبوضات الفترة"}
            </span>
            <span className="p-2 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <ArrowDownLeft className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {totalReceipts.toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-purple-600 dark:text-purple-400 font-medium">
            <FileText className="w-3.5 h-3.5" />
            <span>{receipts.length} سندات قبض مسجلة</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm relative overflow-hidden group hover:border-amber-500/50 transition duration-300">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">
              {(t.treasury as any)?.kpi_pending_transfers || "التحويلات المنفذة"}
            </span>
            <span className="p-2 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <ArrowLeftRight className="w-5 h-5" />
            </span>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {transfers.reduce((acc, curr) => acc + curr.amount, 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
            </span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-amber-600 dark:text-amber-400 font-medium">
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>إيداعات بنكية معتمدة</span>
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
          {(t.treasury as any)?.tab_accounts || "الصناديق والعهد"}
        </button>
        <button
          onClick={() => setActiveTab("receipts")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "receipts"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <ArrowDownLeft className="w-4 h-4" />
          {(t.treasury as any)?.tab_receipts || "سندات القبض"}
        </button>
        <button
          onClick={() => setActiveTab("payments")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "payments"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <ArrowUpRight className="w-4 h-4" />
          {(t.treasury as any)?.tab_payments || "سندات الصرف"}
        </button>
        <button
          onClick={() => setActiveTab("transfers")}
          className={`pb-3 px-4 text-sm font-medium border-b-2 transition flex items-center gap-2 whitespace-nowrap ${
            activeTab === "transfers"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <ArrowLeftRight className="w-4 h-4" />
          {(t.treasury as any)?.tab_transfers || "التحويلات المالية"}
        </button>
      </div>

      {/* Tab Content */}
      {activeTab === "accounts" && (
        <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
                <tr>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_code || "الرمز"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_name || "اسم الصندوق / العهدة"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_type || "النوع"}</th>
                  <th className="py-3.5 px-4 text-start">الفرع المسؤول</th>
                  <th className="py-3.5 px-4 text-end">{(t.treasury as any)?.col_limit || "الحد الأقصى"}</th>
                  <th className="py-3.5 px-4 text-end">{(t.treasury as any)?.col_balance || "الرصيد الحالي"}</th>
                  <th className="py-3.5 px-4 text-center">{(t.treasury as any)?.col_status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {accounts.map((acc) => (
                  <tr key={acc.id} className="hover:bg-muted/20 transition">
                    <td className="py-3.5 px-4 font-mono font-semibold text-primary">{acc.code}</td>
                    <td className="py-3.5 px-4 font-medium text-foreground">
                      {locale === "ar" ? acc.nameAr : acc.nameEn}
                    </td>
                    <td className="py-3.5 px-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-secondary text-secondary-foreground">
                        {acc.type === "PETTY_CASH" ? "عهدة نثرية" : acc.type === "CASH_REGISTER" ? "نقطة بيع" : "خزينة رئيسية"}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-muted-foreground">{acc.branch}</td>
                    <td className="py-3.5 px-4 text-end font-mono text-muted-foreground">
                      {acc.limit.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-end font-mono font-bold text-foreground">
                      {acc.balance.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                        <CheckCircle2 className="w-3 h-3" />
                        {acc.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === "receipts" && (
        <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
                <tr>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_number || "رقم السند"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_date || "التاريخ"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_party || "مستلم من"}</th>
                  <th className="py-3.5 px-4 text-start">الصندوق المستودع</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_method || "طريقة الدفع"}</th>
                  <th className="py-3.5 px-4 text-end">{(t.treasury as any)?.col_amount || "المبلغ"}</th>
                  <th className="py-3.5 px-4 text-center">{(t.treasury as any)?.col_status || "الحالة"}</th>
                  <th className="py-3.5 px-4 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {receipts.map((rc) => (
                  <tr key={rc.id} className="hover:bg-muted/20 transition">
                    <td className="py-3.5 px-4 font-mono font-semibold text-primary">{rc.number}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{rc.date}</td>
                    <td className="py-3.5 px-4 font-medium text-foreground">{rc.party}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{rc.account}</td>
                    <td className="py-3.5 px-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
                        {rc.method}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-end font-mono font-bold text-emerald-600 dark:text-emerald-400">
                      +{rc.amount.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        rc.status === "POSTED"
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                          : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                      }`}>
                        {rc.status === "POSTED" ? "مرحل" : "مسودة"}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      {rc.status === "DRAFT" ? (
                        <button
                          onClick={() => handlePostReceipt(rc.id)}
                          className="px-3 py-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-medium transition"
                        >
                          {(t.treasury as any)?.btn_post || "ترحيل"}
                        </button>
                      ) : (
                        <span className="text-xs text-muted-foreground">مكتمل</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === "payments" && (
        <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
                <tr>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_number || "رقم السند"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_date || "التاريخ"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_party || "المدفوع له"}</th>
                  <th className="py-3.5 px-4 text-start">نوع الصرف</th>
                  <th className="py-3.5 px-4 text-start">صندوق الصرف</th>
                  <th className="py-3.5 px-4 text-end">{(t.treasury as any)?.col_amount || "المبلغ"}</th>
                  <th className="py-3.5 px-4 text-center">{(t.treasury as any)?.col_status || "الحالة"}</th>
                  <th className="py-3.5 px-4 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {payments.map((pmt) => (
                  <tr key={pmt.id} className="hover:bg-muted/20 transition">
                    <td className="py-3.5 px-4 font-mono font-semibold text-primary">{pmt.number}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{pmt.date}</td>
                    <td className="py-3.5 px-4 font-medium text-foreground">{pmt.party}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{pmt.type}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{pmt.account}</td>
                    <td className="py-3.5 px-4 text-end font-mono font-bold text-rose-600 dark:text-rose-400">
                      -{pmt.amount.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        pmt.status === "PAID"
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                          : pmt.status === "APPROVED"
                          ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                          : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                      }`}>
                        {pmt.status === "PAID" ? "تم الصرف" : pmt.status === "APPROVED" ? "معتمد" : "مسودة"}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      {pmt.status !== "PAID" ? (
                        <button
                          onClick={() => handlePayPayment(pmt.id)}
                          className="px-3 py-1 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs font-medium transition"
                        >
                          {(t.treasury as any)?.btn_pay || "صرف"}
                        </button>
                      ) : (
                        <span className="text-xs text-muted-foreground">مكتمل</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === "transfers" && (
        <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
                <tr>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_number || "رقم التحويل"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_date || "التاريخ"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_source || "المصدر"}</th>
                  <th className="py-3.5 px-4 text-start">{(t.treasury as any)?.col_destination || "الوجهة"}</th>
                  <th className="py-3.5 px-4 text-end">{(t.treasury as any)?.col_amount || "المبلغ"}</th>
                  <th className="py-3.5 px-4 text-end">رسوم التحويل</th>
                  <th className="py-3.5 px-4 text-center">{(t.treasury as any)?.col_status || "الحالة"}</th>
                  <th className="py-3.5 px-4 text-center">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {transfers.map((trf) => (
                  <tr key={trf.id} className="hover:bg-muted/20 transition">
                    <td className="py-3.5 px-4 font-mono font-semibold text-primary">{trf.number}</td>
                    <td className="py-3.5 px-4 text-muted-foreground">{trf.date}</td>
                    <td className="py-3.5 px-4 font-medium text-foreground">{trf.source}</td>
                    <td className="py-3.5 px-4 font-medium text-foreground">{trf.destination}</td>
                    <td className="py-3.5 px-4 text-end font-mono font-bold text-foreground">
                      {trf.amount.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-end font-mono text-muted-foreground">
                      {trf.fee.toLocaleString()} SAR
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        trf.status === "COMPLETED"
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                          : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                      }`}>
                        {trf.status === "COMPLETED" ? "مكتمل" : "قيد النقل"}
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-center">
                      {trf.status !== "COMPLETED" ? (
                        <button
                          onClick={() => handleCompleteTransfer(trf.id)}
                          className="px-3 py-1 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg text-xs font-medium transition"
                        >
                          {(t.treasury as any)?.btn_complete || "إتمام"}
                        </button>
                      ) : (
                        <span className="text-xs text-muted-foreground">مكتمل</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

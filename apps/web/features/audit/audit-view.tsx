"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  History,
  ShieldCheck,
  CheckCircle2,
  AlertTriangle,
  Lock,
  Search,
  Filter,
  Eye,
  Hash,
  Link as LinkIcon,
  Server,
  Activity,
  UserCheck,
  RefreshCw,
  X,
  FileCode,
} from "lucide-react";

interface AuditLogItem {
  id: string;
  action: string;
  resourceType: string;
  resourceId: string;
  legalEntity: string;
  user: string;
  ip: string;
  timestamp: string;
  previousHash: string;
  recordHash: string;
  payload: Record<string, any>;
  diff?: {
    before?: Record<string, any>;
    after?: Record<string, any>;
  };
}

const mockEvents: AuditLogItem[] = [
  {
    id: "0191f630-9b34-7a81-9451-b06f1d7a0005",
    action: "POST",
    resourceType: "JOURNAL_ENTRY",
    resourceId: "JE-2026-000042",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    user: "admin (مدير النظام)",
    ip: "192.168.1.105",
    timestamp: "2026-09-15 07:15:22",
    previousHash: "c8e71b29a1b023f4d6e90218ab2c3d4e5f60718293a4b5c6d7e8f901a2b3c4d5",
    recordHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    payload: {
      journal_id: "0191f630-9b34-7a81-9451-b06f1d7a0005",
      entry_number: "JE-2026-000042",
      total_debit: "45000.00",
      total_credit: "45000.00",
      lines_count: 4,
      period: "2026-09",
    },
    diff: {
      before: { status: "DRAFT", posted_at: null },
      after: { status: "POSTED", posted_at: "2026-09-15T07:15:22Z" },
    },
  },
  {
    id: "0191f62e-1c88-721a-810e-a43b2f9c8112",
    action: "CREATE",
    resourceType: "SALES_INVOICE",
    resourceId: "INV-RYD-2026-000089",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    user: "sales.officer (مسؤول المبيعات)",
    ip: "192.168.1.112",
    timestamp: "2026-09-15 07:04:10",
    previousHash: "8f4b23c90d1a4e5b6c7d8e9f0123456789abcdef0123456789abcdef01234567",
    recordHash: "c8e71b29a1b023f4d6e90218ab2c3d4e5f60718293a4b5c6d7e8f901a2b3c4d5",
    payload: {
      invoice_number: "INV-RYD-2026-000089",
      customer: "شركة النهضة للتجارة",
      total_excl_vat: "15000.00",
      vat_amount: "2250.00",
      total_incl_vat: "17250.00",
      zatca_stage: "PHASE_2_COMPLIANT",
    },
  },
  {
    id: "0191f629-4501-789f-b883-99cf3e2d1944",
    action: "ASSIGN_ROLE",
    resourceType: "USER_ROLE_ASSIGNMENT",
    resourceId: "CHIEF_ACCOUNTANT",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    user: "admin (مدير النظام)",
    ip: "192.168.1.101",
    timestamp: "2026-09-15 06:45:00",
    previousHash: "4a5b6c7d8e9f0123456789abcdef0123456789abcdef0123456789abcdef0123",
    recordHash: "8f4b23c90d1a4e5b6c7d8e9f0123456789abcdef0123456789abcdef01234567",
    payload: {
      role: "CHIEF_ACCOUNTANT",
      assigned_to_user_id: "0191f620-1122-3344-5566-778899aabbcc",
      assigned_by: "admin",
      permissions_count: 14,
      scope: "LEGAL_ENTITY",
    },
  },
  {
    id: "0191f620-1122-3344-5566-778899aabbcc",
    action: "CREATE",
    resourceType: "NUMBER_SEQUENCE",
    resourceId: "SEQ-INV-RYD-2026",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    user: "admin (مدير النظام)",
    ip: "127.0.0.1",
    timestamp: "2026-09-15 06:10:00",
    previousHash: "GENESIS",
    recordHash: "4a5b6c7d8e9f0123456789abcdef0123456789abcdef0123456789abcdef0123",
    payload: {
      code: "INV_SALES_RYD",
      prefix: "INV-RYD-2026-",
      current_val: 89,
      concurrency_protection: "PESSIMISTIC_LOCK_FOR_UPDATE",
    },
  },
];

export function AuditView() {
  const { t, locale, dir } = useLocale();

  const [activeTab, setActiveTab] = useState<"events" | "security" | "integrity">("events");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedAction, setSelectedAction] = useState<string>("ALL");
  const [selectedEvent, setSelectedEvent] = useState<AuditLogItem | null>(null);
  const [verifying, setVerifying] = useState(false);
  const [verificationDone, setVerificationDone] = useState(true);

  const handleVerifyChain = () => {
    setVerifying(true);
    setTimeout(() => {
      setVerifying(false);
      setVerificationDone(true);
    }, 600);
  };

  const filteredEvents = mockEvents.filter((ev) => {
    const matchAction = selectedAction === "ALL" || ev.action === selectedAction;
    const matchSearch =
      searchQuery === "" ||
      ev.resourceId.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ev.action.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ev.user.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ev.resourceType.toLowerCase().includes(searchQuery.toLowerCase());
    return matchAction && matchSearch;
  });

  const getActionBadgeColor = (action: string) => {
    switch (action) {
      case "POST":
        return "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20";
      case "CREATE":
        return "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20";
      case "ASSIGN_ROLE":
        return "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20";
      case "UPDATE":
        return "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20";
      case "DELETE":
        return "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20";
      default:
        return "bg-muted text-muted-foreground border-border";
    }
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 mb-2">
            <ShieldCheck className="h-3.5 w-3.5" />
            <span>{t.app.phase}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.audit.title}
          </h1>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {t.audit.subtitle}
          </p>
        </div>

        <div className="flex items-center gap-3 p-3 rounded-xl border bg-card/60 shadow-sm">
          <div className="h-10 w-10 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-600 dark:text-emerald-400">
            <Lock className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-muted-foreground font-medium">مناعة السجلات (Strict Invariant)</div>
            <div className="text-sm font-bold text-foreground flex items-center gap-1.5">
              <span>Append-Only Immutable</span>
              <span className="h-2 w-2 rounded-full bg-emerald-500 inline-block animate-pulse" />
            </div>
          </div>
        </div>
      </div>

      {/* Cryptographic Chain Integrity Card */}
      <div className="p-5 rounded-2xl border bg-gradient-to-r from-card via-card to-emerald-500/5 shadow-sm relative overflow-hidden">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div className="space-y-2 max-w-2xl">
            <div className="flex items-center gap-2">
              <span className="p-1 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                <CheckCircle2 className="h-5 w-5" />
              </span>
              <h2 className="text-base font-bold text-foreground">
                {t.audit.integrity_title}
              </h2>
              <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-semibold border border-emerald-500/20">
                {verificationDone ? "100% SHA-256 Valid" : "Checking..."}
              </span>
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed">
              {verificationDone ? t.audit.integrity_valid : t.audit.integrity_broken}
            </p>

            {/* Quick Metrics */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
              <div className="p-2.5 rounded-lg border bg-background/50">
                <div className="text-[11px] text-muted-foreground">{t.audit.total_records}</div>
                <div className="text-lg font-bold text-foreground">4,892</div>
              </div>
              <div className="p-2.5 rounded-lg border bg-background/50">
                <div className="text-[11px] text-muted-foreground">{t.audit.verified_records}</div>
                <div className="text-lg font-bold text-emerald-600 dark:text-emerald-400">4,892 (100%)</div>
              </div>
              <div className="p-2.5 rounded-lg border bg-background/50">
                <div className="text-[11px] text-muted-foreground">خوارزمية الربط</div>
                <div className="text-lg font-bold text-foreground font-mono">SHA-256</div>
              </div>
              <div className="p-2.5 rounded-lg border bg-background/50">
                <div className="text-[11px] text-muted-foreground">حماية قاعدة البيانات</div>
                <div className="text-lg font-bold text-foreground">Triggers Blocked</div>
              </div>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row lg:flex-col items-start lg:items-end gap-3 justify-center">
            <button
              onClick={handleVerifyChain}
              disabled={verifying}
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-primary text-primary-foreground font-medium text-sm shadow-sm hover:opacity-90 transition-opacity disabled:opacity-50"
            >
              <RefreshCw className={`h-4 w-4 ${verifying ? "animate-spin" : ""}`} />
              <span>{verifying ? t.audit.verifying : t.audit.verify_button}</span>
            </button>
            <span className="text-[11px] text-muted-foreground">
              {t.audit.last_check}: منذ بضع ثوانٍ
            </span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b">
        <button
          onClick={() => setActiveTab("events")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "events"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <History className="h-4 w-4" />
          <span>{t.audit.events_table_title}</span>
        </button>

        <button
          onClick={() => setActiveTab("security")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "security"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Activity className="h-4 w-4" />
          <span>{t.audit.security_events_title}</span>
        </button>

        <button
          onClick={() => setActiveTab("integrity")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "integrity"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Hash className="h-4 w-4" />
          <span>محاكي سلسلة التجزئة التشفيرية</span>
        </button>
      </div>

      {/* Tab 1: Audit Events Table */}
      {activeTab === "events" && (
        <div className="space-y-4">
          {/* Filters & Search */}
          <div className="flex flex-col sm:flex-row gap-3 items-center justify-between">
            <div className="relative w-full sm:w-80">
              <Search className="absolute inset-inline-start-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t.audit.search_placeholder}
                className="w-full ps-9 pe-4 py-2 rounded-xl border bg-card text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/20"
              />
            </div>

            <div className="flex items-center gap-2 w-full sm:w-auto">
              <div className="flex items-center gap-1.5 p-1 rounded-lg border bg-card text-xs">
                {["ALL", "POST", "CREATE", "ASSIGN_ROLE"].map((act) => (
                  <button
                    key={act}
                    onClick={() => setSelectedAction(act)}
                    className={`px-3 py-1.5 rounded-md font-medium transition-colors ${
                      selectedAction === act
                        ? "bg-primary text-primary-foreground shadow-xs"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {act === "ALL" ? t.audit.filter_all : act}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Table */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full text-start text-xs">
                <thead className="bg-muted/40 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-start">{t.audit.action}</th>
                    <th className="px-4 py-3 text-start">{t.audit.resource_type}</th>
                    <th className="px-4 py-3 text-start">{t.audit.resource_id}</th>
                    <th className="px-4 py-3 text-start">{t.audit.user}</th>
                    <th className="px-4 py-3 text-start">{t.audit.ip_address}</th>
                    <th className="px-4 py-3 text-start">{t.audit.timestamp}</th>
                    <th className="px-4 py-3 text-start">{t.audit.record_hash}</th>
                    <th className="px-4 py-3 text-center">{t.audit.view_payload}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60 font-medium">
                  {filteredEvents.map((ev) => (
                    <tr key={ev.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-bold border ${getActionBadgeColor(
                            ev.action
                          )}`}
                        >
                          {ev.action}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-semibold text-foreground">
                        {ev.resourceType}
                      </td>
                      <td className="px-4 py-3 font-mono text-primary font-bold">
                        {ev.resourceId}
                      </td>
                      <td className="px-4 py-3 text-foreground">{ev.user}</td>
                      <td className="px-4 py-3 text-muted-foreground font-mono">{ev.ip}</td>
                      <td className="px-4 py-3 text-muted-foreground font-mono">{ev.timestamp}</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1 font-mono text-[10px] text-muted-foreground">
                          <LinkIcon className="h-3 w-3 text-emerald-500 shrink-0" />
                          <span className="truncate max-w-[120px]">{ev.recordHash}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-center">
                        <button
                          onClick={() => setSelectedEvent(ev)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg border bg-background hover:bg-muted text-xs font-semibold text-foreground transition-colors"
                        >
                          <Eye className="h-3.5 w-3.5 text-primary" />
                          <span>{t.audit.view_payload}</span>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Security Events */}
      {activeTab === "security" && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="p-4 rounded-xl border bg-card space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-muted-foreground">أحداث تسجيل الدخول</span>
              <UserCheck className="h-4 w-4 text-emerald-500" />
            </div>
            <div className="text-2xl font-bold text-foreground">1,240</div>
            <div className="text-xs text-muted-foreground">كافة المحاولات موثقة ومحمية بمعدل خوارزمية Argon2id</div>
          </div>

          <div className="p-4 rounded-xl border bg-card space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-muted-foreground">محاولات غير مصرح بها (403/401)</span>
              <AlertTriangle className="h-4 w-4 text-amber-500" />
            </div>
            <div className="text-2xl font-bold text-foreground">3</div>
            <div className="text-xs text-muted-foreground">تم حظر الوصول وتسجيل بيانات الـ IP تلقائياً</div>
          </div>

          <div className="p-4 rounded-xl border bg-card space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-muted-foreground">متوسط زمن الاستجابة</span>
              <Server className="h-4 w-4 text-primary" />
            </div>
            <div className="text-2xl font-bold text-foreground font-mono">18ms</div>
            <div className="text-xs text-muted-foreground">مسجلة عبر كاش السياسات اللحظي في Go HTTP</div>
          </div>
        </div>
      )}

      {/* Tab 3: Hash Chain Simulator & Integrity Proof */}
      {activeTab === "integrity" && (
        <div className="p-6 rounded-2xl border bg-card space-y-6">
          <div>
            <h3 className="text-lg font-bold text-foreground">
              سلسلة التجزئة المشفرة (Cryptographic Hash Chaining Mechanism)
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              يرتبط كل سجل بالسجل السابق بواسطة تجزئة SHA-256 مشفرة، مما يجعل تعديل أي سجل تاريخي مستحيلاً رياضياً دون كسر كافة السجلات اللاحقة.
            </p>
          </div>

          {/* Visual Chain Diagram */}
          <div className="space-y-4">
            {mockEvents.slice(0, 3).map((item, idx) => (
              <div key={item.id} className="p-4 rounded-xl border bg-background space-y-2 relative">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold font-mono text-primary">
                    كتلة التدقيق #{mockEvents.length - idx}: {item.action} {item.resourceId}
                  </span>
                  <span className="text-[11px] font-mono text-muted-foreground">{item.timestamp}</span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs font-mono pt-1">
                  <div className="p-2 rounded bg-muted/40 border">
                    <span className="text-muted-foreground block text-[10px]">Previous Hash:</span>
                    <span className="truncate block text-foreground font-semibold">{item.previousHash}</span>
                  </div>
                  <div className="p-2 rounded bg-emerald-500/10 border border-emerald-500/20">
                    <span className="text-emerald-700 dark:text-emerald-400 block text-[10px]">Current Record Hash:</span>
                    <span className="truncate block text-emerald-700 dark:text-emerald-300 font-semibold">{item.recordHash}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Payload Details Modal */}
      {selectedEvent && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-2xl rounded-2xl border bg-card shadow-xl p-6 space-y-5 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-4">
              <div className="flex items-center gap-2">
                <span className={`px-2.5 py-0.5 rounded-md text-xs font-bold border ${getActionBadgeColor(selectedEvent.action)}`}>
                  {selectedEvent.action}
                </span>
                <span className="font-bold text-foreground">{selectedEvent.resourceId}</span>
              </div>
              <button
                onClick={() => setSelectedEvent(null)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-3 p-3 rounded-xl bg-muted/30 border">
                <div>
                  <span className="text-muted-foreground block text-[10px]">المستخدم</span>
                  <span className="font-semibold text-foreground">{selectedEvent.user}</span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-[10px]">الكيان القانوني</span>
                  <span className="font-semibold text-foreground">{selectedEvent.legalEntity}</span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-[10px]">عنوان IP</span>
                  <span className="font-mono text-foreground">{selectedEvent.ip}</span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-[10px]">التاريخ والتوقيت</span>
                  <span className="font-mono text-foreground">{selectedEvent.timestamp}</span>
                </div>
              </div>

              {/* Hashes */}
              <div className="space-y-2 p-3 rounded-xl bg-muted/30 border font-mono">
                <div>
                  <span className="text-muted-foreground block text-[10px]">Previous Hash:</span>
                  <span className="text-foreground text-[11px] break-all">{selectedEvent.previousHash}</span>
                </div>
                <div className="pt-1">
                  <span className="text-emerald-600 dark:text-emerald-400 block text-[10px]">Current Record Hash:</span>
                  <span className="text-emerald-600 dark:text-emerald-400 text-[11px] break-all font-bold">
                    {selectedEvent.recordHash}
                  </span>
                </div>
              </div>

              {/* Payload JSON */}
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-muted-foreground font-semibold">
                  <FileCode className="h-4 w-4 text-primary" />
                  <span>بيانات الحمولة الأصلية (Payload JSON)</span>
                </div>
                <pre className="p-3 rounded-xl bg-muted/60 border text-[11px] font-mono overflow-x-auto text-foreground max-h-48">
                  {JSON.stringify(selectedEvent.payload, null, 2)}
                </pre>
              </div>

              {/* State Diff if available */}
              {selectedEvent.diff && (
                <div className="space-y-2 pt-2">
                  <div className="font-semibold text-foreground">تغيرات الحالة قبل وبعد العملية (Entity State Diff):</div>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="p-2.5 rounded-lg border bg-amber-500/5 border-amber-500/20">
                      <span className="text-[10px] text-amber-600 dark:text-amber-400 font-bold block mb-1">
                        الحالة قبل التعديل (Before)
                      </span>
                      <pre className="text-[10px] font-mono text-foreground overflow-x-auto">
                        {JSON.stringify(selectedEvent.diff.before, null, 2)}
                      </pre>
                    </div>
                    <div className="p-2.5 rounded-lg border bg-emerald-500/5 border-emerald-500/20">
                      <span className="text-[10px] text-emerald-600 dark:text-emerald-400 font-bold block mb-1">
                        الحالة بعد التعديل (After)
                      </span>
                      <pre className="text-[10px] font-mono text-foreground overflow-x-auto">
                        {JSON.stringify(selectedEvent.diff.after, null, 2)}
                      </pre>
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="flex justify-end pt-2">
              <button
                onClick={() => setSelectedEvent(null)}
                className="px-4 py-2 rounded-xl bg-primary text-primary-foreground font-medium text-xs shadow-sm hover:opacity-90"
              >
                إغلاق النافذة
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

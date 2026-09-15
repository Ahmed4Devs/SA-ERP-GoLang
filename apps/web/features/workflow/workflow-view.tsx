"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  GitMerge,
  ShieldCheck,
  CheckCircle2,
  XCircle,
  Clock,
  UserCheck,
  UserX,
  FileText,
  Building,
  ArrowRight,
  ArrowLeft,
  Filter,
  Check,
  X,
  Layers,
  Sparkles,
  Calendar,
  AlertOctagon,
  Users,
} from "lucide-react";

interface ApprovalItem {
  id: string;
  documentType: string;
  documentId: string;
  amount: string;
  legalEntity: string;
  initiatorName: string;
  stepNameAr: string;
  stepNameEn: string;
  createdAt: string;
  status: "PENDING" | "APPROVED" | "REJECTED";
}

const initialInbox: ApprovalItem[] = [
  {
    id: "req-001",
    documentType: "SALES_INVOICE",
    documentId: "INV-RYD-2026-000089",
    amount: "78,500.00 SAR",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    initiatorName: "sales.officer (مسؤول المبيعات)",
    stepNameAr: "اعتماد الإدارة المالية للمبالغ الكبرى (> 50,000 ر.س)",
    stepNameEn: "Finance High-Value Approval (> 50,000 SAR)",
    createdAt: "منذ 15 دقيقة",
    status: "PENDING",
  },
  {
    id: "req-002",
    documentType: "JOURNAL_ENTRY",
    documentId: "JE-2026-000042",
    amount: "145,000.00 SAR",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    initiatorName: "senior.accountant (محاسب أول)",
    stepNameAr: "الاعتماد المالي النهائي للمدير المالي (CFO)",
    stepNameEn: "CFO Final Approval",
    createdAt: "منذ ساعتين",
    status: "PENDING",
  },
  {
    id: "req-003",
    documentType: "PURCHASE_ORDER",
    documentId: "PO-RYD-2026-000014",
    amount: "32,000.00 SAR",
    legalEntity: "شركة الأفق المتقدمة (LE-01)",
    initiatorName: "purchasing.lead (مسؤول المشتريات)",
    stepNameAr: "مراجعة واعتماد المشتريات المبدئية",
    stepNameEn: "Purchasing Lead Review",
    createdAt: "منذ 4 ساعات",
    status: "PENDING",
  },
];

export function WorkflowView() {
  const { t, locale, dir } = useLocale();

  const [activeTab, setActiveTab] = useState<"inbox" | "definitions" | "simulator" | "delegations">("inbox");
  const [inboxItems, setInboxItems] = useState<ApprovalItem[]>(initialInbox);
  const [selectedRole, setSelectedRole] = useState<"MAKER" | "CHECKER">("MAKER");
  const [simResult, setSimResult] = useState<{ allowed: boolean; message: string } | null>(null);
  const [processingId, setProcessingId] = useState<string | null>(null);

  const isRTL = dir === "rtl";
  const Arrow = isRTL ? ArrowLeft : ArrowRight;

  const handleDecision = (id: string, decision: "APPROVED" | "REJECTED") => {
    setProcessingId(id);
    setTimeout(() => {
      setInboxItems((prev) =>
        prev.map((item) => (item.id === id ? { ...item, status: decision } : item))
      );
      setProcessingId(null);
    }, 400);
  };

  const handleSimulate = (action: "APPROVE") => {
    if (selectedRole === "MAKER") {
      setSimResult({
        allowed: false,
        message:
          locale === "ar"
            ? "حظر فوري (403 Forbidden): قاعدة فصل المهام (Maker/Checker) تمنع منشئ المعاملة من اعتماد طلبه الخاص (cannot_approve_own_document)"
            : "Blocked (403 Forbidden): Segregation of Duties policy strictly prevents the document initiator from approving their own request (cannot_approve_own_document)",
      });
    } else {
      setSimResult({
        allowed: true,
        message:
          locale === "ar"
            ? "مسموح ومصرح به: تم التحقق من هوية المعتمد واستيفاء الرقابة الثنائية بنجاح وانتقال المستند للخطوة التالية"
            : "Authorized: Maker/Checker verified successfully, dual-control satisfied, and workflow transitioned to next stage",
      });
    }
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/10 text-blue-600 dark:text-blue-400 mb-2">
            <GitMerge className="h-3.5 w-3.5" />
            <span>{t.app.phase}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.workflow.title}
          </h1>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {t.workflow.subtitle}
          </p>
        </div>

        <div className="flex items-center gap-3 p-3 rounded-xl border bg-card/60 shadow-sm">
          <div className="h-10 w-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-600 dark:text-blue-400">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-muted-foreground font-medium">الرقابة الثنائية (Maker/Checker)</div>
            <div className="text-sm font-bold text-foreground flex items-center gap-1.5">
              <span>Segregation of Duties Enforced</span>
              <span className="h-2 w-2 rounded-full bg-blue-500 inline-block animate-pulse" />
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b">
        <button
          onClick={() => setActiveTab("inbox")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "inbox"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Clock className="h-4 w-4" />
          <span>{t.workflow.inbox_title}</span>
          <span className="px-1.5 py-0.2 rounded-full bg-primary/10 text-primary text-xs font-bold">
            {inboxItems.filter((i) => i.status === "PENDING").length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("definitions")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "definitions"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>{t.workflow.definitions_title}</span>
        </button>

        <button
          onClick={() => setActiveTab("simulator")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "simulator"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <AlertOctagon className="h-4 w-4" />
          <span>{t.workflow.simulator_title}</span>
        </button>

        <button
          onClick={() => setActiveTab("delegations")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-colors ${
            activeTab === "delegations"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Users className="h-4 w-4" />
          <span>{t.workflow.delegations_title}</span>
        </button>
      </div>

      {/* Tab 1: Approval Inbox */}
      {activeTab === "inbox" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 gap-4">
            {inboxItems.map((item) => {
              const isPending = item.status === "PENDING";
              const isApproved = item.status === "APPROVED";
              const isRejected = item.status === "REJECTED";

              return (
                <div
                  key={item.id}
                  className={`p-5 rounded-2xl border transition-all ${
                    isPending
                      ? "bg-card hover:border-primary/40 shadow-xs"
                      : isApproved
                      ? "bg-emerald-500/5 border-emerald-500/30"
                      : "bg-red-500/5 border-red-500/30 opacity-75"
                  }`}
                >
                  <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                    <div className="space-y-2">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="px-2 py-0.5 rounded-md text-xs font-bold bg-muted text-foreground border">
                          {item.documentType}
                        </span>
                        <span className="font-mono font-bold text-primary text-sm">
                          {item.documentId}
                        </span>
                        <span className="text-xs text-muted-foreground">({item.legalEntity})</span>
                        <span className="text-xs text-muted-foreground">• {item.createdAt}</span>
                      </div>

                      <div className="text-sm font-semibold text-foreground">
                        {locale === "ar" ? item.stepNameAr : item.stepNameEn}
                      </div>

                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <div>
                          <span>{t.workflow.initiator}: </span>
                          <span className="font-medium text-foreground">{item.initiatorName}</span>
                        </div>
                        <div>
                          <span>{t.workflow.amount}: </span>
                          <span className="font-bold text-foreground font-mono">{item.amount}</span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      {isPending ? (
                        <>
                          <button
                            onClick={() => handleDecision(item.id, "APPROVED")}
                            disabled={processingId === item.id}
                            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold shadow-xs transition-colors disabled:opacity-50"
                          >
                            <Check className="h-4 w-4" />
                            <span>{t.workflow.approve}</span>
                          </button>
                          <button
                            onClick={() => handleDecision(item.id, "REJECTED")}
                            disabled={processingId === item.id}
                            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-red-600 hover:bg-red-700 text-white text-xs font-semibold shadow-xs transition-colors disabled:opacity-50"
                          >
                            <X className="h-4 w-4" />
                            <span>{t.workflow.reject}</span>
                          </button>
                        </>
                      ) : isApproved ? (
                        <div className="flex items-center gap-1.5 text-xs font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-lg border border-emerald-500/20">
                          <CheckCircle2 className="h-4 w-4" />
                          <span>تم الاعتماد بنجاح (Approved)</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5 text-xs font-bold text-red-600 dark:text-red-400 bg-red-500/10 px-3 py-1.5 rounded-lg border border-red-500/20">
                          <XCircle className="h-4 w-4" />
                          <span>تم الرفض (Rejected)</span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Tab 2: Workflow Definitions */}
      {activeTab === "definitions" && (
        <div className="space-y-6">
          {/* Definition 1: Sales Invoice */}
          <div className="p-6 rounded-2xl border bg-card space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div>
                <span className="text-xs font-mono text-primary font-bold">SALES_INVOICE_APPROVAL</span>
                <h3 className="text-base font-bold text-foreground">اعتماد فواتير المبيعات (Sales Invoice Approval)</h3>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                v1 (PUBLISHED)
              </span>
            </div>

            <div className="flex flex-col md:flex-row items-center gap-4 pt-2">
              <div className="p-4 rounded-xl border bg-background flex-1 w-full space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-primary font-mono">الخطوة 1: مراجعة المبيعات</span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-muted">إلزامي للجميع</span>
                </div>
                <div className="text-xs text-foreground font-medium">مراجعة مسؤول المبيعات المبدئية للأسعار والشروط</div>
                <div className="text-[11px] text-muted-foreground">المعتمد: دور مسؤول المبيعات (SALES_OFFICER)</div>
              </div>

              <div className="p-2 rounded-full bg-muted text-muted-foreground shrink-0">
                <Arrow className="h-4 w-4" />
              </div>

              <div className="p-4 rounded-xl border bg-background flex-1 w-full space-y-2 border-amber-500/30">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-amber-600 dark:text-amber-400 font-mono">الخطوة 2: اعتماد الإدارة المالية</span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-amber-500/10 text-amber-600 font-bold border border-amber-500/20">
                    شرطي: المبلغ &gt; 50,000 ر.س
                  </span>
                </div>
                <div className="text-xs text-foreground font-medium">اعتماد رئيس الحسابات أو المدير المالي للفواتير الكبرى</div>
                <div className="text-[11px] text-muted-foreground">المعتمد: دور رئيس الحسابات (CHIEF_ACCOUNTANT)</div>
              </div>
            </div>
          </div>

          {/* Definition 2: Journal Entry */}
          <div className="p-6 rounded-2xl border bg-card space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div>
                <span className="text-xs font-mono text-primary font-bold">JOURNAL_ENTRY_APPROVAL</span>
                <h3 className="text-base font-bold text-foreground">اعتماد وترحيل القيود اليومية (Journal Entry Posting)</h3>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                v1 (PUBLISHED)
              </span>
            </div>

            <div className="flex flex-col md:flex-row items-center gap-4 pt-2">
              <div className="p-4 rounded-xl border bg-background flex-1 w-full space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-primary font-mono">الخطوة 1: تدقيق محاسب أول</span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-muted">إلزامي للجميع</span>
                </div>
                <div className="text-xs text-foreground font-medium">مراجعة توازن المدين والدائن ومراكز التكلفة والحسابات</div>
                <div className="text-[11px] text-muted-foreground">المعتمد: دور رئيس الحسابات (CHIEF_ACCOUNTANT)</div>
              </div>

              <div className="p-2 rounded-full bg-muted text-muted-foreground shrink-0">
                <Arrow className="h-4 w-4" />
              </div>

              <div className="p-4 rounded-xl border bg-background flex-1 w-full space-y-2 border-purple-500/30">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-purple-600 dark:text-purple-400 font-mono">الخطوة 2: الاعتماد المالي النهائي (CFO)</span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-purple-500/10 text-purple-600 font-bold border border-purple-500/20">
                    شرطي: المبلغ &gt; 100,000 ر.س
                  </span>
                </div>
                <div className="text-xs text-foreground font-medium">اعتماد مدير النظام أو المدير المالي للقيود الجوهرية</div>
                <div className="text-[11px] text-muted-foreground">المعتمد: دور مدير النظام الشامل (SUPER_ADMIN)</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: Maker/Checker Simulator */}
      {activeTab === "simulator" && (
        <div className="p-6 rounded-2xl border bg-card space-y-6">
          <div>
            <h3 className="text-lg font-bold text-foreground">
              {t.workflow.simulator_title}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {t.workflow.maker_checker_rule} (Section 55: Segregation of Duties).
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4 p-5 rounded-xl border bg-background/50">
              <span className="text-xs font-bold text-foreground block">اختيار صفة المستخدم الحالي في النظام:</span>
              <div className="space-y-2">
                <label
                  onClick={() => setSelectedRole("MAKER")}
                  className={`flex items-start gap-3 p-3.5 rounded-xl border cursor-pointer transition-all ${
                    selectedRole === "MAKER"
                      ? "border-primary bg-primary/5 ring-2 ring-primary/20"
                      : "border-border hover:bg-muted/40"
                  }`}
                >
                  <input
                    type="radio"
                    checked={selectedRole === "MAKER"}
                    onChange={() => setSelectedRole("MAKER")}
                    className="mt-1"
                  />
                  <div>
                    <div className="text-xs font-bold text-foreground">أنا منشئ الوثيقة (Document Initiator / Maker)</div>
                    <div className="text-[11px] text-muted-foreground mt-0.5">
                      قمت بإنشاء الفاتورة INV-2026-000089 وأحاول اعتمادها بنفسي
                    </div>
                  </div>
                </label>

                <label
                  onClick={() => setSelectedRole("CHECKER")}
                  className={`flex items-start gap-3 p-3.5 rounded-xl border cursor-pointer transition-all ${
                    selectedRole === "CHECKER"
                      ? "border-primary bg-primary/5 ring-2 ring-primary/20"
                      : "border-border hover:bg-muted/40"
                  }`}
                >
                  <input
                    type="radio"
                    checked={selectedRole === "CHECKER"}
                    onChange={() => setSelectedRole("CHECKER")}
                    className="mt-1"
                  />
                  <div>
                    <div className="text-xs font-bold text-foreground">أنا المعتمد المستقل (Independent Checker / Approver)</div>
                    <div className="text-[11px] text-muted-foreground mt-0.5">
                      مسؤول مالي مستقل يحمل الصلاحية ولم يشارك في إنشاء المستند
                    </div>
                  </div>
                </label>
              </div>

              <button
                onClick={() => handleSimulate("APPROVE")}
                className="w-full py-2.5 rounded-xl bg-primary text-primary-foreground text-xs font-bold shadow-sm hover:opacity-90 transition-opacity"
              >
                اختبار تنفيذ إجراء الاعتماد (Test Approval Action)
              </button>
            </div>

            {/* Live Result Output */}
            <div className="p-5 rounded-xl border bg-background/50 flex flex-col justify-between space-y-4">
              <span className="text-xs font-bold text-foreground block">استجابة محرك الرقابة وفصل المهام:</span>

              {simResult ? (
                <div
                  className={`p-4 rounded-xl border space-y-2 animate-in fade-in zoom-in-95 ${
                    simResult.allowed
                      ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-700 dark:text-emerald-300"
                      : "bg-red-500/10 border-red-500/30 text-red-700 dark:text-red-300"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    {simResult.allowed ? (
                      <CheckCircle2 className="h-5 w-5 shrink-0 text-emerald-600 dark:text-emerald-400" />
                    ) : (
                      <AlertOctagon className="h-5 w-5 shrink-0 text-red-600 dark:text-red-400" />
                    )}
                    <span className="font-bold text-xs">
                      {simResult.allowed ? "اعتماد قانوني ناجح" : "انتهاك فصل المهام (Violation Blocked)"}
                    </span>
                  </div>
                  <p className="text-xs leading-relaxed">{simResult.message}</p>
                </div>
              ) : (
                <div className="text-center py-10 text-xs text-muted-foreground">
                  اختر الصفة واضغط على زر الاختبار لمشاهدة استجابة النظام الفورية
                </div>
              )}

              <div className="p-3 rounded-lg border bg-muted/40 text-[11px] text-muted-foreground space-y-1">
                <div className="font-semibold text-foreground">المرجع التنظيمي: Section 55 Segregation of Duties</div>
                <div>Policies: cannot_approve_own_document, cannot_post_own_document</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: Delegations Management */}
      {activeTab === "delegations" && (
        <div className="p-6 rounded-2xl border bg-card space-y-5">
          <div className="flex items-center justify-between border-b pb-3">
            <div>
              <h3 className="text-base font-bold text-foreground">{t.workflow.delegations_title}</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                إدارة الوكالات والتفويضات المؤقتة للاعتمادات المالية والإدارية
              </p>
            </div>
            <button className="px-3.5 py-1.5 rounded-xl bg-primary text-primary-foreground text-xs font-bold shadow-xs">
              + إنشاء تفويض جديد
            </button>
          </div>

          <div className="space-y-3">
            <div className="p-4 rounded-xl border bg-background space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <UserCheck className="h-4 w-4 text-emerald-500" />
                  <span className="font-bold text-xs text-foreground">
                    تفويض من: رئيس الحسابات إلى: المدقق المالي (CFO Proxy)
                  </span>
                </div>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                  {t.workflow.active}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">
                الكيان القانوني: شركة الأفق المتقدمة (LE-01) • السبب: إجازة سنوية
              </div>
              <div className="text-[11px] font-mono text-muted-foreground flex items-center gap-2 pt-1">
                <Calendar className="h-3.5 w-3.5" />
                <span>ساري من: 2026-09-10 حتى: 2026-09-25</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

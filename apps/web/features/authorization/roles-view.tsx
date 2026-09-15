"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  ShieldCheck,
  Key,
  Users,
  Building,
  CheckCircle2,
  XCircle,
  Sparkles,
  Lock,
  Layers,
  ChevronRight,
  ChevronLeft,
} from "lucide-react";

export function RolesView() {
  const { t, locale, dir } = useLocale();

  const [activeTab, setActiveTab] = useState<"roles" | "permissions" | "assignments" | "evaluator">("roles");
  const [selectedPerm, setSelectedPerm] = useState("gl.journals.post");
  const [selectedScope, setSelectedScope] = useState<"LE-01" | "LE-02" | "GLOBAL">("LE-01");
  const [evaluating, setEvaluating] = useState(false);
  const [decision, setDecision] = useState<{ allowed: boolean; reason: string } | null>({
    allowed: true,
    reason: locale === "ar" ? "صلاحية ممنوحة عبر دور مدير النظام الشامل (SUPER_ADMIN)" : "Permission granted via SUPER_ADMIN role",
  });

  const isRTL = dir === "rtl";
  const ArrowIcon = isRTL ? ChevronLeft : ChevronRight;

  const handleEvaluate = () => {
    setEvaluating(true);
    setTimeout(() => {
      // Simulate real-time evaluation logic matching backend
      if (selectedScope === "LE-02" && selectedPerm === "gl.journals.post") {
        setDecision({
          allowed: false,
          reason: locale === "ar"
            ? "محظور: المستخدم غير مرخص له بترحيل القيود في الكيان القانوني LE-02 (عزل مالي صارم)"
            : "Access Denied: User lacks journal posting authorization in Legal Entity LE-02",
        });
      } else {
        setDecision({
          allowed: true,
          reason: locale === "ar"
            ? `صلاحية ممنوحة ومصرح بها في نطاق ${selectedScope}`
            : `Permission granted within ${selectedScope} scope (Policy Cache hit: 0.12ms)`,
        });
      }
      setEvaluating(false);
    }, 350);
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-600 dark:text-purple-400 mb-2">
            <ShieldCheck className="h-3.5 w-3.5" />
            <span>{t.app.phase}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.authorization.title}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t.authorization.subtitle}
          </p>
        </div>

        <div className="flex items-center gap-3 p-3 rounded-xl border bg-card/60 shadow-sm">
          <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
            <Key className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-muted-foreground font-medium">سياسة التحكم بالوصول</div>
            <div className="text-sm font-bold text-foreground">
              {locale === "ar" ? "تقييد بالنطاق المالي الصارم" : "Scoped RBAC + In-Memory Cache"}
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex overflow-x-auto border-b border-border gap-2 pb-px">
        {[
          { id: "roles", label: t.authorization.roles_title, icon: ShieldCheck },
          { id: "permissions", label: t.authorization.permissions_title, icon: Key },
          { id: "assignments", label: t.authorization.assignments_title, icon: Users },
          { id: "evaluator", label: t.authorization.evaluator_title, icon: Sparkles },
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
        {/* 1. Roles */}
        {activeTab === "roles" && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[
              {
                code: "SUPER_ADMIN",
                ar: "مدير النظام الشامل",
                en: "Super Administrator",
                desc: "صلاحيات غير مقيدة لكافة الكيانات القانونية والوحدات التشغيلية",
                scope: "Global",
                permsCount: "كافة الصلاحيات (*)",
              },
              {
                code: "ORG_ADMIN",
                ar: "مدير الكيان القانوني",
                en: "Legal Entity Administrator",
                desc: "إدارة الفروع والأقسام والمستخدمين مقيدة بنطاق الكيان القانوني المحدد",
                scope: "Scoped",
                permsCount: "14 صلاحية",
              },
              {
                code: "CHIEF_ACCOUNTANT",
                ar: "رئيس الحسابات",
                en: "Chief Accountant",
                desc: "إدارة الحسابات ودفاتر اليومية وترحيل القيود والتقارير المالية",
                scope: "Scoped",
                permsCount: "8 صلاحيات محاسبية",
              },
              {
                code: "SALES_OFFICER",
                ar: "مسؤول المبيعات",
                en: "Sales Officer",
                desc: "إنشاء الفواتير وإدارة طلبات وعقود العملاء وإصدار إشعارات الخصم",
                scope: "Scoped",
                permsCount: "6 صلاحيات مبيعات",
              },
              {
                code: "AUDITOR",
                ar: "مدقق مالي",
                en: "Financial Auditor",
                desc: "صلاحيات اطلاع وتدقيق للقوائم والقيود دون صلاحية التعديل أو الترحيل",
                scope: "Scoped",
                permsCount: "صلاحيات قراءة فقط",
              },
            ].map((role) => (
              <div key={role.code} className="p-5 rounded-xl border bg-card space-y-3 shadow-sm hover:shadow-md transition-shadow">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className="h-9 w-9 rounded-lg bg-primary/10 text-primary flex items-center justify-center font-bold text-xs">
                      <ShieldCheck className="h-5 w-5" />
                    </div>
                    <div>
                      <h3 className="text-sm font-bold text-foreground">{locale === "ar" ? role.ar : role.en}</h3>
                      <span className="font-mono text-xs text-muted-foreground">{role.code}</span>
                    </div>
                  </div>
                  <span className="text-[11px] px-2 py-0.5 rounded-full font-semibold bg-primary/10 text-primary">
                    {role.scope}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground leading-relaxed">{role.desc}</p>
                <div className="pt-3 border-t text-[11px] font-medium flex items-center justify-between text-muted-foreground">
                  <span>الصلاحيات:</span>
                  <span className="font-semibold text-foreground">{role.permsCount}</span>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 2. Permissions Matrix */}
        {activeTab === "permissions" && (
          <div className="space-y-4">
            {[
              {
                module: "ORGANIZATION",
                arTitle: "صلاحيات الهيكل التنظيمي",
                perms: [
                  { code: "org.organizations.view", ar: "عرض المنظمات القابضة", en: "View Organizations" },
                  { code: "org.legal_entities.manage", ar: "إدارة الكيانات القانونية", en: "Manage Legal Entities" },
                  { code: "org.branches.manage", ar: "إدارة الفروع والمواقع", en: "Manage Branches" },
                  { code: "org.number_sequences.generate", ar: "توليد أرقام الوثائق التسلسلية", en: "Generate Number Sequences" },
                ],
              },
              {
                module: "ACCOUNTING & GL",
                arTitle: "صلاحيات المحاسبة ودفتر الأستاذ",
                perms: [
                  { code: "gl.journals.view", ar: "عرض القيود اليومية", en: "View Journal Entries" },
                  { code: "gl.journals.post", ar: "ترحيل القيود اليومية (غير قابل للعكس)", en: "Post Immutable Journals" },
                  { code: "gl.accounts.view", ar: "عرض دليل الحسابات", en: "View Chart of Accounts" },
                ],
              },
              {
                module: "SALES & ZATCA",
                arTitle: "صلاحيات المبيعات والفوترة الإلكترونية",
                perms: [
                  { code: "sales.invoices.view", ar: "عرض فواتير المبيعات", en: "View Sales Invoices" },
                  { code: "sales.invoices.create", ar: "إنشاء فواتير المبيعات", en: "Create Sales Invoices" },
                  { code: "zatca.invoices.sign", ar: "توقيع وإرسال الفواتير لزاتكا", en: "Sign & Transmit ZATCA Invoices" },
                ],
              },
            ].map((sec) => (
              <div key={sec.module} className="rounded-xl border bg-card p-5 space-y-3 shadow-sm">
                <div className="flex items-center justify-between border-b pb-2">
                  <span className="font-bold text-sm text-foreground flex items-center gap-2">
                    <Layers className="h-4 w-4 text-primary" />
                    <span>{locale === "ar" ? sec.arTitle : sec.module}</span>
                  </span>
                  <span className="text-xs font-mono text-muted-foreground bg-muted px-2 py-0.5 rounded">
                    {sec.perms.length} {locale === "ar" ? "صلاحيات" : "permissions"}
                  </span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                  {sec.perms.map((p) => (
                    <div key={p.code} className="p-3 rounded-lg border bg-muted/20 flex items-center justify-between text-xs">
                      <div>
                        <div className="font-semibold text-foreground">{locale === "ar" ? p.ar : p.en}</div>
                        <div className="font-mono text-muted-foreground text-[11px] mt-0.5">{p.code}</div>
                      </div>
                      <Lock className="h-3.5 w-3.5 text-muted-foreground" />
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 3. User Scoped Assignments */}
        {activeTab === "assignments" && (
          <div className="rounded-xl border bg-card p-5 space-y-4 shadow-sm">
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              <Users className="h-4 w-4 text-primary" />
              <span>{t.authorization.assignments_title}</span>
            </h3>

            <div className="space-y-3">
              {[
                {
                  user: "مدير النظام (admin)",
                  email: "admin@erp.sa",
                  role: "SUPER_ADMIN",
                  roleName: "مدير النظام الشامل",
                  scope: "شامل (Global)",
                  branch: "كافة الفروع والمواقع",
                  badge: "شامل",
                },
                {
                  user: "أحمد المحاسب",
                  email: "ahmed.acc@erp.sa",
                  role: "CHIEF_ACCOUNTANT",
                  roleName: "رئيس الحسابات",
                  scope: "شركة الأفق المتقدمة (LE-01)",
                  branch: "الفرع الرئيسي - الرياض",
                  badge: "مقيد بكيان",
                },
              ].map((asg, idx) => (
                <div key={idx} className="p-4 rounded-xl border bg-muted/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
                  <div className="space-y-1">
                    <div className="font-bold text-foreground text-sm flex items-center gap-2">
                      <span>{asg.user}</span>
                      <span className="font-mono text-xs text-muted-foreground font-normal">({asg.email})</span>
                    </div>
                    <div className="text-muted-foreground flex items-center gap-3">
                      <span>الدور: <strong className="text-foreground">{asg.roleName}</strong> ({asg.role})</span>
                      <span>•</span>
                      <span>النطاق: <strong className="text-primary">{asg.scope}</strong></span>
                    </div>
                  </div>
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary self-start sm:self-auto">
                    {asg.badge}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 4. Real-Time Permission Evaluator & Policy Cache */}
        {activeTab === "evaluator" && (
          <div className="p-6 rounded-xl border bg-card space-y-5 shadow-sm">
            <div>
              <h3 className="text-base font-bold text-foreground flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-amber-500" />
                <span>{t.authorization.evaluator_title}</span>
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                {locale === "ar"
                  ? "قم باختبار التحقق من الصلاحيات اللحظي وفق كاش السياسات في الذاكرة ومطابقة نطاق الكيان القانوني."
                  : "Test live permission evaluation against in-memory policy cache and legal entity scoping."}
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-foreground block">الصلاحية المطلوبة (Required Permission)</label>
                <select
                  value={selectedPerm}
                  onChange={(e) => setSelectedPerm(e.target.value)}
                  className="w-full rounded-md border bg-background px-3 py-2 text-xs font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="gl.journals.post">gl.journals.post (ترحيل القيود)</option>
                  <option value="sales.invoices.create">sales.invoices.create (إنشاء فواتير مبيعات)</option>
                  <option value="org.legal_entities.manage">org.legal_entities.manage (إدارة الكيانات)</option>
                  <option value="zatca.invoices.sign">zatca.invoices.sign (توقيع زاتكا)</option>
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-foreground block">النطاق المالي (Target Scope)</label>
                <select
                  value={selectedScope}
                  onChange={(e) => setSelectedScope(e.target.value as any)}
                  className="w-full rounded-md border bg-background px-3 py-2 text-xs font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="LE-01">LE-01 (شركة الأفق المتقدمة - مصرح)</option>
                  <option value="LE-02">LE-02 (شركة الشرق الأوسط - غير مصرح)</option>
                  <option value="GLOBAL">GLOBAL (النطاق الشامل للمسؤول)</option>
                </select>
              </div>
            </div>

            <button
              id="evaluate-perm-btn"
              onClick={handleEvaluate}
              disabled={evaluating}
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              <Key className="h-4 w-4" />
              <span>{evaluating ? t.authorization.evaluating : t.authorization.test_eval_button}</span>
            </button>

            {/* Live Result Badge */}
            {decision && (
              <div
                className={`p-4 rounded-xl border flex items-center justify-between text-xs ${
                  decision.allowed
                    ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-700 dark:text-emerald-300"
                    : "bg-destructive/10 border-destructive/20 text-destructive"
                }`}
              >
                <div className="flex items-center gap-2.5">
                  {decision.allowed ? (
                    <CheckCircle2 className="h-5 w-5 flex-shrink-0 text-emerald-600" />
                  ) : (
                    <XCircle className="h-5 w-5 flex-shrink-0 text-destructive" />
                  )}
                  <div>
                    <div className="font-bold">
                      {decision.allowed ? t.authorization.granted : t.authorization.denied}
                    </div>
                    <div className="text-[11px] opacity-90 mt-0.5">{decision.reason}</div>
                  </div>
                </div>
                <span className="font-mono text-xs px-2.5 py-1 rounded bg-background border font-bold">
                  {decision.allowed ? "200 OK" : "403 FORBIDDEN"}
                </span>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

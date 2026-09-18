"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Settings,
  Building2,
  ReceiptText,
  ShieldCheck,
  FileCheck2,
  Save,
  CheckCircle2,
  RefreshCw,
  Globe,
  Lock,
  Database,
  Sliders,
  AlertCircle,
  Clock,
  Coins,
  MapPin,
  Laptop
} from "lucide-react";

export function SettingsView() {
  const { t, locale, dir } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"general" | "zatca" | "financial" | "security">("general");
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [saving, setSaving] = useState(false);

  // Form State
  const [formData, setFormData] = useState({
    companyNameAr: "شركة الأفق المتقدمة القابضة ذ.م.م",
    companyNameEn: "Saudi Advanced Horizon Holding LLC",
    crNumber: "1010998877",
    vatNumber: "310123456700003",
    nationalAddress: "المملكة العربية السعودية، الرياض، حي الملز، طريق الملك عبدالعزيز، مبنى 4210",
    postalCode: "12836",
    baseCurrency: "SAR",
    fiscalYearStart: "01-01",
    fiscalYearEnd: "12-31",
    inventoryMethod: "WEIGHTED_AVERAGE",
    zatcaEnv: "SIMULATION",
    zatcaCSIDStatus: "ACTIVE",
    enforceMFA: true,
    sessionTimeoutMins: "30",
    maxLoginAttempts: "5",
    auditImmutable: true,
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setTimeout(() => {
      setSaving(false);
      setSavedSuccess(true);
      setTimeout(() => setSavedSuccess(false), 4000);
    }, 600);
  };

  return (
    <div className="space-y-6" dir={dir}>
      {/* Header Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary mb-2">
            <Sliders className="h-3.5 w-3.5" />
            <span>{isAr ? "إدارة وتكوين النظام" : "System Configuration"}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Settings className="h-6 w-6 text-primary" />
            <span>{isAr ? "إعدادات النظام والمنشأة" : "System & Enterprise Settings"}</span>
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isAr
              ? "التحكم في بيانات الكيانات القانونية، خيارات الفوترة الإلكترونية، السياسات المحاسبية، وبروتوكولات الأمان"
              : "Manage legal entity profiles, ZATCA e-invoicing parameters, accounting policies, and security controls"}
          </p>
        </div>

        {/* Action Button */}
        <div className="flex items-center gap-3">
          {savedSuccess && (
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-xs font-semibold border border-emerald-500/20 animate-fade-in">
              <CheckCircle2 className="h-4 w-4" />
              <span>{isAr ? "تم حفظ الإعدادات بنجاح!" : "Settings saved successfully!"}</span>
            </div>
          )}
          <button
            onClick={handleSave}
            disabled={saving}
            className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground font-semibold text-sm shadow-sm hover:bg-primary/90 transition-all disabled:opacity-50 cursor-pointer"
          >
            {saving ? (
              <>
                <RefreshCw className="h-4 w-4 animate-spin" />
                <span>{isAr ? "جاري الحفظ..." : "Saving..."}</span>
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                <span>{isAr ? "حفظ التغييرات" : "Save Changes"}</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex items-center gap-2 border-b overflow-x-auto pb-px">
        {[
          { id: "general", label: isAr ? "البيانات المؤسسية" : "Enterprise Profile", icon: Building2 },
          { id: "zatca", label: isAr ? "إعدادات زاتكا والضريبة" : "ZATCA & Tax", icon: ReceiptText },
          { id: "financial", label: isAr ? "السياسات المالية والمحاسبية" : "Financial Policies", icon: Coins },
          { id: "security", label: isAr ? "الأمان والجلسات" : "Security & Access", icon: ShieldCheck },
        ].map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`
                flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-semibold border-b-2 transition-all cursor-pointer whitespace-nowrap
                ${
                  isActive
                    ? "border-primary text-primary bg-primary/5 rounded-t-lg"
                    : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
                }
              `}
            >
              <Icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab 1: General Enterprise Profile */}
      {activeTab === "general" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Building2 className="h-4 w-4 text-primary" />
              <span>{isAr ? "هوية المنشأة الرسمية" : "Official Legal Identity"}</span>
            </h3>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "اسم المنشأة التجاري (بالعربية)" : "Company Name (Arabic)"}
                </label>
                <input
                  type="text"
                  value={formData.companyNameAr}
                  onChange={(e) => setFormData({ ...formData, companyNameAr: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "اسم المنشأة التجاري (بالإنجليزية)" : "Company Name (English)"}
                </label>
                <input
                  type="text"
                  value={formData.companyNameEn}
                  onChange={(e) => setFormData({ ...formData, companyNameEn: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "رقم السجل التجاري (CR Number)" : "Commercial Registration"}
                  </label>
                  <input
                    type="text"
                    value={formData.crNumber}
                    maxLength={10}
                    onChange={(e) => setFormData({ ...formData, crNumber: e.target.value })}
                    className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                  <span className="text-[11px] text-muted-foreground mt-0.5 block">10 أرقام نظامية معتمدة</span>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الرقم الضريبي الموحد (VAT Number)" : "Tax Identification (TIN)"}
                  </label>
                  <input
                    type="text"
                    value={formData.vatNumber}
                    maxLength={15}
                    onChange={(e) => setFormData({ ...formData, vatNumber: e.target.value })}
                    className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                  <span className="text-[11px] text-muted-foreground mt-0.5 block">15 رقماً زكوياً معتمداً</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <MapPin className="h-4 w-4 text-primary" />
              <span>{isAr ? "العنوان الوطني المعتمد (KSA National Address)" : "National Address"}</span>
            </h3>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "العنوان التفصيلي" : "Full Address"}
                </label>
                <textarea
                  rows={3}
                  value={formData.nationalAddress}
                  onChange={(e) => setFormData({ ...formData, nationalAddress: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الرمز البريدي" : "Postal Code"}
                  </label>
                  <input
                    type="text"
                    value={formData.postalCode}
                    onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                    className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الدولة" : "Country"}
                  </label>
                  <input
                    type="text"
                    disabled
                    value={isAr ? "المملكة العربية السعودية (SA)" : "Saudi Arabia (SA)"}
                    className="w-full px-3.5 py-2 rounded-xl border bg-muted text-sm text-muted-foreground cursor-not-allowed"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: ZATCA & Tax */}
      {activeTab === "zatca" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <ReceiptText className="h-4 w-4 text-primary" />
              <span>{isAr ? "تكوين الربط مع هيئة الزكاة والضريبة (فاتورة)" : "ZATCA Fatoora Integration"}</span>
            </h3>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "بيئة الربط الإلكتروني" : "ZATCA Environment"}
                </label>
                <select
                  value={formData.zatcaEnv}
                  onChange={(e) => setFormData({ ...formData, zatcaEnv: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="SIMULATION">بيئة المحاكاة (Simulation Portal)</option>
                  <option value="PRODUCTION">بيئة الإنتاج الحقيقية (Production Portal)</option>
                  <option value="SANDBOX">بيئة فحص المطورين (Developer Sandbox)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "حالة شهادة وحدة الحل (CSID Certificate)" : "CSID Compliance Status"}
                </label>
                <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-between">
                  <div className="flex items-center gap-2 text-emerald-600 dark:text-emerald-400 text-xs font-bold">
                    <CheckCircle2 className="h-4 w-4" />
                    <span>{isAr ? "الشهادة صالحة ومطابقة للمواصفات" : "Valid Production CSID Certificate"}</span>
                  </div>
                  <span className="text-[11px] font-mono text-muted-foreground">Expires: 2027-09</span>
                </div>
              </div>

              <div className="p-4 rounded-xl bg-muted/50 border space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="text-muted-foreground">{isAr ? "خوارزمية التشفير والتوقيع:" : "Signing Algorithm:"}</span>
                  <span className="font-mono font-bold text-foreground">ECDSA secp256k1</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-muted-foreground">{isAr ? "صيغة الفواتير:" : "Invoice Format:"}</span>
                  <span className="font-mono font-bold text-foreground">UBL 2.1 XML + TLV QR</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-muted-foreground">{isAr ? "التحقق من صحة القواعد:" : "Validation Standard:"}</span>
                  <span className="font-bold text-emerald-600 dark:text-emerald-400">BR-KSA Schematron 100%</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Coins className="h-4 w-4 text-primary" />
              <span>{isAr ? "معدلات ضريبة القيمة المضافة (VAT Rates)" : "VAT Rules"}</span>
            </h3>

            <div className="space-y-3">
              <div className="p-3 rounded-xl border flex items-center justify-between">
                <div>
                  <div className="text-sm font-bold text-foreground">{isAr ? "النسبة الأساسية القياسية" : "Standard Rate"}</div>
                  <div className="text-xs text-muted-foreground">{isAr ? "تطبق على معظم السلع والخدمات" : "Applied to standard goods & services"}</div>
                </div>
                <span className="px-3 py-1 rounded-full bg-primary/10 text-primary font-bold text-sm">15%</span>
              </div>

              <div className="p-3 rounded-xl border flex items-center justify-between">
                <div>
                  <div className="text-sm font-bold text-foreground">{isAr ? "النسبة الصفرية (Zero Rated)" : "Zero-Rated"}</div>
                  <div className="text-xs text-muted-foreground">{isAr ? "الصادرات والنقل الدولي المؤهل" : "Exports & qualifying transport"}</div>
                </div>
                <span className="px-3 py-1 rounded-full bg-muted text-muted-foreground font-bold text-sm">0%</span>
              </div>

              <div className="p-3 rounded-xl border flex items-center justify-between">
                <div>
                  <div className="text-sm font-bold text-foreground">{isAr ? "السلع المعفاة (Exempt)" : "Tax Exempt"}</div>
                  <div className="text-xs text-muted-foreground">{isAr ? "الخدمات المالية المحددة وتأجير العقار السكني" : "Financial services & residential leases"}</div>
                </div>
                <span className="px-3 py-1 rounded-full bg-amber-500/10 text-amber-600 font-bold text-sm">معفى</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: Financial & Accounting Policies */}
      {activeTab === "financial" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Coins className="h-4 w-4 text-primary" />
              <span>{isAr ? "العملة والسنة المالية" : "Currency & Fiscal Periods"}</span>
            </h3>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "العملة الأساسية للنظام" : "Base Currency"}
                </label>
                <input
                  type="text"
                  disabled
                  value="SAR - ريال سعودي"
                  className="w-full px-3.5 py-2 rounded-xl border bg-muted font-bold text-sm text-foreground cursor-not-allowed"
                />
                <span className="text-[11px] text-muted-foreground mt-0.5 block">العملة السيادية المعتمدة لكافة دفاتر الأستاذ</span>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "بداية السنة المالية" : "Fiscal Start"}
                  </label>
                  <input
                    type="text"
                    value={formData.fiscalYearStart}
                    onChange={(e) => setFormData({ ...formData, fiscalYearStart: e.target.value })}
                    className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "نهاية السنة المالية" : "Fiscal End"}
                  </label>
                  <input
                    type="text"
                    value={formData.fiscalYearEnd}
                    onChange={(e) => setFormData({ ...formData, fiscalYearEnd: e.target.value })}
                    className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Database className="h-4 w-4 text-primary" />
              <span>{isAr ? "سياسات المخزون والقيود المحاسبية" : "Inventory & Posting Rules"}</span>
            </h3>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "طريقة تقييم المخزون المعتمدة" : "Inventory Valuation Method"}
                </label>
                <select
                  value={formData.inventoryMethod}
                  onChange={(e) => setFormData({ ...formData, inventoryMethod: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="WEIGHTED_AVERAGE">المتوسط المرجح المتحرك (Weighted Average - SOCPA Standard)</option>
                  <option value="FIFO">الوارد أولاً صادر أولاً (FIFO)</option>
                </select>
              </div>

              <div className="p-4 rounded-xl bg-primary/5 border border-primary/20 space-y-1">
                <div className="text-xs font-bold text-primary flex items-center gap-1.5">
                  <ShieldCheck className="h-4 w-4" />
                  <span>{isAr ? "ضمان عدم قابلية تعديل القيود (Immutability Guarantee)" : "Strict Journal Immutability"}</span>
                </div>
                <p className="text-[11px] text-muted-foreground leading-relaxed">
                  {isAr
                    ? "القيود المرحلة لا يمكن حذفها أو تعديلها برمجياً أو عبر واجهة المستخدم، والتصحيح يتم حصراً عبر قيود العكس (Reversals)."
                    : "Posted journals cannot be modified or deleted. Corrections require audited reversing entries."}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: Security & Access */}
      {activeTab === "security" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Lock className="h-4 w-4 text-primary" />
              <span>{isAr ? "أمان الجلسات وسياسة الدخول" : "Session & Authentication"}</span>
            </h3>

            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 rounded-xl border">
                <div>
                  <div className="text-sm font-bold text-foreground">{isAr ? "إلزامية المصادقة الثنائية (MFA)" : "Enforce Multi-Factor Auth"}</div>
                  <div className="text-xs text-muted-foreground">{isAr ? "إلزام الحسابات الإدارية برمز OTP عبر البريد أو التطبيق" : "Required for all privileged users"}</div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.enforceMFA}
                  onChange={(e) => setFormData({ ...formData, enforceMFA: e.target.checked })}
                  className="h-5 w-5 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "مهلة انتهاء الجلسة عند عدم النشاط (بالدقائق)" : "Session Inactivity Timeout (Minutes)"}
                </label>
                <input
                  type="number"
                  value={formData.sessionTimeoutMins}
                  onChange={(e) => setFormData({ ...formData, sessionTimeoutMins: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "الحد الأقصى لمحاولات تسجيل الدخول الفاشلة" : "Max Failed Login Attempts"}
                </label>
                <input
                  type="number"
                  value={formData.maxLoginAttempts}
                  onChange={(e) => setFormData({ ...formData, maxLoginAttempts: e.target.value })}
                  className="w-full px-3.5 py-2 rounded-xl border bg-background font-mono text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>
          </div>

          <div className="p-6 rounded-2xl border bg-card shadow-xs space-y-4">
            <h3 className="text-base font-bold text-foreground flex items-center gap-2">
              <Laptop className="h-4 w-4 text-primary" />
              <span>{isAr ? "معلومات النظام والإصدار الحالي" : "System Environment Info"}</span>
            </h3>

            <div className="space-y-3">
              <div className="flex justify-between items-center p-3 rounded-xl bg-muted/40 border text-xs">
                <span className="text-muted-foreground">{isAr ? "إصدار المنصة:" : "Platform Version:"}</span>
                <span className="font-bold font-mono px-2 py-0.5 rounded bg-primary/10 text-primary">v1.0.0 (Production)</span>
              </div>
              <div className="flex justify-between items-center p-3 rounded-xl bg-muted/40 border text-xs">
                <span className="text-muted-foreground">{isAr ? "محرك قاعدة البيانات:" : "Database Engine:"}</span>
                <span className="font-bold font-mono text-foreground">PostgreSQL 18.2 (pgx/v5)</span>
              </div>
              <div className="flex justify-between items-center p-3 rounded-xl bg-muted/40 border text-xs">
                <span className="text-muted-foreground">{isAr ? "الواجهة البرمجية (API):" : "Core API Runtime:"}</span>
                <span className="font-bold font-mono text-foreground">Go 1.27.1 (Modular Monolith)</span>
              </div>
              <div className="flex justify-between items-center p-3 rounded-xl bg-muted/40 border text-xs">
                <span className="text-muted-foreground">{isAr ? "النطاق الزمني السيادي:" : "Sovereign Timezone:"}</span>
                <span className="font-bold font-mono text-foreground">Asia/Riyadh (AST UTC+3)</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

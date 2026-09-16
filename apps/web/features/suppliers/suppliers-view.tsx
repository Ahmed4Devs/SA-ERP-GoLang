"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Truck,
  Building2,
  ShieldCheck,
  Plus,
  Search,
  CheckCircle2,
  AlertCircle,
  CreditCard,
  X,
  Landmark,
  Phone,
  Mail,
  MapPin,
  ExternalLink,
} from "lucide-react";

export interface SupplierItem {
  id: string;
  supplierNumber: string;
  supplierType: "LOCAL" | "INTERNATIONAL" | "GOVERNMENT";
  nameAr: string;
  nameEn: string;
  taxId?: string;
  crNumber?: string;
  iban?: string;
  bankName?: string;
  city: string;
  district?: string;
  streetName?: string;
  buildingNo?: string;
  postalCode?: string;
  email?: string;
  phone?: string;
  contactPerson?: string;
  paymentTermsCode: string;
  paymentTermsDueDays: number;
  openBalance: number;
  status: "ACTIVE" | "SUSPENDED" | "BLOCKED";
}

const INITIAL_SUPPLIERS: SupplierItem[] = [
  {
    id: "sup-1",
    supplierNumber: "SUP-0001",
    supplierType: "LOCAL",
    nameAr: "شركة الفوزان لمواد البناء",
    nameEn: "Al-Fozan Building Materials Co.",
    taxId: "310123456789013",
    crNumber: "1010345678",
    iban: "SA0380000000123456789012",
    bankName: "مصرف الراجحي",
    city: "الرياض",
    district: "العليا",
    streetName: "طريق الملك فهد",
    buildingNo: "7890",
    postalCode: "12214",
    email: "procurement@alfozan.example.com",
    phone: "+966114567890",
    contactPerson: "م. محمد الفوزان",
    paymentTermsCode: "NET30",
    paymentTermsDueDays: 30,
    openBalance: 45200.0,
    status: "ACTIVE",
  },
  {
    id: "sup-2",
    supplierNumber: "SUP-0002",
    supplierType: "LOCAL",
    nameAr: "الشركة السعودية للكهرباء (SEC)",
    nameEn: "Saudi Electricity Company",
    taxId: "300012345678903",
    crNumber: "1010000012",
    iban: "SA4410000000987654321098",
    bankName: "البنك الأهلي السعودي (SNB)",
    city: "الرياض",
    district: "الملز",
    streetName: "شارع العليا العام",
    buildingNo: "1234",
    postalCode: "11411",
    email: "billing@se.com.sa",
    phone: "+966118001234",
    contactPerson: "إدارة الفوترة والتحصيل",
    paymentTermsCode: "IMMEDIATE",
    paymentTermsDueDays: 0,
    openBalance: 12850.5,
    status: "ACTIVE",
  },
  {
    id: "sup-3",
    supplierNumber: "SUP-0003",
    supplierType: "INTERNATIONAL",
    nameAr: "شركة ديل العالمية للتقنية",
    nameEn: "Dell Technologies EMEA",
    taxId: "310998877600003",
    crNumber: "1010998877",
    iban: "SA5520000000456123789456",
    bankName: "بنك الرياض",
    city: "الرياض",
    district: "واحة غرناطة الرقمية",
    streetName: "طريق الدائري الشرقي",
    buildingNo: "3300",
    postalCode: "13241",
    email: "enterprise.sa@dell.com",
    phone: "+966112003000",
    contactPerson: "Mr. David Miller",
    paymentTermsCode: "NET45",
    paymentTermsDueDays: 45,
    openBalance: 185000.0,
    status: "ACTIVE",
  },
  {
    id: "sup-4",
    supplierNumber: "SUP-0004",
    supplierType: "GOVERNMENT",
    nameAr: "هيئة الزكاة والضريبة والجمارك (ZATCA)",
    nameEn: "Zakat, Tax and Customs Authority",
    taxId: "300000000000003",
    crNumber: "1010000001",
    city: "الرياض",
    district: "المربع",
    streetName: "طريق الملك فيصل",
    buildingNo: "1111",
    postalCode: "11187",
    email: "info@zatca.gov.sa",
    phone: "19993",
    contactPerson: "خدمات المكلفين",
    paymentTermsCode: "IMMEDIATE",
    paymentTermsDueDays: 0,
    openBalance: 0.0,
    status: "ACTIVE",
  },
];

export function SuppliersView() {
  const { locale, t } = useLocale();
  const isAr = locale === "ar";
  const sT = (t as any).suppliers || {};

  const [suppliers, setSuppliers] = useState<SupplierItem[]>(INITIAL_SUPPLIERS);
  const [search, setSearch] = useState("");
  const [selectedType, setSelectedType] = useState<string>("ALL");
  const [selectedStatus, setSelectedStatus] = useState<string>("ALL");
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Form State for new supplier
  const [formData, setFormData] = useState({
    supplierType: "LOCAL" as "LOCAL" | "INTERNATIONAL" | "GOVERNMENT",
    nameAr: "",
    nameEn: "",
    taxId: "",
    crNumber: "",
    iban: "",
    bankName: "",
    city: isAr ? "الرياض" : "Riyadh",
    district: "",
    streetName: "",
    buildingNo: "",
    postalCode: "",
    email: "",
    phone: "",
    contactPerson: "",
    paymentTermsCode: "NET30",
    paymentTermsDueDays: 30,
  });

  const [formErrors, setFormErrors] = useState<{ [key: string]: string }>({});

  const filteredSuppliers = suppliers.filter((s) => {
    const matchSearch =
      s.supplierNumber.toLowerCase().includes(search.toLowerCase()) ||
      s.nameAr.includes(search) ||
      s.nameEn.toLowerCase().includes(search.toLowerCase()) ||
      (s.taxId && s.taxId.includes(search)) ||
      (s.crNumber && s.crNumber.includes(search));

    const matchType = selectedType === "ALL" || s.supplierType === selectedType;
    const matchStatus = selectedStatus === "ALL" || s.status === selectedStatus;

    return matchSearch && matchType && matchStatus;
  });

  const totalSuppliers = suppliers.length;
  const activeSuppliers = suppliers.filter((s) => s.status === "ACTIVE").length;
  const localSuppliers = suppliers.filter((s) => s.supplierType === "LOCAL").length;
  const totalOpenBalance = suppliers.reduce((sum, s) => sum + s.openBalance, 0);

  const validateForm = () => {
    const errors: { [key: string]: string } = {};
    if (!formData.nameAr.trim()) errors.nameAr = isAr ? "الاسم بالعربية مطلوب" : "Arabic name is required";
    if (!formData.nameEn.trim()) errors.nameEn = isAr ? "الاسم بالإنجليزية مطلوب" : "English name is required";
    if (!formData.city.trim()) errors.city = isAr ? "المدينة مطلوبة" : "City is required";

    if (formData.taxId.trim()) {
      const vatRegex = /^3[0-9]{13}3$/;
      if (!vatRegex.test(formData.taxId.trim())) {
        errors.taxId = isAr
          ? "الرقم الضريبي السعودي يجب أن يتكون من 15 رقماً ويبدأ وينتهي بـ 3"
          : "Saudi VAT ID must be 15 digits starting and ending with 3";
      }
    }

    if (formData.crNumber.trim()) {
      const crRegex = /^[0-9]{10}$/;
      if (!crRegex.test(formData.crNumber.trim())) {
        errors.crNumber = isAr
          ? "السجل التجاري يجب أن يتكون من 10 أرقام"
          : "CR Number must be exactly 10 digits";
      }
    }

    if (formData.iban.trim()) {
      const cleanIban = formData.iban.replace(/\s+/g, "").toUpperCase();
      const ibanRegex = /^[A-Z]{2}[0-9]{2}[A-Z0-9]{4,30}$/;
      if (!ibanRegex.test(cleanIban)) {
        errors.iban = isAr ? "صيغة الآيبان غير صحيحة" : "Invalid IBAN format";
      }
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateSupplier = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    const newSup: SupplierItem = {
      id: `sup-${Date.now()}`,
      supplierNumber: `SUP-${String(suppliers.length + 1).padStart(4, "0")}`,
      supplierType: formData.supplierType,
      nameAr: formData.nameAr,
      nameEn: formData.nameEn,
      taxId: formData.taxId.trim() || undefined,
      crNumber: formData.crNumber.trim() || undefined,
      iban: formData.iban.trim() || undefined,
      bankName: formData.bankName.trim() || undefined,
      city: formData.city,
      district: formData.district.trim() || undefined,
      streetName: formData.streetName.trim() || undefined,
      buildingNo: formData.buildingNo.trim() || undefined,
      postalCode: formData.postalCode.trim() || undefined,
      email: formData.email.trim() || undefined,
      phone: formData.phone.trim() || undefined,
      contactPerson: formData.contactPerson.trim() || undefined,
      paymentTermsCode: formData.paymentTermsCode,
      paymentTermsDueDays: formData.paymentTermsDueDays,
      openBalance: 0.0,
      status: "ACTIVE",
    };

    setSuppliers([newSup, ...suppliers]);
    setIsModalOpen(false);
    setFormData({
      supplierType: "LOCAL",
      nameAr: "",
      nameEn: "",
      taxId: "",
      crNumber: "",
      iban: "",
      bankName: "",
      city: isAr ? "الرياض" : "Riyadh",
      district: "",
      streetName: "",
      buildingNo: "",
      postalCode: "",
      email: "",
      phone: "",
      contactPerson: "",
      paymentTermsCode: "NET30",
      paymentTermsDueDays: 30,
    });
    setFormErrors({});
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <Truck className="h-7 w-7 text-primary" />
            {sT.title || (isAr ? "سجل الموردين والدائنين" : "Supplier Master & AP Accounts")}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {sT.subtitle ||
              (isAr
                ? "إدارة بيانات الموردين، التحقق من السجل التجاري والـ 15 رقماً ضريبياً، وحسابات الذمم الدائنة"
                : "Manage suppliers, verify 15-digit ZATCA VAT & CR, and accounts payable subledgers")}
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground font-medium text-sm hover:bg-primary/90 transition shadow-sm"
        >
          <Plus className="h-4 w-4" />
          {sT.new_supplier || (isAr ? "إضافة مورد جديد" : "New Supplier")}
        </button>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {sT.kpi_total || (isAr ? "إجمالي الموردين" : "Total Suppliers")}
            </span>
            <Building2 className="h-4 w-4 text-primary" />
          </div>
          <div className="text-2xl font-bold mt-2">{totalSuppliers}</div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "مسجلين في دليل الموردين" : "Registered in master directory"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {sT.kpi_active || (isAr ? "الموردين النشطين" : "Active Suppliers")}
            </span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-emerald-600 dark:text-emerald-400">
            {activeSuppliers}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "مؤهلون لأوامر الشراء والصرف" : "Qualified for POs and payments"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {sT.kpi_local || (isAr ? "موردين محليين (المملكة)" : "Domestic Suppliers")}
            </span>
            <ShieldCheck className="h-4 w-4 text-sky-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-sky-600 dark:text-sky-400">
            {localSuppliers}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "خاضعون لضريبة 15% وعنوان وطني" : "15% VAT & Saudi address compliant"}
          </p>
        </div>

        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {sT.kpi_open_balance || (isAr ? "إجمالي المديونية القائمة" : "Total Outstanding AP")}
            </span>
            <CreditCard className="h-4 w-4 text-amber-500" />
          </div>
          <div className="text-2xl font-bold mt-2 text-amber-600 dark:text-amber-400">
            {totalOpenBalance.toLocaleString(isAr ? "ar-SA" : "en-US", {
              style: "currency",
              currency: "SAR",
            })}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {isAr ? "رصيد الذمم الدائنة المفتوحة" : "Unpaid invoices balance"}
          </p>
        </div>
      </div>

      {/* Filters and Search */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-3 bg-card border rounded-xl">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={isAr ? "بحث بالاسم، رقم المورد، الضريبي، السجل..." : "Search by name, #, VAT, CR..."}
            className="w-full bg-background border rounded-lg py-2 ps-9 pe-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <select
            value={selectedType}
            onChange={(e) => setSelectedType(e.target.value)}
            className="bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="ALL">{isAr ? "كل التصنيفات" : "All Classifications"}</option>
            <option value="LOCAL">{isAr ? "محلي (LOCAL)" : "Domestic"}</option>
            <option value="INTERNATIONAL">{isAr ? "دولي (INTERNATIONAL)" : "International"}</option>
            <option value="GOVERNMENT">{isAr ? "حكومي (GOVERNMENT)" : "Government"}</option>
          </select>

          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="ALL">{isAr ? "كل الحالات" : "All Statuses"}</option>
            <option value="ACTIVE">{isAr ? "نشط" : "Active"}</option>
            <option value="SUSPENDED">{isAr ? "موقوف" : "Suspended"}</option>
            <option value="BLOCKED">{isAr ? "محظور" : "Blocked"}</option>
          </select>
        </div>
      </div>

      {/* Suppliers Table */}
      <div className="border rounded-xl bg-card overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-xs text-start">
            <thead className="bg-muted/50 border-b text-muted-foreground">
              <tr>
                <th className="py-3 px-4 text-start font-medium">{sT.supplier_number || (isAr ? "رقم المورد" : "Supplier #")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.name || (isAr ? "اسم المورد" : "Supplier Name")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.type || (isAr ? "التصنيف" : "Type")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.tax_id || (isAr ? "الرقم الضريبي / السجل" : "VAT / CR")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.iban || (isAr ? "الآيبان والبنك" : "IBAN & Bank")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.city || (isAr ? "المدينة" : "City")}</th>
                <th className="py-3 px-4 text-start font-medium">{sT.payment_terms || (isAr ? "شروط الدفع" : "Terms")}</th>
                <th className="py-3 px-4 text-end font-medium">{isAr ? "الرصيد المستحق" : "Balance"}</th>
                <th className="py-3 px-4 text-center font-medium">{sT.status || (isAr ? "الحالة" : "Status")}</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {filteredSuppliers.length === 0 ? (
                <tr>
                  <td colSpan={9} className="py-8 text-center text-muted-foreground">
                    {isAr ? "لا يوجد موردين مطابقين لمعايير البحث" : "No suppliers match search criteria"}
                  </td>
                </tr>
              ) : (
                filteredSuppliers.map((s) => (
                  <tr key={s.id} className="hover:bg-muted/30 transition">
                    <td className="py-3 px-4 font-mono font-medium text-foreground">
                      {s.supplierNumber}
                    </td>
                    <td className="py-3 px-4">
                      <div className="font-semibold text-foreground">{isAr ? s.nameAr : s.nameEn}</div>
                      <div className="text-[10px] text-muted-foreground">{isAr ? s.nameEn : s.nameAr}</div>
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium border ${
                          s.supplierType === "LOCAL"
                            ? "bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950 dark:text-sky-300"
                            : s.supplierType === "GOVERNMENT"
                            ? "bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950 dark:text-purple-300"
                            : "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950 dark:text-amber-300"
                        }`}
                      >
                        {s.supplierType}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono">
                      {s.taxId ? (
                        <div className="text-foreground">{s.taxId}</div>
                      ) : (
                        <span className="text-muted-foreground text-[10px]">--</span>
                      )}
                      {s.crNumber && (
                        <div className="text-[10px] text-muted-foreground">CR: {s.crNumber}</div>
                      )}
                    </td>
                    <td className="py-3 px-4">
                      {s.iban ? (
                        <div className="font-mono text-foreground text-[11px]">{s.iban}</div>
                      ) : (
                        <span className="text-muted-foreground">--</span>
                      )}
                      {s.bankName && (
                        <div className="text-[10px] text-muted-foreground flex items-center gap-1">
                          <Landmark className="h-2.5 w-2.5" />
                          {s.bankName}
                        </div>
                      )}
                    </td>
                    <td className="py-3 px-4 text-foreground">
                      {s.city}
                      {s.district && <span className="text-muted-foreground">, {s.district}</span>}
                    </td>
                    <td className="py-3 px-4">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-secondary border">
                        {s.paymentTermsCode} ({s.paymentTermsDueDays}d)
                      </span>
                    </td>
                    <td className="py-3 px-4 text-end font-mono font-medium text-foreground">
                      {s.openBalance.toLocaleString(isAr ? "ar-SA" : "en-US", {
                        minimumFractionDigits: 2,
                      })}{" "}
                      <span className="text-[10px] text-muted-foreground">SAR</span>
                    </td>
                    <td className="py-3 px-4 text-center">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium ${
                          s.status === "ACTIVE"
                            ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
                            : s.status === "SUSPENDED"
                            ? "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300"
                            : "bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
                        }`}
                      >
                        {s.status === "ACTIVE" ? (isAr ? "نشط" : "Active") : s.status}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal: New Supplier */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
            <div className="flex items-center justify-between p-5 border-b sticky top-0 bg-card z-10">
              <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
                <Truck className="h-5 w-5 text-primary" />
                {sT.new_supplier || (isAr ? "إضافة مورد جديد" : "New Supplier")}
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreateSupplier} className="p-5 space-y-4">
              {/* Classification */}
              <div>
                <label className="block text-xs font-semibold mb-1 text-foreground">
                  {sT.type || (isAr ? "تصنيف المورد" : "Supplier Classification")}
                </label>
                <div className="grid grid-cols-3 gap-2">
                  {(["LOCAL", "INTERNATIONAL", "GOVERNMENT"] as const).map((type) => (
                    <button
                      key={type}
                      type="button"
                      onClick={() => setFormData({ ...formData, supplierType: type })}
                      className={`py-2 px-3 rounded-lg border text-xs font-medium text-center transition ${
                        formData.supplierType === type
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-background hover:bg-muted text-foreground"
                      }`}
                    >
                      {type === "LOCAL" ? (isAr ? "محلي (المملكة)" : "Local") : type === "GOVERNMENT" ? (isAr ? "حكومي" : "Government") : (isAr ? "دولي" : "International")}
                    </button>
                  ))}
                </div>
              </div>

              {/* Names */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {isAr ? "الاسم بالعربية *" : "Arabic Name *"}
                  </label>
                  <input
                    type="text"
                    required
                    dir="rtl"
                    value={formData.nameAr}
                    onChange={(e) => setFormData({ ...formData, nameAr: e.target.value })}
                    className={`w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 ${
                      formErrors.nameAr ? "border-destructive ring-destructive" : "focus:ring-primary"
                    }`}
                  />
                  {formErrors.nameAr && (
                    <p className="text-[10px] text-destructive mt-1">{formErrors.nameAr}</p>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {isAr ? "الاسم بالإنجليزية *" : "English Name *"}
                  </label>
                  <input
                    type="text"
                    required
                    dir="ltr"
                    value={formData.nameEn}
                    onChange={(e) => setFormData({ ...formData, nameEn: e.target.value })}
                    className={`w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 ${
                      formErrors.nameEn ? "border-destructive ring-destructive" : "focus:ring-primary"
                    }`}
                  />
                  {formErrors.nameEn && (
                    <p className="text-[10px] text-destructive mt-1">{formErrors.nameEn}</p>
                  )}
                </div>
              </div>

              {/* Tax ID & CR */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {sT.tax_id || (isAr ? "الرقم الضريبي (15 رقماً)" : "Saudi VAT ID (15 digits)")}
                  </label>
                  <input
                    type="text"
                    maxLength={15}
                    placeholder="3xxxxxxxxxxxx03"
                    value={formData.taxId}
                    onChange={(e) => setFormData({ ...formData, taxId: e.target.value })}
                    className={`w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 ${
                      formErrors.taxId ? "border-destructive ring-destructive" : "focus:ring-primary"
                    }`}
                  />
                  {formErrors.taxId && (
                    <p className="text-[10px] text-destructive mt-1">{formErrors.taxId}</p>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {sT.cr_number || (isAr ? "السجل التجاري (10 أرقام)" : "CR Number (10 digits)")}
                  </label>
                  <input
                    type="text"
                    maxLength={10}
                    placeholder="1010xxxxxx"
                    value={formData.crNumber}
                    onChange={(e) => setFormData({ ...formData, crNumber: e.target.value })}
                    className={`w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono focus:outline-none focus:ring-2 ${
                      formErrors.crNumber ? "border-destructive ring-destructive" : "focus:ring-primary"
                    }`}
                  />
                  {formErrors.crNumber && (
                    <p className="text-[10px] text-destructive mt-1">{formErrors.crNumber}</p>
                  )}
                </div>
              </div>

              {/* Bank & IBAN */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {sT.iban || (isAr ? "رقم الآيبان (IBAN)" : "IBAN")}
                  </label>
                  <input
                    type="text"
                    placeholder="SA0000000000000000000000"
                    value={formData.iban}
                    onChange={(e) => setFormData({ ...formData, iban: e.target.value })}
                    className={`w-full bg-background border rounded-lg px-3 py-2 text-xs font-mono uppercase focus:outline-none focus:ring-2 ${
                      formErrors.iban ? "border-destructive ring-destructive" : "focus:ring-primary"
                    }`}
                  />
                  {formErrors.iban && (
                    <p className="text-[10px] text-destructive mt-1">{formErrors.iban}</p>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {isAr ? "اسم البنك" : "Bank Name"}
                  </label>
                  <input
                    type="text"
                    placeholder={isAr ? "مثال: مصرف الراجحي" : "e.g. Al Rajhi Bank"}
                    value={formData.bankName}
                    onChange={(e) => setFormData({ ...formData, bankName: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              {/* Saudi National Address */}
              <div className="border rounded-xl p-3 bg-muted/20 space-y-3">
                <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                  <MapPin className="h-3.5 w-3.5 text-primary" />
                  {isAr ? "العنوان الوطني ومقر التوريد" : "National Address & Location"}
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                  <div>
                    <label className="block text-[11px] text-muted-foreground mb-1">
                      {isAr ? "المدينة *" : "City *"}
                    </label>
                    <input
                      type="text"
                      required
                      value={formData.city}
                      onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                      className="w-full bg-background border rounded-lg px-2.5 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] text-muted-foreground mb-1">
                      {isAr ? "الحي" : "District"}
                    </label>
                    <input
                      type="text"
                      value={formData.district}
                      onChange={(e) => setFormData({ ...formData, district: e.target.value })}
                      className="w-full bg-background border rounded-lg px-2.5 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] text-muted-foreground mb-1">
                      {isAr ? "اسم الشارع" : "Street"}
                    </label>
                    <input
                      type="text"
                      value={formData.streetName}
                      onChange={(e) => setFormData({ ...formData, streetName: e.target.value })}
                      className="w-full bg-background border rounded-lg px-2.5 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] text-muted-foreground mb-1">
                      {isAr ? "رقم المبنى" : "Building #"}
                    </label>
                    <input
                      type="text"
                      maxLength={4}
                      value={formData.buildingNo}
                      onChange={(e) => setFormData({ ...formData, buildingNo: e.target.value })}
                      className="w-full bg-background border rounded-lg px-2.5 py-1.5 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                </div>
              </div>

              {/* Payment Terms & Contact */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {sT.payment_terms || (isAr ? "شروط السداد" : "Payment Terms")}
                  </label>
                  <select
                    value={formData.paymentTermsCode}
                    onChange={(e) => {
                      const code = e.target.value;
                      const days = code === "IMMEDIATE" ? 0 : code === "NET15" ? 15 : code === "NET45" ? 45 : 30;
                      setFormData({ ...formData, paymentTermsCode: code, paymentTermsDueDays: days });
                    }}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="IMMEDIATE">{isAr ? "سداد فوري (Immediate)" : "Immediate"}</option>
                    <option value="NET15">{isAr ? "صافي 15 يوماً (Net 15)" : "Net 15 Days"}</option>
                    <option value="NET30">{isAr ? "صافي 30 يوماً (Net 30)" : "Net 30 Days"}</option>
                    <option value="NET45">{isAr ? "صافي 45 يوماً (Net 45)" : "Net 45 Days"}</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {isAr ? "البريد الإلكتروني" : "Email"}
                  </label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium mb-1 text-foreground">
                    {isAr ? "رقم الهاتف" : "Phone"}
                  </label>
                  <input
                    type="text"
                    value={formData.phone}
                    onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex items-center justify-end gap-2 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 rounded-lg border text-xs font-medium hover:bg-muted transition"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 transition shadow-sm"
                >
                  {isAr ? "حفظ المورد" : "Save Supplier"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

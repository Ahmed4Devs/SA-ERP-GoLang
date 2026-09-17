"use client";

import React, { useState } from "react";
import {
  Layers,
  Building,
  Truck,
  Monitor,
  Armchair,
  Wrench,
  Search,
  Plus,
  RefreshCw,
  Calendar,
  CheckCircle2,
  AlertCircle,
  TrendingDown,
  FileSpreadsheet,
  ArrowRight,
  ArrowLeftRight,
  Trash2,
  Eye,
  SlidersHorizontal,
  DollarSign,
  Download,
  BookOpen,
  MapPin,
  User,
  Clock,
  ShieldAlert,
  HelpCircle,
  X,
  Sparkles,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface FixedAssetItem {
  id: string;
  assetNumber: string;
  nameAr: string;
  nameEn: string;
  categoryCode: string;
  categoryNameAr: string;
  categoryNameEn: string;
  locationCode: string;
  locationNameAr: string;
  locationNameEn: string;
  custodianNameAr?: string;
  custodianNameEn?: string;
  acquisitionDate: string;
  inServiceDate: string;
  acquisitionCost: number;
  salvageValue: number;
  usefulLifeMonths: number;
  depreciationMethod: "STRAIGHT_LINE";
  currentNetBookValue: number;
  accumulatedDepreciation: number;
  status: "ACTIVE" | "FULLY_DEPRECIATED" | "DISPOSED" | "UNDER_MAINTENANCE";
}

interface DepreciationRunItem {
  id: string;
  runNumber: string;
  periodYear: number;
  periodMonth: number;
  bookCode: string;
  bookNameAr: string;
  totalAssets: number;
  totalDepreciation: number;
  runDate: string;
  status: "CALCULATED" | "POSTED";
  journalEntryNumber?: string;
}

interface AssetCategoryItem {
  id: string;
  code: string;
  nameAr: string;
  nameEn: string;
  usefulLifeMonths: number;
  depreciationMethod: string;
  assetAccountCode: string;
  assetAccountNameAr: string;
  accumAccountCode: string;
  accumAccountNameAr: string;
  expenseAccountCode: string;
  expenseAccountNameAr: string;
}

interface AssetTransferItem {
  id: string;
  transferNumber: string;
  assetNumber: string;
  assetNameAr: string;
  fromLocationNameAr: string;
  toLocationNameAr: string;
  custodianNameAr: string;
  transferDate: string;
  reason: string;
  status: "COMPLETED" | "PENDING";
}

export function FixedAssetsView() {
  const { dir, locale } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"assets" | "runs" | "categories" | "transfers">("assets");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("ALL");
  const [selectedStatus, setSelectedStatus] = useState("ALL");

  // Modals state
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [showRunModal, setShowRunModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showDisposeModal, setShowDisposeModal] = useState(false);
  const [showScheduleModal, setShowScheduleModal] = useState(false);
  const [showGLEntryModal, setShowGLEntryModal] = useState(false);

  // Selected asset for actions
  const [selectedAsset, setSelectedAsset] = useState<FixedAssetItem | null>(null);
  const [selectedRun, setSelectedRun] = useState<DepreciationRunItem | null>(null);

  // Form states
  const [newAssetForm, setNewAssetForm] = useState({
    assetNumber: `AST-${Math.floor(1000 + Math.random() * 9000)}`,
    nameAr: "",
    nameEn: "",
    categoryCode: "IT_EQUIP",
    locationCode: "HQ-RUH",
    acquisitionCost: 12000,
    salvageValue: 500,
    usefulLifeMonths: 36,
    acquisitionDate: new Date().toISOString().split("T")[0],
    inServiceDate: new Date().toISOString().split("T")[0],
  });

  const [transferForm, setTransferForm] = useState({
    toLocationCode: "BR-JED",
    custodianNameAr: "م. خالد السبيعي",
    custodianNameEn: "Eng. Khaled Al-Subaie",
    reason: "نقل معدات لتجهيز فرع المنطقة الغربية",
  });

  const [disposeForm, setDisposeForm] = useState({
    disposalType: "SALE" as "SALE" | "SCRAP",
    proceedsAmount: 4500,
    reason: "تحديث أسطول الأجهزة وتقادم المواصفات الفنية",
  });

  const [newRunForm, setNewRunForm] = useState({
    periodYear: 2026,
    periodMonth: 9,
    bookCode: "BOOK-FIN",
  });

  // Seed / Initial Data
  const [assets, setAssets] = useState<FixedAssetItem[]>([
    {
      id: "ast-1",
      assetNumber: "AST-2024-001",
      nameAr: "خادم رئيسي ديل باور إيدج R750",
      nameEn: "Dell PowerEdge R750 Rack Server",
      categoryCode: "IT_EQUIP",
      categoryNameAr: "أجهزة وخوادم تقنية",
      categoryNameEn: "IT & Server Equipment",
      locationCode: "HQ-RUH",
      locationNameAr: "المقر الرئيسي - غرفة الخوادم",
      locationNameEn: "HQ - Data Center",
      custodianNameAr: "م. سلطان القحطاني",
      custodianNameEn: "Eng. Sultan Al-Qahtani",
      acquisitionDate: "2024-01-15",
      inServiceDate: "2024-01-15",
      acquisitionCost: 120000,
      salvageValue: 12000,
      usefulLifeMonths: 36,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 42000,
      accumulatedDepreciation: 78000,
      status: "ACTIVE",
    },
    {
      id: "ast-2",
      assetNumber: "AST-2023-014",
      nameAr: "شاحنة نقل مبردة إيسوزو FTR",
      nameEn: "Isuzu FTR Refrigerated Truck",
      categoryCode: "VEHICLES",
      categoryNameAr: "أسطول النقل والسيارات",
      categoryNameEn: "Vehicles & Transport",
      locationCode: "WH-SUL",
      locationNameAr: "مستودع السلي اللوجستي",
      locationNameEn: "Al-Suly Central Warehouse",
      custodianNameAr: "محمد العتيبي",
      custodianNameEn: "Mohammed Al-Otaibi",
      acquisitionDate: "2023-03-10",
      inServiceDate: "2023-03-10",
      acquisitionCost: 260000,
      salvageValue: 20000,
      usefulLifeMonths: 60,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 92000,
      accumulatedDepreciation: 168000,
      status: "ACTIVE",
    },
    {
      id: "ast-3",
      assetNumber: "AST-2022-005",
      nameAr: "مبنى الإدارة العامة والمكاتب",
      nameEn: "General Administration Building",
      categoryCode: "BUILDINGS",
      categoryNameAr: "المباني والإنشاءات",
      categoryNameEn: "Buildings & Infrastructure",
      locationCode: "HQ-RUH",
      locationNameAr: "حي الصحافة - الرياض",
      locationNameEn: "Al-Sahafa Dist - Riyadh",
      custodianNameAr: "إدارة المرافق والخدمات",
      custodianNameEn: "Facilities Management",
      acquisitionDate: "2022-01-01",
      inServiceDate: "2022-01-01",
      acquisitionCost: 4500000,
      salvageValue: 500000,
      usefulLifeMonths: 240,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 3750000,
      accumulatedDepreciation: 750000,
      status: "ACTIVE",
    },
    {
      id: "ast-4",
      assetNumber: "AST-2023-088",
      nameAr: "طابعة صناعية متعددة المهام كانون",
      nameEn: "Canon Multifunction Production Printer",
      categoryCode: "OFFICE_FURN",
      categoryNameAr: "الأثاث والتجهيزات المكتبية",
      categoryNameEn: "Office Furniture & Equip",
      locationCode: "BR-JED",
      locationNameAr: "فرع الأندلس - جدة",
      locationNameEn: "Al-Andalus Branch - Jeddah",
      custodianNameAr: "فاطمة الغامدي",
      custodianNameEn: "Fatima Al-Ghamdi",
      acquisitionDate: "2023-06-01",
      inServiceDate: "2023-06-01",
      acquisitionCost: 35000,
      salvageValue: 3000,
      usefulLifeMonths: 36,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 6000,
      accumulatedDepreciation: 29000,
      status: "ACTIVE",
    },
    {
      id: "ast-5",
      assetNumber: "AST-2021-002",
      nameAr: "أجهزة حاسوب محمولة لينوفو ثينك باد (دفعة 10)",
      nameEn: "Lenovo ThinkPad Laptops Batch (10 Units)",
      categoryCode: "IT_EQUIP",
      categoryNameAr: "أجهزة وخوادم تقنية",
      categoryNameEn: "IT & Server Equipment",
      locationCode: "HQ-RUH",
      locationNameAr: "المقر الرئيسي - الرياض",
      locationNameEn: "HQ - Riyadh",
      custodianNameAr: "فريق تطوير البرمجيات",
      custodianNameEn: "Dev Engineering Team",
      acquisitionDate: "2021-02-01",
      inServiceDate: "2021-02-01",
      acquisitionCost: 55000,
      salvageValue: 5000,
      usefulLifeMonths: 36,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 5000,
      accumulatedDepreciation: 50000,
      status: "FULLY_DEPRECIATED",
    },
    {
      id: "ast-6",
      assetNumber: "AST-2020-041",
      nameAr: "سيارة سيدان تويوتا كامري 2020",
      nameEn: "Toyota Camry Sedan 2020",
      categoryCode: "VEHICLES",
      categoryNameAr: "أسطول النقل والسيارات",
      categoryNameEn: "Vehicles & Transport",
      locationCode: "BR-DAM",
      locationNameAr: "فرع الدمام الشرقي",
      locationNameEn: "Dammam East Branch",
      custodianNameAr: "إدارة المبيعات الميدانية",
      custodianNameEn: "Field Sales Management",
      acquisitionDate: "2020-05-15",
      inServiceDate: "2020-05-15",
      acquisitionCost: 95000,
      salvageValue: 15000,
      usefulLifeMonths: 48,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: 0,
      accumulatedDepreciation: 80000,
      status: "DISPOSED",
    },
  ]);

  const [runs, setRuns] = useState<DepreciationRunItem[]>([
    {
      id: "run-1",
      runNumber: "RUN-202608-001",
      periodYear: 2026,
      periodMonth: 8,
      bookCode: "BOOK-FIN",
      bookNameAr: "الدفتر المالي القانوني (IFRS)",
      totalAssets: 5,
      totalDepreciation: 31444.44,
      runDate: "2026-08-31",
      status: "POSTED",
      journalEntryNumber: "JV-2026-08-0144",
    },
    {
      id: "run-2",
      runNumber: "RUN-202607-001",
      periodYear: 2026,
      periodMonth: 7,
      bookCode: "BOOK-FIN",
      bookNameAr: "الدفتر المالي القانوني (IFRS)",
      totalAssets: 5,
      totalDepreciation: 31444.44,
      runDate: "2026-07-31",
      status: "POSTED",
      journalEntryNumber: "JV-2026-07-0112",
    },
    {
      id: "run-3",
      runNumber: "RUN-202609-001",
      periodYear: 2026,
      periodMonth: 9,
      bookCode: "BOOK-FIN",
      bookNameAr: "الدفتر المالي القانوني (IFRS)",
      totalAssets: 5,
      totalDepreciation: 31444.44,
      runDate: "2026-09-17",
      status: "CALCULATED",
    },
  ]);

  const categories: AssetCategoryItem[] = [
    {
      id: "cat-1",
      code: "BUILDINGS",
      nameAr: "المباني والإنشاءات",
      nameEn: "Buildings & Infrastructure",
      usefulLifeMonths: 240,
      depreciationMethod: "القسط الثابت (Straight Line)",
      assetAccountCode: "121000",
      assetAccountNameAr: "الأصول الثابتة - مباني وعقارات",
      accumAccountCode: "129000",
      accumAccountNameAr: "مجمع استهلاك الأصول الثابتة",
      expenseAccountCode: "525000",
      expenseAccountNameAr: "مصروف استهلاك الأصول الثابتة",
    },
    {
      id: "cat-2",
      code: "VEHICLES",
      nameAr: "أسطول النقل والسيارات",
      nameEn: "Vehicles & Transport",
      usefulLifeMonths: 60,
      depreciationMethod: "القسط الثابت (Straight Line)",
      assetAccountCode: "122000",
      assetAccountNameAr: "الأصول الثابتة - سيارات وشاحنات",
      accumAccountCode: "129000",
      accumAccountNameAr: "مجمع استهلاك الأصول الثابتة",
      expenseAccountCode: "525000",
      expenseAccountNameAr: "مصروف استهلاك الأصول الثابتة",
    },
    {
      id: "cat-3",
      code: "IT_EQUIP",
      nameAr: "أجهزة وخوادم تقنية",
      nameEn: "IT & Server Equipment",
      usefulLifeMonths: 36,
      depreciationMethod: "القسط الثابت (Straight Line)",
      assetAccountCode: "123000",
      assetAccountNameAr: "الأصول الثابتة - أجهزة تقنية",
      accumAccountCode: "129000",
      accumAccountNameAr: "مجمع استهلاك الأصول الثابتة",
      expenseAccountCode: "525000",
      expenseAccountNameAr: "مصروف استهلاك الأصول الثابتة",
    },
    {
      id: "cat-4",
      code: "OFFICE_FURN",
      nameAr: "الأثاث والتجهيزات المكتبية",
      nameEn: "Office Furniture & Equip",
      usefulLifeMonths: 48,
      depreciationMethod: "القسط الثابت (Straight Line)",
      assetAccountCode: "124000",
      assetAccountNameAr: "الأصول الثابتة - أثاث وديكور",
      accumAccountCode: "129000",
      accumAccountNameAr: "مجمع استهلاك الأصول الثابتة",
      expenseAccountCode: "525000",
      expenseAccountNameAr: "مصروف استهلاك الأصول الثابتة",
    },
  ];

  const [transfers, setTransfers] = useState<AssetTransferItem[]>([
    {
      id: "trf-1",
      transferNumber: "TRF-2026-003",
      assetNumber: "AST-2023-088",
      assetNameAr: "طابعة صناعية متعددة المهام كانون",
      fromLocationNameAr: "المقر الرئيسي - الرياض",
      toLocationNameAr: "فرع الأندلس - جدة",
      custodianNameAr: "فاطمة الغامدي",
      transferDate: "2026-05-10",
      reason: "دعم عمليات الطباعة بفرع جدة الجديد",
      status: "COMPLETED",
    },
    {
      id: "trf-2",
      transferNumber: "TRF-2026-002",
      assetNumber: "AST-2024-001",
      assetNameAr: "خادم رئيسي ديل باور إيدج R750",
      fromLocationNameAr: "مستودع السلي",
      toLocationNameAr: "المقر الرئيسي - غرفة الخوادم",
      custodianNameAr: "م. سلطان القحطاني",
      transferDate: "2026-01-20",
      reason: "تركيب وتشغيل بيئة الخوادم المركزية",
      status: "COMPLETED",
    },
  ]);

  // Aggregate KPIs
  const totalHistoricalCost = assets.reduce((sum, a) => sum + (a.status !== "DISPOSED" ? a.acquisitionCost : 0), 0);
  const totalNetBookValue = assets.reduce((sum, a) => sum + (a.status !== "DISPOSED" ? a.currentNetBookValue : 0), 0);
  const totalAccumulatedDepr = assets.reduce((sum, a) => sum + (a.status !== "DISPOSED" ? a.accumulatedDepreciation : 0), 0);
  const activeAssetsCount = assets.filter((a) => a.status === "ACTIVE").length;

  // Filtered Assets
  const filteredAssets = assets.filter((a) => {
    const matchesSearch =
      a.assetNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
      a.nameAr.includes(searchQuery) ||
      a.nameEn.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = selectedCategory === "ALL" || a.categoryCode === selectedCategory;
    const matchesStatus = selectedStatus === "ALL" || a.status === selectedStatus;
    return matchesSearch && matchesCategory && matchesStatus;
  });

  // Handlers
  const handleRegisterAsset = (e: React.FormEvent) => {
    e.preventDefault();
    const cat = categories.find((c) => c.code === newAssetForm.categoryCode) || categories[2];
    const cost = Number(newAssetForm.acquisitionCost);
    const salvage = Number(newAssetForm.salvageValue);
    const usefulLife = Number(newAssetForm.usefulLifeMonths);

    const created: FixedAssetItem = {
      id: `ast-${Date.now()}`,
      assetNumber: newAssetForm.assetNumber,
      nameAr: newAssetForm.nameAr || "أصل ثابت جديد",
      nameEn: newAssetForm.nameEn || "New Fixed Asset",
      categoryCode: cat.code,
      categoryNameAr: cat.nameAr,
      categoryNameEn: cat.nameEn,
      locationCode: newAssetForm.locationCode,
      locationNameAr: newAssetForm.locationCode === "HQ-RUH" ? "المقر الرئيسي - الرياض" : "فرع جدة التجاري",
      locationNameEn: newAssetForm.locationCode === "HQ-RUH" ? "HQ - Riyadh" : "Jeddah Branch",
      acquisitionDate: newAssetForm.acquisitionDate,
      inServiceDate: newAssetForm.inServiceDate,
      acquisitionCost: cost,
      salvageValue: salvage,
      usefulLifeMonths: usefulLife,
      depreciationMethod: "STRAIGHT_LINE",
      currentNetBookValue: cost,
      accumulatedDepreciation: 0,
      status: "ACTIVE",
    };

    setAssets([created, ...assets]);
    setShowRegisterModal(false);
    setNewAssetForm({
      assetNumber: `AST-${Math.floor(1000 + Math.random() * 9000)}`,
      nameAr: "",
      nameEn: "",
      categoryCode: "IT_EQUIP",
      locationCode: "HQ-RUH",
      acquisitionCost: 12000,
      salvageValue: 500,
      usefulLifeMonths: 36,
      acquisitionDate: new Date().toISOString().split("T")[0],
      inServiceDate: new Date().toISOString().split("T")[0],
    });
  };

  const handlePostRun = (runId: string) => {
    setRuns(
      runs.map((r) =>
        r.id === runId
          ? {
              ...r,
              status: "POSTED",
              journalEntryNumber: `JV-2026-${String(r.periodMonth).padStart(2, "0")}-${Math.floor(1000 + Math.random() * 9000)}`,
            }
          : r
      )
    );
  };

  const handleCreateRun = (e: React.FormEvent) => {
    e.preventDefault();
    const newRun: DepreciationRunItem = {
      id: `run-${Date.now()}`,
      runNumber: `RUN-${newRunForm.periodYear}${String(newRunForm.periodMonth).padStart(2, "0")}-001`,
      periodYear: Number(newRunForm.periodYear),
      periodMonth: Number(newRunForm.periodMonth),
      bookCode: newRunForm.bookCode,
      bookNameAr: "الدفتر المالي القانوني (IFRS)",
      totalAssets: activeAssetsCount,
      totalDepreciation: 31444.44,
      runDate: new Date().toISOString().split("T")[0],
      status: "CALCULATED",
    };
    setRuns([newRun, ...runs]);
    setShowRunModal(false);
  };

  const handleTransferAsset = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAsset) return;

    const newTransfer: AssetTransferItem = {
      id: `trf-${Date.now()}`,
      transferNumber: `TRF-2026-${Math.floor(100 + Math.random() * 900)}`,
      assetNumber: selectedAsset.assetNumber,
      assetNameAr: selectedAsset.nameAr,
      fromLocationNameAr: selectedAsset.locationNameAr,
      toLocationNameAr: transferForm.toLocationCode === "BR-JED" ? "فرع الأندلس - جدة" : "مستودع السلي اللوجستي",
      custodianNameAr: transferForm.custodianNameAr,
      transferDate: new Date().toISOString().split("T")[0],
      reason: transferForm.reason,
      status: "COMPLETED",
    };

    setTransfers([newTransfer, ...transfers]);
    setAssets(
      assets.map((a) =>
        a.id === selectedAsset.id
          ? {
              ...a,
              locationCode: transferForm.toLocationCode,
              locationNameAr: transferForm.toLocationCode === "BR-JED" ? "فرع الأندلس - جدة" : "مستودع السلي اللوجستي",
              custodianNameAr: transferForm.custodianNameAr,
            }
          : a
      )
    );
    setShowTransferModal(false);
  };

  const handleDisposeAsset = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAsset) return;

    const proceeds = Number(disposeForm.proceedsAmount);
    setAssets(
      assets.map((a) =>
        a.id === selectedAsset.id
          ? {
              ...a,
              status: "DISPOSED",
              currentNetBookValue: 0,
            }
          : a
      )
    );
    setShowDisposeModal(false);
  };

  return (
    <div className="p-4 sm:p-6 space-y-6 max-w-7xl mx-auto" dir={dir}>
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-blue-900 via-indigo-900 to-slate-900 p-6 sm:p-8 text-white shadow-xl">
        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/20 text-blue-300 border border-blue-400/30 text-xs font-semibold backdrop-blur-md">
              <Sparkles className="w-3.5 h-3.5 text-blue-400" />
              <span>{isAr ? "المرحلة 23 — محرك الأصول الثابتة والإهلاك" : "Phase 23 — Fixed Assets & Depreciation Engine"}</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">
              {isAr ? "دورة حياة الأصول الثابتة وإهلاكها الآلي" : "Fixed Assets Lifecycle & Depreciation"}
            </h1>
            <p className="text-blue-200 text-sm max-w-2xl leading-relaxed">
              {isAr
                ? "إدارة الأصول الثابتة من الاقتناء حتى الاستبعاد، وتوليد جداول الإهلاك بالقسط الثابت والترحيل الآلي لقيود اليومية لدفتر الأستاذ العام وفق المعايير المحاسبية (SOCPA / IFRS)."
                : "Manage tangible assets from acquisition to disposal, generate deterministic straight-line schedules, and post balanced GL journal entries (Expense 525000 / Accum 129000)."}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={() => setShowRunModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-indigo-600/80 hover:bg-indigo-600 text-white text-sm font-medium border border-indigo-400/30 backdrop-blur-md transition-all shadow-md active:scale-95"
            >
              <Calendar className="w-4 h-4 text-indigo-200" />
              <span>{isAr ? "تشغيل مسير استهلاك" : "Run Depreciation"}</span>
            </button>
            <button
              onClick={() => setShowRegisterModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-blue-500 hover:bg-blue-400 text-white text-sm font-semibold shadow-lg shadow-blue-500/25 transition-all active:scale-95"
            >
              <Plus className="w-4 h-4" />
              <span>{isAr ? "تسجيل أصل جديد" : "Register Asset"}</span>
            </button>
          </div>
        </div>

        {/* Decorative background glow */}
        <div className="absolute -top-24 -right-24 w-96 h-96 bg-blue-500/20 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute -bottom-24 -left-24 w-96 h-96 bg-indigo-500/20 rounded-full blur-3xl pointer-events-none" />
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Cost */}
        <div className="p-5 rounded-2xl bg-card border border-border/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "إجمالي التكلفة التاريخية" : "Total Historical Cost"}
            </span>
            <div className="p-2.5 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <Building className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {totalHistoricalCost.toLocaleString()} <span className="text-xs font-normal text-muted-foreground">{isAr ? "ر.س" : "SAR"}</span>
            </div>
            <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span className="text-blue-600 dark:text-blue-400 font-medium">100%</span>
              <span>{isAr ? "تكلفة الأساس الرأسمالي" : "Capitalized Asset Base"}</span>
            </div>
          </div>
        </div>

        {/* Net Book Value */}
        <div className="p-5 rounded-2xl bg-card border border-border/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "صافي القيمة الدفترية (NBV)" : "Net Book Value (NBV)"}
            </span>
            <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <DollarSign className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
              {totalNetBookValue.toLocaleString()} <span className="text-xs font-normal text-muted-foreground">{isAr ? "ر.س" : "SAR"}</span>
            </div>
            <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span>{isAr ? "القيمة الصافية بعد خصم المجمع" : "Cost less accumulated depr"}</span>
            </div>
          </div>
        </div>

        {/* Accumulated Depreciation */}
        <div className="p-5 rounded-2xl bg-card border border-border/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "مجمع الاستهلاك المتراكم" : "Total Accum. Depreciation"}
            </span>
            <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <TrendingDown className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold tracking-tight text-amber-600 dark:text-amber-400">
              {totalAccumulatedDepr.toLocaleString()} <span className="text-xs font-normal text-muted-foreground">{isAr ? "ر.س" : "SAR"}</span>
            </div>
            <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span className="text-amber-600 dark:text-amber-400 font-medium">
                {totalHistoricalCost > 0 ? ((totalAccumulatedDepr / totalHistoricalCost) * 100).toFixed(1) : 0}%
              </span>
              <span>{isAr ? "نسبة استهلاك الأصول" : "Depreciation Ratio"}</span>
            </div>
          </div>
        </div>

        {/* Active Asset Count */}
        <div className="p-5 rounded-2xl bg-card border border-border/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "الأصول النشطة بالخدمة" : "Active In-Service Assets"}
            </span>
            <div className="p-2.5 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <Layers className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {activeAssetsCount}{" "}
              <span className="text-xs font-normal text-muted-foreground">{isAr ? "أصل تشغيلي" : "Units"}</span>
            </div>
            <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span className="inline-block w-2 h-2 rounded-full bg-emerald-500" />
              <span>{assets.length} {isAr ? "إجمالي السجلات المسجلة" : "Total Registered"}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border pb-2">
        <div className="flex items-center gap-2 overflow-x-auto">
          <button
            onClick={() => setActiveTab("assets")}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              activeTab === "assets"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
            }`}
          >
            <Layers className="w-4 h-4" />
            <span>{isAr ? "سجل الأصول الثابتة" : "Asset Register"}</span>
            <span className="px-1.5 py-0.5 rounded-full text-xs bg-background/20">{assets.length}</span>
          </button>

          <button
            onClick={() => setActiveTab("runs")}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              activeTab === "runs"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
            }`}
          >
            <Calendar className="w-4 h-4" />
            <span>{isAr ? "مسيرات الاستهلاك والترحيل" : "Depreciation Runs & GL"}</span>
            <span className="px-1.5 py-0.5 rounded-full text-xs bg-background/20">{runs.length}</span>
          </button>

          <button
            onClick={() => setActiveTab("categories")}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              activeTab === "categories"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
            }`}
          >
            <BookOpen className="w-4 h-4" />
            <span>{isAr ? "التصنيفات ودليل الحسابات" : "Categories & COA"}</span>
          </button>

          <button
            onClick={() => setActiveTab("transfers")}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              activeTab === "transfers"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
            }`}
          >
            <ArrowLeftRight className="w-4 h-4" />
            <span>{isAr ? "المواقع والمناقلات والعهد" : "Custody & Transfers"}</span>
            <span className="px-1.5 py-0.5 rounded-full text-xs bg-background/20">{transfers.length}</span>
          </button>
        </div>
      </div>

      {/* TAB 1: ASSETS REGISTER */}
      {activeTab === "assets" && (
        <div className="space-y-4">
          {/* Filters and Search Bar */}
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 bg-card p-3.5 rounded-2xl border border-border/60">
            <div className="relative flex-1">
              <Search className="w-4 h-4 absolute top-1/2 -translate-y-1/2 start-3 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={isAr ? "البحث بالرقم التسلسلي، أو اسم الأصل..." : "Search by asset # or name..."}
                className="w-full ps-9 pe-4 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
              />
            </div>

            <div className="flex items-center gap-2">
              <select
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                aria-label={isAr ? "تصفية حسب التصنيف" : "Filter by Category"}
                className="px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl text-foreground focus:outline-none"
              >
                <option value="ALL">{isAr ? "كل التصنيفات" : "All Categories"}</option>
                <option value="BUILDINGS">{isAr ? "مباني وعقارات" : "Buildings"}</option>
                <option value="VEHICLES">{isAr ? "مركبات وشاحنات" : "Vehicles"}</option>
                <option value="IT_EQUIP">{isAr ? "أجهزة تقنية وخوادم" : "IT Equipment"}</option>
                <option value="OFFICE_FURN">{isAr ? "أثاث وتجهيزات" : "Furniture"}</option>
              </select>

              <select
                value={selectedStatus}
                onChange={(e) => setSelectedStatus(e.target.value)}
                aria-label={isAr ? "تصفية حسب الحالة" : "Filter by Status"}
                className="px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl text-foreground focus:outline-none"
              >
                <option value="ALL">{isAr ? "كل الحالات" : "All Statuses"}</option>
                <option value="ACTIVE">{isAr ? "نشط بالخدمة" : "Active"}</option>
                <option value="FULLY_DEPRECIATED">{isAr ? "مستهلك بالكامل" : "Fully Depreciated"}</option>
                <option value="DISPOSED">{isAr ? "مستبعد / مباع" : "Disposed"}</option>
              </select>
            </div>
          </div>

          {/* Assets Table */}
          <div className="overflow-x-auto rounded-2xl border border-border/60 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase border-b border-border">
                <tr>
                  <th className="px-4 py-3.5 text-start">{isAr ? "رمز الأصل" : "Asset #"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "اسم الأصل والتصنيف" : "Asset & Category"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "الموقع والعهدة" : "Location & Custodian"}</th>
                  <th className="px-4 py-3.5 text-end">{isAr ? "تكلفة الشراء" : "Cost (SAR)"}</th>
                  <th className="px-4 py-3.5 text-end">{isAr ? "مجمع الاستهلاك" : "Accum. Depr"}</th>
                  <th className="px-4 py-3.5 text-end">{isAr ? "القيمة الدفترية" : "NBV (SAR)"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الحالة" : "Status"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الإجراءات" : "Actions"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {filteredAssets.map((asset) => (
                  <tr key={asset.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-mono font-medium text-foreground whitespace-nowrap">
                      {asset.assetNumber}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-semibold text-foreground">{isAr ? asset.nameAr : asset.nameEn}</div>
                      <div className="text-xs text-muted-foreground flex items-center gap-1.5 mt-0.5">
                        <span className="px-1.5 py-0.5 rounded bg-muted text-[11px]">
                          {isAr ? asset.categoryNameAr : asset.categoryNameEn}
                        </span>
                        <span>•</span>
                        <span>{asset.usefulLifeMonths} {isAr ? "شهراً" : "mo"}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-xs font-medium text-foreground flex items-center gap-1">
                        <MapPin className="w-3.5 h-3.5 text-blue-500 shrink-0" />
                        <span>{isAr ? asset.locationNameAr : asset.locationNameEn}</span>
                      </div>
                      <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                        <User className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                        <span>{isAr ? asset.custodianNameAr || "عهدة الإدارة" : asset.custodianNameEn || "General Dept"}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-end font-medium text-foreground">
                      {asset.acquisitionCost.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-end font-medium text-amber-600 dark:text-amber-400">
                      {asset.accumulatedDepreciation.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-end font-bold text-emerald-600 dark:text-emerald-400">
                      {asset.currentNetBookValue.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      {asset.status === "ACTIVE" && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          <CheckCircle2 className="w-3 h-3" />
                          <span>{isAr ? "نشط" : "Active"}</span>
                        </span>
                      )}
                      {asset.status === "FULLY_DEPRECIATED" && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
                          <Clock className="w-3 h-3" />
                          <span>{isAr ? "مستهلك بالكامل" : "Fully Depreciated"}</span>
                        </span>
                      )}
                      {asset.status === "DISPOSED" && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-slate-500/10 text-slate-500 border border-slate-500/20">
                          <Trash2 className="w-3 h-3" />
                          <span>{isAr ? "مستبعد / مباع" : "Disposed"}</span>
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      <div className="flex items-center justify-center gap-1.5">
                        <button
                          onClick={() => {
                            setSelectedAsset(asset);
                            setShowScheduleModal(true);
                          }}
                          title={isAr ? "عرض جدول الاستهلاك" : "View Depreciation Schedule"}
                          className="p-1.5 rounded-lg text-muted-foreground hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950/50 transition-colors"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        {asset.status === "ACTIVE" && (
                          <>
                            <button
                              onClick={() => {
                                setSelectedAsset(asset);
                                setShowTransferModal(true);
                              }}
                              title={isAr ? "مناقلة الأصل / تغيير العهدة" : "Transfer Asset"}
                              className="p-1.5 rounded-lg text-muted-foreground hover:text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-950/50 transition-colors"
                            >
                              <ArrowLeftRight className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => {
                                setSelectedAsset(asset);
                                setDisposeForm({
                                  disposalType: "SALE",
                                  proceedsAmount: asset.currentNetBookValue * 1.1,
                                  reason: "استبعاد أو بيع الأصل",
                                });
                                setShowDisposeModal(true);
                              }}
                              title={isAr ? "استبعاد أو بيع الأصل" : "Dispose / Sell Asset"}
                              className="p-1.5 rounded-lg text-muted-foreground hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/50 transition-colors"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: DEPRECIATION RUNS & GL POSTING */}
      {activeTab === "runs" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between bg-card p-4 rounded-2xl border border-border/60">
            <div>
              <h2 className="font-bold text-foreground">{isAr ? "مسيرات الإهلاك الشهرية والترحيل لدفتر الأستاذ" : "Depreciation Runs & GL Integration"}</h2>
              <p className="text-xs text-muted-foreground mt-0.5">
                {isAr
                  ? "يتم توليد قيد اليومية المتوازن آلياً: مدين 525000 (مصروف الاستهلاك) ودائن 129000 (مجمع الاستهلاك)"
                  : "Generates balanced journal entries debiting 525000 (Expense) and crediting 129000 (Accumulated Depr)."}
              </p>
            </div>
            <button
              onClick={() => setShowRunModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-semibold shadow-sm hover:opacity-90"
            >
              <Calendar className="w-4 h-4" />
              <span>{isAr ? "تشغيل مسير جديد" : "New Depreciation Run"}</span>
            </button>
          </div>

          <div className="overflow-x-auto rounded-2xl border border-border/60 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase border-b border-border">
                <tr>
                  <th className="px-4 py-3.5 text-start">{isAr ? "رقم المسير" : "Run #"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "الفترة المالية" : "Period"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "دفتر الإهلاك" : "Depreciation Book"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الأصول المعالجة" : "Assets"}</th>
                  <th className="px-4 py-3.5 text-end">{isAr ? "إجمالي الاستهلاك" : "Total Depr (SAR)"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الحالة" : "Status"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "قيد اليومية" : "Journal Entry"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الإجراءات" : "Action"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {runs.map((run) => (
                  <tr key={run.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-mono font-medium text-foreground">{run.runNumber}</td>
                    <td className="px-4 py-3 font-medium text-foreground">
                      {run.periodYear} / {String(run.periodMonth).padStart(2, "0")}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{run.bookNameAr}</td>
                    <td className="px-4 py-3 text-center font-medium text-foreground">{run.totalAssets}</td>
                    <td className="px-4 py-3 text-end font-bold text-amber-600 dark:text-amber-400">
                      {run.totalDepreciation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </td>
                    <td className="px-4 py-3 text-center">
                      {run.status === "POSTED" ? (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          <CheckCircle2 className="w-3 h-3" />
                          <span>{isAr ? "مرحل محاسبياً" : "POSTED to GL"}</span>
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20">
                          <Clock className="w-3 h-3" />
                          <span>{isAr ? "محسوب (بانتظار الاعتماد)" : "Calculated"}</span>
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center font-mono text-xs">
                      {run.journalEntryNumber ? (
                        <button
                          onClick={() => {
                            setSelectedRun(run);
                            setShowGLEntryModal(true);
                          }}
                          className="text-blue-600 dark:text-blue-400 hover:underline font-semibold"
                        >
                          {run.journalEntryNumber}
                        </button>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center">
                      {run.status === "CALCULATED" ? (
                        <button
                          onClick={() => handlePostRun(run.id)}
                          className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-sm transition-all"
                        >
                          {isAr ? "ترحيل لدفتر الأستاذ" : "Post to GL"}
                        </button>
                      ) : (
                        <button
                          onClick={() => {
                            setSelectedRun(run);
                            setShowGLEntryModal(true);
                          }}
                          className="px-2.5 py-1 rounded-lg bg-muted hover:bg-muted/80 text-foreground text-xs font-medium"
                        >
                          {isAr ? "عرض القيد" : "View JV"}
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: CATEGORIES & CHART OF ACCOUNTS */}
      {activeTab === "categories" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {categories.map((cat) => (
              <div key={cat.id} className="p-5 rounded-2xl bg-card border border-border/60 shadow-sm space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-3 rounded-xl bg-primary/10 text-primary">
                      {cat.code === "BUILDINGS" && <Building className="w-6 h-6" />}
                      {cat.code === "VEHICLES" && <Truck className="w-6 h-6" />}
                      {cat.code === "IT_EQUIP" && <Monitor className="w-6 h-6" />}
                      {cat.code === "OFFICE_FURN" && <Armchair className="w-6 h-6" />}
                    </div>
                    <div>
                      <h3 className="font-bold text-foreground">{isAr ? cat.nameAr : cat.nameEn}</h3>
                      <div className="text-xs font-mono text-muted-foreground mt-0.5">{cat.code}</div>
                    </div>
                  </div>
                  <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-foreground">
                    {cat.usefulLifeMonths} {isAr ? "شهراً" : "Months"} ({cat.usefulLifeMonths / 12} {isAr ? "سنوات" : "Yrs"})
                  </span>
                </div>

                <div className="text-xs space-y-2 pt-2 border-t border-border">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">{isAr ? "طريقة الإهلاك المعتمدة:" : "Depreciation Method:"}</span>
                    <span className="font-semibold text-foreground">{cat.depreciationMethod}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">{isAr ? "حساب الأصل (الميزانية):" : "Asset Account:"}</span>
                    <span className="font-mono text-blue-600 dark:text-blue-400">{cat.assetAccountCode} - {cat.assetAccountNameAr}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">{isAr ? "مجمع الاستهلاك (دائن):" : "Accum. Depr Account:"}</span>
                    <span className="font-mono text-amber-600 dark:text-amber-400">{cat.accumAccountCode} - {cat.accumAccountNameAr}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">{isAr ? "مصروف الاستهلاك (مدين):" : "Expense Account:"}</span>
                    <span className="font-mono text-purple-600 dark:text-purple-400">{cat.expenseAccountCode} - {cat.expenseAccountNameAr}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 4: CUSTODY & TRANSFERS */}
      {activeTab === "transfers" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between bg-card p-4 rounded-2xl border border-border/60">
            <div>
              <h2 className="font-bold text-foreground">{isAr ? "سجل المناقلات وحركة العهد" : "Asset Custody & Transfers History"}</h2>
              <p className="text-xs text-muted-foreground mt-0.5">
                {isAr
                  ? "توثيق حركة انتقال الأصول الثابتة بين الفروع والمستودعات والموظفين المستلمين للعهدة"
                  : "Tracks location movements and employee custody handovers across branches."}
              </p>
            </div>
          </div>

          <div className="overflow-x-auto rounded-2xl border border-border/60 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase border-b border-border">
                <tr>
                  <th className="px-4 py-3.5 text-start">{isAr ? "رقم المناقلة" : "Transfer #"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "الأصل المنقول" : "Asset"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "الموقع السابق" : "From Location"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "الموقع الجديد" : "To Location"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "المستلم (العهدة)" : "Custodian"}</th>
                  <th className="px-4 py-3.5 text-start">{isAr ? "التاريخ والسبب" : "Date & Reason"}</th>
                  <th className="px-4 py-3.5 text-center">{isAr ? "الحالة" : "Status"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {transfers.map((trf) => (
                  <tr key={trf.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-mono font-medium text-foreground">{trf.transferNumber}</td>
                    <td className="px-4 py-3">
                      <div className="font-semibold text-foreground">{trf.assetNameAr}</div>
                      <div className="text-xs font-mono text-muted-foreground">{trf.assetNumber}</div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{trf.fromLocationNameAr}</td>
                    <td className="px-4 py-3 font-medium text-blue-600 dark:text-blue-400">{trf.toLocationNameAr}</td>
                    <td className="px-4 py-3 font-medium text-foreground">{trf.custodianNameAr}</td>
                    <td className="px-4 py-3">
                      <div className="text-xs text-foreground">{trf.transferDate}</div>
                      <div className="text-xs text-muted-foreground">{trf.reason}</div>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                        <CheckCircle2 className="w-3 h-3" />
                        <span>{isAr ? "مكتمل" : "Completed"}</span>
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* MODAL 1: REGISTER NEW ASSET */}
      {showRegisterModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-xl rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-blue-500/10 text-blue-600">
                  <Plus className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "تسجيل أصل ثابت جديد" : "Register New Fixed Asset"}</h3>
                  <p className="text-xs text-muted-foreground">{isAr ? "إدخال بيانات الأصل وتحديد العمر الإنتاجي وقيمة التخريد" : "Enter asset acquisition details and depreciation parameters"}</p>
                </div>
              </div>
              <button onClick={() => setShowRegisterModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleRegisterAsset} className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "رمز الأصل" : "Asset Number"}</label>
                  <input
                    type="text"
                    required
                    value={newAssetForm.assetNumber}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, assetNumber: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl font-mono focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "تصنيف الأصل" : "Category"}</label>
                  <select
                    value={newAssetForm.categoryCode}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, categoryCode: e.target.value })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                  >
                    <option value="IT_EQUIP">{isAr ? "أجهزة وخوادم تقنية (36 شهراً)" : "IT Equipment (36 mo)"}</option>
                    <option value="VEHICLES">{isAr ? "أسطول النقل والسيارات (60 شهراً)" : "Vehicles (60 mo)"}</option>
                    <option value="BUILDINGS">{isAr ? "المباني والإنشاءات (240 شهراً)" : "Buildings (240 mo)"}</option>
                    <option value="OFFICE_FURN">{isAr ? "الأثاث والتجهيزات المكتبية (48 شهراً)" : "Furniture (48 mo)"}</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "اسم الأصل بالعربية" : "Name (Arabic)"}</label>
                  <input
                    type="text"
                    required
                    value={newAssetForm.nameAr}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, nameAr: e.target.value })}
                    placeholder="مثال: لابتوب ماك بوك برو M3"
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "اسم الأصل بالإنجليزية" : "Name (English)"}</label>
                  <input
                    type="text"
                    required
                    value={newAssetForm.nameEn}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, nameEn: e.target.value })}
                    placeholder="e.g. MacBook Pro M3 Laptop"
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "تكلفة الشراء (ر.س)" : "Cost (SAR)"}</label>
                  <input
                    type="number"
                    required
                    min="1"
                    value={newAssetForm.acquisitionCost}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, acquisitionCost: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 font-medium"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "قيمة الخردة المتبقية (ر.س)" : "Salvage Floor"}</label>
                  <input
                    type="number"
                    min="0"
                    value={newAssetForm.salvageValue}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, salvageValue: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "العمر الإنتاجي (شهراً)" : "Life (Months)"}</label>
                  <input
                    type="number"
                    required
                    min="1"
                    value={newAssetForm.usefulLifeMonths}
                    onChange={(e) => setNewAssetForm({ ...newAssetForm, usefulLifeMonths: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
              </div>

              <div className="p-3.5 rounded-xl bg-muted/50 border border-border/80 flex items-center justify-between text-xs">
                <span className="text-muted-foreground">{isAr ? "قسط الإهلاك الشهري التقديري (القسط الثابت):" : "Projected Monthly Depreciation:"}</span>
                <span className="font-bold text-amber-600 dark:text-amber-400 font-mono text-sm">
                  {newAssetForm.usefulLifeMonths > 0
                    ? ((newAssetForm.acquisitionCost - newAssetForm.salvageValue) / newAssetForm.usefulLifeMonths).toFixed(2)
                    : 0}{" "}
                  {isAr ? "ر.س / شهر" : "SAR/mo"}
                </span>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowRegisterModal(false)}
                  className="px-4 py-2 rounded-xl text-sm font-medium text-muted-foreground hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 rounded-xl text-sm font-semibold bg-primary text-primary-foreground shadow hover:opacity-90 transition-all"
                >
                  {isAr ? "حفظ وتوليد الجدول" : "Save & Generate Schedule"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: RUN DEPRECIATION */}
      {showRunModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-600">
                  <Calendar className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "تشغيل مسير استهلاك شهري" : "Run Monthly Depreciation"}</h3>
                  <p className="text-xs text-muted-foreground">{isAr ? "احتساب أقساط الإهلاك لكافة الأصول النشطة" : "Calculate monthly straight-line depreciation for all active assets"}</p>
                </div>
              </div>
              <button onClick={() => setShowRunModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateRun} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "السنة المالية" : "Year"}</label>
                  <input
                    type="number"
                    value={newRunForm.periodYear}
                    onChange={(e) => setNewRunForm({ ...newRunForm, periodYear: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl font-mono focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "الشهر" : "Month"}</label>
                  <select
                    value={newRunForm.periodMonth}
                    onChange={(e) => setNewRunForm({ ...newRunForm, periodMonth: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                  >
                    {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12].map((m) => (
                      <option key={m} value={m}>
                        {m} - {new Date(2026, m - 1, 1).toLocaleString(isAr ? "ar-SA" : "en-US", { month: "long" })}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "دفتر الإهلاك" : "Depreciation Book"}</label>
                <select
                  value={newRunForm.bookCode}
                  onChange={(e) => setNewRunForm({ ...newRunForm, bookCode: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                >
                  <option value="BOOK-FIN">{isAr ? "الدفتر المالي القانوني (IFRS/SOCPA)" : "Statutory Financial Book"}</option>
                  <option value="BOOK-TAX">{isAr ? "الدفتر الضريبي والزكوي (ZATCA)" : "Tax & Zakat Book"}</option>
                </select>
              </div>

              <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs space-y-1.5">
                <div className="flex items-center gap-1.5 text-amber-700 dark:text-amber-400 font-semibold">
                  <AlertCircle className="w-4 h-4" />
                  <span>{isAr ? "ملاحظة تدقيق هامة" : "Audit Notice"}</span>
                </div>
                <p className="text-muted-foreground leading-relaxed">
                  {isAr
                    ? "سيتم احتساب القسط الشهري فقط للأصول التي لم تصل لقيمة الخردة المحددة (Floor Salvage Value). وعند اعتماد المسير، سيتم ترحيل القيد لدفتر الأستاذ فوراً."
                    : "Assets at or below salvage value will not be depreciated. GL Journal entries will be created upon posting."}
                </p>
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowRunModal(false)}
                  className="px-4 py-2 rounded-xl text-sm font-medium text-muted-foreground hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 rounded-xl text-sm font-semibold bg-indigo-600 hover:bg-indigo-500 text-white shadow transition-all"
                >
                  {isAr ? "بدء الاحتساب" : "Execute Calculation"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 3: ASSET TRANSFER */}
      {showTransferModal && selectedAsset && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-purple-500/10 text-purple-600">
                  <ArrowLeftRight className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "مناقلة أصل / تسليم عهدة" : "Transfer Asset / Custody"}</h3>
                  <p className="text-xs text-muted-foreground">{selectedAsset.assetNumber} - {isAr ? selectedAsset.nameAr : selectedAsset.nameEn}</p>
                </div>
              </div>
              <button onClick={() => setShowTransferModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleTransferAsset} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "الموقع الجديد" : "To Location"}</label>
                <select
                  value={transferForm.toLocationCode}
                  onChange={(e) => setTransferForm({ ...transferForm, toLocationCode: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                >
                  <option value="HQ-RUH">{isAr ? "المقر الرئيسي - الرياض" : "HQ - Riyadh"}</option>
                  <option value="BR-JED">{isAr ? "فرع الأندلس - جدة" : "Al-Andalus Branch - Jeddah"}</option>
                  <option value="WH-SUL">{isAr ? "مستودع السلي اللوجستي - الرياض" : "Al-Suly Logistics Warehouse"}</option>
                  <option value="BR-DAM">{isAr ? "فرع الدمام الشرقي" : "Dammam East Branch"}</option>
                </select>
              </div>

              <div>
                <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "المستلم الجديد (الموظف المسؤول عن العهدة)" : "New Custodian Employee"}</label>
                <input
                  type="text"
                  required
                  value={transferForm.custodianNameAr}
                  onChange={(e) => setTransferForm({ ...transferForm, custodianNameAr: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "سبب المناقلة" : "Reason / Purpose"}</label>
                <textarea
                  rows={3}
                  value={transferForm.reason}
                  onChange={(e) => setTransferForm({ ...transferForm, reason: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowTransferModal(false)}
                  className="px-4 py-2 rounded-xl text-sm font-medium text-muted-foreground hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 rounded-xl text-sm font-semibold bg-purple-600 hover:bg-purple-500 text-white shadow transition-all"
                >
                  {isAr ? "تأكيد المناقلة" : "Confirm Transfer"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 4: ASSET DISPOSAL / SALE */}
      {showDisposeModal && selectedAsset && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-rose-500/10 text-rose-600">
                  <Trash2 className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "استبعاد أو بيع الأصل الثابت" : "Dispose / Sell Fixed Asset"}</h3>
                  <p className="text-xs text-muted-foreground">{selectedAsset.assetNumber} - {isAr ? selectedAsset.nameAr : selectedAsset.nameEn}</p>
                </div>
              </div>
              <button onClick={() => setShowDisposeModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleDisposeAsset} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "نوع الاستبعاد" : "Disposal Type"}</label>
                  <select
                    value={disposeForm.disposalType}
                    onChange={(e) => setDisposeForm({ ...disposeForm, disposalType: e.target.value as "SALE" | "SCRAP" })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                  >
                    <option value="SALE">{isAr ? "بيع الأصل (Sale)" : "Sale"}</option>
                    <option value="SCRAP">{isAr ? "تخريد / إتلاف (Scrap)" : "Scrap"}</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "مبلغ المتحصلات (ر.س)" : "Proceeds (SAR)"}</label>
                  <input
                    type="number"
                    min="0"
                    value={disposeForm.proceedsAmount}
                    onChange={(e) => setDisposeForm({ ...disposeForm, proceedsAmount: Number(e.target.value) })}
                    className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl font-medium focus:outline-none"
                  />
                </div>
              </div>

              {/* Live Gain/Loss preview */}
              <div className="p-3.5 rounded-xl bg-muted/50 border border-border space-y-2 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">{isAr ? "صافي القيمة الدفترية الحالية (NBV):" : "Current Net Book Value:"}</span>
                  <span className="font-semibold text-foreground font-mono">{selectedAsset.currentNetBookValue.toLocaleString()} {isAr ? "ر.س" : "SAR"}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">{isAr ? "الأرباح / (الخسائر) الرأسمالية الناتجة:" : "Gain / (Loss) on Disposal:"}</span>
                  <span
                    className={`font-bold font-mono text-sm ${
                      disposeForm.proceedsAmount - selectedAsset.currentNetBookValue >= 0
                        ? "text-emerald-600 dark:text-emerald-400"
                        : "text-rose-600 dark:text-rose-400"
                    }`}
                  >
                    {(disposeForm.proceedsAmount - selectedAsset.currentNetBookValue).toLocaleString()}{" "}
                    {isAr ? "ر.س" : "SAR"}
                  </span>
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-foreground block mb-1">{isAr ? "سبب الاستبعاد" : "Disposal Reason"}</label>
                <textarea
                  rows={2}
                  value={disposeForm.reason}
                  onChange={(e) => setDisposeForm({ ...disposeForm, reason: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-muted/40 border border-border rounded-xl focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowDisposeModal(false)}
                  className="px-4 py-2 rounded-xl text-sm font-medium text-muted-foreground hover:bg-muted"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 rounded-xl text-sm font-semibold bg-rose-600 hover:bg-rose-500 text-white shadow transition-all"
                >
                  {isAr ? "تأكيد الاستبعاد وتصفير الأصل" : "Confirm Disposal"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 5: DEPRECIATION SCHEDULE PROJECTION */}
      {showScheduleModal && selectedAsset && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-2xl rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] flex flex-col">
            <div className="flex items-center justify-between border-b border-border pb-3 shrink-0">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-blue-500/10 text-blue-600">
                  <FileSpreadsheet className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "جدول استهلاك الأصل المتوقع (القسط الثابت)" : "Projected Depreciation Schedule"}</h3>
                  <p className="text-xs text-muted-foreground">{selectedAsset.assetNumber} - {isAr ? selectedAsset.nameAr : selectedAsset.nameEn}</p>
                </div>
              </div>
              <button onClick={() => setShowScheduleModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Asset Summary info */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-muted/40 p-3.5 rounded-xl text-xs shrink-0">
              <div>
                <span className="text-muted-foreground block">{isAr ? "تكلفة الشراء:" : "Cost:"}</span>
                <span className="font-bold text-foreground font-mono">{selectedAsset.acquisitionCost.toLocaleString()} {isAr ? "ر.س" : "SAR"}</span>
              </div>
              <div>
                <span className="text-muted-foreground block">{isAr ? "قيمة الخردة:" : "Salvage Floor:"}</span>
                <span className="font-bold text-foreground font-mono">{selectedAsset.salvageValue.toLocaleString()} {isAr ? "ر.س" : "SAR"}</span>
              </div>
              <div>
                <span className="text-muted-foreground block">{isAr ? "القسط الشهري:" : "Monthly Depr:"}</span>
                <span className="font-bold text-amber-600 dark:text-amber-400 font-mono">
                  {((selectedAsset.acquisitionCost - selectedAsset.salvageValue) / selectedAsset.usefulLifeMonths).toFixed(2)}
                </span>
              </div>
              <div>
                <span className="text-muted-foreground block">{isAr ? "العمر الإنتاجي:" : "Useful Life:"}</span>
                <span className="font-bold text-foreground">{selectedAsset.usefulLifeMonths} {isAr ? "شهراً" : "Months"}</span>
              </div>
            </div>

            {/* Schedule Table */}
            <div className="overflow-y-auto flex-1 border border-border rounded-xl">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/80 sticky top-0 border-b border-border text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2 text-center">#</th>
                    <th className="px-3 py-2 text-start">{isAr ? "الفترة المالية" : "Period"}</th>
                    <th className="px-3 py-2 text-end">{isAr ? "القيمة أول المدة" : "Beginning NBV"}</th>
                    <th className="px-3 py-2 text-end">{isAr ? "قسط الاستهلاك" : "Depreciation"}</th>
                    <th className="px-3 py-2 text-end">{isAr ? "مجمع الاستهلاك" : "Accumulated"}</th>
                    <th className="px-3 py-2 text-end">{isAr ? "صافي القيمة آخر المدة" : "Ending NBV"}</th>
                    <th className="px-3 py-2 text-center">{isAr ? "الحالة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {Array.from({ length: Math.min(selectedAsset.usefulLifeMonths, 24) }).map((_, idx) => {
                    const monthNumber = idx + 1;
                    const monthly = (selectedAsset.acquisitionCost - selectedAsset.salvageValue) / selectedAsset.usefulLifeMonths;
                    const accum = Math.min(monthly * monthNumber, selectedAsset.acquisitionCost - selectedAsset.salvageValue);
                    const begNBV = selectedAsset.acquisitionCost - monthly * (monthNumber - 1);
                    const endNBV = selectedAsset.acquisitionCost - accum;
                    const isPosted = accum <= selectedAsset.accumulatedDepreciation;

                    return (
                      <tr key={idx} className={isPosted ? "bg-emerald-500/5" : ""}>
                        <td className="px-3 py-2 text-center font-mono text-muted-foreground">{monthNumber}</td>
                        <td className="px-3 py-2 font-mono">
                          {2024 + Math.floor((idx + 1) / 12)} / {String(((idx + 1) % 12) + 1).padStart(2, "0")}
                        </td>
                        <td className="px-3 py-2 text-end font-mono">{begNBV.toFixed(2)}</td>
                        <td className="px-3 py-2 text-end font-mono text-amber-600 dark:text-amber-400 font-semibold">
                          {monthly.toFixed(2)}
                        </td>
                        <td className="px-3 py-2 text-end font-mono">{accum.toFixed(2)}</td>
                        <td className="px-3 py-2 text-end font-mono font-bold text-foreground">{endNBV.toFixed(2)}</td>
                        <td className="px-3 py-2 text-center">
                          {isPosted ? (
                            <span className="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-600 text-[10px] font-semibold">
                              {isAr ? "مرحل" : "POSTED"}
                            </span>
                          ) : (
                            <span className="px-1.5 py-0.5 rounded bg-muted text-muted-foreground text-[10px]">
                              {isAr ? "متوقع" : "PROJECTED"}
                            </span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            <div className="flex items-center justify-end shrink-0 pt-2">
              <button
                onClick={() => setShowScheduleModal(false)}
                className="px-4 py-2 rounded-xl text-sm font-medium bg-muted hover:bg-muted/80 text-foreground"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 6: GL JOURNAL ENTRY DETAIL */}
      {showGLEntryModal && selectedRun && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-xl rounded-2xl bg-card border border-border p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-600">
                  <BookOpen className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-lg text-foreground">{isAr ? "تفاصيل قيد اليومية المحاسبي لدفتر الأستاذ" : "General Ledger Journal Entry"}</h3>
                  <p className="text-xs text-muted-foreground font-mono">{selectedRun.journalEntryNumber || "JV-DRAFT"}</p>
                </div>
              </div>
              <button onClick={() => setShowGLEntryModal(false)} className="p-1 rounded-lg text-muted-foreground hover:bg-muted">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between text-xs bg-muted/40 p-3 rounded-xl">
                <div>
                  <span className="text-muted-foreground">{isAr ? "الفترة المحاسبية: " : "Period: "}</span>
                  <span className="font-bold text-foreground">{selectedRun.periodYear} / {selectedRun.periodMonth}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">{isAr ? "المرجع: " : "Ref: "}</span>
                  <span className="font-mono text-foreground">{selectedRun.runNumber}</span>
                </div>
              </div>

              {/* Balanced Entry Table */}
              <div className="border border-border rounded-xl overflow-hidden">
                <table className="w-full text-xs">
                  <thead className="bg-muted text-muted-foreground border-b border-border">
                    <tr>
                      <th className="px-3 py-2 text-start">{isAr ? "رقم الحساب" : "Account #"}</th>
                      <th className="px-3 py-2 text-start">{isAr ? "اسم الحساب والتفصيل" : "Account Name"}</th>
                      <th className="px-3 py-2 text-end">{isAr ? "مدين (SAR)" : "Debit"}</th>
                      <th className="px-3 py-2 text-end">{isAr ? "دائن (SAR)" : "Credit"}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/60">
                    <tr>
                      <td className="px-3 py-2.5 font-mono text-blue-600 dark:text-blue-400 font-medium">525000</td>
                      <td className="px-3 py-2.5">
                        <div className="font-semibold text-foreground">{isAr ? "مصروف استهلاك الأصول الثابتة" : "Depreciation Expense"}</div>
                        <div className="text-[11px] text-muted-foreground">{isAr ? "إثبات استهلاك شهر " : "Depreciation of month "}{selectedRun.periodMonth}</div>
                      </td>
                      <td className="px-3 py-2.5 text-end font-mono font-bold text-foreground">
                        {selectedRun.totalDepreciation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </td>
                      <td className="px-3 py-2.5 text-end font-mono text-muted-foreground">0.00</td>
                    </tr>
                    <tr>
                      <td className="px-3 py-2.5 font-mono text-amber-600 dark:text-amber-400 font-medium">129000</td>
                      <td className="px-3 py-2.5">
                        <div className="font-semibold text-foreground">{isAr ? "مجمع استهلاك الأصول الثابتة" : "Accumulated Depreciation"}</div>
                        <div className="text-[11px] text-muted-foreground">{isAr ? "حساب مقابل الأصول الثابتة (Contra Asset)" : "Contra Asset Account"}</div>
                      </td>
                      <td className="px-3 py-2.5 text-end font-mono text-muted-foreground">0.00</td>
                      <td className="px-3 py-2.5 text-end font-mono font-bold text-foreground">
                        {selectedRun.totalDepreciation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </td>
                    </tr>
                  </tbody>
                  <tfoot className="bg-muted/60 border-t border-border font-bold">
                    <tr>
                      <td colSpan={2} className="px-3 py-2 text-start text-muted-foreground">{isAr ? "الإجمالي المتوازن:" : "Balanced Total:"}</td>
                      <td className="px-3 py-2 text-end font-mono text-emerald-600 dark:text-emerald-400">
                        {selectedRun.totalDepreciation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </td>
                      <td className="px-3 py-2 text-end font-mono text-emerald-600 dark:text-emerald-400">
                        {selectedRun.totalDepreciation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>

            <div className="flex items-center justify-end pt-2">
              <button
                onClick={() => setShowGLEntryModal(false)}
                className="px-4 py-2 rounded-xl text-sm font-medium bg-muted hover:bg-muted/80 text-foreground"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

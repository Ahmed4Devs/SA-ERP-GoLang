"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Package,
  Barcode,
  FolderTree,
  Scale,
  Plus,
  Search,
  CheckCircle2,
  AlertCircle,
  Tag,
  Layers,
  ArrowRightLeft,
  X,
  Sparkles,
  DollarSign,
  ShieldCheck,
  Smartphone,
  Cpu,
  Boxes,
} from "lucide-react";

interface Item {
  id: string;
  code: string;
  name_ar: string;
  name_en: string;
  category: string;
  item_type: "STOCK" | "SERVICE" | "NON_STOCK" | "ASSET";
  tracking_type: "NONE" | "LOT" | "SERIAL";
  base_uom: string;
  cost_price: number;
  list_price: number;
  barcode: string;
  tax_code: string;
  is_active: boolean;
}

const INITIAL_ITEMS: Item[] = [
  {
    id: "itm-001",
    code: "IT-LAPTOP-01",
    name_ar: "جهاز حاسب محمول للأعمال 16 بوصة",
    name_en: "Enterprise Business Laptop 16-inch",
    category: "IT & Computing Hardware",
    item_type: "STOCK",
    tracking_type: "SERIAL",
    base_uom: "PCS",
    cost_price: 3800.0,
    list_price: 4800.0,
    barcode: "6281000123456",
    tax_code: "VAT-STD-15 (15%)",
    is_active: true,
  },
  {
    id: "itm-002",
    code: "SRV-CONSULT-01",
    name_ar: "استشارات هندسة النظم السحابية",
    name_en: "Cloud Architecture Advisory",
    category: "Professional Services",
    item_type: "SERVICE",
    tracking_type: "NONE",
    base_uom: "HOUR",
    cost_price: 250.0,
    list_price: 600.0,
    barcode: "6282000654321",
    tax_code: "VAT-STD-15 (15%)",
    is_active: true,
  },
  {
    id: "itm-003",
    code: "PKG-PAPER-50",
    name_ar: "ورق فواتير حراري (كرتون 50 لفة)",
    name_en: "Thermal POS Paper Rolls Box",
    category: "General Supplies",
    item_type: "STOCK",
    tracking_type: "LOT",
    base_uom: "BOX",
    cost_price: 65.0,
    list_price: 110.0,
    barcode: "6283000987654",
    tax_code: "VAT-STD-15 (15%)",
    is_active: true,
  },
  {
    id: "itm-004",
    code: "NET-SWITCH-24",
    name_ar: "موزع شبكي 24 منفذ مُدار",
    name_en: "24-Port Managed Gigabit Switch",
    category: "IT & Computing Hardware",
    item_type: "ASSET",
    tracking_type: "SERIAL",
    base_uom: "PCS",
    cost_price: 1200.0,
    list_price: 1750.0,
    barcode: "6284000112233",
    tax_code: "VAT-STD-15 (15%)",
    is_active: true,
  },
];

export function ProductView() {
  const { t, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"items" | "scanner" | "categories" | "uoms">("items");
  const [searchQuery, setSearchQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("ALL");
  const [items, setItems] = useState<Item[]>(INITIAL_ITEMS);

  // Scanner Simulator state
  const [scanCode, setScanCode] = useState("");
  const [matchedItem, setMatchedItem] = useState<Item | null>(null);
  const [scanError, setScanError] = useState("");

  // UOM Converter state
  const [convertFrom, setConvertFrom] = useState("KG");
  const [convertTo, setConvertTo] = useState("G");
  const [convertQty, setConvertQty] = useState<number>(2.5);
  const [convertResult, setConvertResult] = useState<number | null>(2500);

  // Modal State
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newCode, setNewCode] = useState("");
  const [newNameAr, setNewNameAr] = useState("");
  const [newNameEn, setNewNameEn] = useState("");
  const [newType, setNewType] = useState<"STOCK" | "SERVICE" | "NON_STOCK" | "ASSET">("STOCK");
  const [newTracking, setNewTracking] = useState<"NONE" | "LOT" | "SERIAL">("NONE");
  const [newBaseUOM, setNewBaseUOM] = useState("PCS");
  const [newCost, setNewCost] = useState(100);
  const [newList, setNewList] = useState(150);
  const [newBarcode, setNewBarcode] = useState("");

  // Filter items
  const filteredItems = items.filter((it) => {
    const matchesType = typeFilter === "ALL" || it.item_type === typeFilter;
    const q = searchQuery.toLowerCase();
    const matchesSearch =
      !searchQuery ||
      it.code.toLowerCase().includes(q) ||
      it.name_ar.includes(q) ||
      it.name_en.toLowerCase().includes(q) ||
      it.barcode.includes(q);
    return matchesType && matchesSearch;
  });

  // Handle barcode lookup
  const handleScan = () => {
    setScanError("");
    setMatchedItem(null);
    if (!scanCode.trim()) return;

    const found = items.find((i) => i.barcode === scanCode.trim() || i.code.toLowerCase() === scanCode.trim().toLowerCase());
    if (found) {
      setMatchedItem(found);
    } else {
      setScanError(dir === "rtl" ? "لم يتم العثور على صنف مطابق لهذا الباركود" : "No product found matching this barcode");
    }
  };

  // Handle UOM conversion
  const handleConvert = () => {
    let rate = 1;
    if (convertFrom === "KG" && convertTo === "G") rate = 1000;
    else if (convertFrom === "G" && convertTo === "KG") rate = 0.001;
    else if (convertFrom === "BOX" && convertTo === "PCS") rate = 10;
    else if (convertFrom === "PCS" && convertTo === "BOX") rate = 0.1;
    else if (convertFrom === "M" && convertTo === "CM") rate = 100;
    else if (convertFrom === "CM" && convertTo === "M") rate = 0.01;
    else if (convertFrom === convertTo) rate = 1;
    else {
      setConvertResult(null);
      return;
    }
    setConvertResult(Number((convertQty * rate).toFixed(4)));
  };

  // Handle Add Item
  const handleCreateItem = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCode || !newNameAr || !newNameEn) return;

    const item: Item = {
      id: "itm-" + Date.now(),
      code: newCode.toUpperCase(),
      name_ar: newNameAr,
      name_en: newNameEn,
      category: "General Hardware",
      item_type: newType,
      tracking_type: newTracking,
      base_uom: newBaseUOM,
      cost_price: Number(newCost),
      list_price: Number(newList),
      barcode: newBarcode || "628" + Math.floor(1000000000 + Math.random() * 9000000000),
      tax_code: "VAT-STD-15 (15%)",
      is_active: true,
    };

    setItems([item, ...items]);
    setIsCreateOpen(false);
    // Reset
    setNewCode("");
    setNewNameAr("");
    setNewNameEn("");
    setNewBarcode("");
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 p-6 rounded-2xl bg-card border shadow-sm relative overflow-hidden">
        <div className="space-y-1.5 z-10">
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20 flex items-center gap-1">
              <Sparkles className="h-3 w-3" />
              {dir === "rtl" ? "المرحلة 11: دليل المنتجات والأصناف" : "Phase 11: Products Master"}
            </span>
            <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
              ZATCA Tax Linked
            </span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.products.title}
          </h1>
          <p className="text-sm text-muted-foreground max-w-2xl">
            {t.products.subtitle}
          </p>
        </div>

        <div className="flex items-center gap-3 z-10">
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition-all shadow-sm"
          >
            <Plus className="h-4 w-4" />
            {t.products.create_item}
          </button>
        </div>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-xl border bg-card/60 backdrop-blur shadow-sm space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-medium">{t.products.total_items}</span>
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Package className="h-4 w-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">{items.length}</span>
            <span className="text-xs text-muted-foreground">
              {dir === "rtl" ? "صنف معتمد" : "Registered"}
            </span>
          </div>
        </div>

        <div className="p-5 rounded-xl border bg-card/60 backdrop-blur shadow-sm space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-medium">{t.products.stock_items}</span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600">
              <Boxes className="h-4 w-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {items.filter((i) => i.item_type === "STOCK").length}
            </span>
            <span className="text-xs text-blue-600 font-medium">
              {dir === "rtl" ? "تتبع جرد وباركود" : "Stock Managed"}
            </span>
          </div>
        </div>

        <div className="p-5 rounded-xl border bg-card/60 backdrop-blur shadow-sm space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-medium">{t.products.service_items}</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600">
              <Cpu className="h-4 w-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">
              {items.filter((i) => i.item_type === "SERVICE").length}
            </span>
            <span className="text-xs text-amber-600 font-medium">
              {dir === "rtl" ? "خدمات فورية" : "Service Units"}
            </span>
          </div>
        </div>

        <div className="p-5 rounded-xl border bg-card/60 backdrop-blur shadow-sm space-y-2">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-medium">{t.products.active_skus}</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-600">
              <Barcode className="h-4 w-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-foreground">{items.length + 3}</span>
            <span className="text-xs text-emerald-600 font-medium">
              {dir === "rtl" ? "مع الباركود والمتغيرات" : "Active Barcodes"}
            </span>
          </div>
        </div>
      </div>

      {/* Tabs Bar */}
      <div className="flex items-center gap-2 border-b border-border/60 pb-px overflow-x-auto">
        <button
          onClick={() => setActiveTab("items")}
          className={`inline-flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap ${
            activeTab === "items"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Package className="h-4 w-4" />
          {t.products.tab_items}
        </button>

        <button
          onClick={() => setActiveTab("scanner")}
          className={`inline-flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap ${
            activeTab === "scanner"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Barcode className="h-4 w-4" />
          {t.products.tab_scanner}
        </button>

        <button
          onClick={() => setActiveTab("categories")}
          className={`inline-flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap ${
            activeTab === "categories"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FolderTree className="h-4 w-4" />
          {t.products.tab_categories}
        </button>

        <button
          onClick={() => setActiveTab("uoms")}
          className={`inline-flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap ${
            activeTab === "uoms"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Scale className="h-4 w-4" />
          {t.products.tab_uoms}
        </button>
      </div>

      {/* TAB 1: Product Master List */}
      {activeTab === "items" && (
        <div className="space-y-4">
          {/* Filters Bar */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t.products.search_placeholder}
                className="w-full rounded-xl border bg-card ps-9 pe-4 py-2 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring shadow-sm"
              />
            </div>

            <div className="flex items-center gap-2 overflow-x-auto">
              {["ALL", "STOCK", "SERVICE", "ASSET", "NON_STOCK"].map((type) => (
                <button
                  key={type}
                  onClick={() => setTypeFilter(type)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                    typeFilter === type
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-card text-muted-foreground hover:text-foreground hover:bg-muted border-border"
                  }`}
                >
                  {type === "ALL"
                    ? dir === "rtl"
                      ? "الكل"
                      : "All"
                    : type}
                </button>
              ))}
            </div>
          </div>

          {/* Table */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="py-3 px-4 text-start">{t.products.code}</th>
                    <th className="py-3 px-4 text-start">{dir === "rtl" ? t.products.name_ar : t.products.name_en}</th>
                    <th className="py-3 px-4 text-start">{t.products.type}</th>
                    <th className="py-3 px-4 text-start">{t.products.tracking}</th>
                    <th className="py-3 px-4 text-start">{t.products.base_uom}</th>
                    <th className="py-3 px-4 text-end">{t.products.cost_price}</th>
                    <th className="py-3 px-4 text-end">{t.products.list_price}</th>
                    <th className="py-3 px-4 text-start">{t.products.barcode}</th>
                    <th className="py-3 px-4 text-center">{dir === "rtl" ? "الحالة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredItems.map((item) => (
                    <tr key={item.id} className="hover:bg-muted/40 transition-colors">
                      <td className="py-3 px-4 font-mono font-bold text-primary">
                        {item.code}
                      </td>
                      <td className="py-3 px-4">
                        <div className="font-medium text-foreground">
                          {dir === "rtl" ? item.name_ar : item.name_en}
                        </div>
                        <div className="text-[11px] text-muted-foreground">
                          {dir === "rtl" ? item.name_en : item.name_ar} • {item.category}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                            item.item_type === "STOCK"
                              ? "bg-blue-500/10 text-blue-600 border border-blue-500/20"
                              : item.item_type === "SERVICE"
                              ? "bg-amber-500/10 text-amber-600 border border-amber-500/20"
                              : "bg-purple-500/10 text-purple-600 border border-purple-500/20"
                          }`}
                        >
                          {item.item_type}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-muted-foreground font-mono text-[11px]">
                          {item.tracking_type}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono font-medium">
                        {item.base_uom}
                      </td>
                      <td className="py-3 px-4 text-end font-mono text-muted-foreground">
                        {item.cost_price.toFixed(2)} SAR
                      </td>
                      <td className="py-3 px-4 text-end font-mono font-bold text-foreground">
                        {item.list_price.toFixed(2)} SAR
                      </td>
                      <td className="py-3 px-4 font-mono text-[11px] text-muted-foreground">
                        {item.barcode}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span className="inline-flex items-center gap-1 text-[11px] text-emerald-600 font-medium">
                          <CheckCircle2 className="h-3.5 w-3.5" />
                          {dir === "rtl" ? "نشط" : "Active"}
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

      {/* TAB 2: Live Barcode Scanner Simulator */}
      {activeTab === "scanner" && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="p-6 rounded-2xl border bg-card shadow-sm space-y-5">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                <Barcode className="h-5 w-5" />
              </div>
              <div>
                <h2 className="text-base font-bold text-foreground">
                  {t.products.tab_scanner}
                </h2>
                <p className="text-xs text-muted-foreground">
                  {dir === "rtl"
                    ? "محاكاة قراءة الباركود السريع لأجهزة نقاط البيع ومسدسات الليزر المحمولة"
                    : "Laser barcode gun and POS scanner lookup simulator"}
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <label className="text-xs font-medium text-foreground block">
                {t.products.barcode}
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={scanCode}
                  onChange={(e) => setScanCode(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleScan()}
                  placeholder={t.products.scanner_placeholder}
                  className="flex-1 rounded-xl border bg-background px-4 py-2 text-xs font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                />
                <button
                  onClick={handleScan}
                  className="px-4 py-2 rounded-xl text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition-colors flex items-center gap-1.5"
                >
                  <Search className="h-3.5 w-3.5" />
                  {t.products.scan_now}
                </button>
              </div>

              {/* Quick Barcode Test Buttons */}
              <div className="space-y-1.5 pt-2">
                <span className="text-[11px] text-muted-foreground block">
                  {dir === "rtl" ? "أكواد تجريبية سريعة:" : "Quick test barcodes:"}
                </span>
                <div className="flex flex-wrap gap-2">
                  {items.map((it) => (
                    <button
                      key={it.id}
                      onClick={() => {
                        setScanCode(it.barcode);
                        setMatchedItem(it);
                        setScanError("");
                      }}
                      className="px-2.5 py-1 rounded-md text-[11px] font-mono border bg-muted/50 hover:bg-muted text-foreground transition-colors"
                    >
                      {it.barcode} ({it.code})
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Matched Product Output Display */}
          <div className="p-6 rounded-2xl border bg-card/60 backdrop-blur shadow-sm space-y-4">
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              <Package className="h-4 w-4 text-primary" />
              {t.products.scanned_product}
            </h3>

            {scanError && (
              <div className="p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive flex items-center gap-2 text-xs">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{scanError}</span>
              </div>
            )}

            {matchedItem ? (
              <div className="p-5 rounded-xl border bg-background/80 space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-primary/10 text-primary border border-primary/20">
                      {matchedItem.code}
                    </span>
                    <h4 className="text-base font-bold text-foreground mt-1.5">
                      {dir === "rtl" ? matchedItem.name_ar : matchedItem.name_en}
                    </h4>
                    <span className="text-xs text-muted-foreground">
                      {dir === "rtl" ? matchedItem.name_en : matchedItem.name_ar}
                    </span>
                  </div>
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                    MATCH CONFIRMED
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-3 pt-2 border-t border-border">
                  <div>
                    <span className="text-[10px] text-muted-foreground block">{t.products.list_price}</span>
                    <span className="text-lg font-bold font-mono text-primary">
                      {matchedItem.list_price.toFixed(2)} SAR
                    </span>
                  </div>
                  <div>
                    <span className="text-[10px] text-muted-foreground block">{t.products.base_uom}</span>
                    <span className="text-sm font-semibold text-foreground font-mono">
                      {matchedItem.base_uom}
                    </span>
                  </div>
                  <div>
                    <span className="text-[10px] text-muted-foreground block">{t.products.type}</span>
                    <span className="text-xs font-semibold text-foreground">
                      {matchedItem.item_type} ({matchedItem.tracking_type})
                    </span>
                  </div>
                  <div>
                    <span className="text-[10px] text-muted-foreground block">ZATCA Tax Profile</span>
                    <span className="text-xs font-semibold text-foreground">
                      {matchedItem.tax_code}
                    </span>
                  </div>
                </div>
              </div>
            ) : !scanError ? (
              <div className="p-8 text-center border border-dashed rounded-xl text-muted-foreground text-xs">
                {dir === "rtl"
                  ? "قم بمسح الباركود أو اختيار أحد الأكواد التجريبية أعلاه لعرض التفاصيل الفورية"
                  : "Scan a barcode or choose a test code above to display product details"}
              </div>
            ) : null}
          </div>
        </div>
      )}

      {/* TAB 3: Hierarchical Categories Tree */}
      {activeTab === "categories" && (
        <div className="p-6 rounded-2xl border bg-card shadow-sm space-y-5">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-base font-bold text-foreground">
                {t.products.tab_categories}
              </h2>
              <p className="text-xs text-muted-foreground">
                {dir === "rtl"
                  ? "الهيكل الشجري لتصنيف الأصناف مع توريث الخصائص الضريبية والمحاسبية"
                  : "Hierarchical product categories with tax & GL account inheritance"}
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 rounded-xl border bg-background/50 space-y-3">
              <div className="flex items-center gap-2 text-primary font-bold text-xs">
                <FolderTree className="h-4 w-4" />
                <span>CAT-IT: أجهزة وتقنيات الحاسب الآلي (IT Hardware)</span>
              </div>
              <div className="ps-6 border-s-2 border-primary/20 space-y-2">
                <div className="p-2.5 rounded-lg border bg-card text-xs flex justify-between items-center">
                  <span className="font-medium">IT-LAPTOPS: الحواسيب المحمولة والأجهزة اللوحية</span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-mono">1 Item</span>
                </div>
                <div className="p-2.5 rounded-lg border bg-card text-xs flex justify-between items-center">
                  <span className="font-medium">IT-NETWORKING: البنية التحتية والشبكات</span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-mono">1 Item</span>
                </div>
              </div>
            </div>

            <div className="p-4 rounded-xl border bg-background/50 space-y-3">
              <div className="flex items-center gap-2 text-emerald-600 font-bold text-xs">
                <FolderTree className="h-4 w-4" />
                <span>CAT-SERV: الخدمات والاستشارات التقنية (Professional Services)</span>
              </div>
              <div className="ps-6 border-s-2 border-emerald-500/20 space-y-2">
                <div className="p-2.5 rounded-lg border bg-card text-xs flex justify-between items-center">
                  <span className="font-medium">SRV-ADVISORY: الاستشارات والتحول الرقمي</span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-mono">1 Item</span>
                </div>
                <div className="p-2.5 rounded-lg border bg-card text-xs flex justify-between items-center">
                  <span className="font-medium">SRV-SUPPORT: عقود الصيانة والدعم الفني السنوي</span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-mono">0 Items</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 4: Units of Measure (UOM) & Live Converter */}
      {activeTab === "uoms" && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* UOM Master List */}
          <div className="p-6 rounded-2xl border bg-card shadow-sm space-y-4">
            <h2 className="text-base font-bold text-foreground">
              {t.products.tab_uoms}
            </h2>
            <div className="space-y-2">
              {[
                { code: "PCS", name_ar: "حبة / قطعة", name_en: "Piece", type: "COUNT", symbol: "pcs" },
                { code: "BOX", name_ar: "صندوق / كرتون", name_en: "Box", type: "COUNT", symbol: "bx" },
                { code: "KG", name_ar: "كيلوجرام", name_en: "Kilogram", type: "WEIGHT", symbol: "kg" },
                { code: "G", name_ar: "جرام", name_en: "Gram", type: "WEIGHT", symbol: "g" },
                { code: "M", name_ar: "متر", name_en: "Meter", type: "LENGTH", symbol: "m" },
                { code: "HOUR", name_ar: "ساعة عمل", name_en: "Labor Hour", type: "TIME", symbol: "hr" },
              ].map((u) => (
                <div key={u.code} className="flex items-center justify-between p-3 rounded-xl border bg-background/60 text-xs">
                  <div className="flex items-center gap-3">
                    <span className="font-mono font-bold text-primary w-12">{u.code}</span>
                    <div>
                      <span className="font-medium text-foreground">{dir === "rtl" ? u.name_ar : u.name_en}</span>
                      <span className="text-[11px] text-muted-foreground ms-2">({u.symbol})</span>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-muted text-muted-foreground">
                    {u.type}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Real-time UOM Converter */}
          <div className="p-6 rounded-2xl border bg-card shadow-sm space-y-5">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                <ArrowRightLeft className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-foreground">
                  {t.products.uom_converter_title}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {dir === "rtl"
                    ? "تحويل الكميات بين الوحدات المتوافقة بدقة حسابية تامة"
                    : "Exact physical unit quantity conversion calculator"}
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <label className="text-xs font-medium text-foreground block mb-1.5">
                  {t.products.quantity}
                </label>
                <input
                  type="number"
                  step="any"
                  value={convertQty}
                  onChange={(e) => setConvertQty(parseFloat(e.target.value) || 0)}
                  className="w-full rounded-xl border bg-background px-4 py-2 text-xs font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-foreground block mb-1.5">
                    {t.products.convert_from}
                  </label>
                  <select
                    value={convertFrom}
                    onChange={(e) => setConvertFrom(e.target.value)}
                    className="w-full rounded-xl border bg-background px-3 py-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                  >
                    <option value="KG">KG (كيلوجرام)</option>
                    <option value="G">G (جرام)</option>
                    <option value="BOX">BOX (كرتون)</option>
                    <option value="PCS">PCS (حبة)</option>
                    <option value="M">M (متر)</option>
                    <option value="CM">CM (سنتيمتر)</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-medium text-foreground block mb-1.5">
                    {t.products.convert_to}
                  </label>
                  <select
                    value={convertTo}
                    onChange={(e) => setConvertTo(e.target.value)}
                    className="w-full rounded-xl border bg-background px-3 py-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                  >
                    <option value="G">G (جرام)</option>
                    <option value="KG">KG (كيلوجرام)</option>
                    <option value="PCS">PCS (حبة)</option>
                    <option value="BOX">BOX (كرتون)</option>
                    <option value="CM">CM (سنتيمتر)</option>
                    <option value="M">M (متر)</option>
                  </select>
                </div>
              </div>

              <button
                onClick={handleConvert}
                className="w-full py-2.5 rounded-xl text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                {t.products.convert_btn}
              </button>

              {convertResult !== null && (
                <div className="p-4 rounded-xl bg-primary/5 border border-primary/20 flex items-center justify-between">
                  <span className="text-xs font-medium text-muted-foreground">
                    {t.products.converted_result}:
                  </span>
                  <span className="text-lg font-bold font-mono text-primary">
                    {convertResult} {convertTo}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* CREATE ITEM MODAL */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-lg rounded-2xl border bg-card p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-base font-bold text-foreground flex items-center gap-2">
                <Plus className="h-4 w-4 text-primary" />
                {t.products.create_item}
              </h3>
              <button
                onClick={() => setIsCreateOpen(false)}
                className="p-1 rounded-lg hover:bg-muted text-muted-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleCreateItem} className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.code}</label>
                  <input
                    type="text"
                    required
                    value={newCode}
                    onChange={(e) => setNewCode(e.target.value)}
                    placeholder="IT-DESK-01"
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs font-mono uppercase focus:outline-none focus:ring-1 focus:ring-ring"
                  />
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.type}</label>
                  <select
                    value={newType}
                    onChange={(e) => setNewType(e.target.value as any)}
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                  >
                    <option value="STOCK">{t.products.type_stock}</option>
                    <option value="SERVICE">{t.products.type_service}</option>
                    <option value="NON_STOCK">{t.products.type_non_stock}</option>
                    <option value="ASSET">{t.products.type_asset}</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="font-medium text-foreground block mb-1">{t.products.name_ar}</label>
                <input
                  type="text"
                  required
                  value={newNameAr}
                  onChange={(e) => setNewNameAr(e.target.value)}
                  placeholder="مثال: شاشة عرض 27 بوصة بدقة 4K"
                  className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <div>
                <label className="font-medium text-foreground block mb-1">{t.products.name_en}</label>
                <input
                  type="text"
                  required
                  value={newNameEn}
                  onChange={(e) => setNewNameEn(e.target.value)}
                  placeholder="e.g. 27-inch 4K Professional Display"
                  className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.tracking}</label>
                  <select
                    value={newTracking}
                    onChange={(e) => setNewTracking(e.target.value as any)}
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                  >
                    <option value="NONE">{t.products.tracking_none}</option>
                    <option value="LOT">{t.products.tracking_lot}</option>
                    <option value="SERIAL">{t.products.tracking_serial}</option>
                  </select>
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.base_uom}</label>
                  <select
                    value={newBaseUOM}
                    onChange={(e) => setNewBaseUOM(e.target.value)}
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                  >
                    <option value="PCS">PCS (قطعة)</option>
                    <option value="BOX">BOX (صندوق)</option>
                    <option value="KG">KG (كيلوجرام)</option>
                    <option value="HOUR">HOUR (ساعة)</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.cost_price} (SAR)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={newCost}
                    onChange={(e) => setNewCost(parseFloat(e.target.value) || 0)}
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                  />
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">{t.products.list_price} (SAR)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={newList}
                    onChange={(e) => setNewList(parseFloat(e.target.value) || 0)}
                    className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                  />
                </div>
              </div>

              <div>
                <label className="font-medium text-foreground block mb-1">{t.products.barcode} (اختياري)</label>
                <input
                  type="text"
                  value={newBarcode}
                  onChange={(e) => setNewBarcode(e.target.value)}
                  placeholder="628..."
                  className="w-full rounded-lg border bg-background px-3 py-1.5 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="px-4 py-2 rounded-xl text-xs font-medium border hover:bg-muted text-muted-foreground"
                >
                  {t.products.cancel}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90"
                >
                  {t.products.save}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

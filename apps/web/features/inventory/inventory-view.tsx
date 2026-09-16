"use client";

import React, { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  Boxes,
  Warehouse,
  ArrowRightLeft,
  SlidersHorizontal,
  History,
  ShieldCheck,
  Plus,
  Search,
  CheckCircle2,
  AlertCircle,
  Package,
  Layers,
  Building2,
  Lock,
  ArrowUpRight,
  ArrowDownLeft,
  FileText,
  Clock,
  Check,
  X,
  Sparkles,
  QrCode,
  Tag,
  AlertTriangle,
  Calculator,
  Coins,
  TrendingUp,
  TrendingDown,
  RefreshCw,
} from "lucide-react";

interface StockBalanceItem {
  id: string;
  itemCode: string;
  nameAr: string;
  nameEn: string;
  warehouseCode: string;
  warehouseNameAr: string;
  warehouseNameEn: string;
  binCode: string;
  onHand: number;
  reserved: number;
  available: number;
  averageCost: number;
  totalValue: number;
  trackingType: "NONE" | "LOT" | "SERIAL";
  uom: string;
}

interface StockMovementItem {
  id: string;
  movementNumber: string;
  movementType: "OPENING" | "PURCHASE_RECEIPT" | "SALES_ISSUE" | "TRANSFER_IN" | "TRANSFER_OUT" | "ADJUSTMENT_IN" | "ADJUSTMENT_OUT";
  itemCode: string;
  itemNameAr: string;
  itemNameEn: string;
  warehouseCode: string;
  binCode: string;
  quantity: number;
  unitCost: number;
  totalCost: number;
  occurredAt: string;
  sourceDoc: string;
}

interface WarehouseEntity {
  id: string;
  code: string;
  nameAr: string;
  nameEn: string;
  type: "STANDARD" | "TRANSIT" | "COLD_STORAGE" | "QUARANTINE";
  city: string;
  binCount: number;
  totalStockItems: number;
  isTransit: boolean;
  isActive: boolean;
}

interface TransferOrderEntity {
  id: string;
  orderNumber: string;
  fromWarehouse: string;
  toWarehouse: string;
  status: "DRAFT" | "REQUESTED" | "APPROVED" | "DISPATCHED" | "RECEIVED";
  itemsCount: number;
  totalQuantity: number;
  createdAt: string;
  notes: string;
}

interface AdjustmentEntity {
  id: string;
  adjustmentNumber: string;
  warehouse: string;
  adjustmentType: "INCREASE" | "DECREASE" | "SCRAP" | "FOUND" | "DAMAGED";
  reasonCode: string;
  status: "DRAFT" | "APPROVED" | "POSTED";
  totalVarianceValue: number;
  createdAt: string;
}

interface CostLayerItem {
  id: string;
  itemCode: string;
  itemNameAr: string;
  itemNameEn: string;
  warehouseCode: string;
  layerDate: string;
  originalQuantity: number;
  remainingQuantity: number;
  unitCost: number;
  totalRemainingValue: number;
  sourceDoc: string;
  status: "ACTIVE" | "EXHAUSTED";
}

interface RevaluationItem {
  id: string;
  revaluationNumber: string;
  itemCode: string;
  itemNameAr: string;
  warehouseCode: string;
  quantity: number;
  oldUnitCost: number;
  newUnitCost: number;
  varianceAmount: number;
  reason: string;
  performedAt: string;
}

const INITIAL_BALANCES: StockBalanceItem[] = [
  {
    id: "bal-1",
    itemCode: "IT-LAPTOP-01",
    nameAr: "جهاز حاسب محمول للأعمال 16 بوصة",
    nameEn: "Enterprise Business Laptop 16-inch",
    warehouseCode: "WH-MAIN",
    warehouseNameAr: "المستودع الرئيسي - الرياض",
    warehouseNameEn: "Riyadh Central Warehouse",
    binCode: "A-01-RACK1",
    onHand: 45,
    reserved: 5,
    available: 40,
    averageCost: 3800,
    totalValue: 171000,
    trackingType: "SERIAL",
    uom: "PCS",
  },
  {
    id: "bal-2",
    itemCode: "IND-PUMP-HD",
    nameAr: "مضخة هيدروليكية ضغط عالي",
    nameEn: "Heavy Duty Hydraulic Pump",
    warehouseCode: "WH-MAIN",
    warehouseNameAr: "المستودع الرئيسي - الرياض",
    warehouseNameEn: "Riyadh Central Warehouse",
    binCode: "B-03-PALLET",
    onHand: 18,
    reserved: 2,
    available: 16,
    averageCost: 4500,
    totalValue: 81000,
    trackingType: "LOT",
    uom: "UNIT",
  },
  {
    id: "bal-3",
    itemCode: "IND-VALVE-50",
    nameAr: "صمام أمان صناعي 50 ملم",
    nameEn: "Industrial Safety Valve 50mm",
    warehouseCode: "WH-DAMMAM",
    warehouseNameAr: "مستودع فرع الدمام",
    warehouseNameEn: "Dammam Branch Depot",
    binCode: "D-01-SHELF",
    onHand: 120,
    reserved: 15,
    available: 105,
    averageCost: 280,
    totalValue: 33600,
    trackingType: "NONE",
    uom: "PCS",
  },
  {
    id: "bal-4",
    itemCode: "CHEM-SOLVENT-01",
    nameAr: "مذيب كيميائي صناعي (درجة أولى)",
    nameEn: "Industrial Cleaning Solvent Grade A",
    warehouseCode: "WH-COLD",
    warehouseNameAr: "مستودع التخزين المبرد والمواد الحساسة",
    warehouseNameEn: "Cold Storage & Chemical Facility",
    binCode: "C-01-COLD",
    onHand: 60,
    reserved: 0,
    available: 60,
    averageCost: 150,
    totalValue: 9000,
    trackingType: "LOT",
    uom: "LITER",
  },
  {
    id: "bal-5",
    itemCode: "PPE-HELMET-PRO",
    nameAr: "خوذة سلامة موقعية مطابقة للمعايير",
    nameEn: "Industrial Safety Helmet ANSI Certified",
    warehouseCode: "WH-MAIN",
    warehouseNameAr: "المستودع الرئيسي - الرياض",
    warehouseNameEn: "Riyadh Central Warehouse",
    binCode: "A-04-BOX",
    onHand: 350,
    reserved: 20,
    available: 330,
    averageCost: 45,
    totalValue: 15750,
    trackingType: "NONE",
    uom: "PCS",
  },
];

const INITIAL_MOVEMENTS: StockMovementItem[] = [
  {
    id: "mov-1",
    movementNumber: "MOV-2026-0091",
    movementType: "PURCHASE_RECEIPT",
    itemCode: "IT-LAPTOP-01",
    itemNameAr: "جهاز حاسب محمول للأعمال 16 بوصة",
    itemNameEn: "Enterprise Business Laptop 16-inch",
    warehouseCode: "WH-MAIN",
    binCode: "A-01-RACK1",
    quantity: 20,
    unitCost: 3800,
    totalCost: 76000,
    occurredAt: "2026-09-16 10:15",
    sourceDoc: "PO-2026-0045",
  },
  {
    id: "mov-2",
    movementNumber: "MOV-2026-0090",
    movementType: "SALES_ISSUE",
    itemCode: "IND-VALVE-50",
    itemNameAr: "صمام أمان صناعي 50 ملم",
    itemNameEn: "Industrial Safety Valve 50mm",
    warehouseCode: "WH-DAMMAM",
    binCode: "D-01-SHELF",
    quantity: -15,
    unitCost: 280,
    totalCost: 4200,
    occurredAt: "2026-09-16 09:30",
    sourceDoc: "SO-2026-0120",
  },
  {
    id: "mov-3",
    movementNumber: "MOV-2026-0089",
    movementType: "TRANSFER_OUT",
    itemCode: "IND-PUMP-HD",
    itemNameAr: "مضخة هيدروليكية ضغط عالي",
    itemNameEn: "Heavy Duty Hydraulic Pump",
    warehouseCode: "WH-MAIN",
    binCode: "B-03-PALLET",
    quantity: -2,
    unitCost: 4500,
    totalCost: 9000,
    occurredAt: "2026-09-15 16:45",
    sourceDoc: "TRF-2026-0012",
  },
  {
    id: "mov-4",
    movementNumber: "MOV-2026-0088",
    movementType: "ADJUSTMENT_IN",
    itemCode: "PPE-HELMET-PRO",
    itemNameAr: "خوذة سلامة موقعية مطابقة للمعايير",
    itemNameEn: "Industrial Safety Helmet ANSI Certified",
    warehouseCode: "WH-MAIN",
    binCode: "A-04-BOX",
    quantity: 10,
    unitCost: 45,
    totalCost: 450,
    occurredAt: "2026-09-15 11:20",
    sourceDoc: "ADJ-2026-0004",
  },
  {
    id: "mov-5",
    movementNumber: "MOV-2026-0087",
    movementType: "OPENING",
    itemCode: "CHEM-SOLVENT-01",
    itemNameAr: "مذيب كيميائي صناعي (درجة أولى)",
    itemNameEn: "Industrial Cleaning Solvent Grade A",
    warehouseCode: "WH-COLD",
    binCode: "C-01-COLD",
    quantity: 60,
    unitCost: 150,
    totalCost: 9000,
    occurredAt: "2026-09-01 08:00",
    sourceDoc: "SYS-INIT-BALANCE",
  },
];

const INITIAL_WAREHOUSES: WarehouseEntity[] = [
  {
    id: "wh-1",
    code: "WH-MAIN",
    nameAr: "المستودع المركزي الرئيسي - الرياض",
    nameEn: "Riyadh Central Distribution Warehouse",
    type: "STANDARD",
    city: "الرياض / حي السلي",
    binCount: 120,
    totalStockItems: 1450,
    isTransit: false,
    isActive: true,
  },
  {
    id: "wh-2",
    code: "WH-DAMMAM",
    nameAr: "مستودع المنطقة الشرقية - الدمام",
    nameEn: "Eastern Province Logistics Center",
    type: "STANDARD",
    city: "الدمام / الخالدية",
    binCount: 65,
    totalStockItems: 680,
    isTransit: false,
    isActive: true,
  },
  {
    id: "wh-3",
    code: "WH-COLD",
    nameAr: "مستودع التبريد والتخزين المتخصص",
    nameEn: "Specialized Cold Storage Facility",
    type: "COLD_STORAGE",
    city: "الرياض / المدينة الصناعية الثانية",
    binCount: 30,
    totalStockItems: 210,
    isTransit: false,
    isActive: true,
  },
  {
    id: "wh-4",
    code: "WH-TRANSIT",
    nameAr: "مستودع بضاعة قيد الطريق والنقل",
    nameEn: "In-Transit Goods Virtual Warehouse",
    type: "TRANSIT",
    city: "افتراضي عبر المملكة",
    binCount: 1,
    totalStockItems: 14,
    isTransit: true,
    isActive: true,
  },
];

const INITIAL_TRANSFERS: TransferOrderEntity[] = [
  {
    id: "trf-1",
    orderNumber: "TRF-2026-0012",
    fromWarehouse: "WH-MAIN (الرياض)",
    toWarehouse: "WH-DAMMAM (الدمام)",
    status: "DISPATCHED",
    itemsCount: 3,
    totalQuantity: 25,
    createdAt: "2026-09-15",
    notes: "شحنة دورية لتعزيز مخزون فرع الشرقية",
  },
  {
    id: "trf-2",
    orderNumber: "TRF-2026-0013",
    fromWarehouse: "WH-MAIN (الرياض)",
    toWarehouse: "WH-COLD (الصناعية 2)",
    status: "RECEIVED",
    itemsCount: 1,
    totalQuantity: 50,
    createdAt: "2026-09-14",
    notes: "نقل مواد حساسة إلى مستودع التبريد",
  },
  {
    id: "trf-3",
    orderNumber: "TRF-2026-0014",
    fromWarehouse: "WH-DAMMAM (الدمام)",
    toWarehouse: "WH-MAIN (الرياض)",
    status: "REQUESTED",
    itemsCount: 2,
    totalQuantity: 12,
    createdAt: "2026-09-16",
    notes: "إعادة توازن مخزون المضخات",
  },
];

const INITIAL_ADJUSTMENTS: AdjustmentEntity[] = [
  {
    id: "adj-1",
    adjustmentNumber: "ADJ-2026-0004",
    warehouse: "WH-MAIN (الرياض)",
    adjustmentType: "FOUND",
    reasonCode: "CYCLE_COUNT_SURPLUS",
    status: "POSTED",
    totalVarianceValue: 450,
    createdAt: "2026-09-15",
  },
  {
    id: "adj-2",
    adjustmentNumber: "ADJ-2026-0005",
    warehouse: "WH-DAMMAM (الدمام)",
    adjustmentType: "DAMAGED",
    reasonCode: "SHIPPING_DAMAGE",
    status: "APPROVED",
    totalVarianceValue: -1120,
    createdAt: "2026-09-16",
  },
];

const INITIAL_COST_LAYERS: CostLayerItem[] = [
  {
    id: "layer-1",
    itemCode: "IT-LAPTOP-01",
    itemNameAr: "جهاز حاسب محمول للأعمال 16 بوصة",
    itemNameEn: "Enterprise Business Laptop 16-inch",
    warehouseCode: "WH-MAIN",
    layerDate: "2026-09-01 09:30",
    originalQuantity: 30,
    remainingQuantity: 15,
    unitCost: 3750,
    totalRemainingValue: 56250,
    sourceDoc: "PO-2026-0012",
    status: "ACTIVE",
  },
  {
    id: "layer-2",
    itemCode: "IT-LAPTOP-01",
    itemNameAr: "جهاز حاسب محمول للأعمال 16 بوصة",
    itemNameEn: "Enterprise Business Laptop 16-inch",
    warehouseCode: "WH-MAIN",
    layerDate: "2026-09-10 14:15",
    originalQuantity: 30,
    remainingQuantity: 30,
    unitCost: 3850,
    totalRemainingValue: 115500,
    sourceDoc: "PO-2026-0045",
    status: "ACTIVE",
  },
  {
    id: "layer-3",
    itemCode: "IND-PUMP-HD",
    itemNameAr: "مضخة هيدروليكية ضغط عالي",
    itemNameEn: "Heavy Duty Hydraulic Pump",
    warehouseCode: "WH-MAIN",
    layerDate: "2026-08-20 11:00",
    originalQuantity: 20,
    remainingQuantity: 18,
    unitCost: 4500,
    totalRemainingValue: 81000,
    sourceDoc: "PO-2026-0005",
    status: "ACTIVE",
  },
  {
    id: "layer-4",
    itemCode: "IND-VALVE-50",
    itemNameAr: "صمام أمان صناعي 50 ملم",
    itemNameEn: "Industrial Safety Valve 50mm",
    warehouseCode: "WH-DAMMAM",
    layerDate: "2026-08-15 08:45",
    originalQuantity: 50,
    remainingQuantity: 0,
    unitCost: 260,
    totalRemainingValue: 0,
    sourceDoc: "PO-2026-0002",
    status: "EXHAUSTED",
  },
  {
    id: "layer-5",
    itemCode: "IND-VALVE-50",
    itemNameAr: "صمام أمان صناعي 50 ملم",
    itemNameEn: "Industrial Safety Valve 50mm",
    warehouseCode: "WH-DAMMAM",
    layerDate: "2026-09-05 13:20",
    originalQuantity: 150,
    remainingQuantity: 120,
    unitCost: 280,
    totalRemainingValue: 33600,
    sourceDoc: "PO-2026-0033",
    status: "ACTIVE",
  },
];

const INITIAL_REVALUATIONS: RevaluationItem[] = [
  {
    id: "rev-1",
    revaluationNumber: "REV-2026-0001",
    itemCode: "IT-LAPTOP-01",
    itemNameAr: "جهاز حاسب محمول للأعمال 16 بوصة",
    warehouseCode: "WH-MAIN",
    quantity: 45,
    oldUnitCost: 3700,
    newUnitCost: 3800,
    varianceAmount: 4500,
    reason: "تحديث تكلفة الاستيراد السنوية والرسوم الجمركية",
    performedAt: "2026-09-12 11:30",
  },
  {
    id: "rev-2",
    revaluationNumber: "REV-2026-0002",
    itemCode: "CHEM-SOLVENT-01",
    itemNameAr: "مذيب كيميائي صناعي (درجة أولى)",
    warehouseCode: "WH-COLD",
    quantity: 60,
    oldUnitCost: 165,
    newUnitCost: 150,
    varianceAmount: -900,
    reason: "انخفاض أسعار السوق العالمية وتصحيح تكلفة التخزين",
    performedAt: "2026-09-14 16:00",
  },
];

export function InventoryView() {
  const { locale, t } = useLocale();
  const isAr = locale === "ar";
  const inv = (t as any).inventory || {};

  const [activeTab, setActiveTab] = useState<"balances" | "movements" | "warehouses" | "transfers" | "adjustments" | "costing">("balances");
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedWarehouseFilter, setSelectedWarehouseFilter] = useState("ALL");

  // State collections
  const [balances, setBalances] = useState<StockBalanceItem[]>(INITIAL_BALANCES);
  const [movements, setMovements] = useState<StockMovementItem[]>(INITIAL_MOVEMENTS);
  const [transfers, setTransfers] = useState<TransferOrderEntity[]>(INITIAL_TRANSFERS);
  const [costLayers, setCostLayers] = useState<CostLayerItem[]>(INITIAL_COST_LAYERS);
  const [revaluations, setRevaluations] = useState<RevaluationItem[]>(INITIAL_REVALUATIONS);

  // Modals
  const [showMovementModal, setShowMovementModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showRevalModal, setShowRevalModal] = useState(false);
  const [notification, setNotification] = useState<string | null>(null);

  // New Movement Form
  const [movItemCode, setMovItemCode] = useState("IT-LAPTOP-01");
  const [movWarehouse, setMovWarehouse] = useState("WH-MAIN");
  const [movType, setMovType] = useState<StockMovementItem["movementType"]>("PURCHASE_RECEIPT");
  const [movQty, setMovQty] = useState(10);
  const [movUnitCost, setMovUnitCost] = useState(3800);
  const [movDoc, setMovDoc] = useState("MANUAL-ENTRY-01");

  // Revaluation Form
  const [revalItemCode, setRevalItemCode] = useState("IT-LAPTOP-01");
  const [revalWarehouse, setRevalWarehouse] = useState("WH-MAIN");
  const [revalNewCost, setRevalNewCost] = useState(3850);
  const [revalReason, setRevalReason] = useState("");

  // Filtered Balances
  const filteredBalances = balances.filter((b) => {
    const matchesSearch =
      b.itemCode.toLowerCase().includes(searchTerm.toLowerCase()) ||
      b.nameAr.includes(searchTerm) ||
      b.nameEn.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesWH = selectedWarehouseFilter === "ALL" || b.warehouseCode === selectedWarehouseFilter;
    return matchesSearch && matchesWH;
  });

  // Filtered Movements
  const filteredMovements = movements.filter((m) => {
    const matchesSearch =
      m.movementNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      m.itemCode.toLowerCase().includes(searchTerm.toLowerCase()) ||
      m.itemNameAr.includes(searchTerm) ||
      m.sourceDoc.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesWH = selectedWarehouseFilter === "ALL" || m.warehouseCode === selectedWarehouseFilter;
    return matchesSearch && matchesWH;
  });

  // Filtered Cost Layers
  const filteredCostLayers = costLayers.filter((layer) => {
    const matchesSearch =
      layer.itemCode.toLowerCase().includes(searchTerm.toLowerCase()) ||
      layer.itemNameAr.includes(searchTerm) ||
      layer.sourceDoc.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesWH = selectedWarehouseFilter === "ALL" || layer.warehouseCode === selectedWarehouseFilter;
    return matchesSearch && matchesWH;
  });

  // Filtered Revaluations
  const filteredRevaluations = revaluations.filter((rev) => {
    const matchesSearch =
      rev.revaluationNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
      rev.itemCode.toLowerCase().includes(searchTerm.toLowerCase()) ||
      rev.itemNameAr.includes(searchTerm) ||
      rev.reason.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesWH = selectedWarehouseFilter === "ALL" || rev.warehouseCode === selectedWarehouseFilter;
    return matchesSearch && matchesWH;
  });

  // Revaluation active selection calculations
  const selectedRevalBalance = balances.find(
    (b) => b.itemCode === revalItemCode && b.warehouseCode === revalWarehouse
  ) || balances.find((b) => b.itemCode === revalItemCode);
  const revalCurrentQty = selectedRevalBalance?.onHand || 0;
  const revalCurrentCost = selectedRevalBalance?.averageCost || 0;
  const revalUnitVariance = revalNewCost - revalCurrentCost;
  const revalTotalVariance = revalUnitVariance * revalCurrentQty;

  const handlePostRevaluation = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedRevalBalance || revalCurrentQty <= 0) {
      alert(isAr ? "لا يمكن إعادة تقييم صنف بدون رصيد فعلي" : "Cannot revalue item with zero on-hand quantity");
      return;
    }
    const newRev: RevaluationItem = {
      id: `rev-${Date.now()}`,
      revaluationNumber: `REV-2026-${Math.floor(1000 + Math.random() * 9000)}`,
      itemCode: selectedRevalBalance.itemCode,
      itemNameAr: selectedRevalBalance.nameAr,
      warehouseCode: selectedRevalBalance.warehouseCode,
      quantity: revalCurrentQty,
      oldUnitCost: revalCurrentCost,
      newUnitCost: revalNewCost,
      varianceAmount: revalTotalVariance,
      reason: revalReason || (isAr ? "تسوية تقييم مخزني دورية" : "Periodic inventory revaluation"),
      performedAt: new Date().toISOString().replace("T", " ").substring(0, 16),
    };

    setRevaluations([newRev, ...revaluations]);
    setBalances((prev) =>
      prev.map((b) => {
        if (b.itemCode === selectedRevalBalance.itemCode && b.warehouseCode === selectedRevalBalance.warehouseCode) {
          return {
            ...b,
            averageCost: revalNewCost,
            totalValue: b.onHand * revalNewCost,
          };
        }
        return b;
      })
    );

    setShowRevalModal(false);
    setRevalReason("");
    setNotification(inv.success_revaluation || "تمت إعادة تقييم المخزون وتحديث الأرصدة بنجاح");
    setTimeout(() => setNotification(null), 4000);
  };

  // Total valuation calculation
  const totalValuation = balances.reduce((sum, b) => sum + b.totalValue, 0);
  const totalOnHand = balances.reduce((sum, b) => sum + b.onHand, 0);
  const totalReserved = balances.reduce((sum, b) => sum + b.reserved, 0);

  const handlePostMovement = (e: React.FormEvent) => {
    e.preventDefault();
    const isOut = movType === "SALES_ISSUE" || movType === "TRANSFER_OUT" || movType === "ADJUSTMENT_OUT";
    const deltaQty = isOut ? -Math.abs(movQty) : Math.abs(movQty);

    const targetItem = balances.find((b) => b.itemCode === movItemCode) || balances[0];

    const newMov: StockMovementItem = {
      id: `mov-${Date.now()}`,
      movementNumber: `MOV-2026-${Math.floor(1000 + Math.random() * 9000)}`,
      movementType: movType,
      itemCode: targetItem.itemCode,
      itemNameAr: targetItem.nameAr,
      itemNameEn: targetItem.nameEn,
      warehouseCode: movWarehouse,
      binCode: "DEFAULT-BIN",
      quantity: deltaQty,
      unitCost: movUnitCost,
      totalCost: Math.abs(deltaQty) * movUnitCost,
      occurredAt: new Date().toISOString().replace("T", " ").substring(0, 16),
      sourceDoc: movDoc || "MANUAL-TX",
    };

    // Update Movements (Append-Only)
    setMovements([newMov, ...movements]);

    // Update Balances projection
    setBalances((prev) =>
      prev.map((b) => {
        if (b.itemCode === targetItem.itemCode && b.warehouseCode === movWarehouse) {
          const newOnHand = Math.max(0, b.onHand + deltaQty);
          const newAvailable = Math.max(0, newOnHand - b.reserved);
          const newTotalVal = newOnHand * b.averageCost;
          return {
            ...b,
            onHand: newOnHand,
            available: newAvailable,
            totalValue: newTotalVal,
          };
        }
        return b;
      })
    );

    setShowMovementModal(false);
    setNotification(inv.success_movement || "تم تسجيل الحركة بنجاح");
    setTimeout(() => setNotification(null), 4000);
  };

  const getMovementBadge = (type: StockMovementItem["movementType"]) => {
    switch (type) {
      case "OPENING":
        return "bg-blue-500/10 text-blue-500 border-blue-500/30";
      case "PURCHASE_RECEIPT":
        return "bg-emerald-500/10 text-emerald-500 border-emerald-500/30";
      case "SALES_ISSUE":
        return "bg-rose-500/10 text-rose-500 border-rose-500/30";
      case "TRANSFER_IN":
        return "bg-teal-500/10 text-teal-500 border-teal-500/30";
      case "TRANSFER_OUT":
        return "bg-amber-500/10 text-amber-500 border-amber-500/30";
      case "ADJUSTMENT_IN":
        return "bg-purple-500/10 text-purple-500 border-purple-500/30";
      case "ADJUSTMENT_OUT":
        return "bg-red-500/10 text-red-500 border-red-500/30";
      default:
        return "bg-muted text-muted-foreground";
    }
  };

  return (
    <div className="space-y-6">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed bottom-5 end-5 z-50 flex items-center gap-3 bg-emerald-600 text-white px-4 py-3 rounded-lg shadow-xl animate-in slide-in-from-bottom-5">
          <CheckCircle2 className="h-5 w-5" />
          <span className="text-sm font-medium">{notification}</span>
        </div>
      )}

      {/* Header Section */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
              <Boxes className="h-7 w-7 text-primary" />
              {inv.title || "إدارة المستودعات وحركات المخزون وسجل التدقيق"}
            </h1>
            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
              <ShieldCheck className="h-3.5 w-3.5" />
              {inv.immutable_badge || "سجل تدقيق غير قابل للتعديل"}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {inv.subtitle || "متابعة أرصدة المخزون لحظياً، المستودعات والمواقع، أوامر النقل الداخلي، الحجوزات، وتسويات الجرد وفق سجل غير قابل للتعديل"}
          </p>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2.5">
          <button
            onClick={() => setShowRevalModal(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold border bg-card hover:bg-muted text-foreground transition shadow-sm"
          >
            <Coins className="h-4 w-4 text-primary" />
            {inv.new_revaluation || "إعادة تقييم المخزون"}
          </button>
          <button
            onClick={() => setShowMovementModal(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm"
          >
            <Plus className="h-4 w-4" />
            {inv.new_movement || "تسجيل حركة مخزنية"}
          </button>
        </div>
      </div>

      {/* Immutability Banner */}
      <div className="flex items-center justify-between p-3.5 rounded-xl border border-primary/20 bg-primary/5 text-primary">
        <div className="flex items-center gap-3 text-xs font-medium">
          <Lock className="h-4 w-4 shrink-0 text-primary" />
          <span>
            {inv.ledger_guarantee || "حركات المخزون محمية بقاعدة البيانات: القيود المباشرة (Append-Only) تمنع أي تعديل أو حذف (PostgreSQL Trigger Protected)"}
          </span>
        </div>
        <span className="text-[11px] font-mono bg-primary/10 px-2.5 py-0.5 rounded-md border border-primary/20 hidden md:inline-block">
          PESSIMISTIC_ROW_LOCK (SELECT FOR UPDATE)
        </span>
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Card 1: Valuation */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{inv.total_stock_value || "إجمالي تقييم المخزون"}</span>
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Boxes className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">
              {totalValuation.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
            <span className="text-xs text-muted-foreground">{inv.currency || "ر.س"}</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600">
            <ArrowUpRight className="h-3.5 w-3.5" />
            <span>+4.8% مقارنة بالشهر السابق (Moving Avg Cost)</span>
          </div>
        </div>

        {/* Card 2: Warehouses */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{inv.active_warehouses || "المستودعات التشغيلية"}</span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600">
              <Warehouse className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">4</span>
            <span className="text-xs text-muted-foreground">مستودعات و 216 موقع تخزين (Bins)</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
            <Check className="h-3.5 w-3.5 text-emerald-600" />
            <span>مستودع رئيسي، مستودع عبور، تبريد، ومستودع شرقية</span>
          </div>
        </div>

        {/* Card 3: Reserved */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{inv.reserved_stock || "الكميات المحجوزة"}</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600">
              <Lock className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">{totalReserved}</span>
            <span className="text-xs text-muted-foreground">وحدة محجوزة لأوامر البيع</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
            <span>من إجمالي رصيد متاح {totalOnHand} وحدة</span>
          </div>
        </div>

        {/* Card 4: Lots & Serials */}
        <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow transition">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{inv.tracked_lots || "دفعات وتشغيلات متبعة"}</span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-600">
              <QrCode className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono tracking-tight">100%</span>
            <span className="text-xs text-muted-foreground">تتبع كامل للأرقام التسلسلية</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-purple-600">
            <CheckCircle2 className="h-3.5 w-3.5" />
            <span>مطابق لمعايير هيئة الزكاة والمواصفات السعودية</span>
          </div>
        </div>
      </div>

      {/* Tabs Bar & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b pb-3">
        {/* Navigation Tabs */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
          <button
            onClick={() => setActiveTab("balances")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "balances"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Boxes className="h-4 w-4" />
            {inv.tab_balances || "أرصدة المخزون اللحظية"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {filteredBalances.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("movements")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "movements"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <History className="h-4 w-4" />
            {inv.tab_movements || "سجل الحركات المخزنية (Ledger)"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {filteredMovements.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("warehouses")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "warehouses"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Warehouse className="h-4 w-4" />
            {inv.tab_warehouses || "المستودعات والمواقع (Bins)"}
          </button>

          <button
            onClick={() => setActiveTab("transfers")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "transfers"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <ArrowRightLeft className="h-4 w-4" />
            {inv.tab_transfers || "أوامر النقل الداخلي"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              {transfers.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab("adjustments")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "adjustments"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <SlidersHorizontal className="h-4 w-4" />
            {inv.tab_adjustments || "تسويات وجرد المخزون"}
          </button>

          <button
            onClick={() => setActiveTab("costing")}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition shrink-0 ${
              activeTab === "costing"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Calculator className="h-4 w-4" />
            {inv.tab_costing || "محاسبة التكاليف والتقييم (Costing)"}
            <span className="ms-1 px-1.5 py-0.2 rounded-full text-[10px] bg-background/20 font-mono">
              FIFO / AVG
            </span>
          </button>
        </div>

        {/* Filter and Search */}
        <div className="flex items-center gap-2.5">
          {/* Warehouse Dropdown Filter */}
          <select
            value={selectedWarehouseFilter}
            onChange={(e) => setSelectedWarehouseFilter(e.target.value)}
            className="text-xs bg-card border rounded-lg px-2.5 py-1.5 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          >
            <option value="ALL">{inv.filter_warehouse || "كل المستودعات"}</option>
            <option value="WH-MAIN">WH-MAIN (الرياض)</option>
            <option value="WH-DAMMAM">WH-DAMMAM (الدمام)</option>
            <option value="WH-COLD">WH-COLD (المبرد)</option>
            <option value="WH-TRANSIT">WH-TRANSIT (قيد النقل)</option>
          </select>

          {/* Search Input */}
          <div className="relative">
            <Search className="absolute inset-inline-start-2.5 top-2 h-3.5 w-3.5 text-muted-foreground" />
            <input
              type="text"
              placeholder={inv.search_placeholder || "بحث برقم الصنف، الاسم، أو الحركة..."}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="text-xs bg-card border rounded-lg ps-8 pe-3 py-1.5 w-48 sm:w-64 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
        </div>
      </div>

      {/* Tab 1: Live Balances View */}
      {activeTab === "balances" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{inv.item || "الصنف"}</th>
                  <th className="px-4 py-3 text-start">{inv.warehouse || "المستودع"}</th>
                  <th className="px-4 py-3 text-start">{inv.bin || "موقع التخزين (Bin)"}</th>
                  <th className="px-4 py-3 text-end">{inv.onHand || "الرصيد الفعلي"}</th>
                  <th className="px-4 py-3 text-end">{inv.reserved || "المحجوز"}</th>
                  <th className="px-4 py-3 text-end">{inv.available || "المتاح للصرف"}</th>
                  <th className="px-4 py-3 text-end">{inv.avg_cost || "متوسط التكلفة"}</th>
                  <th className="px-4 py-3 text-end">{inv.total_val || "القيمة الإجمالية"}</th>
                  <th className="px-4 py-3 text-center">{inv.status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredBalances.map((item) => (
                  <tr key={item.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3">
                      <div>
                        <div className="font-mono font-semibold text-primary">{item.itemCode}</div>
                        <div className="text-foreground font-medium">{isAr ? item.nameAr : item.nameEn}</div>
                        <div className="text-[10px] text-muted-foreground flex items-center gap-1.5 mt-0.5">
                          <span className="px-1.5 py-0.2 rounded bg-muted font-mono">{item.uom}</span>
                          <span>•</span>
                          <span>تتبع: {item.trackingType}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-semibold text-foreground">{item.warehouseCode}</div>
                      <div className="text-[11px] text-muted-foreground">{isAr ? item.warehouseNameAr : item.warehouseNameEn}</div>
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">
                      <span className="px-2 py-0.5 rounded bg-muted/60 border font-mono text-[11px]">
                        {item.binCode}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                      {item.onHand.toLocaleString(isAr ? "ar-SA" : "en-US")}
                    </td>
                    <td className="px-4 py-3 text-end font-mono text-amber-600 font-semibold">
                      {item.reserved > 0 ? item.reserved.toLocaleString(isAr ? "ar-SA" : "en-US") : "0"}
                    </td>
                    <td className="px-4 py-3 text-end font-mono text-emerald-600 font-bold">
                      {item.available.toLocaleString(isAr ? "ar-SA" : "en-US")}
                    </td>
                    <td className="px-4 py-3 text-end font-mono text-muted-foreground">
                      {item.averageCost.toFixed(2)} {inv.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                      {item.totalValue.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                    </td>
                    <td className="px-4 py-3 text-center">
                      {item.available > 10 ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                          <CheckCircle2 className="h-3 w-3" />
                          متوفر كافٍ
                        </span>
                      ) : item.available > 0 ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/10 text-amber-600 border border-amber-500/20">
                          <AlertTriangle className="h-3 w-3" />
                          مخزون منخفض
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-rose-500/10 text-rose-600 border border-rose-500/20">
                          <X className="h-3 w-3" />
                          نفد المخزون
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Stock Movements Ledger (Append-Only) */}
      {activeTab === "movements" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="p-3.5 bg-muted/30 border-b flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center gap-2">
              <History className="h-4 w-4 text-primary" />
              <span className="font-semibold text-foreground">سجل الحركات التاريخي الشامل (Append-Only Audit Trail)</span>
            </div>
            <span className="text-[11px] font-mono">Total Movements: {filteredMovements.length}</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{inv.movement_number || "رقم الحركة"}</th>
                  <th className="px-4 py-3 text-start">{inv.movement_type || "نوع الحركة"}</th>
                  <th className="px-4 py-3 text-start">{inv.item || "الصنف"}</th>
                  <th className="px-4 py-3 text-start">{inv.warehouse || "المستودع"}</th>
                  <th className="px-4 py-3 text-end">{inv.quantity || "الكمية"}</th>
                  <th className="px-4 py-3 text-end">{inv.unit_cost || "تكلفة الوحدة"}</th>
                  <th className="px-4 py-3 text-end">{inv.total_val || "التكلفة الإجمالية"}</th>
                  <th className="px-4 py-3 text-start">{inv.source_doc || "المستند المصدري"}</th>
                  <th className="px-4 py-3 text-start">{inv.occurred_at || "تاريخ ووقت الحركة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredMovements.map((mov) => {
                  const isPositive = mov.quantity > 0;
                  return (
                    <tr key={mov.id} className="hover:bg-muted/30 transition">
                      <td className="px-4 py-3">
                        <div className="font-mono font-semibold text-primary">{mov.movementNumber}</div>
                        <div className="text-[10px] text-muted-foreground flex items-center gap-1">
                          <Lock className="h-3 w-3 text-emerald-600" />
                          <span>موثق بنظام الحظر</span>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-semibold border ${getMovementBadge(mov.movementType)}`}>
                          {mov.movementType}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-mono font-medium text-foreground">{mov.itemCode}</div>
                        <div className="text-[11px] text-muted-foreground">{isAr ? mov.itemNameAr : mov.itemNameEn}</div>
                      </td>
                      <td className="px-4 py-3">
                        <span className="font-mono font-semibold text-foreground">{mov.warehouseCode}</span>
                        <span className="text-[10px] text-muted-foreground block font-mono">{mov.binCode}</span>
                      </td>
                      <td className="px-4 py-3 text-end font-mono font-bold">
                        <span className={isPositive ? "text-emerald-600" : "text-rose-600"}>
                          {isPositive ? `+${mov.quantity}` : mov.quantity}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-end font-mono text-muted-foreground">
                        {mov.unitCost.toFixed(2)} {inv.currency || "ر.س"}
                      </td>
                      <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                        {mov.totalCost.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                      </td>
                      <td className="px-4 py-3">
                        <span className="px-2 py-0.5 rounded bg-muted text-foreground font-mono text-[11px] border">
                          {mov.sourceDoc}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono text-muted-foreground text-[11px]">
                        {mov.occurredAt}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 3: Warehouses & Bins Directory */}
      {activeTab === "warehouses" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {INITIAL_WAREHOUSES.map((wh) => (
            <div key={wh.id} className="p-4 rounded-xl border bg-card shadow-sm hover:border-primary/40 transition">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-bold text-primary px-2 py-0.5 rounded bg-primary/10 border border-primary/20">
                      {wh.code}
                    </span>
                    <h3 className="font-bold text-sm text-foreground">{isAr ? wh.nameAr : wh.nameEn}</h3>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                    <Building2 className="h-3.5 w-3.5" />
                    <span>الموقع الجغرافي: {wh.city}</span>
                  </p>
                </div>
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                  wh.isTransit
                    ? "bg-amber-500/10 text-amber-600 border-amber-500/20"
                    : "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                }`}>
                  {wh.type}
                </span>
              </div>

              <div className="mt-4 pt-3 border-t grid grid-cols-2 gap-3 text-xs">
                <div className="p-2 rounded-lg bg-muted/40">
                  <span className="text-muted-foreground block text-[10px]">مواقع التخزين (Bins)</span>
                  <span className="text-base font-bold font-mono text-foreground">{wh.binCount} موقع</span>
                </div>
                <div className="p-2 rounded-lg bg-muted/40">
                  <span className="text-muted-foreground block text-[10px]">إجمالي الأصناف المودعة</span>
                  <span className="text-base font-bold font-mono text-primary">{wh.totalStockItems} صنف</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Tab 4: Transfer Orders */}
      {activeTab === "transfers" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="p-3.5 bg-muted/30 border-b flex items-center justify-between">
            <span className="text-xs font-semibold text-foreground flex items-center gap-2">
              <ArrowRightLeft className="h-4 w-4 text-primary" />
              أوامر النقل والترحيل الداخلي بين المستودعات
            </span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">{inv.transfer_number || "رقم أمر النقل"}</th>
                  <th className="px-4 py-3 text-start">{inv.from_warehouse || "من مستودع"}</th>
                  <th className="px-4 py-3 text-start">{inv.to_warehouse || "إلى مستودع"}</th>
                  <th className="px-4 py-3 text-end">عدد الأصناف</th>
                  <th className="px-4 py-3 text-end">إجمالي الكميات</th>
                  <th className="px-4 py-3 text-center">{inv.status || "الحالة"}</th>
                  <th className="px-4 py-3 text-start">ملاحظات الشحنة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {transfers.map((trf) => (
                  <tr key={trf.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-bold text-primary">{trf.orderNumber}</td>
                    <td className="px-4 py-3 font-semibold text-foreground">{trf.fromWarehouse}</td>
                    <td className="px-4 py-3 font-semibold text-foreground">{trf.toWarehouse}</td>
                    <td className="px-4 py-3 text-end font-mono">{trf.itemsCount}</td>
                    <td className="px-4 py-3 text-end font-mono font-bold">{trf.totalQuantity}</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                        trf.status === "RECEIVED"
                          ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                          : trf.status === "DISPATCHED"
                          ? "bg-blue-500/10 text-blue-600 border-blue-500/20"
                          : "bg-amber-500/10 text-amber-600 border-amber-500/20"
                      }`}>
                        {trf.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{trf.notes}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 5: Stock Adjustments & Counts */}
      {activeTab === "adjustments" && (
        <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
          <div className="p-3.5 bg-muted/30 border-b flex items-center justify-between">
            <span className="text-xs font-semibold text-foreground flex items-center gap-2">
              <SlidersHorizontal className="h-4 w-4 text-primary" />
              تسويات الجرد الدورية وعمليات المطابقة الفعلية
            </span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-start">
              <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                <tr>
                  <th className="px-4 py-3 text-start">رقم التسوية</th>
                  <th className="px-4 py-3 text-start">{inv.warehouse || "المستودع"}</th>
                  <th className="px-4 py-3 text-start">نوع التسوية</th>
                  <th className="px-4 py-3 text-start">رمز السبب</th>
                  <th className="px-4 py-3 text-end">فارق التقييم</th>
                  <th className="px-4 py-3 text-center">{inv.status || "الحالة"}</th>
                  <th className="px-4 py-3 text-start">تاريخ التسوية</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {INITIAL_ADJUSTMENTS.map((adj) => (
                  <tr key={adj.id} className="hover:bg-muted/30 transition">
                    <td className="px-4 py-3 font-mono font-bold text-primary">{adj.adjustmentNumber}</td>
                    <td className="px-4 py-3 font-semibold text-foreground">{adj.warehouse}</td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-0.5 rounded bg-muted text-foreground font-mono text-[10px]">
                        {adj.adjustmentType}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground text-[11px]">{adj.reasonCode}</td>
                    <td className="px-4 py-3 text-end font-mono font-bold">
                      <span className={adj.totalVarianceValue >= 0 ? "text-emerald-600" : "text-rose-600"}>
                        {adj.totalVarianceValue >= 0 ? `+${adj.totalVarianceValue}` : adj.totalVarianceValue} {inv.currency || "ر.س"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold bg-purple-500/10 text-purple-600 border border-purple-500/20">
                        {adj.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{adj.createdAt}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 6: Costing & Valuation Engine */}
      {activeTab === "costing" && (
        <div className="space-y-6">
          {/* Methodology Comparison Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Card 1: Moving Weighted Average */}
            <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm relative overflow-hidden">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-600">
                    <Calculator className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-foreground">
                      {inv.method_avg || "متوسط التكلفة المرجح (WEIGHTED_AVG)"}
                    </h4>
                    <span className="text-[10px] text-emerald-600 font-semibold bg-emerald-500/10 px-1.5 py-0.2 rounded border border-emerald-500/20">
                      محرك التكلفة النشط الافتراضي
                    </span>
                  </div>
                </div>
              </div>
              <p className="mt-3 text-[11px] text-muted-foreground leading-relaxed">
                يتم احتساب التكلفة ديناميكياً مع كل حركة استلام توريد بقسمة إجمالي القيمة التراكمية على الرصيد المتوفر، وصرف المبيعات بأحدث متوسط تكلفة مرجح لحظي.
              </p>
              <div className="mt-3 pt-3 border-t flex items-center justify-between text-[11px]">
                <span className="text-muted-foreground">صيغة التحديث:</span>
                <span className="font-mono font-semibold text-foreground dir-ltr">
                  (Val_old + In_val) / (Qty_old + In_qty)
                </span>
              </div>
            </div>

            {/* Card 2: FIFO Discrete Layers */}
            <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm relative overflow-hidden">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600">
                    <Layers className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-foreground">
                      {inv.method_fifo || "الوارد أولاً يصرف أولاً (FIFO)"}
                    </h4>
                    <span className="text-[10px] text-blue-600 font-semibold bg-blue-500/10 px-1.5 py-0.2 rounded border border-blue-500/20">
                      تتبع الطبقات المنفصلة (Layers)
                    </span>
                  </div>
                </div>
              </div>
              <p className="mt-3 text-[11px] text-muted-foreground leading-relaxed">
                يتم تسجيل كل دفعة استلام مشتريات كطبقة تكلفة مستقلة. عند صرف البضاعة، يتم استنفاد أقدم طبقة أولاً بدقة وإقفالها عند استهلاك كامل كميتها.
              </p>
              <div className="mt-3 pt-3 border-t flex items-center justify-between text-[11px]">
                <span className="text-muted-foreground">طبقات التكلفة النشطة:</span>
                <span className="font-mono font-bold text-blue-600">
                  {costLayers.filter((l) => l.status === "ACTIVE").length} طبقة نشطة
                </span>
              </div>
            </div>

            {/* Card 3: Standard Costing */}
            <div className="p-4 rounded-xl border bg-card text-card-foreground shadow-sm relative overflow-hidden">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="p-2 rounded-lg bg-purple-500/10 text-purple-600">
                    <Coins className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-foreground">
                      {inv.method_std || "التكلفة المعيارية (STANDARD)"}
                    </h4>
                    <span className="text-[10px] text-purple-600 font-semibold bg-purple-500/10 px-1.5 py-0.2 rounded border border-purple-500/20">
                      معايير التكلفة وفروقات الأسعار
                    </span>
                  </div>
                </div>
              </div>
              <p className="mt-3 text-[11px] text-muted-foreground leading-relaxed">
                تحديد تكلفة ثابتة مسبقة للأصناف المصنعة أو المستوردة، وترحيل أي انحرافات في أسعار التوريد الفعلية تلقائياً إلى حساب فروقات الشراء (PPV).
              </p>
              <div className="mt-3 pt-3 border-t flex items-center justify-between text-[11px]">
                <span className="text-muted-foreground">تسجيل الفروقات:</span>
                <span className="font-mono font-bold text-purple-600">PPV Auto-Reconciled</span>
              </div>
            </div>
          </div>

          {/* FIFO Layers Section */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="p-3.5 bg-muted/30 border-b flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Layers className="h-4 w-4 text-primary" />
                <span className="text-xs font-bold text-foreground">
                  {inv.cost_layers_title || "طبقات تكلفة الوارد أولاً يصرف أولاً (FIFO Layers)"}
                </span>
              </div>
              <span className="text-[11px] text-muted-foreground font-mono">
                {filteredCostLayers.length} {isAr ? "طبقة مسجلة" : "layers tracked"}
              </span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                  <tr>
                    <th className="px-4 py-3 text-start">{inv.item || "الصنف"}</th>
                    <th className="px-4 py-3 text-start">{inv.warehouse || "المستودع"}</th>
                    <th className="px-4 py-3 text-start">تاريخ الاستلام</th>
                    <th className="px-4 py-3 text-end">{inv.original_qty || "الكمية الأصلية"}</th>
                    <th className="px-4 py-3 text-center">{inv.remaining_qty || "المتبقي ونسبة الاستهلاك"}</th>
                    <th className="px-4 py-3 text-end">{inv.unit_cost || "تكلفة الوحدة"}</th>
                    <th className="px-4 py-3 text-end">قيمة الرصيد المتبقي</th>
                    <th className="px-4 py-3 text-start">{inv.source_doc || "المستند المصدري"}</th>
                    <th className="px-4 py-3 text-center">{inv.status || "الحالة"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredCostLayers.map((layer) => {
                    const pctRemaining = layer.originalQuantity > 0 
                      ? Math.round((layer.remainingQuantity / layer.originalQuantity) * 100) 
                      : 0;
                    return (
                      <tr key={layer.id} className="hover:bg-muted/30 transition">
                        <td className="px-4 py-3">
                          <div className="font-mono font-semibold text-primary">{layer.itemCode}</div>
                          <div className="text-foreground font-medium">{isAr ? layer.itemNameAr : layer.itemNameEn}</div>
                        </td>
                        <td className="px-4 py-3 font-semibold text-foreground">{layer.warehouseCode}</td>
                        <td className="px-4 py-3 font-mono text-muted-foreground text-[11px]">{layer.layerDate}</td>
                        <td className="px-4 py-3 text-end font-mono font-semibold">{layer.originalQuantity}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-center gap-2">
                            <div className="w-24 bg-muted h-2 rounded-full overflow-hidden border border-border">
                              <div
                                className={`h-full rounded-full ${
                                  pctRemaining > 50
                                    ? "bg-emerald-500"
                                    : pctRemaining > 15
                                    ? "bg-amber-500"
                                    : "bg-rose-500"
                                }`}
                                style={{ width: `${pctRemaining}%` }}
                              />
                            </div>
                            <span className="font-mono text-[10px] font-bold text-foreground w-12 text-end">
                              {layer.remainingQuantity} ({pctRemaining}%)
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-end font-mono font-bold">
                          {layer.unitCost.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                        </td>
                        <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                          {layer.totalRemainingValue.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                        </td>
                        <td className="px-4 py-3 font-mono text-primary text-[11px]">{layer.sourceDoc}</td>
                        <td className="px-4 py-3 text-center">
                          <span
                            className={`inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                              layer.status === "ACTIVE"
                                ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                                : "bg-muted text-muted-foreground border-border"
                            }`}
                          >
                            {layer.status === "ACTIVE" ? (isAr ? "نشطة" : "ACTIVE") : (isAr ? "مستنفدة" : "EXHAUSTED")}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {/* Inventory Revaluations Section */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="p-3.5 bg-muted/30 border-b flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Coins className="h-4 w-4 text-primary" />
                <span className="text-xs font-bold text-foreground">
                  {inv.revaluation_title || "إعادة تقييم المخزون وتعديل التكلفة"}
                </span>
                <span className="text-[10px] text-muted-foreground">
                  (سجل التدقيق المحاسبي لفروقات التقييم)
                </span>
              </div>
              <button
                onClick={() => setShowRevalModal(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm w-fit"
              >
                <Plus className="h-3.5 w-3.5" />
                {inv.new_revaluation || "إعادة تقييم صنف"}
              </button>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/60 text-muted-foreground font-semibold border-b">
                  <tr>
                    <th className="px-4 py-3 text-start">رقم المستند</th>
                    <th className="px-4 py-3 text-start">{inv.item || "الصنف"}</th>
                    <th className="px-4 py-3 text-start">{inv.warehouse || "المستودع"}</th>
                    <th className="px-4 py-3 text-end">الكمية الخاضعة للتقييم</th>
                    <th className="px-4 py-3 text-end">{inv.old_unit_cost || "التكلفة السابقة"}</th>
                    <th className="px-4 py-3 text-end">{inv.new_unit_cost || "التكلفة الجديدة"}</th>
                    <th className="px-4 py-3 text-end">{inv.variance || "فارق التقييم الإجمالي"}</th>
                    <th className="px-4 py-3 text-start">{inv.revaluation_reason || "السبب المحاسبي"}</th>
                    <th className="px-4 py-3 text-start">تاريخ التنفيذ</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredRevaluations.map((rev) => {
                    const isPositive = rev.varianceAmount >= 0;
                    return (
                      <tr key={rev.id} className="hover:bg-muted/30 transition">
                        <td className="px-4 py-3 font-mono font-bold text-primary">{rev.revaluationNumber}</td>
                        <td className="px-4 py-3">
                          <div className="font-mono font-semibold text-foreground">{rev.itemCode}</div>
                          <div className="text-[11px] text-muted-foreground">{rev.itemNameAr}</div>
                        </td>
                        <td className="px-4 py-3 font-semibold text-foreground">{rev.warehouseCode}</td>
                        <td className="px-4 py-3 text-end font-mono font-semibold">{rev.quantity}</td>
                        <td className="px-4 py-3 text-end font-mono text-muted-foreground">
                          {rev.oldUnitCost.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                        </td>
                        <td className="px-4 py-3 text-end font-mono font-bold text-foreground">
                          {rev.newUnitCost.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                        </td>
                        <td className="px-4 py-3 text-end font-mono font-bold">
                          <span
                            className={`inline-flex items-center gap-1 ${
                              isPositive ? "text-emerald-600" : "text-rose-600"
                            }`}
                          >
                            {isPositive ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                            {isPositive ? `+${rev.varianceAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}` : rev.varianceAmount.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-muted-foreground">{rev.reason}</td>
                        <td className="px-4 py-3 font-mono text-muted-foreground text-[11px]">{rev.performedAt}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Inventory Revaluation Modal */}
      {showRevalModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg rounded-xl border bg-card p-6 shadow-2xl animate-in zoom-in-95">
            <div className="flex items-center justify-between pb-3 border-b">
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-primary/10 text-primary">
                  <Coins className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="font-bold text-base text-foreground">
                    {inv.new_revaluation || "إعادة تقييم المخزون وتعديل التكلفة"}
                  </h3>
                  <p className="text-[11px] text-muted-foreground">
                    تعديل تكلفة الوحدة دفترياً وترحيل فارق التقييم مع حفظ سجل التدقيق
                  </p>
                </div>
              </div>
              <button
                onClick={() => setShowRevalModal(false)}
                className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handlePostRevaluation} className="space-y-4 mt-4 text-xs">
              {/* Item Selection */}
              <div>
                <label className="font-medium text-foreground block mb-1">{inv.item || "الصنف المراد تقييمه"}</label>
                <select
                  value={revalItemCode}
                  onChange={(e) => setRevalItemCode(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  {balances.map((b) => (
                    <option key={b.id} value={b.itemCode}>
                      {b.itemCode} - {isAr ? b.nameAr : b.nameEn}
                    </option>
                  ))}
                </select>
              </div>

              {/* Warehouse Selection */}
              <div>
                <label className="font-medium text-foreground block mb-1">{inv.warehouse || "المستودع"}</label>
                <select
                  value={revalWarehouse}
                  onChange={(e) => setRevalWarehouse(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="WH-MAIN">WH-MAIN (الرياض المركزي)</option>
                  <option value="WH-DAMMAM">WH-DAMMAM (فرع الدمام)</option>
                  <option value="WH-COLD">WH-COLD (مستودع التبريد)</option>
                </select>
              </div>

              {/* Current Status Box (Readonly) */}
              <div className="grid grid-cols-3 gap-2.5 p-3 rounded-lg bg-muted/40 border text-center">
                <div>
                  <span className="text-[10px] text-muted-foreground block">الرصيد الفعلي الحالي</span>
                  <span className="font-mono font-bold text-foreground text-sm">{revalCurrentQty}</span>
                </div>
                <div>
                  <span className="text-[10px] text-muted-foreground block">{inv.old_unit_cost || "التكلفة السابقة"}</span>
                  <span className="font-mono font-bold text-foreground text-sm">
                    {revalCurrentCost.toFixed(2)} {inv.currency || "ر.س"}
                  </span>
                </div>
                <div>
                  <span className="text-[10px] text-muted-foreground block">القيمة الإجمالية السابقة</span>
                  <span className="font-mono font-bold text-foreground text-sm">
                    {(revalCurrentQty * revalCurrentCost).toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                  </span>
                </div>
              </div>

              {/* New Unit Cost Input */}
              <div>
                <label className="font-medium text-foreground block mb-1">
                  {inv.new_unit_cost || "التكلفة الجديدة للوحدة"} ({inv.currency || "ر.س"})
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={revalNewCost}
                  onChange={(e) => setRevalNewCost(parseFloat(e.target.value) || 0)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono font-bold focus:outline-none focus:ring-1 focus:ring-primary"
                  required
                />
              </div>

              {/* Live Variance Calculation Display */}
              <div className="p-3 rounded-lg border bg-card flex items-center justify-between">
                <div>
                  <span className="text-[11px] text-muted-foreground block">{inv.variance || "فارق التقييم الإجمالي"}</span>
                  <span className="text-[10px] text-muted-foreground">
                    ({revalCurrentQty} وحدة × {revalUnitVariance >= 0 ? `+${revalUnitVariance.toFixed(2)}` : revalUnitVariance.toFixed(2)} ر.س)
                  </span>
                </div>
                <div className="text-end">
                  <span
                    className={`font-mono text-base font-bold flex items-center justify-end gap-1 ${
                      revalTotalVariance >= 0 ? "text-emerald-600" : "text-rose-600"
                    }`}
                  >
                    {revalTotalVariance >= 0 ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />}
                    {revalTotalVariance >= 0 ? `+${revalTotalVariance.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })}` : revalTotalVariance.toLocaleString(isAr ? "ar-SA" : "en-US", { minimumFractionDigits: 2 })} {inv.currency || "ر.س"}
                  </span>
                  <span className="text-[10px] font-medium text-muted-foreground block">
                    {revalTotalVariance >= 0 ? "أرباح تقييم دفتري" : "خسائر هبوط قيمة المخزون"}
                  </span>
                </div>
              </div>

              {/* Reason Input */}
              <div>
                <label className="font-medium text-foreground block mb-1">
                  {inv.revaluation_reason || "سبب التعديل المحاسبي"}
                </label>
                <input
                  type="text"
                  placeholder="مثال: تحديث تكلفة التوريد والرسوم، تصحيح أسعار المورد..."
                  value={revalReason}
                  onChange={(e) => setRevalReason(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  required
                />
              </div>

              {/* Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setShowRevalModal(false)}
                  className="px-4 py-2 rounded-lg border hover:bg-muted text-foreground transition"
                >
                  {inv.cancel || "إلغاء"}
                </button>
                <button
                  type="submit"
                  disabled={revalCurrentQty <= 0}
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition shadow-sm disabled:opacity-50"
                >
                  {inv.submit_revaluation || "اعتماد إعادة التقييم وتحديث الرصيد"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Post Movement Modal */}
      {showMovementModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg rounded-xl border bg-card p-6 shadow-2xl animate-in zoom-in-95">
            <div className="flex items-center justify-between pb-3 border-b">
              <div className="flex items-center gap-2">
                <Boxes className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-base text-foreground">
                  {inv.post_movement_dialog || "تسجيل حركة وارد / صادر فورية"}
                </h3>
              </div>
              <button
                onClick={() => setShowMovementModal(false)}
                className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handlePostMovement} className="space-y-4 mt-4 text-xs">
              {/* Movement Type */}
              <div>
                <label className="font-medium text-foreground block mb-1">
                  {inv.movement_type || "نوع الحركة المخزنية"}
                </label>
                <select
                  value={movType}
                  onChange={(e) => setMovType(e.target.value as any)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="PURCHASE_RECEIPT">{inv.type_purchase || "استلام مشتريات (PURCHASE_RECEIPT)"}</option>
                  <option value="SALES_ISSUE">{inv.type_sales || "صرف مبيعات (SALES_ISSUE)"}</option>
                  <option value="TRANSFER_IN">{inv.type_transfer_in || "وارد تحويل (TRANSFER_IN)"}</option>
                  <option value="TRANSFER_OUT">{inv.type_transfer_out || "صادر تحويل (TRANSFER_OUT)"}</option>
                  <option value="ADJUSTMENT_IN">{inv.type_adj_in || "تسوية إضافة (ADJUSTMENT_IN)"}</option>
                  <option value="ADJUSTMENT_OUT">{inv.type_adj_out || "تسوية خصم (ADJUSTMENT_OUT)"}</option>
                  <option value="OPENING">{inv.type_opening || "رصيد افتتاحي (OPENING)"}</option>
                </select>
              </div>

              {/* Item Selection */}
              <div>
                <label className="font-medium text-foreground block mb-1">{inv.item || "الصنف"}</label>
                <select
                  value={movItemCode}
                  onChange={(e) => setMovItemCode(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  {balances.map((b) => (
                    <option key={b.id} value={b.itemCode}>
                      {b.itemCode} - {isAr ? b.nameAr : b.nameEn}
                    </option>
                  ))}
                </select>
              </div>

              {/* Warehouse */}
              <div>
                <label className="font-medium text-foreground block mb-1">{inv.warehouse || "المستودع"}</label>
                <select
                  value={movWarehouse}
                  onChange={(e) => setMovWarehouse(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="WH-MAIN">WH-MAIN (الرياض المركزي)</option>
                  <option value="WH-DAMMAM">WH-DAMMAM (فرع الدمام)</option>
                  <option value="WH-COLD">WH-COLD (مستودع التبريد)</option>
                </select>
              </div>

              {/* Quantity & Unit Cost */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="font-medium text-foreground block mb-1">{inv.quantity || "الكمية"}</label>
                  <input
                    type="number"
                    min="1"
                    value={movQty}
                    onChange={(e) => setMovQty(parseFloat(e.target.value) || 1)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                    required
                  />
                </div>
                <div>
                  <label className="font-medium text-foreground block mb-1">
                    {inv.unit_cost || "تكلفة الوحدة"} ({inv.currency || "ر.س"})
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={movUnitCost}
                    onChange={(e) => setMovUnitCost(parseFloat(e.target.value) || 0)}
                    className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                    required
                  />
                </div>
              </div>

              {/* Source Document Reference */}
              <div>
                <label className="font-medium text-foreground block mb-1">{inv.source_doc || "المستند المصدري"}</label>
                <input
                  type="text"
                  placeholder="PO-2026-XXXX أو SO-2026-XXXX"
                  value={movDoc}
                  onChange={(e) => setMovDoc(e.target.value)}
                  className="w-full rounded-lg border bg-background px-3 py-2 text-foreground font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                  required
                />
              </div>

              {/* Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setShowMovementModal(false)}
                  className="px-4 py-2 rounded-lg border hover:bg-muted text-foreground transition"
                >
                  {inv.cancel || "إلغاء"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-semibold hover:bg-primary/90 transition shadow-sm"
                >
                  {inv.submit || "تنفيذ وترحيل الحركة"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

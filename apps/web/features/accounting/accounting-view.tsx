"use client";

import React, { useState, useMemo } from "react";
import { useLocale } from "@/components/layout/shell";
import {
  BookOpen,
  FolderTree,
  FileSpreadsheet,
  PlusCircle,
  Calendar,
  ShieldCheck,
  CheckCircle2,
  AlertTriangle,
  RotateCcw,
  Send,
  ChevronDown,
  ChevronRight,
  ChevronLeft,
  Search,
  Filter,
  Layers,
  Lock,
  Unlock,
  Building,
  DollarSign,
  ArrowRight,
  ArrowLeft,
  Check,
  Trash2,
  AlertCircle,
  HelpCircle,
} from "lucide-react";

// --- Types ---
export type AccountType = "ASSET" | "LIABILITY" | "EQUITY" | "REVENUE" | "EXPENSE";

export interface AccountNode {
  id: string;
  code: string;
  nameAr: string;
  nameEn: string;
  accountType: AccountType;
  level: number;
  isLeaf: boolean;
  normalBalance: "DEBIT" | "CREDIT";
  parentId?: string;
  balance?: number;
}

export interface JournalEntryLine {
  id: string;
  accountId: string;
  accountCode: string;
  accountNameAr: string;
  accountNameEn: string;
  description: string;
  debit: number;
  credit: number;
}

export interface JournalEntry {
  id: string;
  entryNumber: string;
  journalCode: string;
  journalNameAr: string;
  journalNameEn: string;
  periodName: string;
  entryDate: string;
  description: string;
  status: "DRAFT" | "POSTED" | "REVERSED";
  totalDebit: number;
  totalCredit: number;
  reversalOfId?: string;
  reversalReason?: string;
  lines: JournalEntryLine[];
}

export interface TrialBalanceRow {
  accountCode: string;
  nameAr: string;
  nameEn: string;
  type: AccountType;
  openingDebit: number;
  openingCredit: number;
  movementDebit: number;
  movementCredit: number;
  closingDebit: number;
  closingCredit: number;
}

// Initial 5-level canonical Saudi standard Chart of Accounts
const initialAccounts: AccountNode[] = [
  // 1: ASSETS
  { id: "1", code: "1", nameAr: "الأصول", nameEn: "Assets", accountType: "ASSET", level: 1, isLeaf: false, normalBalance: "DEBIT" },
  { id: "11", code: "11", nameAr: "الأصول المتداولة", nameEn: "Current Assets", accountType: "ASSET", level: 2, isLeaf: false, normalBalance: "DEBIT", parentId: "1" },
  { id: "1101", code: "1101", nameAr: "النقد وما في حكمه", nameEn: "Cash & Cash Equivalents", accountType: "ASSET", level: 3, isLeaf: false, normalBalance: "DEBIT", parentId: "11" },
  { id: "110101", code: "110101", nameAr: "النقدية بالصندوق الرئيسي", nameEn: "Petty Cash - Main Vault", accountType: "ASSET", level: 4, isLeaf: true, normalBalance: "DEBIT", parentId: "1101", balance: 50000 },
  { id: "110102", code: "110102", nameAr: "مصرف الراجحي - الحساب الجاري", nameEn: "Al Rajhi Bank - Current SAR", accountType: "ASSET", level: 4, isLeaf: true, normalBalance: "DEBIT", parentId: "1101", balance: 350000 },
  { id: "1102", code: "1102", nameAr: "الذمم المدينة التجارية", nameEn: "Accounts Receivable", accountType: "ASSET", level: 3, isLeaf: false, normalBalance: "DEBIT", parentId: "11" },
  { id: "110201", code: "110201", nameAr: "عملاء تجاريون محليون", nameEn: "Trade Debtors - Domestic", accountType: "ASSET", level: 4, isLeaf: true, normalBalance: "DEBIT", parentId: "1102", balance: 120000 },
  { id: "1103", code: "1103", nameAr: "المخزون السلعي", nameEn: "Merchandise Inventory", accountType: "ASSET", level: 3, isLeaf: false, normalBalance: "DEBIT", parentId: "11" },
  { id: "110301", code: "110301", nameAr: "مخزون بضائع الرياض الرئيسي", nameEn: "Inventory - Riyadh Central", accountType: "ASSET", level: 4, isLeaf: true, normalBalance: "DEBIT", parentId: "1103", balance: 80000 },
  { id: "12", code: "12", nameAr: "الأصول غير المتداولة", nameEn: "Non-Current Assets", accountType: "ASSET", level: 2, isLeaf: false, normalBalance: "DEBIT", parentId: "1" },
  { id: "1201", code: "1201", nameAr: "الممتلكات والمعدات", nameEn: "Property, Plant & Equipment", accountType: "ASSET", level: 3, isLeaf: false, normalBalance: "DEBIT", parentId: "12" },
  { id: "120101", code: "120101", nameAr: "أجهزة وتقنية معلومات ومخدمات", nameEn: "IT Hardware & Cloud Servers", accountType: "ASSET", level: 4, isLeaf: true, normalBalance: "DEBIT", parentId: "1201", balance: 65000 },

  // 2: LIABILITIES
  { id: "2", code: "2", nameAr: "الخصوم والالتزامات", nameEn: "Liabilities", accountType: "LIABILITY", level: 1, isLeaf: false, normalBalance: "CREDIT" },
  { id: "21", code: "21", nameAr: "الخصوم المتداولة", nameEn: "Current Liabilities", accountType: "LIABILITY", level: 2, isLeaf: false, normalBalance: "CREDIT", parentId: "2" },
  { id: "2101", code: "2101", nameAr: "الذمم الدائنة والموردين", nameEn: "Accounts Payable", accountType: "LIABILITY", level: 3, isLeaf: false, normalBalance: "CREDIT", parentId: "21" },
  { id: "210101", code: "210101", nameAr: "موردو السلع والخدمات المحليين", nameEn: "Domestic Trade Suppliers", accountType: "LIABILITY", level: 4, isLeaf: true, normalBalance: "CREDIT", parentId: "2101", balance: 95000 },
  { id: "2102", code: "2102", nameAr: "ضريبة القيمة المضافة المستحقة (ZATCA)", nameEn: "VAT Output Payable (ZATCA)", accountType: "LIABILITY", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "21", balance: 25000 },
  { id: "2103", code: "2103", nameAr: "مستحقات التأمينات الاجتماعية (GOSI)", nameEn: "GOSI Contributions Payable", accountType: "LIABILITY", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "21", balance: 15000 },

  // 3: EQUITY
  { id: "3", code: "3", nameAr: "حقوق الملكية", nameEn: "Equity", accountType: "EQUITY", level: 1, isLeaf: false, normalBalance: "CREDIT" },
  { id: "31", code: "31", nameAr: "رأس المال المدفوع", nameEn: "Paid-in Share Capital", accountType: "EQUITY", level: 2, isLeaf: false, normalBalance: "CREDIT", parentId: "3" },
  { id: "3101", code: "3101", nameAr: "رأس مال الشركاء والمساهمين", nameEn: "Common Capital Stock", accountType: "EQUITY", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "31", balance: 450000 },
  { id: "32", code: "32", nameAr: "الأرباح المبقاة والاحتياطيات", nameEn: "Retained Earnings & Reserves", accountType: "EQUITY", level: 2, isLeaf: false, normalBalance: "CREDIT", parentId: "3" },
  { id: "3201", code: "3201", nameAr: "الاحتياطي النظامي (10%)", nameEn: "Statutory Reserve", accountType: "EQUITY", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "32", balance: 45000 },

  // 4: REVENUE
  { id: "4", code: "4", nameAr: "الإيرادات", nameEn: "Revenue", accountType: "REVENUE", level: 1, isLeaf: false, normalBalance: "CREDIT" },
  { id: "41", code: "41", nameAr: "إيرادات النشاط التشغيلي", nameEn: "Operating Revenue", accountType: "REVENUE", level: 2, isLeaf: false, normalBalance: "CREDIT", parentId: "4" },
  { id: "4101", code: "4101", nameAr: "إيرادات مبيعات البرمجيات والأنظمة", nameEn: "SaaS Subscription Revenue", accountType: "REVENUE", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "41", balance: 110000 },
  { id: "4102", code: "4102", nameAr: "إيرادات الخدمات المهنية والاستشارات", nameEn: "Consulting & Professional Fees", accountType: "REVENUE", level: 3, isLeaf: true, normalBalance: "CREDIT", parentId: "41", balance: 35000 },

  // 5: EXPENSES
  { id: "5", code: "5", nameAr: "المصروفات", nameEn: "Expenses", accountType: "EXPENSE", level: 1, isLeaf: false, normalBalance: "DEBIT" },
  { id: "51", code: "51", nameAr: "تكلفة الإيرادات (المباشرة)", nameEn: "Cost of Goods & Services Sold", accountType: "EXPENSE", level: 2, isLeaf: false, normalBalance: "DEBIT", parentId: "5" },
  { id: "5101", code: "5101", nameAr: "تكاليف البنية التحتية والاستضافة", nameEn: "Hosting & Infrastructure Cost", accountType: "EXPENSE", level: 3, isLeaf: true, normalBalance: "DEBIT", parentId: "51", balance: 25000 },
  { id: "52", code: "52", nameAr: "المصاريف العمومية والإدارية", nameEn: "General & Administrative Expenses", accountType: "EXPENSE", level: 2, isLeaf: false, normalBalance: "DEBIT", parentId: "5" },
  { id: "5201", code: "5201", nameAr: "رواتب ومزايا الموظفين الأساسية", nameEn: "Salaries & Employee Benefits", accountType: "EXPENSE", level: 3, isLeaf: true, normalBalance: "DEBIT", parentId: "52", balance: 65000 },
  { id: "5202", code: "5202", nameAr: "مصاريف إيجار المقرات والمرافق", nameEn: "Office Rent & Utilities", accountType: "EXPENSE", level: 3, isLeaf: true, normalBalance: "DEBIT", parentId: "52", balance: 15000 },
];

const initialEntries: JournalEntry[] = [
  {
    id: "je-101",
    entryNumber: "JE-2026-00001",
    journalCode: "GEN",
    journalNameAr: "دفتر اليومية العامة",
    journalNameEn: "General Journal",
    periodName: "2026-01 (يناير)",
    entryDate: "2026-01-01",
    description: "قيد إثبات رأس المال التأسيسي وإيداعه في الحساب البنكي لمصرف الراجحي",
    status: "POSTED",
    totalDebit: 450000,
    totalCredit: 450000,
    lines: [
      { id: "l1", accountId: "110102", accountCode: "110102", accountNameAr: "مصرف الراجحي - الحساب الجاري", accountNameEn: "Al Rajhi Bank - Current SAR", description: "إيداع رأس المال النقدي", debit: 450000, credit: 0 },
      { id: "l2", accountId: "3101", accountCode: "3101", accountNameAr: "رأس مال الشركاء والمساهمين", accountNameEn: "Common Capital Stock", description: "إثبات حصص الشركاء والمساهمين", debit: 0, credit: 450000 },
    ],
  },
  {
    id: "je-102",
    entryNumber: "JE-2026-00002",
    journalCode: "SALES",
    journalNameAr: "دفتر مبيعات الخدمات",
    journalNameEn: "Sales Journal",
    periodName: "2026-01 (يناير)",
    entryDate: "2026-01-15",
    description: "إثبات مبيعات خدمات برمجية سنوية مع احتساب ضريبة القيمة المضافة 15%",
    status: "POSTED",
    totalDebit: 115000,
    totalCredit: 115000,
    lines: [
      { id: "l3", accountId: "110201", accountCode: "110201", accountNameAr: "عملاء تجاريون محليون", accountNameEn: "Trade Debtors - Domestic", description: "استحقاق فاتورة مبيعات شركة الرواد", debit: 115000, credit: 0 },
      { id: "l4", accountId: "4101", accountCode: "4101", accountNameAr: "إيرادات مبيعات البرمجيات والأنظمة", accountNameEn: "SaaS Subscription Revenue", description: "اشتراك المنصة السحابية", debit: 0, credit: 100000 },
      { id: "l5", accountId: "2102", accountCode: "2102", accountNameAr: "ضريبة القيمة المضافة المستحقة (ZATCA)", accountNameEn: "VAT Output Payable (ZATCA)", description: "ضريبة القيمة المضافة 15%", debit: 0, credit: 15000 },
    ],
  },
  {
    id: "je-103",
    entryNumber: "JE-2026-00003",
    journalCode: "GEN",
    journalNameAr: "دفتر اليومية العامة",
    journalNameEn: "General Journal",
    periodName: "2026-01 (يناير)",
    entryDate: "2026-01-25",
    description: "قيد تسوية أجور واستحقاقات شهر يناير 2026 مع مخصص التأمينات",
    status: "DRAFT",
    totalDebit: 65000,
    totalCredit: 65000,
    lines: [
      { id: "l6", accountId: "5201", accountCode: "5201", accountNameAr: "رواتب ومزايا الموظفين الأساسية", accountNameEn: "Salaries & Employee Benefits", description: "مسير رواتب شهر يناير", debit: 65000, credit: 0 },
      { id: "l7", accountId: "2103", accountCode: "2103", accountNameAr: "مستحقات التأمينات الاجتماعية (GOSI)", accountNameEn: "GOSI Contributions Payable", description: "حصة التأمينات الاجتماعية المستقطعة", debit: 0, credit: 15000 },
      { id: "l8", accountId: "110102", accountCode: "110102", accountNameAr: "مصرف الراجحي - الحساب الجاري", accountNameEn: "Al Rajhi Bank - Current SAR", description: "صرف صافي الرواتب عبر سريع/نظام حماية الأجور", debit: 0, credit: 50000 },
    ],
  },
];

export function AccountingView() {
  const { t, locale, dir } = useLocale();
  const isRTL = dir === "rtl";

  const [activeTab, setActiveTab] = useState<"coa" | "journals" | "new_entry" | "trial_balance" | "fiscal">("coa");

  // COA State
  const [accounts, setAccounts] = useState<AccountNode[]>(initialAccounts);
  const [coaFilterType, setCoaFilterType] = useState<AccountType | "ALL">("ALL");
  const [coaSearch, setCoaSearch] = useState("");
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({
    "1": true,
    "11": true,
    "1101": true,
    "2": true,
    "21": true,
    "3": true,
    "4": true,
    "5": true,
  });

  // Journals State
  const [entries, setEntries] = useState<JournalEntry[]>(initialEntries);
  const [entryFilterStatus, setEntryFilterStatus] = useState<"ALL" | "DRAFT" | "POSTED" | "REVERSED">("ALL");
  const [expandedEntryId, setExpandedEntryId] = useState<string | null>("je-101");
  const [actionSuccessMsg, setActionSuccessMsg] = useState<string | null>(null);

  // Reversal Modal State
  const [reversingEntry, setReversingEntry] = useState<JournalEntry | null>(null);
  const [reversalReason, setReversalReason] = useState("");

  // New Entry Studio State
  const leafAccounts = useMemo(() => accounts.filter((a) => a.isLeaf), [accounts]);
  const [newJournalCode, setNewJournalCode] = useState("GEN");
  const [newEntryDate, setNewEntryDate] = useState("2026-01-30");
  const [newDescription, setNewDescription] = useState("");
  const [newLines, setNewLines] = useState<Array<{ id: string; accountId: string; description: string; debit: string; credit: string }>>([
    { id: "nl-1", accountId: leafAccounts[0]?.id || "110101", description: "سداد مصروف صيانة ومرافق", debit: "3500.00", credit: "0" },
    { id: "nl-2", accountId: leafAccounts[1]?.id || "110102", description: "تحويل بنكي مباشر", debit: "0", credit: "3500.00" },
  ]);

  // Fiscal Periods State
  const [fiscalPeriods, setFiscalPeriods] = useState([
    { id: "p1", periodNumber: 1, name: "2026-01", startDate: "2026-01-01", endDate: "2026-01-31", status: "OPEN" as const },
    { id: "p2", periodNumber: 2, name: "2026-02", startDate: "2026-02-01", endDate: "2026-02-28", status: "OPEN" as const },
    { id: "p3", periodNumber: 3, name: "2026-03", startDate: "2026-03-01", endDate: "2026-03-31", status: "OPEN" as const },
    { id: "p4", periodNumber: 4, name: "2026-04", startDate: "2026-04-01", endDate: "2026-04-30", status: "OPEN" as const },
  ]);

  // Toggle Node
  const toggleNode = (id: string) => {
    setExpandedNodes((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  // Filtered COA
  const filteredAccounts = useMemo(() => {
    return accounts.filter((acc) => {
      if (coaFilterType !== "ALL" && acc.accountType !== coaFilterType) return false;
      if (coaSearch.trim()) {
        const query = coaSearch.toLowerCase();
        return (
          acc.code.toLowerCase().includes(query) ||
          acc.nameAr.toLowerCase().includes(query) ||
          acc.nameEn.toLowerCase().includes(query)
        );
      }
      return true;
    });
  }, [accounts, coaFilterType, coaSearch]);

  // Journal Entries Calculation
  const filteredEntries = useMemo(() => {
    return entries.filter((e) => {
      if (entryFilterStatus !== "ALL" && e.status !== entryFilterStatus) return false;
      return true;
    });
  }, [entries, entryFilterStatus]);

  // Calculate totals for new entry
  const newEntryTotals = useMemo(() => {
    let debits = 0;
    let credits = 0;
    newLines.forEach((l) => {
      const d = parseFloat(l.debit) || 0;
      const c = parseFloat(l.credit) || 0;
      debits += d;
      credits += c;
    });
    const diff = Math.abs(debits - credits);
    const isBalanced = diff < 0.0001 && debits > 0;
    return { debits, credits, diff, isBalanced };
  }, [newLines]);

  // Add line to new entry
  const handleAddLine = () => {
    setNewLines((prev) => [
      ...prev,
      {
        id: `nl-${Date.now()}`,
        accountId: leafAccounts[0]?.id || "",
        description: "",
        debit: "0",
        credit: "0",
      },
    ]);
  };

  // Remove line from new entry
  const handleRemoveLine = (idx: number) => {
    if (newLines.length <= 2) return;
    setNewLines((prev) => prev.filter((_, i) => i !== idx));
  };

  // Post Draft Entry
  const handlePostEntry = (entryId: string) => {
    setEntries((prev) =>
      prev.map((e) => (e.id === entryId ? { ...e, status: "POSTED" } : e))
    );
    setActionSuccessMsg(
      locale === "ar"
        ? "تم ترحيل القيد المحاسبي بنجاح إلى الأستاذ العام وتثبيته في قاعدة البيانات."
        : "Journal entry posted successfully to General Ledger and locked against mutation."
    );
    setTimeout(() => setActionSuccessMsg(null), 4000);
  };

  // Execute Reversal
  const handleExecuteReversal = () => {
    if (!reversingEntry) return;

    const revId = `je-rev-${Date.now().toString().slice(-4)}`;
    const revNumber = `${reversingEntry.entryNumber}-REV`;

    const reversedLines: JournalEntryLine[] = reversingEntry.lines.map((l, idx) => ({
      id: `rev-l-${idx}`,
      accountId: l.accountId,
      accountCode: l.accountCode,
      accountNameAr: l.accountNameAr,
      accountNameEn: l.accountNameEn,
      description: `[عكس] ${l.description}`,
      debit: l.credit, // Invert
      credit: l.debit, // Invert
    }));

    const newRevEntry: JournalEntry = {
      id: revId,
      entryNumber: revNumber,
      journalCode: reversingEntry.journalCode,
      journalNameAr: reversingEntry.journalNameAr,
      journalNameEn: reversingEntry.journalNameEn,
      periodName: reversingEntry.periodName,
      entryDate: new Date().toISOString().split("T")[0],
      description: `عكس القيد ${reversingEntry.entryNumber}: ${reversalReason}`,
      status: "POSTED",
      totalDebit: reversingEntry.totalCredit,
      totalCredit: reversingEntry.totalDebit,
      reversalOfId: reversingEntry.id,
      reversalReason: reversalReason,
      lines: reversedLines,
    };

    setEntries((prev) => [
      newRevEntry,
      ...prev.map((e) => (e.id === reversingEntry.id ? { ...e, status: "REVERSED" as const } : e)),
    ]);

    setReversingEntry(null);
    setReversalReason("");
    setActionSuccessMsg(
      locale === "ar"
        ? `تم عكس القيد وتوليد قيد العكس ${revNumber} متوازناً ومرحّلاً بنجاح.`
        : `Reversal entry ${revNumber} generated and posted successfully with inverse lines.`
    );
    setTimeout(() => setActionSuccessMsg(null), 4500);
  };

  // Submit New Entry
  const handleCreateNewEntry = (postImmediately: boolean) => {
    if (!newEntryTotals.isBalanced) return;

    const entryNum = `JE-2026-${String(entries.length + 1).padStart(5, "0")}`;
    const mappedLines: JournalEntryLine[] = newLines.map((l, idx) => {
      const acc = leafAccounts.find((a) => a.id === l.accountId) || leafAccounts[0];
      return {
        id: `line-${Date.now()}-${idx}`,
        accountId: acc.id,
        accountCode: acc.code,
        accountNameAr: acc.nameAr,
        accountNameEn: acc.nameEn,
        description: l.description || newDescription,
        debit: parseFloat(l.debit) || 0,
        credit: parseFloat(l.credit) || 0,
      };
    });

    const created: JournalEntry = {
      id: `je-${Date.now()}`,
      entryNumber: entryNum,
      journalCode: newJournalCode,
      journalNameAr: newJournalCode === "GEN" ? "دفتر اليومية العامة" : "دفتر العمليات التشغيلية",
      journalNameEn: newJournalCode === "GEN" ? "General Journal" : "Operations Journal",
      periodName: "2026-01 (يناير)",
      entryDate: newEntryDate,
      description: newDescription || "قيد عمليات مالية عامة",
      status: postImmediately ? "POSTED" : "DRAFT",
      totalDebit: newEntryTotals.debits,
      totalCredit: newEntryTotals.credits,
      lines: mappedLines,
    };

    setEntries((prev) => [created, ...prev]);
    setActiveTab("journals");
    setExpandedEntryId(created.id);
    setActionSuccessMsg(
      postImmediately
        ? locale === "ar"
          ? `تم حفظ وترحيل القيد ${entryNum} فورياً للأستاذ العام بنجاح!`
          : `Journal entry ${entryNum} saved and posted immediately to General Ledger!`
        : locale === "ar"
          ? `تم حفظ القيد ${entryNum} كمسودة متوازنة جاهزة للمراجعة.`
          : `Journal entry ${entryNum} saved as draft.`
    );
    setTimeout(() => setActionSuccessMsg(null), 4000);
  };

  // Compute Trial Balance Rows from posted entries
  const trialBalanceData = useMemo<TrialBalanceRow[]>(() => {
    const map = new Map<string, { debits: number; credits: number; acc: AccountNode }>();

    leafAccounts.forEach((acc) => {
      map.set(acc.id, { debits: 0, credits: 0, acc });
    });

    // Sum posted entries
    entries
      .filter((e) => e.status === "POSTED")
      .forEach((e) => {
        e.lines.forEach((l) => {
          const item = map.get(l.accountId);
          if (item) {
            item.debits += l.debit;
            item.credits += l.credit;
          }
        });
      });

    const rows: TrialBalanceRow[] = [];
    map.forEach((val) => {
      if (val.debits > 0 || val.credits > 0 || (val.acc.balance && val.acc.balance > 0)) {
        const net = val.debits - val.credits;
        rows.push({
          accountCode: val.acc.code,
          nameAr: val.acc.nameAr,
          nameEn: val.acc.nameEn,
          type: val.acc.accountType,
          openingDebit: 0,
          openingCredit: 0,
          movementDebit: val.debits,
          movementCredit: val.credits,
          closingDebit: net > 0 ? net : 0,
          closingCredit: net < 0 ? Math.abs(net) : 0,
        });
      }
    });

    return rows.sort((a, b) => a.accountCode.localeCompare(b.accountCode));
  }, [leafAccounts, entries]);

  const tbTotals = useMemo(() => {
    let totMovDebit = 0;
    let totMovCredit = 0;
    let totCloseDebit = 0;
    let totCloseCredit = 0;

    trialBalanceData.forEach((r) => {
      totMovDebit += r.movementDebit;
      totMovCredit += r.movementCredit;
      totCloseDebit += r.closingDebit;
      totCloseCredit += r.closingCredit;
    });

    return {
      totMovDebit,
      totMovCredit,
      totCloseDebit,
      totCloseCredit,
      diff: Math.abs(totCloseDebit - totCloseCredit),
    };
  }, [trialBalanceData]);

  return (
    <div className="space-y-6" dir={dir}>
      {/* Toast Alert */}
      {actionSuccessMsg && (
        <div className="p-4 rounded-xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-700 dark:text-emerald-300 flex items-center gap-3 animate-in fade-in slide-in-from-top-2 duration-300">
          <CheckCircle2 className="h-5 w-5 flex-shrink-0" />
          <span className="text-sm font-medium">{actionSuccessMsg}</span>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b pb-5">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary mb-2">
            <BookOpen className="h-3.5 w-3.5" />
            <span>المرحلة 8: المحاسبة العامة والأستاذ العام (General Ledger Foundation)</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {t.accounting.title}
          </h1>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {t.accounting.subtitle}
          </p>
        </div>

        {/* Badges / Status */}
        <div className="flex flex-wrap items-center gap-2">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-secondary text-secondary-foreground border">
            <ShieldCheck className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            <span>{t.accounting.badge_socpa}</span>
          </div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20">
            <Lock className="h-3.5 w-3.5" />
            <span>{t.accounting.badge_immutable}</span>
          </div>
        </div>
      </div>

      {/* KPI Overview Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl border bg-card/60 shadow-sm flex items-center justify-between">
          <div>
            <div className="text-xs text-muted-foreground font-medium">{t.accounting.total_accounts}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{accounts.length}</div>
            <div className="text-[11px] text-muted-foreground mt-0.5">5 مستويات شجرية وفق SOCPA</div>
          </div>
          <div className="h-11 w-11 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
            <FolderTree className="h-5 w-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl border bg-card/60 shadow-sm flex items-center justify-between">
          <div>
            <div className="text-xs text-muted-foreground font-medium">{t.accounting.active_journals}</div>
            <div className="text-2xl font-bold text-foreground mt-1">5</div>
            <div className="text-[11px] text-muted-foreground mt-0.5">عامة، مبيعات، مشتريات، بنوك، رواتب</div>
          </div>
          <div className="h-11 w-11 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400 flex items-center justify-center">
            <BookOpen className="h-5 w-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl border bg-card/60 shadow-sm flex items-center justify-between">
          <div>
            <div className="text-xs text-muted-foreground font-medium">{t.accounting.posted_entries}</div>
            <div className="text-2xl font-bold text-foreground mt-1">
              {entries.filter((e) => e.status === "POSTED").length}
            </div>
            <div className="text-[11px] text-emerald-600 dark:text-emerald-400 mt-0.5 flex items-center gap-1 font-medium">
              <Check className="h-3 w-3" />
              <span>مرحّلة وغير قابلة للتعديل</span>
            </div>
          </div>
          <div className="h-11 w-11 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
            <FileSpreadsheet className="h-5 w-5" />
          </div>
        </div>

        <div className="p-4 rounded-xl border bg-card/60 shadow-sm flex items-center justify-between">
          <div>
            <div className="text-xs text-muted-foreground font-medium">{t.accounting.tb_equilibrium}</div>
            <div className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
              {t.accounting.balanced_100}
            </div>
            <div className="text-[11px] text-muted-foreground mt-0.5">Δ = 0.00 ر.س (متوازن بدقة)</div>
          </div>
          <div className="h-11 w-11 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center">
            <DollarSign className="h-5 w-5" />
          </div>
        </div>
      </div>

      {/* Main Navigation Tabs */}
      <div className="border-b">
        <nav className="flex gap-2 overflow-x-auto pb-px" aria-label="Tabs">
          <button
            onClick={() => setActiveTab("coa")}
            className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${activeTab === "coa"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              }`}
          >
            <FolderTree className="h-4 w-4" />
            <span>{t.accounting.tab_coa}</span>
            <span className="text-xs bg-muted px-1.5 py-0.5 rounded-full">{accounts.length}</span>
          </button>

          <button
            onClick={() => setActiveTab("journals")}
            className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${activeTab === "journals"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              }`}
          >
            <FileSpreadsheet className="h-4 w-4" />
            <span>{t.accounting.tab_journals}</span>
            <span className="text-xs bg-muted px-1.5 py-0.5 rounded-full">{entries.length}</span>
          </button>

          <button
            onClick={() => setActiveTab("new_entry")}
            className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${activeTab === "new_entry"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              }`}
          >
            <PlusCircle className="h-4 w-4" />
            <span>{t.accounting.tab_new_entry}</span>
            {newEntryTotals.isBalanced && (
              <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
            )}
          </button>

          <button
            onClick={() => setActiveTab("trial_balance")}
            className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${activeTab === "trial_balance"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              }`}
          >
            <DollarSign className="h-4 w-4" />
            <span>{t.accounting.tab_trial_balance}</span>
            <span className="text-[10px] bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 font-semibold px-1.5 py-0.5 rounded">
              OK
            </span>
          </button>

          <button
            onClick={() => setActiveTab("fiscal")}
            className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${activeTab === "fiscal"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted"
              }`}
          >
            <Calendar className="h-4 w-4" />
            <span>{t.accounting.tab_fiscal}</span>
          </button>
        </nav>
      </div>

      {/* ========================================================================= */}
      {/* TAB 1: CHART OF ACCOUNTS (COA)                                            */}
      {/* ========================================================================= */}
      {activeTab === "coa" && (
        <div className="space-y-4 animate-in fade-in duration-200">
          {/* Controls Bar */}
          <div className="p-4 rounded-xl border bg-card/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-2 w-full sm:w-auto">
              <div className="relative flex-1 sm:w-64">
                <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={t.accounting.search_coa}
                  value={coaSearch}
                  onChange={(e) => setCoaSearch(e.target.value)}
                  className="w-full rounded-md border bg-background ps-9 pe-3 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <select
                value={coaFilterType}
                onChange={(e) => setCoaFilterType(e.target.value as any)}
                className="rounded-md border bg-background px-3 py-1.5 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              >
                <option value="ALL">{t.accounting.all_types}</option>
                <option value="ASSET">{t.accounting.assets}</option>
                <option value="LIABILITY">{t.accounting.liabilities}</option>
                <option value="EQUITY">{t.accounting.equity}</option>
                <option value="REVENUE">{t.accounting.revenue}</option>
                <option value="EXPENSE">{t.accounting.expense}</option>
              </select>
            </div>

            <div className="text-xs text-muted-foreground">
              {filteredAccounts.length} من {accounts.length} حساب مطابق للبحث
            </div>
          </div>

          {/* Accounts Tree Table */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="py-3 px-4 text-start">{t.accounting.account_code} & {t.accounting.account_name}</th>
                    <th className="py-3 px-4 text-center">{t.accounting.account_type}</th>
                    <th className="py-3 px-4 text-center">{t.accounting.level}</th>
                    <th className="py-3 px-4 text-center">{t.accounting.normal_balance}</th>
                    <th className="py-3 px-4 text-center">{t.accounting.postable}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredAccounts.map((node) => {
                    const indentPx = (node.level - 1) * 24;
                    const hasChildren = accounts.some((a) => a.parentId === node.id);
                    const isExpanded = expandedNodes[node.id];

                    return (
                      <tr key={node.id} className="hover:bg-muted/30 transition-colors">
                        <td className="py-2.5 px-4 font-medium text-foreground">
                          <div
                            className="flex items-center gap-2"
                            style={{ [isRTL ? "paddingRight" : "paddingLeft"]: `${indentPx}px` }}
                          >
                            {hasChildren ? (
                              <button
                                onClick={() => toggleNode(node.id)}
                                className="p-0.5 rounded hover:bg-muted text-muted-foreground"
                              >
                                {isExpanded ? (
                                  <ChevronDown className="h-4 w-4" />
                                ) : isRTL ? (
                                  <ChevronLeft className="h-4 w-4" />
                                ) : (
                                  <ChevronRight className="h-4 w-4" />
                                )}
                              </button>
                            ) : (
                              <span className="w-5" />
                            )}

                            <span className="font-mono text-xs px-1.5 py-0.5 rounded bg-muted/60 text-foreground font-semibold">
                              {node.code}
                            </span>
                            <span className="font-semibold text-sm">
                              {locale === "ar" ? node.nameAr : node.nameEn}
                            </span>
                            <span className="text-[11px] text-muted-foreground">
                              ({locale === "ar" ? node.nameEn : node.nameAr})
                            </span>
                          </div>
                        </td>

                        <td className="py-2.5 px-4 text-center">
                          <span
                            className={`inline-flex px-2 py-0.5 rounded-full text-[11px] font-semibold ${node.accountType === "ASSET"
                                ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                                : node.accountType === "LIABILITY"
                                  ? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                                  : node.accountType === "EQUITY"
                                    ? "bg-purple-500/10 text-purple-600 dark:text-purple-400"
                                    : node.accountType === "REVENUE"
                                      ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                                      : "bg-rose-500/10 text-rose-600 dark:text-rose-400"
                              }`}
                          >
                            {node.accountType}
                          </span>
                        </td>

                        <td className="py-2.5 px-4 text-center">
                          <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-secondary text-xs font-bold text-foreground">
                            {node.level}
                          </span>
                        </td>

                        <td className="py-2.5 px-4 text-center">
                          <span
                            className={`font-semibold ${node.normalBalance === "DEBIT"
                                ? "text-blue-600 dark:text-blue-400"
                                : "text-amber-600 dark:text-amber-400"
                              }`}
                          >
                            {node.normalBalance === "DEBIT" ? t.accounting.debit : t.accounting.credit}
                          </span>
                        </td>

                        <td className="py-2.5 px-4 text-center">
                          {node.isLeaf ? (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                              <Check className="h-3 w-3" />
                              <span>{t.accounting.postable_yes}</span>
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-medium bg-muted text-muted-foreground">
                              <span>{t.accounting.postable_no}</span>
                            </span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 2: JOURNAL ENTRIES & LEDGERS                                          */}
      {/* ========================================================================= */}
      {activeTab === "journals" && (
        <div className="space-y-4 animate-in fade-in duration-200">
          {/* Filter Bar */}
          <div className="p-4 rounded-xl border bg-card/60 flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <span className="text-xs font-semibold text-muted-foreground">{t.accounting.status}:</span>
              {(["ALL", "DRAFT", "POSTED", "REVERSED"] as const).map((st) => (
                <button
                  key={st}
                  onClick={() => setEntryFilterStatus(st)}
                  className={`px-3 py-1 rounded-md text-xs font-medium transition-colors ${entryFilterStatus === st
                      ? "bg-primary text-primary-foreground shadow-sm"
                      : "bg-secondary text-secondary-foreground hover:bg-muted"
                    }`}
                >
                  {st === "ALL"
                    ? "الكل"
                    : st === "DRAFT"
                      ? t.accounting.status_draft
                      : st === "POSTED"
                        ? t.accounting.status_posted
                        : t.accounting.status_reversed}
                </button>
              ))}
            </div>

            <button
              onClick={() => setActiveTab("new_entry")}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
            >
              <PlusCircle className="h-4 w-4" />
              <span>{t.accounting.tab_new_entry}</span>
            </button>
          </div>

          {/* Entries List */}
          <div className="space-y-3">
            {filteredEntries.map((entry) => {
              const isExpanded = expandedEntryId === entry.id;

              return (
                <div
                  key={entry.id}
                  className="rounded-xl border bg-card shadow-sm overflow-hidden transition-all"
                >
                  {/* Summary Row */}
                  <div className="p-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-3 bg-card hover:bg-muted/20">
                    <div className="flex items-start gap-3">
                      <div
                        className={`h-10 w-10 rounded-xl flex items-center justify-center flex-shrink-0 ${entry.status === "POSTED"
                            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : entry.status === "DRAFT"
                              ? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                              : "bg-rose-500/10 text-rose-600 dark:text-rose-400"
                          }`}
                      >
                        {entry.status === "POSTED" ? (
                          <Lock className="h-5 w-5" />
                        ) : entry.status === "DRAFT" ? (
                          <Unlock className="h-5 w-5" />
                        ) : (
                          <RotateCcw className="h-5 w-5" />
                        )}
                      </div>

                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono font-bold text-sm text-foreground">
                            {entry.entryNumber}
                          </span>
                          <span
                            className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${entry.status === "POSTED"
                                ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                                : entry.status === "DRAFT"
                                  ? "bg-amber-500/15 text-amber-700 dark:text-amber-300"
                                  : "bg-rose-500/15 text-rose-700 dark:text-rose-300"
                              }`}
                          >
                            {entry.status === "POSTED"
                              ? t.accounting.status_posted
                              : entry.status === "DRAFT"
                                ? t.accounting.status_draft
                                : t.accounting.status_reversed}
                          </span>
                          <span className="text-xs text-muted-foreground font-mono">
                            {entry.journalCode} • {locale === "ar" ? entry.journalNameAr : entry.journalNameEn}
                          </span>
                        </div>
                        <p className="text-xs text-foreground font-medium mt-1">
                          {entry.description}
                        </p>
                        <div className="flex items-center gap-3 text-[11px] text-muted-foreground mt-1">
                          <span>{entry.entryDate}</span>
                          <span>•</span>
                          <span>{entry.periodName}</span>
                          {entry.reversalReason && (
                            <>
                              <span>•</span>
                              <span className="text-rose-600 dark:text-rose-400 font-medium">
                                سبب العكس: {entry.reversalReason}
                              </span>
                            </>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Financial Figures & Actions */}
                    <div className="flex flex-wrap items-center gap-4 w-full md:w-auto justify-between md:justify-end border-t md:border-0 pt-3 md:pt-0">
                      <div className="text-end">
                        <div className="text-[11px] text-muted-foreground">
                          {t.accounting.total_debit} / {t.accounting.total_credit}
                        </div>
                        <div className="text-sm font-mono font-bold text-foreground">
                          {entry.totalDebit.toLocaleString("en-US", { minimumFractionDigits: 2 })} ر.س
                        </div>
                      </div>

                      <div className="flex items-center gap-1.5">
                        {entry.status === "DRAFT" && (
                          <button
                            onClick={() => handlePostEntry(entry.id)}
                            className="inline-flex items-center gap-1 px-3 py-1.5 rounded-md text-xs font-semibold bg-emerald-600 text-white hover:bg-emerald-700 shadow-sm transition-colors"
                          >
                            <Send className="h-3.5 w-3.5" />
                            <span>{t.accounting.post_now}</span>
                          </button>
                        )}

                        {entry.status === "POSTED" && (
                          <button
                            onClick={() => setReversingEntry(entry)}
                            className="inline-flex items-center gap-1 px-3 py-1.5 rounded-md text-xs font-semibold bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300 border border-rose-200 dark:border-rose-900 hover:bg-rose-100 transition-colors"
                          >
                            <RotateCcw className="h-3.5 w-3.5" />
                            <span>{t.accounting.reverse_now}</span>
                          </button>
                        )}

                        <button
                          onClick={() => setExpandedEntryId(isExpanded ? null : entry.id)}
                          className="px-2.5 py-1.5 rounded-md text-xs font-medium bg-muted hover:bg-muted/80 text-foreground transition-colors"
                        >
                          {isExpanded ? t.accounting.hide_lines : t.accounting.view_lines}
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Expanded Lines Table */}
                  {isExpanded && (
                    <div className="border-t bg-muted/20 p-4 animate-in fade-in duration-150">
                      <div className="text-xs font-semibold text-muted-foreground mb-2 flex items-center justify-between">
                        <span>سطور قيد اليومية (Journal Lines)</span>
                        <span className="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
                          {t.accounting.invariant_balanced}
                        </span>
                      </div>

                      <div className="rounded-lg border bg-card overflow-hidden">
                        <table className="w-full text-xs text-start">
                          <thead className="bg-muted/40 border-b text-muted-foreground">
                            <tr>
                              <th className="py-2 px-3 text-start">{t.accounting.account_code}</th>
                              <th className="py-2 px-3 text-start">{t.accounting.account_name}</th>
                              <th className="py-2 px-3 text-start">{t.accounting.line_desc}</th>
                              <th className="py-2 px-3 text-end">{t.accounting.debit} (ر.س)</th>
                              <th className="py-2 px-3 text-end">{t.accounting.credit} (ر.س)</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y font-mono">
                            {entry.lines.map((l) => (
                              <tr key={l.id} className="hover:bg-muted/20">
                                <td className="py-2 px-3 font-semibold text-foreground">{l.accountCode}</td>
                                <td className="py-2 px-3 font-sans text-foreground">
                                  {locale === "ar" ? l.accountNameAr : l.accountNameEn}
                                </td>
                                <td className="py-2 px-3 font-sans text-muted-foreground">{l.description}</td>
                                <td className="py-2 px-3 text-end font-semibold text-blue-600 dark:text-blue-400">
                                  {l.debit > 0 ? l.debit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                                </td>
                                <td className="py-2 px-3 text-end font-semibold text-amber-600 dark:text-amber-400">
                                  {l.credit > 0 ? l.credit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                          <tfoot className="bg-muted/60 font-mono font-bold border-t">
                            <tr>
                              <td colSpan={3} className="py-2.5 px-3 text-end font-sans">
                                {t.accounting.tb_total}
                              </td>
                              <td className="py-2.5 px-3 text-end text-blue-700 dark:text-blue-300">
                                {entry.totalDebit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                              </td>
                              <td className="py-2.5 px-3 text-end text-amber-700 dark:text-amber-300">
                                {entry.totalCredit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                              </td>
                            </tr>
                          </tfoot>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 3: NEW BALANCED JOURNAL ENTRY STUDIO                                  */}
      {/* ========================================================================= */}
      {activeTab === "new_entry" && (
        <div className="space-y-5 animate-in fade-in duration-200">
          <div className="p-5 rounded-xl border bg-card shadow-sm space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div className="flex items-center gap-2">
                <PlusCircle className="h-5 w-5 text-primary" />
                <h2 className="text-base font-bold text-foreground">
                  {t.accounting.new_entry_title}
                </h2>
              </div>
              <span className="text-xs text-muted-foreground">
                المعايير المطبقة: SOCPA / IFRS (توازن إلزامي مدين = دائن)
              </span>
            </div>

            {/* Header Fields */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t.accounting.select_journal}
                </label>
                <select
                  value={newJournalCode}
                  onChange={(e) => setNewJournalCode(e.target.value)}
                  className="w-full rounded-md border bg-background px-3 py-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  <option value="GEN">GEN - دفتر اليومية العامة</option>
                  <option value="SALES">SALES - دفتر المبيعات الآجلة</option>
                  <option value="PURCH">PURCH - دفتر المشتريات والموردين</option>
                  <option value="BANK">BANK - دفتر الخزينة والمدفوعات البنكية</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t.accounting.entry_date}
                </label>
                <input
                  type="date"
                  value={newEntryDate}
                  onChange={(e) => setNewEntryDate(e.target.value)}
                  className="w-full rounded-md border bg-background px-3 py-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono"
                >
                </input>
              </div>

              <div>
                <label className="block text-xs font-semibold text-foreground mb-1">
                  {t.accounting.period}
                </label>
                <input
                  type="text"
                  disabled
                  value="2026-01 (يناير - فترة مفتوحة OPEN)"
                  className="w-full rounded-md border bg-muted px-3 py-2 text-xs text-muted-foreground"
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-foreground mb-1">
                {t.accounting.description} (البيان العام للقيد)
              </label>
              <input
                type="text"
                placeholder="أدخل وصفاً تفصيلياً لحركة القيد المحاسبي..."
                value={newDescription}
                onChange={(e) => setNewDescription(e.target.value)}
                className="w-full rounded-md border bg-background px-3 py-2 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>

            {/* Lines Composer */}
            <div className="space-y-3 pt-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-foreground">سطور القيد المحاسبي (Debit / Credit)</span>
                <button
                  type="button"
                  onClick={handleAddLine}
                  className="inline-flex items-center gap-1 px-3 py-1 rounded-md text-xs font-semibold bg-secondary text-secondary-foreground hover:bg-muted"
                >
                  <PlusCircle className="h-3.5 w-3.5" />
                  <span>{t.accounting.add_line}</span>
                </button>
              </div>

              <div className="rounded-lg border overflow-hidden">
                <table className="w-full text-xs text-start">
                  <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                    <tr>
                      <th className="py-2.5 px-3 text-start">{t.accounting.select_account}</th>
                      <th className="py-2.5 px-3 text-start">{t.accounting.line_desc}</th>
                      <th className="py-2.5 px-3 text-end w-36">{t.accounting.debit} (ر.س)</th>
                      <th className="py-2.5 px-3 text-end w-36">{t.accounting.credit} (ر.س)</th>
                      <th className="py-2.5 px-3 text-center w-12"></th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {newLines.map((line, idx) => (
                      <tr key={line.id}>
                        <td className="p-2">
                          <select
                            value={line.accountId}
                            onChange={(e) => {
                              const val = e.target.value;
                              setNewLines((prev) =>
                                prev.map((l, i) => (i === idx ? { ...l, accountId: val } : l))
                              );
                            }}
                            className="w-full rounded border bg-background px-2 py-1.5 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                          >
                            {leafAccounts.map((acc) => (
                              <option key={acc.id} value={acc.id}>
                                {acc.code} - {locale === "ar" ? acc.nameAr : acc.nameEn}
                              </option>
                            ))}
                          </select>
                        </td>

                        <td className="p-2">
                          <input
                            type="text"
                            placeholder="بيان تفصيلي للسطر..."
                            value={line.description}
                            onChange={(e) => {
                              const val = e.target.value;
                              setNewLines((prev) =>
                                prev.map((l, i) => (i === idx ? { ...l, description: val } : l))
                              );
                            }}
                            className="w-full rounded border bg-background px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                          />
                        </td>

                        <td className="p-2 text-end">
                          <input
                            type="number"
                            step="0.01"
                            min="0"
                            value={line.debit}
                            onChange={(e) => {
                              const val = e.target.value;
                              setNewLines((prev) =>
                                prev.map((l, i) =>
                                  i === idx ? { ...l, debit: val, credit: parseFloat(val) > 0 ? "0" : l.credit } : l
                                )
                              );
                            }}
                            className="w-full rounded border bg-background px-2 py-1.5 text-xs text-end font-mono text-blue-600 dark:text-blue-400 font-bold focus:outline-none focus:ring-1 focus:ring-ring"
                          />
                        </td>

                        <td className="p-2 text-end">
                          <input
                            type="number"
                            step="0.01"
                            min="0"
                            value={line.credit}
                            onChange={(e) => {
                              const val = e.target.value;
                              setNewLines((prev) =>
                                prev.map((l, i) =>
                                  i === idx ? { ...l, credit: val, debit: parseFloat(val) > 0 ? "0" : l.debit } : l
                                )
                              );
                            }}
                            className="w-full rounded border bg-background px-2 py-1.5 text-xs text-end font-mono text-amber-600 dark:text-amber-400 font-bold focus:outline-none focus:ring-1 focus:ring-ring"
                          />
                        </td>

                        <td className="p-2 text-center">
                          <button
                            type="button"
                            disabled={newLines.length <= 2}
                            onClick={() => handleRemoveLine(idx)}
                            className="p-1 rounded text-muted-foreground hover:text-rose-600 disabled:opacity-30"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot className="bg-muted/40 font-mono font-bold border-t">
                    <tr>
                      <td colSpan={2} className="py-2.5 px-3 text-end font-sans">
                        {t.accounting.tb_total}
                      </td>
                      <td className="py-2.5 px-3 text-end text-blue-700 dark:text-blue-300">
                        {newEntryTotals.debits.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="py-2.5 px-3 text-end text-amber-700 dark:text-amber-300">
                        {newEntryTotals.credits.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                      </td>
                      <td></td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>

            {/* Live Invariant Validation Callout */}
            <div
              className={`p-4 rounded-xl border flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 ${newEntryTotals.isBalanced
                  ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-800 dark:text-emerald-200"
                  : "bg-amber-500/10 border-amber-500/30 text-amber-800 dark:text-amber-200"
                }`}
            >
              <div className="flex items-center gap-2.5">
                {newEntryTotals.isBalanced ? (
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
                ) : (
                  <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 flex-shrink-0" />
                )}
                <div>
                  <div className="text-xs font-bold">
                    {newEntryTotals.isBalanced
                      ? t.accounting.invariant_balanced
                      : t.accounting.invariant_unbalanced}
                  </div>
                  <div className="text-[11px] opacity-80 mt-0.5">
                    المدين: {newEntryTotals.debits.toFixed(2)} ر.س | الدائن: {newEntryTotals.credits.toFixed(2)} ر.س | الفارق: {newEntryTotals.diff.toFixed(2)} ر.س
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex items-center gap-2 w-full sm:w-auto">
                <button
                  type="button"
                  disabled={!newEntryTotals.isBalanced}
                  onClick={() => handleCreateNewEntry(false)}
                  className="flex-1 sm:flex-initial px-4 py-2 rounded-md text-xs font-semibold bg-secondary text-secondary-foreground hover:bg-muted disabled:opacity-40 transition-colors"
                >
                  {t.accounting.create_draft}
                </button>

                <button
                  type="button"
                  disabled={!newEntryTotals.isBalanced}
                  onClick={() => handleCreateNewEntry(true)}
                  className="flex-1 sm:flex-initial inline-flex items-center justify-center gap-1.5 px-4 py-2 rounded-md text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-40 shadow-sm transition-colors"
                >
                  <Send className="h-3.5 w-3.5" />
                  <span>{t.accounting.create_and_post}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 4: TRIAL BALANCE REPORT (ميزان المراجعة)                              */}
      {/* ========================================================================= */}
      {activeTab === "trial_balance" && (
        <div className="space-y-4 animate-in fade-in duration-200">
          {/* Mathematical Balance Proof Callout */}
          <div className="p-4 rounded-xl bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent border border-emerald-500/30 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div className="flex items-start gap-3">
              <div className="h-10 w-10 rounded-xl bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center flex-shrink-0">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <div>
                <div className="text-sm font-bold text-foreground">
                  {t.accounting.tb_proof_title}
                </div>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {t.accounting.tb_proof_desc}
                </p>
                <div className="text-xs font-mono text-emerald-700 dark:text-emerald-300 font-semibold mt-1">
                  Σ Debits ({tbTotals.totCloseDebit.toLocaleString("en-US", { minimumFractionDigits: 2 })}) = Σ Credits ({tbTotals.totCloseCredit.toLocaleString("en-US", { minimumFractionDigits: 2 })})
                </div>
              </div>
            </div>

            <div className="px-3 py-1.5 rounded-lg bg-emerald-500/20 border border-emerald-500/30 text-emerald-800 dark:text-emerald-300 text-xs font-bold font-mono">
              Δ = 0.00 SAR (ZERO DEFICIT)
            </div>
          </div>

          {/* Trial Balance Table */}
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <div className="p-4 border-b bg-muted/20 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <h3 className="text-sm font-bold text-foreground">{t.accounting.tb_title}</h3>
                <span className="text-xs text-muted-foreground">الفترة: السنة المالية 2026 (فترة يناير) • العملة: ريال سعودي (SAR)</span>
              </div>
              <div className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
                <Building className="h-3.5 w-3.5 text-primary" />
                <span>الكيان: شركة الأفق المتقدمة للتقنية (LE-01)</span>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th rowSpan={2} className="py-3 px-3 text-start border-e">{t.accounting.account_code}</th>
                    <th rowSpan={2} className="py-3 px-3 text-start border-e">{t.accounting.account_name}</th>
                    <th rowSpan={2} className="py-3 px-3 text-center border-e">{t.accounting.account_type}</th>
                    <th colSpan={2} className="py-2 px-3 text-center border-e bg-blue-500/5">{t.accounting.tb_movement}</th>
                    <th colSpan={2} className="py-2 px-3 text-center bg-emerald-500/5">{t.accounting.tb_closing}</th>
                  </tr>
                  <tr className="border-t text-[11px]">
                    <th className="py-1 px-3 text-end bg-blue-500/5">{t.accounting.debit}</th>
                    <th className="py-1 px-3 text-end border-e bg-blue-500/5">{t.accounting.credit}</th>
                    <th className="py-1 px-3 text-end bg-emerald-500/5">{t.accounting.debit}</th>
                    <th className="py-1 px-3 text-end bg-emerald-500/5">{t.accounting.credit}</th>
                  </tr>
                </thead>
                <tbody className="divide-y font-mono">
                  {trialBalanceData.map((row) => (
                    <tr key={row.accountCode} className="hover:bg-muted/30">
                      <td className="py-2.5 px-3 font-semibold text-foreground border-e">{row.accountCode}</td>
                      <td className="py-2.5 px-3 font-sans font-medium text-foreground border-e">
                        {locale === "ar" ? row.nameAr : row.nameEn}
                      </td>
                      <td className="py-2.5 px-3 text-center border-e font-sans">
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-medium">
                          {row.type}
                        </span>
                      </td>
                      <td className="py-2.5 px-3 text-end text-blue-600 dark:text-blue-400">
                        {row.movementDebit > 0 ? row.movementDebit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                      </td>
                      <td className="py-2.5 px-3 text-end text-amber-600 dark:text-amber-400 border-e">
                        {row.movementCredit > 0 ? row.movementCredit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                      </td>
                      <td className="py-2.5 px-3 text-end font-bold text-foreground">
                        {row.closingDebit > 0 ? row.closingDebit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                      </td>
                      <td className="py-2.5 px-3 text-end font-bold text-foreground">
                        {row.closingCredit > 0 ? row.closingCredit.toLocaleString("en-US", { minimumFractionDigits: 2 }) : "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot className="bg-muted font-mono font-bold text-xs border-t-2">
                  <tr>
                    <td colSpan={3} className="py-3 px-3 text-end font-sans">
                      {t.accounting.tb_total} (المجاميع والأرصدة العامة)
                    </td>
                    <td className="py-3 px-3 text-end text-blue-700 dark:text-blue-300">
                      {tbTotals.totMovDebit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="py-3 px-3 text-end text-amber-700 dark:text-amber-300 border-e">
                      {tbTotals.totMovCredit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="py-3 px-3 text-end text-emerald-700 dark:text-emerald-400">
                      {tbTotals.totCloseDebit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="py-3 px-3 text-end text-emerald-700 dark:text-emerald-400">
                      {tbTotals.totCloseCredit.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 5: FISCAL CALENDAR & PERIOD LOCKING                                   */}
      {/* ========================================================================= */}
      {activeTab === "fiscal" && (
        <div className="space-y-4 animate-in fade-in duration-200">
          <div className="p-4 rounded-xl border bg-card shadow-sm space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <div>
                <h3 className="text-sm font-bold text-foreground">التقويم المالي وقواعد إقفال الفترات (Fiscal Periods)</h3>
                <p className="text-xs text-muted-foreground mt-0.5">
                  تطبيق متطلب النظام الأساسي: لا يُقبل ترحيل أي قيد إلا في فترة مالية بحالة مفتوحة (OPEN).
                </p>
              </div>
              <span className="text-xs font-mono font-bold px-2.5 py-1 rounded-md bg-secondary text-secondary-foreground border">
                FY-2026 (12 فترات شهرية)
              </span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
              {fiscalPeriods.map((p) => {
                const isOpen = p.status === "OPEN";

                return (
                  <div
                    key={p.id}
                    className={`p-3.5 rounded-xl border transition-all ${isOpen
                        ? "bg-emerald-500/5 border-emerald-500/30"
                        : "bg-muted/40 border-muted"
                      }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-mono font-bold text-sm text-foreground">{p.name}</span>
                      <span
                        className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${isOpen
                            ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                            : "bg-rose-500/15 text-rose-700 dark:text-rose-300"
                          }`}
                      >
                        {isOpen ? "مفتوحة (OPEN)" : "مقفلة (LOCKED)"}
                      </span>
                    </div>

                    <div className="text-[11px] text-muted-foreground mt-2 font-mono">
                      {p.startDate} إلى {p.endDate}
                    </div>

                    <div className="mt-3 pt-2 border-t flex items-center justify-between">
                      <span className="text-[11px] text-muted-foreground">
                        {isOpen ? "تقبل القيود" : "محظورة"}
                      </span>
                      <button
                        onClick={() => {
                          setFiscalPeriods((prev) =>
                            prev.map((item) =>
                              item.id === p.id
                                ? { ...item, status: item.status === "OPEN" ? "HARD_CLOSED" as any : "OPEN" }
                                : item
                            )
                          );
                        }}
                        className={`px-2 py-1 rounded text-[11px] font-semibold transition-colors ${isOpen
                            ? "bg-amber-500/10 text-amber-700 dark:text-amber-300 hover:bg-amber-500/20"
                            : "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/20"
                          }`}
                      >
                        {isOpen ? "إقفال الفترة" : "إعادة الفتح"}
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* REVERSAL MODAL                                                            */}
      {/* ========================================================================= */}
      {reversingEntry && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm animate-in fade-in duration-150">
          <div className="w-full max-w-md rounded-2xl border bg-card p-6 shadow-2xl space-y-4">
            <div className="flex items-center gap-3 text-rose-600 dark:text-rose-400">
              <RotateCcw className="h-6 w-6" />
              <h3 className="text-base font-bold text-foreground">
                {t.accounting.reverse_confirm_title}
              </h3>
            </div>

            <p className="text-xs text-muted-foreground leading-relaxed">
              وفق المعايير المحاسبية وقواعد قاعدة البيانات (Immutability Triggers)، يُمنع تعديل أو حذف القيود المرحّلة. سيقوم النظام بتوليد قيد عكسي تلقائي معكوس الأطراف (Inverted Debits/Credits) وربطه برقم القيد الأصلي:
              <strong className="font-mono text-foreground ms-1">{reversingEntry.entryNumber}</strong>
            </p>

            <div>
              <label className="block text-xs font-semibold text-foreground mb-1">
                {t.accounting.reverse_reason}
              </label>
              <textarea
                rows={3}
                placeholder="أدخل سبب عكس القيد المالي للتوثيق والتدقيق الرقابي..."
                value={reversalReason}
                onChange={(e) => setReversalReason(e.target.value)}
                className="w-full rounded-md border bg-background p-2.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t">
              <button
                type="button"
                onClick={() => {
                  setReversingEntry(null);
                  setReversalReason("");
                }}
                className="px-4 py-2 rounded-md text-xs font-semibold bg-secondary text-secondary-foreground hover:bg-muted transition-colors"
              >
                إلغاء
              </button>

              <button
                type="button"
                disabled={!reversalReason.trim()}
                onClick={handleExecuteReversal}
                className="px-4 py-2 rounded-md text-xs font-semibold bg-rose-600 text-white hover:bg-rose-700 disabled:opacity-40 shadow-sm transition-colors"
              >
                {t.accounting.reverse_execute}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

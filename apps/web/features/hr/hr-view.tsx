"use client";

import React, { useState } from "react";
import {
  Users,
  FileCheck2,
  Briefcase,
  Award,
  ShieldCheck,
  Plus,
  Search,
  CheckCircle2,
  AlertCircle,
  TrendingUp,
  UserCheck,
  Building2,
  Eye,
  EyeOff,
  Percent,
  Sparkles,
  ArrowUpRight,
  Filter
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface EmployeeItem {
  id: string;
  employeeNumber: string;
  firstNameEn: string;
  lastNameEn: string;
  firstNameAr: string;
  lastNameAr: string;
  idType: "NATIONAL_ID" | "IQAMA";
  idNumberMasked: string;
  nationalityCode: string;
  isSaudi: boolean;
  jobTitleEn: string;
  jobTitleAr: string;
  department: string;
  hireDate: string;
  status: "ACTIVE" | "PROBATION" | "ON_LEAVE" | "TERMINATED";
  ibanMasked: string;
  isWpsCompliant: boolean;
}

interface ContractItem {
  id: string;
  contractNumber: string;
  employeeNumber: string;
  employeeNameEn: string;
  employeeNameAr: string;
  contractType: "UNSPECIFIED_TERM" | "SPECIFIED_TERM";
  startDate: string;
  endDate?: string;
  probationDays: number;
  basicSalary: number;
  housingAllowance: number;
  transportAllowance: number;
  otherAllowances: number;
  grossSalary: number;
  versionNumber: number;
  status: "ACTIVE" | "PENDING_APPROVAL" | "EXPIRED" | "TERMINATED";
}

interface PositionItem {
  id: string;
  code: string;
  jobTitleEn: string;
  jobTitleAr: string;
  jobFamily: string;
  department: string;
  headcountTarget: number;
  headcountCurrent: number;
  status: "VACANT" | "FILLED" | "FROZEN";
}

export function HRView() {
  const { t, locale, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"employees" | "contracts" | "positions" | "saudization">("employees");

  // Initial Seeded/Demo State
  const [employees, setEmployees] = useState<EmployeeItem[]>([
    {
      id: "emp-1",
      employeeNumber: "EMP-2025-001",
      firstNameEn: "Mohammed",
      lastNameEn: "Al-Otaibi",
      firstNameAr: "محمد",
      lastNameAr: "العتيبي",
      idType: "NATIONAL_ID",
      idNumberMasked: "1087****21",
      nationalityCode: "SA",
      isSaudi: true,
      jobTitleEn: "Senior HR Specialist",
      jobTitleAr: "أخصائي موارد بشرية أول",
      department: "Human Resources / الموارد البشرية",
      hireDate: "2025-01-01",
      status: "ACTIVE",
      ibanMasked: "SA4480000204****1234",
      isWpsCompliant: true,
    },
    {
      id: "emp-2",
      employeeNumber: "EMP-2025-002",
      firstNameEn: "Sarah",
      lastNameEn: "Al-Ghamdi",
      firstNameAr: "سارة",
      lastNameAr: "الغامدي",
      idType: "NATIONAL_ID",
      idNumberMasked: "1054****89",
      nationalityCode: "SA",
      isSaudi: true,
      jobTitleEn: "Lead Software Engineer",
      jobTitleAr: "كبير مهندسي البرمجيات",
      department: "Information Technology / تقنية المعلومات",
      hireDate: "2025-02-15",
      status: "ACTIVE",
      ibanMasked: "SA8810000101****4321",
      isWpsCompliant: true,
    },
    {
      id: "emp-3",
      employeeNumber: "EMP-2025-003",
      firstNameEn: "Tariq",
      lastNameEn: "Mansour",
      firstNameAr: "طارق",
      lastNameAr: "منصور",
      idType: "NATIONAL_ID",
      idNumberMasked: "1092****33",
      nationalityCode: "SA",
      isSaudi: true,
      jobTitleEn: "Senior Financial Analyst",
      jobTitleAr: "محلل مالي أول",
      department: "Finance & Accounting / الإدارة المالية",
      hireDate: "2025-03-01",
      status: "ACTIVE",
      ibanMasked: "SA5520000001****9876",
      isWpsCompliant: true,
    },
    {
      id: "emp-4",
      employeeNumber: "EMP-2025-004",
      firstNameEn: "Ahmed",
      lastNameEn: "Hassan",
      firstNameAr: "أحمد",
      lastNameAr: "حسن",
      idType: "IQAMA",
      idNumberMasked: "2450****67",
      nationalityCode: "EG",
      isSaudi: false,
      jobTitleEn: "Database Architect",
      jobTitleAr: "مهندس قواعد بيانات",
      department: "Information Technology / تقنية المعلومات",
      hireDate: "2025-04-10",
      status: "ACTIVE",
      ibanMasked: "SA3312000000****5544",
      isWpsCompliant: true,
    },
  ]);

  const [contracts, setContracts] = useState<ContractItem[]>([
    {
      id: "cnt-1",
      contractNumber: "QIWA-2025-00101",
      employeeNumber: "EMP-2025-001",
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      contractType: "UNSPECIFIED_TERM",
      startDate: "2025-01-01",
      probationDays: 90,
      basicSalary: 12000,
      housingAllowance: 3000,
      transportAllowance: 1000,
      otherAllowances: 500,
      grossSalary: 16500,
      versionNumber: 2,
      status: "ACTIVE",
    },
    {
      id: "cnt-2",
      contractNumber: "QIWA-2025-00102",
      employeeNumber: "EMP-2025-002",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      contractType: "UNSPECIFIED_TERM",
      startDate: "2025-02-15",
      probationDays: 90,
      basicSalary: 18000,
      housingAllowance: 4500,
      transportAllowance: 1500,
      otherAllowances: 1000,
      grossSalary: 25000,
      versionNumber: 1,
      status: "ACTIVE",
    },
    {
      id: "cnt-3",
      contractNumber: "QIWA-2025-00103",
      employeeNumber: "EMP-2025-004",
      employeeNameEn: "Ahmed Hassan",
      employeeNameAr: "أحمد حسن",
      contractType: "SPECIFIED_TERM",
      startDate: "2025-04-10",
      endDate: "2027-04-09",
      probationDays: 90,
      basicSalary: 15000,
      housingAllowance: 3750,
      transportAllowance: 1250,
      otherAllowances: 0,
      grossSalary: 20000,
      versionNumber: 1,
      status: "ACTIVE",
    },
  ]);

  const [positions, setPositions] = useState<PositionItem[]>([
    {
      id: "pos-1",
      code: "POS-HR-001",
      jobTitleEn: "Senior HR Specialist",
      jobTitleAr: "أخصائي موارد بشرية أول",
      jobFamily: "HR",
      department: "Human Resources",
      headcountTarget: 1,
      headcountCurrent: 1,
      status: "FILLED",
    },
    {
      id: "pos-2",
      code: "POS-ENG-001",
      jobTitleEn: "Lead Software Engineer",
      jobTitleAr: "كبير مهندسي البرمجيات",
      jobFamily: "Engineering",
      department: "Information Technology",
      headcountTarget: 2,
      headcountCurrent: 1,
      status: "VACANT",
    },
    {
      id: "pos-3",
      code: "POS-FIN-001",
      jobTitleEn: "Senior Financial Analyst",
      jobTitleAr: "محلل مالي أول",
      jobFamily: "Finance",
      department: "Finance & Accounting",
      headcountTarget: 1,
      headcountCurrent: 1,
      status: "FILLED",
    },
    {
      id: "pos-4",
      code: "POS-OPS-001",
      jobTitleEn: "Logistics Coordinator",
      jobTitleAr: "منسق عمليات لوجستية",
      jobFamily: "Operations",
      department: "Operations & Supply",
      headcountTarget: 1,
      headcountCurrent: 0,
      status: "VACANT",
    },
  ]);

  // Modals state
  const [showNewEmpModal, setShowNewEmpModal] = useState(false);
  const [showNewContractModal, setShowNewContractModal] = useState(false);
  const [showAmendModal, setShowAmendModal] = useState(false);
  const [selectedContract, setSelectedContract] = useState<ContractItem | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("ALL");

  // Form states
  const [newEmpForm, setNewEmpForm] = useState({
    firstNameEn: "",
    lastNameEn: "",
    firstNameAr: "",
    lastNameAr: "",
    nationalityCode: "SA",
    idNumber: "",
    jobTitleEn: "Operations Coordinator",
    jobTitleAr: "منسق عمليات",
    department: "Operations & Services",
    iban: "SA5510000001234567890123",
  });

  const [newContractForm, setNewContractForm] = useState({
    employeeNumber: "EMP-2025-003",
    contractType: "UNSPECIFIED_TERM" as const,
    basicSalary: 10000,
    housingAllowance: 2500,
    transportAllowance: 1000,
    otherAllowances: 0,
    probationDays: 90,
  });

  const [amendForm, setAmendForm] = useState({
    newBasicSalary: 14000,
    newHousing: 3500,
    newTransport: 1000,
    newOther: 500,
    reason: "ANNUAL_MERIT_INCREASE",
  });

  // Calculate Metrics
  const totalEmployees = employees.length;
  const saudiEmployees = employees.filter((e) => e.isSaudi).length;
  const expatEmployees = totalEmployees - saudiEmployees;
  const saudizationRate = totalEmployees > 0 ? ((saudiEmployees / totalEmployees) * 100).toFixed(1) : "0.0";
  const vacantPositions = positions.filter((p) => p.status === "VACANT").length;
  const totalPayroll = contracts.reduce((acc, c) => acc + c.grossSalary, 0);

  // Nitaqat Tier calculation
  const getTierBadge = (rate: number) => {
    if (rate >= 50) return { label: "بلاتيني (Platinum)", color: "bg-emerald-600/10 text-emerald-600 border-emerald-500/20" };
    if (rate >= 40) return { label: "أخضر مرتفع (High Green)", color: "bg-green-600/10 text-green-600 border-green-500/20" };
    if (rate >= 25) return { label: "أخضر متوسط (Mid Green)", color: "bg-lime-600/10 text-lime-600 border-lime-500/20" };
    if (rate >= 10) return { label: "أخضر منخفض (Low Green)", color: "bg-amber-600/10 text-amber-600 border-amber-500/20" };
    return { label: "نطاق أحمر (Red Tier)", color: "bg-red-600/10 text-red-600 border-red-500/20" };
  };

  const currentTier = getTierBadge(parseFloat(saudizationRate));

  // Handlers
  const handleCreateEmployee = (e: React.FormEvent) => {
    e.preventDefault();
    const isSaudi = newEmpForm.nationalityCode === "SA";
    const newEmp: EmployeeItem = {
      id: `emp-${Date.now()}`,
      employeeNumber: `EMP-2025-00${employees.length + 1}`,
      firstNameEn: newEmpForm.firstNameEn,
      lastNameEn: newEmpForm.lastNameEn,
      firstNameAr: newEmpForm.firstNameAr,
      lastNameAr: newEmpForm.lastNameAr,
      idType: isSaudi ? "NATIONAL_ID" : "IQAMA",
      idNumberMasked: newEmpForm.idNumber ? `${newEmpForm.idNumber.slice(0, 4)}****${newEmpForm.idNumber.slice(-2)}` : "1099****01",
      nationalityCode: newEmpForm.nationalityCode,
      isSaudi,
      jobTitleEn: newEmpForm.jobTitleEn,
      jobTitleAr: newEmpForm.jobTitleAr,
      department: newEmpForm.department,
      hireDate: new Date().toISOString().split("T")[0],
      status: "ACTIVE",
      ibanMasked: `${newEmpForm.iban.slice(0, 8)}****${newEmpForm.iban.slice(-4)}`,
      isWpsCompliant: true,
    };

    setEmployees([...employees, newEmp]);
    setShowNewEmpModal(false);
  };

  const handleCreateContract = (e: React.FormEvent) => {
    e.preventDefault();
    const emp = employees.find((x) => x.employeeNumber === newContractForm.employeeNumber);
    const gross =
      Number(newContractForm.basicSalary) +
      Number(newContractForm.housingAllowance) +
      Number(newContractForm.transportAllowance) +
      Number(newContractForm.otherAllowances);

    const newContract: ContractItem = {
      id: `cnt-${Date.now()}`,
      contractNumber: `QIWA-2025-00${contracts.length + 101}`,
      employeeNumber: newContractForm.employeeNumber,
      employeeNameEn: emp ? `${emp.firstNameEn} ${emp.lastNameEn}` : "Employee",
      employeeNameAr: emp ? `${emp.firstNameAr} ${emp.lastNameAr}` : "موظف",
      contractType: newContractForm.contractType,
      startDate: new Date().toISOString().split("T")[0],
      probationDays: Number(newContractForm.probationDays),
      basicSalary: Number(newContractForm.basicSalary),
      housingAllowance: Number(newContractForm.housingAllowance),
      transportAllowance: Number(newContractForm.transportAllowance),
      otherAllowances: Number(newContractForm.otherAllowances),
      grossSalary: gross,
      versionNumber: 1,
      status: "ACTIVE",
    };

    setContracts([...contracts, newContract]);
    setShowNewContractModal(false);
  };

  const handleAmendContract = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedContract) return;

    const gross =
      Number(amendForm.newBasicSalary) +
      Number(amendForm.newHousing) +
      Number(amendForm.newTransport) +
      Number(amendForm.newOther);

    const updated = contracts.map((c) => {
      if (c.id === selectedContract.id) {
        return {
          ...c,
          basicSalary: Number(amendForm.newBasicSalary),
          housingAllowance: Number(amendForm.newHousing),
          transportAllowance: Number(amendForm.newTransport),
          otherAllowances: Number(amendForm.newOther),
          grossSalary: gross,
          versionNumber: c.versionNumber + 1,
        };
      }
      return c;
    });

    setContracts(updated);
    setShowAmendModal(false);
    setSelectedContract(null);
  };

  const filteredEmployees = employees.filter((emp) => {
    const fullName = `${emp.firstNameAr} ${emp.lastNameAr} ${emp.firstNameEn} ${emp.lastNameEn} ${emp.employeeNumber}`.toLowerCase();
    const matchesSearch = fullName.includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === "ALL" || emp.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight">{t.hr.title}</h1>
            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              <ShieldCheck className="h-3 w-3" />
              Qiwa & Labor Law Compliant
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">{t.hr.description}</p>
        </div>

        <div className="flex items-center gap-2">
          {activeTab === "employees" && (
            <button
              onClick={() => setShowNewEmpModal(true)}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90 transition-colors"
            >
              <Plus className="h-4 w-4" />
              {t.hr.btn_new_employee}
            </button>
          )}
          {activeTab === "contracts" && (
            <button
              onClick={() => setShowNewContractModal(true)}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90 transition-colors"
            >
              <Plus className="h-4 w-4" />
              {t.hr.btn_new_contract}
            </button>
          )}
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        {/* Total Employees */}
        <div className="rounded-xl border bg-card p-4 shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.hr.kpi_total_employees}</span>
            <Users className="h-4 w-4 text-primary" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold">{totalEmployees}</span>
            <span className="text-xs text-muted-foreground">
              ({saudiEmployees} SA / {expatEmployees} EXP)
            </span>
          </div>
          <div className="mt-2 text-[11px] text-muted-foreground">نشطون في مختلف الفروع</div>
        </div>

        {/* Saudization Rate */}
        <div className="rounded-xl border bg-card p-4 shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.hr.kpi_saudization_rate}</span>
            <Percent className="h-4 w-4 text-emerald-600" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-emerald-600">{saudizationRate}%</span>
          </div>
          <div className="w-full bg-secondary h-1.5 rounded-full mt-2 overflow-hidden">
            <div className="bg-emerald-600 h-full rounded-full" style={{ width: `${saudizationRate}%` }} />
          </div>
        </div>

        {/* Nitaqat Tier */}
        <div className="rounded-xl border bg-card p-4 shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.hr.kpi_nitaqat_tier}</span>
            <Award className="h-4 w-4 text-amber-500" />
          </div>
          <div className="mt-2">
            <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold border ${currentTier.color}`}>
              {currentTier.label}
            </span>
          </div>
          <div className="mt-2 text-[11px] text-muted-foreground">متوافق مع نسب التوطين الوزارية</div>
        </div>

        {/* Active Contracts */}
        <div className="rounded-xl border bg-card p-4 shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.hr.kpi_active_contracts}</span>
            <FileCheck2 className="h-4 w-4 text-primary" />
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold">{contracts.length}</span>
            <span className="text-xs text-emerald-600 font-medium">100% موثقة بقوى</span>
          </div>
          <div className="mt-2 text-[11px] text-muted-foreground">شاملة تفكيك الرواتب والبدلات</div>
        </div>

        {/* Total Payroll */}
        <div className="rounded-xl border bg-card p-4 shadow-sm relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">مسير الرواتب الشهري</span>
            <TrendingUp className="h-4 w-4 text-primary" />
          </div>
          <div className="mt-2 flex items-baseline gap-1">
            <span className="text-2xl font-bold">{totalPayroll.toLocaleString()}</span>
            <span className="text-xs font-semibold text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 text-[11px] text-muted-foreground">{vacantPositions} شواغر متاحة للتوظيف</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b space-x-1 rtl:space-x-reverse">
        <button
          onClick={() => setActiveTab("employees")}
          className={`px-4 py-2.5 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
            activeTab === "employees" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Users className="h-4 w-4" />
          {t.hr.tab_employees}
          <span className="ms-1.5 px-1.5 py-0.2 rounded-full text-[10px] bg-muted font-bold">{employees.length}</span>
        </button>

        <button
          onClick={() => setActiveTab("contracts")}
          className={`px-4 py-2.5 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
            activeTab === "contracts" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileCheck2 className="h-4 w-4" />
          {t.hr.tab_contracts}
          <span className="ms-1.5 px-1.5 py-0.2 rounded-full text-[10px] bg-muted font-bold">{contracts.length}</span>
        </button>

        <button
          onClick={() => setActiveTab("positions")}
          className={`px-4 py-2.5 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
            activeTab === "positions" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Briefcase className="h-4 w-4" />
          {t.hr.tab_positions}
          <span className="ms-1.5 px-1.5 py-0.2 rounded-full text-[10px] bg-muted font-bold">{positions.length}</span>
        </button>

        <button
          onClick={() => setActiveTab("saudization")}
          className={`px-4 py-2.5 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
            activeTab === "saudization" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Award className="h-4 w-4" />
          {t.hr.tab_saudization}
        </button>
      </div>

      {/* TAB 1: EMPLOYEES */}
      {activeTab === "employees" && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-card p-3 rounded-lg border">
            <div className="relative w-full sm:w-72">
              <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="بحث برقم الموظف، الاسم، أو الهوية..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full rounded-md border bg-background ps-9 pe-3 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>

            <div className="flex items-center gap-2 w-full sm:w-auto">
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="text-xs rounded-md border bg-background px-3 py-1.5 focus:outline-none"
              >
                <option value="ALL">جميع الحالات</option>
                <option value="ACTIVE">نشط (Active)</option>
                <option value="PROBATION">تحت التجربة</option>
                <option value="TERMINATED">منتهي الخدمة</option>
              </select>
            </div>
          </div>

          <div className="rounded-lg border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="p-3 text-start">{t.hr.col_emp_num}</th>
                    <th className="p-3 text-start">{t.hr.col_emp_name}</th>
                    <th className="p-3 text-start">{t.hr.col_identity}</th>
                    <th className="p-3 text-start">{t.hr.col_nationality}</th>
                    <th className="p-3 text-start">{t.hr.col_position}</th>
                    <th className="p-3 text-start">{t.hr.col_department}</th>
                    <th className="p-3 text-start">الآيبان البنكي (WPS)</th>
                    <th className="p-3 text-start">{t.hr.col_hire_date}</th>
                    <th className="p-3 text-start">{t.hr.col_status}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredEmployees.map((emp) => (
                    <tr key={emp.id} className="hover:bg-muted/30 transition-colors">
                      <td className="p-3 font-mono font-medium text-primary">{emp.employeeNumber}</td>
                      <td className="p-3">
                        <div className="font-semibold text-foreground">
                          {locale === "ar" ? `${emp.firstNameAr} ${emp.lastNameAr}` : `${emp.firstNameEn} ${emp.lastNameEn}`}
                        </div>
                        <div className="text-[10px] text-muted-foreground">
                          {locale === "ar" ? `${emp.firstNameEn} ${emp.lastNameEn}` : `${emp.firstNameAr} ${emp.lastNameAr}`}
                        </div>
                      </td>
                      <td className="p-3">
                        <div className="flex items-center gap-1.5 font-mono">
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-muted">
                            {emp.idType === "NATIONAL_ID" ? "هوية وطنية" : "إقامة"}
                          </span>
                          <span>{emp.idNumberMasked}</span>
                        </div>
                      </td>
                      <td className="p-3">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                            emp.isSaudi
                              ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                              : "bg-blue-500/10 text-blue-600 border-blue-500/20"
                          }`}
                        >
                          {emp.isSaudi ? "🇸🇦 سعودي" : `🌐 ${emp.nationalityCode}`}
                        </span>
                      </td>
                      <td className="p-3 font-medium">{locale === "ar" ? emp.jobTitleAr : emp.jobTitleEn}</td>
                      <td className="p-3 text-muted-foreground">{emp.department}</td>
                      <td className="p-3">
                        <div className="font-mono text-[11px] text-muted-foreground">{emp.ibanMasked}</div>
                        {emp.isWpsCompliant && (
                          <span className="inline-flex items-center gap-1 text-[10px] text-emerald-600">
                            <CheckCircle2 className="h-3 w-3" /> متوافق مع حماية الأجور
                          </span>
                        )}
                      </td>
                      <td className="p-3 font-mono">{emp.hireDate}</td>
                      <td className="p-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                            emp.status === "ACTIVE"
                              ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                              : "bg-muted text-muted-foreground"
                          }`}
                        >
                          {emp.status}
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

      {/* TAB 2: CONTRACTS */}
      {activeTab === "contracts" && (
        <div className="space-y-4">
          <div className="rounded-lg border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="p-3 text-start">{t.hr.col_contract_num}</th>
                    <th className="p-3 text-start">{t.hr.col_emp_name}</th>
                    <th className="p-3 text-start">{t.hr.col_contract_type}</th>
                    <th className="p-3 text-start">{t.hr.col_basic_salary}</th>
                    <th className="p-3 text-start">{t.hr.col_housing}</th>
                    <th className="p-3 text-start">{t.hr.col_transport}</th>
                    <th className="p-3 text-start">بدلات أخرى</th>
                    <th className="p-3 text-start font-bold text-foreground">{t.hr.col_gross_salary}</th>
                    <th className="p-3 text-start">الإصدار</th>
                    <th className="p-3 text-start">{t.hr.col_probation}</th>
                    <th className="p-3 text-start">{t.hr.col_actions}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {contracts.map((cnt) => (
                    <tr key={cnt.id} className="hover:bg-muted/30 transition-colors">
                      <td className="p-3 font-mono font-medium text-primary">
                        <div className="flex items-center gap-1.5">
                          <FileCheck2 className="h-3.5 w-3.5 text-primary" />
                          <span>{cnt.contractNumber}</span>
                        </div>
                      </td>
                      <td className="p-3">
                        <div className="font-semibold text-foreground">{locale === "ar" ? cnt.employeeNameAr : cnt.employeeNameEn}</div>
                        <div className="font-mono text-[10px] text-muted-foreground">{cnt.employeeNumber}</div>
                      </td>
                      <td className="p-3">
                        <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-muted">
                          {cnt.contractType === "UNSPECIFIED_TERM" ? "غير محدد المدة" : "محدد المدة"}
                        </span>
                      </td>
                      <td className="p-3 font-mono">{cnt.basicSalary.toLocaleString()} SAR</td>
                      <td className="p-3 font-mono text-muted-foreground">{cnt.housingAllowance.toLocaleString()} SAR</td>
                      <td className="p-3 font-mono text-muted-foreground">{cnt.transportAllowance.toLocaleString()} SAR</td>
                      <td className="p-3 font-mono text-muted-foreground">{cnt.otherAllowances.toLocaleString()} SAR</td>
                      <td className="p-3 font-mono font-bold text-emerald-600">{cnt.grossSalary.toLocaleString()} SAR</td>
                      <td className="p-3">
                        <span className="px-2 py-0.5 rounded-full bg-primary/10 text-primary font-bold text-[10px]">
                          v{cnt.versionNumber}
                        </span>
                      </td>
                      <td className="p-3 text-muted-foreground">{cnt.probationDays} يوم</td>
                      <td className="p-3">
                        <button
                          onClick={() => {
                            setSelectedContract(cnt);
                            setAmendForm({
                              newBasicSalary: cnt.basicSalary + 1000,
                              newHousing: cnt.housingAllowance + 250,
                              newTransport: cnt.transportAllowance,
                              newOther: cnt.otherAllowances,
                              reason: "ANNUAL_MERIT_INCREASE",
                            });
                            setShowAmendModal(true);
                          }}
                          className="px-2.5 py-1 rounded bg-secondary text-secondary-foreground text-[10px] font-semibold hover:bg-muted transition-colors border"
                        >
                          تعديل / ترقية
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

      {/* TAB 3: POSITIONS & JOBS */}
      {activeTab === "positions" && (
        <div className="space-y-4">
          <div className="rounded-lg border bg-card overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-start">
                <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
                  <tr>
                    <th className="p-3 text-start">رمز المنصب (Code)</th>
                    <th className="p-3 text-start">المسمى الوظيفي</th>
                    <th className="p-3 text-start">العائلة الوظيفية</th>
                    <th className="p-3 text-start">الإدارة / القسم</th>
                    <th className="p-3 text-start">المستهدف / الفعلي</th>
                    <th className="p-3 text-start">الحالة</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {positions.map((pos) => (
                    <tr key={pos.id} className="hover:bg-muted/30 transition-colors">
                      <td className="p-3 font-mono font-medium text-primary">{pos.code}</td>
                      <td className="p-3 font-semibold">{locale === "ar" ? pos.jobTitleAr : pos.jobTitleEn}</td>
                      <td className="p-3">
                        <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-muted">{pos.jobFamily}</span>
                      </td>
                      <td className="p-3 text-muted-foreground">{pos.department}</td>
                      <td className="p-3 font-mono">
                        <span className="font-bold">{pos.headcountCurrent}</span> / {pos.headcountTarget}
                      </td>
                      <td className="p-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                            pos.status === "FILLED"
                              ? "bg-emerald-500/10 text-emerald-600 border-emerald-500/20"
                              : "bg-amber-500/10 text-amber-600 border-amber-500/20"
                          }`}
                        >
                          {pos.status === "FILLED" ? "شاغر مكتمل (FILLED)" : "شاغر متاح (VACANT)"}
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

      {/* TAB 4: SAUDIZATION & NITAQAT */}
      {activeTab === "saudization" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Gauge & Metrics Card */}
          <div className="rounded-xl border bg-card p-6 shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <h3 className="font-bold text-sm">مؤشر نطاقات والتوطين (MHRSD / Qiwa)</h3>
                <Award className="h-5 w-5 text-amber-500" />
              </div>
              <p className="text-xs text-muted-foreground mt-1">تقييم النطاق الآلي بناء على نسبة القوى العاملة الوطنية</p>

              <div className="mt-8 text-center">
                <div className="inline-flex flex-col items-center justify-center h-36 w-36 rounded-full border-4 border-emerald-500 bg-emerald-500/5">
                  <span className="text-3xl font-black text-emerald-600">{saudizationRate}%</span>
                  <span className="text-xs font-semibold text-muted-foreground mt-1">نسبة السعودة</span>
                </div>

                <div className="mt-4">
                  <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-bold border ${currentTier.color}`}>
                    {currentTier.label}
                  </span>
                </div>
              </div>
            </div>

            <div className="mt-6 pt-4 border-t space-y-2 text-xs">
              <div className="flex justify-between">
                <span className="text-muted-foreground">عدد الموظفين السعوديين:</span>
                <span className="font-bold font-mono">{saudiEmployees}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">عدد الموظفين الوافدين:</span>
                <span className="font-bold font-mono">{expatEmployees}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">إجمالي القوى العاملة:</span>
                <span className="font-bold font-mono">{totalEmployees}</span>
              </div>
            </div>
          </div>

          {/* Rules & Benefits */}
          <div className="rounded-xl border bg-card p-6 shadow-sm lg:col-span-2 space-y-4">
            <h3 className="font-bold text-sm">مميزات النطاق البلاتيني والمرتفع وفق وزارة الموارد البشرية</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
              <div className="p-3.5 rounded-lg border bg-muted/20 space-y-1">
                <div className="font-semibold text-emerald-600 flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4" />
                  إصدار وتجديد رخص العمل فوري
                </div>
                <p className="text-muted-foreground">الاستفادة من تسهيلات منصة قوى دون انتظار الفترات البينية.</p>
              </div>

              <div className="p-3.5 rounded-lg border bg-muted/20 space-y-1">
                <div className="font-semibold text-emerald-600 flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4" />
                  نقل الخدمات والتأشيرات التعويضية
                </div>
                <p className="text-muted-foreground">صلاحية نقل خدمات العمالة وإصدار التأشيرات للمهن التخصصية بسهولة.</p>
              </div>

              <div className="p-3.5 rounded-lg border bg-muted/20 space-y-1">
                <div className="font-semibold text-emerald-600 flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4" />
                  حماية الخصوصية الوطنية
                </div>
                <p className="text-muted-foreground">حجب أرقام الهويات الوطنية والآيبان البنكي عبر النظام لحماية بيانات الموظفين.</p>
              </div>

              <div className="p-3.5 rounded-lg border bg-muted/20 space-y-1">
                <div className="font-semibold text-emerald-600 flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4" />
                  الربط المستقبلي مع التأمينات (GOSI)
                </div>
                <p className="text-muted-foreground">تفكيك الراتب (أساسي، سكن، نقل) يجهز المنشأة لاحتساب اشتراكات التأمينات بدقة.</p>
              </div>
            </div>

            {/* Nitaqat Scale Breakdown */}
            <div className="mt-4 pt-4 border-t">
              <h4 className="font-semibold text-xs mb-3">توزيع النطاقات المعتمدة:</h4>
              <div className="space-y-2 text-xs">
                <div className="flex items-center justify-between p-2 rounded bg-emerald-500/10 text-emerald-700">
                  <span className="font-semibold">النطاق البلاتيني (Platinum)</span>
                  <span className="font-mono font-bold">50% فأكثر (حالتك الحالية)</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded bg-green-500/10 text-green-700">
                  <span>الأخضر المرتفع (High Green)</span>
                  <span className="font-mono">40% - 49.9%</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded bg-lime-500/10 text-lime-700">
                  <span>الأخضر المتوسط (Mid Green)</span>
                  <span className="font-mono">25% - 39.9%</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded bg-amber-500/10 text-amber-700">
                  <span>الأخضر المنخفض (Low Green)</span>
                  <span className="font-mono">10% - 24.9%</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded bg-red-500/10 text-red-700">
                  <span>النطاق الأحمر (Red Tier)</span>
                  <span className="font-mono">أقل من 10% (إيقاف خدمات قوى)</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: ONBOARD EMPLOYEE */}
      {showNewEmpModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-xl rounded-xl border shadow-2xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-base font-bold flex items-center gap-2">
                <UserCheck className="h-5 w-5 text-primary" />
                تعيين موظف جديد وتوثيق الهوية
              </h3>
              <button onClick={() => setShowNewEmpModal(false)} className="text-muted-foreground hover:text-foreground">
                ✕
              </button>
            </div>

            <form onSubmit={handleCreateEmployee} className="space-y-3.5 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">الاسم الأول (بالعربي)</label>
                  <input
                    required
                    type="text"
                    value={newEmpForm.firstNameAr}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, firstNameAr: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-ring"
                    placeholder="مثال: فيصل"
                  />
                </div>
                <div>
                  <label className="block font-semibold mb-1">اسم العائلة (بالعربي)</label>
                  <input
                    required
                    type="text"
                    value={newEmpForm.lastNameAr}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, lastNameAr: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-ring"
                    placeholder="مثال: الدوسري"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">First Name (English)</label>
                  <input
                    required
                    type="text"
                    value={newEmpForm.firstNameEn}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, firstNameEn: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-ring"
                    placeholder="e.g. Faisal"
                  />
                </div>
                <div>
                  <label className="block font-semibold mb-1">Last Name (English)</label>
                  <input
                    required
                    type="text"
                    value={newEmpForm.lastNameEn}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, lastNameEn: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-ring"
                    placeholder="e.g. Al-Dossari"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">الجنسية</label>
                  <select
                    value={newEmpForm.nationalityCode}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, nationalityCode: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                  >
                    <option value="SA">🇸🇦 المملكة العربية السعودية (SA)</option>
                    <option value="EG">🇪🇬 مصر (EG)</option>
                    <option value="JO">🇯🇴 الأردن (JO)</option>
                    <option value="IN">🇮🇳 الهند (IN)</option>
                    <option value="PK">🇵🇰 باكستان (PK)</option>
                  </select>
                </div>
                <div>
                  <label className="block font-semibold mb-1">
                    {newEmpForm.nationalityCode === "SA" ? "رقم الهوية الوطنية (10 أرقام)" : "رقم الإقامة (10 أرقام)"}
                  </label>
                  <input
                    required
                    type="text"
                    maxLength={10}
                    value={newEmpForm.idNumber}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, idNumber: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                    placeholder="10XXXXXXXX"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">المسمى الوظيفي</label>
                  <input
                    type="text"
                    value={newEmpForm.jobTitleAr}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, jobTitleAr: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block font-semibold mb-1">الإدارة / القسم</label>
                  <input
                    type="text"
                    value={newEmpForm.department}
                    onChange={(e) => setNewEmpForm({ ...newEmpForm, department: e.target.value })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block font-semibold mb-1">الآيبان البنكي المعتمد لحماية الأجور (WPS IBAN)</label>
                <input
                  required
                  type="text"
                  value={newEmpForm.iban}
                  onChange={(e) => setNewEmpForm({ ...newEmpForm, iban: e.target.value })}
                  className="w-full rounded-md border bg-background px-3 py-1.5 font-mono focus:outline-none"
                  placeholder="SA44..."
                />
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setShowNewEmpModal(false)}
                  className="px-4 py-2 rounded-md border text-xs font-semibold hover:bg-muted"
                >
                  إلغاء
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90"
                >
                  حفظ وتعيين الموظف
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW CONTRACT */}
      {showNewContractModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-lg rounded-xl border shadow-2xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-base font-bold flex items-center gap-2">
                <FileCheck2 className="h-5 w-5 text-primary" />
                إنشاء عقد عمل موحد (قوى)
              </h3>
              <button onClick={() => setShowNewContractModal(false)} className="text-muted-foreground hover:text-foreground">
                ✕
              </button>
            </div>

            <form onSubmit={handleCreateContract} className="space-y-3.5 text-xs">
              <div>
                <label className="block font-semibold mb-1">اختيار الموظف</label>
                <select
                  value={newContractForm.employeeNumber}
                  onChange={(e) => setNewContractForm({ ...newContractForm, employeeNumber: e.target.value })}
                  className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                >
                  {employees.map((emp) => (
                    <option key={emp.id} value={emp.employeeNumber}>
                      {emp.employeeNumber} - {emp.firstNameAr} {emp.lastNameAr} ({emp.isSaudi ? "سعودي" : "وافد"})
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">نوع العقد</label>
                  <select
                    value={newContractForm.contractType}
                    onChange={(e) => setNewContractForm({ ...newContractForm, contractType: e.target.value as any })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                  >
                    <option value="UNSPECIFIED_TERM">غير محدد المدة (للسعوديين)</option>
                    <option value="SPECIFIED_TERM">محدد المدة (للوافدين أو محدد)</option>
                  </select>
                </div>
                <div>
                  <label className="block font-semibold mb-1">فترة التجربة (أيام - حد أقصى 180)</label>
                  <input
                    type="number"
                    max={180}
                    value={newContractForm.probationDays}
                    onChange={(e) => setNewContractForm({ ...newContractForm, probationDays: Number(e.target.value) })}
                    className="w-full rounded-md border bg-background px-3 py-1.5 focus:outline-none"
                  />
                </div>
              </div>

              <div className="p-3 bg-muted/20 border rounded-lg space-y-2.5">
                <div className="font-semibold text-primary">تفكيك الراتب والبدلات التعاقدية (SAR):</div>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-muted-foreground mb-0.5">الراتب الأساسي</label>
                    <input
                      type="number"
                      required
                      value={newContractForm.basicSalary}
                      onChange={(e) => setNewContractForm({ ...newContractForm, basicSalary: Number(e.target.value) })}
                      className="w-full rounded border bg-background px-2.5 py-1 font-mono focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-muted-foreground mb-0.5">بدل السكن</label>
                    <input
                      type="number"
                      value={newContractForm.housingAllowance}
                      onChange={(e) => setNewContractForm({ ...newContractForm, housingAllowance: Number(e.target.value) })}
                      className="w-full rounded border bg-background px-2.5 py-1 font-mono focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-muted-foreground mb-0.5">بدل النقل</label>
                    <input
                      type="number"
                      value={newContractForm.transportAllowance}
                      onChange={(e) => setNewContractForm({ ...newContractForm, transportAllowance: Number(e.target.value) })}
                      className="w-full rounded border bg-background px-2.5 py-1 font-mono focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-muted-foreground mb-0.5">بدلات أخرى</label>
                    <input
                      type="number"
                      value={newContractForm.otherAllowances}
                      onChange={(e) => setNewContractForm({ ...newContractForm, otherAllowances: Number(e.target.value) })}
                      className="w-full rounded border bg-background px-2.5 py-1 font-mono focus:outline-none"
                    />
                  </div>
                </div>

                <div className="flex justify-between items-center pt-2 border-t font-bold">
                  <span>إجمالي الراتب التعاقدي:</span>
                  <span className="font-mono text-emerald-600 text-sm">
                    {(
                      Number(newContractForm.basicSalary) +
                      Number(newContractForm.housingAllowance) +
                      Number(newContractForm.transportAllowance) +
                      Number(newContractForm.otherAllowances)
                    ).toLocaleString()}{" "}
                    SAR
                  </span>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setShowNewContractModal(false)}
                  className="px-4 py-2 rounded-md border text-xs font-semibold hover:bg-muted"
                >
                  إلغاء
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90"
                >
                  توثيق العقد
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: AMEND CONTRACT / SALARY RAISE */}
      {showAmendModal && selectedContract && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-md rounded-xl border shadow-2xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-base font-bold flex items-center gap-2">
                <Sparkles className="h-5 w-5 text-amber-500" />
                تعديل بنود العقد وترقية الراتب (إصدار جديد)
              </h3>
              <button onClick={() => setShowAmendModal(false)} className="text-muted-foreground hover:text-foreground">
                ✕
              </button>
            </div>

            <form onSubmit={handleAmendContract} className="space-y-3.5 text-xs">
              <div className="p-2.5 bg-muted/40 rounded-lg text-xs space-y-1">
                <div>
                  <span className="text-muted-foreground">الموظف: </span>
                  <span className="font-bold">{locale === "ar" ? selectedContract.employeeNameAr : selectedContract.employeeNameEn}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">رقم العقد: </span>
                  <span className="font-mono">{selectedContract.contractNumber}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">الإصدار الحالي: </span>
                  <span className="font-bold">v{selectedContract.versionNumber}</span>
                  <span className="text-primary font-bold ms-2">→ سيتم إنشاء الإصدار v{selectedContract.versionNumber + 1}</span>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">الراتب الأساسي الجديد</label>
                  <input
                    type="number"
                    required
                    value={amendForm.newBasicSalary}
                    onChange={(e) => setAmendForm({ ...amendForm, newBasicSalary: Number(e.target.value) })}
                    className="w-full rounded border bg-background px-3 py-1.5 font-mono focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block font-semibold mb-1">بدل السكن الجديد</label>
                  <input
                    type="number"
                    value={amendForm.newHousing}
                    onChange={(e) => setAmendForm({ ...amendForm, newHousing: Number(e.target.value) })}
                    className="w-full rounded border bg-background px-3 py-1.5 font-mono focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold mb-1">بدل النقل الجديد</label>
                  <input
                    type="number"
                    value={amendForm.newTransport}
                    onChange={(e) => setAmendForm({ ...amendForm, newTransport: Number(e.target.value) })}
                    className="w-full rounded border bg-background px-3 py-1.5 font-mono focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block font-semibold mb-1">بدلات أخرى</label>
                  <input
                    type="number"
                    value={amendForm.newOther}
                    onChange={(e) => setAmendForm({ ...amendForm, newOther: Number(e.target.value) })}
                    className="w-full rounded border bg-background px-3 py-1.5 font-mono focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex justify-between items-center p-2.5 rounded bg-emerald-500/10 border border-emerald-500/20 font-bold">
                <span>إجمالي الراتب الجديد:</span>
                <span className="font-mono text-emerald-600 text-sm">
                  {(
                    Number(amendForm.newBasicSalary) +
                    Number(amendForm.newHousing) +
                    Number(amendForm.newTransport) +
                    Number(amendForm.newOther)
                  ).toLocaleString()}{" "}
                  SAR
                </span>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setShowAmendModal(false)}
                  className="px-4 py-2 rounded-md border text-xs font-semibold hover:bg-muted"
                >
                  إلغاء
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-md bg-primary text-primary-foreground text-xs font-semibold shadow hover:bg-primary/90"
                >
                  اعتماد التعديل والإصدار
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

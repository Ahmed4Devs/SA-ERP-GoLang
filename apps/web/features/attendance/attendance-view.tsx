"use client";

import React, { useState } from "react";
import {
  Clock,
  Calendar,
  UserCheck,
  AlertTriangle,
  Plus,
  Search,
  CheckCircle2,
  XCircle,
  TrendingUp,
  Moon,
  Coffee,
  ShieldCheck,
  Building2,
  Users,
  Timer,
  FileEdit,
  Sparkles,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface ShiftItem {
  id: string;
  nameEn: string;
  nameAr: string;
  startTime: string;
  endTime: string;
  gracePeriodMins: number;
  breakDurationMins: number;
  isRamadan: boolean;
  activeEmployeesCount: number;
}

interface AttendanceRecordItem {
  id: string;
  employeeId: string;
  employeeNameEn: string;
  employeeNameAr: string;
  shiftNameEn: string;
  shiftNameAr: string;
  workDate: string;
  clockIn: string | null;
  clockOut: string | null;
  totalWorkHours: number;
  lateArrivalMinutes: number;
  earlyDepartureMinutes: number;
  overtimeHours: number;
  status: "PRESENT" | "LATE" | "ABSENT" | "HALF_DAY";
}

interface OvertimeItem {
  id: string;
  employeeId: string;
  employeeNameEn: string;
  employeeNameAr: string;
  workDate: string;
  requestedHours: number;
  approvedHours: number | null;
  rateMultiplier: number;
  reason: string;
  status: "PENDING" | "APPROVED" | "REJECTED";
}

interface AdjustmentItem {
  id: string;
  recordId: string;
  employeeNameEn: string;
  employeeNameAr: string;
  adjustedClockIn: string;
  adjustedClockOut: string;
  reason: string;
  status: "PENDING" | "APPROVED" | "REJECTED";
  requestedBy: string;
  createdAt: string;
}

export function AttendanceView() {
  const { t, locale, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"records" | "shifts" | "overtime" | "adjustments">("records");

  // State: Shifts
  const [shifts, setShifts] = useState<ShiftItem[]>([
    {
      id: "sh-1",
      nameEn: "Standard Morning Shift",
      nameAr: "الفترة الصباحية العامة",
      startTime: "08:00",
      endTime: "17:00",
      gracePeriodMins: 15,
      breakDurationMins: 60,
      isRamadan: false,
      activeEmployeesCount: 42,
    },
    {
      id: "sh-2",
      nameEn: "Evening Shift",
      nameAr: "الفترة المسائية",
      startTime: "16:00",
      endTime: "00:00",
      gracePeriodMins: 10,
      breakDurationMins: 45,
      isRamadan: false,
      activeEmployeesCount: 18,
    },
    {
      id: "sh-3",
      nameEn: "Ramadan Special Shift (6 Hours)",
      nameAr: "فترة شهر رمضان المبارك (6 ساعات عمل)",
      startTime: "10:00",
      endTime: "16:00",
      gracePeriodMins: 20,
      breakDurationMins: 30,
      isRamadan: true,
      activeEmployeesCount: 60,
    },
  ]);

  // State: Attendance Records
  const [records, setRecords] = useState<AttendanceRecordItem[]>([
    {
      id: "rec-1",
      employeeId: "EMP-2025-001",
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      shiftNameEn: "Standard Morning Shift",
      shiftNameAr: "الفترة الصباحية العامة",
      workDate: "2025-05-18",
      clockIn: "08:05",
      clockOut: "17:02",
      totalWorkHours: 8.0,
      lateArrivalMinutes: 0,
      earlyDepartureMinutes: 0,
      overtimeHours: 0,
      status: "PRESENT",
    },
    {
      id: "rec-2",
      employeeId: "EMP-2025-002",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      shiftNameEn: "Standard Morning Shift",
      shiftNameAr: "الفترة الصباحية العامة",
      workDate: "2025-05-18",
      clockIn: "08:25",
      clockOut: "17:30",
      totalWorkHours: 8.0,
      lateArrivalMinutes: 10, // 25 - 15 grace
      earlyDepartureMinutes: 0,
      overtimeHours: 0.5,
      status: "LATE",
    },
    {
      id: "rec-3",
      employeeId: "EMP-2025-003",
      employeeNameEn: "Tariq Mansour",
      employeeNameAr: "طارق منصور",
      shiftNameEn: "Evening Shift",
      shiftNameAr: "الفترة المسائية",
      workDate: "2025-05-18",
      clockIn: "16:02",
      clockOut: "00:05",
      totalWorkHours: 7.25,
      lateArrivalMinutes: 0,
      earlyDepartureMinutes: 0,
      overtimeHours: 0,
      status: "PRESENT",
    },
    {
      id: "rec-4",
      employeeId: "EMP-2025-004",
      employeeNameEn: "Reem Al-Dossary",
      employeeNameAr: "ريم الدوسري",
      shiftNameEn: "Standard Morning Shift",
      shiftNameAr: "الفترة الصباحية العامة",
      workDate: "2025-05-18",
      clockIn: null,
      clockOut: null,
      totalWorkHours: 0,
      lateArrivalMinutes: 0,
      earlyDepartureMinutes: 0,
      overtimeHours: 0,
      status: "ABSENT",
    },
  ]);

  // State: Overtime Requests (Art 107 - 1.5x)
  const [overtimes, setOvertimes] = useState<OvertimeItem[]>([
    {
      id: "ot-1",
      employeeId: "EMP-2025-002",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      workDate: "2025-05-18",
      requestedHours: 2.5,
      approvedHours: 2.5,
      rateMultiplier: 1.5,
      reason: "Emergency software release & cloud deployment / نشر تحديث برمجي طارئ",
      status: "APPROVED",
    },
    {
      id: "ot-2",
      employeeId: "EMP-2025-003",
      employeeNameEn: "Tariq Mansour",
      employeeNameAr: "طارق منصور",
      workDate: "2025-05-19",
      requestedHours: 3.0,
      approvedHours: null,
      rateMultiplier: 1.5,
      reason: "Month-end physical inventory audit support / جرد المستودعات الشهري",
      status: "PENDING",
    },
  ]);

  // State: Adjustments
  const [adjustments, setAdjustments] = useState<AdjustmentItem[]>([
    {
      id: "adj-1",
      recordId: "rec-4",
      employeeNameEn: "Reem Al-Dossary",
      employeeNameAr: "ريم الدوسري",
      adjustedClockIn: "08:10",
      adjustedClockOut: "17:00",
      reason: "Biometric scanner power outage at main entrance / عطل في جهاز البصمة عند المدخل",
      status: "APPROVED",
      requestedBy: "Admin / مدير النظام",
      createdAt: "2025-05-18 10:15",
    },
  ]);

  // Search filter
  const [searchTerm, setSearchTerm] = useState("");

  // Modals state
  const [showClockModal, setShowClockModal] = useState(false);
  const [clockType, setClockType] = useState<"IN" | "OUT">("IN");
  const [selectedEmpClock, setSelectedEmpClock] = useState("EMP-2025-001");

  const [showNewShiftModal, setShowNewShiftModal] = useState(false);
  const [newShiftData, setNewShiftData] = useState({
    nameEn: "",
    nameAr: "",
    startTime: "08:00",
    endTime: "16:00",
    gracePeriodMins: 15,
    breakDurationMins: 60,
    isRamadan: false,
  });

  const [showOTModal, setShowOTModal] = useState(false);
  const [newOTData, setNewOTData] = useState({
    employeeId: "EMP-2025-001",
    workDate: new Date().toISOString().split("T")[0],
    hours: 2.0,
    reason: "",
  });

  const [showAdjustModal, setShowAdjustModal] = useState(false);
  const [newAdjustData, setNewAdjustData] = useState({
    employeeName: "Reem Al-Dossary",
    adjustedClockIn: "08:00",
    adjustedClockOut: "17:00",
    reason: "",
  });

  // Clock in/out handler
  const handleClockAction = (e: React.FormEvent) => {
    e.preventDefault();
    const timeNow = new Date().toTimeString().slice(0, 5);
    const todayStr = new Date().toISOString().split("T")[0];

    const empObj = records.find((r) => r.employeeId === selectedEmpClock) || {
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      shiftNameEn: "Standard Morning Shift",
      shiftNameAr: "الفترة الصباحية العامة",
    };

    if (clockType === "IN") {
      const existing = records.find((r) => r.employeeId === selectedEmpClock && r.workDate === todayStr);
      if (existing) {
        setRecords(
          records.map((r) =>
            r.id === existing.id
              ? { ...r, clockIn: timeNow, status: "PRESENT" }
              : r
          )
        );
      } else {
        setRecords([
          {
            id: `rec-${Date.now()}`,
            employeeId: selectedEmpClock,
            employeeNameEn: empObj.employeeNameEn,
            employeeNameAr: empObj.employeeNameAr,
            shiftNameEn: empObj.shiftNameEn,
            shiftNameAr: empObj.shiftNameAr,
            workDate: todayStr,
            clockIn: timeNow,
            clockOut: null,
            totalWorkHours: 0,
            lateArrivalMinutes: 0,
            earlyDepartureMinutes: 0,
            overtimeHours: 0,
            status: "PRESENT",
          },
          ...records,
        ]);
      }
    } else {
      setRecords(
        records.map((r) =>
          r.employeeId === selectedEmpClock
            ? { ...r, clockOut: timeNow, totalWorkHours: 8.0 }
            : r
        )
      );
    }
    setShowClockModal(false);
  };

  // Add Shift handler
  const handleCreateShift = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newShiftData.nameEn || !newShiftData.nameAr) return;
    const created: ShiftItem = {
      id: `sh-${Date.now()}`,
      nameEn: newShiftData.nameEn,
      nameAr: newShiftData.nameAr,
      startTime: newShiftData.startTime,
      endTime: newShiftData.endTime,
      gracePeriodMins: Number(newShiftData.gracePeriodMins),
      breakDurationMins: Number(newShiftData.breakDurationMins),
      isRamadan: newShiftData.isRamadan,
      activeEmployeesCount: 0,
    };
    setShifts([...shifts, created]);
    setShowNewShiftModal(false);
    setNewShiftData({
      nameEn: "",
      nameAr: "",
      startTime: "08:00",
      endTime: "16:00",
      gracePeriodMins: 15,
      breakDurationMins: 60,
      isRamadan: false,
    });
  };

  // Overtime request handler
  const handleRequestOT = (e: React.FormEvent) => {
    e.preventDefault();
    const emp = records.find((r) => r.employeeId === newOTData.employeeId) || {
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
    };
    const created: OvertimeItem = {
      id: `ot-${Date.now()}`,
      employeeId: newOTData.employeeId,
      employeeNameEn: emp.employeeNameEn,
      employeeNameAr: emp.employeeNameAr,
      workDate: newOTData.workDate,
      requestedHours: Number(newOTData.hours),
      approvedHours: null,
      rateMultiplier: 1.5,
      reason: newOTData.reason,
      status: "PENDING",
    };
    setOvertimes([created, ...overtimes]);
    setShowOTModal(false);
    setNewOTData({
      employeeId: "EMP-2025-001",
      workDate: new Date().toISOString().split("T")[0],
      hours: 2.0,
      reason: "",
    });
  };

  // Overtime decision
  const handleOTDecision = (id: string, decision: "APPROVED" | "REJECTED") => {
    setOvertimes(
      overtimes.map((ot) =>
        ot.id === id
          ? {
              ...ot,
              status: decision,
              approvedHours: decision === "APPROVED" ? ot.requestedHours : 0,
            }
          : ot
      )
    );
  };

  // Add adjustment handler
  const handleAddAdjustment = (e: React.FormEvent) => {
    e.preventDefault();
    const created: AdjustmentItem = {
      id: `adj-${Date.now()}`,
      recordId: "manual",
      employeeNameEn: newAdjustData.employeeName,
      employeeNameAr: newAdjustData.employeeName,
      adjustedClockIn: newAdjustData.adjustedClockIn,
      adjustedClockOut: newAdjustData.adjustedClockOut,
      reason: newAdjustData.reason,
      status: "APPROVED",
      requestedBy: "HR Admin / مسؤول الحضور",
      createdAt: new Date().toLocaleString(),
    };
    setAdjustments([created, ...adjustments]);
    setShowAdjustModal(false);
    setNewAdjustData({
      employeeName: "Reem Al-Dossary",
      adjustedClockIn: "08:00",
      adjustedClockOut: "17:00",
      reason: "",
    });
  };

  return (
    <div className="space-y-6 pb-12">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-blue-950 via-indigo-900 to-slate-900 p-6 md:p-8 text-white shadow-xl">
        <div className="relative z-10 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <span className="px-3 py-1 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-200 border border-blue-400/30 flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5 text-blue-400" />
                {locale === "ar" ? "المرحلة 20: الحضور والورديات" : "Phase 20: Attendance & Shifts"}
              </span>
              <span className="px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-200 border border-emerald-400/30 flex items-center gap-1">
                <Sparkles className="h-3.5 w-3.5 text-emerald-300" />
                {locale === "ar" ? "المادة 107 (معامل 1.5x)" : "Labor Law Art. 107 (1.5x)"}
              </span>
            </div>
            <h1 className="text-2xl md:text-3xl font-bold tracking-tight">
              {(t as any).attendance?.title || "إدارة الحضور والانصراف والورديات"}
            </h1>
            <p className="mt-2 text-sm md:text-base text-blue-200/80 max-w-3xl">
              {(t as any).attendance?.description ||
                "تسجيل الحضور والانصراف الذكي، إدارة المناوبات وتوزيع ساعات العمل، احتساب التأخير والانصراف المبكر، وموافقات العمل الإضافي"}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={() => {
                setClockType("IN");
                setShowClockModal(true);
              }}
              className="px-4 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-sm flex items-center gap-2 shadow-lg shadow-emerald-950/40 transition-all cursor-pointer"
            >
              <UserCheck className="h-4 w-4" />
              {(t as any).attendance?.btn_clock_in || "تسجيل حضور"}
            </button>
            <button
              onClick={() => {
                setClockType("OUT");
                setShowClockModal(true);
              }}
              className="px-4 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium text-sm flex items-center gap-2 shadow-lg shadow-blue-950/40 transition-all cursor-pointer"
            >
              <Clock className="h-4 w-4" />
              {(t as any).attendance?.btn_clock_out || "تسجيل انصراف"}
            </button>
          </div>
        </div>

        {/* Decorative background glow */}
        <div className="absolute -bottom-10 -right-10 w-64 h-64 bg-blue-500/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute top-0 -left-10 w-64 h-64 bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* KPI 1: Active Shifts */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).attendance?.kpi_active_shifts || "المناوبات المجدولة"}
            </span>
            <div className="p-2.5 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <Clock className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold">{shifts.length}</div>
            <div className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
              <Moon className="h-3.5 w-3.5 text-amber-500" />
              <span>{shifts.filter((s) => s.isRamadan).length} {locale === "ar" ? "مناوبة رمضانية خاصة" : "Ramadan special shift"}</span>
            </div>
          </div>
        </div>

        {/* KPI 2: Today Records */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).attendance?.kpi_today_records || "تسجيلات الحضور اليومية"}
            </span>
            <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <UserCheck className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-emerald-600 dark:text-emerald-400">
              {records.filter((r) => r.status === "PRESENT" || r.status === "LATE").length}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "من أصل 4 موظفين مجدولين اليوم" : "of 4 employees scheduled today"}
            </div>
          </div>
        </div>

        {/* KPI 3: Late Arrivals */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).attendance?.kpi_late_arrivals || "حالات التأخير المرصودة"}
            </span>
            <div className="p-2.5 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <AlertTriangle className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">
              {records.filter((r) => r.status === "LATE").length}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "تجاوزت فترة السماح (15 دقيقة)" : "Exceeded grace period (15m)"}
            </div>
          </div>
        </div>

        {/* KPI 4: Overtime Approved */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).attendance?.kpi_approved_ot || "ساعات الإضافي المعتمدة (1.5x)"}
            </span>
            <div className="p-2.5 rounded-lg bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
              {overtimes
                .filter((ot) => ot.status === "APPROVED")
                .reduce((acc, curr) => acc + (curr.approvedHours || 0), 0)}{" "}
              <span className="text-sm font-normal text-muted-foreground">{locale === "ar" ? "ساعة" : "hrs"}</span>
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "معامل الاحتساب: 150% من الأجر" : "Multiplier: 150% standard pay"}
            </div>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-border space-x-1 sm:space-x-2">
        <button
          onClick={() => setActiveTab("records")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "records"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Clock className="h-4 w-4" />
          {(t as any).attendance?.tab_records || "سجل الحضور اليومي"}
        </button>

        <button
          onClick={() => setActiveTab("shifts")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "shifts"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Calendar className="h-4 w-4" />
          {(t as any).attendance?.tab_shifts || "المناوبات وتوزيع الورديات"}
        </button>

        <button
          onClick={() => setActiveTab("overtime")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "overtime"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Timer className="h-4 w-4" />
          {(t as any).attendance?.tab_overtime || "طلبات العمل الإضافي (المادة 107)"}
        </button>

        <button
          onClick={() => setActiveTab("adjustments")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "adjustments"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileEdit className="h-4 w-4" />
          {(t as any).attendance?.tab_adjustments || "تسويات وتعديلات الحضور"}
        </button>
      </div>

      {/* TAB 1: DAILY ATTENDANCE RECORDS */}
      {activeTab === "records" && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="relative w-full sm:w-80">
              <Search className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground pointer-events-none" />
              <input
                type="text"
                placeholder={locale === "ar" ? "بحث بالاسم أو الرقم الوظيفي..." : "Search by employee..."}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-3 pr-9 py-2 rounded-xl border border-input bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="flex items-center gap-2 w-full sm:w-auto">
              <span className="text-xs text-muted-foreground">
                {locale === "ar" ? "التاريخ: 2025-05-18 (اليوم)" : "Date: 2025-05-18 (Today)"}
              </span>
            </div>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).attendance?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_shift || "المناوبة"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_date || "التاريخ"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_clock_in || "وقت الحضور"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_clock_out || "وقت الانصراف"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_work_hours || "ساعات العمل"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_late_mins || "دقائق التأخير"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_ot_hours || "الإضافي"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_status || "الحالة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {records
                  .filter(
                    (r) =>
                      r.employeeNameEn.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      r.employeeNameAr.includes(searchTerm) ||
                      r.employeeId.toLowerCase().includes(searchTerm.toLowerCase())
                  )
                  .map((rec) => (
                    <tr key={rec.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3.5 font-medium">
                        <div>{locale === "ar" ? rec.employeeNameAr : rec.employeeNameEn}</div>
                        <div className="text-xs text-muted-foreground font-mono">{rec.employeeId}</div>
                      </td>
                      <td className="px-4 py-3.5 text-muted-foreground">
                        {locale === "ar" ? rec.shiftNameAr : rec.shiftNameEn}
                      </td>
                      <td className="px-4 py-3.5 font-mono text-xs">{rec.workDate}</td>
                      <td className="px-4 py-3.5 font-mono font-medium">
                        {rec.clockIn ? (
                          <span className="text-emerald-600 dark:text-emerald-400">{rec.clockIn}</span>
                        ) : (
                          <span className="text-muted-foreground">--:--</span>
                        )}
                      </td>
                      <td className="px-4 py-3.5 font-mono font-medium">
                        {rec.clockOut ? (
                          <span className="text-blue-600 dark:text-blue-400">{rec.clockOut}</span>
                        ) : (
                          <span className="text-muted-foreground">--:--</span>
                        )}
                      </td>
                      <td className="px-4 py-3.5 font-semibold">{rec.totalWorkHours}h</td>
                      <td className="px-4 py-3.5">
                        {rec.lateArrivalMinutes > 0 ? (
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-amber-500/15 text-amber-600 dark:text-amber-400">
                            +{rec.lateArrivalMinutes} {locale === "ar" ? "دقيقة" : "min"}
                          </span>
                        ) : (
                          <span className="text-emerald-600 dark:text-emerald-400 text-xs font-medium">
                            {locale === "ar" ? "في الموعد" : "On time"}
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3.5 font-medium">
                        {rec.overtimeHours > 0 ? (
                          <span className="text-purple-600 dark:text-purple-400 font-semibold">
                            +{rec.overtimeHours}h
                          </span>
                        ) : (
                          "-"
                        )}
                      </td>
                      <td className="px-4 py-3.5">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
                            rec.status === "PRESENT"
                              ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                              : rec.status === "LATE"
                              ? "bg-amber-500/15 text-amber-600 dark:text-amber-400"
                              : "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                          }`}
                        >
                          {rec.status === "PRESENT" && (locale === "ar" ? "حاضر" : "Present")}
                          {rec.status === "LATE" && (locale === "ar" ? "متأخر" : "Late")}
                          {rec.status === "ABSENT" && (locale === "ar" ? "غائب" : "Absent")}
                        </span>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: SHIFTS & ROSTERS */}
      {activeTab === "shifts" && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-base font-semibold">
              {locale === "ar" ? "المناوبات ونظام الورديات المعتمد" : "Configured Work Shifts & Rosters"}
            </h3>
            <button
              onClick={() => setShowNewShiftModal(true)}
              className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium flex items-center gap-2 shadow hover:bg-primary/90 transition-all cursor-pointer"
            >
              <Plus className="h-4 w-4" />
              {(t as any).attendance?.btn_new_shift || "إضافة مناوبة جديدة"}
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {shifts.map((shift) => (
              <div
                key={shift.id}
                className="p-5 rounded-xl border border-border bg-card shadow-sm hover:shadow-md transition-all flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span
                      className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                        shift.isRamadan
                          ? "bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30"
                          : "bg-blue-500/15 text-blue-600 dark:text-blue-400"
                      }`}
                    >
                      {shift.isRamadan
                        ? locale === "ar"
                          ? "مناوبة رمضان (6 ساعات)"
                          : "Ramadan Shift (6h)"
                        : locale === "ar"
                        ? "مناوبة اعتيادية"
                        : "Standard Shift"}
                    </span>
                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                      <Users className="h-3.5 w-3.5" />
                      {shift.activeEmployeesCount} {locale === "ar" ? "موظف" : "employees"}
                    </span>
                  </div>

                  <h4 className="font-bold text-lg">
                    {locale === "ar" ? shift.nameAr : shift.nameEn}
                  </h4>
                  <div className="text-xs text-muted-foreground mt-0.5">
                    {locale === "ar" ? shift.nameEn : shift.nameAr}
                  </div>

                  <div className="mt-4 p-3 rounded-lg bg-muted/40 space-y-2 text-xs">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        {locale === "ar" ? "مواعيد العمل:" : "Working Hours:"}
                      </span>
                      <span className="font-mono font-bold">
                        {shift.startTime} - {shift.endTime}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        {locale === "ar" ? "فترة السماح:" : "Grace Period:"}
                      </span>
                      <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                        {shift.gracePeriodMins} {locale === "ar" ? "دقيقة" : "mins"}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">
                        {locale === "ar" ? "مدة الراحة والصلاة:" : "Break Duration:"}
                      </span>
                      <span className="font-semibold">{shift.breakDurationMins} {locale === "ar" ? "دقيقة" : "mins"}</span>
                    </div>
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-border flex justify-end">
                  <span className="text-xs text-primary font-medium flex items-center gap-1">
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    {locale === "ar" ? "نشطة وجاهزة للتوزيع" : "Active & Ready"}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 3: OVERTIME REQUESTS (ART 107) */}
      {activeTab === "overtime" && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-base font-semibold">
                {locale === "ar"
                  ? "طلبات العمل الإضافي وفق المادة 107 من نظام العمل السعودي"
                  : "Saudi Labor Law Article 107 Overtime Requests"}
              </h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {locale === "ar"
                  ? "يستحق العامل أجراً إضافياً يوازي أجر الساعة مضافاً إليه 50% من أجره الأساسي (المعامل 1.5x)"
                  : "Employees are compensated normal hourly wage plus 50% basic wage (1.5x multiplier)"}
              </p>
            </div>
            <button
              onClick={() => setShowOTModal(true)}
              className="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium flex items-center gap-2 shadow transition-all cursor-pointer"
            >
              <Plus className="h-4 w-4" />
              {(t as any).attendance?.btn_request_ot || "طلب عمل إضافي"}
            </button>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).attendance?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_date || "التاريخ"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "الساعات المطلوبة" : "Req. Hours"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "الساعات المعتمدة" : "Appr. Hours"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_multiplier || "معامل الأجر"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "سبب التكليف" : "Reason"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_status || "الحالة"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_actions || "الإجراءات"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {overtimes.map((ot) => (
                  <tr key={ot.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium">
                      <div>{locale === "ar" ? ot.employeeNameAr : ot.employeeNameEn}</div>
                      <div className="text-xs text-muted-foreground font-mono">{ot.employeeId}</div>
                    </td>
                    <td className="px-4 py-3.5 font-mono text-xs">{ot.workDate}</td>
                    <td className="px-4 py-3.5 font-bold">{ot.requestedHours}h</td>
                    <td className="px-4 py-3.5 font-bold text-purple-600 dark:text-purple-400">
                      {ot.approvedHours !== null ? `${ot.approvedHours}h` : "-"}
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
                        {ot.rateMultiplier.toFixed(2)}x
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-xs text-muted-foreground max-w-xs truncate">
                      {ot.reason}
                    </td>
                    <td className="px-4 py-3.5">
                      <span
                        className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
                          ot.status === "APPROVED"
                            ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                            : ot.status === "PENDING"
                            ? "bg-amber-500/15 text-amber-600 dark:text-amber-400"
                            : "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                        }`}
                      >
                        {ot.status === "APPROVED" && (locale === "ar" ? "معتمد" : "Approved")}
                        {ot.status === "PENDING" && (locale === "ar" ? "معلق" : "Pending")}
                        {ot.status === "REJECTED" && (locale === "ar" ? "مرفوض" : "Rejected")}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      {ot.status === "PENDING" ? (
                        <div className="flex items-center gap-1.5">
                          <button
                            onClick={() => handleOTDecision(ot.id, "APPROVED")}
                            className="p-1.5 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 transition-colors"
                            title={locale === "ar" ? "اعتماد الطلب" : "Approve"}
                          >
                            <CheckCircle2 className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => handleOTDecision(ot.id, "REJECTED")}
                            className="p-1.5 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 transition-colors"
                            title={locale === "ar" ? "رفض الطلب" : "Reject"}
                          >
                            <XCircle className="h-4 w-4" />
                          </button>
                        </div>
                      ) : (
                        <span className="text-xs text-muted-foreground">--</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 4: ADJUSTMENTS */}
      {activeTab === "adjustments" && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-base font-semibold">
                {locale === "ar" ? "سجل تسويات وتعديلات الحضور اليدوية" : "Attendance Adjustments Log"}
              </h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {locale === "ar"
                  ? "تسجيل التعديلات الاستثنائية (أعطال البصمة، النسيان، المهام الخارجية) مع سجل التدقيق"
                  : "Manual adjustment workflow with audit reason for scanner outage or forgotten badges"}
              </p>
            </div>
            <button
              onClick={() => setShowAdjustModal(true)}
              className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium flex items-center gap-2 shadow hover:bg-primary/90 transition-all cursor-pointer"
            >
              <Plus className="h-4 w-4" />
              {(t as any).attendance?.btn_adjust_record || "طلب تسوية حضور"}
            </button>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).attendance?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "وقت الحضور المعدل" : "Adjusted In"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "وقت الانصراف المعدل" : "Adjusted Out"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "سبب التسوية" : "Reason"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "مقدم الطلب" : "Requested By"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_status || "الحالة"}</th>
                  <th className="px-4 py-3">{(t as any).attendance?.col_date || "التاريخ"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {adjustments.map((adj) => (
                  <tr key={adj.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium">
                      {locale === "ar" ? adj.employeeNameAr : adj.employeeNameEn}
                    </td>
                    <td className="px-4 py-3.5 font-mono text-emerald-600 dark:text-emerald-400 font-medium">
                      {adj.adjustedClockIn}
                    </td>
                    <td className="px-4 py-3.5 font-mono text-blue-600 dark:text-blue-400 font-medium">
                      {adj.adjustedClockOut}
                    </td>
                    <td className="px-4 py-3.5 text-xs text-muted-foreground max-w-sm truncate">
                      {adj.reason}
                    </td>
                    <td className="px-4 py-3.5 text-xs">{adj.requestedBy}</td>
                    <td className="px-4 py-3.5">
                      <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
                        {adj.status === "APPROVED" ? (locale === "ar" ? "معتمد" : "Approved") : adj.status}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-xs text-muted-foreground font-mono">{adj.createdAt}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* MODAL 1: CLOCK IN / OUT */}
      {showClockModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-md rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              {clockType === "IN"
                ? locale === "ar"
                  ? "تسجيل حضور جديد (Clock In)"
                  : "Clock In Simulation"
                : locale === "ar"
                ? "تسجيل انصراف (Clock Out)"
                : "Clock Out Simulation"}
            </h3>
            <form onSubmit={handleClockAction} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "اختر الموظف" : "Select Employee"}
                </label>
                <select
                  value={selectedEmpClock}
                  onChange={(e) => setSelectedEmpClock(e.target.value)}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                >
                  <option value="EMP-2025-001">محمد العتيبي (EMP-2025-001)</option>
                  <option value="EMP-2025-002">سارة الغامدي (EMP-2025-002)</option>
                  <option value="EMP-2025-003">طارق منصور (EMP-2025-003)</option>
                  <option value="EMP-2025-004">ريم الدوسري (EMP-2025-004)</option>
                </select>
              </div>

              <div className="p-3 rounded-xl bg-muted/40 text-xs space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{locale === "ar" ? "الوقت الحالي:" : "Current Time:"}</span>
                  <span className="font-mono font-bold">{new Date().toTimeString().slice(0, 5)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{locale === "ar" ? "الموقع الجغرافي (Geofence):" : "Geofence Status:"}</span>
                  <span className="text-emerald-600 font-semibold">{locale === "ar" ? "داخل المقر الرئيسي (مطابق)" : "Verified On-Site"}</span>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowClockModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90"
                >
                  {clockType === "IN"
                    ? locale === "ar"
                      ? "تأكيد تسجيل الحضور"
                      : "Confirm Clock-In"
                    : locale === "ar"
                    ? "تأكيد تسجيل الانصراف"
                    : "Confirm Clock-Out"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: NEW SHIFT */}
      {showNewShiftModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-lg rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <Calendar className="h-5 w-5 text-primary" />
              {locale === "ar" ? "إضافة مناوبة عمل جديدة" : "Create New Work Shift"}
            </h3>
            <form onSubmit={handleCreateShift} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "اسم المناوبة (عربي)" : "Shift Name (Arabic)"}
                  </label>
                  <input
                    type="text"
                    required
                    value={newShiftData.nameAr}
                    onChange={(e) => setNewShiftData({ ...newShiftData, nameAr: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                    placeholder="وردية الفجر / الوردية الصباحية"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "اسم المناوبة (إنجليزي)" : "Shift Name (English)"}
                  </label>
                  <input
                    type="text"
                    required
                    value={newShiftData.nameEn}
                    onChange={(e) => setNewShiftData({ ...newShiftData, nameEn: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                    placeholder="Morning Shift A"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "وقت البداية" : "Start Time"}
                  </label>
                  <input
                    type="time"
                    required
                    value={newShiftData.startTime}
                    onChange={(e) => setNewShiftData({ ...newShiftData, startTime: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "وقت النهاية" : "End Time"}
                  </label>
                  <input
                    type="time"
                    required
                    value={newShiftData.endTime}
                    onChange={(e) => setNewShiftData({ ...newShiftData, endTime: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "فترة السماح (دقائق)" : "Grace Period (mins)"}
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="60"
                    value={newShiftData.gracePeriodMins}
                    onChange={(e) =>
                      setNewShiftData({ ...newShiftData, gracePeriodMins: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "مدة الراحة (دقائق)" : "Break Duration (mins)"}
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="120"
                    value={newShiftData.breakDurationMins}
                    onChange={(e) =>
                      setNewShiftData({ ...newShiftData, breakDurationMins: Number(e.target.value) })
                    }
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="ramadanCheck"
                  checked={newShiftData.isRamadan}
                  onChange={(e) => setNewShiftData({ ...newShiftData, isRamadan: e.target.checked })}
                  className="rounded border-input text-primary focus:ring-primary h-4 w-4"
                />
                <label htmlFor="ramadanCheck" className="text-xs font-medium cursor-pointer">
                  {locale === "ar"
                    ? "مناوبة خاصة بشهر رمضان المبارك (تحديد ساعات العمل بـ 6 ساعات يومياً)"
                    : "Ramadan Special Shift (capped at 6 hours/day under Saudi Labor Law)"}
                </label>
              </div>

              <div className="flex justify-end gap-2 pt-3">
                <button
                  type="button"
                  onClick={() => setShowNewShiftModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90"
                >
                  {locale === "ar" ? "حفظ المناوبة" : "Save Shift"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 3: REQUEST OVERTIME */}
      {showOTModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-md rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <Timer className="h-5 w-5 text-purple-600" />
              {locale === "ar" ? "طلب تكليف بعمل إضافي (المادة 107)" : "Request Overtime (Art. 107)"}
            </h3>
            <form onSubmit={handleRequestOT} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "الموظف المكلف" : "Employee"}
                </label>
                <select
                  value={newOTData.employeeId}
                  onChange={(e) => setNewOTData({ ...newOTData, employeeId: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                >
                  <option value="EMP-2025-001">محمد العتيبي (EMP-2025-001)</option>
                  <option value="EMP-2025-002">سارة الغامدي (EMP-2025-002)</option>
                  <option value="EMP-2025-003">طارق منصور (EMP-2025-003)</option>
                </select>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "التاريخ" : "Date"}
                  </label>
                  <input
                    type="date"
                    required
                    value={newOTData.workDate}
                    onChange={(e) => setNewOTData({ ...newOTData, workDate: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "عدد الساعات" : "Hours"}
                  </label>
                  <input
                    type="number"
                    step="0.5"
                    min="0.5"
                    max="12"
                    required
                    value={newOTData.hours}
                    onChange={(e) => setNewOTData({ ...newOTData, hours: Number(e.target.value) })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "سبب التكليف ومبرراته" : "Justification & Business Need"}
                </label>
                <textarea
                  required
                  rows={3}
                  value={newOTData.reason}
                  onChange={(e) => setNewOTData({ ...newOTData, reason: e.target.value })}
                  placeholder={
                    locale === "ar"
                      ? "إغلاق حسابات نهاية الشهر أو استكمال أعمال الصيانة الطارئة..."
                      : "Month-end closing, critical support..."
                  }
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                />
              </div>

              <div className="p-2.5 rounded-xl bg-purple-500/10 text-xs text-purple-700 dark:text-purple-300 font-medium">
                {locale === "ar"
                  ? "سيتم احتساب أجر الساعات بمعامل 150% (1.5x) آلياً وفق المادة 107"
                  : "Compensation automatically set to 150% (1.5x) per Saudi Labor Law Art. 107"}
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowOTModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium"
                >
                  {locale === "ar" ? "إرسال طلب التكليف" : "Submit Request"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 4: ATTENDANCE ADJUSTMENT */}
      {showAdjustModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-md rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <FileEdit className="h-5 w-5 text-primary" />
              {locale === "ar" ? "طلب تسوية حضور يدوية" : "Manual Attendance Adjustment"}
            </h3>
            <form onSubmit={handleAddAdjustment} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "الموظف" : "Employee"}
                </label>
                <input
                  type="text"
                  required
                  value={newAdjustData.employeeName}
                  onChange={(e) =>
                    setNewAdjustData({ ...newAdjustData, employeeName: e.target.value })
                  }
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "وقت الحضور المعدل" : "Adjusted Clock-In"}
                  </label>
                  <input
                    type="time"
                    required
                    value={newAdjustData.adjustedClockIn}
                    onChange={(e) =>
                      setNewAdjustData({ ...newAdjustData, adjustedClockIn: e.target.value })
                    }
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "وقت الانصراف المعدل" : "Adjusted Clock-Out"}
                  </label>
                  <input
                    type="time"
                    required
                    value={newAdjustData.adjustedClockOut}
                    onChange={(e) =>
                      setNewAdjustData({ ...newAdjustData, adjustedClockOut: e.target.value })
                    }
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "سبب التعديل الاستثنائي" : "Adjustment Reason"}
                </label>
                <textarea
                  required
                  rows={3}
                  value={newAdjustData.reason}
                  onChange={(e) => setNewAdjustData({ ...newAdjustData, reason: e.target.value })}
                  placeholder={
                    locale === "ar"
                      ? "عطل بجهاز البصمة أو مهمة عمل خارجية..."
                      : "Scanner hardware failure, off-site mission..."
                  }
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAdjustModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90"
                >
                  {locale === "ar" ? "اعتماد وحفظ التسوية" : "Save Adjustment"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

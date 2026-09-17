"use client";

import React, { useState } from "react";
import {
  Bell,
  CheckCheck,
  Filter,
  Search,
  Mail,
  Smartphone,
  AlertTriangle,
  CheckCircle2,
  Info,
  Clock,
  Sparkles,
  Send,
  Trash2,
  ExternalLink,
  ShieldAlert,
  Layers,
  FileCheck2,
  Receipt,
  Boxes,
  Banknote,
  Settings,
  X,
  Plus,
  RefreshCw,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

export interface NotificationItem {
  id: string;
  titleAr: string;
  titleEn: string;
  bodyAr: string;
  bodyEn: string;
  channel: "IN_APP" | "EMAIL" | "SMS";
  category: "WORKFLOW" | "ZATCA" | "FINANCE" | "INVENTORY" | "PAYROLL" | "SYSTEM";
  priority: "LOW" | "NORMAL" | "HIGH" | "URGENT";
  actionUrl?: string;
  status: "PENDING" | "SENT" | "FAILED" | "READ";
  createdAt: string;
  recipientEmail?: string;
}

const initialNotifications: NotificationItem[] = [
  {
    id: "notif-001",
    titleAr: "تم اعتماد الفاتورة الضريبية من هيئة الزكاة (ZATCA Cleared)",
    titleEn: "Tax Invoice Cleared by ZATCA Portal",
    bodyAr: "تم إصدار الختم الرقمي وتخليص الفاتورة رقم INV-2026-00042 بنجاح من منصة فاتورة (FATOORA).",
    bodyEn: "Cryptographic stamp generated and tax invoice #INV-2026-00042 cleared by ZATCA FATOORA portal.",
    channel: "IN_APP",
    category: "ZATCA",
    priority: "HIGH",
    actionUrl: "/zatca",
    status: "SENT",
    createdAt: "2026-03-17T09:30:00Z",
  },
  {
    id: "notif-002",
    titleAr: "طلب اعتماد أمر شراء جديد بانتظار موافقتكم",
    titleEn: "Purchase Order Approval Required",
    bodyAr: "قام قسم المشتريات برفع أمر الشراء رقم PO-2026-018 بمبلغ 145,000 ريال ويتطلب اعتماد المدير المالي.",
    bodyEn: "Procurement dept submitted PO-2026-018 for SAR 145,000 requiring CFO financial approval.",
    channel: "EMAIL",
    category: "WORKFLOW",
    priority: "URGENT",
    actionUrl: "/purchasing",
    status: "SENT",
    recipientEmail: "cfo@saudi-enterprise.sa",
    createdAt: "2026-03-17T08:15:00Z",
  },
  {
    id: "notif-003",
    titleAr: "تحذير: وصول مادة أولية إلى حد إعادة الطلب",
    titleEn: "Warning: Stock Level Below Reorder Point",
    bodyAr: "وصل رصيد صنف 'حبيبات البوليمر الصناعي' في المستودع المركزي إلى 250 كجم، وهو أقل من حد الأمان.",
    bodyEn: "Polymer Granules stock in Central Warehouse dropped to 250 KG, below safety threshold.",
    channel: "IN_APP",
    category: "INVENTORY",
    priority: "HIGH",
    actionUrl: "/inventory",
    status: "SENT",
    createdAt: "2026-03-16T16:45:00Z",
  },
  {
    id: "notif-004",
    titleAr: "تم احتساب مسير رواتب شهر مارس 2026 وجاهزية ملف WPS",
    titleEn: "March 2026 Payroll Processed & WPS Ready",
    bodyAr: "اكتمل احتساب مسير رواتب 48 موظفاً وتوليد ملف حماية الأجور ومطابقة استقطاعات التأمينات الاجتماعية GOSI.",
    bodyEn: "Monthly payroll for 48 employees processed with WPS file generation and GOSI deductions validated.",
    channel: "IN_APP",
    category: "PAYROLL",
    priority: "NORMAL",
    actionUrl: "/payroll",
    status: "READ",
    createdAt: "2026-03-16T11:00:00Z",
  },
  {
    id: "notif-005",
    titleAr: "إشعار تجاوز سقف الميزانية التشغيلية (OPEX)",
    titleEn: "OPEX Budget Threshold Advisory",
    bodyAr: "بلغت مصاريف تقنية المعلومات والاتصالات نسبة 92% من الميزانية المعتمدة للربع الأول 2026.",
    bodyEn: "IT & Telecommunications expenditure has reached 92% of Q1 2026 approved budget allocation.",
    channel: "EMAIL",
    category: "FINANCE",
    priority: "HIGH",
    actionUrl: "/budgets",
    status: "SENT",
    recipientEmail: "finance.controller@saudi-enterprise.sa",
    createdAt: "2026-03-15T14:20:00Z",
  },
  {
    id: "notif-006",
    titleAr: "تحديث أمني: جاهزية مفاتيح التشفير ECDSA secp256k1",
    titleEn: "Security Notice: Cryptographic Keys Generated",
    bodyAr: "تم تجديد أزواج المفاتيح وشهادات CSID لأجهزة الـ EGS في فرع الرياض بنجاح.",
    bodyEn: "Key pairs and CSID certificates successfully renewed for Riyadh branch EGS solution units.",
    channel: "IN_APP",
    category: "SYSTEM",
    priority: "LOW",
    actionUrl: "/zatca",
    status: "READ",
    createdAt: "2026-03-14T10:00:00Z",
  },
];

export function NotificationsView() {
  const { locale } = useLocale();
  const isAr = locale === "ar";

  const [notifications, setNotifications] = useState<NotificationItem[]>(initialNotifications);
  const [filterType, setFilterType] = useState<"ALL" | "UNREAD">("ALL");
  const [selectedCategory, setSelectedCategory] = useState<string>("ALL");
  const [searchQuery, setSearchQuery] = useState("");
  const [isSendOpen, setIsSendOpen] = useState(false);

  // Form states
  const [newTitleAr, setNewTitleAr] = useState("");
  const [newTitleEn, setNewTitleEn] = useState("");
  const [newBodyAr, setNewBodyAr] = useState("");
  const [newBodyEn, setNewBodyEn] = useState("");
  const [newCategory, setNewCategory] = useState<NotificationItem["category"]>("SYSTEM");
  const [newPriority, setNewPriority] = useState<NotificationItem["priority"]>("NORMAL");
  const [newChannel, setNewChannel] = useState<NotificationItem["channel"]>("IN_APP");
  const [newRecipientEmail, setNewRecipientEmail] = useState("");
  const [newActionUrl, setNewActionUrl] = useState("");
  const [isSending, setIsSending] = useState(false);

  // Counts
  const totalCount = notifications.length;
  const unreadCount = notifications.filter((n) => n.status !== "READ").length;
  const emailCount = notifications.filter((n) => n.channel === "EMAIL").length;
  const workflowCount = notifications.filter((n) => n.category === "WORKFLOW").length;

  const handleMarkAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, status: "READ" } : n))
    );
  };

  const handleMarkAllAsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, status: "READ" })));
  };

  const handleDelete = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const handleSendSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitleAr.trim() && !newTitleEn.trim()) return;

    setIsSending(true);
    setTimeout(() => {
      const newItem: NotificationItem = {
        id: `notif-${Date.now()}`,
        titleAr: newTitleAr || newTitleEn,
        titleEn: newTitleEn || newTitleAr,
        bodyAr: newBodyAr || newBodyEn,
        bodyEn: newBodyEn || newBodyAr,
        category: newCategory,
        priority: newPriority,
        channel: newChannel,
        recipientEmail: newRecipientEmail,
        actionUrl: newActionUrl || "/dashboard",
        status: "SENT",
        createdAt: new Date().toISOString(),
      };

      setNotifications([newItem, ...notifications]);
      setIsSending(false);
      setIsSendOpen(false);

      // Reset form
      setNewTitleAr("");
      setNewTitleEn("");
      setNewBodyAr("");
      setNewBodyEn("");
      setNewRecipientEmail("");
      setNewActionUrl("");
    }, 400);
  };

  // Filter logic
  const filteredNotifications = notifications.filter((n) => {
    if (filterType === "UNREAD" && n.status === "READ") return false;
    if (selectedCategory !== "ALL" && n.category !== selectedCategory) return false;

    if (searchQuery.trim() !== "") {
      const q = searchQuery.toLowerCase();
      const matchTitle = n.titleAr.toLowerCase().includes(q) || n.titleEn.toLowerCase().includes(q);
      const matchBody = n.bodyAr.toLowerCase().includes(q) || n.bodyEn.toLowerCase().includes(q);
      return matchTitle || matchBody;
    }
    return true;
  });

  const getCategoryIcon = (category: NotificationItem["category"]) => {
    switch (category) {
      case "ZATCA":
        return <FileCheck2 className="h-4 w-4 text-emerald-600" />;
      case "WORKFLOW":
        return <Layers className="h-4 w-4 text-blue-600" />;
      case "FINANCE":
        return <Receipt className="h-4 w-4 text-amber-600" />;
      case "INVENTORY":
        return <Boxes className="h-4 w-4 text-orange-600" />;
      case "PAYROLL":
        return <Banknote className="h-4 w-4 text-purple-600" />;
      default:
        return <Settings className="h-4 w-4 text-slate-500" />;
    }
  };

  const getPriorityBadge = (priority: NotificationItem["priority"]) => {
    switch (priority) {
      case "URGENT":
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-bold bg-destructive text-destructive-foreground animate-pulse">
            {isAr ? "عاجل جداً" : "Urgent"}
          </span>
        );
      case "HIGH":
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300 border border-rose-200 dark:border-rose-800">
            {isAr ? "أولوية عالية" : "High"}
          </span>
        );
      case "NORMAL":
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border border-blue-200 dark:border-blue-800">
            {isAr ? "عادي" : "Normal"}
          </span>
        );
      case "LOW":
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-muted text-muted-foreground">
            {isAr ? "منخفض" : "Low"}
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* 1. Header & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight">
              {isAr ? "مركز التنبيهات والإشعارات الذكية" : "Notification & Messaging Center"}
            </h1>
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              Phase 28
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            {isAr
              ? "إدارة ومتابعة تنبيهات مسارات العمل، الامتثال لزاتكا، التقارير المالية، والتراسل البريدي الفوري"
              : "Enterprise messaging engine for workflow alerts, ZATCA clearance notices, finance, and email dispatch"}
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            onClick={handleMarkAllAsRead}
            disabled={unreadCount === 0}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg border bg-background hover:bg-muted text-xs font-medium transition disabled:opacity-50"
          >
            <CheckCheck className="h-4 w-4 text-emerald-600" />
            {isAr ? "تحديد الكل كمقروء" : "Mark All as Read"}
          </button>

          <button
            onClick={() => setIsSendOpen(true)}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-xs font-semibold hover:bg-primary/90 transition shadow-sm"
          >
            <Send className="h-3.5 w-3.5" />
            {isAr ? "إرسال تنبيه جديد" : "Dispatch Alert"}
          </button>
        </div>
      </div>

      {/* 2. KPI Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "إجمالي التنبيهات" : "Total Alerts"}
            </span>
            <div className="p-2 rounded-lg bg-primary/10 text-primary">
              <Bell className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{totalCount}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">{isAr ? "إشعار مسجل" : "notifications"}</span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "التنبيهات غير المقروءة" : "Unread Alerts"}
            </span>
            <div className="p-2 rounded-lg bg-rose-500/10 text-rose-600 dark:text-rose-400">
              <AlertTriangle className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold text-rose-600 dark:text-rose-400">{unreadCount}</span>
            <span className="text-xs text-muted-foreground">{isAr ? "تتطلب الإجراء" : "pending read"}</span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "رسائل البريد المرسلة" : "Dispatched Emails"}
            </span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <Mail className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{emailCount}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">{isAr ? "عبر Mailer" : "via mail platform"}</span>
          </div>
        </div>

        <div className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {isAr ? "طلبات اعتماد مسارات العمل" : "Workflow Actions"}
            </span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <Layers className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-bold">{workflowCount}</span>
            <span className="text-xs text-muted-foreground ml-2 mr-2">{isAr ? "طلب اعتماد" : "approval alerts"}</span>
          </div>
        </div>
      </div>

      {/* 3. Filter Bar */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 bg-card p-3 rounded-xl border">
        {/* Status Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0">
          <button
            onClick={() => setFilterType("ALL")}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition whitespace-nowrap ${
              filterType === "ALL"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            {isAr ? "كافة الإشعارات" : "All Alerts"} ({totalCount})
          </button>
          <button
            onClick={() => setFilterType("UNREAD")}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition whitespace-nowrap ${
              filterType === "UNREAD"
                ? "bg-primary text-primary-foreground shadow-sm"
                : "bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            {isAr ? "غير المقروءة فقط" : "Unread Only"} ({unreadCount})
          </button>

          <div className="h-4 w-px bg-border mx-1" />

          {/* Category Dropdown */}
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="px-2.5 py-1.5 text-xs rounded-lg border bg-background text-foreground"
          >
            <option value="ALL">{isAr ? "كافة التصنيفات" : "All Categories"}</option>
            <option value="WORKFLOW">{isAr ? "مسارات العمل" : "Workflow"}</option>
            <option value="ZATCA">{isAr ? "الفوترة وزاتكا" : "ZATCA"}</option>
            <option value="FINANCE">{isAr ? "المالية والميزانيات" : "Finance"}</option>
            <option value="INVENTORY">{isAr ? "المخزون" : "Inventory"}</option>
            <option value="PAYROLL">{isAr ? "الرواتب والأجور" : "Payroll"}</option>
            <option value="SYSTEM">{isAr ? "النظام" : "System"}</option>
          </select>
        </div>

        {/* Search Input */}
        <div className="relative w-full md:w-64">
          <Search className="absolute right-3 top-2.5 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
          <input
            type="text"
            placeholder={isAr ? "بحث في التنبيهات..." : "Search alerts..."}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pr-8 pl-3 py-1.5 text-xs rounded-lg border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          />
        </div>
      </div>

      {/* 4. Notifications Feed List */}
      <div className="space-y-3">
        {filteredNotifications.length === 0 ? (
          <div className="rounded-xl border bg-card p-12 text-center">
            <CheckCircle2 className="h-10 w-10 text-muted-foreground mx-auto mb-3 opacity-50" />
            <h3 className="font-semibold text-sm">
              {isAr ? "لا توجد تنبيهات جديدة في هذه القائمة" : "No notifications found"}
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "كافة العمليات والموافقات مكتملة بنجاح" : "You're all caught up with your tasks and alerts"}
            </p>
          </div>
        ) : (
          filteredNotifications.map((n) => {
            const isRead = n.status === "READ";
            return (
              <div
                key={n.id}
                className={`rounded-xl border p-4 transition shadow-sm flex flex-col sm:flex-row sm:items-start justify-between gap-4 ${
                  isRead
                    ? "bg-card/60 opacity-85 hover:opacity-100"
                    : "bg-card border-primary/30 shadow-primary/5 ring-1 ring-primary/10"
                }`}
              >
                <div className="flex items-start gap-3.5">
                  <div className="p-2.5 rounded-xl bg-muted/80 mt-0.5 shrink-0">
                    {getCategoryIcon(n.category)}
                  </div>

                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className={`font-bold text-sm ${isRead ? "text-foreground/80" : "text-foreground"}`}>
                        {isAr ? n.titleAr : n.titleEn}
                      </span>
                      {getPriorityBadge(n.priority)}

                      {n.channel === "EMAIL" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium bg-blue-100/60 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
                          <Mail className="h-3 w-3" />
                          {isAr ? "بريد إلكتروني" : "Email"}
                        </span>
                      )}

                      {!isRead && (
                        <span className="h-2 w-2 rounded-full bg-primary animate-pulse" />
                      )}
                    </div>

                    <p className="text-xs text-muted-foreground leading-relaxed">
                      {isAr ? n.bodyAr : n.bodyEn}
                    </p>

                    <div className="flex items-center gap-4 text-[11px] text-muted-foreground pt-1">
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {new Date(n.createdAt).toLocaleDateString(isAr ? "ar-SA" : "en-US", {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </span>

                      {n.recipientEmail && (
                        <span className="font-mono text-[10px] bg-muted px-1.5 py-0.5 rounded">
                          {n.recipientEmail}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Card Actions */}
                <div className="flex items-center gap-2 self-end sm:self-center shrink-0">
                  {n.actionUrl && (
                    <a
                      href={n.actionUrl}
                      className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-primary/10 text-primary text-xs font-semibold hover:bg-primary/20 transition"
                    >
                      <span>{isAr ? "عرض التفاصيل" : "View Details"}</span>
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  )}

                  {!isRead && (
                    <button
                      onClick={() => handleMarkAsRead(n.id)}
                      className="p-1.5 rounded-lg border hover:bg-muted text-muted-foreground hover:text-foreground transition"
                      title={isAr ? "تحديد كمقروء" : "Mark as Read"}
                    >
                      <CheckCheck className="h-4 w-4" />
                    </button>
                  )}

                  <button
                    onClick={() => handleDelete(n.id)}
                    className="p-1.5 rounded-lg hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition"
                    title={isAr ? "حذف التنبيه" : "Delete"}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* 5. Modal: Send Notification */}
      {isSendOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <div className="bg-card border rounded-2xl w-full max-w-lg shadow-xl overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div className="flex items-center gap-2">
                <Send className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-base">
                  {isAr ? "إرسال تنبيه أو بريد جديد" : "Dispatch New Alert or Email"}
                </h3>
              </div>
              <button
                onClick={() => setIsSendOpen(false)}
                className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleSendSubmit} className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1">
                    {isAr ? "قناة الإرسال *" : "Channel *"}
                  </label>
                  <select
                    value={newChannel}
                    onChange={(e) => setNewChannel(e.target.value as any)}
                    className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground"
                  >
                    <option value="IN_APP">{isAr ? "إشعار في النظام (In-App)" : "In-App Notification"}</option>
                    <option value="EMAIL">{isAr ? "بريد إلكتروني (Email)" : "Email Message"}</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1">
                    {isAr ? "الأولوية *" : "Priority *"}
                  </label>
                  <select
                    value={newPriority}
                    onChange={(e) => setNewPriority(e.target.value as any)}
                    className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground"
                  >
                    <option value="LOW">{isAr ? "منخفضة" : "Low"}</option>
                    <option value="NORMAL">{isAr ? "عادية" : "Normal"}</option>
                    <option value="HIGH">{isAr ? "عالية" : "High"}</option>
                    <option value="URGENT">{isAr ? "عاجلة جداً" : "Urgent"}</option>
                  </select>
                </div>
              </div>

              {newChannel === "EMAIL" && (
                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1">
                    {isAr ? "البريد الإلكتروني للمستلم *" : "Recipient Email *"}
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="recipient@company.sa"
                    value={newRecipientEmail}
                    onChange={(e) => setNewRecipientEmail(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground font-mono"
                  />
                </div>
              )}

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1">
                    {isAr ? "التصنيف *" : "Category *"}
                  </label>
                  <select
                    value={newCategory}
                    onChange={(e) => setNewCategory(e.target.value as any)}
                    className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground"
                  >
                    <option value="WORKFLOW">{isAr ? "مسارات العمل" : "Workflow"}</option>
                    <option value="ZATCA">{isAr ? "الفوترة وزاتكا" : "ZATCA"}</option>
                    <option value="FINANCE">{isAr ? "المالية" : "Finance"}</option>
                    <option value="INVENTORY">{isAr ? "المخزون" : "Inventory"}</option>
                    <option value="PAYROLL">{isAr ? "الرواتب" : "Payroll"}</option>
                    <option value="SYSTEM">{isAr ? "النظام" : "System"}</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-semibold text-foreground block mb-1">
                    {isAr ? "رابط الإجراء (Action URL)" : "Action URL"}
                  </label>
                  <input
                    type="text"
                    placeholder="/sales or /workflow"
                    value={newActionUrl}
                    onChange={(e) => setNewActionUrl(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground font-mono"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1">
                  {isAr ? "العنوان بالعربية *" : "Title (Arabic) *"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="مثال: طلب اعتماد مستند جديد"
                  value={newTitleAr}
                  onChange={(e) => setNewTitleAr(e.target.value)}
                  className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground"
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1">
                  {isAr ? "العنوان بالإنجليزية" : "Title (English)"}
                </label>
                <input
                  type="text"
                  placeholder="e.g. New Document Approval Request"
                  value={newTitleEn}
                  onChange={(e) => setNewTitleEn(e.target.value)}
                  className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground"
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground block mb-1">
                  {isAr ? "نص الإشعار بالعربية *" : "Body Text (Arabic) *"}
                </label>
                <textarea
                  rows={2}
                  required
                  placeholder="محتوى الإشعار وتفاصيل العملية..."
                  value={newBodyAr}
                  onChange={(e) => setNewBodyAr(e.target.value)}
                  className="w-full px-3 py-2 text-xs rounded-lg border bg-background text-foreground resize-none"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsSendOpen(false)}
                  className="px-4 py-2 text-xs font-medium rounded-lg border hover:bg-muted transition"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  disabled={isSending}
                  className="px-5 py-2 text-xs font-semibold rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition shadow-sm flex items-center gap-1.5"
                >
                  {isSending ? (
                    <>
                      <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                      {isAr ? "جاري الإرسال..." : "Sending..."}
                    </>
                  ) : (
                    <>
                      <Send className="h-3.5 w-3.5" />
                      {isAr ? "إرسال التنبيه" : "Dispatch"}
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

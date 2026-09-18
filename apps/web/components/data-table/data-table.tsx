"use client";

import React, { useState, useMemo } from "react";
import {
  ChevronLeft,
  ChevronRight,
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Inbox,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export interface ColumnDef<T> {
  key: string;
  headerAr: string;
  headerEn: string;
  render?: (row: T) => React.ReactNode;
  sortable?: boolean;
  align?: "start" | "center" | "end";
}

interface DataTableProps<T> {
  data: T[];
  columns: ColumnDef<T>[];
  searchPlaceholderAr?: string;
  searchPlaceholderEn?: string;
  searchKey?: keyof T;
  pageSize?: number;
  isArabic?: boolean;
  emptyTitleAr?: string;
  emptyTitleEn?: string;
  className?: string;
  actionsSlot?: React.ReactNode;
}

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  searchPlaceholderAr = "بحث في السجلات...",
  searchPlaceholderEn = "Search records...",
  searchKey,
  pageSize = 10,
  isArabic = true,
  emptyTitleAr = "لا توجد سجلات مطابقة للبحث",
  emptyTitleEn = "No matching records found",
  className,
  actionsSlot,
}: DataTableProps<T>) {
  const [query, setQuery] = useState("");
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [currentPage, setCurrentPage] = useState(1);

  // 1. Filter
  const filteredData = useMemo(() => {
    if (!query.trim()) return data;
    const lower = query.toLowerCase();

    return data.filter((item) => {
      if (searchKey) {
        const val = item[searchKey];
        return String(val || "").toLowerCase().includes(lower);
      }
      return Object.values(item).some((val) =>
        String(val || "").toLowerCase().includes(lower)
      );
    });
  }, [data, query, searchKey]);

  // 2. Sort
  const sortedData = useMemo(() => {
    if (!sortKey) return filteredData;

    return [...filteredData].sort((a, b) => {
      const valA = a[sortKey];
      const valB = b[sortKey];

      if (valA === valB) return 0;
      if (valA === null || valA === undefined) return 1;
      if (valB === null || valB === undefined) return -1;

      if (typeof valA === "number" && typeof valB === "number") {
        return sortDirection === "asc" ? valA - valB : valB - valA;
      }

      return sortDirection === "asc"
        ? String(valA).localeCompare(String(valB))
        : String(valB).localeCompare(String(valA));
    });
  }, [filteredData, sortKey, sortDirection]);

  // 3. Paginate
  const totalPages = Math.ceil(sortedData.length / pageSize) || 1;
  const paginatedData = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return sortedData.slice(start, start + pageSize);
  }, [sortedData, currentPage, pageSize]);

  const handleSort = (key: string) => {
    if (sortKey === key) {
      if (sortDirection === "asc") {
        setSortDirection("desc");
      } else {
        setSortKey(null);
      }
    } else {
      setSortKey(key);
      setSortDirection("asc");
    }
  };

  return (
    <div className={cn("space-y-4", className)}>
      {/* Top Controls: Search + Actions slot */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        <div className="w-full sm:max-w-xs">
          <Input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setCurrentPage(1);
            }}
            placeholder={isArabic ? searchPlaceholderAr : searchPlaceholderEn}
            startIcon={<Search className="h-4 w-4" />}
          />
        </div>
        {actionsSlot && <div className="flex items-center gap-2 self-end">{actionsSlot}</div>}
      </div>

      {/* Table Box */}
      <div className="rounded-xl border bg-card overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-xs text-start">
            <thead className="bg-muted/50 border-b text-muted-foreground font-semibold">
              <tr>
                {columns.map((col) => {
                  const isSorted = sortKey === col.key;
                  const alignClass =
                    col.align === "center"
                      ? "text-center"
                      : col.align === "end"
                      ? "text-end"
                      : "text-start";

                  return (
                    <th
                      key={col.key}
                      className={cn("px-4 py-3 whitespace-nowrap", alignClass)}
                    >
                      {col.sortable ? (
                        <button
                          type="button"
                          onClick={() => handleSort(col.key)}
                          className="inline-flex items-center gap-1.5 hover:text-foreground transition-colors font-semibold"
                        >
                          <span>{isArabic ? col.headerAr : col.headerEn}</span>
                          {isSorted ? (
                            sortDirection === "asc" ? (
                              <ArrowUp className="h-3.5 w-3.5 text-primary" />
                            ) : (
                              <ArrowDown className="h-3.5 w-3.5 text-primary" />
                            )
                          ) : (
                            <ArrowUpDown className="h-3 w-3 opacity-40" />
                          )}
                        </button>
                      ) : (
                        <span>{isArabic ? col.headerAr : col.headerEn}</span>
                      )}
                    </th>
                  );
                })}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {paginatedData.length === 0 ? (
                <tr>
                  <td colSpan={columns.length} className="py-12 text-center text-muted-foreground">
                    <Inbox className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <span>{isArabic ? emptyTitleAr : emptyTitleEn}</span>
                  </td>
                </tr>
              ) : (
                paginatedData.map((row, idx) => (
                  <tr
                    key={row.id || idx}
                    className="hover:bg-muted/40 transition-colors"
                  >
                    {columns.map((col) => {
                      const alignClass =
                        col.align === "center"
                          ? "text-center"
                          : col.align === "end"
                          ? "text-end"
                          : "text-start";

                      return (
                        <td
                          key={col.key}
                          className={cn("px-4 py-3 text-foreground", alignClass)}
                        >
                          {col.render ? col.render(row) : row[col.key]}
                        </td>
                      );
                    })}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t bg-muted/20 text-xs text-muted-foreground">
          <div>
            {isArabic
              ? `إجمالي السجلات: ${sortedData.length} (الصفحة ${currentPage} من ${totalPages})`
              : `Total records: ${sortedData.length} (Page ${currentPage} of ${totalPages})`}
          </div>
          <div className="flex items-center gap-1.5">
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage <= 1}
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              className="gap-1 h-7 px-2 text-xs"
            >
              <ChevronLeft className="h-3.5 w-3.5 rtl:rotate-180" />
              <span>{isArabic ? "السابق" : "Prev"}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage >= totalPages}
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              className="gap-1 h-7 px-2 text-xs"
            >
              <span>{isArabic ? "التالي" : "Next"}</span>
              <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

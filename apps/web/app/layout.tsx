import type { Metadata } from "next";
import { Inter, Cairo } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { ShellProvider, Shell } from "@/components/layout/shell";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const cairo = Cairo({
  subsets: ["arabic", "latin"],
  variable: "--font-cairo",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Saudi ERP Platform | منصة تخطيط الموارد للمنشآت السعودية",
  description:
    "Commercial Production-Grade ERP Platform for Saudi Arabia, fully compliant with ZATCA Phase 2 E-Invoicing and GOSI rules.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ar" dir="rtl" suppressHydrationWarning className={`${inter.variable} ${cairo.variable}`}>
      <body className="antialiased min-h-screen font-sans">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <ShellProvider>
            <Shell>{children}</Shell>
          </ShellProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}

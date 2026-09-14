import React from "react";
import { LoginForm } from "@/features/auth/login-form";

export const metadata = {
  title: "تسجيل الدخول | منصة تخطيط الموارد للمنشآت السعودية",
  description: "تسجيل الدخول الآمن لنظام تخطيط الموارد ERP",
};

export default function LoginPage() {
  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4 py-12">
      <LoginForm />
    </div>
  );
}

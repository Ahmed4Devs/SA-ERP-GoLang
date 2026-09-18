"use client";

import { useCallback, useEffect, useState } from "react";
import { AlertCircle, CheckCircle2, Copy, KeyRound, ShieldCheck } from "lucide-react";
import { apiFetch } from "@/lib/api-client";
import { useLocale } from "@/components/layout/shell";

interface MFAEnrollment {
  secret: string;
  otpauth_uri: string;
  recovery_codes: string[];
}

export function MFASettings() {
  const { locale, dir } = useLocale();
  const isAr = locale === "ar";
  const [enabled, setEnabled] = useState<boolean | null>(null);
  const [enrollment, setEnrollment] = useState<MFAEnrollment | null>(null);
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [copying, setCopying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const loadStatus = useCallback(async () => {
    try {
      const status = await apiFetch<{ enabled: boolean }>("/api/v1/auth/mfa/status");
      setEnabled(status.enabled);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to load MFA status");
    }
  }, []);

  useEffect(() => {
    void loadStatus();
  }, [loadStatus]);

  const enroll = async () => {
    setLoading(true);
    setError(null);
    setMessage(null);
    try {
      const result = await apiFetch<MFAEnrollment>("/api/v1/auth/mfa/enroll", { method: "POST" });
      setEnrollment(result);
      setCode("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to start MFA enrollment");
    } finally {
      setLoading(false);
    }
  };

  const confirm = async () => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch("/api/v1/auth/mfa/confirm", {
        method: "POST",
        body: JSON.stringify({ code }),
      });
      setEnabled(true);
      setCode("");
      setMessage(isAr ? "تم تفعيل المصادقة متعددة العوامل." : "Multi-factor authentication is enabled.");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to confirm MFA");
    } finally {
      setLoading(false);
    }
  };

  const disable = async () => {
    setLoading(true);
    setError(null);
    setMessage(null);
    try {
      await apiFetch("/api/v1/auth/mfa/disable", {
        method: "POST",
        body: JSON.stringify({ code }),
      });
      setEnabled(false);
      setEnrollment(null);
      setCode("");
      setMessage(isAr ? "تم تعطيل المصادقة متعددة العوامل." : "Multi-factor authentication is disabled.");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to disable MFA");
    } finally {
      setLoading(false);
    }
  };

  const copyRecoveryCodes = async () => {
    if (!enrollment) return;
    setCopying(true);
    setError(null);
    try {
      await navigator.clipboard.writeText(enrollment.recovery_codes.join("\n"));
      setMessage(isAr ? "تم نسخ رموز الاسترداد." : "Recovery codes copied.");
    } catch {
      setError(isAr ? "تعذر نسخ الرموز. انسخها يدويًا قبل المتابعة." : "Could not copy the codes. Copy them manually before continuing.");
    } finally {
      setCopying(false);
    }
  };

  return (
    <section className="rounded-2xl border bg-card p-6 shadow-sm" dir={dir} aria-labelledby="mfa-heading">
      <div className="flex items-start gap-3">
        <div className="rounded-xl bg-primary/10 p-2 text-primary"><ShieldCheck className="h-5 w-5" /></div>
        <div className="flex-1">
          <h2 id="mfa-heading" className="text-lg font-bold">
            {isAr ? "المصادقة متعددة العوامل" : "Multi-factor authentication"}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {isAr
              ? "احمِ الحساب برمز مؤقت من تطبيق المصادقة، مع رموز استرداد تستخدم مرة واحدة."
              : "Protect the account with time-based authenticator codes and one-time recovery codes."}
          </p>
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-semibold ${enabled ? "bg-emerald-500/10 text-emerald-600" : "bg-muted text-muted-foreground"}`}>
          {enabled === null ? "…" : enabled ? (isAr ? "مفعّلة" : "Enabled") : (isAr ? "غير مفعّلة" : "Disabled")}
        </span>
      </div>

      {error && <div role="alert" className="mt-4 flex gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"><AlertCircle className="h-4 w-4 shrink-0" />{error}</div>}
      {message && <div role="status" className="mt-4 flex gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm text-emerald-600"><CheckCircle2 className="h-4 w-4 shrink-0" />{message}</div>}

      {!enabled && !enrollment && (
        <button type="button" onClick={() => void enroll()} disabled={loading || enabled === null} className="mt-5 inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50">
          <KeyRound className="h-4 w-4" />{isAr ? "بدء الإعداد" : "Start setup"}
        </button>
      )}

      {enrollment && (
        <div className="mt-5 space-y-4">
          {!enabled && (
            <div className="rounded-lg border bg-muted/30 p-4 text-sm">
              <p className="font-semibold">{isAr ? "أضف المفتاح إلى تطبيق المصادقة:" : "Add this key to your authenticator app:"}</p>
              <code className="mt-2 block break-all rounded bg-background p-2 font-mono text-xs" data-testid="mfa-secret">{enrollment.secret}</code>
              <details className="mt-2 text-xs text-muted-foreground"><summary>{isAr ? "رابط الإعداد المتقدم" : "Advanced setup URI"}</summary><code className="mt-2 block break-all">{enrollment.otpauth_uri}</code></details>
            </div>
          )}
          <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-4">
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-semibold">{isAr ? "احفظ رموز الاسترداد الآن" : "Save recovery codes now"}</p>
              <button type="button" onClick={() => void copyRecoveryCodes()} disabled={copying} className="inline-flex items-center gap-1 text-xs text-primary disabled:opacity-50"><Copy className="h-3.5 w-3.5" />{isAr ? "نسخ" : "Copy"}</button>
            </div>
            <div className="mt-3 grid grid-cols-2 gap-2 font-mono text-xs" data-testid="recovery-codes">
              {enrollment.recovery_codes.map((recoveryCode) => <code key={recoveryCode}>{recoveryCode}</code>)}
            </div>
          </div>
          {!enabled ? (
            <>
              <CodeInput value={code} onChange={setCode} isAr={isAr} />
              <button type="button" onClick={() => void confirm()} disabled={loading || !code.trim()} className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50">
                {isAr ? "تأكيد وتفعيل" : "Confirm and enable"}
              </button>
            </>
          ) : (
            <button type="button" onClick={() => setEnrollment(null)} className="rounded-lg border px-4 py-2 text-sm font-semibold">
              {isAr ? "حفظت رموز الاسترداد" : "I saved the recovery codes"}
            </button>
          )}
        </div>
      )}

      {enabled && !enrollment && (
        <div className="mt-5 space-y-3">
          <CodeInput value={code} onChange={setCode} isAr={isAr} />
          <button type="button" onClick={() => void disable()} disabled={loading || !code.trim()} className="rounded-lg border border-destructive px-4 py-2 text-sm font-semibold text-destructive disabled:opacity-50">
            {isAr ? "تعطيل MFA" : "Disable MFA"}
          </button>
        </div>
      )}
    </section>
  );
}

function CodeInput({ value, onChange, isAr }: { value: string; onChange: (value: string) => void; isAr: boolean }) {
  return (
    <div className="space-y-1.5">
      <label htmlFor="mfa-settings-code" className="block text-xs font-semibold">{isAr ? "رمز المصادقة أو الاسترداد" : "Authenticator or recovery code"}</label>
      <input id="mfa-settings-code" value={value} onChange={(event) => onChange(event.target.value)} autoComplete="one-time-code" autoCapitalize="characters" spellCheck={false} className="w-full max-w-sm rounded-md border bg-background px-3 py-2 text-sm tracking-wider outline-none focus:ring-2 focus:ring-primary/20" />
    </div>
  );
}

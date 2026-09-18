"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useLocale } from "@/components/layout/shell";
import { Lock, User, Eye, EyeOff, AlertCircle, CheckCircle2, Shield } from "lucide-react";
import { apiFetch, APIError } from "@/lib/api-client";
import { useAuth } from "@/lib/auth-context";
import { getMFAChallenge, getSafeReturnPath } from "@/lib/auth-flow";

export function LoginForm() {
  const { t, dir } = useLocale();
  const router = useRouter();
  const { user, loading: sessionLoading, refresh } = useAuth();

  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [challengeToken, setChallengeToken] = useState<string | null>(null);
  const [challengeExpiresAt, setChallengeExpiresAt] = useState<string | null>(null);
  const [mfaCode, setMFACode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!sessionLoading && user && !success) {
      router.replace(getSafeReturnPath(window.location.search));
    }
  }, [router, sessionLoading, success, user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      if (!challengeToken) {
        const response = await apiFetch<unknown>("/api/v1/auth/login", {
          method: "POST",
          body: JSON.stringify({ identifier: identifier.trim(), password }),
        });
        const challenge = getMFAChallenge(response);
        if (challenge) {
          setChallengeToken(challenge.challengeToken);
          setChallengeExpiresAt(challenge.expiresAt);
          setMFACode("");
          return;
        }
      } else {
        await apiFetch("/api/v1/auth/login/mfa", {
          method: "POST",
          body: JSON.stringify({ challenge_token: challengeToken, code: mfaCode.trim() }),
        });
      }

      await refresh();
      setSuccess(true);
      router.replace(getSafeReturnPath(window.location.search));
    } catch (err: unknown) {
      if (err instanceof APIError && err.code === "INVALID_CREDENTIALS") {
        setError(t.auth.invalid_credentials);
      } else if (err instanceof APIError && err.code === "INVALID_MFA_CODE") {
        setError("رمز التحقق غير صحيح أو منتهي الصلاحية / Invalid or expired verification code");
      } else {
        setError(err instanceof Error ? err.message : "Failed to connect to authentication server");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto" dir={dir}>
      <div className="rounded-2xl border bg-card p-6 sm:p-8 shadow-lg space-y-6">
        {/* Header */}
        <div className="text-center space-y-2">
          <div className="inline-flex h-12 w-12 rounded-xl bg-primary text-primary-foreground items-center justify-center shadow-md mb-2">
            <Shield className="h-6 w-6" />
          </div>
          <h2 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
            {t.auth.login_title}
          </h2>
          <p className="text-xs sm:text-sm text-muted-foreground">
            {t.auth.login_subtitle}
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="p-3 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-xs flex items-center gap-2">
            <AlertCircle className="h-4 w-4 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Success Alert */}
        {success && (
          <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
            <span>تم التحقق وتسجيل الدخول بنجاح / Successfully Authenticated</span>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {!challengeToken ? (
            <>
              {/* Identifier field */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground block">
              {t.auth.identifier}
            </label>
            <div className="relative">
              <User className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                id="login-identifier"
                type="text"
                autoComplete="username"
                required
                value={identifier}
                onChange={(e) => setIdentifier(e.target.value)}
                placeholder={t.auth.identifier_placeholder}
                className="w-full rounded-md border bg-background ps-9 pe-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
              />
            </div>
          </div>

          {/* Password field */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground block">
              {t.auth.password}
            </label>
            <div className="relative">
              <Lock className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                id="login-password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t.auth.password_placeholder}
                className="w-full rounded-md border bg-background ps-9 pe-10 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-inline-end-3 top-2.5 text-muted-foreground hover:text-foreground"
                aria-label="Toggle password visibility"
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>
            </>
          ) : (
            <div className="space-y-3">
              <div className="rounded-lg border bg-muted/40 p-3 text-xs text-muted-foreground">
                أدخل رمز تطبيق المصادقة المكوّن من 6 أرقام أو أحد رموز الاسترداد.
                <br />
                Enter the 6-digit authenticator code or a recovery code.
                {challengeExpiresAt && (
                  <span className="mt-1 block" data-testid="mfa-expiry">
                    {`Valid until ${new Date(challengeExpiresAt).toLocaleTimeString()}`}
                  </span>
                )}
              </div>
              <label className="text-xs font-semibold text-foreground block" htmlFor="login-mfa-code">
                رمز التحقق / Verification code
              </label>
              <div className="relative">
                <Shield className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <input
                  id="login-mfa-code"
                  type="text"
                  autoComplete="one-time-code"
                  autoCapitalize="characters"
                  spellCheck={false}
                  required
                  autoFocus
                  value={mfaCode}
                  onChange={(event) => setMFACode(event.target.value)}
                  className="w-full rounded-md border bg-background ps-9 pe-3 py-2 text-sm tracking-widest text-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
                />
              </div>
              <button
                type="button"
                onClick={() => {
                  setChallengeToken(null);
                  setChallengeExpiresAt(null);
                  setMFACode("");
                  setError(null);
                }}
                className="text-xs text-primary hover:underline"
              >
                العودة لتسجيل الدخول / Back to sign in
              </button>
            </div>
          )}

          {/* Submit Button */}
          <button
            id="login-submit-btn"
            type="submit"
            disabled={loading || success}
            className="w-full mt-2 inline-flex items-center justify-center rounded-md bg-primary py-2.5 px-4 text-sm font-semibold text-primary-foreground shadow hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/20 disabled:opacity-50 transition-colors"
          >
            {loading ? t.auth.signing_in : challengeToken ? "تحقق / Verify" : t.auth.sign_in}
          </button>
        </form>

      </div>
    </div>
  );
}

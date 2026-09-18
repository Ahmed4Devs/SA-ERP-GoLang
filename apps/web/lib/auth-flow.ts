export interface MFALoginChallenge {
  challengeToken: string;
  expiresAt: string | null;
}

export function getMFAChallenge(payload: unknown): MFALoginChallenge | null {
  if (!payload || typeof payload !== "object") return null;

  const response = payload as Record<string, unknown>;
  if (response.mfa_required !== true) return null;
  if (typeof response.challenge_token !== "string" || response.challenge_token.trim() === "") {
    throw new Error("Authentication server returned an invalid MFA challenge");
  }

  return {
    challengeToken: response.challenge_token,
    expiresAt: typeof response.expires_at === "string" ? response.expires_at : null,
  };
}

export function getSafeReturnPath(search: string, fallback = "/"): string {
  const requested = new URLSearchParams(search).get("next");
  if (!requested || !requested.startsWith("/") || requested.startsWith("//") || requested.includes("\\")) {
    return fallback;
  }

  try {
    const base = new URL("https://erp.local");
    const target = new URL(requested, base);
    return target.origin === base.origin
      ? `${target.pathname}${target.search}${target.hash}`
      : fallback;
  } catch {
    return fallback;
  }
}

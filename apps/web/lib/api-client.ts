export class APIError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

export function apiURL(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  const configuredBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";
  return `${configuredBase}${normalizedPath}`;
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type") && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(apiURL(path), {
    ...init,
    headers,
    credentials: "include",
    cache: "no-store",
  });

  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    const message =
      typeof payload === "object" && payload !== null
        ? payload.error?.message ?? payload.error ?? `Request failed (${response.status})`
        : payload || `Request failed (${response.status})`;
    const code = typeof payload === "object" && payload !== null ? payload.error?.code : undefined;
    throw new APIError(String(message), response.status, code);
  }

  return payload as T;
}

export function apiBaseUrl() {
  const isServer = typeof window === "undefined";
  const root = isServer
    ? (process.env.API_INTERNAL_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080");
  const version = process.env.NEXT_PUBLIC_API_VERSION ?? "v1";
  return `${root.replace(/\/$/, "")}/${version}`;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const url = `${apiBaseUrl()}${path}`;
  const headers = new Headers(init?.headers ?? {});
  headers.set("Content-Type", "application/json");
  const res = await fetch(url, {
    ...init,
    headers,
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `Request failed with status ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

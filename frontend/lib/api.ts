import { getToken } from "@/lib/auth";

const DEFAULT_API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080";

function normalizeBase(base: string) {
  return base.replace(/\/+$/, "");
}

export function apiUrl(path: string) {
  const base = normalizeBase(DEFAULT_API_BASE);
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${base}${normalized}`;
}

type JsonBody = Record<string, unknown> | unknown[];
type FetchOptions = Omit<RequestInit, "body"> & {
  body?: JsonBody | string | null;
};

export async function fetchJson<T>(
  path: string,
  options: FetchOptions = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  const token = getToken();

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  let resolvedBody: BodyInit | undefined;
  const body = options.body;
  const isStringBody = typeof body === "string";
  if (body !== null && body !== undefined) {
    if (isStringBody) {
      resolvedBody = body;
    } else {
      headers.set("Content-Type", "application/json");
      resolvedBody = JSON.stringify(body);
    }
  }

  const response = await fetch(apiUrl(path), {
    ...options,
    headers,
    body: resolvedBody,
  });

  if (!response.ok) {
    const errorText = await response.text();
    const message = errorText || `Request failed with ${response.status}`;
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get("content-type") ?? "";
  const rawText = await response.text();

  if (contentType.includes("application/json")) {
    return JSON.parse(rawText) as T;
  }

  try {
    return JSON.parse(rawText) as T;
  } catch {
    return rawText as T;
  }
}

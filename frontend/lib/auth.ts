import type { CompanyProfile } from "@/lib/types";

const TOKEN_KEY = "monitor.jwt";
const COMPANY_KEY = "monitor.company";

export function getToken() {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(TOKEN_KEY);
}

export function getCompanyProfile(): CompanyProfile | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(COMPANY_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as CompanyProfile;
  } catch {
    return null;
  }
}

export function setCompanyProfile(profile: CompanyProfile) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(COMPANY_KEY, JSON.stringify(profile));
}

export function clearCompanyProfile() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(COMPANY_KEY);
}

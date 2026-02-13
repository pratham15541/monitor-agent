"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { fetchJson } from "@/lib/api";
import { getCompanyProfile, getToken, setCompanyProfile } from "@/lib/auth";
import type { CompanyProfile } from "@/lib/types";

export default function RegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [registered, setRegistered] = useState<CompanyProfile | null>(
    getCompanyProfile(),
  );

  useEffect(() => {
    if (getToken()) {
      router.replace("/dashboard");
    }
  }, [router]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const profile = await fetchJson<CompanyProfile>("/auth/register", {
        method: "POST",
        body: { name, email, password },
      });
      setCompanyProfile(profile);
      setRegistered(profile);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  function handleCopyToken() {
    if (!registered?.apiToken) return;
    void navigator.clipboard.writeText(registered.apiToken);
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-6 py-12">
      <Card className="w-full max-w-xl border-0 bg-white/85 shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl">Create your company</CardTitle>
          <p className="text-sm text-muted-foreground">
            Register a company account and grab the API token for your agents.
          </p>
        </CardHeader>
        <CardContent className="space-y-6">
          {registered ? (
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4">
              <p className="text-sm font-medium text-emerald-700">
                Company registered
              </p>
              <div className="mt-3 space-y-2 text-sm text-emerald-900">
                <div className="flex items-center justify-between">
                  <span>Company</span>
                  <span className="font-medium">{registered.name}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span>Email</span>
                  <span className="font-medium">{registered.email}</span>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <span>API token</span>
                  <span className="break-all font-mono text-xs">
                    {registered.apiToken}
                  </span>
                </div>
                <Button
                  className="mt-3 w-full"
                  type="button"
                  onClick={handleCopyToken}
                >
                  Copy API token
                </Button>
              </div>
            </div>
          ) : null}

          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="name">
                Company name
              </label>
              <Input
                id="name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="Ops Division"
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="email">
                Email
              </label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="ops@company.com"
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="password">
                Password
              </label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Create a password"
                required
              />
            </div>
            {error ? (
              <p className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
                {error}
              </p>
            ) : null}
            <Button className="w-full" disabled={loading} type="submit">
              {loading ? "Registering..." : "Register company"}
            </Button>
          </form>

          <div className="text-center text-sm text-muted-foreground">
            Already have an account?{" "}
            <Link className="font-medium text-primary" href="/login">
              Sign in
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

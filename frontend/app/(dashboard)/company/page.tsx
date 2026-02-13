"use client";

import Link from "next/link";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { clearCompanyProfile, getCompanyProfile } from "@/lib/auth";

export default function CompanyPage() {
  const [profile, setProfile] = useState(getCompanyProfile());

  function handleCopyToken() {
    if (!profile?.apiToken) return;
    void navigator.clipboard.writeText(profile.apiToken);
  }

  function handleClear() {
    clearCompanyProfile();
    setProfile(null);
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
      <Card>
        <CardHeader>
          <CardTitle>Company profile</CardTitle>
          <p className="text-sm text-muted-foreground">
            Store the API token locally to speed up agent onboarding.
          </p>
        </CardHeader>
        <CardContent>
          {profile ? (
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Company</span>
                <span className="font-medium">{profile.name}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Email</span>
                <span className="font-medium">{profile.email}</span>
              </div>
              <div className="flex items-center justify-between gap-4">
                <span className="text-muted-foreground">API token</span>
                <span className="break-all font-mono text-xs">
                  {profile.apiToken}
                </span>
              </div>
              <div className="flex flex-wrap gap-3 pt-3">
                <Button onClick={handleCopyToken}>Copy API token</Button>
                <Button variant="outline" onClick={handleClear}>
                  Clear local profile
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4 text-sm text-muted-foreground">
              <p>No local company profile found.</p>
              <Link className="text-primary" href="/register">
                Register a company to generate an API token
              </Link>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-dashed">
        <CardHeader>
          <CardTitle>Next steps</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          <p>1. Register company and copy the API token.</p>
          <p>2. Register agents with hostname + IP.</p>
          <p>3. Watch the dashboard update with live metrics.</p>
        </CardContent>
      </Card>
    </div>
  );
}

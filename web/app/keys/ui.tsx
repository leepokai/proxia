"use client";

import { useState } from "react";
import { createSupabaseBrowserClient } from "@/lib/supabase/browser";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type ApiKeyRow = {
  id: string;
  name: string | null;
  status: string;
  key_prefix: string;
  rate_limit_rps: number;
  rate_limit_burst: number;
  created_at: string;
  expires_at: string | null;
};

export default function KeysUI({
  initialKeys,
}: {
  initialKeys: ApiKeyRow[];
}) {
  const [keys, setKeys] = useState<ApiKeyRow[]>(initialKeys);
  const [rawKey, setRawKey] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");

  async function load() {
    setLoading(true);
    setErr("");
    const res = await fetch("/api/keys", { cache: "no-store" });
    const json = await res.json();
    if (!res.ok) {
      setErr(json.error ?? "failed");
      setLoading(false);
      return;
    }
    setKeys(json.keys ?? []);
    setLoading(false);
  }

  async function createKey() {
    setErr("");
    setRawKey("");
    const res = await fetch("/api/keys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: "Default", rate_limit_rps: 3, rate_limit_burst: 10 }),
    });
    const json = await res.json();
    if (!res.ok) {
      setErr(json.error ?? "failed");
      return;
    }
    setRawKey(json.raw_key ?? "");
    await load();
  }

  async function revoke(id: string) {
    setErr("");
    const res = await fetch("/api/keys", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id }),
    });
    const json = await res.json();
    if (!res.ok) {
      setErr(json.error ?? "failed");
      return;
    }
    await load();
  }

  async function signOut() {
    const supabase = createSupabaseBrowserClient();
    await supabase.auth.signOut();
    window.location.href = "/login";
  }

  return (
    <main className="mx-auto max-w-3xl p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">API Keys</h1>
        <Button variant="ghost" className="text-sm" onClick={signOut}>
          Sign out
        </Button>
      </div>

      <Card>
        <CardContent className="space-y-3 pt-4">
          <div className="flex items-center gap-3">
            <Button onClick={createKey}>Create key</Button>
            <span className="text-sm text-gray-600">Default limits: 3 rps / burst 10</span>
          </div>
          {rawKey ? (
            <div className="rounded bg-yellow-50 border border-yellow-200 p-3">
              <div className="text-sm font-medium mb-1">Copy this key now (shown once):</div>
              <pre className="text-xs overflow-auto p-2 bg-white border rounded">{rawKey}</pre>
              <div className="text-xs text-gray-600 mt-2">
                Use it with: <code>Authorization: Bearer {rawKey}</code>
              </div>
            </div>
          ) : null}
          {err ? <p className="text-sm text-red-600">{err}</p> : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Your keys</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-sm text-gray-600">Loading...</p>
          ) : keys.length === 0 ? (
            <p className="text-sm text-gray-600">No keys yet.</p>
          ) : (
            <ul className="space-y-3">
              {keys.map((k) => (
                <li key={k.id} className="flex items-center justify-between rounded border p-3">
                  <div className="space-y-1">
                    <div className="text-sm font-medium">
                      {k.name ?? "Unnamed"}{" "}
                      <span className="text-xs text-gray-500">({k.status})</span>
                    </div>
                    <div className="text-xs text-gray-600">
                      Prefix: <code>{k.key_prefix}</code>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    className="text-sm"
                    onClick={() => revoke(k.id)}
                    disabled={k.status !== "active"}
                  >
                    Revoke
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </main>
  );
}



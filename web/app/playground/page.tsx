"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function PlaygroundPage() {
  const [gatewayBase, setGatewayBase] = useState("http://localhost:8080");
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState("gpt-4o-mini");
  const [message, setMessage] = useState("Hello from Proxia playground");
  const [provider, setProvider] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string>("");
  const [error, setError] = useState<string>("");

  async function callChat() {
    setLoading(true);
    setResult("");
    setError("");
    try {
      const body: Record<string, unknown> = {
        model,
        messages: [{ role: "user", content: message }],
      };
      if (provider) {
        body.provider = provider;
      }
      if (model.toLowerCase().startsWith("claude") && body.max_tokens == null) {
        body.max_tokens = 256;
      }
      const res = await fetch(`${gatewayBase}/v1/chat`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(body),
      });
      const text = await res.text();
      if (!res.ok) {
        setError(text || `HTTP ${res.status}`);
      } else {
        setResult(text);
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : "request failed";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-neutral-950 via-neutral-930 to-neutral-900 text-neutral-100">
      <header className="border-b border-neutral-900 bg-neutral-950/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded bg-black text-white flex items-center justify-center text-sm font-semibold">
              GX
            </div>
            <div>
              <div className="text-sm text-gray-500">Proxia</div>
              <div className="text-base font-semibold">Playground</div>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm text-gray-700">
            <Link className="hover:text-black" href="/">
              Home
            </Link>
            <Link className="hover:text-black" href="/keys">
              Keys
            </Link>
            <Link className="hover:text-black font-medium" href="/playground">
              Playground
            </Link>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-10 space-y-6">
        <section className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardContent className="space-y-2">
              <CardTitle className="text-sm">Bring your own key</CardTitle>
              <CardDescription>
                Use a gateway key (rate-limited per account) or a direct provider key (OpenAI / Anthropic).
              </CardDescription>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="space-y-2">
              <CardTitle className="text-sm">Routing</CardTitle>
              <CardDescription>
                Model prefix auto-routes (<code>claude*</code> → Claude, <code>gpt*</code> / <code>o1*</code> → OpenAI).
                Set <code>provider</code> to force.
              </CardDescription>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="space-y-2">
              <CardTitle className="text-sm">Anthropic note</CardTitle>
              <CardDescription>
                Claude calls require <code>max_tokens</code>. We auto-fill 256 if the model starts with <code>claude</code>.
              </CardDescription>
            </CardContent>
          </Card>
        </section>

        <Card>
          <CardHeader className="flex items-center justify-between">
            <div>
              <CardTitle>Send a test call</CardTitle>
              <CardDescription>Calls /v1/chat directly</CardDescription>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="block text-sm font-medium">Gateway base URL</label>
                <Input
                  value={gatewayBase}
                  onChange={(e) => setGatewayBase(e.target.value)}
                  placeholder="http://localhost:8080"
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium">API key (gateway or provider)</label>
                <Input
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  type="password"
                  placeholder="Paste your key"
                />
                <div className="text-xs text-gray-500">
                  Gateway keys are rate-limited; provider keys go straight upstream.
                </div>
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="block text-sm font-medium">Model</label>
                <Input
                  value={model}
                  onChange={(e) => setModel(e.target.value)}
                  placeholder="gpt-4o-mini or claude-haiku-4-5"
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium">
                  Provider (optional, auto-route if empty)
                </label>
                <Input
                  value={provider}
                  onChange={(e) => setProvider(e.target.value)}
                  placeholder="openai | claude"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="block text-sm font-medium">User message</label>
              <Textarea
                rows={4}
                value={message}
                onChange={(e) => setMessage(e.target.value)}
              />
            </div>

            <div className="flex items-center gap-3">
              <Button onClick={callChat} disabled={loading || !apiKey}>
                {loading ? "Calling..." : "Send test call"}
              </Button>
              <div className="text-xs text-gray-500">
                Gateway-only fields (provider/base_url) are stripped before sending upstream.
              </div>
            </div>

            {error ? (
              <pre className="text-sm text-red-600 whitespace-pre-wrap rounded border border-red-200 bg-red-50 p-3">
                {error}
              </pre>
            ) : null}
            {result ? (
              <pre className="text-xs whitespace-pre-wrap rounded border bg-zinc-50 p-3 overflow-auto">
                {result}
              </pre>
            ) : null}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

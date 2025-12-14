"use client";

import { useState } from "react";
import { createSupabaseBrowserClient } from "@/lib/supabase/browser";

export default function LoginForm() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");
  const [message, setMessage] = useState<string>("");

  const siteUrl = (process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000").replace(/\/$/, "");
  const redirectTo = `${siteUrl}/auth/callback`;

  async function signInWithGoogle() {
    setStatus("sending");
    setMessage("");
    const supabase = createSupabaseBrowserClient();
    const { error } = await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        // Use a stable allowlisted URL and exchange the code server-side
        redirectTo,
      },
    });
    if (error) {
      setStatus("error");
      setMessage(error.message);
      return;
    }
    // Redirect is handled by Supabase
    setStatus("idle");
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("sending");
    setMessage("");

    const supabase = createSupabaseBrowserClient();
    const { error } = await supabase.auth.signInWithOtp({
      email,
      options: {
        emailRedirectTo: redirectTo,
      },
    });

    if (error) {
      setStatus("error");
      setMessage(error.message);
      return;
    }
    setStatus("sent");
    setMessage("Magic link sent. Check your email.");
  }

  return (
    <form onSubmit={onSubmit} className="space-y-3">
      <button
        className="w-full rounded border px-4 py-2"
        type="button"
        onClick={signInWithGoogle}
        disabled={status === "sending"}
      >
        Continue with Google
      </button>

      <div className="flex items-center gap-3 text-xs text-gray-500">
        <div className="h-px flex-1 bg-gray-200" />
        OR
        <div className="h-px flex-1 bg-gray-200" />
      </div>

      <label className="block text-sm font-medium">Email</label>
      <input
        className="w-full rounded border px-3 py-2"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="you@example.com"
        type="email"
        required
      />
      <button
        className="rounded bg-black px-4 py-2 text-white disabled:opacity-50"
        disabled={status === "sending"}
        type="submit"
      >
        {status === "sending" ? "Sending..." : "Send magic link"}
      </button>
      {message ? <p className="text-sm text-gray-600">{message}</p> : null}
    </form>
  );
}



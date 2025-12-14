import Link from "next/link";
import { Button } from "@/components/ui/button";

const Badge = ({ children }: { children: React.ReactNode }) => (
  <span className="inline-flex items-center gap-2 rounded-full border border-neutral-800 bg-neutral-900/60 px-3 py-1 text-xs uppercase tracking-[0.18em] text-neutral-300 shadow-[0_0_0_1px_rgba(255,255,255,0.05)]">
    {children}
  </span>
);

const Glow = () => (
  <div className="absolute inset-0 -z-10">
    <div className="absolute -left-32 top-10 h-64 w-64 rounded-full bg-indigo-500/20 blur-3xl" />
    <div className="absolute right-0 bottom-0 h-72 w-72 rounded-full bg-fuchsia-500/20 blur-3xl" />
  </div>
);

const FeatureCard = ({ title, desc, icon }: { title: string; desc: string; icon: string }) => (
  <div className="rounded-2xl border border-neutral-800 bg-neutral-900/70 p-6 shadow-[0_30px_80px_-40px_rgba(0,0,0,0.8)]">
    <div className="text-lg">{icon}</div>
    <div className="mt-3 text-sm font-semibold text-white">{title}</div>
    <p className="mt-2 text-sm text-neutral-400">{desc}</p>
  </div>
);

const Step = ({ title, desc, number }: { title: string; desc: string; number: number }) => (
  <div className="relative overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-950/60 p-5 shadow-[0_20px_60px_-40px_rgba(0,0,0,0.7)]">
    <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(99,102,241,0.08),transparent_35%)]" />
    <div className="relative flex items-start gap-3">
      <span className="flex h-8 w-8 items-center justify-center rounded-full border border-neutral-700 bg-neutral-900 text-xs font-semibold text-neutral-200">
        {number}
      </span>
      <div>
        <div className="text-sm font-semibold text-white">{title}</div>
        <p className="mt-1 text-sm text-neutral-400">{desc}</p>
      </div>
    </div>
  </div>
);

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-neutral-950 via-neutral-930 to-neutral-900 text-neutral-100">
      <header className="border-b border-neutral-900 bg-neutral-950/80 backdrop-blur">
        <div className="mx-auto max-w-6xl px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="h-9 w-9 rounded bg-indigo-500 text-white flex items-center justify-center text-sm font-semibold">
              GX
            </div>
            <div className="font-semibold">Proxia</div>
          </div>
          <nav className="flex items-center gap-4 text-sm text-neutral-300">
            <Link className="hover:text-white transition" href="/playground">
              Playground
            </Link>
            <Link className="hover:text-white transition" href="/login">
              Login
            </Link>
            <Link
              className="rounded bg-indigo-500 px-3 py-2 text-white hover:bg-indigo-400 transition shadow-lg shadow-indigo-500/20"
              href="/keys"
            >
              Dashboard
            </Link>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-14 space-y-14">
        <section className="relative overflow-hidden rounded-3xl border border-neutral-800 bg-gradient-to-br from-neutral-950 via-neutral-900 to-neutral-950 p-10 shadow-2xl shadow-indigo-500/10">
          <Glow />
          <div className="relative space-y-5">
            <Badge>Unified AI Gateway</Badge>
            <h1 className="text-4xl md:text-5xl font-semibold leading-tight text-white">
              OpenAI-compatible, multi-provider routing,
              <br />
              with your own keys.
            </h1>
            <p className="max-w-2xl text-neutral-300">
              Bring your OpenAI or Anthropic key, or issue gateway keys with per-key rate limits. One endpoint,
              automatic model-based routing, normalized responses.
            </p>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link href="/login">
                <Button className="w-full sm:w-auto">Sign in with Google</Button>
              </Link>
              <Link href="/playground">
                <Button variant="outline" className="w-full sm:w-auto">
                  Try the Playground
                </Button>
              </Link>
            </div>
          </div>
        </section>

        <section className="grid gap-4 md:grid-cols-3">
          <FeatureCard
            icon="🔑"
            title="Gateway & BYOK"
            desc="Use gateway-issued keys or your own upstream keys (OpenAI / Anthropic)."
          />
          <FeatureCard
            icon="⚡"
            title="Per-key limits"
            desc="Rate limit per key (RPS + burst) with consistent error messaging."
          />
          <FeatureCard
            icon="🌉"
            title="Model-aware routing"
            desc="Auto-route by model prefix; override with provider/base_url when needed."
          />
        </section>

        <section className="space-y-6">
          <div className="space-y-2">
            <Badge>Flow</Badge>
            <h2 className="text-3xl font-semibold text-white">Three steps to live traffic</h2>
            <p className="text-neutral-300">Issue a key, drop it into your OpenAI client, and start sending.</p>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            <Step number={1} title="Create a gateway key" desc="Go to /keys and mint a key (or use your upstream key directly)." />
            <Step number={2} title="Pick your model" desc="gpt-* routes to OpenAI, claude-* to Anthropic. Add max_tokens for Claude." />
            <Step number={3} title="Call /v1/chat" desc="Keep the OpenAI payload shape; gateway strips provider/base_url before sending." />
          </div>
        </section>

        <section className="relative overflow-hidden rounded-3xl border border-neutral-800 bg-neutral-950/80 p-8 shadow-xl shadow-black/30">
          <Glow />
          <div className="relative flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="space-y-2">
              <Badge>Ready to try?</Badge>
              <h3 className="text-2xl font-semibold text-white">Spin up a request in under a minute.</h3>
              <p className="text-neutral-300 max-w-xl">
                Use the Playground to verify keys and models, then drop the same payloads into your app.
              </p>
            </div>
            <div className="flex gap-3">
              <Link href="/playground">
                <Button className="w-full sm:w-auto">Open Playground</Button>
              </Link>
              <Link href="/keys">
                <Button variant="outline" className="w-full sm:w-auto">
                  Issue a key
                </Button>
              </Link>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-neutral-900 bg-neutral-950">
        <div className="mx-auto max-w-6xl px-6 py-6 text-xs text-neutral-500">
          Proxia — unified AI gateway. Playground-ready.
        </div>
      </footer>
    </div>
  );
}

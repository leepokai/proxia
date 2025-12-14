import Link from "next/link";
import KeysUI from "./ui";
import { createSupabaseServerClient } from "@/lib/supabase/server";

export default async function KeysPage() {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  // middleware should already protect this route, but keep it safe
  if (!user) {
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
                <div className="text-base font-semibold">Keys</div>
              </div>
            </div>
            <nav className="flex items-center gap-4 text-sm text-neutral-300">
              <Link className="hover:text-white" href="/">
                Home
              </Link>
              <Link className="hover:text-white font-medium" href="/keys">
                Keys
              </Link>
              <Link className="hover:text-white" href="/playground">
                Playground
              </Link>
            </nav>
          </div>
        </header>
        <main className="mx-auto max-w-3xl p-6">
          <p className="text-sm text-neutral-300">Not signed in.</p>
        </main>
      </div>
    );
  }

  const { data: keys } = await supabase
    .from("api_keys")
    .select("id,name,status,key_prefix,rate_limit_rps,rate_limit_burst,created_at,expires_at")
    .order("created_at", { ascending: false });

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
              <div className="text-base font-semibold">Keys</div>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm text-neutral-300">
            <Link className="hover:text-white" href="/">
              Home
            </Link>
            <Link className="hover:text-white font-medium" href="/keys">
              Keys
            </Link>
            <Link className="hover:text-white" href="/playground">
              Playground
            </Link>
          </nav>
        </div>
      </header>

      <KeysUI initialKeys={keys ?? []} />
    </div>
  );
}



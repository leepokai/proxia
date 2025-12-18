import { NextResponse } from "next/server";
import crypto from "crypto";
import { createSupabaseServerClient } from "@/lib/supabase/server";

function derive32Key(secret: string): Buffer {
  // Accept base64 32 bytes, else sha256(secret)
  try {
    const b = Buffer.from(secret, "base64");
    if (b.length === 32) return b;
  } catch {
    // ignore
  }
  return crypto.createHash("sha256").update(secret).digest();
}

function encryptString(plaintext: string, secret: string) {
  const key = derive32Key(secret);
  const iv = crypto.randomBytes(12); // GCM nonce
  const cipher = crypto.createCipheriv("aes-256-gcm", key, iv);
  const ct = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()]);
  const tag = cipher.getAuthTag();
  // match Go: base64(nonce||ciphertext||tag) - but Go used gcm.Seal which appends tag to ct automatically.
  // Node separates tag; we append tag to ciphertext to match gcm output.
  const out = Buffer.concat([iv, ct, tag]);
  return out.toString("base64");
}

export async function GET() {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const { data, error } = await supabase
    .from("provider_credentials")
    .select("id,provider,name,created_at")
    .order("created_at", { ascending: false });

  if (error) return NextResponse.json({ error: error.message }, { status: 500 });
  return NextResponse.json({ credentials: data ?? [] });
}

export async function POST(request: Request) {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const body = (await request.json().catch(() => ({}))) as {
    provider?: string;
    name?: string;
    api_key?: string;
  };
  const provider = (body.provider ?? "").trim();
  const name = (body.name ?? "default").trim() || "default";
  const apiKey = (body.api_key ?? "").trim();

  if (!provider || !apiKey) {
    return NextResponse.json({ error: "missing provider/api_key" }, { status: 400 });
  }

  const encKey = process.env.ENCRYPTION_KEY;
  if (!encKey) {
    return NextResponse.json(
      { error: "missing ENCRYPTION_KEY on server" },
      { status: 500 },
    );
  }

  const encrypted_key = encryptString(apiKey, encKey);

  const { data, error } = await supabase
    .from("provider_credentials")
    .upsert(
      {
        user_id: user.id,
        provider,
        name,
        encrypted_key,
      },
      { onConflict: "user_id,provider,name" },
    )
    .select("id,provider,name,created_at")
    .single();

  if (error) return NextResponse.json({ error: error.message }, { status: 500 });
  return NextResponse.json({ credential: data });
}






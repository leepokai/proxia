import { NextResponse } from "next/server";
import crypto from "crypto";
import { createSupabaseServerClient } from "@/lib/supabase/server";

function generateGatewayKey() {
  const raw = `sk_live_${crypto.randomBytes(24).toString("hex")}`;
  const keyHash = crypto.createHash("sha256").update(raw).digest("hex");
  const keyPrefix = raw.slice(0, 12);
  return { raw, keyHash, keyPrefix };
}

export async function GET() {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
    error: userErr,
  } = await supabase.auth.getUser();

  if (userErr || !user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const { data, error } = await supabase
    .from("api_keys")
    .select("id,name,status,key_prefix,rate_limit_rps,rate_limit_burst,created_at,expires_at")
    .order("created_at", { ascending: false });

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }
  return NextResponse.json({ keys: data ?? [] });
}

export async function POST(request: Request) {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
    error: userErr,
  } = await supabase.auth.getUser();

  if (userErr || !user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = (await request.json().catch(() => ({}))) as {
    name?: string;
    rate_limit_rps?: number;
    rate_limit_burst?: number;
  };

  const { raw, keyHash, keyPrefix } = generateGatewayKey();

  const payload = {
    user_id: user.id,
    name: body.name ?? "Default",
    status: "active",
    key_prefix: keyPrefix,
    key_hash: keyHash,
    rate_limit_rps: body.rate_limit_rps ?? 3,
    rate_limit_burst: body.rate_limit_burst ?? 10,
  };

  const { data, error } = await supabase.from("api_keys").insert(payload).select().single();
  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  // IMPORTANT: return raw key ONLY once
  return NextResponse.json({ key: data, raw_key: raw });
}

export async function DELETE(request: Request) {
  const supabase = await createSupabaseServerClient();
  const {
    data: { user },
    error: userErr,
  } = await supabase.auth.getUser();

  if (userErr || !user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = (await request.json().catch(() => ({}))) as { id?: string };
  if (!body.id) {
    return NextResponse.json({ error: "missing id" }, { status: 400 });
  }

  const { error } = await supabase
    .from("api_keys")
    .update({ status: "revoked" })
    .eq("id", body.id);

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  return NextResponse.json({ ok: true });
}



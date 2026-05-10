export type RemoteJoinInfo = {
  host: string;
  port: string;
  code: string;
};

const inviteScheme = "sakiko-remote://join";

export function buildRemoteJoinText(info: RemoteJoinInfo): string {
  const host = info.host.trim();
  const port = info.port.trim();
  const code = info.code.trim();
  if (!host || !port || !code) {
    return "";
  }

  const params = new URLSearchParams({
    host,
    port,
    code,
  });
  return `${inviteScheme}?${params.toString()}`;
}

export function parseRemoteJoinText(input: string): RemoteJoinInfo | null {
  const raw = input.trim();
  if (!raw) {
    return null;
  }

  const fromURL = parseRemoteJoinURL(raw);
  if (fromURL) {
    return fromURL;
  }

  const values = new Map<string, string>();
  const pattern = /\b(host|ip|address|port|code|pairingCode|oneTimeCode)\s*[:=]\s*([^\s,;]+)/gi;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(raw)) !== null) {
    values.set(match[1].toLowerCase(), trimToken(match[2]));
  }

  const host = values.get("host") || values.get("ip") || values.get("address") || "";
  const port = values.get("port") || "";
  const code = values.get("code") || values.get("pairingcode") || values.get("onetimecode") || "";
  if (!host || !port || !code) {
    return null;
  }
  return { host, port, code };
}

function parseRemoteJoinURL(raw: string): RemoteJoinInfo | null {
  try {
    const parsed = new URL(raw);
    if (parsed.protocol !== "sakiko-remote:" || parsed.hostname !== "join") {
      return null;
    }
    const host = parsed.searchParams.get("host")?.trim() || "";
    const port = parsed.searchParams.get("port")?.trim() || "";
    const code = parsed.searchParams.get("code")?.trim() || "";
    if (!host || !port || !code) {
      return null;
    }
    return { host, port, code };
  } catch {
    return null;
  }
}

function trimToken(value: string): string {
  return value.trim().replace(/^["']|["']$/g, "");
}

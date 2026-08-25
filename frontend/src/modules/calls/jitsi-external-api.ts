const INVALID_ORIGIN_MESSAGE = "Invalid Jitsi external API origin";
const LABEL_PATTERN = /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/;

function invalidOrigin(): never {
  throw new Error(INVALID_ORIGIN_MESSAGE);
}

function isSelfHostedJitsiHostname(hostname: string): boolean {
  const labels = hostname.split(".");

  return (
    hostname.length <= 253 &&
    labels.length > 1 &&
    labels[0] === "meet" &&
    labels.every((label) => label.length > 0 && label.length <= 63 && LABEL_PATTERN.test(label))
  );
}

export function getJitsiExternalAPIScriptURL(origin: string): string {
  let parsed: URL;

  try {
    parsed = new URL(origin);
  } catch {
    return invalidOrigin();
  }

  const hostname = parsed.hostname;

  if (
    origin !== `https://${hostname}` ||
    hostname === "meet.jit.si" ||
    hostname.endsWith(".meet.jit.si") ||
    !isSelfHostedJitsiHostname(hostname)
  ) {
    return invalidOrigin();
  }

  return `${origin}/external_api.js`;
}

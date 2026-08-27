import { getJitsiExternalAPIScriptURL } from "./jitsi-external-api";

const INVALID_CONTRACT_MESSAGE = "Invalid Jitsi embed contract";

export interface JitsiEmbedContractInput {
  provider: string;
  roomName: string;
  serviceURL: string;
  token: string;
}

export interface JitsiEmbedContract {
  externalApiScriptURL: string;
  domain: string;
  roomName: string;
  jwt: string;
}

function invalidContract(): never {
  throw new Error(INVALID_CONTRACT_MESSAGE);
}

function isCanonicalNonBlank(value: string): boolean {
  return value.length > 0 && value === value.trim();
}

export function buildJitsiEmbedContract(input: JitsiEmbedContractInput): JitsiEmbedContract {
  try {
    if (
      input.provider !== "jitsi" ||
      !isCanonicalNonBlank(input.roomName) ||
      !isCanonicalNonBlank(input.token)
    ) {
      return invalidContract();
    }

    const externalApiScriptURL = getJitsiExternalAPIScriptURL(input.serviceURL);
    const domain = new URL(input.serviceURL).hostname;

    if (domain.length === 0) {
      return invalidContract();
    }

    return {
      externalApiScriptURL,
      domain,
      roomName: input.roomName,
      jwt: input.token,
    };
  } catch {
    return invalidContract();
  }
}

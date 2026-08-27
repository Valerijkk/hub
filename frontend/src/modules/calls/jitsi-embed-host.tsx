import { forwardRef } from "react";

import {
  buildJitsiEmbedContract,
  type JitsiEmbedContractInput,
} from "./jitsi-embed-contract";

export interface JitsiEmbedHostProps {
  credential: JitsiEmbedContractInput | null;
}

export const JitsiEmbedHost = forwardRef<HTMLDivElement, JitsiEmbedHostProps>(
  function JitsiEmbedHost({ credential }, ref) {
    if (credential === null) {
      return null;
    }

    try {
      buildJitsiEmbedContract(credential);
    } catch {
      return null;
    }

    return <div ref={ref} data-jitsi-embed-host="" />;
  },
);

JitsiEmbedHost.displayName = "JitsiEmbedHost";

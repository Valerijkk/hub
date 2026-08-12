import { createFileRoute } from "@tanstack/react-router";

/**
 * A13-017: stub — CRM page lives in `crm.lazy.tsx` (code-split from main graph).
 */
export const Route = createFileRoute("/_authenticated/crm")({});

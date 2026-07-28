/**
 * TASK-052 / P2-FE-08: compile-fail harness.
 * Omit `version` must be a TS error. If `version` becomes optional,
 * `@ts-expect-error` becomes unused → `tsc` RED.
 */
import { updateBase, updateRecord } from "@/api/bases";

declare const baseId: string;
declare const tableId: string;
declare const recordId: string;

// @ts-expect-error version is required on updateBase
updateBase(baseId, { name: "x" });

// @ts-expect-error version is required on updateRecord
updateRecord(baseId, tableId, recordId, { cells: {} });

/**
 * TASK-130 / A3W11-2: compile-fail harness.
 * Omit `version` must be a TS error. If `version` becomes optional,
 * `@ts-expect-error` becomes unused → `tsc` RED (mutation proof).
 */
import { updateProject } from "@/api/projects";
import { updateSprint } from "@/api/sprint";

declare const projectId: string;
declare const sprintId: string;

// @ts-expect-error version is required on updateProject
updateProject(projectId, { name: "x" });

// @ts-expect-error version is required on updateSprint
updateSprint(projectId, sprintId, { name: "x" });

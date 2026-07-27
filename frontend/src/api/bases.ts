import { api } from "@/api/client";

export type ColumnType =
	| "text"
	| "number"
	| "date"
	| "select"
	| "checkbox"
	| "url"
	| "user_ref";

export interface BaseSummary {
	id: string;
	name: string;
	description?: string;
	version: number;
	owner_id: string;
	owner_name?: string;
	table_count: number;
	created_at: string;
	updated_at: string;
}

export interface TableSummary {
	id: string;
	name: string;
	sort_order: number;
	version: number;
}

export interface ColumnDef {
	id: string;
	name: string;
	col_type: ColumnType;
	options?: Record<string, unknown>;
	sort_order: number;
	version: number;
}

export interface TableRecord {
	id: string;
	cells: Record<string, unknown>;
	version: number;
	created_by?: string;
	created_at: string;
	updated_at: string;
}

export interface BaseDetail extends BaseSummary {
	tables: TableSummary[];
}

export interface TableView {
	base_id: string;
	table: TableSummary;
	columns: ColumnDef[];
	records: TableRecord[];
}

export function listBases() {
	return api.get<{ bases: BaseSummary[] }>("/bases").then((r) => r.bases);
}

export function getBase(id: string) {
	return api.get<BaseDetail>(`/bases/${id}`);
}

export function createBase(data: { name: string; description?: string }) {
	return api.post<BaseDetail>("/bases", data);
}

export function updateBase(
	baseId: string,
	data: { name?: string; description?: string; version?: number },
) {
	return api.put<BaseDetail>(`/bases/${baseId}`, data);
}

export function deleteBase(baseId: string) {
	return api.delete<void>(`/bases/${baseId}`);
}

export function getTableView(baseId: string, tableId: string) {
	return api.get<TableView>(`/bases/${baseId}/tables/${tableId}`);
}

export function createTable(baseId: string, data: { name: string }) {
	return api.post<TableSummary>(`/bases/${baseId}/tables`, data);
}

export function createColumn(
	baseId: string,
	tableId: string,
	data: { name: string; col_type?: ColumnType },
) {
	return api.post<ColumnDef>(
		`/bases/${baseId}/tables/${tableId}/columns`,
		data,
	);
}

export function createRecord(
	baseId: string,
	tableId: string,
	cells: Record<string, unknown>,
) {
	return api.post<TableRecord>(`/bases/${baseId}/tables/${tableId}/records`, {
		cells,
	});
}

export function updateRecord(
	baseId: string,
	tableId: string,
	recordId: string,
	data: { cells: Record<string, unknown>; version?: number },
) {
	return api.put<TableRecord>(
		`/bases/${baseId}/tables/${tableId}/records/${recordId}`,
		data,
	);
}

export function deleteRecord(
	baseId: string,
	tableId: string,
	recordId: string,
) {
	return api.delete<void>(
		`/bases/${baseId}/tables/${tableId}/records/${recordId}`,
	);
}

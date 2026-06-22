import { api } from "./client";

export type LeaveType = "vacation" | "sick" | "personal" | "other";
export type LeaveStatus = "pending" | "approved" | "rejected" | "cancelled";

export interface LeaveRequest {
	id: string;
	leave_type: LeaveType;
	status: LeaveStatus;
	start_date: string;
	end_date: string;
	reason?: string;
	review_note?: string;
	requester_id: string;
	requester_name?: string;
	reviewer_id?: string;
	reviewer_name?: string;
	reviewed_at?: string;
	version: number;
	created_at: string;
	updated_at: string;
}

export const LEAVE_TYPE_LABELS: Record<LeaveType, string> = {
	vacation: "Отпуск",
	sick: "Больничный",
	personal: "Отгул",
	other: "Другое",
};

export const LEAVE_STATUS_LABELS: Record<LeaveStatus, string> = {
	pending: "На согласовании",
	approved: "Одобрено",
	rejected: "Отклонено",
	cancelled: "Отменено",
};

export function listMyLeave(): Promise<{ items: LeaveRequest[] }> {
	return api.get<{ items: LeaveRequest[] }>("/hr/leave");
}

export function listPendingLeave(): Promise<{ items: LeaveRequest[] }> {
	return api.get<{ items: LeaveRequest[] }>("/hr/leave/pending");
}

export function createLeave(data: {
	leave_type: LeaveType;
	start_date: string;
	end_date: string;
	reason?: string;
}): Promise<LeaveRequest> {
	return api.post<LeaveRequest>("/hr/leave", data);
}

export function approveLeave(
	id: string,
	version: number,
	review_note?: string,
): Promise<LeaveRequest> {
	return api.post<LeaveRequest>(`/hr/leave/${id}/approve`, {
		version,
		review_note,
	});
}

export function rejectLeave(
	id: string,
	version: number,
	review_note?: string,
): Promise<LeaveRequest> {
	return api.post<LeaveRequest>(`/hr/leave/${id}/reject`, {
		version,
		review_note,
	});
}

export function cancelLeave(
	id: string,
	version: number,
): Promise<LeaveRequest> {
	return api.post<LeaveRequest>(`/hr/leave/${id}/cancel`, { version });
}

import { api } from "./client";

/** Запись бэкапа. */
export interface BackupRecord {
  id: string;
  s3_key: string;
  file_size: number;
  status: "in_progress" | "completed" | "failed";
  error_message?: string;
  created_at: string;
}

/** Получить список бэкапов. */
export function listBackups() {
  return api.get<BackupRecord[]>("/admin/backups");
}

/** Запустить бэкап вручную. */
export function triggerBackup() {
  return api.post<BackupRecord>("/admin/backups");
}

/** Восстановить из бэкапа. */
export function restoreBackup(id: string) {
  return api.post<{ message: string }>(`/admin/backups/${id}/restore`, {
    confirm: "RESTORE",
  });
}

/** Удалить бэкап. */
export function deleteBackup(id: string) {
  return api.delete<void>(`/admin/backups/${id}`);
}

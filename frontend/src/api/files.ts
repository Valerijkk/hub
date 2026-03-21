import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Ответ с presigned URL для загрузки файла. */
interface UploadUrlResponse {
  file_id: string;
  upload_url: string;
}

/** Ответ с URL для скачивания файла. */
interface DownloadUrlResponse {
  downloadUrl: string;
}

// ---------------------------------------------------------------------------
// API-функции: Файлы
// ---------------------------------------------------------------------------

/** Endpoint: POST /files/upload-url — получить presigned URL для загрузки файла в хранилище. */
export function getUploadUrl(
  filename: string,
  contentType: string,
  projectId: string,
) {
  return api.post<UploadUrlResponse>("/files/upload-url", {
    filename,
    contentType,
    projectId,
  });
}

/** Endpoint: POST /files/:id/confirm — подтвердить загрузку файла и привязать к сущности. */
export function confirmUpload(
  fileId: string,
  entityType: string,
  entityId: string,
) {
  return api.post<{ status: string }>(`/files/${fileId}/confirm`, {
    entityType,
    entityId,
  });
}

/** Endpoint: GET /files/:id/url — получить URL для скачивания файла. */
export function getDownloadUrl(fileId: string) {
  return api.get<DownloadUrlResponse>(`/files/${fileId}/url`);
}

/** Endpoint: DELETE /files/:id — удалить файл. */
export function deleteFile(fileId: string) {
  return api.delete<{ status: string }>(`/files/${fileId}`);
}

/**
 * Полный цикл загрузки файла: получение presigned URL, PUT файла в хранилище, подтверждение.
 */
export async function uploadFile(
  file: File,
  projectId: string,
  entityType: string,
  entityId: string,
): Promise<string> {
  // 1. Get presigned upload URL
  const { file_id, upload_url } = await getUploadUrl(
    file.name,
    file.type || "application/octet-stream",
    projectId,
  );

  // 2. PUT file directly to storage (presigned URL)
  const putRes = await fetch(upload_url, {
    method: "PUT",
    headers: { "Content-Type": file.type || "application/octet-stream" },
    body: file,
  });

  if (!putRes.ok) {
    throw new Error(`Upload failed: ${putRes.status}`);
  }

  // 3. Confirm upload
  await confirmUpload(file_id, entityType, entityId);

  return file_id;
}

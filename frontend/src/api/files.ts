import { api } from "./client";

// --- Response types ---

interface UploadUrlResponse {
  file_id: string;
  upload_url: string;
}

interface DownloadUrlResponse {
  downloadUrl: string;
}

// --- API functions ---

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

export function getDownloadUrl(fileId: string) {
  return api.get<DownloadUrlResponse>(`/files/${fileId}/url`);
}

export function deleteFile(fileId: string) {
  return api.delete<{ status: string }>(`/files/${fileId}`);
}

/**
 * Upload a file end-to-end: get presigned URL, PUT file, confirm upload.
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

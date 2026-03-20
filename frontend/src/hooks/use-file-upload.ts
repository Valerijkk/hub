import { useState, useCallback, useRef, useEffect } from "react";
import { uploadFile } from "@/api/files";

interface UseFileUploadOptions {
  projectId: string;
  entityType: string;
  entityId: string;
  onUpload?: (fileId: string, file: File) => void;
  onError?: (error: unknown) => void;
}

interface UseFileUploadReturn {
  isDragging: boolean;
  isUploading: boolean;
  upload: (files: FileList | File[]) => Promise<void>;
  dropZoneProps: {
    onDragOver: (e: React.DragEvent) => void;
    onDragLeave: (e: React.DragEvent) => void;
    onDrop: (e: React.DragEvent) => void;
  };
}

/**
 * useFileUpload provides drag-and-drop and paste upload functionality.
 *
 * Handles dragover, drop events, and Ctrl+V paste events. Uploads files
 * via the presigned URL flow (get URL -> PUT -> confirm).
 */
export function useFileUpload(
  options: UseFileUploadOptions,
): UseFileUploadReturn {
  const { projectId, entityType, entityId, onUpload, onError } = options;
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const dragCounter = useRef(0);

  const upload = useCallback(
    async (files: FileList | File[]) => {
      const fileArray = Array.from(files);
      if (fileArray.length === 0) return;

      setIsUploading(true);
      try {
        for (const file of fileArray) {
          const fileId = await uploadFile(file, projectId, entityType, entityId);
          onUpload?.(fileId, file);
        }
      } catch (err) {
        onError?.(err);
      } finally {
        setIsUploading(false);
      }
    },
    [projectId, entityType, entityId, onUpload, onError],
  );

  // Handle Ctrl+V paste events globally.
  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;

      const files: File[] = [];
      for (const item of items) {
        if (item.kind === "file") {
          const file = item.getAsFile();
          if (file) files.push(file);
        }
      }

      if (files.length > 0) {
        e.preventDefault();
        void upload(files);
      }
    };

    document.addEventListener("paste", handlePaste);
    return () => document.removeEventListener("paste", handlePaste);
  }, [upload]);

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current++;
    setIsDragging(true);
  }, []);

  const onDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounter.current--;
    if (dragCounter.current === 0) {
      setIsDragging(false);
    }
  }, []);

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounter.current = 0;
      setIsDragging(false);

      if (e.dataTransfer.files.length > 0) {
        void upload(e.dataTransfer.files);
      }
    },
    [upload],
  );

  return {
    isDragging,
    isUploading,
    upload,
    dropZoneProps: { onDragOver, onDragLeave, onDrop },
  };
}

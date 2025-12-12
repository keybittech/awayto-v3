import { useCallback, useMemo, useState } from 'react';
import { BufferResponse, IFile } from './api';
import { useUtil } from './useUtil';
import { useAppSelector } from './store';
import { decryptData, encryptData } from './pqc';

export type UseFileContents = () => {
  fileContents: IFile | undefined;
  postFileContents: (uploadId: string, fileRef: File[], existingIds: string[], overwriteIds: string[]) => Promise<string[]>;
  getFileContents: (fileRef: Partial<IFile>, download?: boolean) => Promise<BufferResponse | undefined>;
}

export interface IPreviewFile extends File {
  preview?: string;
}

export interface OrderedFiles {
  name: string;
  order: number;
  files: IFile[];
}

export const useFileContents: UseFileContents = () => {

  const { setSnack } = useUtil();
  const [fileContents, setFileContents] = useState<IFile | undefined>();
  const { vaultKey, sessionId } = useAppSelector(state => state.auth);

  // postFileContents and getFileContents are implemented manually instead of using RTK Query generated methods, in order to support binary transfer
  const postFileContents = useCallback<ReturnType<UseFileContents>['postFileContents']>(async (uploadId, fileRef, existingIds, overwriteIds) => {
    if (!vaultKey || !sessionId) return [];

    const fd = new FormData();
    fd.append('uploadId', uploadId);
    fd.append('overwriteIds', overwriteIds.join(","));
    fd.append('existingIds', existingIds.join(","));
    for (const f of fileRef) fd.append('contents', f);

    const payloadResponse = new Response(fd);
    const payloadBuffer = new Uint8Array(await payloadResponse.arrayBuffer());
    const multipartContentType = payloadResponse.headers.get('Content-Type') || '';

    const crypto = encryptData(vaultKey, sessionId, payloadBuffer);

    if (!crypto) {
      setSnack({ snackType: 'error', snackOn: "Encryption Failed" });
      return [];
    }

    const response = await fetch('/api/v1/files/content', {
      method: 'POST',
      body: crypto.blobBytes,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/x-awayto-vault',
        'X-Original-Content-Type': multipartContentType,
        'X-Tz': Intl.DateTimeFormat().resolvedOptions().timeZone
      }
    });

    const responseBody = await response.text();
    const decrypted = decryptData(crypto.secretB64, sessionId, responseBody);

    if (!decrypted) {
      setSnack({ snackType: 'error', snackOn: "Decryption Failed" });
      return [];
    }

    if (response.status !== 200) {
      console.error("Decrypt issue", decrypted.string);
      setSnack({ snackType: 'warning', snackOn: "File Decryption Unsuccessful" });
      return [];
    }

    try {
      const { ids } = JSON.parse(decrypted.string);
      return ids;
    } catch (e) {
      return [];
    }
  }, [vaultKey, sessionId]);

  const getFileContents = useCallback<ReturnType<UseFileContents>['getFileContents']>(async (fileRef, download) => {
    if (!fileRef.uuid || !fileRef.mimeType || !vaultKey || !sessionId) {
      setFileContents(undefined);
      return undefined;
    }

    const crypto = encryptData(vaultKey, sessionId, ' ');
    if (!crypto) return undefined;

    const response = await fetch(`/api/v1/files/content/${fileRef.uuid}`, {
      credentials: 'include',
      headers: {
        'X-Awayto-Vault': crypto.blobB64,
        'X-Tz': Intl.DateTimeFormat().resolvedOptions().timeZone,
      },
    });

    if (response.status !== 200) return undefined;

    const encryptedBody = await response.text();
    const decrypted = decryptData(crypto.secretB64, sessionId, encryptedBody);

    if (!decrypted) {
      setSnack({ snackType: 'error', snackOn: "Read File Decryption Failed" });
      return undefined;
    }

    const fileBlob = new Blob([decrypted.bytes], { type: 'application/pdf' });
    fileRef.url = window.URL.createObjectURL(fileBlob);

    setFileContents(fileRef as IFile);

    if (download) {
      const link = document.createElement('a');
      link.href = fileRef.url || "";
      link.setAttribute('download', 'downloaded-' + fileRef.name);
      link.click();
    }

    return fileRef as IFile;
  }, [vaultKey, sessionId]);

  return useMemo(() => ({ fileContents, postFileContents, getFileContents }), [fileContents]);
}

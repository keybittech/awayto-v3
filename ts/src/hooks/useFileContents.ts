import { useCallback, useMemo, useState } from 'react';
import { IFile } from './api';
import { keycloak, refreshToken } from './auth';
import { UseFileContents } from './util';
import { useUtil } from './useUtil';

export const useFileContents: UseFileContents = () => {

  const { setSnack } = useUtil();

  const [fileContents, setFileContents] = useState<IFile | undefined>();

  // postFileContents and getFileContents are implemented manually instead of using RTK Query generated methods, in order to support binary transfer

  const postFileContents = useCallback<ReturnType<UseFileContents>['postFileContents']>(async (uploadId, fileRef, existingIds, overwriteIds) => {
    const fd = new FormData();

    fd.append('uploadId', uploadId);
    fd.append('overwriteIds', overwriteIds.join(","));
    fd.append('existingIds', existingIds.join(","));

    for (const f of fileRef) {
      fd.append('contents', f);
    }

    await refreshToken();

    const headers = {
      'Authorization': `Bearer ${keycloak.token as string}`
    }

    const res = await fetch('/api/v1/files/content', {
      body: fd,
      method: 'POST',
      headers
    });

    if (200 !== res.status) {
      const errText = await res.text();
      setSnack({ snackType: 'warning', snackOn: errText });
      return [];
    }

    const { ids } = await res.json() as { ids: string[] };

    return ids;
  }, []);

  const getFileContents = useCallback<ReturnType<UseFileContents>['getFileContents']>(async (fileRef, download) => {
    if (!fileRef.uuid || !fileRef.mimeType) {
      setFileContents(undefined);
      return undefined;
    }

    await refreshToken();

    const headers = {
      'Authorization': `Bearer ${keycloak.token as string}`
    }

    const response = await fetch(`/api/v1/files/content/${fileRef.uuid}`, {
      headers
    });

    const fileBlob = await response.blob();

    fileRef.url = window.URL.createObjectURL(fileBlob);

    setFileContents(fileRef as IFile);

    if (download) {
      const link = document.createElement('a');
      link.id = 'site-file-downloader';
      link.href = fileRef.url || "";
      link.setAttribute('download', 'downloaded-' + fileRef.name); // or any other extension
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }

    return fileRef as IFile;
  }, []);

  return useMemo(() => ({ fileContents, postFileContents, getFileContents }), [fileContents]);
}

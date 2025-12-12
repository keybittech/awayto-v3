const enc = new TextEncoder();
const dec = new TextDecoder();
const toB64 = (bytes: Uint8Array) => btoa(String.fromCharCode(...bytes));
const fromB64 = (str: string) => Uint8Array.from(atob(str), c => c.charCodeAt(0));

export function encryptData(pubKeyB64: string, sessionId: string, data: string | Uint8Array) {
  if (!window.pqcEncrypt) throw new Error("WASM not loaded");

  const keyBytes = fromB64(pubKeyB64);
  const dataBytes = typeof data === 'string' ? enc.encode(data) : data;

  const result = window.pqcEncrypt(keyBytes, sessionId, dataBytes);
  if (!result) return null;

  return {
    blobBytes: result.blob,
    blobB64: toB64(result.blob),
    secretBytes: result.secret,
    secretB64: toB64(result.secret)
  };
}

export function decryptData(secretB64: string, sessionId: string, blob: string | Uint8Array) {
  if (!window.pqcDecrypt) throw new Error("WASM not loaded");

  const secretBytes = fromB64(secretB64);
  const blobBytes = typeof blob === 'string' ? fromB64(blob) : blob;

  const decrypted = window.pqcDecrypt(secretBytes, sessionId, blobBytes);
  if (!decrypted) return null;

  return {
    bytes: decrypted,
    string: dec.decode(decrypted)
  };
}

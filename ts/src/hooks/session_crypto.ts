async function getKey(vaultKey: string, userId: string) {
  const enc = new TextEncoder();
  const keyMaterial = await window.crypto.subtle.importKey(
    "raw",
    enc.encode(vaultKey),
    { name: "PBKDF2" },
    false,
    ["deriveKey"]
  );

  return window.crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt: enc.encode("RTK_CACHE_SALT" + userId),
      iterations: 10000,
      hash: "SHA-256",
    },
    keyMaterial,
    { name: "AES-GCM", length: 256 },
    false,
    ["encrypt", "decrypt"]
  );
}

export async function encryptCacheData(data: any, vaultKey: string, userId: string): Promise<string | null> {
  try {
    const key = await getKey(vaultKey, userId);
    const iv = window.crypto.getRandomValues(new Uint8Array(12));
    const encodedData = new TextEncoder().encode(JSON.stringify(data));

    const encryptedContent = await window.crypto.subtle.encrypt(
      { name: "AES-GCM", iv: iv },
      key,
      encodedData
    );

    // Store as JSON string: { iv: [bytes], content: [bytes] }
    const buffer = new Uint8Array(encryptedContent);
    return JSON.stringify({
      iv: Array.from(iv),
      content: Array.from(buffer)
    });
  } catch (e) {
    console.error("Encryption failed", e);
    return null;
  }
}

export async function decryptCacheData(storedString: string, vaultKey: string, userId: string): Promise<any | null> {
  try {
    const { iv, content } = JSON.parse(storedString);
    const key = await getKey(vaultKey, userId);

    const decryptedContent = await window.crypto.subtle.decrypt(
      { name: "AES-GCM", iv: new Uint8Array(iv) },
      key,
      new Uint8Array(content)
    );

    const decoded = new TextDecoder().decode(decryptedContent);
    return JSON.parse(decoded);
  } catch (e) {
    // This happens naturally on first load if vaultKey is wrong/missing
    return null;
  }
}

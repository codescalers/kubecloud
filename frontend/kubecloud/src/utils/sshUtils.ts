/**
 * Converts an ArrayBuffer to a Base64 encoded string using a concise method.
 * @param buffer The ArrayBuffer to encode.
 * @returns The Base64 string.
 */
export const arrayBufferToBase64 = (buffer: ArrayBuffer): string =>
  btoa(String.fromCharCode(...new Uint8Array(buffer)));

/**
 * Converts a Base64 URL string to a Uint8Array, handling padding automatically.
 * @param base64url The Base64 URL string to decode.
 * @returns The decoded Uint8Array.
 */
export const base64urlToUint8Array = (base64url: string): Uint8Array => {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
  const padded = base64.padEnd(base64.length + (4 - base64.length % 4) % 4, '=');
  return Uint8Array.from(atob(padded), c => c.charCodeAt(0));
};

/**
 * Formats a CryptoKey to the OpenSSH public key format, including MPI encoding.
 * @param publicKey The CryptoKey to format.
 * @returns A Promise that resolves to the formatted public key string.
 */
export const formatPublicKey = async (publicKey: CryptoKey): Promise<string> => {
  const jwk = await crypto.subtle.exportKey('jwk', publicKey);
  const e = base64urlToUint8Array(jwk.e!);
  let n = base64urlToUint8Array(jwk.n!);

  // Prepend zero if high bit is set for MPI encoding.
  if (n[0] & 0x80) {
    n = new Uint8Array([0, ...n]);
  }

  // Helper to create a 4-byte big-endian length buffer.
  const encodeLength = (length: number): Uint8Array => {
    const buffer = new ArrayBuffer(4);
    new DataView(buffer).setUint32(0, length, false);
    return new Uint8Array(buffer);
  };

  const sshRsaBytes = new TextEncoder().encode('ssh-rsa');

  // Concatenate all parts into a single buffer.
  const buffer = new Uint8Array([
    ...encodeLength(sshRsaBytes.length), ...sshRsaBytes,
    ...encodeLength(e.length), ...e,
    ...encodeLength(n.length), ...n
  ]);

  return `ssh-rsa ${arrayBufferToBase64(buffer.buffer)}`;
};

/**
 * Formats a PKCS#8 ArrayBuffer private key into PEM format.
 * @param buffer The ArrayBuffer of the private key.
 * @returns The private key in PEM format.
 */
export const formatPrivateKey = (buffer: ArrayBuffer): string => {
  const base64 = arrayBufferToBase64(buffer);
  const chunks = base64.match(/.{1,64}/g) || [];
  return `-----BEGIN PRIVATE KEY-----\n${chunks.join('\n')}\n-----END PRIVATE KEY-----\n`;
};

/**
 * Generates an SSH key pair (public and private key) using the Web Crypto API.
 * The private key is returned in PEM format, and the public key in OpenSSH format.
 * @returns A Promise that resolves to an object containing the formatted public and private keys.
 */
export const generateSSHKey = async (): Promise<{ publicKey: string; privateKey: string }> => {
  // Generate a new 4096-bit RSA key pair for strong security.
  const keyPair = await crypto.subtle.generateKey(
    {
      name: 'RSASSA-PKCS1-v1_5',
      modulusLength: 4096,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: 'SHA-256',
    },
    true,
    ['sign', 'verify']
  );

  // Concurrently format both keys for efficiency.
  const [publicKey, privateKeyBuffer] = await Promise.all([
    formatPublicKey(keyPair.publicKey),
    crypto.subtle.exportKey('pkcs8', keyPair.privateKey)
  ]);

  // Convert the private key buffer to the correct PEM format.
  return { publicKey, privateKey: formatPrivateKey(privateKeyBuffer) };
};

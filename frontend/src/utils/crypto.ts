import CryptoJS from 'crypto-js';

const hasSubtle = typeof window !== 'undefined' && !!(window.crypto && window.crypto.subtle);

function sha256HexFromArrayBuffer(buf: ArrayBuffer): string {
  const wordArray = CryptoJS.lib.WordArray.create(new Uint8Array(buf) as any);
  return CryptoJS.SHA256(wordArray).toString(CryptoJS.enc.Hex);
}

// 计算文件的MD5哈希值
export async function calculateFileHash(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = async (event) => {
      try {
        const arrayBuffer = event.target?.result as ArrayBuffer;
        if (hasSubtle) {
          const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
          const hashArray = Array.from(new Uint8Array(hashBuffer));
          const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
          resolve(hashHex);
        } else {
          const hashHex = sha256HexFromArrayBuffer(arrayBuffer);
          resolve(hashHex);
        }
      } catch (error) {
        reject(error);
      }
    };
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(file);
  });
}

// 计算分片的哈希值
export async function calculateChunkHash(chunk: ArrayBuffer): Promise<string> {
  if (hasSubtle) {
    const hashBuffer = await crypto.subtle.digest('SHA-256', chunk);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return hashHex;
  }
  return sha256HexFromArrayBuffer(chunk);
}

// 生成UUID
export function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

// 将 ArrayBuffer 转换为 Base64 编码字符串
export function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  bytes.forEach((b) => {
    binary += String.fromCharCode(b);
  });
  return btoa(binary);
}

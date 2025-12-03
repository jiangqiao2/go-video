import CryptoJS from 'crypto-js';

const hasSubtle = typeof window !== 'undefined' && !!(window.crypto && window.crypto.subtle);

function sha256HexFromArrayBuffer(buf: ArrayBuffer): string {
  const wordArray = CryptoJS.lib.WordArray.create(new Uint8Array(buf) as any);
  return CryptoJS.SHA256(wordArray).toString(CryptoJS.enc.Hex);
}

// 计算文件的SHA256哈希（分片读取，避免一次性占用大内存导致浏览器 RangeError）
export async function calculateFileHash(file: File): Promise<string> {
  // 对大文件统一用流式 CryptoJS，避免一次 readAsArrayBuffer 触发 "Invalid array length"
  const CHUNK_SIZE = 4 * 1024 * 1024; // 4MB
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    const sha = CryptoJS.algo.SHA256.create();
    let offset = 0;

    const readNext = () => {
      const end = Math.min(offset + CHUNK_SIZE, file.size);
      reader.readAsArrayBuffer(file.slice(offset, end));
    };

    reader.onload = (event) => {
      const arrayBuffer = event.target?.result as ArrayBuffer | null;
      if (arrayBuffer) {
        const wordArray = CryptoJS.lib.WordArray.create(new Uint8Array(arrayBuffer) as any);
        sha.update(wordArray);
        offset += arrayBuffer.byteLength;
      }
      if (offset < file.size) {
        readNext();
      } else {
        const hashHex = sha.finalize().toString(CryptoJS.enc.Hex);
        resolve(hashHex);
      }
    };
    reader.onerror = () => reject(reader.error);

    readNext();
  });
}

// 计算分片的哈希值
export async function calculateChunkHash(chunk: ArrayBuffer): Promise<string> {
  if (hasSubtle) {
    const hashBuffer = await crypto.subtle.digest('SHA-256', chunk);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
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

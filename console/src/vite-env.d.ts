/// <reference types="vite/client" />

declare module 'sm-crypto' {
  export const sm2: {
    doEncrypt(msg: string, publicKey: string, cipherMode?: number): string
    doDecrypt(encryptData: string, privateKey: string, cipherMode?: number): string
    generateKeyPairHex(): { publicKey: string; privateKey: string }
  }
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  /** 版本类型：community（社区版） | enterprise（企业版） */
  readonly VITE_EDITION?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

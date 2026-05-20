/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  /** 版本类型：community（社区版） | enterprise（企业版） */
  readonly VITE_EDITION?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

# Contful

> 🏠 [返回首頁](../README.md) &nbsp;|&nbsp; 🇨🇳 [简体中文](README_zh-CN.md) &nbsp;|&nbsp; 🇭🇰 [繁體中文](README_zh-TW.md) &nbsp;|&nbsp; 🇺🇸 [English](README_en.md) &nbsp;|&nbsp; 🇰🇷 [한국어](README_ko.md) &nbsp;|&nbsp; 🇯🇵 [日本語](README_ja.md)

開源 Headless CMS，支援多站點管理。

## 技術棧

| 層級 | 技術 |
|------|------|
| 後端 | Go 1.25 / Gin / GORM |
| 前端 | Vue 3.5 / TDesign / Vite 8 |
| 資料庫 | PostgreSQL 18 |
| 快取 | Valkey 9 |

## 專案結構

```
contful/
├── admin/            # Admin API 服務（:9080）
├── openapi/          # Open API 服務（:8080）
├── console/          # Vue 3 控制台（:3000）
├── db/               # 資料庫初始化腳本（init_pg.sql：DDL + 種子資料）
├── docker/           # Docker 設定（Dockerfile + docker-compose.yaml）
├── shell/            # 建構腳本
├── build/            # 編譯產物（.gitignore）
├── logs/             # 日誌檔案（.gitignore）
└── uploads/          # 使用者上傳（.gitignore）
```

## 快速開始

### 預設帳號

首次部署後使用以下帳號登入管理後台：

| 欄位 | 值 |
|------|-----|
| 信箱 | `admin@contful.com` |
| 密碼 | `contful@com` |

> ⚠️ **安全提示**：首次登入後請立即修改密碼。

### 前置條件

- PostgreSQL 18
- Valkey 9+
- Go 1.25+
- Node.js 24+

### 方式一：Docker 部署

```bash
# 1. 建構映像檔（在 contful/ 目錄執行）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 2. 編輯設定檔
#    - conf/console.yaml   # Console 服務設定
#    - conf/openapi.yaml   # Open API 服務設定
#    設定檔中已預置預設值，只需修改資料庫密碼等敏感資訊

# 3. 啟動服務
docker-compose -f docker/docker-compose.yaml up -d

# 存取
#   管理後台:  http://localhost         (Console + Admin API)
#   Open API: http://localhost:8080/   (直連)
```

> **提示**：建構命令在 `contful/` 目錄執行，建構上下文為當前目錄。

### 方式二：本地開發

```bash
# 1. 複製環境變數設定
cp .env.example .env

# 2. 啟動資料庫和快取（使用遠端或 Docker 本地）
docker run -d --name contful-postgres -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:18-alpine
docker run -d --name contful-redis -p 6379:6379 redis:7-alpine

# 2. 初始化資料庫
psql -h <host> -U <user> -d contful -f db/init_pg.sql


# 3. 建構
./shell/build.sh

# 4. 啟動服務
./shell/dev.sh start

# 存取
#   管理後台:  http://localhost:3000   (Console + Admin API :9080)
#   Open API: http://localhost:8080/
```

### 方式三：分別啟動

```bash
# 建構
./shell/build.sh console    # 建構 Console（Admin API + 前端）
./shell/build.sh openapi    # 建構 Open API

# 單獨啟動某個服務
./shell/dev.sh logs admin   # 查看 Admin API 日誌
./shell/dev.sh status       # 查看服務狀態
./shell/dev.sh stop         # 停止所有服務
```

## 腳本說明

| 腳本 | 用途 |
|------|------|
| `./shell/build-image.sh` | 建構 Docker 映像檔（用於部署） |
| `./shell/build.sh` | 本地編譯（生成 build/ 目錄產物） |
| `./shell/dev.sh` | 本地開發啟動（編譯 + 執行） |

### 建構參數

```bash
# PostgreSQL 版本（預設）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .


# ARM64 平台
docker build -f docker/Dockerfile.console -t contful/console:pg-arm64 --platform linux/arm64 .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-arm64 --platform linux/arm64 .
```

> **注意**：建構命令在 `contful/` 目錄執行。使用 `--platform` 參數交叉編譯時，`TARGETOS` 和 `TARGETARCH` 會自動適配。

## 服務說明

| 服務 | 連接埠 | 說明 |
|------|--------|------|
| Console | 3000 | Vue 管理後台（開發模式） / 80（Docker） |
| Admin API | 9080 | 管理後台 API |
| Open API | 8080 | 內容 API，可水平擴展 |

## 站點預設設定

新站點建立時會自動寫入以下預設設定（儲存在 `contful_system_config` 表）：

| 設定項 | 預設值 | 說明 |
|--------|--------|------|
| `storage.driver` | `local` | 儲存驅動：`local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | 本地儲存根目錄 |
| `storage.local.base_url` | `/uploads` | 本地儲存存取路徑 |
| `integrity.enabled` | `false` | 是否啟用資料簽章（HMAC-SHA256） |
| `integrity.algorithm` | `HMAC-SHA256` | 簽章演算法 |
| `integrity.signing_key` | _(空) | 簽章密鑰，AES-256-GCM 加密儲存；`integrity.enabled=true` 時自動生成 |

> **提示**：敏感設定（`integrity.signing_key` 等）透過 `CONTFUL_CONFIG_MASTER_KEY` 環境變數加密儲存。
> 生產環境請設定 32 位元組隨機字串作為主密鑰：
> ```bash
> openssl rand -hex 32
> ```

## 文件

- [快速開始](https://contful.com/guide/quickstart)
- [部署指南](https://contful.com/guide/deploy/)
- [系統架構](https://contful.com/guide/architecture/overview)
- [Admin API 文件](https://contful.com/api/admin-api/overview)
- [Open API 文件](https://contful.com/api/open-api/overview)
- [貢獻指南](https://contful.com/about/developers)
- [更新日誌](https://contful.com/guide/changelog)

# Contful

> 🏠 [ルートに戻る](../README.md) &nbsp;|&nbsp; 🇨🇳 [简体中文](README_zh-CN.md) &nbsp;|&nbsp; 🇭🇰 [繁體中文](README_zh-TW.md) &nbsp;|&nbsp; 🇺🇸 [English](README_en.md) &nbsp;|&nbsp; 🇰🇷 [한국어](README_ko.md) &nbsp;|&nbsp; 🇯🇵 [日本語](README_ja.md)

オープンソースのHeadless CMS、マルチサイト管理に対応しています。

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| バックエンド | Go 1.22+ / Gin / GORM |
| フロントエンド | Vue 3.4+ / TDesign |
| データベース | PostgreSQL 16+ |
| キャッシュ | Valkey 9+ |

## プロジェクト構造

```
contful/
├── admin/            # Admin API サービス（:9080）
├── openapi/          # Open API サービス（:8080）
├── console/          # Vue 3 コンソール（:3000）
├── sql/              # データベース初期化 SQL
├── docker/           # Docker 設定（Dockerfile + docker-compose.yaml）
├── shell/            # ビルドスクリプト
├── build/            # ビルド成果物（.gitignore）
├── logs/             # ログファイル（.gitignore）
└── uploads/          # ユーザーアップロード（.gitignore）
```

## クイックスタート

### デフォルトアカウント

初回デプロイ後に以下の認証情報で管理ダッシュボードにログインしてください：

| フィールド | 値 |
|----------|-----|
| メールアドレス | `admin@contful.com` |
| パスワード | `contful@com` |

> ⚠️ **セキュリティ注意**: 初回ログイン後、必ずパスワードを変更してください。

### 前提条件

- PostgreSQL 18
- Valkey 9+
- Go 1.22+
- Node.js 18+

### 方法1：Docker デプロイ

```bash
# 1. イメージをビルド（contful/ ディレクトリで実行）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 2. 設定ファイルを編集
#    - conf/console.yaml   # Console サービス設定
#    - conf/openapi.yaml   # Open API サービス設定
#    デフォルト値が事前設定済みのため、DBパスワードなどの機密情報のみ変更

# 3. サービスを起動
docker-compose -f docker/docker-compose.yaml up -d

# アクセス
#   管理ダッシュボード:  http://localhost         (Console + Admin API)
#   Open API:          http://localhost:8080/   (直接接続)
```

> **補足**: ビルドコマンドは `contful/` ディレクトリで実行し、現在のディレクトリをビルドコンテキストとします。

### 方法2：ローカル開発

```bash
# 1. 環境変数設定をコピー
cp conf/.env.example .env

# 2. データベースとキャッシュを起動（リモートまたはDockerローカルを使用）
docker run -d --name contful-postgres -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:18-alpine
docker run -d --name contful-redis -p 6379:6379 redis:7-alpine

# 2. データベースを初期化
psql -h <host> -U <user> -d contful -f sql/init_pg.sql


# 3. ビルド
./shell/build.sh

# 4. サービスを起動
./shell/dev.sh start

# アクセス
#   管理ダッシュボード:  http://localhost:3000   (Console + Admin API :9080)
#   Open API:          http://localhost:8080/
```

### 方法3：個別に起動

```bash
# ビルド
./shell/build.sh console    # Console をビルド（Admin API + フロントエンド）
./shell/build.sh openapi    # Open API をビルド

# 特定のサービスを起動
./shell/dev.sh logs admin   # Admin API ログを確認
./shell/dev.sh status       # サービス状態を確認
./shell/dev.sh stop         # 全サービスを停止
```

## スクリプト一覧

| スクリプト | 用途 |
|----------|------|
| `./shell/build-image.sh` | Docker イメージをビルド（デプロイ用） |
| `./shell/build.sh` | ローカルコンパイル（build/ ディレクトリに出力） |
| `./shell/dev.sh` | ローカル開発起動（コンパイル + 実行） |

### ビルド引数

```bash
# PostgreSQL バージョン（デフォルト）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .


# ARM64 プラットフォーム
docker build -f docker/Dockerfile.console -t contful/console:pg-arm64 --platform linux/arm64 .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-arm64 --platform linux/arm64 .
```

> **注意**: ビルドコマンドは `contful/` ディレクトリで実行します。`--platform` でクロスコンパイルする場合、`TARGETOS` と `TARGETARCH` は自動的に設定されます。

## サービス一覧

| サービス | ポート | 説明 |
|---------|--------|------|
| Console | 3000 | Vue 管理ダッシュボード（開発モード）/ 80（Docker） |
| Admin API | 9080 | 管理ダッシュボード API |
| Open API | 8080 | コンテンツ API、水平スケール対応 |

## サイトデフォルト設定

新規サイトが作成されると、以下のデフォルト設定が自動的に書き込まれます（`site_configs` テーブルに保存）：

| 設定キー | デフォルト値 | 説明 |
|---------|------------|------|
| `storage.driver` | `local` | ストレージドライバー：`local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | ローカルストレージルートディレクトリ |
| `storage.local.base_url` | `/uploads` | ローカルストレージアクセスパス |
| `integrity.enabled` | `false` | データ完全性署名を有効にするか（HMAC-SHA256） |
| `integrity.algorithm` | `HMAC-SHA256` | 署名アルゴリズム |
| `integrity.signing_key` | _(空) | 署名鍵、AES-256-GCM で暗号化保存；`integrity.enabled=true` の場合自動生成 |

> **ヒント**: 機密設定（`integrity.signing_key` など）は `CONTFUL_CONFIG_MASTER_KEY` 環境変数で暗号化されて保存されます。
> 本番環境では、マスターキーとして32バイトのランダム文字列を設定してください：
> ```bash
> openssl rand -hex 32
> ```

## ドキュメント

- [クイックスタート](https://contful.com/docs/getting-started)
- [デプロイメントガイド](https://contful.com/docs/deployment)
- [アーキテクチャ概要](https://contful.com/docs/architecture/overview)
- [Admin API リファレンス](https://contful.com/docs/api/admin-api/overview)
- [Open API リファレンス](https://contful.com/docs/api/open-api/overview)
- [データベーススキーマ](https://contful.com/docs/database/schema)
- [コントリビューションガイド](https://contful.com/docs/community/contributing)
- [リリースノート](https://contful.com/guide/release)

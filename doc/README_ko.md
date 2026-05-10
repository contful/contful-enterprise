# Contful

> 🏠 [루트로 돌아가기](../README.md) &nbsp;|&nbsp; 🇨🇳 [简体中文](README_zh-CN.md) &nbsp;|&nbsp; 🇭🇰 [繁體中文](README_zh-TW.md) &nbsp;|&nbsp; 🇺🇸 [English](README_en.md) &nbsp;|&nbsp; 🇰🇷 [한국어](README_ko.md) &nbsp;|&nbsp; 🇯🇵 [日本語](README_ja.md)

오픈소스 Headless CMS, 멀티 사이트 관리 지원.

## 기술 스택

| 레이어 | 기술 |
|--------|------|
| 백엔드 | Go 1.22+ / Gin / GORM |
| 프론트엔드 | Vue 3.4+ / TDesign |
| 데이터베이스 | PostgreSQL 16+ |
| 캐시 | Valkey 9+ |

## 프로젝트 구조

```
contful/
├── admin/            # Admin API 서비스 (:9080)
├── openapi/          # Open API 서비스 (:8080)
├── console/          # Vue 3 콘솔 (:3000)
├── sql/              # 데이터베이스 초기화 SQL
├── docker/           # Docker 설정 (Dockerfile + docker-compose.yaml)
├── shell/            # 빌드 스크립트
├── build/            # 빌드 산출물 (.gitignore)
├── logs/             # 로그 파일 (.gitignore)
└── uploads/          # 사용자 업로드 (.gitignore)
```

## 빠른 시작

### 기본 계정

첫 배포 후 다음 자격 증명으로 관리 대시보드에 로그인하세요:

| 필드 | 값 |
|------|-----|
| 이메일 | `admin@contful.com` |
| 비밀번호 | `contful@com` |

> ⚠️ **보안 알림**: 첫 로그인 후 즉시 비밀번호를 변경하세요.

### prerequisites

- PostgreSQL 18
- Valkey 9+
- Go 1.22+
- Node.js 18+

### 방법 1: Docker 배포

```bash
# 1. 이미지 빌드 (contful/ 디렉토리에서 실행)
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 2. 설정 파일 편집
#    - conf/console.yaml   # Console 서비스 설정
#    - conf/openapi.yaml   # Open API 서비스 설정
#    기본값이 사전 설정되어 있으므로 DB 비밀번호 등 민감 정보만 수정

# 3. 서비스 시작
docker-compose -f docker/docker-compose.yaml up -d

# 접속
#   관리 대시보드:  http://localhost         (Console + Admin API)
#   Open API:     http://localhost:8080/   (직접)
```

> **참고**: 빌드 명령은 `contful/` 디렉토리에서 현재 디렉토리를 빌드 컨텍스트로 실행합니다.

### 방법 2: 로컬 개발

```bash
# 1. 환경 변수 설정 복사
cp conf/.env.example .env

# 2. 데이터베이스 및 캐시 시작 (원격 또는 Docker 로컬 사용)
docker run -d --name contful-postgres -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:18-alpine
docker run -d --name contful-redis -p 6379:6379 redis:7-alpine

# 2. 데이터베이스 초기화
psql -h <host> -U <user> -d contful -f sql/init_pg.sql


# 3. 빌드
./shell/build.sh

# 4. 서비스 시작
./shell/dev.sh start

# 접속
#   관리 대시보드:  http://localhost:3000   (Console + Admin API :9080)
#   Open API:     http://localhost:8080/
```

### 방법 3: 개별 서비스 시작

```bash
# 빌드
./shell/build.sh console    # Console 빌드 (Admin API + 프론트엔드)
./shell/build.sh openapi    # Open API 빌드

# 특정 서비스 시작
./shell/dev.sh logs admin   # Admin API 로그 확인
./shell/dev.sh status       # 서비스 상태 확인
./shell/dev.sh stop         # 모든 서비스 중지
```

## 스크립트 참조

| 스크립트 | 용도 |
|----------|------|
| `./shell/build-image.sh` | Docker 이미지 빌드 (배포용) |
| `./shell/build.sh` | 로컬 컴파일 (build/ 디렉토리 산출물) |
| `./shell/dev.sh` | 로컬 개발 시작 (컴파일 + 실행) |

### 빌드 인수

```bash
# PostgreSQL 버전 (기본)
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .


# ARM64 플랫폼
docker build -f docker/Dockerfile.console -t contful/console:pg-arm64 --platform linux/arm64 .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-arm64 --platform linux/arm64 .
```

> **참고**: 빌드 명령은 `contful/` 디렉토리에서 실행합니다. `--platform`으로 크로스 컴파일 시 `TARGETOS`와 `TARGETARCH`가 자동으로 설정됩니다.

## 서비스 참조

| 서비스 | 포트 | 설명 |
|--------|------|------|
| Console | 3000 | Vue 관리 대시보드 (개발 모드) / 80 (Docker) |
| Admin API | 9080 | 관리 대시보드 API |
| Open API | 8080 | 콘텐츠 API, 수평 확장 가능 |

## 사이트 기본 설정

새 사이트가 생성되면 다음 기본 설정이 자동으로 기록됩니다 (`site_configs` 테이블에 저장):

| 설정 키 | 기본값 | 설명 |
|---------|--------|------|
| `storage.driver` | `local` | 스토리지 드라이버: `local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | 로컬 스토리지 루트 디렉토리 |
| `storage.local.base_url` | `/uploads` | 로컬 스토리지 접속 경로 |
| `integrity.enabled` | `false` | 데이터 무결성 서명 활성화 여부 (HMAC-SHA256) |
| `integrity.algorithm` | `HMAC-SHA256` | 서명 알고리즘 |
| `integrity.signing_key` | _(비어있음) | 서명 키, AES-256-GCM 암호화 저장; `integrity.enabled=true`일 때 자동 생성 |

> **팁**: 민감한 설정 (`integrity.signing_key` 등)은 `CONTFUL_CONFIG_MASTER_KEY` 환경 변수로 암호화되어 저장됩니다.
> 프로덕션 환경에서는 마스터 키로 32바이트 무작위 문자열을 설정하세요:
> ```bash
> openssl rand -hex 32
> ```

## 문서

- [시작하기](https://contful.com/docs/getting-started)
- [배포 가이드](https://contful.com/docs/deployment)
- [아키텍처 개요](https://contful.com/docs/architecture/overview)
- [Admin API 레퍼런스](https://contful.com/docs/api/admin-api/overview)
- [Open API 레퍼런스](https://contful.com/docs/api/open-api/overview)
- [데이터베이스 스키마](https://contful.com/docs/database/schema)
- [기여 가이드](https://contful.com/docs/community/contributing)
- [릴리스 노트](https://contful.com/guide/release)

# gpf — Greenfield Port Forwarding

Go와 Bubble Tea로 구축된 빠르고 현대적인 SSH 포트 포워딩 CLI 및 TUI 도구입니다.

![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue)
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

[English](README.md) | **한국어**

## 소개

**gpf** (Greenfield Port Forwarding)은 SSH 포트 포워딩을 쉽게 관리할 수 있는 터미널 네이티브 도구입니다. 간단한 CLI 모드와 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 기반의 대화형 TUI를 모두 지원합니다.

### TUI 미리보기

```
 ┌── gpf — Greenfield Port Forwarding ───────────────────────────┐
 │                                                                │
 │   Active Forwards                                            │
 │   ───────────────────────────────────────────────────────────  │
 │   ┌────────────┬──────────────────┬───────────┬───────┐       │
 │   │ Local      │ Remote           │ State     │       │       │
 │   ├────────────┼──────────────────┼───────────┼───────┤       │
 │   │ :8080      │ server.local:3000│ connected │       │       │
 │   │ :5432      │ db.local:5432    │ connected │       │       │
 │   │ :6379      │ cache.local:6379 │ stopping  │       │       │
 │   └────────────┴──────────────────┴───────────┴───────┘       │
 │                                                                │
 │   [New] [Connect] [Disconnect] [Quit]                         │
 └────────────────────────────────────────────────────────────────┘
```

## 주요 기능

- **대화형 TUI** — 모든 포트 포워딩을 한눈에 관리하는 시각적 대시보드
- **CLI 모드** — CI/CD 및 자동화를 위한 스크립트 실행 가능한 원라인 명령어
- **다중 터널** — 여러 포트를 동시에 다른 호스트로 포워딩
- **영구 세션** — 연결 끊김 시 자동 재연결
- **크로스 플랫폼** — Linux, macOS, Windows 지원
- **의존성 없음** — 단일 정적 링크 바이너리, 런타임 의존성 불필요
- **빠른 시작** — Go로 작성되어 즉시 실행

## 설치

### 방법 1: GitHub Releases (권장)

빌드된 바이너리는 [GitHub Releases](https://github.com/turbobit/gpf/releases)에서 다운로드할 수 있습니다.

플랫폼에 맞는 바이너리를 다운로드하세요:

| 플랫폼 | 바이너리 |
|--------|----------|
| Linux amd64 | `gpf_linux_amd64` |
| Linux arm64 | `gpf_linux_arm64` |
| macOS arm64 | `gpf_darwin_arm64` |
| Windows amd64 | `gpf_windows_amd64.exe` |
| Windows arm64 | `gpf_windows_arm64.exe` |

```bash
# 예시: Linux amd64
VERSION=v0.1.0
curl -LO "https://github.com/turbobit/gpf/releases/download/${VERSION}/gpf_linux_amd64"
chmod +x gpf_linux_amd64
sudo mv gpf_linux_amd64 /usr/local/bin/gpf
```

### 방법 2: Go install

```bash
go install github.com/turbobit/gpf@latest
```

### 방법 3: Unix 설치 스크립트

```bash
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/main/install/unix.sh | sh -s -- v0.1.0
```

최신 버전 설치 (버전 생략 가능):

```bash
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/main/install/unix.sh | sh
```

### 방법 4: Windows PowerShell

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/main/install/windows.ps1" -UseBasicParsing | Invoke-Expression
```

또는 특정 버전:

```powershell
.\install\windows.ps1 v0.1.0
```

## 사용 방법

### CLI 모드

#### 단일 포트 포워딩

```bash
# 로컬 :8080을 원격 :3000으로 포워딩
gpf forward localhost:8080 server:3000

# 커스텀 SSH 키 사용
gpf forward --key ~/.ssh/id_ed25519 localhost:8080 server:3000

# 커스텀 사용자 사용
gpf forward --user admin localhost:8080 server:3000

# 특정 인터페이스에 바인딩
gpf forward --bind 0.0.0.0 localhost:8080 server:3000
```

#### 여러 포트 포워딩

```bash
# 여러 로컬-원격 매핑
gpf forward localhost:8080 web:3000 localhost:5432 db:5432 localhost:6379 cache:6379

# 또는 설정 파일 사용
gpf forward --config forwards.yaml
```

#### 포워딩 관리

```bash
# 활성 포워딩 목록
gpf list

# 특정 포워딩 종료
gpf disconnect 8080

# 모든 포워딩 종료
gpf disconnect --all
```

#### TUI 모드

```bash
# 대화형 터미널 UI 시작
gpf tui
```

### 명령어

```
Usage:
  gpf [command]

Available Commands:
  forward     SSH 포트 포워딩 생성
  disconnect  활성 포워딩 종료
  list        활성 포트 포워딩 목록
  tui         대화형 TUI 시작
  version     버전 정보 출력

Flags:
  -h, --help      Help for gpf
  -v, --version   Version for gpf
```

## 예제

### 개발 워크플로우

```bash
# 애플리케이션, 데이터베이스, 캐시를 한 명령어로 포워딩
gpf forward \
  localhost:8080 app:3000 \
  localhost:5432 db:5432 \
  localhost:6379 redis:6379
```

### SSH 설정 통합

```bash
# 특정 SSH 인증서 사용
gpf forward --key ~/.ssh/deploy_key \
  localhost:8443 staging:443
```

### 일회성 포워딩

```bash
# 연결 후 매핑 출력 및 종료
gpf forward --once localhost:9090 server:80
```

## 설정

gpf는 다음 순서로 설정 파일을 찾습니다:

1. `./gpf.yaml` (현재 디렉토리)
2. `$HOME/.config/gpf/config.yaml`
3. `$HOME/.gpf/config.yaml`

설정 예시:

```yaml
ssh:
  user: deploy
  key: ~/.ssh/id_ed25519
  timeout: 10s

forwards:
  - local: "localhost:8080"
    remote: "production:3000"
  - local: "localhost:5432"
    remote: "production-db:5432"
```

## 소스에서 빌드

```bash
git clone https://github.com/turbobit/gpf.git
cd gpf
go build -o gpf .
```

## 릴리즈

gpf는 [GoReleaser](https://goreleaser.com/)와 GitHub Actions를 사용하여 릴리즈를 자동화합니다. `v*` 태그가 푸시되면 CI가 Linux, macOS, Windows용 바이너리를 자동으로 빌드하여 GitHub Releases에 게시합니다.

### 릴리즈 워크플로우

1. `v*` 태그 푸시 (예: `git tag v0.1.0 && git push origin v0.1.0`)
2. **Release** GitHub Actions 워크플로우 실행
3. GoReleaser가 모든 지원 플랫폼용 크로스 컴파일
4. 변경 로그와 함께 GitHub Release에 바이너리 업로드

### 지원 플랫폼

| OS | 아키텍처 |
|----|---------|
| Linux | amd64, arm64 |
| macOS | arm64 |
| Windows | amd64, arm64 |

## 다국어 (i18n)

gpf는 `i18n/` 디렉토리의 JSON 번역 파일을 통해 다국어를 지원합니다.

### 지원 언어

| 언어 | 파일 |
|------|------|
| English | `i18n/en.json` |
| 한국어 | `i18n/ko.json` |

### 새로운 언어 추가

1. `i18n/en.json`을 `i18n/<locale>.json`으로 복사 (예: `i18n/ja.json`, `i18n/fr.json`)
2. 키는 그대로 유지하고 문자열 값만 번역
3. UTF-8 인코딩의 JSON으로 저장
4. 새로운 파일과 함께 Pull Request 생성

### 번역 기여

커뮤니티 번역을 환영합니다! 기여하려면:

- 리포지토리를 포크하세요
- `i18n/<locale>.json` 파일을 추가하세요
- 언어 및 참고 사항과 함께 PR을 제출하세요

자세한 내용은 `i18n/README.md`를 참조하세요.

## 감사의 말

gpf는 훌륭한 SSH 설정 헬퍼인 [ggh](https://github.com/byawitz/ggh)에서 영감을 받았습니다. 영감을 주신 [@byawitz](https://github.com/byawitz)님께 감사드립니다.

## 라이선스

MIT

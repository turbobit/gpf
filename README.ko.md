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
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/master/install/unix.sh | sh -s -- v0.1.0
```

최신 버전 설치 (버전 생략 가능):

```bash
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/master/install/unix.sh | sh
```

### 방법 4: Windows PowerShell

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/master/install/windows.ps1" -UseBasicParsing | Invoke-Expression
```

또는 특정 버전:

```powershell
.\install\windows.ps1 v0.1.0
```

## 사용 방법

### 빠른 시작

```bash
# ~/.ssh/config의 모든 서버 표시
gpf

# 키워드로 서버 검색 (이름, 호스트, 사용자 부분 일치)
gpf mac
gpf prod
gpf - macbook

# 특정 언어로 실행
gpf --lang ko
gpf -l en mac

# 서버의 리스닝 포트 스캔
gpf ports myserver

# 포트 포워딩 생성
gpf forward myserver 3000        # 원격 :3000 → 자동 할당된 로컬 포트
gpf forward myserver 3000 8080   # 원격 :3000 → 로컬 :8080

# 활성 터널 보기
gpf tunnels

# 터널 종료
gpf stop 12345                   # PID로 종료
gpf stop-all
```

### 명령어

| 명령어 | 설명 |
|--------|------|
| `gpf` | 모든 SSH 서버 표시 (대화형 TUI) |
| `gpf <키워드>` | 서버 검색 (부분 일치, `%키워드%` 형태) |
| `gpf - <키워드>` | 위와 동일 |
| `gpf ports <별칭>` | 서버의 리스닝 포트 스캔 |
| `gpf forward <별칭> <원격-포트> [로컬-포트]` | 포트 포워딩 생성 |
| `gpf tunnels` | 활성 터널 보기 및 관리 |
| `gpf stop <pid>` | PID로 터널 종료 |
| `gpf stop-all` | 모든 터널 종료 |
| `gpf version` | 버전 정보 표시 |
| `--lang <로케일>` | UI 언어 설정 (`en`, `ko`) |
| `-l <로케일>` | `--lang`의 약어 |

### TUI 키보드 단축키

| 키 | 동작 |
|-----|------|
| `↑` / `↓` | 서버 목록 이동 |
| `Enter` | 액션 선택 (포트 포워딩 / SSH) |
| `f` | 선택한 포트 포워딩 |
| `s` | 선택한 서버에 SSH 접속 |
| `k` | 선택한 터널 종료 |
| `Ctrl+U` | 모든 터널 종료 |
| `r` | 터널 목록 새로고침 |
| `/` | 서버 필터링 |
| `Esc` | 뒤로 가기 |
| `q` | 종료 (모든 터널 중지) |

## 예제

### 빠른 포트 포워딩

```bash
# 프로덕션 웹 서버 포워딩
gpf forward prod-web 3000

# 특정 로컬 포트로 포워딩
gpf forward prod-db 5432 5432
```

### 검색 및 연결

```bash
# 이름에 "mac"이 포함된 모든 서버 찾기
gpf mac

# 호스트 또는 사용자명으로 서버 찾기
gpf staging
gpf deploy
```

## 설정

gpf는 기존 `~/.ssh/config` 파일을 읽으므로 별도의 설정 파일이 필요하지 않습니다.

```
Host mac
  HostName 192.168.1.100
  User ubuntu
  Port 22
  IdentityFile ~/.ssh/id_ed25519

Host prod-web
  HostName web.example.com
  User deploy
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

gpf는 다국어를 지원하며, UI 언어는 시스템 로케일을 **자동 감지**하여 적용됩니다.

### 언어 감지 방식

gpf는 다음 우선순위로 언어를 결정합니다:

1. `--lang` / `-l` 플래그 (예: `gpf --lang ko`)
2. `~/.gpf/lang`에 저장된 설정 (마지막 `--lang` 사용 시 자동 저장)
3. `LANG` 환경 변수 (예: `ko_KR.UTF-8` → 한국어)
4. `LANGUAGE`, `LC_ALL`, `LC_MESSAGES`
5. 영어 (기본)

**`--lang`을 한 번 사용하면 설정이 저장되어** 재시작 후에도 자동으로 적용됩니다. 매번 입력할 필요가 없습니다.

### 언어 변경 방법

**방법 1: `--lang` 플래그 (권장)**

```bash
# 한국어
gpf --lang ko
gpf -l ko mac          # 한국어 UI, "mac" 검색
gpf tunnels --lang en  # 영어 UI

# 플래그는 명령어 어디에나 사용 가능
gpf forward prod 3000 --lang ko
```

**방법 2: 환경 변수**

`LANG` 환경 변수를 설정합니다:

```bash
LANG=ko_KR.UTF-8 gpf
LANG=en_US.UTF-8 gpf

# 영구적으로 설정하려면 셸 프로필(~/.bashrc, ~/.zshrc)에 추가
export LANG=ko_KR.UTF-8
```

### 지원 언어

| 언어 | 파일 | 로케일 예시 |
|------|------|-------------|
| English | `i18n/en.json` | `en`, `en_US`, `en_US.UTF-8` |
| 중국어 | `i18n/zh.json` | `zh`, `zh_CN`, `zh_CN.UTF-8` |

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

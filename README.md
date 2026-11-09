# DevOn Articles

Gin으로 만든 독립 실행형 글 커뮤니티 REST API입니다. 회원가입 → 로그인 → 글 발행 → 태그 검색 → 댓글 대화 → 즐겨찾기를 실제 SQLite에 저장합니다. 인증 여부와 작성자 권한, 중첩 URL의 소속 검증, 페이지 수 계산을 코드와 통합 테스트에서 함께 살펴볼 수 있는 학습·포트폴리오용 프로젝트입니다.

DevOn의 GitHub 저장소 분석·채용 공고 매칭·코드 기반 모의 면접을 테스트하는 데모 프로젝트입니다. 저장소 공개 후 활용 순서는 [DevOn 테스트 가이드](docs/devon-testing.md)를 참고하세요.

## 구현 기능

- 회원가입, 이메일 로그인, bcrypt 비밀번호 해시, 24시간 JWT
- 글 작성·상세·목록·수정·삭제, 읽기 쉬운 고유 slug, 태그 중복 정리
- 태그·작성자 동시 필터, limit/offset 페이지, 전체 결과 수
- 댓글 작성·조회·수정·삭제, 댓글 작성자 권한 및 부모 글 검증
- 사용자별 즐겨찾기 추가·해제, DB 복합 기본 키로 중복 방지
- 익명 읽기, 모든 변경 작업의 인증, 구조화한 오류 응답
- 외래 키와 연쇄 삭제, 원자적이고 반복 가능한 시드, 실제 SQLite HTTP 통합 테스트

## 기술 선택과 구조

Go 1.25, Gin 1.11.0, GORM 1.31.1, golang-jwt/jwt v5.3.0, x/crypto bcrypt를 사용합니다. SQLite 드라이버는 `github.com/glebarez/sqlite v1.11.0`으로, C 컴파일러 없이 Windows와 `CGO_ENABLED=0` 환경에서 빌드할 수 있습니다. 직접·간접 의존성 버전은 go.mod와 go.sum에 고정했습니다.

| 위치 | 책임 |
| --- | --- |
| `cmd/server/main.go` | 환경변수, HTTP 서버, 종료 처리 |
| `cmd/manage/main.go` | DB 초기화·시드·개수 확인 |
| `internal/application` | 의존성 조립, HTTP 통합 테스트 |
| `internal/httpapi` | 라우팅, JSON 파싱, JWT 미들웨어, HTTP 상태 |
| `internal/service` | 입력 규칙, 작성자 권한, 응답 모델, slug |
| `internal/repository` | GORM 쿼리, 트랜잭션, 페이지 계산 |
| `internal/model` | User·Article·Tag·Comment·Favorite 관계 |
| `internal/security` | HS256 JWT 발급·검증 |
| `internal/seed` | 학습용 글과 관계 데이터, 반복 시드 테스트 |
| `scripts` | PowerShell 실행·실제 HTTP 시나리오 |
| `docs/verification.md` | 실제 수행한 검증과 미검증 범위 |

요청은 Gin → 선택적 JWT 확인 → 변경 작업의 로그인 확인 → 핸들러의 JSON 파싱 → 서비스의 입력·작성자 검사 → 저장소의 GORM 쿼리 → 응답 모델 순서로 처리됩니다. 저장소는 HTTP 상태 코드를 모르며, 외부 응답에는 비밀번호 해시나 DB 오류를 내보내지 않습니다. 공개 글·댓글의 author에는 username만 포함합니다.

글은 사용자 한 명에게 속하고 여러 태그를 연결합니다. 댓글은 글과 작성자를 참조하며 즐겨찾기는 `(article_id, user_id)`를 기본 키로 사용합니다. 글을 삭제하면 댓글·즐겨찾기·태그 연결이 함께 삭제됩니다. 태그 이름 행은 남을 수 있지만 `/api/tags`에는 현재 글에 연결된 태그만 나옵니다.

## 로컬 설치와 실행

필수 도구는 Go 1.25.7 이상입니다. 별도 DB 서버, Node.js, C 컴파일러가 필요하지 않습니다. 프로젝트 루트에서 실행하세요. `.env`는 서버가 자동으로 읽지 않으며 아래 PowerShell 실행 스크립트가 읽습니다.

```powershell
go mod download
Copy-Item .env.example .env
# .env의 JWT_SECRET을 충분히 긴 무작위 값으로 바꿉니다.
$randomBytes = New-Object byte[] 48
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($randomBytes)
$rng.Dispose()
$secret = [Convert]::ToBase64String($randomBytes)
(Get-Content .env) -replace '^JWT_SECRET=.*$', "JWT_SECRET=$secret" | Set-Content .env
go run ./cmd/manage init
go run ./cmd/manage seed
go run ./cmd/manage seed
go run ./cmd/manage stats
./scripts/run.ps1
```

기본 DB는 프로젝트의 `data/articles.db`에 생성됩니다. 초기화는 GORM AutoMigrate를 사용하며, 서버 시작 시에도 적용합니다. `init`은 기존 데이터를 지우지 않습니다. 시드는 하나의 트랜잭션으로 처리하며 기존 시드 글·계정을 덮어쓰지 않습니다. 사용자 직접 추가 데이터가 있다면 아래 개수보다 많을 수 있습니다.

Unix 셸에서는 환경변수를 직접 지정할 수 있습니다.

```sh
export JWT_SECRET="$(openssl rand -base64 48)"
export DB_PATH=data/articles.db
export HTTP_ADDR=127.0.0.1:8082
go run ./cmd/manage init
go run ./cmd/manage seed
go run ./cmd/server
```

| 변수 | 기본값·규칙 |
| --- | --- |
| `HTTP_ADDR` | `127.0.0.1:8082`, Docker에서는 `0.0.0.0:8082` |
| `DB_PATH` | `data/articles.db`, 상위 디렉터리 자동 생성 |
| `JWT_SECRET` | 필수, 최소 32바이트. 예제 문자열을 실제 환경에서 사용하지 마세요. |
| `GIN_MODE` | 실행 서버 기본 `release`, 테스트 권장 `test` |

`.env`, DB, 실행 파일, coverage, vendor 디렉터리는 Git에서 제외합니다. 비밀번호와 토큰을 실제 계정 정보로 바꾸어 저장할 필요가 없습니다.

## 시드 계정과 시나리오

세 계정은 모두 가상의 example.com 계정이며 비밀번호는 학습용 `DemoPass123!`입니다.

| 사용자 | 이메일 | 역할 |
| --- | --- | --- |
| hana | hana@example.com | 서비스 설계·JWT 글 작성자 |
| min | min@example.com | 통합 테스트·시드 글 작성자 |
| june | june@example.com | 페이지 설계 글 작성자·댓글 참여자 |

빈 DB에서 시드를 두 번 실행해도 **사용자 3, 글 5, 댓글 6, 즐겨찾기 7, 태그 6**입니다. 고정 slug는 `go-service-boundaries`, `sqlite-integration-tests`, `jwt-ownership-checks`, `stable-pagination`, `idempotent-seeds`입니다. 시드 내용은 이 프로젝트의 설계·테스트 주제에 맞춘 원본 한국어 글입니다.

서버를 켠 다음 별도 터미널에서 실제 HTTP 시나리오를 실행하세요.

```powershell
./scripts/smoke.ps1
```

스크립트는 가상 사용자 두 명을 생성하고 로그인, 글 CRUD, 다른 사용자의 수정 차단, 태그·작성자 필터, 댓글 수정·삭제 권한, 즐겨찾기 중복 방지, 잘못된 JWT, 삭제 후 404를 검사합니다. 생성한 글·댓글·즐겨찾기는 정리하고 가상 계정 두 개는 남깁니다. 시드 기준 개수를 보존하려면 별도 `DB_PATH`에서 실행하세요.

## API 계약

인증 헤더는 `Authorization: Bearer <token>`입니다. 읽기 API는 익명 호출할 수 있지만 인증 헤더를 보냈다면 유효해야 합니다. JWT의 HS256 알고리즘, 서명, exp, iss, aud, sub와 현재 사용자 존재 여부를 확인합니다.

| 메서드·경로 | 인증 | 응답 |
| --- | --- | --- |
| `GET /healthz` | 불필요 | DB 연결 확인, 200 |
| `POST /api/users` | 불필요 | 회원가입, 201 `{user}` |
| `POST /api/users/login` | 불필요 | 로그인, 200 `{user}` |
| `GET /api/articles` | 선택 | 200 `{articles, articlesCount}` |
| `GET /api/articles/:slug` | 선택 | 200 `{article}` |
| `POST /api/articles` | 필수 | 201 `{article}` |
| `PUT /api/articles/:slug` | 글 작성자 | 200 `{article}` |
| `DELETE /api/articles/:slug` | 글 작성자 | 204 |
| `GET /api/articles/:slug/comments` | 선택 | 200 `{comments}` |
| `POST /api/articles/:slug/comments` | 필수 | 201 `{comment}` |
| `PUT /api/articles/:slug/comments/:id` | 댓글 작성자 | 200 `{comment}` |
| `DELETE /api/articles/:slug/comments/:id` | 댓글 작성자 | 204 |
| `POST /api/articles/:slug/favorite` | 필수 | 200 `{article}` |
| `DELETE /api/articles/:slug/favorite` | 필수 | 200 `{article}` |
| `GET /api/tags` | 선택 | 200 `{tags}` |

목록은 최신 생성 ID 순입니다. `tag=go&author=hana`처럼 두 조건을 함께 적용합니다. `limit` 기본 20·범위 1~100, `offset` 기본 0·범위 0~1,000,000이며 중복 쿼리 키는 400입니다. `articlesCount`는 필터 후 전체 수로, 페이지 길이와 다릅니다. 빈 배열은 `[]`로 반환합니다.

새 글의 slug는 제목을 정리한 문자열과 암호학적 난수 접미사로 생성합니다. 제목이 같아도 서로 다른 글을 만들며, 제목 수정 시 slug는 유지합니다. 즐겨찾기 추가·해제는 반복 호출해도 결과가 유지됩니다. 댓글 삭제·수정은 URL의 글과 댓글 소속이 다르면 404입니다. 글 작성자에게 타인의 댓글 수정·삭제 권한을 주지 않습니다.

### 요청 예시

PowerShell에서는 JSON을 객체로 만드는 방식이 따옴표 이스케이프를 피하기 쉽습니다.

```powershell
$base = 'http://127.0.0.1:8082'
$loginBody = @{ user = @{ email='hana@example.com'; password='DemoPass123!' } } | ConvertTo-Json
$login = Invoke-RestMethod "$base/api/users/login" -Method Post -ContentType 'application/json' -Body $loginBody
$headers = @{ Authorization = "Bearer $($login.user.token)" }
$body = @{ article = @{ title='My Go article'; description='A short overview'; body='A complete explanation'; tagList=@('go','testing') } } | ConvertTo-Json -Depth 5
$created = Invoke-RestMethod "$base/api/articles" -Method Post -Headers $headers -ContentType 'application/json' -Body $body
$slug = $created.article.slug
Invoke-RestMethod "$base/api/articles?tag=go&author=hana&limit=2&offset=0"
Invoke-RestMethod "$base/api/articles/$slug" -Method Put -Headers $headers -ContentType 'application/json' -Body '{"article":{"title":"Updated title"}}'
$comment = Invoke-RestMethod "$base/api/articles/$slug/comments" -Method Post -Headers $headers -ContentType 'application/json' -Body '{"comment":{"body":"A useful discussion"}}'
$commentId = $comment.comment.id
Invoke-RestMethod "$base/api/articles/$slug/comments/$commentId" -Method Put -Headers $headers -ContentType 'application/json' -Body '{"comment":{"body":"Updated discussion"}}'
Invoke-RestMethod "$base/api/articles/$slug/favorite" -Method Post -Headers $headers
Invoke-RestMethod "$base/api/articles/$slug/favorite" -Method Delete -Headers $headers
Invoke-RestMethod "$base/api/articles/$slug/comments/$commentId" -Method Delete -Headers $headers
Invoke-RestMethod "$base/api/articles/$slug" -Method Delete -Headers $headers
```

Unix curl 예시도 같은 JSON 계약을 사용합니다. `TOKEN`은 로그인 응답의 `user.token`으로 설정합니다.

```sh
curl -i http://127.0.0.1:8082/api/users -H 'Content-Type: application/json' \
  -d '{"user":{"username":"reader","email":"reader@example.com","password":"DemoPass123!"}}'
curl -i http://127.0.0.1:8082/api/users/login -H 'Content-Type: application/json' \
  -d '{"user":{"email":"hana@example.com","password":"DemoPass123!"}}'
curl 'http://127.0.0.1:8082/api/articles?tag=go&author=hana&limit=2&offset=0'
curl -i http://127.0.0.1:8082/api/articles -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"article":{"title":"Go notes","description":"Short notes","body":"Useful notes","tagList":["go"]}}'
```

### 입력·오류 규칙

username은 영문·숫자·밑줄 3~32자이며 대소문자를 구분합니다. 이메일은 소문자로 정규화하고 고유하게 저장합니다. 비밀번호는 UTF-8 8~72바이트입니다. 제목 1~200자, 설명 1~500자, 본문 1~50,000자, 댓글 1~5,000자이며 공백만 있는 값은 거부합니다. 태그는 최대 10개, 소문자 영문·숫자·하이픈 1~32자이고 공백 제거·소문자화·중복 제거를 합니다. PUT은 일부 필드만 바꿀 수 있지만 빈 객체는 거부합니다. 최대 JSON 본문은 1MiB이며 알려지지 않은 JSON 필드, 잘못된 타입, 연속 JSON은 400입니다.

```json
{"error":{"code":"validation_error","message":"입력값을 확인하세요.","fields":{"title":"제목은 1~200자입니다."}}}
```

400은 파싱·쿼리 오류, 401은 누락·잘못된·만료 인증, 403은 작성자 불일치, 404는 없는 리소스·부모 불일치, 409는 이메일·username 중복, 422는 도메인 입력 오류입니다. 내부 오류는 500으로 응답하며 DB 메시지를 공개하지 않습니다.

## 테스트·빌드

```powershell
$env:GIN_MODE = 'test'
go mod verify
go test -count=1 -v ./...
go vet ./...
go build ./...
# 실행 파일이 필요한 경우
New-Item -ItemType Directory -Force bin | Out-Null
go build -o bin/articles.exe ./cmd/server
go build -o bin/manage.exe ./cmd/manage
```

테스트는 `t.TempDir()`의 독립 SQLite 파일과 실제 Gin 라우터를 사용하며 mock DB를 쓰지 않습니다. 인증·만료 JWT·존재하지 않는 사용자·전체 변경 API 보호, 글 CRUD·작성자 격리·동일 제목 slug, 필터와 페이지, 댓글 소속·수정·삭제, 즐겨찾기 중복·해제, 연쇄 삭제, 잘못된 JSON·빈 값, 반복 시드·해시를 검증합니다. GitHub Actions에도 테스트·vet·빌드를 구성했습니다.

현재 자동 테스트는 HTTP 통합 10개와 시드 통합 1개입니다. 실제 실행 증거와 환경 제약은 [검증 기록](docs/verification.md)에 정리했습니다.

## Docker 실행

```sh
# .env에 무작위 JWT_SECRET을 설정한 뒤
docker compose build
docker compose run --rm articles manage init
docker compose run --rm articles manage seed
docker compose up -d
curl http://127.0.0.1:8082/healthz
```

DB는 named volume에 보관합니다. 기본 이미지에는 비루트 사용자를 설정했습니다. 이 작업 환경에는 Docker가 없어 Docker 빌드·컨테이너 실행은 검증하지 않았습니다. 로컬 Windows 실행 결과는 검증 문서에 기록합니다.

## 범위와 한계

완성 범위는 이 문서의 글 커뮤니티 API와 로컬 실행·시드·테스트입니다. 프런트엔드, 프로필 수정·팔로우·피드, 이메일 인증·비밀번호 재설정, 토큰 갱신·개별 로그아웃, 관리자·신고·검색 전문 인덱스는 포함하지 않습니다. 전체 RealWorld 호환성을 주장하지 않습니다. AutoMigrate는 학습용 초기화 방식이며 운영용 버전 마이그레이션 도구는 별도로 도입해야 합니다. SQLite 단일 연결은 데모 동작과 데이터 무결성을 단순하게 검증하기 위한 선택이며 대규모 동시 쓰기 부하를 측정하지 않았습니다. TLS, 인증 요청 제한, 운영 백업·관측도 이 데모 범위 밖입니다.

## 참고와 라이선스

[golang-gin-realworld-example-app](https://github.com/realworld-apps/golang-gin-realworld-example-app)의 글·댓글·태그·즐겨찾기와 JWT 흐름을 참고했습니다(MIT). 참고 프로젝트의 `users`, `articles`, `common` 역할과 Gin/GORM/JWT 구성을 검토한 뒤 이 저장소의 계층·검증·시드는 원본 코드로 작성했으며 소스 파일을 복사하지 않았습니다. 참고 프로젝트를 포크한 저장소는 아닙니다. 이 프로젝트 자체는 [MIT](LICENSE)이며 의존성은 각 패키지 라이선스를 따릅니다.

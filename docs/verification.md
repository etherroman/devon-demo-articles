# Articles 실행 검증 기록

검증일: 2026-11-09 (Asia/Seoul). Windows PowerShell, `go version go1.25.7 windows/amd64`, `CGO_ENABLED=0` 환경에서 확인했습니다. 아래는 설계상 기대치와 실제 실행 결과를 구분한 기록입니다.

## 자동 검증

| 실행 명령 | 실제 결과 |
| --- | --- |
| `go test -count=1 -v ./...` (`GIN_MODE=test`) | **11개 테스트 통과**: application 10개, seed 1개 |
| `go vet ./...` | 종료 코드 0, 진단 없음 |
| `go build ./...` | 종료 코드 0 |
| `go build -o bin/articles.exe ./cmd/server` | 종료 코드 0, 실제 실행 파일 생성 |
| `go build -o bin/manage.exe ./cmd/manage` | 종료 코드 0, 실제 관리 명령 생성 |
| `go mod verify` | `all modules verified` |

최종 전체 테스트 실행에서는 application 2.351초, seed 2.019초로 종료했습니다. root 검토자가 별도 실행한 `CGO_ENABLED=0 go test -count=1 ./...`, `go vet ./...`, `go build ./...`도 종료 코드 0이었습니다. 환경·캐시 상태에 따라 시간은 달라집니다.

HTTP 통합 테스트는 실제 Gin 라우터와 테스트별 임시 SQLite 파일을 연결합니다. 검증 범위는 다음과 같습니다.

- 회원가입·대소문자 이메일 로그인, 중복 username/email 409, 잘못된 로그인 401, 응답 비밀번호 비노출
- 익명 쓰기와 전체 변경 경로 보호, 깨진 JWT·서명 불일치·만료·없는 사용자 거부
- 글 CRUD, 작성자 격리, 제목 변경 시 slug 유지, 동일 제목의 고유 slug
- 태그 정규화, 태그·작성자 동시 필터, 전체 수와 페이지 길이, 빈 페이지
- 댓글 생성·조회·수정·삭제, 댓글 작성자 권한, URL의 글과 댓글 소속 불일치 404
- 즐겨찾기 반복 추가·해제의 멱등성, 사용자별 고유성, 익명·로그인 응답 구분
- 글 삭제 후 댓글·즐겨찾기·태그 연결의 실제 DB 연쇄 삭제
- 빈 값·길이·태그 형식, 잘못된 JSON·알 수 없는 필드·연속 JSON, 쿼리 범위·중복 키
- 시드 두 번 실행 후 동일 개수, 저장된 bcrypt 해시와 샘플 비밀번호 검증

첫 구현 전 테스트 실행에서는 9개 HTTP 시나리오 모두 미구현 경로 404로 실패하는 것을 확인했습니다. 댓글 수정 테스트는 임시 미구현 응답 때문에 기대 403 대신 404로 실패했고, 시드 테스트는 기대 사용자 3명 대신 0명으로 실패한 뒤 구현하여 통과했습니다. 이후 전체 변경 경로 인증 테스트 및 태그 연결 연쇄 삭제 검증을 추가했습니다.

## 빈 DB 초기화와 반복 시드

기존 파일이 없는 `data/verification-20261109.db`를 지정하고 실행했습니다. DB 파일은 Git 제외 대상입니다.

```powershell
$env:DB_PATH = 'data/verification-20261109.db'
./bin/manage.exe init
./bin/manage.exe seed
./bin/manage.exe seed
./bin/manage.exe stats
```

실제 출력:

```text
database initialized
{"articles":5,"comments":6,"favorites":7,"tags":6,"users":3}
{"articles":5,"comments":6,"favorites":7,"tags":6,"users":3}
{"articles":5,"comments":6,"favorites":7,"tags":6,"users":3}
```

## 실제 TCP HTTP 검증

빌드한 `bin/articles.exe`를 `127.0.0.1:8082`에서 위 DB에 연결해 실행했습니다. JWT 비밀값은 프로세스 환경에만 무작위로 설정했습니다. `scripts/smoke.ps1`의 실제 19개 HTTP 요청이 모두 기대 상태 코드로 통과했습니다.

| 흐름 | 확인한 상태·데이터 |
| --- | --- |
| DB health | GET 200 |
| 가상 사용자 두 명 가입·로그인 | POST 201, 201, 200 |
| 글 작성·타인 수정 거부·작성자 수정 | POST 201, PUT 403, PUT 200 |
| 태그+작성자+페이지 조회 | GET 200, 전체 1·해당 slug 일치 |
| 댓글 작성·작성자 수정·익명 목록·타인 삭제 거부 | POST 201, PUT 200, GET 200, DELETE 403 |
| 즐겨찾기 두 번 추가·한 번 해제 | POST 200, POST 200, DELETE 200, 추가마다 count=1 |
| 잘못된 JWT 목록 조회 | GET 401 |
| 댓글·글 삭제 후 상세 조회 | DELETE 204, DELETE 204, GET 404 |

마지막 출력은 `Smoke passed: auth, ownership, CRUD, filter/page, comment, unique favorite, cleanup.`입니다. 스크립트는 만든 글·댓글·즐겨찾기를 삭제하며, 가입한 가상 example.com 사용자 두 명은 남깁니다. 따라서 smoke 이후 사용자 수는 시드 기준보다 2명 많습니다.

이 도구 환경의 sandbox는 localhost 소켓 접근을 제한하여 첫 연결이 차단됐습니다. API 서버와 smoke 프로세스를 승인된 로컬 실행 환경에서 실행한 뒤 위 검증을 완료했습니다. 이는 코드 실패와 구분되는 실행 환경 제약입니다.

## 확인하지 않은 범위

- Docker가 설치되지 않아 Dockerfile·Compose의 실제 이미지 빌드와 컨테이너 실행은 수행하지 않았습니다.
- 최초 게시 커밋에서 [GitHub Actions](https://github.com/etherroman/devon-demo-articles/actions/runs/42632374064)가 성공했습니다. Ubuntu에서 의존성 검증·전체 테스트·vet·build를 실행했습니다.
- 독립 Public GitHub 저장소를 생성하고 push했습니다. API 수집 조건 확인은 [publication.md](publication.md)에 기록합니다. 실제 DevOn 서비스 내부의 수집·분석 실행은 별도입니다.
- 전체 RealWorld 호환성 테스트, 부하·동시 쓰기 테스트, 운영 TLS·요청 제한·배포는 수행하지 않았습니다.
- 프런트엔드와 브라우저 화면은 이 API 데모의 구현 범위가 아닙니다.

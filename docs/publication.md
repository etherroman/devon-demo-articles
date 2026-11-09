# GitHub 게시와 수집 조건 검증

2026-11-09 (Asia/Seoul)에 사용자의 요청으로 테스트 계정 **etherroman** 소유의 독립 Public 저장소를 생성하고 main 브랜치를 push했다. 원본 참고 저장소의 fork나 부모 DevOn 저장소의 이력을 사용하지 않았다.

- 저장소: [etherroman/devon-demo-articles](https://github.com/etherroman/devon-demo-articles)
- 기본 브랜치: main
- 최초 소스 커밋: `f653319ff1e6bf70a926d87fa6b145356c6da38e`
- 비밀번호·토큰·실제 .env·DB·의존성 캐시·빌드 결과는 커밋하지 않았다.

## 최초 게시 직후 API 확인

GitHub REST `GET /repos/etherroman/devon-demo-articles`와 main 브랜치 API를 조회해 로컬 커밋과 원격 커밋의 일치를 확인했다.

| 필드 | 실제 반환값 |
| --- | --- |
| private | false |
| fork | false |
| archived | false |
| language | Go |
| size | 0 (게시 직후 집계값) |

최초 조회에서는 size가 0이었다. 로컬 파일 크기나 Git 객체 크기를 GitHub API의 size 값으로 대신 사용하지 않았다. 시간 단위 집계가 반영된 뒤의 실제 판정은 아래 최종 재조회 결과를 참조한다.

~~~powershell
gh api repos/etherroman/devon-demo-articles --jq '{full_name,private,fork,archived,language,size}'
gh api repos/etherroman/devon-demo-articles --jq '(.private == false) and (.fork == false) and (.archived == false) and (.language != null) and (.size >= 50)'
~~~

코드 실행 검증은 [verification.md](verification.md)를 참조한다. 실제 DevOn 서비스 내부의 수집·분석·매칭·면접 실행은 이 GitHub 게시 작업과 별도다.

## 최종 재조회 결과

확인 시각: **2026-11-10 00:13:27 KST**. 인증 없는 공개 REST API 직접 조회와 인증된 GitHub CLI 조회를 교차 확인했다.

| DevOn 수집 조건 | 실제 API 값 | 판정 |
| --- | --- | --- |
| Public | private=false | 통과 |
| 원본 독립 저장소 | fork=false | 통과 |
| 보관되지 않음 | archived=false | 통과 |
| 주 언어 존재 | Go | 통과 |
| size >= 50 | 44 | 미달 |

**전체 판정: 현재 size 조건 미달로 전체 수집 조건 미충족.** 저장소 생성과 push는 완료했지만, 이 API 스냅샷 기준으로는 DevOn 기본 자동 수집 대상이 아니다.

GitHub의 [공식 OpenAPI 명세](https://github.com/github/rest-api-description/blob/main/descriptions/api.github.com/api.github.com.json)는 size를 KB 단위로 시간마다 계산하며 새 저장소 초기값은 0이라고 설명한다. 첫 push 뒤 약 한 시간 후 실제 값이 반영된 것을 관찰했다. 로컬 바이트 수는 판정에 사용하지 않았다. 위 숫자는 확인 시점의 값이며 향후 Git 이력과 GitHub 재집계에 따라 달라질 수 있다.


## 원격 CI

[GitHub Actions 검증](https://github.com/etherroman/devon-demo-articles/actions/runs/42632374064)이 최초 게시 커밋에서 성공했다. ubuntu-latest에서 의존성 검증, HTTP/시드 통합 테스트, go vet, go build를 실행했다. Docker 빌드 검증과는 별개다.


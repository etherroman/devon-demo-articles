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

최초 조회에서는 size가 0이었다. 이 값이 50 이상으로 확인되기 전에는 전체 DevOn 수집 조건을 충족했다고 판정하지 않는다. 로컬 파일 크기나 Git 객체 크기를 GitHub API의 size 값으로 대신 사용하지 않는다. 최종 재조회 결과는 아래에 추가한다.

~~~powershell
gh api repos/etherroman/devon-demo-articles --jq '{full_name,private,fork,archived,language,size}'
gh api repos/etherroman/devon-demo-articles --jq '(.private == false) and (.fork == false) and (.archived == false) and (.language != null) and (.size >= 50)'
~~~

코드 실행 검증은 [verification.md](verification.md)를 참조한다. 실제 DevOn 서비스 내부의 수집·분석·매칭·면접 실행은 이 GitHub 게시 작업과 별도다.

## 원격 CI

[GitHub Actions 검증](https://github.com/etherroman/devon-demo-articles/actions/runs/42632374064)이 최초 게시 커밋에서 성공했다. ubuntu-latest에서 의존성 검증, HTTP/시드 통합 테스트, go vet, go build를 실행했다. Docker 빌드 검증과는 별개다.


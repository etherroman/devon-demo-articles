package seed

type articleFixture struct {
	Slug, Title, Description, Body, Author string
	Tags                                   []string
}

var articleFixtures = []articleFixture{
	{Slug: "go-service-boundaries", Title: "Go 서비스의 책임을 나누는 방법", Description: "HTTP 처리와 비즈니스 규칙, 저장소의 경계를 실전 예제로 살펴봅니다.", Author: "hana", Tags: []string{"go", "api", "design"}, Body: "핸들러는 JSON을 읽고 HTTP 응답을 만듭니다. 서비스는 작성자 권한과 입력 규칙을 판단합니다. 저장소는 GORM 쿼리를 담당합니다. 예를 들어 글 수정은 서비스를 통과해야 하며, 핸들러에서 작성자 검사를 생략해도 저장소를 바로 호출하지 않는 것이 중요합니다. 경계가 분명하면 테스트에서 실제 SQLite를 연결하면서 규칙을 검증할 수 있습니다."},
	{Slug: "sqlite-integration-tests", Title: "SQLite로 빠른 통합 테스트 만들기", Description: "테스트마다 별도 데이터베이스를 만들어 격리와 재현성을 확보합니다.", Author: "min", Tags: []string{"go", "testing", "sqlite"}, Body: "메모리 DB는 연결마다 데이터가 달라질 수 있습니다. 이 예제는 t.TempDir 아래 파일을 만들고 단일 연결을 사용합니다. 외래 키를 활성화해 글 삭제 시 댓글과 즐겨찾기가 함께 정리되는지도 검증합니다. 실제 HTTP 요청을 보내 상태 코드와 저장된 결과를 확인하면 라우팅, 인증, 쿼리가 함께 올바르게 동작하는지 알 수 있습니다."},
	{Slug: "jwt-ownership-checks", Title: "JWT 인증 다음에 필요한 작성자 검사", Description: "로그인 여부와 특정 리소스 수정 권한은 서로 다른 질문입니다.", Author: "hana", Tags: []string{"go", "security"}, Body: "미들웨어는 서명, 만료, 발급자, 수신자를 확인한 후 현재 사용자를 찾습니다. 서비스는 글의 AuthorID와 현재 사용자 ID를 비교합니다. 댓글 삭제에서는 댓글 ID만 찾지 않고 URL의 글 ID에도 속하는지 확인해야 합니다. 유효한 JWT가 있어도 다른 사람의 글을 수정할 수 없어야 합니다."},
	{Slug: "stable-pagination", Title: "필터와 페이지 수를 함께 설계하기", Description: "태그와 작성자를 동시에 검색하고 전체 결과 수를 제공합니다.", Author: "june", Tags: []string{"api", "design"}, Body: "목록 응답의 articlesCount는 현재 페이지 길이가 아니라 필터에 맞는 전체 글 수입니다. 정렬은 생성 ID 내림차순으로 고정해 같은 시각에 생성된 글도 순서가 안정적입니다. limit은 1부터 100까지이며 offset은 0부터 시작합니다. 태그와 작성자 조건을 함께 적용한 뒤 개수를 계산하고 페이지를 읽습니다."},
	{Slug: "idempotent-seeds", Title: "여러 번 실행해도 중복되지 않는 시드", Description: "고정 식별자와 관계 제약을 이용해 학습 데이터를 안전하게 준비합니다.", Author: "min", Tags: []string{"sqlite", "testing"}, Body: "학습용 시드는 같은 명령을 두 번 실행해도 사용자가 늘어나면 안 됩니다. 이메일과 slug로 기존 데이터를 찾고 즐겨찾기는 복합 기본 키로 중복을 막습니다. 전체 시드를 하나의 트랜잭션으로 처리해 일부만 저장되는 상황을 피합니다. 기존 시드 글의 본문을 수정한 경우 다시 시드를 실행해도 편집 내용은 보존됩니다."},
}
var commentFixtures = []struct{ Slug, Author, Body string }{
	{"go-service-boundaries", "min", "저장소에서 HTTP 상태 코드를 모르게 한 점이 이해하기 쉬워요."},
	{"go-service-boundaries", "june", "권한 검사를 서비스 테스트에서도 확인하면 좋겠습니다."},
	{"sqlite-integration-tests", "hana", "외래 키를 켜지 않으면 삭제 테스트가 놓칠 수 있겠네요."},
	{"jwt-ownership-checks", "min", "다른 글의 댓글 ID를 넣는 경우도 테스트에 추가했습니다."},
	{"stable-pagination", "hana", "전체 수와 페이지 길이의 차이가 명확하네요."},
	{"idempotent-seeds", "june", "시드를 두 번 실행한 뒤 개수를 비교해 보았습니다."},
}
var favoriteFixtures = []struct{ Slug, User string }{
	{"go-service-boundaries", "min"}, {"go-service-boundaries", "june"},
	{"sqlite-integration-tests", "hana"}, {"sqlite-integration-tests", "june"},
	{"jwt-ownership-checks", "min"}, {"stable-pagination", "hana"}, {"idempotent-seeds", "june"},
}

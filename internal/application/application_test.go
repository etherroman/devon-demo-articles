package application_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devon-demo-articles/internal/application"
	"devon-demo-articles/internal/model"
	"devon-demo-articles/internal/repository"
)

const testSecret = "integration-test-secret-at-least-32-characters"

type client struct {
	t       *testing.T
	handler http.Handler
	repo    *repository.Repository
}

func setup(t *testing.T) client {
	t.Helper()
	a, err := application.New(filepath.Join(t.TempDir(), "articles.db"), testSecret)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	})
	return client{t: t, handler: a.Handler, repo: a.Repo}
}

func (c client) call(method, path, token string, body any, want int) map[string]any {
	c.t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
	}
	return c.raw(method, path, token, string(raw), want)
}

func (c client) raw(method, path, token, body string, want int) map[string]any {
	c.t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	if w.Code != want {
		c.t.Fatalf("%s %s: want %d got %d: %s", method, path, want, w.Code, w.Body.String())
	}
	if want == http.StatusNoContent {
		if w.Body.Len() != 0 {
			c.t.Fatal("204 has response body")
		}
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		c.t.Fatalf("invalid JSON: %v: %s", err, w.Body.String())
	}
	if want >= 400 {
		e, ok := result["error"].(map[string]any)
		if !ok || e["code"] == "" || e["message"] == "" {
			c.t.Fatalf("unstructured error: %#v", result)
		}
	}
	return result
}

func (c client) signup(name string) string {
	c.t.Helper()
	r := c.call("POST", "/api/users", "", map[string]any{"user": map[string]any{"username": name, "email": name + "@example.com", "password": "DemoPass123!"}}, 201)
	return r["user"].(map[string]any)["token"].(string)
}

func (c client) article(token, title string, tags ...string) string {
	c.t.Helper()
	r := c.call("POST", "/api/articles", token, map[string]any{"article": map[string]any{"title": title, "description": "A practical article", "body": "A carefully written body.", "tagList": tags}}, 201)
	return r["article"].(map[string]any)["slug"].(string)
}

func TestSignupLoginAndDuplicateIdentity(t *testing.T) {
	c := setup(t)
	token := c.signup("alice")
	if token == "" {
		t.Fatal("missing token")
	}
	r := c.call("POST", "/api/users/login", "", map[string]any{"user": map[string]any{"email": "ALICE@example.com", "password": "DemoPass123!"}}, 200)
	u := r["user"].(map[string]any)
	if u["username"] != "alice" || u["email"] != "alice@example.com" || u["token"] == "" {
		t.Fatalf("wrong user: %#v", u)
	}
	if strings.Contains(fmt.Sprint(r), "DemoPass123") || u["password"] != nil || u["passwordHash"] != nil {
		t.Fatal("password leaked")
	}
	c.call("POST", "/api/users", "", map[string]any{"user": map[string]any{"username": "alice", "email": "other@example.com", "password": "DemoPass123!"}}, 409)
	c.call("POST", "/api/users", "", map[string]any{"user": map[string]any{"username": "other", "email": "alice@example.com", "password": "DemoPass123!"}}, 409)
	c.call("POST", "/api/users/login", "", map[string]any{"user": map[string]any{"email": "alice@example.com", "password": "WrongPass123!"}}, 401)
	c.call("POST", "/api/users/login", "", map[string]any{"user": map[string]any{"email": "absent@example.com", "password": "WrongPass123!"}}, 401)
}

func signJWT(secret string, expires int64, subject string) string {
	enc := base64.RawURLEncoding.EncodeToString
	header := enc([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"sub": subject, "iss": "devon-articles", "aud": "devon-articles-api", "exp": expires, "iat": time.Now().Add(-time.Hour).Unix()})
	payload := header + "." + enc(claims)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(payload))
	return payload + "." + enc(h.Sum(nil))
}

func TestJWTRejectsAnonymousMalformedExpiredAndUnknownUser(t *testing.T) {
	c := setup(t)
	for _, token := range []string{"", "broken", signJWT("wrong-secret", time.Now().Add(time.Hour).Unix(), "1"), signJWT(testSecret, time.Now().Add(-time.Minute).Unix(), "1"), signJWT(testSecret, time.Now().Add(time.Hour).Unix(), "999")} {
		c.call("POST", "/api/articles", token, map[string]any{"article": map[string]any{"title": "Auth", "description": "Auth", "body": "Auth"}}, 401)
	}
	c.call("GET", "/api/articles", "broken", nil, 401)
	c.call("GET", "/api/articles", "", nil, 200)
}

func TestArticleCRUDOwnershipAndStableSlug(t *testing.T) {
	c := setup(t)
	alice := c.signup("alice")
	bob := c.signup("bob")
	slug := c.article(alice, "Go HTTP Patterns", " Go ", "api", "go")
	r := c.call("GET", "/api/articles/"+slug, "", nil, 200)["article"].(map[string]any)
	if r["author"].(map[string]any)["username"] != "alice" || len(r["tagList"].([]any)) != 2 {
		t.Fatalf("wrong article: %#v", r)
	}
	if strings.Contains(fmt.Sprint(r), "alice@example.com") {
		t.Fatal("public author exposes email")
	}
	path := "/api/articles/" + slug
	c.call("PUT", path, bob, map[string]any{"article": map[string]any{"title": "Hijacked"}}, 403)
	c.call("DELETE", path, bob, nil, 403)
	r = c.call("PUT", path, alice, map[string]any{"article": map[string]any{"title": "Revised title", "body": "Revised body", "tagList": []string{"testing"}}}, 200)["article"].(map[string]any)
	if r["slug"] != slug || r["title"] != "Revised title" || r["body"] != "Revised body" || len(r["tagList"].([]any)) != 1 {
		t.Fatalf("update failed: %#v", r)
	}
	c.call("DELETE", path, alice, nil, 204)
	c.call("GET", path, "", nil, 404)
	c.call("DELETE", path, alice, nil, 404)
}

func TestArticleTagAuthorFilterAndPagination(t *testing.T) {
	c := setup(t)
	alice := c.signup("alice")
	bob := c.signup("bob")
	c.article(alice, "First Go", "go", "api")
	second := c.article(alice, "Second Go", "go")
	c.article(bob, "Other Go", "go")
	c.article(bob, "Java article", "java")
	r := c.call("GET", "/api/articles?tag=go&author=alice&limit=1&offset=0", "", nil, 200)
	if r["articlesCount"] != float64(2) || len(r["articles"].([]any)) != 1 || r["articles"].([]any)[0].(map[string]any)["slug"] != second {
		t.Fatalf("bad first page: %#v", r)
	}
	r = c.call("GET", "/api/articles?tag=go&author=alice&limit=1&offset=1", "", nil, 200)
	if r["articlesCount"] != float64(2) || len(r["articles"].([]any)) != 1 || r["articles"].([]any)[0].(map[string]any)["slug"] == second {
		t.Fatalf("bad second page: %#v", r)
	}
	r = c.call("GET", "/api/articles?tag=missing&offset=99", "", nil, 200)
	if r["articlesCount"] != float64(0) || len(r["articles"].([]any)) != 0 {
		t.Fatalf("bad empty page: %#v", r)
	}
	c.call("GET", "/api/tags", "", nil, 200)
}

func TestSameTitleCreatesUniqueSlugs(t *testing.T) {
	c := setup(t)
	token := c.signup("alice")
	one := c.article(token, "Repeated title")
	two := c.article(token, "Repeated title")
	if one == two || !strings.HasPrefix(one, "repeated-title-") {
		t.Fatalf("slugs collide or unreadable: %q %q", one, two)
	}
}

func TestCommentsEnforceAuthorAndParent(t *testing.T) {
	c := setup(t)
	alice := c.signup("alice")
	bob := c.signup("bob")
	slug := c.article(alice, "Commented")
	other := c.article(alice, "Other article")
	r := c.call("POST", "/api/articles/"+slug+"/comments", bob, map[string]any{"comment": map[string]any{"body": "Helpful explanation"}}, 201)
	id := int(r["comment"].(map[string]any)["id"].(float64))
	r = c.call("GET", "/api/articles/"+slug+"/comments", "", nil, 200)
	if len(r["comments"].([]any)) != 1 {
		t.Fatal("comment missing")
	}
	path := fmt.Sprintf("/api/articles/%s/comments/%d", slug, id)
	c.call("PUT", path, alice, map[string]any{"comment": map[string]any{"body": "Hijacked"}}, 403)
	c.call("PUT", fmt.Sprintf("/api/articles/%s/comments/%d", other, id), bob, map[string]any{"comment": map[string]any{"body": "Wrong parent"}}, 404)
	r = c.call("PUT", path, bob, map[string]any{"comment": map[string]any{"body": "Updated explanation"}}, 200)
	if r["comment"].(map[string]any)["body"] != "Updated explanation" {
		t.Fatal("comment was not updated")
	}
	c.call("DELETE", path, alice, nil, 403)
	c.call("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", other, id), bob, nil, 404)
	c.call("DELETE", path, bob, nil, 204)
	c.call("DELETE", path, bob, nil, 404)
	c.call("GET", "/api/articles/missing/comments", "", nil, 404)
	c.call("POST", "/api/articles/missing/comments", bob, map[string]any{"comment": map[string]any{"body": "hello"}}, 404)
}

func TestFavoriteIsUniquePerUserAndDeleteIsIdempotent(t *testing.T) {
	c := setup(t)
	alice := c.signup("alice")
	bob := c.signup("bob")
	slug := c.article(alice, "Favorites")
	path := "/api/articles/" + slug + "/favorite"
	for i := 0; i < 2; i++ {
		r := c.call("POST", path, bob, nil, 200)["article"].(map[string]any)
		if r["favoritesCount"] != float64(1) || r["favorited"] != true {
			t.Fatalf("duplicate favorite: %#v", r)
		}
	}
	r := c.call("GET", "/api/articles/"+slug, "", nil, 200)["article"].(map[string]any)
	if r["favorited"] != false || r["favoritesCount"] != float64(1) {
		t.Fatalf("bad anonymous favorite view: %#v", r)
	}
	r = c.call("POST", path, alice, nil, 200)["article"].(map[string]any)
	if r["favoritesCount"] != float64(2) {
		t.Fatal("favorites must be per user")
	}
	for i := 0; i < 2; i++ {
		r = c.call("DELETE", path, bob, nil, 200)["article"].(map[string]any)
		if r["favoritesCount"] != float64(1) || r["favorited"] != false {
			t.Fatalf("delete failed: %#v", r)
		}
	}
	c.call("POST", "/api/articles/missing/favorite", bob, nil, 404)
}

func TestArticleDeletionCascadesCommentsAndFavorites(t *testing.T) {
	c := setup(t)
	token := c.signup("alice")
	slug := c.article(token, "Cascade", "go", "testing")
	c.call("POST", "/api/articles/"+slug+"/comments", token, map[string]any{"comment": map[string]any{"body": "A comment"}}, 201)
	c.call("POST", "/api/articles/"+slug+"/favorite", token, nil, 200)
	c.call("DELETE", "/api/articles/"+slug, token, nil, 204)
	c.call("GET", "/api/articles/"+slug+"/comments", "", nil, 404)
	c.call("POST", "/api/articles/"+slug+"/favorite", token, nil, 404)
	for _, table := range []any{&model.Comment{}, &model.Favorite{}} {
		var count int64
		if err := c.repo.DB.Model(table).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("orphaned row: count=%d err=%v", count, err)
		}
	}
	var links int64
	if err := c.repo.DB.Table("article_tags").Count(&links).Error; err != nil || links != 0 {
		t.Fatalf("orphaned tags: %d %v", links, err)
	}
}

func TestEveryWriteRequiresAuthentication(t *testing.T) {
	c := setup(t)
	for _, route := range []struct{ method, path string }{{"POST", "/api/articles"}, {"PUT", "/api/articles/some-slug"}, {"DELETE", "/api/articles/some-slug"}, {"POST", "/api/articles/some-slug/comments"}, {"PUT", "/api/articles/some-slug/comments/1"}, {"DELETE", "/api/articles/some-slug/comments/1"}, {"POST", "/api/articles/some-slug/favorite"}, {"DELETE", "/api/articles/some-slug/favorite"}} {
		c.call(route.method, route.path, "", nil, 401)
	}
}

func TestValidationAndMalformedJSON(t *testing.T) {
	c := setup(t)
	for _, user := range []map[string]any{{"username": "x", "email": "bad", "password": "short"}, {"username": "valid", "email": "not-email", "password": "DemoPass123!"}, {"username": "valid", "email": "valid@example.com", "password": strings.Repeat("x", 73)}} {
		c.call("POST", "/api/users", "", map[string]any{"user": user}, 422)
	}
	c.raw("POST", "/api/users", "", `{"user":`, 400)
	c.raw("POST", "/api/users", "", `{"user":{},"surprise":true}`, 400)
	c.raw("POST", "/api/users", "", `{} {}`, 400)
	token := c.signup("alice")
	for _, article := range []map[string]any{{"title": "  ", "description": "Desc", "body": "Body"}, {"title": "Title", "description": "  ", "body": "Body"}, {"title": "Title", "description": "Desc", "body": " "}, {"title": "Title", "description": "Desc", "body": "Body", "tagList": []string{"bad tag"}}} {
		c.call("POST", "/api/articles", token, map[string]any{"article": article}, 422)
	}
	slug := c.article(token, "Validation")
	c.call("PUT", "/api/articles/"+slug, token, map[string]any{"article": map[string]any{}}, 422)
	c.call("PUT", "/api/articles/"+slug, token, map[string]any{"article": map[string]any{"title": " "}}, 422)
	c.call("POST", "/api/articles/"+slug+"/comments", token, map[string]any{"comment": map[string]any{"body": " "}}, 422)
	c.call("DELETE", "/api/articles/"+slug+"/comments/abc", token, nil, 400)
	for _, query := range []string{"limit=0", "limit=101", "limit=nope", "offset=-1", "offset=nope", "limit=1&limit=2"} {
		c.call("GET", "/api/articles?"+query, "", nil, 400)
	}
	c.call("GET", "/api/articles/missing", "", nil, 404)
}

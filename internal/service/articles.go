package service

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"devon-demo-articles/internal/model"
	"devon-demo-articles/internal/repository"
)

type ArticleInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}
type ArticleUpdate struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Body        *string   `json:"body"`
	TagList     *[]string `json:"tagList"`
}
type AuthorView struct {
	Username string `json:"username"`
}
type ArticleView struct {
	Slug           string     `json:"slug"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Body           string     `json:"body"`
	TagList        []string   `json:"tagList"`
	Author         AuthorView `json:"author"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	Favorited      bool       `json:"favorited"`
	FavoritesCount int        `json:"favoritesCount"`
}

func articleView(a model.Article, userID uint) ArticleView {
	tags := make([]string, 0, len(a.Tags))
	for _, t := range a.Tags {
		tags = append(tags, t.Name)
	}
	sort.Strings(tags)
	favorited := false
	for _, f := range a.Favorites {
		if userID != 0 && f.UserID == userID {
			favorited = true
		}
	}
	return ArticleView{Slug: a.Slug, Title: a.Title, Description: a.Description, Body: a.Body, TagList: tags, Author: AuthorView{a.Author.Username}, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, Favorited: favorited, FavoritesCount: len(a.Favorites)}
}

func validText(value string, max int) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(value))
	return n > 0 && n <= max
}

var tagPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) > 10 {
		return nil, Invalid(map[string]string{"tagList": "태그는 최대 10개입니다."})
	}
	seen := map[string]bool{}
	normalized := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if !tagPattern.MatchString(tag) {
			return nil, Invalid(map[string]string{"tagList": "태그는 영문 소문자, 숫자, 하이픈 1~32자입니다."})
		}
		if !seen[tag] {
			seen[tag] = true
			normalized = append(normalized, tag)
		}
	}
	return normalized, nil
}

func newSlug(title string) (string, error) {
	var out strings.Builder
	for _, r := range []rune(strings.ToLower(title)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
		} else {
			out.WriteByte('-')
		}
	}
	parts := strings.FieldsFunc(out.String(), func(r rune) bool { return r == '-' })
	base := []rune(strings.Join(parts, "-"))
	if len(base) > 80 {
		base = base[:80]
	}
	if len(base) == 0 {
		base = []rune("article")
	}
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return strings.TrimRight(string(base), "-") + "-" + hex.EncodeToString(suffix[:]), nil
}

func (s *Service) CreateArticle(userID uint, in ArticleInput) (ArticleView, error) {
	fields := map[string]string{}
	if !validText(in.Title, 200) {
		fields["title"] = "제목은 1~200자입니다."
	}
	if !validText(in.Description, 500) {
		fields["description"] = "설명은 1~500자입니다."
	}
	if !validText(in.Body, 50000) {
		fields["body"] = "본문은 1~50,000자입니다."
	}
	if len(fields) > 0 {
		return ArticleView{}, Invalid(fields)
	}
	tags, err := normalizeTags(in.TagList)
	if err != nil {
		return ArticleView{}, err
	}
	slug, err := newSlug(in.Title)
	if err != nil {
		return ArticleView{}, storageError(err)
	}
	a := model.Article{Slug: slug, Title: strings.TrimSpace(in.Title), Description: strings.TrimSpace(in.Description), Body: strings.TrimSpace(in.Body), AuthorID: userID}
	if err := s.Repo.CreateArticle(&a, tags); err != nil {
		return ArticleView{}, storageError(err)
	}
	return s.Article(slug, userID)
}

func (s *Service) Article(slug string, userID uint) (ArticleView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return ArticleView{}, storageError(err)
	}
	return articleView(a, userID), nil
}

func (s *Service) ListArticles(filter repository.ArticleFilter, userID uint) ([]ArticleView, int64, error) {
	articles, count, err := s.Repo.ListArticles(filter)
	if err != nil {
		return nil, 0, storageError(err)
	}
	views := make([]ArticleView, 0, len(articles))
	for _, article := range articles {
		views = append(views, articleView(article, userID))
	}
	return views, count, nil
}

func (s *Service) UpdateArticle(slug string, userID uint, in ArticleUpdate) (ArticleView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return ArticleView{}, storageError(err)
	}
	if a.AuthorID != userID {
		return ArticleView{}, Forbidden()
	}
	changes := map[string]any{}
	fields := map[string]string{}
	for _, field := range []struct {
		name  string
		value *string
		max   int
	}{{"title", in.Title, 200}, {"description", in.Description, 500}, {"body", in.Body, 50000}} {
		if field.value != nil {
			if !validText(*field.value, field.max) {
				fields[field.name] = "빈 값 또는 최대 길이를 초과한 값입니다."
			} else {
				changes[field.name] = strings.TrimSpace(*field.value)
			}
		}
	}
	if len(changes) == 0 && in.TagList == nil && len(fields) == 0 {
		return ArticleView{}, Invalid(map[string]string{"article": "변경할 필드가 필요합니다."})
	}
	if len(fields) > 0 {
		return ArticleView{}, Invalid(fields)
	}
	if in.TagList != nil {
		tags, err := normalizeTags(*in.TagList)
		if err != nil {
			return ArticleView{}, err
		}
		in.TagList = &tags
	}
	changes["updated_at"] = time.Now().UTC()
	if err := s.Repo.UpdateArticle(&a, changes, in.TagList); err != nil {
		return ArticleView{}, storageError(err)
	}
	return s.Article(slug, userID)
}

func (s *Service) DeleteArticle(slug string, userID uint) error {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return storageError(err)
	}
	if a.AuthorID != userID {
		return Forbidden()
	}
	return storageError(s.Repo.DeleteArticle(&a))
}
func (s *Service) Tags() ([]string, error) {
	tags, err := s.Repo.Tags()
	return tags, storageError(err)
}

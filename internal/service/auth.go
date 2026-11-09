package service

import (
	"net/mail"
	"regexp"
	"strings"

	"devon-demo-articles/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type SignupInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserView struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func validEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email && len(email) <= 254
}

func (s *Service) Signup(in SignupInput) (UserView, error) {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	fields := map[string]string{}
	if !usernamePattern.MatchString(in.Username) {
		fields["username"] = "영문, 숫자, 밑줄 3~32자를 입력하세요."
	}
	if !validEmail(in.Email) {
		fields["email"] = "올바른 이메일을 입력하세요."
	}
	if len(in.Password) < 8 || len(in.Password) > 72 {
		fields["password"] = "비밀번호는 UTF-8 기준 8~72바이트여야 합니다."
	}
	if len(fields) > 0 {
		return UserView{}, Invalid(fields)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserView{}, storageError(err)
	}
	user := model.User{Username: in.Username, Email: in.Email, PasswordHash: string(hash)}
	if err := s.Repo.CreateUser(&user); err != nil {
		return UserView{}, storageError(err)
	}
	return s.userView(user)
}

func (s *Service) Login(in LoginInput) (UserView, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !validEmail(email) || len(in.Password) < 8 || len(in.Password) > 72 {
		return UserView{}, Unauthorized()
	}
	user, err := s.Repo.UserByEmail(email)
	if err != nil {
		return UserView{}, Unauthorized()
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		return UserView{}, Unauthorized()
	}
	return s.userView(user)
}

func (s *Service) userView(user model.User) (UserView, error) {
	token, err := s.Tokens.Issue(user.ID)
	if err != nil {
		return UserView{}, storageError(err)
	}
	return UserView{ID: user.ID, Username: user.Username, Email: user.Email, Token: token}, nil
}

func (s *Service) Authenticate(raw string) (uint, error) {
	id, err := s.Tokens.Parse(raw)
	if err != nil {
		return 0, Unauthorized()
	}
	if _, err := s.Repo.UserByID(id); err != nil {
		return 0, Unauthorized()
	}
	return id, nil
}

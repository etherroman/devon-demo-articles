package service

import "devon-demo-articles/internal/repository"

type Error struct {
	Status        int
	Code, Message string
	Fields        map[string]string
	Cause         error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }
func Invalid(fields map[string]string) error {
	return &Error{Status: 422, Code: "validation_error", Message: "입력값을 확인하세요.", Fields: fields}
}
func Unauthorized() error {
	return &Error{Status: 401, Code: "unauthorized", Message: "유효한 인증이 필요합니다."}
}
func Forbidden() error {
	return &Error{Status: 403, Code: "forbidden", Message: "작성자만 변경하거나 삭제할 수 있습니다."}
}
func Missing(resource string) error {
	return &Error{Status: 404, Code: "not_found", Message: resource + "을(를) 찾을 수 없습니다."}
}
func storageError(err error) error {
	if err == nil {
		return nil
	}
	if repository.IsNotFound(err) {
		return Missing("리소스")
	}
	if repository.IsConflict(err) {
		return &Error{Status: 409, Code: "conflict", Message: "이미 사용 중인 값입니다."}
	}
	return &Error{Status: 500, Code: "internal_error", Message: "요청 처리 중 오류가 발생했습니다.", Cause: err}
}

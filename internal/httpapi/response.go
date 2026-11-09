package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"devon-demo-articles/internal/service"
	"github.com/gin-gonic/gin"
)

func fail(c *gin.Context, err error) {
	var appError *service.Error
	if !errors.As(err, &appError) {
		appError = &service.Error{Status: 500, Code: "internal_error", Message: "요청 처리 중 오류가 발생했습니다.", Cause: err}
	}
	if appError.Status >= 500 {
		log.Printf("request failed: %v (cause: %v)", appError, appError.Cause)
	}
	body := gin.H{"code": appError.Code, "message": appError.Message}
	if len(appError.Fields) > 0 {
		body["fields"] = appError.Fields
	}
	c.AbortWithStatusJSON(appError.Status, gin.H{"error": body})
}

func badRequest(c *gin.Context, message string) {
	fail(c, &service.Error{Status: 400, Code: "bad_request", Message: message})
}

func decode(c *gin.Context, destination any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(destination); err != nil {
		badRequest(c, "올바른 JSON 본문을 보내세요. 최대 1MiB입니다.")
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		badRequest(c, "JSON 값 하나만 보내세요.")
		return false
	}
	return true
}

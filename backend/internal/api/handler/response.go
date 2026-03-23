package handler

import "github.com/gin-gonic/gin"

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type errorEnvelope struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Errors  []fieldError `json:"errors,omitempty"`
}

type fieldError struct {
	Field  string `json:"field"`
	Detail string `json:"detail"`
}

func ok(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, envelope{Success: true, Message: message, Data: data})
}

func fail(c *gin.Context, status int, message string, errs ...fieldError) {
	c.JSON(status, errorEnvelope{Success: false, Message: message, Errors: errs})
}

func fieldErr(field, detail string) fieldError {
	return fieldError{Field: field, Detail: detail}
}

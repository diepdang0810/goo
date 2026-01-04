package response

import (
	"net/http"

	"go1/pkg/apperrors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, err interface{}) {
	c.JSON(status, Response{
		Success: false,
		Error:   err,
	})
}

func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   errors,
	})
}

func HandleBindingError(c *gin.Context, err error) {
	if err.Error() == "EOF" {
		Error(c, http.StatusBadRequest, "Request body cannot be empty")
		return
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errs := make(map[string]string)
		for _, e := range validationErrors {
			errs[e.Field()] = msgForTag(e.Tag())
		}
		ValidationError(c, errs)
		return
	}

	Error(c, http.StatusBadRequest, err.Error())
}

func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Status, Response{
			Success: false,
			Message: appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: err.Error(),
	})
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	}
	return "Invalid value"
}

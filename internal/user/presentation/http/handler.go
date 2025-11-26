package http

import (
	"go1/internal/user/usecase"
	"go1/pkg/response"
	"go1/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(usecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: usecase}
}



func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindingError(c, err)
		return
	}

	input := usecase.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := h.usecase.Create(c.Request.Context(), input); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, gin.H{"message": "User created successfully"})
}

func (h *UserHandler) Fetch(c *gin.Context) {
	users, err := h.usecase.Fetch(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	response.Success(c, ToUserResponses(users))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseInt(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	user, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, ToUserResponse(user))
}



package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/core-go/core"
	"github.com/gin-gonic/gin"

	"go-service/internal/user/model"
	sv "go-service/internal/user/service"
)

type UserHandler struct {
	service  sv.UserService
	Validate func(context.Context, interface{}) ([]core.ErrorMessage, error)
	LogError func(context.Context, string, ...map[string]interface{})
	jsonMap  map[string]int
}

func NewUserHandler(service sv.UserService, validate func(context.Context, interface{}) ([]core.ErrorMessage, error), logError func(context.Context, string, ...map[string]interface{})) *UserHandler {
	userType := reflect.TypeOf(model.User{})
	_, jsonMap, _ := core.BuildMapField(userType)
	return &UserHandler{service: service, Validate: validate, LogError: logError, jsonMap: jsonMap}
}

func (h *UserHandler) All(c *gin.Context) {
	res, err := h.service.All(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) Load(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Id cannot be empty")
		return
	}

	res, err := h.service.Load(c.Request.Context(), id)
	if err != nil {
		h.LogError(c.Request.Context(), fmt.Sprintf("Error to get user %s: %s", id, err.Error()))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, res)
	} else {
		c.JSON(http.StatusOK, res)
	}

}

func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	er1 := c.ShouldBindJSON(&user)

	defer c.Request.Body.Close()
	if er1 != nil {
		c.String(http.StatusInternalServerError, er1.Error())
		return
	}

	errors, er2 := h.Validate(c.Request.Context(), &user)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, errors)
		return
	}

	res, er2 := h.service.Create(c.Request.Context(), &user)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res > 0 {
		c.JSON(http.StatusCreated, user)
	} else {
		c.JSON(http.StatusConflict, res)
	}
}

func (h *UserHandler) Update(c *gin.Context) {
	var user model.User
	er1 := c.BindJSON(&user)
	defer c.Request.Body.Close()

	if er1 != nil {
		c.String(http.StatusInternalServerError, er1.Error())
		return
	}

	id := c.Param("id")
	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Id cannot be empty")
		return
	}

	if len(user.Id) == 0 {
		user.Id = id
	} else if id != user.Id {
		c.String(http.StatusBadRequest, "Id not match")
		return
	}

	errors, er2 := h.Validate(c.Request.Context(), &user)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, errors)
		return
	}

	res, er2 := h.service.Update(c.Request.Context(), &user)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res > 0 {
		c.JSON(http.StatusOK, user)
	} else if res == 0 {
		c.JSON(http.StatusNotFound, res)
	} else {
		c.JSON(http.StatusConflict, res)
	}
}

func (h *UserHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Id cannot be empty")
		return
	}

	r := c.Request
	var user model.User
	body, er0 := core.BuildMapAndStruct(r, &user)
	if er0 != nil {
		h.LogError(c.Request.Context(), er0.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if len(user.Id) == 0 {
		user.Id = id
	} else if id != user.Id {
		c.String(http.StatusBadRequest, "Id not match")
		return
	}

	errors, er2 := h.Validate(c.Request.Context(), &user)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	errors = core.RemoveRequiredError(errors)
	if len(errors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, errors)
		return
	}

	jsonObj, er1 := core.BodyToJsonMap(r, user, body, []string{"id"}, h.jsonMap)
	if er1 != nil {
		h.LogError(c.Request.Context(), er1.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}

	res, er2 := h.service.Patch(r.Context(), jsonObj)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(jsonObj))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res > 0 {
		c.JSON(http.StatusOK, jsonObj)
	} else if res == 0 {
		c.JSON(http.StatusNotFound, res)
	} else {
		c.JSON(http.StatusConflict, res)
	}
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Id cannot be empty")
		return
	}

	res, err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		h.LogError(c.Request.Context(), fmt.Sprintf("Error to delete user %s: %s", id, err.Error()))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res > 0 {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusNotFound, res)
	}
}

package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/core-go/core"
	s "github.com/core-go/search"
	"github.com/gin-gonic/gin"

	"go-service/internal/user/model"
	sv "go-service/internal/user/service"
)

type UserHandler struct {
	service     sv.UserService
	Validate    func(context.Context, interface{}) ([]core.ErrorMessage, error)
	LogError    func(context.Context, string, ...map[string]interface{})
	Map         map[string]int
	ParamIndex  map[string]int
	FilterIndex int
}

func NewUserHandler(service sv.UserService, logError func(context.Context, string, ...map[string]interface{}), validate func(context.Context, interface{}) ([]core.ErrorMessage, error)) *UserHandler {
	_, jsonMap, _ := core.BuildMapField(reflect.TypeOf(model.User{}))
	paramIndex, filterIndex := s.BuildParams(reflect.TypeOf(model.UserFilter{}))
	return &UserHandler{service: service, Validate: validate, LogError: logError, Map: jsonMap, ParamIndex: paramIndex, FilterIndex: filterIndex}
}

func (h *UserHandler) All(c *gin.Context) {
	users, err := h.service.All(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Load(c *gin.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.String(http.StatusBadRequest, "Id cannot be empty")
		return
	}

	user, err := h.service.Load(c.Request.Context(), id)
	if err != nil {
		h.LogError(c.Request.Context(), fmt.Sprintf("Error to get user %s: %s", id, err.Error()))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, user)
	} else {
		c.JSON(http.StatusOK, user)
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

	jsonUser, er1 := core.BodyToJsonMap(r, user, body, []string{"id"}, h.Map)
	if er1 != nil {
		h.LogError(c.Request.Context(), er1.Error(), core.MakeMap(user))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}

	res, er2 := h.service.Patch(r.Context(), jsonUser)
	if er2 != nil {
		h.LogError(c.Request.Context(), er2.Error(), core.MakeMap(jsonUser))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	if res > 0 {
		c.JSON(http.StatusOK, jsonUser)
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

func (h *UserHandler) Search(c *gin.Context) {
	filter := model.UserFilter{Filter: &s.Filter{}}
	err := s.Decode(c.Request, &filter, h.ParamIndex, h.FilterIndex)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	offset := s.GetOffset(filter.Limit, filter.Page)
	users, total, err := h.service.Search(c.Request.Context(), &filter, filter.Limit, offset)
	if err != nil {
		h.LogError(c.Request.Context(), fmt.Sprintf("Error to to search %v: %s", filter, err.Error()))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	c.JSON(http.StatusOK, &s.Result{List: &users, Total: total})
}

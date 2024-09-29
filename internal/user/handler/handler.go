package handler

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/core-go/core"
	g "github.com/core-go/core/handler/gin"
	"github.com/core-go/search"
	"github.com/gin-gonic/gin"

	"go-service/internal/user/model"
	"go-service/internal/user/service"
)

type UserHandler struct {
	service  service.UserService
	Validate core.Validate
	*core.Attributes
	*search.Parameters
}

func NewUserHandler(service service.UserService, logError core.Log, validate core.Validate, action *core.ActionConfig) *UserHandler {
	userType := reflect.TypeOf(model.User{})
	parameters := search.CreateParameters(reflect.TypeOf(model.UserFilter{}), userType)
	attributes := core.CreateAttributes(userType, logError, action)
	return &UserHandler{service: service, Validate: validate, Attributes: attributes, Parameters: parameters}
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
	id, err := g.GetRequiredString(c)
	if err == nil {
		user, err := h.service.Load(c.Request.Context(), id)
		if err != nil {
			h.Error(c.Request.Context(), fmt.Sprintf("Error to get user '%s': %s", id, err.Error()))
			c.String(http.StatusInternalServerError, core.InternalServerError)
			return
		}
		c.JSON(core.IsFound(user), user)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	er1 := g.Decode(c, &user)
	if er1 == nil {
		errors, er2 := h.Validate(c.Request.Context(), &user)
		if !g.HasError(c, errors, er2, h.Error, user, h.Log, h.Resource, h.Action.Create) {
			res, er3 := h.service.Create(c.Request.Context(), &user)
			g.AfterCreated(c, &user, res, er3, h.Error)
		}
	}
}

func (h *UserHandler) Update(c *gin.Context) {
	var user model.User
	er1 := g.DecodeAndCheckId(c, &user, h.Keys, h.Indexes)
	if er1 == nil {
		errors, er2 := h.Validate(c.Request.Context(), &user)
		if !g.HasError(c, errors, er2, h.Error, user, h.Log, h.Resource, h.Action.Update) {
			res, er3 := h.service.Update(c.Request.Context(), &user)
			if er3 != nil {
				h.Error(c.Request.Context(), er2.Error(), core.MakeMap(user))
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
	}
}

func (h *UserHandler) Patch(c *gin.Context) {
	var user model.User
	jsonUser, er1 := g.BuildMapAndCheckId(c, &user, h.Keys, h.Indexes)
	if er1 == nil {
		errors, er2 := h.Validate(c.Request.Context(), &user)
		if !g.HasError(c, errors, er2, h.Error, user, h.Log, h.Resource, h.Action.Update) {
			res, er3 := h.service.Patch(c.Request.Context(), jsonUser)
			g.AfterSaved(c, jsonUser, res, er3, h.Error)
		}
	}
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := g.GetRequiredString(c)
	if err == nil {
		res, err := h.service.Delete(c.Request.Context(), id)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error to delete user '%s': %s", id, err.Error()))
			return
		}
		if res > 0 {
			c.JSON(http.StatusOK, res)
		} else {
			c.JSON(http.StatusNotFound, res)
		}
	}
}

func (h *UserHandler) Search(c *gin.Context) {
	filter := model.UserFilter{Filter: &search.Filter{}}
	err := search.Decode(c.Request, &filter, h.ParamIndex, h.FilterIndex)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	offset := search.GetOffset(filter.Limit, filter.Page)
	users, total, err := h.service.Search(c.Request.Context(), &filter, filter.Limit, offset)
	if err != nil {
		h.Error(c.Request.Context(), fmt.Sprintf("Error to to search %v: %s", filter, err.Error()))
		c.String(http.StatusInternalServerError, core.InternalServerError)
		return
	}
	c.JSON(http.StatusOK, &search.Result{List: &users, Total: total})
}

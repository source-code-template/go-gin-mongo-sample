package handler

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/core-go/core"
	g "github.com/core-go/core/gin"
	"github.com/core-go/search"
	"github.com/gin-gonic/gin"

	"go-service/internal/user/model"
	"go-service/internal/user/service"
)

type UserHandler struct {
	service  service.UserService
	Validate core.Validate[*model.User]
	*core.Attributes
	*search.Parameters
}

func NewUserHandler(service service.UserService, logError core.Log, validate core.Validate[*model.User], action *core.ActionConfig) *UserHandler {
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
	user, er1 := g.Decode[model.User](c)
	if er1 == nil {
		errors, er2 := h.Validate(c.Request.Context(), &user)
		if !g.HasError(c, errors, er2, h.Error, user, h.Log, h.Resource, h.Action.Create) {
			res, er3 := h.service.Create(c.Request.Context(), &user)
			g.AfterCreated(c, &user, res, er3, h.Error)
		}
	}
}

func (h *UserHandler) Update(c *gin.Context) {
	user, er1 := g.DecodeAndCheckId[model.User](c, h.Keys, h.Indexes)
	if er1 == nil {
		errors, er2 := h.Validate(c.Request.Context(), &user)
		if !g.HasError(c, errors, er2, h.Error, user, h.Log, h.Resource, h.Action.Update) {
			res, er3 := h.service.Update(c.Request.Context(), &user)
			g.AfterSaved(c, &user, res, er3, h.Error)
		}
	}
}

func (h *UserHandler) Patch(c *gin.Context) {
	user, jsonUser, er1 := g.BuildMapAndCheckId[model.User](c, h.Keys, h.Indexes)
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
		g.AfterDeleted(c, res, err, h.Error)
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

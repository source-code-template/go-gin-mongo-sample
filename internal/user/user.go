package user

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/core-go/core"
	v "github.com/core-go/core/v10"
	"github.com/gin-gonic/gin"

	"go-service/internal/user/handler"
	"go-service/internal/user/repository/adapter"
	"go-service/internal/user/repository/query"
	"go-service/internal/user/service"
)

type UserTransport interface {
	All(*gin.Context)
	Load(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	Patch(*gin.Context)
	Delete(*gin.Context)
	Search(*gin.Context)
}

func NewUserHandler(db *mongo.Database, logError core.Log, action *core.ActionConfig) (UserTransport, error) {
	validator, err := v.NewValidator()
	if err != nil {
		return nil, err
	}

	userRepository := adapter.NewUserAdapter(db, query.BuildQuery)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService, logError, validator.Validate, action)
	return userHandler, nil
}

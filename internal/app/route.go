package app

import (
	"context"

	"github.com/gin-gonic/gin"
)

func Route(ctx context.Context, g *gin.Engine, cfg Config) error {
	app, err := NewApp(ctx, cfg)
	if err != nil {
		return err
	}

	g.GET("/health", app.Health.Check)

	userPath := g.Group("/users")
	{
		userPath.GET("", app.User.All)
		userPath.GET("/:id", app.User.Load)
		userPath.POST("", app.User.Create)
		userPath.PUT("/:id", app.User.Update)
		userPath.PATCH("/:id", app.User.Patch)
		userPath.DELETE("/:id", app.User.Delete)
		userPath.GET("/search", app.User.Search)
		userPath.POST("/search", app.User.Search)
	}

	return nil
}

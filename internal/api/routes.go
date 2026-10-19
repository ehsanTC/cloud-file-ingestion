package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterEndPoints(r *gin.Engine, services []Registrable) {
	registerHealth(r)

	for _, service := range services {
		service.RegisterRoute(r)
	}
}

func registerHealth(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}

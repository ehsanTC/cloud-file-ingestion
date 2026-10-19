package api

import "github.com/gin-gonic/gin"

type Registrable interface {
	RegisterRoute(r *gin.Engine)
}

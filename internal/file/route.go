package file

import (
	"cloud-file-ingestion/common"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (h *FileService) RegisterRoute(r *gin.Engine) {
	group := r.Group(fmt.Sprintf("/files/%s", common.ApiVersion1))
	{
		group.POST("", h.Upload)
	}
}

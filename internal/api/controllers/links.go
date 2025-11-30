package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	apiModels "link-availability-checker/internal/api/models"
	"link-availability-checker/internal/config"
	"link-availability-checker/internal/services"
	"link-availability-checker/internal/utils/files"
	"link-availability-checker/pkg/filestore"
)

type LinkController struct {
	engine      *gin.Engine
	LinkService services.LinkService
}

func NewLinkController(engine *gin.Engine, ls services.LinkService) *LinkController {
	return &LinkController{
		engine:      engine,
		LinkService: ls,
	}
}

func (ctrl *LinkController) RegisterRoutes() {
	basePath := ctrl.engine.Group(viper.GetString(config.ApiBasePath))
	linkRoutes := basePath.Group("/links")
	{
		linkRoutes.POST("/check", ctrl.CheckLinksInSet)
		linkRoutes.POST("/get_report", ctrl.GetLinkSetAsPDF)
	}
}

func (ctrl *LinkController) CheckLinksInSet(ctx *gin.Context) {
	var req apiModels.CheckLinkSetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.Error{Error: "Invalid request payload"})
		return
	}

	set, err := ctrl.LinkService.CheckLinkSet(&req)
	if err != nil {
		if errors.Is(err, services.ErrServiceStopping) {
			ctx.JSON(http.StatusServiceUnavailable, apiModels.Error{Error: "Service is restarting; task queued, fetch result later"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, apiModels.Error{Error: "Failed to process link set"})
		return
	}

	ctx.JSON(http.StatusOK, apiModels.CheckLinkSetResponse{
		Links:    set.ConvertLinksToStrMap(),
		LinksNum: set.Number,
	})
}

func (ctrl *LinkController) GetLinkSetAsPDF(ctx *gin.Context) {
	var req apiModels.GetLinkSetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.Error{Error: "Invalid request payload"})
		return
	}

	filePath, err := ctrl.LinkService.GetLinkSetAsPDF(ctx.Request.Context(), req.LinksList)
	if err != nil {
		if errors.Is(err, filestore.ErrSetNotFound) {
			ctx.JSON(http.StatusBadRequest, apiModels.Error{Error: "Requested set not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, apiModels.Error{Error: "Failed to generate PDF"})
		}
		return
	}

	ctx.FileAttachment(filePath, "report.pdf")

	files.Delete(filePath)
}

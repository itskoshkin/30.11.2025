package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"link-availability-checker/internal/api/middlewares"
	"link-availability-checker/internal/config"
	"link-availability-checker/internal/utils/signals"
)

type SystemController struct {
	engine *gin.Engine
}

func NewSystemController(e *gin.Engine) *SystemController { return &SystemController{engine: e} }

func (ctrl *SystemController) RegisterRoutes() {
	basePath := ctrl.engine.Group(viper.GetString(config.ApiBasePath))
	systemRoutes := basePath.Group("/system").Use(middlewares.AskPassword())
	{
		systemRoutes.GET("/stop", ctrl.StopService)
		systemRoutes.GET("/restart", ctrl.RestartService)
	}
}

func (ctrl *SystemController) StopService(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Service is stopped"})
	signals.SendInterruptSignal()
}

func (ctrl *SystemController) RestartService(ctx *gin.Context) {
	if err := signals.RestartSelf(); err != nil {
		log.Printf("Failed to restart: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to restart service"})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"message": "Service is restarting..."})
		time.Sleep(100 * time.Millisecond)
		signals.SendInterruptSignal()
	}
}

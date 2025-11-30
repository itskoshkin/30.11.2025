package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"

	"link-availability-checker/internal/api/controllers"
	"link-availability-checker/internal/config"
	"link-availability-checker/internal/logger"
	"link-availability-checker/internal/services"
)

func NewEngine() *gin.Engine {
	if viper.GetBool(config.MuteGinDebug) {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		logger.CustomGinLogger(io.MultiWriter(os.Stdout, logger.GetLogFile())),
		gin.RecoveryWithWriter(io.MultiWriter(os.Stdout, logger.GetLogFile())),
	)
	_ = engine.SetTrustedProxies(nil) // No error expected since no proxy addresses were passed

	return engine
}

func RegisterRoutes(sc *controllers.SystemController, lc *controllers.LinkController) {
	sc.RegisterRoutes()
	lc.RegisterRoutes()
}

func Run(lc fx.Lifecycle, engine *gin.Engine, svc services.LinkService) {
	addr := "0.0.0.0:" + viper.GetString(config.ApiPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Printf("Starting web server on %s... ", addr)
			l, err := net.Listen("tcp", addr)
			if err != nil {
				fmt.Printf("\nFailed to listen on %s: %v", viper.GetString(config.ApiPort), err)
				os.Exit(1)
			}
			fmt.Printf("Done.\n")
			go func() {
				//if err = engine.RunListener(l); err != nil {
				if err = srv.Serve(l); err != nil && !errors.Is(http.ErrServerClosed, err) {
					log.Fatalf("Gin failed to start: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("[FX] Shutdown signal received.")
			//<-queueDone // Was a bad idea
			if err := svc.Shutdown(ctx); err != nil {
				log.Printf("Service shutdown error: %v", err)
			}

			log.Println("[GIN] Stopping web server...")
			srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if err := srv.Shutdown(srvCtx); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					log.Printf("Web server shutdown error: %v", err)
				}
			}
			return nil
		},
	})
}

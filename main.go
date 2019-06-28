package main

import (
	"context"
	"flag"
	"fmt"
	"git.thinkinpower.net/bindb/bdata"
	"git.thinkinpower.net/bindb/middleware"
	"git.thinkinpower.net/bindb/route"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setMode(mode string) {
	switch mode {
	case bdata.RunModeDev:
		gin.SetMode(gin.DebugMode)
	case bdata.RunModeTest:
		gin.SetMode(gin.TestMode)
	case bdata.RunModeRelease:
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	logger.SetFormatter(&logger.TextFormatter{FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logger.InfoLevel)

	port := flag.Int("p", 8080, "-p 8080")
	mode := flag.String("m", "dev", "-m [dev|test|release]")
	dataDir := flag.String("d", "./test", "-d /home/testuser/bindata")
	flag.Parse()

	if *dataDir != "" {
		go func() { bdata.WatchBinDataDir(*dataDir) }()
	}
	bdata.SetBinDatabaseMode(bdata.BinDatabaseModeMemory)

	//启动http服务
	logger.Info("启动http服务...")
	setMode(*mode)
	r := gin.New()
	r.Use(middleware.Log())
	r.Use(middleware.Recovery())
	route.Register(r)

	// Listen and serve on 0.0.0.0:8080
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", *port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			logger.Infof("listen: %s", err.Error())
		}
		logger.Infof("启动http服务成功, port: %d", port)
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown failure.", err)
	}
	logger.Info("Server exit.")
}

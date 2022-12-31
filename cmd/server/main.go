package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	cp "github.com/vongdatcuong/api-gateway/music-streaming/modules/connection_pool"
	"github.com/vongdatcuong/api-gateway/music-streaming/modules/jwtAuth"
	"github.com/vongdatcuong/api-gateway/music-streaming/transport/grpc"
	musicStreaming "github.com/vongdatcuong/api-gateway/music-streaming/transport/grpc"
)

func Run() error {
	jwtAuthService := jwtAuth.NewService(os.Getenv("JWT_SECRET_KEY"), 6*time.Hour)

	// Initiate Connection Pool
	cpInterceptor := cp.NewConnectionPoolInterceptor()
	connectionPool, err := cp.NewConnectionPool(cpInterceptor, os.Getenv("AUTHENTICATION_SERVICE_ADDRESS"))
	defer connectionPool.CloseAll()

	if err != nil {
		return err
	}

	authInterceptor := grpc.NewAuthInterceptor(jwtAuthService, connectionPool)

	musicStreamingHandler := musicStreaming.NewHandler(authInterceptor, connectionPool)
	logrus.Info("API Gateway is initiated")
	if err := musicStreamingHandler.Server(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		logrus.Errorln(err)
	}
}

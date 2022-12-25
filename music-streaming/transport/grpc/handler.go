package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	cp "github.com/vongdatcuong/api-gateway/music-streaming/modules/connection_pool"
	grpcPbV1 "github.com/vongdatcuong/music-streaming-protos/go/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type Handler struct {
	grpcPbV1.UnimplementedPermissionServiceServer
	grpcPbV1.UnimplementedUserServiceServer
	authInterceptor *AuthInterceptor
	connectionPool  *cp.ConnectionPool
}

func NewHandler(authInterceptor *AuthInterceptor, connectionPool *cp.ConnectionPool) *Handler {
	h := &Handler{authInterceptor: authInterceptor, connectionPool: connectionPool}

	return h
}

func (h *Handler) RunRestServer(port string, channel chan error, authenticationAddress string, musicAddress string) {
	// Initiate mux
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true, // Rest Server to return the same fields as protobuf
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	muxCtx, cancelMuxCtx := context.WithCancel(context.Background())
	defer cancelMuxCtx()

	// Register Permission service
	err := grpcPbV1.RegisterPermissionServiceHandlerFromEndpoint(muxCtx, gwmux, authenticationAddress, []grpc.DialOption{grpc.WithInsecure()})

	if err != nil {
		channel <- fmt.Errorf("Failed to register Permission Rest endpoints: %w", err)
		return
	}

	// Register User service
	err = grpcPbV1.RegisterUserServiceHandlerFromEndpoint(muxCtx, gwmux, authenticationAddress, []grpc.DialOption{grpc.WithInsecure()})

	if err != nil {
		channel <- fmt.Errorf("Failed to register User Rest endpoints: %w", err)
		return
	}

	// Register Song service
	err = grpcPbV1.RegisterSongServiceHandlerFromEndpoint(muxCtx, gwmux, musicAddress, []grpc.DialOption{grpc.WithInsecure()})

	if err != nil {
		channel <- fmt.Errorf("Failed to register Song Rest endpoints: %w", err)
		return
	}

	// Register Playlist service
	err = grpcPbV1.RegisterPlaylistServiceHandlerFromEndpoint(muxCtx, gwmux, musicAddress, []grpc.DialOption{grpc.WithInsecure()})

	if err != nil {
		channel <- fmt.Errorf("Failed to register Playlist Rest endpoints: %w", err)
		return
	}

	logrus.Infof("API Gateway is about to Listen on port: %d", port)
	restLis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		channel <- fmt.Errorf("could not listen on port %s: %w", port, err)
		return
	}
	httpMux := http.NewServeMux()
	httpMux.Handle("/", h.authInterceptor.HttpMiddleware(gwmux))

	if err := http.Serve(restLis, httpMux); err != nil {
		channel <- fmt.Errorf("could not serve Rest server on port %s: %w", port, err)
	}
}

func (h *Handler) Server() error {
	restChannel := make(chan error)
	go h.RunRestServer(os.Getenv("REST_PORT"), restChannel, os.Getenv("AUTHENTICATION_SERVICE_ADDRESS"), os.Getenv("MUSIC_SERVICE_ADDRESS"))

	select {
	case restError := <-restChannel:
		return restError
	}
}

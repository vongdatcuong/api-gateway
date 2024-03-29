package grpc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cp "github.com/vongdatcuong/api-gateway/music-streaming/modules/connection_pool"
	"github.com/vongdatcuong/api-gateway/music-streaming/modules/jwtAuth"
	grpcPbV1 "github.com/vongdatcuong/music-streaming-protos/go/v1"
)

type AuthInterceptor struct {
	jwtService     *jwtAuth.JwtService
	connectionPool *cp.ConnectionPool
}

func NewAuthInterceptor(jwtService *jwtAuth.JwtService, connectionPool *cp.ConnectionPool) *AuthInterceptor {
	return &AuthInterceptor{jwtService: jwtService, connectionPool: connectionPool}
}

func (interceptor *AuthInterceptor) HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, errCode := interceptor.authorize(r.Context(), r.Header["Authorization"], r.URL.Path, HttpEndPointNoAuthentication)

		if err != nil {
			if errCode == 1 {
				sendErrorResponse(w, http.StatusInternalServerError, errCode, err.Error())
			} else {
				sendErrorResponse(w, int(errCode), errCode, err.Error())
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, authHeader []string, path string, noAuthenMap map[string]bool) (error, uint32) {
	if noAuthenMap[path] {
		return nil, 0
	}

	accessToken, err := parseAuthorizationHeader(authHeader)

	if err != nil {
		return err, uint32(ERROR_CODE_UNAUTHORIZED)
	}

	claims, err := interceptor.jwtService.ValidateToken(accessToken)

	if err != nil {
		return err, uint32(ERROR_CODE_UNAUTHORIZED)
	}

	res, err := interceptor.connectionPool.UserClient.Authenticate(ctx, &grpcPbV1.AuthenticateRequest{UserId: claims.UserID})

	if err != nil {
		return err, uint32(ERROR_CODE_UNAUTHORIZED)
	}

	if res == nil || res.IsAuthenticated == nil || !*res.IsAuthenticated {
		return fmt.Errorf("invalid token"), uint32(ERROR_CODE_UNAUTHORIZED)
	}

	return nil, 0
}

func parseAuthorizationHeader(values []string) (string, error) {
	if values == nil || len(values) == 0 {
		return "", fmt.Errorf("invalid authorization header")
	}
	authHeader := values[0]
	authHeaderParts := strings.Split(authHeader, " ")

	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header")
	}

	return authHeaderParts[1], nil
}

/*TODO Next Step
1. Find way to pass accessToken from incoming interceptor to grpc handler -> Every grpc.Dial to other services will have to pass along this token
2. Improve the connection pool interceptor above*/

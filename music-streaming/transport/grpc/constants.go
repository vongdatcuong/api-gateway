package grpc

// Http
const httpPath = "/api/gateway/v1"
const httpPermissionPath = httpPath + "/permission"
const httpUserPath = httpPath + "/user"
const httpAuthPath = httpPath + "/auth"

var HttpEndPointNoAuthentication map[string]bool = map[string]bool{
	httpAuthPath + "/login":       true,
	httpUserPath + "/create_user": true,
}

start:
	docker-compose up --build
start-service:
	docker-compose build service && docker-compose up service
down:
	docker-compose down -v
remove:
	docker-compose rm -fsv
prune:
	docker image prune -f
gen-protos:
	protoc --go_out=. --grpc-gateway_out=. --go-grpc_out=. protos/**/*.proto 
export_go_path:
	export GO_PATH=~/go && export PATH=$PATH:/$GO_PATH/bin
export_private_github:
	export GOPRIVATE="github.com/vongdatcuong/*"
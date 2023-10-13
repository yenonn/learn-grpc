gen:
	protoc --proto_path=proto \
		--go_opt=paths=source_relative --go_out=pb \
		--go-grpc_opt=paths=source_relative --go-grpc_out=pb \
		proto/*.proto

clean:
	rm pb/*.go

fieldalignment:
	fieldalignment -fix github.com/yenonn/pcbook/pb

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

test:
		go test -cover -race ./...

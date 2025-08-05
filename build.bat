protoc --proto_path=. --go_opt=paths=source_relative --go_opt=Mexplore-service.proto=github.com/jackdelahunt/protoexplore/explore  --go_out=explore --go-grpc_opt=paths=source_relative --go-grpc_opt=Mexplore-service.proto=github.com/jackdelahunt/protoexplore/explore --go-grpc_out=explore explore-service.proto

go build -o bin/ github.com/jackdelahunt/protoexplore/server
go build -o bin/ github.com/jackdelahunt/protoexplore/cli

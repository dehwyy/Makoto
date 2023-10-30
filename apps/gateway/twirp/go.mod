module github.com/dehwyy/makoto/apps/gateway/twirp

go 1.21.2

require (
	github.com/dehwyy/makoto/libs/config v0.0.0-00010101000000-000000000000
	github.com/dehwyy/makoto/libs/grpc v0.0.0-00010101000000-000000000000
	github.com/dehwyy/makoto/libs/logger v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/golang/protobuf v1.5.3
	github.com/twitchtv/twirp v8.1.3+incompatible
)

require (
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace (
	github.com/dehwyy/makoto/libs/config => ../../../libs/config
	github.com/dehwyy/makoto/libs/database => ../../../libs/database
	github.com/dehwyy/makoto/libs/grpc => ../../../libs/grpc
	github.com/dehwyy/makoto/libs/logger => ../../../libs/logger
	github.com/dehwyy/makoto/libs/middleware => ../../../libs/middleware
)

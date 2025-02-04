package main

import (
	"net/http"
	"strings"

	twirp "github.com/dehwyy/makoto/apps/gateway/twirp/internal/impl"
	"github.com/dehwyy/makoto/apps/gateway/twirp/internal/middleware"
	makoto_config "github.com/dehwyy/makoto/libs/config"
	"github.com/dehwyy/makoto/libs/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

func main() {
	log := logger.New()
	config := makoto_config.New()

	rds := redis.NewClient(&redis.Options{
		Addr: config.GatewayRedisUrl,
		DB:   0,
	})

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//  middleware that reads the `Authorization` header (as twirp doesn't give access to it directly)
	md_with_authorization_header := middleware.NewMiddleware_WithAuthorizationHeader()
	// md_with_authorization := middleware.NewMiddleware_WithAuthorization(config.AuthUrl, log)
	md_only_with_authorization := middleware.NewMiddleware_OnlyAuthorized(config.AuthUrl, rds, log)

	// services
	authorization_service := twirp.NewAuthorizationService(config.AuthUrl, config.UserUrl, twirp.TwirpAuthorizationService{
		ReadHeader:             md_with_authorization_header.Read,
		SetAuthorizationHeader: md_with_authorization_header.SetAuthorizationHeader,
	})
	hashmap_service := twirp.NewHashmapService(config.HashmapUrl, twirp.TwirpHashmapService{
		ReadAuthorizationData: md_only_with_authorization.Read,
	})

	user_service := twirp.NewUserService(config.UserUrl, twirp.TwirpUserService{
		ReadAuthorizationData: md_only_with_authorization.Read,
	})

	// mount
	r.Mount(authorization_service.PathPrefix(), md_with_authorization_header.Middleware(authorization_service))
	r.Mount(hashmap_service.PathPrefix(), md_only_with_authorization.Middleware(hashmap_service))
	r.Mount(user_service.PathPrefix(), md_only_with_authorization.Middleware(user_service))

	// as TwirpGatewayUrl looks like `http://{host}:{port}/*`
	port := ":" + strings.Split(config.TwirpGatewayUrl, ":")[2]

	log.Infof("Gateway server started on port %s", port)
	log.Errorf("server shutdown, %v", http.ListenAndServe(port, r))
}

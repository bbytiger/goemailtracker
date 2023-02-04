package emailtracker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// meant to be an internal service run behind webserver like NGINX
func (tracker *EmailTracker) RunDefaultServer(
	baseUrl string,
	port int,
	appendPath string,
	servePath string,
	redisAddr string,
) {
	// setup encoder
	rsaEncoder := NewRSAEncoder()
	rsaEncoder.GeneratePrivateKey()

	// setup connector
	defaultCtx := context.Background()

	redisConnector := NewRedisConnector(defaultCtx, &redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// setup logger
	logger := DefaultLogger{}

	tr := NewEmailTracker(
		baseUrl,
		appendPath,
		servePath,
		*rsaEncoder,
		*redisConnector,
		logger,
	)
	http.ListenAndServe(fmt.Sprintf(":%d", port), tr.GetQuickstartServer())
}

func (tracker *EmailTracker) GetQuickstartServer() *mux.Router {
	// sample implementation for example of tracking server
	r := mux.NewRouter()
	r.HandleFunc(tracker.ActionToURLPath[ServePixel], tracker.ServePixelHandler)
	r.HandleFunc(tracker.ActionToURLPath[AppendPixel], tracker.ServeAppendHandler)
	return r
}

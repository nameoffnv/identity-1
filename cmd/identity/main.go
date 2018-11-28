package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/endpass/identity"
)

const (
	shutdownTimeout = 5 * time.Second
)

var config = map[string]string{
	"IDENTITY_KEYSTORE_PATH": "",
	"IDENTITY_HTTP_HOST":     ":8080",
}

func envConfig() map[string]string {
	for k := range config {
		v, ok := os.LookupEnv(k)
		if ok {
			config[k] = v
		}
	}
	return config
}

func main() {
	envConfig()

	keystores, err := identity.LoadKeystores(config["IDENTITY_KEYSTORE_PATH"])
	if err != nil {
		log.Fatal(err)
	}

	service := identity.New()
	service.SetKeystores(keystores)

	server := http.Server{
		Addr:    config["IDENTITY_HTTP_HOST"],
		Handler: service.Router(),
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Println("shutting down server")

		server.Shutdown(ctx)
	}()

	log.Print("start web server on ", config["IDENTITY_HTTP_HOST"])

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

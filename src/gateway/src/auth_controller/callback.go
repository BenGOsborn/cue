package auth_controller

import (
	"log"
	"net/http"
)

func Callback(logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {}
}

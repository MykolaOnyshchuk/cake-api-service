package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"github.com/gorilla/mux"
)

func getCakeHandler(w http.ResponseWriter, r *http.Request, u User) {
	w.Write([]byte(u.FavoriteCake))
}

func wrapJwt(
	jwt *JWTService,
	f func(http.ResponseWriter, *http.Request, *JWTService),
) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(rw, r, jwt)
	}
}

type ProtectedHandler func(rw http.ResponseWriter, r *http.Request, u User)

func main() {
	r := mux.NewRouter()
	users := NewInMemoryUserStorage()
	userService := UserService{repository: users}
	jwtService, err := NewJWTService("pubkey.rsa", "privkey.rsa")
	if err != nil {
		panic(err)
	}
	r.HandleFunc("/cake", logRequest(jwtService.jwtAuth(users, getCakeHandler))).
	Methods(http.MethodGet)

	r.HandleFunc("/user/register", logRequest(userService.
	Register)).
	Methods(http.MethodPost)
	r.HandleFunc("/user/jwt", logRequest(wrapJwt(jwtService,
	userService.JWT))).
	Methods(http.MethodPost)
	r.HandleFunc("/user/favorite_cake", logRequest(jwtService.AuthenticationJWT(users, userService.UpdateCake))).
		Methods(http.MethodPut)
	r.HandleFunc("/user/email", logRequest(jwtService.AuthenticationJWT(users, userService.UpdateEmail))).
		Methods(http.MethodPut)
	r.HandleFunc("/user/password", logRequest(jwtService.AuthenticationJWT(users, userService.UpdatePassword))).
		Methods(http.MethodPut)
	r.HandleFunc("/user/me", logRequest(jwtService.AuthenticationJWT(users, userService.GetMe)))


	srv := http.Server{
		Addr: ":8080",
		Handler: r,
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()
	log.Println("Server started, hit Ctrl+C to stop")
	err = srv.ListenAndServe()
	if err != nil {
		log.Println("Server exited with error:", err)
	}
	log.Println("Good bye :)")
}

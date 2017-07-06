package main

//go:generate sqlboiler --wipe postgres

import (
	"flag"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"database/sql"

	"github.com/irth/google-token-verifier"
	"github.com/rs/cors"
)

type App struct {
	DB       *sql.DB
	Verifier *googleVerifier.Verifier
	Hub      *Hub
}

func main() {
	address := flag.String("addr", ":3000", "address:port to listen on")
	flag.Parse()

	db, err := sql.Open("postgres", "password=mysecretpassword sslmode=disable user=postgres")

	if err != nil {
		panic(err)
	}

	app := App{
		db,
		&googleVerifier.Verifier{ClientID: "41009918331-5jiap87h9iaaag4qi597siluelvq3706.apps.googleusercontent.com"},
		newHub(),
	}

	go app.Hub.Run()

	app.Verifier.FetchKeys()

	mux := http.NewServeMux()

	app.registerAuthHandlers(mux)
	app.registerProfileHandlers(mux)
	app.registerWebsocketHandlers(mux)

	handler := cors.Default().Handler(mux)

	log.Print("Listening on ", *address)
	log.Fatal(http.ListenAndServe(*address, handler))
}

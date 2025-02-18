package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type JSONResponse struct {
	Status  int
	Message string
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteJSONError(w http.ResponseWriter, code int, message string) {
	resp := &JSONResponse{Status: code, Message: message}

	json, err := json.Marshal(resp)

	if err != nil {
		io.WriteString(w, "Something went wrong.")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	io.Writer.Write(w, json)
}

func getConnectionPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	check(err)

	return pool
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://github.com/itswilliboy/images-go", http.StatusPermanentRedirect)
}

func getID() (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	id, err := gonanoid.Generate(chars, 10)

	if err != nil {
		return "", err
	}

	return id, nil
}

var Pool *pgxpool.Pool

func main() {
	Pool = getConnectionPool()
	defer Pool.Close()

	http.HandleFunc("/", Index)
	http.HandleFunc("/upload", Upload)
	http.HandleFunc("/{id}", Get)

	log.Println("Listening and serving on port 3000")
	err := http.ListenAndServe(":3000", nil)

	check(err)
}

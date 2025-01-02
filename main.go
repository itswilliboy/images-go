package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteJSONError(w http.ResponseWriter, code int, message string) {
	json := `{"status": %d, "message": "%s"}`
	formatted := fmt.Sprintf(json, code, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	io.WriteString(w, formatted)
}

func getConnection() (conn *pgx.Conn) {
	connString, err := pgx.ParseConnectionString(os.Getenv("DATABASE_URL"))
	check(err)

	conn, err = pgx.Connect(connString)
	check(err)

	return
}

func Ping(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world.")
}

func getID() (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	id, err := gonanoid.Generate(chars, 5)

	if err != nil {
		return "", err
	}

	return id, nil
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	r.ParseMultipartForm(100 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file:")
		fmt.Println(err)
		return
	}
	defer file.Close()

	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size: %+v\n", handler.Size)
	log.Printf("MIME Header: %+v\n", handler.Header)

	data, err := io.ReadAll(file)
	check(err)

	os.WriteFile("./tmp/test.png", data, 0644)

	w.Header().Set("Content-Type", "application/json")
	id, err := getID()
	check(err)

	json := fmt.Sprintf(`{"id": "%s"}`, id)
	io.WriteString(w, json)
}

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		WriteJSONError(w, http.StatusMethodNotAllowed, "Method not alowed.")
		return
	}

	id := r.PathValue("id")
	split := strings.Split(id, ".")

	var imageData []byte
	var mimetype string
	if err := Conn.QueryRow("SELECT image_data, mimetype FROM images WHERE id = $1", split[0]).Scan(&imageData, &mimetype); err != nil {
		WriteJSONError(w, http.StatusNotFound, "Not found.")
		return
	}

	w.Header().Set("Content-Type", mimetype)
	w.Write(imageData)
	w.Header().Set("Content-Type", "image/png")
}

var Conn *pgx.Conn

func main() {
	Conn = getConnection()
	defer Conn.Close()

	http.HandleFunc("/", Ping)
	http.HandleFunc("/upload", Upload)
	http.HandleFunc("/{id}", Get)

	log.Println("Listening and serving on port 3000")
	err := http.ListenAndServe(":3000", nil)

	check(err)
}

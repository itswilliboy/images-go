package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
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
	id, err := gonanoid.Generate(chars, 10)

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
	if r.Header.Get("Authorisation") != os.Getenv("AUTH") {
		WriteJSONError(w, http.StatusUnauthorized, "Unauthorised.")
		return
	}

	r.ParseMultipartForm(100 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error while retrieving file: %v\n", err)
		return
	}
	defer file.Close()

	id, err := getID()
	if err != nil {
		log.Printf("Error while creating ID: %v\n", err)
		WriteJSONError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	mimetype := mimetype.Detect(data)
	_, err = Conn.Exec("INSERT INTO images (id, image_data, mimetype) VALUES ($1, $2, $3)", id, data, mimetype.String())
	if err != nil {
		log.Printf("Database error: %s\n", err)
		WriteJSONError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	io.WriteString(w, fmt.Sprintf(`{"url": "https://i.itswilli.dev/%s%s"}`, id, mimetype.Extension()))

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

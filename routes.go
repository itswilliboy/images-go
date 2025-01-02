package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/net/context"
)

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
	_, err = Pool.Exec(context.Background(), "INSERT INTO images (id, image_data, mimetype) VALUES ($1, $2, $3)", id, data, mimetype.String())
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
	err := Pool.QueryRow(context.Background(), "SELECT image_data, mimetype FROM images WHERE id = $1", split[0]).Scan(&imageData, &mimetype)
	if err != nil {
		WriteJSONError(w, http.StatusNotFound, "Not found.")
		return
	}

	w.Header().Set("Content-Type", mimetype)
	w.Write(imageData)
	w.Header().Set("Content-Type", "image/png")
}

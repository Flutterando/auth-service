package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/auth_server/authjwt"

	"github.com/gorilla/mux"
)

type check struct {
	auth bool
}

func main() {
	r := mux.NewRouter()
	r.Handle("/v1/check", authjwt.JwtMiddleware.Handler(checkauth))
	r.Handle("/v1/gettoken", authjwt.GetTokenHandler)
	r.HandleFunc("/v1/upload", uploadFile)
	r.HandleFunc("/v1/uploads/{img}", webHandler)

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal("ListenAndServe:", nil)
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	img, err := os.Open("uploads/" + vars["img"])
	if err != nil {
		log.Fatal(err) // perhaps handle this nicer
	}
	defer img.Close()
	w.Header().Set("Content-Type", "image/*") // <-- set the content-type header
	io.Copy(w, img)
}

var checkauth = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	c := &check{
		auth: true,
	}
	json.NewEncoder(w).Encode(c)
})

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	var extension = filepath.Ext(handler.Filename)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("uploads", "upload-*"+extension)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!

	jsonReturn := make(map[string]string)
	jsonReturn["file"] = tempFile.Name()
	// fmt.Fprintf(w, jsonReturn)
	js, err := json.Marshal(jsonReturn)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Copyright 2016 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"webapp/datastore"
)

type handler struct {
	stdlog *log.Logger
	errlog *log.Logger
	env    map[string]string

	datastore.DB
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.stdlog.Println("Received Request: ", r.URL.Path)
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			fmt.Fprintln(w, "Bookstore: use '/books' to list all books")
		case "/books":
			h.books(w)
		default:
			writeError(w, http.StatusNotFound)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (h *handler) books(w http.ResponseWriter) {
	books, err := h.DB.AllBooks()
	if err != nil {
		h.errlog.Println(err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprint(w, "Books\n=====\n\n")
	for _, book := range books {
		fmt.Fprintf(w, "%d, %s, %s, %s\n", book.ID, book.Isbn, book.Title, book.Author)
	}
}

func writeError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func Setup() (srv http.Server, err error) {
	keys := []string{
		"WEBAPP_SERVICE_PORT", "WEBDB_SERVICE_HOST",
		"WEBDB_SERVICE_PORT", "WEBDB_SERVICE_NAME",
		"WEBDB_SERVICE_USERNAME", "WEBDB_SERVICE_PASSWORD",
	}
	h := &handler{
		stdlog: log.New(os.Stdout, "[WEBAPP:INFO]: ", log.LstdFlags),
		errlog: log.New(os.Stderr, "[WEBAPP:ERROR]: ", log.LstdFlags),
		env:    make(map[string]string, len(keys)),
	}
	for _, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			err = fmt.Errorf("%s environment variable was not set", key)
			return
		}
		h.env[key] = value
	}
	h.DB, err = datastore.New(
		h.env["WEBDB_SERVICE_USERNAME"] +
			":" + h.env["WEBDB_SERVICE_PASSWORD"] +
			"@tcp(" + h.env["WEBDB_SERVICE_HOST"] +
			":" + h.env["WEBDB_SERVICE_PORT"] +
			")/" + h.env["WEBDB_SERVICE_NAME"],
	)
	srv.Addr = ":" + h.env["WEBAPP_SERVICE_PORT"]
	srv.Handler = h
	return
}

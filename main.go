package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/handlers"
)

func main() {
	logrus.SetReportCaller(true)

	templateMap, err := LoadTemplates()
	if err != nil {
		logrus.Fatal(err)
	}

	h := handlers.New(templateMap)

	// API handler

	// Page handler
	http.HandleFunc("/home", h.PageHome)
	http.HandleFunc("/about", h.PageHome)

	port := ":16000"
	logrus.Infof("Listening on port %s", port)
	logrus.Fatal(http.ListenAndServe(port, nil))
}

package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/api_handlers"
	"github.com/umarkotak/go-kubeseal-gui/page_handlers"
)

func main() {
	logrus.SetReportCaller(true)

	templateMap, err := LoadTemplates()
	if err != nil {
		logrus.Fatal(err)
	}

	ph := page_handlers.New(templateMap)
	ah := api_handlers.New()

	mux := http.NewServeMux()

	// API handler
	mux.HandleFunc("GET /api/config", ah.GetConfig)
	mux.HandleFunc("GET /api/kubectl/get_contexts", ah.GetKubectlContexts)

	// Page handler
	mux.HandleFunc("GET /home", ph.Home)
	mux.HandleFunc("GET /about", ph.Home)

	port := ":16000"
	logrus.Infof("Listening on port %s", port)
	logrus.Fatal(http.ListenAndServe(port, mux))
}

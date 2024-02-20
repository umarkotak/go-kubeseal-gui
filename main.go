package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/api_handlers"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/page_handlers"
)

func main() {
	logrus.SetReportCaller(true)

	config.Load()

	templateMap, err := LoadTemplates()
	if err != nil {
		logrus.Fatal(err)
	}

	ph := page_handlers.New(templateMap)
	ah := api_handlers.New()

	mux := http.NewServeMux()

	// API handler
	mux.HandleFunc("GET /api/config", ah.GetConfig)
	mux.HandleFunc("POST /api/config/clusters/add", ah.AddClustersConfig)
	mux.HandleFunc("POST /api/config/clusters/{alias}/delete", ah.RemoveClustersConfig)
	mux.HandleFunc("POST /api/config/clusters/{alias}/enable_secrets", ah.ClusterEnableSecrets)
	mux.HandleFunc("POST /api/config/controller", ah.SetupConfigController)
	mux.HandleFunc("GET /api/kubectl/get_contexts", ah.GetKubectlContexts)
	mux.HandleFunc("GET /api/kubectl/{alias_name}/secret/{secret_name}/read", ah.KubectlSecretRead)

	// Page handler
	mux.HandleFunc("GET /home", ph.Home)
	mux.HandleFunc("GET /about", ph.Home)
	mux.HandleFunc("GET /config", ph.Config)

	port := ":16000"
	logrus.Infof("Listening on port %s", port)
	logrus.Fatal(http.ListenAndServe(port, mux))
}

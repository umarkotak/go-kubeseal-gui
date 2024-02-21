package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/go-kubeseal-gui/api_handlers"
	"github.com/umarkotak/go-kubeseal-gui/config"
	"github.com/umarkotak/go-kubeseal-gui/page_handlers"

	_ "github.com/go-git/go-billy/v5"
	_ "github.com/go-git/go-billy/v5/memfs"
	_ "github.com/go-git/go-git/v5/plumbing/transport/http"
	_ "github.com/go-git/go-git/v5/storage/memory"
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
	mux.HandleFunc("POST /api/config/clusters/{alias}/sync_secrets", ah.ClusterSyncSecrets)
	mux.HandleFunc("POST /api/config/clusters/{alias}/enable_secrets", ah.ClusterEnableSecrets)
	mux.HandleFunc("POST /api/config/controller", ah.SetupConfigController)
	mux.HandleFunc("POST /api/config/git_integration", ah.SetupGitIntegration)
	mux.HandleFunc("GET /api/kubectl/get_contexts", ah.GetKubectlContexts)
	mux.HandleFunc("GET /api/kubectl/{alias_name}/secrets", ah.GetKubectlSecrets)
	mux.HandleFunc("GET /api/kubectl/{alias_name}/secret/{secret_name}/read", ah.KubectlSecretRead)
	mux.HandleFunc("POST /api/kubectl/{alias_name}/secret/seal", ah.KubectlSeal)
	mux.HandleFunc("POST /api/kubectl/{alias_name}/secret/compare_diff", ah.KubectlSecretDiff)
	mux.HandleFunc("POST /api/kubectl/{alias_name}/secret/seal_and_push", ah.KubectlSealAndPush)

	// Page handler
	mux.HandleFunc("GET /home", ph.Home)
	mux.HandleFunc("GET /about", ph.Home)
	mux.HandleFunc("GET /config", ph.Config)

	port := ":16000"
	logrus.Infof("Open dashboard: http://localhost%s/home", port)
	logrus.Infof("Listening on port %s", port)
	logrus.Fatal(http.ListenAndServe(port, mux))
}

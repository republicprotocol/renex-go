package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/republicprotocol/renex-go/contract"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables.
	port := os.Getenv("PORT")
	network := os.Getenv("NETWORK")
	kyberKey := os.Getenv("KYBER_KEY")
	wyreKey := os.Getenv("WYRE_KEY")
	infuraKey := os.Getenv("INFURA_KEY")
	sentryDSN := os.Getenv("SENTRY_DSN")
	if network == "" {
		log.Fatalf("cannot read network environment")
	}

	latestCommit, err := ioutil.ReadFile("env/latest_commit.txt")
	if err != nil {
		log.Fatalf("cannot read latest commit: %v", err)
	}

	// Load configuration file based on specified network environment.
	config, err := loadConfig(fmt.Sprintf("env/%v/config.json", network))
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	r := mux.NewRouter().StrictSlash(true)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := path.Join("./ui/", r.URL.Path)
		if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
			// A file exists so serve it statically.
			http.FileServer(http.Dir("./ui")).ServeHTTP(w, r)
		} else {
			// A file does not exist so serve the UI template.
			serveTemplate(w, r, config, latestCommit, kyberKey, wyreKey, infuraKey, sentryDSN)
		}
	})

	http.Handle("/", cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	}).Handler(r))

	log.Printf("listening on port %v...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatalf("cannot listen on port %v: %v", port, err)
	}
}

func loadConfig(configFile string) (contract.Config, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return contract.Config{}, err
	}
	var data contract.Config
	json.Unmarshal(file, &data)
	return data, nil
}

func serveTemplate(w http.ResponseWriter, r *http.Request, config interface{}, latestCommit []byte, kyberKey, wyreKey, infuraKey, sentryDSN string) {
	networkData, err := json.Marshal(config)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("cannot marshal network data: " + err.Error()))
		return
	}

	// Create template file and insert environment variables.
	tmpl, err := template.ParseFiles("./ui/index.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("cannot parse layout template: %v", err)))
		return
	}

	if len(networkData) == 0 {
		w.WriteHeader(500)
		w.Write([]byte("invalid data received"))
		return
	}

	if tmpl, err = tmpl.Parse(`
	{{define "env"}}
	<script type="text/javascript">
		console.log('renex-js commit hash: ` + strings.TrimSpace(string(latestCommit)) + `');
		window.KYBER_KEY="` + kyberKey + `";
		window.WYRE_KEY="` + wyreKey + `";
		window.INFURA_KEY="` + infuraKey + `";
		window.SENTRY_DSN="` + sentryDSN + `";
		window.NETWORK=` + string(networkData) + `;
		if (window.NETWORK.ethNetwork !== 'main') {
			document.title = 'RenEx Beta (' + window.NETWORK.ethNetworkLabel + ' Test Network)';
		}
	</script>
	{{end}}
	`); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("cannot execute template: %v", err)))
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("cannot execute template: %v", err)))
		return
	}
}

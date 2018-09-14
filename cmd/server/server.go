package main

import (
	"bytes"
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
	"github.com/rs/cors"
)

type authRequest struct {
	Code   string `json:"code"`
	URI    string `json:"redirect_uri"`
	Key    string `json:"client_id"`
	Secret string `json:"client_secret"`
}

func main() {
	// Load environment variables.
	port := os.Getenv("PORT")
	network := os.Getenv("NETWORK")
	infuraKey := os.Getenv("INFURA_KEY")
	kyberSecret := os.Getenv("KYBER_SECRET")
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
			serveTemplate(w, r, config, latestCommit, infuraKey)
		}
	})

	r.HandleFunc("/kyber", func(w http.ResponseWriter, r *http.Request) {
		// Decode POST request data.
		decoder := json.NewDecoder(r.Body)
		var data authRequest
		err := decoder.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("cannot decode data: %v", err)))
			return
		}

		// Construct new request object with Kyber secret key.
		data.Secret = kyberSecret
		byteArray, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("cannot marshal data: %v", err)))
			return
		}

		// Forward updated request data to Kyber.
		url := "https://kyber.network/oauth/token"
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(byteArray))
		request.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("unable to forward request: %v", err)))
			return
		}
		defer resp.Body.Close()
	}).Methods("POST")

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

func loadConfig(configFile string) (interface{}, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var data interface{}
	json.Unmarshal(file, &data)
	return data, nil
}

func serveTemplate(w http.ResponseWriter, r *http.Request, config interface{}, latestCommit []byte, infuraKey string) {
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
		window.INFURA_KEY="` + infuraKey + `";
		window.NETWORK=` + string(networkData) + `;
		if (window.NETWORK.ethNetwork !== 'mainnet') {
			document.title = 'RenEx (' + window.NETWORK.ethNetworkLabel + ' Test Network)';
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

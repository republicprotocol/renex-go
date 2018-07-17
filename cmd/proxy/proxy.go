package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables.
	port := 3001    // os.Getenv("PORT")
	network := "f∅" // os.Getenv("NETWORK")
	if network == "" {
		log.Fatalf("cannot read network environment")
	}

	// Load configuration file based on specified network environment.
	config, err := loadConfig(fmt.Sprintf("env/%v/config.json", network))
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, r, config)
	})
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
	http.Handle("/", cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET"},
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

func serveTemplate(w http.ResponseWriter, r *http.Request, config interface{}) {
	networkData, err := json.Marshal(config)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("cannot marshal network data: " + err.Error()))
		return
	}

	// Create template file and insert environment variables.
	tmpl, err := template.ParseFiles("./public/index.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("cannot parse layout template: %v", err)))
		return
	}

	// TODO: Add validation for networkData (non-empty, etc.)
	if tmpl, err = tmpl.Parse(`{{define "env"}}<script type="text/javascript">window.NETWORK=` + string(networkData) + `;</script>{{end}}`); err != nil {
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

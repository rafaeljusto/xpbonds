package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rafaeljusto/xpbonds"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return

	} else if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var rates xpbonds.BondReport
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rates); err != nil {
		log.Printf("failed to parse body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bonds, err := xpbonds.FindBestBonds(r.Context(), rates)
	if err != nil {
		log.Printf("failed to find the best bonds: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bondsRaw, err := json.Marshal(bonds)
	if err != nil {
		log.Printf("failed to encode bonds: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bondsRaw)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8090", nil))
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Mensaje string `json:"mensaje"`
}

func main() {
	http.HandleFunc("/api2/201902125/llamar-api1", func(w http.ResponseWriter, r *http.Request) {
		Mensaje := Response{
			Mensaje: "Hola, responde la API: API1 en la VM1, desarrollada por el estudiante Christian Josué Hernández Juárez con carnet 201902125",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Mensaje)
	})

	http.HandleFunc("/api2/201902125/llamar-api3", func(w http.ResponseWriter, r *http.Request) {
		Mensaje := Response{
			Mensaje: "Hola, responde la API: API3 en la VM2, desarrollada por el estudiante Christian Josué Hernández Juárez con carnet 201902125",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Mensaje)
	})
	fmt.Println("API2 escuchando en el puerto 8081...")
	http.ListenAndServe(":8081", nil)
}

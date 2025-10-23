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
	http.HandleFunc("/api3/201902125/llamar-api1", func(w http.ResponseWriter, r *http.Request) {
		Mensaje := Response{
			Mensaje: "Hola, responde la API: API1 en la VM1, desarrollada por el estudiante Christian Josué Hernández Juárez con carnet 201902125",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Mensaje)
	})

	http.HandleFunc("/api3/201902125/llamar-api2", func(w http.ResponseWriter, r *http.Request) {
		Mensaje := Response{
			Mensaje: "Hola, responde la API: API2 en la VM1, desarrollada por el estudiante Christian Josué Hernández Juárez con carnet 201902125",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Mensaje)
	})
	fmt.Println("API3 escuchando en el puerto 8082...")
	http.ListenAndServe(":8082", nil)
}

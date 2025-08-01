package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BrasilApiResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
	Location     struct {
		Type        string `json:"type"`
		Coordinates struct {
			Longitude string `json:"longitude"`
			Latitude  string `json:"latitude"`
		} `json:"coordinates"`
	} `json:"location"`
}

type ViaCepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World using chi!"))
    })
	r.Get("/cep/{cep}", handleCepSearch)

    http.ListenAndServe(":8080", r)
}

func handleCepSearch (w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")
	fmt.Printf("Received CEP: %s\n", cep)

	urlBrasilApi := "https://brasilapi.com.br/api/cep/v1/" + cep

	reqBrasilApi, err := http.NewRequest("GET", urlBrasilApi, nil)
	if err != nil {
		http.Error(w, "Error creating request to Brasil API", http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	respBrasilApi, err := client.Do(reqBrasilApi)
	if err != nil {
		http.Error(w, "Error making request to Brasil API", http.StatusInternalServerError)
		return
	}

	jsonData := &BrasilApiResponse{}
	if err := json.NewDecoder(respBrasilApi.Body).Decode(jsonData); err != nil {
		http.Error(w, "Error decoding response from Brasil API", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Decoded JSON data: %+v\n", jsonData)
	defer respBrasilApi.Body.Close()
	if respBrasilApi.StatusCode != http.StatusOK {
		http.Error(w, "Error fetching data from Brasil API", respBrasilApi.StatusCode)
		return
	}

	urlViaCep := "https://viacep.com.br/ws/" + cep + "/json/"
	reqViaCep, err := http.NewRequest("GET", urlViaCep, nil)
	if err != nil {
		http.Error(w, "Error creating request to ViaCep API", http.StatusInternalServerError)
		return
	}
	respViaCep, err := client.Do(reqViaCep)
	if err != nil {
		http.Error(w, "Error making request to ViaCep API", http.StatusInternalServerError)
		return
	}
	defer respViaCep.Body.Close()
	if respViaCep.StatusCode != http.StatusOK {
		http.Error(w, "Error fetching data from ViaCep API", respViaCep.StatusCode)
		return
	}
	jsonDataViaCep := &ViaCepResponse{}
	if err := json.NewDecoder(respViaCep.Body).Decode(jsonDataViaCep); err != nil {
		http.Error(w, "Error decoding response from ViaCep API", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Decoded ViaCep JSON data: %+v\n", jsonDataViaCep)


}
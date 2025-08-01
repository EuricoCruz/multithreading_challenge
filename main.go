package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type CombinedResponse struct {
	BrasilApi *BrasilApiResponse `json:"brasil_api"`
	ViaCep    *ViaCepResponse    `json:"via_cep"`
}

type BrasilApiResponse struct {
	Api          string `json:"api"`
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
	Api         string `json:"api"`
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

	log.Fatal(http.ListenAndServe(":8080", r))

}

func handleCepSearch(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	fmt.Printf("Received CEP: %s\n", cep)
	ch := make(chan any, 2)

	go searchCepInBrasilApi(cep, ctx, ch)
	go searchCepInViaCepApi(cep, ctx, ch)

	select {
	case result := <-ch:
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("Erro ao formatar JSON:", err)
		} else {
			fmt.Println("Resposta recebida:")
			fmt.Println(string(prettyJSON))
		}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(result)
	case <-ctx.Done():
		fmt.Println("Timeout: nenhuma API respondeu em 1 segundo")
		http.Error(w, "Timeout: nenhuma API respondeu em 1 segundo", http.StatusGatewayTimeout)
	}
}

func searchCepInBrasilApi(cep string, ctx context.Context, ch chan<- any) {
	urlBrasilApi := "https://brasilapi.com.br/api/cep/v1/" + cep

	reqBrasilApi, _ := http.NewRequestWithContext(ctx, "GET", urlBrasilApi, nil)
	respBrasilApi, err := http.DefaultClient.Do(reqBrasilApi)
	if err != nil || respBrasilApi.StatusCode != http.StatusOK {
		return
	}
	defer respBrasilApi.Body.Close()

	var data BrasilApiResponse
	if err := json.NewDecoder(respBrasilApi.Body).Decode(&data); err != nil {
		return
	}
	data.Api = "BrasilAPI"
	ch <- data
}

func searchCepInViaCepApi(cep string, ctx context.Context, ch chan<- any) {
	urlViaCep := "https://viacep.com.br/ws/" + cep + "/json/"

	reqViaCepApi, _ := http.NewRequestWithContext(ctx, "GET", urlViaCep, nil)
	respViaCepApi, err := http.DefaultClient.Do(reqViaCepApi)
	if err != nil || respViaCepApi.StatusCode != http.StatusOK {
		return
	}
	defer respViaCepApi.Body.Close()

	var data ViaCepResponse
	if err := json.NewDecoder(respViaCepApi.Body).Decode(&data); err != nil {
		return
	}

	data.Api = "ViaCep"
	ch <- data

}

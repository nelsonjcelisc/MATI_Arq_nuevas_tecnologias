package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "voltera-p2p-go/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrdenDTO struct {
	ID           string  `json:"id"`
	ProsumidorID string  `json:"prosumidor_id"`
	ComunidadID  string  `json:"comunidad_id"`
	Tipo         string  `json:"tipo"` // "OFERTA" o "DEMANDA"
	Kwh          float64 `json:"kwh"`
	PrecioKwh    float64 `json:"precio_kwh"`
}

var matchingClient pb.MatchingEngineClient

func handleOrden(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	tsRecepcion := time.Now().UnixNano()

	var dto OrdenDTO
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decodificando JSON: %v", err), http.StatusBadRequest)
		return
	}

	if dto.ComunidadID == "" {
		dto.ComunidadID = "GLOBAL"
	}

	req := &pb.OrdenEnergyRequest{
		OrdenId:      dto.ID,
		ProsumidorId: dto.ProsumidorID,
		ComunidadId:  dto.ComunidadID,
		Tipo:         dto.Tipo,
		Kwh:          dto.Kwh,
		PrecioKwh:    dto.PrecioKwh,
		TsRecepcion:  tsRecepcion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := matchingClient.CrearOrden(ctx, req)
	if err != nil {
		log.Printf("[API-HANDLER] Error gRPC: %v", err)
		http.Error(w, fmt.Sprintf("Error procesando en MatchingEngine: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exito":    resp.Exito,
		"mensaje":  resp.Mensaje,
		"match_id": resp.MatchId,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	matchingHost := os.Getenv("MATCHING_ENGINE_HOST")
	if matchingHost == "" {
		matchingHost = "matching-engine:8080"
	}

	conn, err := grpc.NewClient(matchingHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[API-HANDLER] Error conectando a MatchingEngine (%s): %v", matchingHost, err)
	}
	matchingClient = pb.NewMatchingEngineClient(conn)
	log.Printf("[API-HANDLER] Conectado a Matching Engine gRPC en %s", matchingHost)

	http.HandleFunc("/api/orden", handleOrden)
	http.HandleFunc("/health", healthHandler)

	port := ":8000"
	log.Printf("[API-HANDLER] HTTP Gateway escuchando en puerto %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error en HTTP Gateway: %v", err)
	}
}

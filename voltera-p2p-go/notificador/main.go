package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	pb "voltera-p2p-go/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	matchesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "voltera_p2p_matches_procesados_total",
		Help: "Número total de emparejamientos P2P procesados",
	})

	kwhTransadosTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "voltera_p2p_kwh_transados_total",
		Help: "Total de kWh transados en el mercado P2P",
	})

	latenciaMatchingHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "voltera_p2p_latencia_matching_segundos",
		Help:    "Tiempo transcurrido en el motor de emparejamiento",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	})

	latenciaRedHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "voltera_p2p_latencia_red_grpc_segundos",
		Help:    "Tiempo de tránsito por la red (gRPC)",
		Buckets: []float64{0.0005, 0.001, 0.002, 0.005, 0.01, 0.025, 0.05},
	})

	latenciaE2EHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "voltera_p2p_latencia_e2e_segundos",
		Help:    "Tiempo total desde recepción de orden hasta notificación de match",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.3},
	})
)

func init() {
	prometheus.MustRegister(matchesProcessed)
	prometheus.MustRegister(kwhTransadosTotal)
	prometheus.MustRegister(latenciaMatchingHistogram)
	prometheus.MustRegister(latenciaRedHistogram)
	prometheus.MustRegister(latenciaE2EHistogram)
}

type server struct {
	pb.UnimplementedNotificadorServer
}

func (s *server) EnviarNotificacion(ctx context.Context, in *pb.MatchRequest) (*pb.NotificacionResponse, error) {
	ahora := time.Now().UnixNano()

	diffMatching := float64(in.TsEngineMatch-in.TsApiRecepcion) / 1e9
	diffRed := float64(ahora-in.TsApiSalida) / 1e9
	diffE2E := float64(ahora-in.TsApiRecepcion) / 1e9

	if diffMatching > 0 {
		latenciaMatchingHistogram.Observe(diffMatching)
	}
	if diffRed > 0 {
		latenciaRedHistogram.Observe(diffRed)
	}
	if diffE2E > 0 {
		latenciaE2EHistogram.Observe(diffE2E)
	}

	matchesProcessed.Inc()
	kwhTransadosTotal.Add(in.KwhTransados)

	log.Printf("[NOTIFICADOR] Match ID: %s | kWh: %.2f | Precio: $%.2f | Engine: %.2fms | Red: %.2fms | E2E: %.2fms",
		in.MatchId, in.KwhTransados, in.PrecioFinalKwh, diffMatching*1000, diffRed*1000, diffE2E*1000)

	return &pb.NotificacionResponse{Exito: true, Mensaje: "Notificación P2P procesada exitosamente"}, nil
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[NOTIFICADOR] Servidor de métricas Prometheus escuchando en :2112")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al escuchar en :50051: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNotificadorServer(s, &server{})
	reflection.Register(s)

	log.Println("[NOTIFICADOR] Servidor gRPC escuchando en :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir gRPC: %v", err)
	}
}

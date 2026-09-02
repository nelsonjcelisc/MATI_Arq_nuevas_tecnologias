package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	pb "voltera-p2p-go/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	contratosValidados = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "voltera_p2p_contratos_validados_total",
		Help: "Número total de contratos validados exitosamente en caliente",
	})

	contratosInvalidos = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "voltera_p2p_contratos_invalidos_total",
		Help: "Número total de contratos rechazados por anomalías o firmas inválidas",
	})
)

func init() {
	prometheus.MustRegister(contratosValidados)
	prometheus.MustRegister(contratosInvalidos)
}

type server struct {
	pb.UnimplementedContractValidatorServer
}

func (s *server) ValidarContrato(ctx context.Context, in *pb.OrdenEnergyRequest) (*pb.ValidacionResponse, error) {
	// Regla Zero-Trust 1: Sanidad de Datos Físicos
	if in.Kwh <= 0 || in.Kwh > 500 {
		contratosInvalidos.Inc()
		log.Printf("[VALIDATOR ALERT] Contrato Rechazado: kWh anómalos (%.2f) en Orden %s", in.Kwh, in.OrdenId)
		return &pb.ValidacionResponse{
			EsValido:    false,
			CodigoError: "ANOMALOUS_KWH",
			Mensaje:     "Volumen energético fuera de rango permitido para medidor residencial/pyme",
		}, nil
	}

	// Regla Zero-Trust 2: Rango de Tarifa Regulatoria CREG
	if in.PrecioKwh < 50 || in.PrecioKwh > 2000 {
		contratosInvalidos.Inc()
		log.Printf("[VALIDATOR ALERT] Contrato Rechazado: Tarifa anómala ($%.2f) en Orden %s", in.PrecioKwh, in.OrdenId)
		return &pb.ValidacionResponse{
			EsValido:    false,
			CodigoError: "INVALID_TARIFF",
			Mensaje:     "Tarifa fuera del marco regulatorio de precio P2P",
		}, nil
	}

	// Regla Zero-Trust 3: Validación de Firma / Token de Contrato (Simulada para test de anomalías)
	if strings.HasPrefix(in.FirmaDigital, "INVALID") || strings.HasPrefix(in.FirmaDigital, "ATTACK") {
		contratosInvalidos.Inc()
		log.Printf("[VALIDATOR ALERT - ZERO TRUST] Firma Digital Manipulada en Orden %s por Prosumidor %s", in.OrdenId, in.ProsumidorId)
		return &pb.ValidacionResponse{
			EsValido:    false,
			CodigoError: "SIGNATURE_TAMPERED",
			Mensaje:     "Firma digital del contrato o credenciales del medidor no coinciden",
		}, nil
	}

	contratosValidados.Inc()
	return &pb.ValidacionResponse{
		EsValido:    true,
		CodigoError: "OK",
		Mensaje:     "Contrato P2P verificado exitosamente",
	}, nil
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[CONTRACT VALIDATOR] Servidor de métricas escuchando en :2113")
		log.Fatal(http.ListenAndServe(":2113", nil))
	}()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("[CONTRACT VALIDATOR] Error al escuchar en :50052: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterContractValidatorServer(s, &server{})
	reflection.Register(s)

	log.Println("[CONTRACT VALIDATOR] Servidor gRPC escuchando en :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("[CONTRACT VALIDATOR] Error en gRPC: %v", err)
	}
}

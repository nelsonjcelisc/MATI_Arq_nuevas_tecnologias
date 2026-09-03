package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "voltera-p2p-go/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type OrderBook struct {
	mu       sync.Mutex
	ofertas  []*pb.OrdenEnergyRequest // Venta
	demandas []*pb.OrdenEnergyRequest // Compra
}

var (
	book          = &OrderBook{}
	notifClient   pb.NotificadorClient
	matchCounter  int64
	counterMutex  sync.Mutex
)

func generateMatchID() string {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	matchCounter++
	return fmt.Sprintf("M-%d-%d", time.Now().Unix(), matchCounter)
}

type server struct {
	pb.UnimplementedMatchingEngineServer
}

func (s *server) CrearOrden(ctx context.Context, in *pb.OrdenEnergyRequest) (*pb.OrdenEnergyResponse, error) {
	log.Printf("[MATCHING ENGINE] Orden Recibida: ID=%s | Tipo=%s | kWh=%.2f | Precio=$%.2f | Comunidad=%s | Firma=%s",
		in.OrdenId, in.Tipo, in.Kwh, in.PrecioKwh, in.ComunidadId, in.FirmaDigital)

	book.mu.Lock()
	defer book.mu.Unlock()

	matched := false
	var matchedID string

	if in.Tipo == "OFERTA" { // Venta de Energía
		for i, d := range book.demandas {
			if d.PrecioKwh >= in.PrecioKwh && (d.ComunidadId == in.ComunidadId || in.ComunidadId == "GLOBAL") {
				matched = true
				matchedID = generateMatchID()
				kwhTransados := min(in.Kwh, d.Kwh)
				precioFinal := (in.PrecioKwh + d.PrecioKwh) / 2.0

				log.Printf("[MATCHING ENGINE] MATCH ENCONTRADO! MatchID=%s | Oferente=%s | Demandante=%s | kWh=%.2f",
					matchedID, in.ProsumidorId, d.ProsumidorId, kwhTransados)

				if d.Kwh <= in.Kwh {
					book.demandas = append(book.demandas[:i], book.demandas[i+1:]...)
				} else {
					d.Kwh -= kwhTransados
				}

				go notificarMatch(matchedID, in.OrdenId, d.OrdenId, in.ProsumidorId, d.ProsumidorId, kwhTransados, precioFinal, in.TsRecepcion)
				break
			}
		}
		if !matched {
			book.ofertas = append(book.ofertas, in)
		}
	} else { // Demanda de Energía
		for i, o := range book.ofertas {
			if in.PrecioKwh >= o.PrecioKwh && (o.ComunidadId == in.ComunidadId || in.ComunidadId == "GLOBAL") {
				matched = true
				matchedID = generateMatchID()
				kwhTransados := min(in.Kwh, o.Kwh)
				precioFinal := (in.PrecioKwh + o.PrecioKwh) / 2.0

				log.Printf("[MATCHING ENGINE] MATCH ENCONTRADO! MatchID=%s | Oferente=%s | Demandante=%s | kWh=%.2f",
					matchedID, o.ProsumidorId, in.ProsumidorId, kwhTransados)

				if o.Kwh <= in.Kwh {
					book.ofertas = append(book.ofertas[:i], book.ofertas[i+1:]...)
				} else {
					o.Kwh -= kwhTransados
				}

				go notificarMatch(matchedID, o.OrdenId, in.OrdenId, o.ProsumidorId, in.ProsumidorId, kwhTransados, precioFinal, in.TsRecepcion)
				break
			}
		}
		if !matched {
			book.demandas = append(book.demandas, in)
		}
	}

	msg := "Orden agregada al libro de ofertas/demandas"
	if matched {
		msg = fmt.Sprintf("Orden emparejada exitosamente (MatchID: %s)", matchedID)
	}

	return &pb.OrdenEnergyResponse{
		Exito:   true,
		Mensaje: msg,
		MatchId: matchedID,
	}, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func notificarMatch(matchID, ofertaID, demandaID, oferenteID, demandanteID string, kwh, precio float64, tsApiRecepcion int64) {
	if notifClient == nil {
		return
	}
	tsEngineMatch := time.Now().UnixNano()
	tsApiSalida := time.Now().UnixNano()

	req := &pb.MatchRequest{
		MatchId:                matchID,
		OfertaId:               ofertaID,
		DemandaId:              demandaID,
		ProsumidorOferenteId:   oferenteID,
		ProsumidorDemandanteId: demandanteID,
		KwhTransados:           kwh,
		PrecioFinalKwh:         precio,
		TsEngineMatch:          tsEngineMatch,
		TsApiRecepcion:         tsApiRecepcion,
		TsApiSalida:            tsApiSalida,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := notifClient.EnviarNotificacion(ctx, req)
	if err != nil {
		log.Printf("[MATCHING ENGINE] Error enviando notificación gRPC: %v", err)
	}
}

func main() {
	notifHost := os.Getenv("NOTIFICADOR_HOST")
	if notifHost == "" {
		notifHost = "notificaciones:50051"
	}

	conn, err := grpc.NewClient(notifHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[MATCHING ENGINE] Advertencia: No se pudo conectar al Notificador (%v)", err)
	} else {
		notifClient = pb.NewNotificadorClient(conn)
		log.Printf("[MATCHING ENGINE] Conectado exitosamente a Notificador gRPC en %s", notifHost)
	}

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error al escuchar en puerto :8080: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMatchingEngineServer(s, &server{})
	reflection.Register(s)

	log.Println("[MATCHING ENGINE] Servidor gRPC escuchando en :8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir gRPC: %v", err)
	}
}

# Bitácora Reto 1: Requisitos de Entorno y Estrategia Docker

## 1. Verificación del Entorno de Ejecución
- **OS:** Linux (Ubuntu/Debian host).
- **Go Runtime:** `go1.23.0 linux/amd64` (instalado en el sistema).
- **Docker Engine:** `Docker 29.7.2`.
- **Docker Compose:** `v5.5.0`.

---

## 2. Reutilización de la Imagen Docker (Multi-Stage Build en Go)

Se adopta exactamente la **misma estrategia multi-etapa** del proyecto de referencia (`golang:1.23-alpine` + `alpine:latest`).

### Ventajas de la Imagen Elegida:
1. **Multi-Plataforma:** Compila de forma transparente en arquitecturas **Mac (Apple Silicon ARM64)**, **Linux (amd64)** y **Windows**.
2. **Auto-Contenida:** Genera el código Protobuf gRPC internamente durante el build con `protoc` y compila los binarios en la etapa de construcción (`builder`).
3. **Ultra-Ligera:** La imagen final utiliza `alpine:latest`, resultando en un contenedor de **menos de 20MB** con tiempo de arranque de milisegundos.

---

## 3. Estructura Inicial del Proyecto P2P en Go

```
/mnt/data/runtime/Nuevas_tec/voltera-p2p-go/
├── proto/
│   └── p2p.proto               # Contrato gRPC (MatchRequest, OrdenEnergyRequest, etc.)
├── api-handler/                # Microservicio de Ingesta HTTP REST -> gRPC
├── matching-engine/            # Microservicio con OrderBook P2P en Memoria
└── notificador/                # Microservicio gRPC de Métricas y Observabilidad (Prometheus)
```

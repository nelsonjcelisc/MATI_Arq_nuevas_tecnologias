# Voltera P2P Solar Energy Trading Platform Engine (Go Microservices)

Sistema distribuido de emparejamiento (*Matching Engine*) de energía solar Peer-to-Peer (P2P) desacoplado en microservicios mediante **Go**, **gRPC**, **Zero-Trust Validation**, **Nginx Load Balancer** y **Prometheus**.

---

## 🏛️ Arquitectura del Sistema

La arquitectura sigue una topología desacoplada orientada a eventos e ingesta balanceada. El flujo perimetral valida contratos energéticos bajo un esquema **Zero-Trust** antes de permitir la entrada de órdenes al libro de ofertas/demandas en memoria (*OrderBook*), despachando notificaciones gRPC asíncronas a ambas partes del emparejamiento (*Oferente* y *Demandante*).

```mermaid
graph TD
    subgraph Clientes_Pruebas["Clientes y Pruebas"]
        ART["Artillery Load Tester / Prosumidores"]
    end

    subgraph Perimetro_Ingesta["Perímetro de Ingesta y Balanceo"]
        NGX["Nginx Load Balancer (:8000)<br/>Least-Conn Strategy"]
    end

    subgraph Cluster_API_Gateway["Cluster API Gateway Elastic (Go)"]
        API1["API Handler Node 1"]
        API2["API Handler Node 2"]
        API3["API Handler Node 3"]
    end

    subgraph Servicios_gRPC["Servicios Internos gRPC (Go)"]
        VAL["contract-validator (:50052)<br/>Zero-Trust Contract Engine"]
        ENG["matching-engine (:8080)<br/>OrderBook & Price-Time Priority"]
        NOT["notificaciones (:50051)<br/>Bidirectional Dispatcher"]
    end

    subgraph Monitoreo_Observabilidad["Monitoreo y Observabilidad"]
        PROM["Prometheus Server (:9090)<br/>Metrics & Telemetry"]
    end

    %% Flujos de Red
    ART -->|HTTP POST /api/v1/ordenes| NGX
    NGX -->|Round-Robin / Least-Conn| API1
    NGX -->|Round-Robin / Least-Conn| API2
    NGX -->|Round-Robin / Least-Conn| API3

    API1 -->|1. gRPC ValidarContrato| VAL
    API2 -->|1. gRPC ValidarContrato| VAL
    API3 -->|1. gRPC ValidarContrato| VAL

    API1 -->|2. gRPC CrearOrden (Si es válido)| ENG
    API2 -->|2. gRPC CrearOrden (Si es válido)| ENG
    API3 -->|2. gRPC CrearOrden (Si es válido)| ENG

    ENG -->|3. Goroutine gRPC NotificarMatch| NOT
    NOT -->|Notifica Oferente y Demandante| ART

    %% Telemetría Prometheus
    NOT -.->|Metrics :2112| PROM
    VAL -.->|Metrics :2113| PROM
```

---

## 🎯 Atributos de Calidad y Escenarios ASR (Architectural Significant Requirements)

### 🚀 ASR 1: Desempeño y Tolerancia a Picos de Emparejamiento (Performance & Scalability)
* **Fuente del Estímulo:** Prosumidores solares (residenciales y comerciales) durante picos de generación fotovoltaica o demanda.
* **Estímulo:** Ráfaga masiva de órdenes de emparejamiento de energía P2P que multiplica por $10\times$ el volumen habitual, alcanzando hasta **2 millones de solicitudes por minuto**.
* **Artefacto:** Cluster de API Gateways (`api-handler`) + Nginx LB + `matching-engine` en Go.
* **Entorno:** Operación normal del sistema sometido a picos extremos de tráfico.
* **Respuesta:** Procesar la ingesta REST, validar contratos en caliente, ejecutar el calce en el OrderBook y emitir las notificaciones bidireccionales.
* **Medida de Respuesta:** Latencia End-to-End en el **percentil 95 ($p95$) $\le 300\text{ ms}$** garantizando una tasa de procesamiento sin pérdida de datos.

### 🛡️ ASR 2: Ciberseguridad y Detección de Contratos Maliciosos (Security & Grid Safety)
* **Fuente del Estímulo:** Actor o dispositivo malicioso intentando inyectar contratos energéticos corruptos o manipulados en la red eléctrica.
* **Estímulo:** Solicitud HTTP/gRPC con firmas digitales alteradas (`INVALID_SIGNATURE`), valores de energía anómalos ($\le 0$ kWh o $> 500$ kWh) o tarifas no autorizadas.
* **Artefacto:** Microservicio perimetral `contract-validator` en Go Docker (`:50052`).
* **Entorno:** Operación normal de ingesta en caliente.
* **Respuesta:** Detectar la firma o parámetro alterado, descartar la orden inmediatamente antes de ingresar al OrderBook del motor de emparejamiento e incrementar las métricas de alerta.
* **Medida de Respuesta:** Tiempo de identificación y rechazo **$\le 200\text{ ms}$** desde la recepción de la solicitud en la frontera, retornando un código `HTTP 400 Bad Request` y registrando el incidente en Prometheus (`voltera_p2p_contratos_invalidos_total`).

---

## 📈 Catálogo de Métricas Expuestas en Prometheus

Los microservicios instrumentan y exponen métricas nativas de Prometheus a través de los endpoints de telemetría `:2112` (`notificador`) y `:2113` (`contract-validator`).

| Nombre de la Métrica | Tipo | Descripción y Propósito de Negocio / Técnico |
| :--- | :---: | :--- |
| `voltera_p2p_matches_procesados_total` | `Counter` | Número total de emparejamientos P2P concretados con éxito por el algoritmo. |
| `voltera_p2p_notificaciones_oferente_total` | `Counter` | Total de notificaciones despachadas al prosumidor **Oferente** (vendedor de energía). |
| `voltera_p2p_notificaciones_demandante_total` | `Counter` | Total de notificaciones despachadas al prosumidor **Demandante** (comprador de energía). |
| `voltera_p2p_kwh_transados_total` | `Counter` | Acumulado total de kilovatios-hora (kWh) solares comercializados en la red. |
| `voltera_p2p_contratos_invalidos_total` | `Counter` | **Métrica de Ciberseguridad (ASR 2 / Zero-Trust):** Registra contratos rechazados por firmas alteradas o valores energéticos anómalos. |
| `voltera_p2p_latencia_matching_segundos` | `Histogram` | Distribución del tiempo (segundos) dedicado exclusivamente a resolver la prioridad precio-tiempo en el OrderBook. |
| `voltera_p2p_latencia_red_grpc_segundos` | `Histogram` | Tiempo de tránsito y serialización de mensajes en la red interna gRPC. |
| `voltera_p2p_latencia_e2e_segundos` | `Histogram` | Latencia End-to-End completa desde la ingesta REST hasta la notificación bidireccional. |

---

## 📊 Consultas Frecuentes en PromQL

Accede a la interfaz de Prometheus en `http://localhost:9090` para graficar:

* **Latencia $p95$ de Emparejamiento (ASR 1):**
  ```promql
  histogram_quantile(0.95, sum(rate(voltera_p2p_latencia_matching_segundos_bucket[1m])) by (le)) * 1000
  ```
* **Throughput de Emparejamientos por Segundo:**
  ```promql
  rate(voltera_p2p_matches_procesados_total[1m])
  ```
* **Alertas de Ciberseguridad / Contratos Inválidos Rechazados (ASR 2):**
  ```promql
  sum(voltera_p2p_contratos_invalidos_total)
  ```
* **Total de Energía Comercializada en la Red (kWh):**
  ```promql
  voltera_p2p_kwh_transados_total
  ```

---

## 🛠️ Estructura de Microservicios

El repositorio contiene los siguientes módulos desarrollados en **Go (golang:1.23)**:

```
voltera-p2p-go/
├── api-handler/          # Gateway HTTP REST (Escalable Horizontalmente)
├── contract-validator/   # Microservicio Zero-Trust de Validación de Contratos (gRPC :50052, Metrics :2113)
├── matching-engine/      # Motor P2P con OrderBook en Memoria (gRPC :8080)
├── notificador/          # Despachador de Notificaciones Bidireccionales (gRPC :50051, Metrics :2112)
├── nginx/                # Configuración de Load Balancer Nginx (Puerto :8000)
├── proto/                # Definiciones Protocol Buffers (p2p.proto)
├── artillery-tests/      # Suites de Pruebas de Carga y Ráfagas (Artillery)
└── docker-compose.yml    # Orquestación de Malla de Contenedores y Prometheus
```

---

## 🚀 Despliegue y Ejecución con Docker Compose

Para levantar toda la arquitectura distribuida con 3 réplicas del API Gateway y Nginx LB:

```bash
# 1. Clonar el repositorio y navegar al proyecto
cd voltera-p2p-go

# 2. Desplegar el cluster completo con Docker Compose
docker compose up -d --scale api-handler=3 --build

# 3. Verificar estado de los contenedores
docker ps
```

# Bitácora Reto 1: Arquitectura Detallada del Motor P2P Voltera (Go)

## 1. Vista General de Arquitectura (C4 - Nivel de Contenedores)

```
                                +---------------------------------------------------+
                                |              CLIENTES / SIMULADOR                 |
                                |  - App Prosumidores / IoT Medidores (AMI)         |
                                |  - Generador de Carga (Artillery / Go Client)     |
                                +-------------------------+-------------------------+
                                                          |
                                                          | HTTP / REST (JSON)
                                                          v
                                +---------------------------------------------------+
                                |                 NGINX LOAD BALANCER               |
                                |  - Reverse Proxy & Round Robin Load Balancer      |
                                |  - Puerto HTTP Exp: 80 / 8080                     |
                                +-------------------------+-------------------------+
                                                          |
                                           +--------------+--------------+
                                           |                             |
                                           v                             v
                        +-----------------------------------+  +-----------------------------------+
                        |   API Handler / Ingestion (Go)   |  |   API Handler / Ingestion (Go)   |
                        |   Instancia 1                     |  |   Instancia 2 ... N               |
                        +-----------------+-----------------+  +-----------------+-----------------+
                                          |                                      |
                                          +------------------+-------------------+
                                                             |
                                                             | gRPC (Internal IPC)
                                                             v
                                        +-----------------------------------------+
                                        |      Matching Engine P2P (Go)           |
                                        |  - Order Books (Oferta / Demanda)       |
                                        |  - Algoritmo de Calce Energético        |
                                        |  - Event Loops con Worker Pools         |
                                        +----+-------------------------------+----+
                                             |                               |
                                gRPC Async   |                               | SQL / Driver
                                             v                               v
                        +----------------------------+   +----------------------------------------+
                        |  Notificador & Metrics     |   |   PostgreSQL / MongoDB (Journal DB)     |
                        |  Service (Go)              |   |   - Audit Log & Histórico Transacciones    |
                        |  - Exporter Prometheus     |   |   - Persistencia de Matches                |
                        |  - Puerto: 2112 / 50051    |   +----------------------------------------+
                        +--------------+-------------+
                                       |
                                       | Pull / Scraping (:2112/metrics)
                                       v
                        +----------------------------+
                        |   Prometheus Server        |
                        |   - Dashboard & Alerts     |
                        |   - Puerto: 9090           |
                        +----------------------------+
```

---

## 2. Flujo de Datos y Componentes Detallados

### A. Componente API Handler (Ingesta & API REST)
* **Función:** Recibir peticiones de prosumidores (Ofertas de Excedente de Energía $kWh$ / Demandas de Consumo $kWh$).
* **Protocolo de Entrada:** HTTP/REST (JSON) balanceado por Nginx.
* **Manejo de Carga:** Desacoplado mediante Worker Pools concurrentes en Go. Convierte solicitudes HTTP en mensajes Protobuf binarios y los envía vía gRPC al Matching Engine.

### B. Componente Matching Engine P2P (Núcleo de Negocio en Go)
* **Función:** Mantener los **Order Books en Memoria** para emparejamiento ultra-rápido ($p95 \le 300\text{ms}$).
* **Estructura de Datos Interna:**
  - **Libro de Ofertas (Sell Orders):** Ordenadas por menor precio ($/kWh$), cercanía geográfica/comunidad y antigüedad (FIFO).
  - **Libro de Demandas (Buy Orders):** Ordenadas por mayor disponibilidad a pagar ($/kWh$) y antigüedad.
* **Proceso de Emparejamiento (Match Execution):**
  1. Al recibir una oferta/demanda, ejecuta el calce continuo.
  2. Si hay coincidencia de tarifa y comunidad energética, calcula los $kWh$ transados.
  3. Genera un evento `MatchExecutedEvent` con métricas de tiempo de ejecución (`ts_engine_match`).
  4. Envía de forma asíncrona la notificación al microservicio de métricas y persiste en la base de datos de auditoría.

### C. Componente Notificador & Servicio de Métricas (Go + Prometheus)
* **Función:** Registrar eventos de auditoría y exponer métricas de latencia de arquitectura de baja sobrecarga.
* **Protocolo:** Servidor **gRPC** en puerto `:50051` y servidor HTTP Prometheus en puerto `:2112`.
* **Métricas Clave Expuestas:**
  1. `voltera_p2p_matches_total`: Contador total de transacciones energéticas completadas.
  2. `voltera_p2p_matching_latency_seconds`: Histograma del tiempo transcurrido desde la recepción de la oferta hasta el calce.
  3. `voltera_p2p_grpc_transit_latency_seconds`: Latencia de red entre servicios gRPC (Go $\rightarrow$ Go).

### D. Componente Base de Datos / Journal (PostgreSQL / MongoDB)
* **Función:** Registro inmutable de transacciones auditables ($kWh$ transferidos, valores monetarios y timestamping para posterior liquidación o certificados de origen Blockchain).

---

## 3. Matriz de Atributos de Calidad vs Decisión de Arquitectura

| Atributo de Calidad | Meta / Escenario | Decisión de Arquitectura en Go |
| :--- | :--- | :--- |
| **Latencia** | $p95 \le 300\text{ms}$ en calce P2P | Uso de Go goroutines + canales + estructuras en memoria (OrderBook en Heap/Mutex sin I/O bloqueante en el thread principal). |
| **Escalabilidad** | $\ge 100.000$ ofertas/min | Escalabilidad horizontal del API Handler con Nginx Round-Robin y gRPC multiplexing (HTTP/2 persistent connections). |
| **Disponibilidad** | $\ge 99.99\%$ mensual | Degradación elegante: si el Notificador/Prometheus falla, el Matching Engine buffering eventos en memoria sin bloquear la transacción. |
| **Seguridad Zero-Trust** | Integridad de transacciones | Canales gRPC TLS/mTLS entre contenedores y firma de payloads de prosumidores. |
| **Modificabilidad** | Nuevos algoritmos de tarifa | Aislamiento de lógica mediante Interfaces en Go (`MatchingStrategy` pattern). |


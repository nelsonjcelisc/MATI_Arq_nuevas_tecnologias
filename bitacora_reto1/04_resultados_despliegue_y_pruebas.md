# Bitácora Reto 1: Resultados del Despliegue y Pruebas P2P (Go)

## 1. Estado del Despliegue en Docker

Se construyeron y desplegaron exitosamente los 4 contenedores en Go utilizando **Go Workspaces (`go.work`)**:

- **`api_handler_voltera`**: Escuchando HTTP en `:8000`.
- **`matching_engine_voltera`**: Escuchando gRPC en `:8080`.
- **`notificador_voltera`**: Escuchando gRPC en `:50051` y Métricas HTTP Prometheus en `:2112`.
- **`prometheus_voltera`**: Dashboard de monitoreo en `:9090`.

---

## 2. Prueba Funcional de Emparejamiento P2P (CURL)

### Paso A: Publicación de Demanda (Compra de Energía)
- **Prosumidor:** `VECINO-C1`
- **Comunidad:** `COM-NORTE`
- **Demanda:** 50.0 kWh a un precio máximo de $650.0 / kWh.
- **Resultado:** Orden registrada exitosamente en el libro de demandas.

```json
{"exito":true,"match_id":"","mensaje":"Orden agregada al libro de ofertas/demandas"}
```

### Paso B: Publicación de Oferta (Venta de Excedentes Solares)
- **Prosumidor:** `PROSUMIDOR-V1`
- **Comunidad:** `COM-NORTE`
- **Oferta:** 30.0 kWh a un precio de $600.0 / kWh.
- **Resultado:** **Match Inmediato** detectado por el motor de emparejamiento.

```json
{"exito":true,"match_id":"M-1788319026-1","mensaje":"Orden emparejada exitosamente (MatchID: M-1788319026-1)"}
```

---

## 3. Validación de Latencias vs ASRs

Métricas registradas por el microservicio notificador y exportadas a **Prometheus (`:2112`)**:

| Métrica de Arquitectura | Valor Medido | Meta ASR | Estado |
| :--- | :--- | :--- | :--- |
| **Tiempo en Matching Engine (Go)** | **1.28 ms** | $\le 300\text{ ms}$ | ✅ Excelente |
| **Latencia Red (gRPC IPC)** | **14.08 ms** | $\le 50\text{ ms}$ | ✅ Excelente |
| **Latencia E2E (Ingesta $\rightarrow$ Match)** | **15.36 ms** | $\le 500\text{ ms}$ | ✅ Excelente |
| **Energía Transada Total** | **30.0 kWh** | N/A | Transacción completada |

---

## 4. Control de Versiones en Git

- **Rama:** `v1.0.0-p2p-engine-go`
- **Commit:** `d197067` - *feat: primera version del motor de emparejamiento p2p voltera en go*

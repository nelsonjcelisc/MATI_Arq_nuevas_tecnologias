# Bitácora Reto 1: Motor de Emparejamiento P2P Voltera (Go)

## Contexto del Proyecto
- **Empresa:** Voltera (Comercializadora digital de energía distribuida / Energytech).
- **Materia:** ARTI4208 - Arquitecturas de Nueva Generación (MATI - Universidad de los Andes).
- **Objetivo:** Migrar / Implementar el **Motor de Emparejamiento P2P (Matching Engine)** y los componentes auxiliares (Métricas/Notificaciones) en **Go**, comunicados mediante **gRPC**, dockerizados y monitoreados con **Prometheus**.

---

## 1. Atributos de Calidad y ASRs Identificados (Caso Voltera)

### ASR-1: Latencia de Emparejamiento P2P (Escenario Principal)
- **Estímulo:** Publicación continua de ofertas de excedentes solares y demandas de energía por parte de prosumidores.
- **Fuente:** Medidores inteligentes (AMI/IoT) y Apps de Prosumidores.
- **Artefacto:** `Matching Engine P2P` (Go).
- **Entorno:** Operación normal y horas pico de generación solar.
- **Respuesta Esperada:** Emparejar oferta y demanda de la comunidad de forma continua y confiable.
- **Medida de Respuesta (Métrica):**
  - **Latencia de emparejamiento p95 ≤ 300 ms** (alineado con la especificación pág. 12 del caso Voltera).
  - **Latencia de red/comunicación gRPC p95 ≤ 50 ms**.

### ASR-2: Escalabilidad Masiva en Hora Pico Solar
- **Estímulo:** Ráfagas masivas al mediodía (máxima oferta de excedentes).
- **Fuente:** 2.000.000 de medidores conectados.
- **Respuesta Esperada:** Ingesta sostenida de lecturas (≥ 2.200 lecturas/s) y emparejamiento de ≥ 100.000 ofertas/minuto sin pérdida de datos ni contrapresión descontrolada.

---

## 2. Decisiones Arquitectónicas

1. **Stack Tecnológico:**
   - Lenguaje principal: **Go** (Golang) para baja latencia, alto rendimiento y bajo consumo de memoria.
   - Protocolo de Inter-Servicios: **gRPC** sobre HTTP/2 (contratos con Protobuf).
   - Observabilidad: **Prometheus** + Exporter de métricas HTTP (`:2112`).
   - Contenerización: **Docker Compose** con control de recursos CPU/Memory.

2. **Dominio de Datos (Voltera P2P):**
   - **Prosumidor:** Identificador de usuario/medidor y comunidad energética.
   - **Oferta de Energía (Venta):** Excedentes en $kWh$, tarifa $/kWh$, timestamp.
   - **Demanda de Energía (Compra):** Requerimiento en $kWh$, tarifa máx $/kWh$, timestamp.
   - **Match / Transacción P2P:** Transacción emparejada entre oferente y demandante.

---

## 3. Plan de Trabajo Incremental

- [x] **Paso 1:** Análisis del caso Voltera y definición de ASRs.
- [ ] **Paso 2:** Diseño del contrato gRPC (`voltera_p2p.proto`) adaptado a Voltera.
- [ ] **Paso 3:** Implementación del microservicio `matching-engine` en Go.
- [ ] **Paso 4:** Implementación del microservicio `notificaciones-go` / Métricas.
- [ ] **Paso 5:** Orquestación con `docker-compose.yml` y configuración de Prometheus.
- [ ] **Paso 6:** Pruebas de carga y validación de ASRs (p95 ≤ 300 ms).

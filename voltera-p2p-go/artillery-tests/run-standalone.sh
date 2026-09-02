#!/bin/bash

echo "================================================================="
echo "⚡ VOLTERA P2P - STANDALONE ARTILLERY BENCHMARK (SIN AGY INTERACTIVO)"
echo "================================================================="
echo "Este script corre la prueba de carga en background para no saturar"
echo "la CPU del sistema con la consola interactiva."
echo ""
echo "Instrucciones de uso:"
echo " 1. Ejecuta: ./run-standalone.sh"
echo " 2. Mira los logs en vivo con: tail -f benchmark-standalone.log"
echo " 3. Revisa los datos de Prometheus en http://localhost:9090"
echo "================================================================="

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="benchmark_${TIMESTAMP}.log"

echo "Iniciando prueba estocástica de 39,980 solicitudes a las $(date)..." | tee -a "$LOG_FILE"
npx artillery run test-p2p.yml | tee -a "$LOG_FILE"

echo "" | tee -a "$LOG_FILE"
echo "Prueba completada a las $(date). Revisa las métricas en Prometheus." | tee -a "$LOG_FILE"

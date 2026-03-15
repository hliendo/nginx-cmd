#!/usr/bin/env bash
# ==============================================================================
# nginx-cmd (Docker Host Wrapper)
# ==============================================================================
# Este script permite ejecutar el comando 'nginx-cmd' dentro del contenedor 
# de Docker desde el host sin tener que hacer 'docker exec' manualmente.
# ==============================================================================

CONTAINER_NAME="nginx"

# Verificar si Docker está instalado
if ! command -v docker &> /dev/null; then
    echo "Error: docker no está instalado. http://docker.com"
    exit 1
fi

# Detectar si es una terminal interactiva (tty) para pasar flags de docker
DOCKER_OPTS="-i"
if [ -t 0 ]; then
    DOCKER_OPTS="-it"
fi

# Ejecutar el comando dentro del contenedor
docker exec $DOCKER_OPTS "$CONTAINER_NAME" nginx-cmd "$@"

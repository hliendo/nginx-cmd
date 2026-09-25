#!/bin/sh
# entrypoint.sh
# Inicia el daemon interno de renovación de certificados (nginx-cmd serve) en
# background y luego nginx en foreground.
#
# Antes esto usaba crond + /etc/crontabs/root. Se reemplazó porque depender
# de un cron del sistema operativo es un punto de falla silencioso: si la
# ruta del binario o el directorio de log quedan desincronizados, el cron
# falla todos los días sin que nadie lo note. "nginx-cmd serve" es el mismo
# proceso que ya corre en este contenedor, se duerme y se despierta solo, y
# deja un latido en los logs aunque no haya nada para renovar.

echo "🚀 Iniciando nginx-cmd serve (renovación automática, sin cron externo)..."
nginx-cmd serve &

echo "🚀 Iniciando Nginx (Foreground)..."
exec /opt/nginx/sbin/nginx -g "daemon off;"

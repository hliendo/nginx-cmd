#!/bin/sh
# entrypoint.sh
# Inicia crond en background y luego nginx en foreground

echo "🚀 Iniciando Crond (Background)..."
crond -b -L /var/log/cron.log

echo "🚀 Iniciando Nginx (Foreground)..."
exec /opt/nginx/sbin/nginx -g "daemon off;"

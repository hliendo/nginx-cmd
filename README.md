# Nginx Command Tool (nginx-cmd)

Este proyecto no busca reemplazar la genialidad de Caddy ni la comodidad de Nginx Proxy Manager. Ambos son logros enormes de la comunidad. nginx-cmd se plantea como una alternativa minimalista para quienes necesitan control absoluto sobre Nginx, modularidad mediante includes y un ciclo de vida limpio sin deuda técnica.


**`nginx-cmd`** no es un motor de plantillas de IA, ni un clon de Nginx-Proxy-Manager lleno de dependencias web, ni pretende reemplazar el poder de escritura humana en un archivo `.conf`.

Es un orquestador ágil, estricto e inmutable, escrito en **Golang**, nacido en el seno de la arquitectura CDC para eliminar la fragilidad de los scripts históricos en Bash, gestionando la inserción segura, inicialización y ciclado vital (Bootstrap ACME y trust stores) de los dominios, **sin ceder ni una coma de control estructural**.

## 🏗️ La Historia y la Verdad Industrial

Originalmente, administrábamos Nginx crudo instalando Certbot y lidiando a mano con dependencias (o usando bash-scripts).
Pasamos por **Caddy**: brillante, fácil, pero perdíamos el control granular superior de Nginx para arquitecturas complejas (módulos, gRPC pesado, caching estricto).
Pasamos por **Nginx-Proxy-Manager**: pesado, inserta una capa visual y una base de datos que expanden la superficie de ataque, sacando la "verdad" de los archivos `.conf` puros.

Así nació este proyecto: **la unión de la magia desatendida de Caddy con el poder intocable y crudo de Nginx**.

## ⚖️ Doctrina Arquitectónica de `nginx-cmd`

Para escalar este proyecto a la comunidad, se aplican dos reglas de oro inquebrantables:

### 1. La Configuración de Negocio es Humana ("Features como Includes")
`nginx-cmd` genera Server Blocks ("Stubs") idiotamente simples. Toda la inteligencia operativa real (Buffer sizes, WebSocket upgrades, proxy timeouts, CORS) no se le inyecta con variables flag por CLI, ni la procesa Go. **Esa lógica la retiene el Sysadmin en su disco local**, mediante archivos modulares montados en `conf.d/domains/includes/` (ej: `proxy-websocket.conf`). El binario solo inyecta la instrucción `include` hacia ellos. Esto garantiza portabilidad total y cero necesidad de recompilar el binario ante cada ajuste de infraestructura.

### 2. Idempotencia y Destrucción: La Regla "No Edit"
`nginx-cmd` **no tiene** comando `edit`.
Modificar dominios parcializados mediante CLI implica parsear bidireccionalmente y manejar variables de estado que se corrompen (ej. si falla la escritura a la mitad, Nginx muere). 
Si te equivocaste al crear un dominio (ej. escribiste mal el target o pusiste SPA en vez de Proxy):
```bash
# Correcto:
nginx-cmd remove mi-web.com
nginx-cmd add mi-web.com proxy --target 10.0.0.5:80
```
**Destrucción y Recreación.** Cero deuda, estado siempre limpio.

## 🚀 Instalación y Uso (Docker First)

`nginx-cmd` está diseñado para correr estrictamente dentro de su contenedor Docker (Alpine con soporte Brotli y ACME). 

### 1. Despliegue con Docker Compose
Clona el repositorio, configura tu `.env` (basado en `.env.example`) y levanta:
```bash
docker compose up -d --build
```

### 2. Ejecutar Comandos desde el Host
Para administrar dominios sin entrar al contenedor (`docker exec`), usa el script wrapper incluido en la raíz de este proyecto:

```bash
# Dar permisos (si no los tiene)
chmod +x nginx-cmd.sh

# Listar dominios
./nginx-cmd.sh list

# Agregar un dominio (Reverse Proxy)
./nginx-cmd.sh add midominio.com proxy --target 10.0.0.5:8080
```

*Próximamente documentación detallada sobre auto-escalado TLS en [USAGE.md](USAGE.md).*
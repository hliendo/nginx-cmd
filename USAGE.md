# Manual de Uso: nginx-cmd (Árbol de Dominios)

`nginx-cmd` es el orquestador estricto de la arquitectura CDC para Nginx. Su función es inyectar configuraciones basadas en **dos únicos paradigmas de red**: **SPA** (Frontend Estático) y **Proxy** (Intermediario de Backends).

Al momento de ejecutar `nginx-cmd add <dominio> <tipo>`, el binario inyecta el árbol de cabeceras ('Includes') predefinidas en el servidor.

---

## 🌳 Árbol Estructural de nginx-cmd

```text
nginx-cmd
│
├── 1. SPA (Single Page Application)
│   └── Comando: nginx-cmd add dominio.com spa
│   └── (Sin Switches Adicionales)
│
└── 2. Proxy (Reverse Proxy)
    ├── Comando: nginx-cmd add dominio.com proxy --target <IP:PUERTO>
    │
    ├── [+ Switch Opcional] --ws     (Soporte para WebSockets)
    ├── [+ Switch Opcional] --grpc   (Soporte para Protocolo gRPC)
    └── [+ Switch Opcional] --stream (Buffering-Off para SSE/Video)
```

---

## 1. Sitios Tipo: SPA (Single Page Application)
**Uso típico:** Sitios web estáticos desarrollados en React, Vue, Svelte, o simple HTML/JS.
**Comando Base:** `nginx-cmd add <dominio.com> spa`

Un SPA requiere servidor web clásico. Nginx funciona como "Servidor de Archivos", leyendo los `.html` desde el disco duro del contenedor (`/opt/nginx/html/<dominio>`) e implementando la lógica de enrutamiento necesaria para frameworks JS (`try_files $uri /index.html;`).

### ⚙️ Switches Disponibles (SPA)
**NINGUNO.**
- **¿Por qué?** Porque en un SPA la comunicación cliente-servidor se da exclusivamente por peticiones HTTP GET a un archivo estático en disco duro local. Nginx *no reenvía* el tráfico a un tercero. Por lo tanto, switches de tunelización como WebSocket, gRPC o Chunking (Streaming) no tienen entidad lógica porque no hay un "Backend" conectado en tiempo real al `location /` raíz de Nginx.

---

## 2. Sitios Tipo: Proxy (Reverse Proxy)
**Uso típico:** APIs en Node.js, Python, Rust, Go, bases de datos expuestas, o servicios como Portainer, Gitea, etc.
**Comando Base:** `nginx-cmd add <dominio.com> proxy --target <IP:PORT>`

Un Proxy requiere que Nginx funcione como un "Túnel" ciego y delegador. Nginx recibe el tráfico de internet, y lo re-ensambla hacia una IP interna de nuestra red.

### ⚙️ Switches Obligatorios
- `--target <HOST:PORT>` (Ej: `--target 10.10.10.200:8080`)
  - **Función:** Le indica a Nginx hacia qué IP y puerto exacto enviar el tráfico re-ensamblado.

### ⚙️ Switches Opcionales (Features inyectadas como Includes)

El tipo Proxy admite tres (3) modificadores de comportamiento (switches) para adaptar Nginx al requerimiento técnico del backend al que conectamos.

#### A. Switch `--ws` (WebSocket Support)
- **Función:** Inyecta en Nginx el archivo `proxy-websocket.conf`.
- **Efecto Técnico:** Fuerza a Nginx a interceptar el HTTP Request del cliente, buscar el campo `Upgrade: websocket` y transformar la petición de "Pregunta-Respuesta (HTTP)" en una **"Tubería Bidireccional de Larga Duración"** inyectando Cabeceras.
- **¿Por qué?** Porque si no se envía este "Upgrade Header", el servidor Nginx corta el túnel pensando que el request inicial HTTP finalizó la comunicación. Fundamental para chats o sincronización P2P en tiempo real.

#### B. Switch `--grpc` (Protocolo gRPC Support)
- **Función:** Activa el enrutamiento y proxy-pass de gRPC (`proxy-grpc.conf`).
- **Efecto Técnico:** Reemplaza el túnel tradicional `http://` por `grpc://`. Adjudica el soporte estricto de HTTP/2 que exige gRPC y omite cabeceras de proxy obsoletas que corromperían los paquetes Protobuf de Google (usados en alta sincronización entre microservicios).
- **¿Por qué?** Porque un backend escrito en gRPC no entiende HTTP genérico; lee binario directo empaquetado. Nginx necesita una orden quirúrgica para dejar de comportarse como un servidor Web y convertirse en un orquestador gRPC ciego.

#### C. Switch `--stream` (Server-Sent Events / Chunking)
- **Función:** Desactiva el buffering de red (`proxy-stream.conf`).
- **Efecto Técnico:** Instruye a Nginx (`proxy_buffering off;`) para que no intente guardar la respuesta del backend en su RAM hasta empaquetarla toda junta. Todo byte que Nginx recibe del Backend, lo empuja instantáneamente al navegador del usuario final. Además amplía los Timeouts al máximo (3600s).
- **¿Por qué?** Nginx por defecto es "Ansioso": si le pedís una película de 5GB, intentará descargar los 5GB en su memoria y **recién después enviárselo al usuario**. Con este switch prendido, si tenés APIs que avisan eventos gota a gota (SSE, ChatGPT, Streaming de video/IP-Cams), Nginx libera el freno central.

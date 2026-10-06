# Finance App

Aplicación personal de gestión financiera en Go, desarrollada por etapas.
La etapa 1 contiene únicamente la estructura base y un servidor HTTP con
`net/http`, sin dependencias externas ni almacenamiento implementado.

## Iniciar el servidor

Desde la carpeta `finance-app`, con Go 1.27 o posterior:

```sh
go run ./cmd/server
```

El puerto predeterminado es `8080`. Para cambiarlo, configurar la variable
de entorno `APP_PORT` antes de iniciar el servidor. En PowerShell:

```powershell
$env:APP_PORT = "9090"
go run ./cmd/server
```

## Health check

`GET http://localhost:8080/health` responde HTTP 200 con
`Content-Type: application/json` y el cuerpo:

```json
{"status":"ok"}
```

El servidor registra su inicio y cierre, y se apaga de forma ordenada al recibir
SIGINT (Ctrl+C) o SIGTERM. La carpeta `data/` queda reservada para
`finanzas.xlsx` en una etapa posterior; el archivo todavía no se crea.

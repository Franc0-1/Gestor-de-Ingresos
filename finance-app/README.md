# Finance App

Aplicación personal de gestión financiera en Go, desarrollada por etapas.
El servidor utiliza `net/http`. La etapa 2 incorpora modelos y repositorios
Excel con `github.com/xuri/excelize/v2`, sin endpoints financieros todavía.

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
SIGINT (Ctrl+C) o SIGTERM. Al iniciar se crea `data/finanzas.xlsx` si no existe,
con las hojas `Cuentas`, `Categorias` y `Movimientos`.

## Repositorios Excel

`repository.NewExcelRepository(path)` expone `Accounts`, `Categories` y
`Transactions`, cada uno con `Create`, `GetAll`, `GetByID`, `Update` y `Delete`.
Los modelos se validan antes de escribir; no se verifica aún la existencia
de las cuentas o categorías referenciadas por un movimiento.

Los montos son `int64` en unidad monetaria mínima: `$10.500,25` se guarda
como `1050025`. IDs y montos se almacenan en celdas de texto decimal para
evitar pérdida de precisión. Los booleanos usan `true`/`false` y las fechas
RFC3339Nano. `Create` asigna el ID y completa timestamps vacíos; `Update`
conserva `CreatedAt` y renueva `UpdatedAt`.

Una hoja oculta `_Metadata` conserva los últimos IDs asignados, incluso
después de eliminar registros o reabrir el archivo. No eliminar esa hoja.
Las operaciones usan `sync.RWMutex` compartido por ruta dentro del proceso;
no se admite escritura simultánea desde otros procesos o Excel.
El archivo se abre y cierra por operación, con reemplazo mediante archivo
temporal al guardar. Los tests utilizan exclusivamente `t.TempDir()`.

```sh
go test ./...
go build ./...
```

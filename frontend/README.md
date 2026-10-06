# Mis Finanzas · Frontend

React + TypeScript estricto + Vite, con CSS propio. Consume la API existente
y no incluye datos ficticios.
Los temas claro y oscuro se cambian desde la navegación y la preferencia se
guarda en `localStorage`. La primera visita usa el tema del sistema. Todos los
componentes comparten tokens de color, contraste y espaciado; no se agregaron
dependencias de diseño.
`/movimientos` permite listar, buscar, filtrar,
ordenar, paginar y gestionar ingresos/gastos con la API existente.
`/transferencias` permite gestionar transferencias entre cuentas con búsqueda,
filtros, orden y paginación del backend, y confirmación antes de eliminar.

Desde `finance-app/`, iniciar el backend en una terminal:

```powershell
go run ./cmd/server
```

Desde `frontend/`, en otra terminal:

```powershell
npm.cmd install
npm.cmd run dev
```

Abrir http://localhost:5173. En PowerShell se usa `npm.cmd` para evitar
restricciones de ejecución de `npm.ps1`. En otras shells puede usarse `npm`.

Vite envía `/api` a http://localhost:8080. Para otro puerto del backend,
crear `frontend/.env.local` con `API_PROXY_TARGET=http://localhost:9090`.
El proxy también está configurado para `npm.cmd run preview`. Un despliegue
de `dist/` debe servir `/api` desde el mismo origen mediante un proxy.

```powershell
npm.cmd run build
```

El build verifica TypeScript antes de generar `dist/`. Los importes se
decodifican como `bigint` y se formatean con `Intl.NumberFormat`, sin perder
precisión en int64. Para importes fuera del rango seguro de JavaScript se
requiere un navegador con soporte del contexto `source` de `JSON.parse`;
los navegadores anteriores muestran un error en lugar de redondearlos.
Las monedas se mantienen separadas. Los movimientos de cuentas inactivas
o eliminadas muestran «moneda sin identificar» si el dashboard no incluye
la cuenta. Las fechas se muestran en UTC, igual que los filtros del backend.

Los controles de período consultan `GET /api/dashboard?from=…&to=…`.
«Este mes» consulta el período predeterminado del backend. Al volver al
Dashboard se consultan nuevamente los balances, incluyendo transferencias
creadas, editadas o eliminadas. Cuentas y Categorías quedan para próximas etapas.

En Movimientos, «Nuevo movimiento» abre el formulario y las acciones de
cada fila permiten editar o eliminar con confirmación. El monto se escribe
sin separadores de miles, con coma o punto decimal y hasta dos decimales.
Solo las cuentas activas se ofrecen en el formulario; el historial muestra
también cuentas inactivas. Los filtros y la paginación se procesan en Go.

En Transferencias se eligen dos cuentas activas diferentes. El backend permite
monedas distintas y registra el mismo importe en ambas cuentas sin conversión;
el formulario y el historial lo indican cuando corresponde. El historial
conserva los nombres de cuentas inactivas disponibles en la API.

Los tests de importes y API usan únicamente el runner de Node (Node 22.6+
con soporte para eliminar tipos TypeScript): `npm.cmd test`.

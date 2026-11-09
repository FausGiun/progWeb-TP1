# Running Tracker - Trabajo de Cursada TP4

Registro de zapatillas y entrenamientos con **HTML, CSS y JavaScript**, una **API REST en Go**, **PostgreSQL** y acceso a datos generado por **sqlc**. La interfaz permite listar, crear y eliminar ambas entidades sin recargar la página.

## Organización

- `static/index.html`: formularios, contenedores de listados y tabla del historial.
- `static/app.js`: `fetch`, altas JSON, actualización del DOM y borrado con delegación de eventos.
- `static/styles.css`: estilos, Grid para la distribución principal y Flexbox para alineaciones.
- `main.go`: conexión, rutas, archivos estáticos y arranque del servidor.
- `handlers_zapatillas.go`, `handlers_entrenamientos.go`: operaciones de cada recurso.
- `handlers.go`: respuestas, lectura de IDs y transacciones compartidas.
- `middleware.go`: registro de las peticiones.
- `logic/`: validaciones y cálculos del dominio.
- `db/schema`, `db/queries`, `db/sqlc`: esquema, consultas y código generado.

Las pantallas anteriores de `templates/` se conservan como antecedente; la interfaz de TP4 es `static/index.html`.

## Requisitos y configuración

Go **1.26.6** (versión de `go.mod`), sqlc, Docker con Compose, Bash, curl y, para los atajos, make. Los comandos se ejecutan desde `ProgWeb`; en Windows puede usarse Git Bash con Docker Desktop iniciado.

Si todavía no existe `.env`, copiar la plantilla:

```bash
cp .env.example .env
```

Configurar en `.env` el usuario, contraseña, base y puerto de PostgreSQL. `DATABASE_URL` debe representar esos mismos datos. La plantilla contiene valores locales de ejemplo; `.env` queda excluido de Git. Make exporta la configuración al proceso Go y Compose la utiliza para PostgreSQL.

## Ejecutar la aplicación

```bash
make run
```

Este comando genera sqlc, compila todos los archivos Go del paquete, inicia PostgreSQL y arranca el servidor. Abrir **http://localhost:8080/**. Ctrl+C detiene los servicios y conserva el volumen con los datos.

Alternativa en Git Bash si no se dispone de make:

```bash
set -a
. ./.env
set +a
sqlc generate
docker compose up -d
# Esperar a que PostgreSQL acepte conexiones antes del último comando.
docker compose exec database pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"
go run .
```

Al finalizar esa alternativa, detener Go con Ctrl+C y ejecutar `docker compose down` para detener la base conservando los datos.

## Comprobar el TP4 desde el navegador

1. Agregar una zapatilla con kilómetros iniciales. Debe aparecer en el listado y en el selector del entrenamiento.
2. Registrar una salida eligiendo esa zapatilla. Deben aparecer el historial, el ritmo en **min/km**, las calorías y el kilometraje actualizado.
3. Eliminar el entrenamiento. Debe desaparecer de la tabla y descontarse su distancia del par utilizado.
4. Eliminar la zapatilla. Debe salir del listado activo y del selector. Si tiene entrenamientos, éstos conservan la referencia al par retirado.
5. Comprobar la lista vacía, campos obligatorios, distancia positiva y fecha no futura.
6. En las herramientas de red del navegador, verificar los `GET`, `POST` y `DELETE`; el documento no se recarga al usar los formularios.

El ritmo y las calorías se calculan en Go. El ritmo también se recalcula al leer para mostrar correctamente registros anteriores que lo almacenaban con otra unidad.

El kilometraje representa **kilómetros iniciales + distancias de los entrenamientos registrados**. Corregir distancia o zapatilla con la API ajusta los totales; borrar un entrenamiento descuenta su distancia. Cada operación se confirma por completo o se revierte mediante una transacción. Una zapatilla retirada conserva su historial y no admite nuevas salidas.

## Pruebas automatizadas

```bash
make test-api
# Alternativa sin make:
bash ./test-api.sh
```

`test-api.sh` usa exclusivamente el proyecto Docker `progweb-tp4-tests`, con puertos separados (por defecto **55432** para la base y **18080** para HTTP). Genera y compila, ejecuta `go test -p 1 -count=1 -v ./...`, levanta la API y comprueba las peticiones de `requests.bash`. Al terminar detiene su servidor y elimina solamente los contenedores y el volumen de pruebas. La base de desarrollo no se reinicia.

Las pruebas cubren CRUD, validaciones, ritmo, archivos de la interfaz, correcciones de kilometraje, cambio de zapatilla, rechazo de pares retirados y reversión completa ante un fallo de escritura. `requests.bash` comprueba los códigos HTTP reales y usa los IDs devueltos por la API.

También puede ejecutarse `bash ./requests.bash` contra un servidor ya iniciado. Crea sus propios registros de ejemplo y los elimina al completar el recorrido. `API_URL` permite elegir otra dirección.

El permiso ejecutable de `test-api.sh` debe conservarse al clonar el repositorio. Si se copia como archivo en Linux, se puede restaurar con `chmod +x test-api.sh`. `.gitattributes` mantiene los finales de línea LF de los scripts.

## Relación con el material de la cátedra

Los números siguientes son **páginas del PDF**, que pueden diferir del número de diapositiva impreso. Los documentos están en la carpeta superior de este espacio de trabajo.

| Implementación | Material local | Ubicación |
| --- | --- | --- |
| Formularios, listados, fetch GET/POST/DELETE, refresco y delegación | `tp4.pdf` | pp. 2-3, sección Trabajo de Cursada |
| Selección, creación de nodos, texto y eventos del DOM | `08-view.pdf` | pp. 21, 24, 26-27 |
| Grid y Flexbox | `08-view.pdf` | pp. 71-81 |
| JSON a HTML, `textContent`, carga inicial y eventos delegados | `10-view-dynamic.pdf` | pp. 35, 44-45 |
| Archivos estáticos, separación de handlers y configuración con `os.Getenv` | `09-view-static.pdf` | pp. 22-23, 33, 36-37 |
| `.env`, variables de Compose y exclusión de credenciales de Git | `03-docker.pdf` | pp. 17 y 20 |
| Operaciones atómicas con Begin/Commit/Rollback | `12-diseño.pdf` | pp. 50-53 |
| Permisos del script, `.env.example` y separación de handlers | Devolución del docente del 28 de octubre | Captura proporcionada en el chat |

Las reglas de kilometraje y la corrección de min/km son adaptaciones al dominio de running del código existente. Se conservaron las tecnologías del proyecto, sin agregar dependencias.

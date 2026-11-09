# Runner Log - Tracker - Trabajo de Cursada TP5

Aplicación Go con vistas **templ**, acceso a datos mediante **sqlc** y **PostgreSQL**. Permite registrar zapatillas y entrenamientos y eliminar registros mediante formularios HTML.

## Cómo funciona

1. El navegador pide una página con GET.
2. El handler consulta la base con sqlc y renderiza el HTML con templ.
3. El usuario envía un formulario con POST.
4. Go usa `r.ParseForm()`, valida y guarda los datos.
5. `http.Redirect` devuelve una redirección 303 y el navegador pide el listado actualizado con GET. Es el patrón **Post/Redirect/Get (PRG)**.

Las páginas funcionan sin JavaScript. `/` muestra el formulario y el historial de entrenamientos. `/mis-zapatillas` tiene el formulario y listado de calzado. La dirección anterior `/mis-entrenamientos` redirige a `/`.

## Archivos principales

| Archivo | Responsabilidad |
| --- | --- |
| `main.go` | Conexión a la base, rutas y arranque. |
| `handlers_entrenamientos.go` | Vista, alta y borrado de entrenamientos. |
| `handlers_zapatillas.go` | Vista, alta y baja de zapatillas. |
| `handlers.go` | Errores, lectura del ID de los formularios y transacciones ya utilizadas en TP4. |
| `views/layout.templ` | HTML base, navegación, CSS y pie. |
| `views/entity_form.templ` | Los dos formularios. |
| `views/entity_list.templ` | Listado de zapatillas y tabla de entrenamientos. |
| `views/index.templ` | Composición de las dos páginas. |
| `static/styles.css` | Diseño de las pantallas. |
| `logic/running.go` | Validaciones, ritmo en min/km y calorías. |
| `db/schema`, `db/queries`, `db/sqlc` | Esquema, consultas y código generado por sqlc. |

`templ generate` crea los archivos `views/*_templ.go`, que se incluyen como código generado. Se regeneran antes de compilar: los archivos que se editan son los `.templ`. Las vistas anteriores de `templates/` se conservan como antecedente; las vistas activas están en `views/`. La entrega con HTML y JavaScript del TP4 está en la rama `tp4`.

## Preparación

Se necesitan Go **1.26.6**, sqlc, Docker con Compose, Bash y curl. Make es opcional. Instalar el generador de la misma versión que figura en `go.mod`:

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1100
```

Crear `.env` si todavía no existe:

```bash
test -f .env || cp .env.example .env
```

La plantilla contiene valores locales de ejemplo. `DATABASE_URL` debe coincidir con el usuario, contraseña, base y puerto configurados para PostgreSQL. `.env` queda excluido de Git.

## Ejecutar

Con Docker iniciado, desde la carpeta del proyecto:

```bash
make run
```

Make genera sqlc y templ, compila el paquete completo e inicia PostgreSQL y el servidor. Abrir **http://localhost:8080/**. Ctrl+C detiene los servicios conservando los datos.

Alternativa sin make, desde Git Bash en Windows:

```bash
set -a
. ./.env
set +a
sqlc generate
templ generate
docker compose up -d --wait
go run .
```

En esa alternativa, Ctrl+C detiene Go. Después se puede ejecutar `docker compose down` para detener PostgreSQL conservando su volumen.

## Comprobar el TP5

1. Abrir Zapatillas y agregar un par. El formulario hace POST y vuelve a su listado mediante una redirección.
2. Entrar a Entrenamientos: el par ya aparece en el selector incluido en el HTML.
3. Registrar 10 km en 50 minutos. Deben aparecer 5 min/km, 650 kcal y la salida nueva después de la recarga completa.
4. Actualizar la página: la salida no se vuelve a crear, porque la última petición es GET.
5. Eliminar la salida mediante su formulario. Se descuenta su distancia de la zapatilla.
6. Eliminar una zapatilla la retira del listado activo y conserva las referencias del historial.

Las validaciones se hacen en Go. Los errores de entrada devuelven 400 y se puede volver al formulario con el botón Atrás del navegador. Los cambios de entrenamiento y kilometraje usan una misma transacción para evitar escrituras parciales.

## Pruebas

```bash
make test
# Alternativa sin make:
bash ./test-api.sh
```

El script conserva su nombre anterior, pero prueba la aplicación web del TP5. Genera el código, ejecuta las pruebas de Go e inicia una instancia para comprobar formularios y redirecciones con curl. Usa el proyecto Docker `progweb-tp5-tests`, con los puertos de pruebas configurados en `.env`. Al finalizar elimina solamente los contenedores y el volumen de ese proyecto temporal.

Se comprueban HTML con datos reales, separación de páginas, formularios válidos e inválidos, redirecciones, escape del texto ingresado, kilometraje, bajas y reversión ante fallos. `requests.bash` se ejecuta desde ese script contra la instancia temporal.

## Material de la cátedra utilizado

Los números son páginas del PDF. Los documentos están en la carpeta superior de este espacio de trabajo.

| Punto | Material local | Ubicación |
| --- | --- | --- |
| Componentes, consulta con sqlc, formularios PRG y automatización | `tp5.pdf` | p. 3, Trabajo de Cursada |
| Componentes tipados y renderizado | `09-view-static.pdf` | pp. 15-24 y 33-35 |
| CSS servido como archivo estático | `09-view-static.pdf` | p. 22 |
| Conexión y configuración | `09-view-static.pdf` | pp. 36-37 |
| PostgreSQL con Docker y `.env` | `03-docker.pdf` | pp. 17-20 |
| Transacciones conservadas del TP4 | `12-diseño.pdf` | pp. 50-53 |

Se mantiene el dominio de running y la base PostgreSQL que ya tenía el proyecto. El borrado por formulario conserva una función del TP4; el objetivo central del TP5 es listar y crear mediante renderizado en el servidor.

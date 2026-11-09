# Runner Log - Tracker - Trabajo de Cursada TP5

Entrega en la rama **`tp5`**. Aplicación Go con vistas **templ**, consultas **sqlc** y **PostgreSQL**. Registra zapatillas y entrenamientos mediante formularios HTML, con renderizado del lado del servidor y Post/Redirect/Get (PRG).

## Requisitos

- Go **1.26.6** (versión declarada en `go.mod`).
- Docker iniciado y Docker Compose v2 con soporte de `up --wait`.
- Bash. En Windows se puede usar **Git Bash**.
- Los generadores `sqlc` y `templ` disponibles en `PATH`.
- Make es opcional. Para las pruebas HTTP también se necesita curl.

Instalación de los generadores, si todavía no están instalados:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
go install github.com/a-h/templ/cmd/templ@v0.3.1100
```

El directorio de ejecutables de Go (`go env GOPATH`, subcarpeta `bin`) debe estar en `PATH`.

## Ejecutar después de clonar

Desde la carpeta del proyecto, en Bash:

```bash
bash ./run-app.sh
```

Alternativa con Make:

```bash
make run-app
```

`make run` es un alias del mismo ciclo. Abrir **http://localhost:8080/** con la configuración de ejemplo. Presionar **Ctrl+C** para detener la aplicación y completar la limpieza.

Si falta `.env`, el script lo crea desde `.env.example`. Sus valores son ejemplos locales. Si se modifica el usuario, la contraseña, la base o `DB_PORT`, también debe actualizarse `DATABASE_URL`. `HTTP_ADDR` configura dónde escucha Go. `.env` está excluido de Git.

### Ciclo completo de la entrega

El mismo script realiza todas las tareas:

1. Verifica las herramientas y carga la configuración.
2. Elimina contenedores y volúmenes anteriores del proyecto Docker **`progweb-tp5-app`**.
3. Ejecuta **`sqlc generate`** y **`templ generate`**.
4. Compila todo el paquete Go en `tmp/runner-log` (`.exe` en Windows).
5. Levanta PostgreSQL y espera el resultado de su **healthcheck**, con un límite de 60 segundos.
6. Ejecuta la aplicación y espera su finalización.
7. Al salir normalmente, con Ctrl+C o por un error posterior a la preparación, detiene el proceso y elimina los contenedores y volúmenes de **ese mismo proyecto Docker**.

**La base de esta entrega empieza vacía en cada ejecución y sus datos se eliminan al salir**, como exige la consigna. La limpieza usa un nombre de proyecto explícito; no elimina los recursos de otros proyectos Docker. Para una limpieza manual del mismo entorno: `make clean`.

Los puertos de ejemplo son 8080 para HTTP y 5432 para PostgreSQL. Deben estar disponibles o cambiarse en `.env`. Los números de puertos de pruebas son elecciones de esta implementación, no una exigencia de la cátedra.

## Cómo funciona

1. El navegador pide una página con **GET**.
2. El handler consulta PostgreSQL con sqlc y renderiza el HTML con templ.
3. El navegador envía un formulario nativo con **POST**.
4. Go usa `r.ParseForm()`, convierte y valida los campos, y guarda los datos.
5. Devuelve **303 See Other** con `http.Redirect`. El navegador hace otro GET y recibe el listado actualizado. Así recargar la página de resultado no vuelve a crear el registro.

**No hay archivos JavaScript ni dependencia de JavaScript para usar la aplicación.** La entrega anterior con HTML y JavaScript está en la rama `tp4`. Las vistas del TP5 están en `views/`, y el único archivo estático activo es `static/styles.css`.

### Rutas

| Método y ruta | Operación |
| --- | --- |
| `GET /` | Formulario e historial de entrenamientos. |
| `GET /mis-zapatillas` | Formulario y listado de zapatillas activas. |
| `GET /mis-entrenamientos` | Redirección a `/`, para la dirección anterior. |
| `POST /zapatillas` | Alta de zapatilla y redirección a su página. |
| `POST /entrenamientos` | Alta de salida y redirección al inicio. |
| `POST /zapatillas/eliminar` | Baja lógica del par y redirección. |
| `POST /entrenamientos/eliminar` | Borrado de salida y redirección. |
| `GET /static/styles.css` | Hoja de estilos. |

Los formularios no envían JSON. Una entrada inválida devuelve 400; un registro inexistente, 404; un método no permitido, 405; un fallo interno, 500. Los errores de carga se muestran como texto y se puede volver al formulario con Atrás.

### Reglas del registro

- La zapatilla admite kilómetros previos. Crear una salida suma su distancia; borrarla la descuenta.
- El alta o borrado de la salida y el cambio de kilometraje usan **una misma transacción**. Un fallo revierte ambas escrituras.
- El ritmo se calcula como minutos/distancia, en **min/km**. Las calorías son una estimación fija de distancia × 65.
- Eliminar un par cambia `estado` a falso: conserva su fila e historial, pero lo quita de los activos y no permite nuevas salidas con él.
- Si no hay zapatillas activas, el formulario de entrenamiento queda deshabilitado. Go vuelve a validar aunque se envíe una petición manual.

## Archivos principales

| Archivo | Responsabilidad |
| --- | --- |
| `run-app.sh` / `Makefile` | Ciclo de ejecución y limpieza exigido para la entrega. |
| `main.go` | Conexión a la base, rutas y arranque. |
| `handlers_entrenamientos.go` | GET, lectura del formulario, alta y borrado de salidas. |
| `handlers_zapatillas.go` | Vista, alta y baja de zapatillas. |
| `handlers.go` / `middleware.go` | Errores, ID, transacciones y log de peticiones. |
| `views/layout.templ` | Documento base, navegación, CSS y pie. |
| `views/entity_form.templ` | Formularios HTML. |
| `views/entity_list.templ` | Listados y formularios de borrado. |
| `views/index.templ` | Composición de las dos páginas. |
| `logic/running.go` | Validaciones, ritmo y calorías. |
| `db/schema`, `db/queries`, `db/sqlc` | Esquema, consultas y código generado por sqlc. |
| `test-app.sh`, `requests.bash`, `*_test.go` | Pruebas de la aplicación web. |

`templ generate` crea los `views/*_templ.go`, incluidos en Git. `sqlc generate` obtiene el código de acceso a datos. Se editan las fuentes `.templ` y `.sql`, no los archivos generados. Las plantillas y estilos antiguos sin uso se quitaron de esta rama.

## Comprobación manual para la defensa

1. Ejecutar el comando de entrega y abrir Zapatillas. Agregar un par con 12,5 km iniciales.
2. Ir a Entrenamientos y registrar 10 km en 50 minutos con ese par.
3. Comprobar 5 min/km y 650 kcal. En Zapatillas deben aparecer 22,5 km.
4. Recargar: la salida no se vuelve a crear porque la última petición es GET.
5. Borrar la salida: el par debe volver a 12,5 km.
6. Retirar el par: deja de aparecer en los activos.
7. Finalizar con Ctrl+C: se eliminan los contenedores y el volumen de la entrega. Al volver a ejecutar comienza una base nueva.

## Pruebas

Después de disponer de `.env` (lo crea `run-app.sh` si falta):

```bash
bash ./test-app.sh
# Alternativa:
make test
```

Estas pruebas comprueban **HTML, formularios y redirecciones**, no la API JSON anterior. El script genera código, compila, ejecuta los tests de Go e inicia una instancia para probar 8 peticiones HTTP con curl.

Usa el proyecto Docker **`progweb-tp5-tests`**, con puertos independientes (`TEST_DB_PORT` y `TEST_HTTP_PORT`, por defecto 55432 y 18080). Al terminar elimina su proceso, contenedores y volumen temporal. Los tests limpian las tablas de esa base de pruebas.

Se verifican datos reales en el HTML, escape del texto ingresado, separación de páginas, validaciones, PRG, ritmo/calorías, kilometraje, bajas y rollback ante un fallo. No se deben ejecutar los tests contra una base con datos que se quieran conservar.

## Material de la cátedra utilizado

Los números son páginas del PDF; los archivos están en la carpeta superior de este espacio de trabajo.

| Tema | Material local | Páginas |
| --- | --- | --- |
| Componentes, sqlc, formularios PRG y automatización | `tp5.pdf`, Trabajo de Cursada | 3 |
| Componentes tipados y renderizado | `09-view-static.pdf` | 15-24 y 33-35 |
| Archivos estáticos | `09-view-static.pdf` | 22 |
| Conexión y configuración | `09-view-static.pdf` | 36-37 |
| Docker, variables y comandos de limpieza | `03-docker.pdf` | 17-20 |
| Transacciones conservadas del TP4 | `12-diseño.pdf` | 50-53 |

El ciclo con limpieza de contenedores y volúmenes antes y después, el nombre de rama y la ausencia de `.js` siguen además las condiciones de entrega proporcionadas en el chat.

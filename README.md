# Runner Log - Tracker (TP6)

Registro de entrenamientos y kilometraje de zapatillas. El TP6 usa Go, templ y PostgreSQL. HTMX actualiza los listados al guardar o borrar un registro.

## Organización del código

- `web/` recibe las peticiones y llama a `application/`. `views/` genera la página completa o el fragmento para HTMX.
- `application/` define los casos de uso y las interfaces que necesita para guardar y consultar datos.
- `logic/` reúne los modelos, las validaciones y los cálculos de ritmo y calorías.
- `db/postgres/` ejecuta las consultas de sqlc y maneja las transacciones. `main.go` conecta las capas.

## Requisitos

- **Go** 1.26.6, según `go.mod`.
- **Docker** y **Docker Compose**
- **sqlc** `v1.31.1` y **templ** `v0.3.1100` en el `PATH`.
- **Bash** y, opcionalmente, **make**. Las pruebas HTTP usan **curl**.
- Conexión a Internet desde el navegador para cargar HTMX 1.9.10 desde el CDN.

## Cómo ejecutar

Desde la raíz del proyecto, con Docker iniciado, ejecutá este comando en Bash (Git Bash en Windows).

```bash
bash ./run-app.sh
```

También podés usar `make run-app`. El script crea `.env` desde `.env.example` si hace falta, genera el código, compila y espera a PostgreSQL. Abrí <http://localhost:8080/> y cargá primero un par en **Zapatillas**.

La base de la entrega arranca vacía. Al presionar `Ctrl+C`, el script detiene el servidor y elimina el contenedor y el volumen de la app.

## Pruebas

Con `.env` presente, ejecutá las pruebas con este comando o con `make test`. Usan un proyecto de Docker separado y eliminan su volumen al terminar.

```bash
bash ./test-app.sh
```

Las pruebas de Go cubren los formularios comunes y las transacciones. `requests.bash` comprueba las altas y bajas de HTMX sobre el servidor en ejecución.

`make generate` regenera el código de sqlc y templ. `make clean` elimina los recursos Docker de la app y su binario temporal.

## Rutas

| Método | Ruta | Acción | Respuesta |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Obtiene zapatillas y entrenamientos de la BD y renderiza `views.IndexPage` | `200 OK` (`text/html`) |
| `GET` | `/mis-zapatillas` | Muestra el formulario y el calzado activo | `200 OK` (`text/html`) |
| `POST` | `/zapatillas` | Guarda un par | `200` con el listado para HTMX o `303` a `/mis-zapatillas` |
| `POST` | `/entrenamientos` | Guarda un entrenamiento y suma sus kilómetros | `200` con el historial para HTMX o `303` a `/` |
| `DELETE` | `/zapatillas/{id}` | Retira un par del listado activo | `200` con cuerpo vacío |
| `DELETE` | `/entrenamientos/{id}` | Borra un entrenamiento y descuenta sus kilómetros | `200` con cuerpo vacío |
| `POST` | `/zapatillas/eliminar` | Retira un par desde un formulario común | `303` a `/mis-zapatillas` |
| `POST` | `/entrenamientos/eliminar` | Borra una salida desde un formulario común | `303` a `/` |

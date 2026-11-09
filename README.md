# Running Tracker - Vistas Dinámicas con Templ y SQLc (TP5)

Aplicación web desarrollada en **Go** para el registro de entrenamientos de running y control del desgaste de calzado deportivo.

En esta etapa (**TP5**), la arquitectura implementa **Server-Side Rendering (SSR)** integrando la capa de presentación directamente en el servidor Go mediante **Templ** (componentes HTML tipados y compilados) y **SQLc** para la persistencia en PostgreSQL. El manejo de formularios implementa el patrón clásico **Post/Redirect/Get (PRG)**.

## Arquitectura del Proyecto

El proyecto se organiza en capas bien definidas siguiendo el modelo de 3 capas (Presentación, Lógica de Negocio y Acceso a Datos):

- **`views/` (Capa de Presentación - SSR con Templ):**
  - `layout.templ`: Estructura base del documento HTML (`<!DOCTYPE html>`, `<head>`, `<body>`) estilizada con una hoja de estilo propia (`/static/styles.css`).
  - `index.templ`: Página principal (`IndexPage`) que compone el layout junto con los formularios y las tablas de datos.
  - `entity_form.templ`: Componentes de formularios HTML clásicos (`POST /zapatillas` y `POST /entrenamientos`).
  - `entity_list.templ`: Componentes que reciben slices tipados de `db.Zapatilla` y `db.Entrenamiento` y renderizan las tablas de registros con acciones de eliminación.
- **`main.go` y `handlers*.go` (Controladores HTTP):** Reciben las peticiones del navegador, consultan la base de datos con `sqlc`, validan las reglas de negocio, renderizan los componentes de `templ` en `GET /` y aplican redirecciones `303 See Other` tras procesar formularios `POST` (Patrón PRG).
- **`logic/` (Capa de Lógica de Negocio):** Funciones puras en Go (`ValidateZapatilla`, `ValidateEntrenamiento`, `CalcularRitmo`, `CalcularCalorias`) sin dependencias externas.
- **`db/` (Capa de Acceso a Datos):** Esquema SQL (`db/schema/schema.sql`), consultas (`db/queries/queries.sql`) y código Go seguro generado por `sqlc`.

## Requisitos Previos

- **Go** (1.22+)
- **Docker** y **Docker Compose**
- **sqlc**: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **templ**: `go install github.com/a-h/templ/cmd/templ@latest`
- **make**

## Ejecución y Flujo de Trabajo (Makefile)

El proyecto incluye un `Makefile` que automatiza la generación de código de `sqlc` y `templ`, la compilación del servidor y la gestión del contenedor de PostgreSQL en Docker.

### 1. Ejecutar el Servidor

Levanta el contenedor de PostgreSQL, espera a que acepte conexiones, ejecuta `sqlc generate` y `templ generate`, compila el binario e inicia el servidor en `http://localhost:8080`:

```bash
make run
```

Una vez iniciado, abrir en el navegador:
👉 **`http://localhost:8080`**

*(Al presionar `Ctrl + C`, el servidor se detiene y apaga el contenedor de Docker automáticamente).*

### 2. Ejecutar Pruebas Automatizadas

Inicia una base de datos limpia en Docker, ejecuta las pruebas unitarias de lógica de negocio (`logic/`), las pruebas de integración de base de datos (`db/sqlc/`) y las pruebas de renderizado SSR y flujo PRG (`main_test.go`), eliminando los contenedores al terminar:

```bash
make test
```

### 3. Regenerar Código Manualmente (SQLc + Templ)

Si se modifican archivos `.sql` o `.templ`, se puede regenerar el código Go ejecutando:

```bash
make generate
```

### 4. Limpiar el Entorno

Detiene y elimina los contenedores y volúmenes de Docker, y borra los binarios temporales:

```bash
make clean
```

## Rutas Principales del Servidor SSR

| Método | Ruta | Acción | Respuesta |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Obtiene zapatillas y entrenamientos de la BD y renderiza `views.IndexPage` | `200 OK` (`text/html`) |
| `POST` | `/zapatillas` | Parsea el formulario, valida y guarda una nueva zapatilla | `303 See Other` (Redirect a `/mis-zapatillas`) |
| `POST` | `/zapatillas/eliminar` | Elimina una zapatilla desde el formulario de la tabla | `303 See Other` (Redirect a `/mis-zapatillas`) |
| `POST` | `/entrenamientos` | Parsea el formulario, calcula ritmo/calorías, guarda el entrenamiento y suma Kms | `303 See Other` (Redirect a `/`) |
| `POST` | `/entrenamientos/eliminar` | Elimina un entrenamiento desde el formulario de la tabla | `303 See Other` (Redirect a `/`) |

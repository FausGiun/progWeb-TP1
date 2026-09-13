# Runner Log - Tracker

Aplicación web para registrar entrenamientos de running y llevar el control del desgaste del equipamiento.

## Documentación del Proyecto (Persistencia)

El proyecto utiliza una base de datos relacional **PostgreSQL** para garantizar la persistencia de los datos.

La arquitectura de datos se compone de dos tablas principales:

- **zapatillas:** Almacena el calzado registrado, llevando un acumulador del kilometraje total de cada par.
- **entrenamientos:** Registra cada sesión de running, vinculándose con la tabla de zapatillas mediante una relación de clave foránea (`zapatilla_id`).

Para la interacción entre la aplicación en Go y la base de datos se utiliza sqlc. Esta herramienta genera código Go fuertemente tipado a partir de consultas SQL puras (`queries.sql`) y el esquema de la base de datos (`schema.sql`), evitando el uso de ORMs pesados y garantizando consultas seguras.

## Testing automatizado

Para evaluar el proyecto y correr los tests automatizados (que incluyen la creación de un entorno limpio con Docker, la ejecución de pruebas unitarias sobre las rutas HTTP y operaciones CRUD de la base de datos), simplemente ejecute:

` ` `bash
make test
` ` `
_Nota: Este comando requiere tener Docker y docker-compose instalados en el sistema._

## Cómo ejecutar la aplicación localmente

Si desea levantar la aplicación de forma manual:

1. Asegurarse de tener Go instalado en el equipo y una base de datos PostgreSQL corriendo.
2. Abrir una terminal en la carpeta raíz del proyecto.
3. Ordenar y descargar las dependencias del módulo ejecutando:
   ` ` `bash
go mod tidy
` ` `
4. Levantar el servidor ejecutando:
   ` ` `bash
go run main.go
` ` `
5. Abrir un navegador web e ingresar a: `http://localhost:8080`

## Estructura del proyecto

- `/db`: Contiene los esquemas SQL (`schema.sql`), las consultas (`queries.sql`) y el código Go autogenerado por `sqlc` (`/sqlc`).
- `/static`: Archivos CSS globales y específicos (`style.css`, `index.css`, `formularios.css`, `historial.css`).
- `/templates`: Vistas HTML dinámicas para el historial, registro de entrenamientos y carga de zapatillas.
- `main.go`: Archivo principal con la inicialización del servidor y los controladores de rutas (handlers).
- `Makefile`: Script de automatización para levantar contenedores, regenerar modelos y correr los tests del paquete `testing`.

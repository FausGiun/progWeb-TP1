# Runner Log - Tracker

Aplicación web para registrar entrenamientos de running y llevar el control del desgaste del equipo

## Cómo ejecutar

1. Asegurarse de tener Go instalado en el equipo.
2. Abrir una terminal en la carpeta raíz del proyecto.
3. Ordenar y descargar las dependencias del modulo ejecutando
   go mod tidy
4. Ejecutar el siguiente comando para levantar el servidor:
   go run main.go
5. Abrir un navegador web e ingresar a: http://localhost:8080

## Estructura del proyecto

- `/db`: Contiene los esquemas SQL (`schema.sql`), las consultas (`queries.sql`) y el código Go autogenerado por `sqlc` (`/sqlc`).
- `/static`: Archivos CSS globales y específicos (`style.css`, `index.css`, `formularios.css`, `historial.css`).
- `/templates`: Vistas HTML dinámicas para el historial, registro de entrenamientos y carga de zapatillas.
- `main.go`: Archivo principal con la inicialización del servidor y los controladores de rutas (handlers).

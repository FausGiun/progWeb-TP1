# Runner Log - Tracker

Aplicación web para registrar entrenamientos de running y llevar el control del desgaste del equipamiento.

## Prerrequisitos e Instalación

Para ejecutar y testear este proyecto, es necesario contar con:

- **Go** (instalado y configurado en el PATH).
- **Docker** y el plugin **Docker Compose** (para aislar la base de datos PostgreSQL).
  - _Cómo instalar en Linux (Ubuntu / Pop!\_OS / Debian):_
    Puede instalar el motor de Docker y el plugin de Compose directamente desde la terminal ejecutando:
    ```bash
    sudo apt update
    sudo apt install docker.io docker-compose-v2
    ```
  - _Nota sobre permisos:_ Para poder ejecutar los comandos de Docker o el `make test` sin tener que usar `sudo` cada vez, recuerde agregar su usuario al grupo de Docker ejecutando `sudo usermod -aG docker $USER` y luego reiniciar su sesión.
  - _Nota sobre Compose:_ Asegúrese de usar el comando `docker compose` (con espacio) nativo de la versión 2, en lugar del antiguo `docker-compose` (con guion).

## Documentación del Proyecto (Persistencia)

El proyecto utiliza una base de datos relacional **PostgreSQL** para garantizar la persistencia de los datos.

La arquitectura de datos se compone de dos tablas principales:

- **zapatillas:** Almacena el calzado registrado, llevando un acumulador del kilometraje total de cada par.
- **entrenamientos:** Registra cada sesión de running, vinculándose con la tabla de zapatillas mediante una relación de clave foránea (`zapatilla_id`).

Para la interacción entre la aplicación en Go y la base de datos se utiliza **sqlc**. Esta herramienta genera código Go fuertemente tipado a partir de consultas SQL puras (`queries.sql`) y el esquema de la base de datos (`schema.sql`), evitando el uso de ORMs pesados y garantizando consultas seguras.

## Testing automatizado

Para evaluar el proyecto y correr los tests automatizados (que incluyen la creación de un entorno limpio con Docker, la ejecución de pruebas unitarias sobre las rutas HTTP y operaciones CRUD de la base de datos), simplemente abra una terminal en la raíz del proyecto y ejecute:

```bash
make test
```

## Cómo ejecutar la aplicación localmente

Si desea levantar la aplicación de forma manual para probarla en el navegador:

1. Abrir una terminal en la carpeta raíz del proyecto.
2. Levantar el contenedor de la base de datos PostgreSQL:
   ```bash
   docker compose up -d
   ```
3. Ordenar y descargar las dependencias del módulo ejecutando:
   ```bash
   go mod tidy
   ```
4. Levantar el servidor ejecutando:
   ```bash
   go run main.go
   ```
5. Abrir un navegador web e ingresar a: `http://localhost:8080`
6. Para apagar la base de datos al terminar, ejecutar:
   ```bash
   docker compose down
   ```

## Estructura del proyecto

- `/db`: Contiene los esquemas SQL (`schema.sql`), las consultas (`queries.sql`) y el código Go autogenerado por `sqlc` (`/sqlc`).
- `/static`: Archivos CSS globales y específicos (`style.css`, `index.css`, `formularios.css`, `historial.css`).
- `/templates`: Vistas HTML dinámicas para el historial, registro de entrenamientos y carga de zapatillas.
- `main.go`: Archivo principal con la inicialización del servidor y los controladores de rutas (handlers).
- `docker-compose.yml`: Archivo de configuración para levantar la base de datos PostgreSQL.
- `Makefile`: Script de automatización para levantar contenedores, regenerar modelos y correr los tests del paquete `testing`.

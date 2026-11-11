# Running Tracker - Renderizado del Lado del Servidor (TP5)

Aplicación web desarrollada en **Go** para el registro y seguimiento de entrenamientos de running y control del kilometraje del calzado deportivo.

En esta etapa (TP5), se reemplaza la interfaz con JavaScript y la API JSON del TP4 por páginas renderizadas en el servidor mediante **templ**, manteniendo **PostgreSQL** y el acceso a datos con **sqlc**. Los formularios utilizan el patrón **Post/Redirect/Get (PRG)**: envían los datos mediante POST y, después de guardarlos, el servidor redirige al listado actualizado.

## Arquitectura del Proyecto

El proyecto mantiene la separación de responsabilidades en tres capas:

- **`views/` y controladores HTTP (Capa de Presentación):** Componentes templ para el layout, formularios y listados. `main.go` configura las rutas y el servidor; los archivos `handlers*.go` procesan los formularios y renderizan las páginas.
- **`logic/` (Capa de Lógica de Negocio / Dominio):** Validaciones de zapatillas y entrenamientos, cálculo de ritmo y calorías.
- **`db/` (Capa de Datos):** Esquema relacional (`db/schema/schema.sql`), consultas parametrizadas (`db/queries/queries.sql`) y código Go generado mediante sqlc.

## Requisitos Previos

- **Go** (1.26.6 o superior).
- **Docker** y **Docker Compose**.
- **sqlc**: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **templ**: `go install github.com/a-h/templ/cmd/templ@v0.3.1100`
- **make**, **Bash** y **curl**.

Los comandos se ejecutan desde la raíz del proyecto en Bash. En Windows se puede utilizar Git Bash con Docker Desktop iniciado. Las herramientas deben estar disponibles en el PATH.

## Configuración

Antes de la primera ejecución, si todavía no existe `.env`, copiar la configuración de ejemplo:

```bash
cp .env.example .env
```

El usuario, contraseña, base y puerto de PostgreSQL deben coincidir con los indicados en `DATABASE_URL`. La plantilla incluye valores para uso local. El archivo `.env` queda excluido de Git.

## Cómo Ejecutar el Proyecto

Con Docker iniciado, ejecutar:

```bash
make run
```

El comando genera el código de sqlc y templ, compila la aplicación, inicia PostgreSQL y ejecuta el servidor.

Abrir http://localhost:8080/. Primero se debe registrar un par desde la sección **Zapatillas** y luego se pueden cargar entrenamientos.

Al presionar **Ctrl+C**, se detienen el servidor y los contenedores, conservando los datos.

## Cómo Probar el Proyecto

Para generar el código y ejecutar las pruebas de lógica, base de datos, renderizado HTML y formularios:

```bash
make test
```

Las pruebas utilizan una base de datos separada de la aplicación. Al finalizar, se eliminan los contenedores y el volumen de pruebas, conservando los datos de desarrollo.
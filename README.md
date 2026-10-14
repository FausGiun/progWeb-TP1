# Running Tracker - API REST (TP3)

Aplicación backend desarrollada en **Go** para el registro y seguimiento de entrenamientos de running y control de desgaste de calzado deportivo.
En esta etapa (TP3), el sistema expone su lógica de negocio a través de una **API REST JSON**, conectando las reglas de dominio con una base de datos **PostgreSQL** mediante código tipado generado con **sqlc**.

## Arquitectura del Proyecto

El proyecto aplica separación de responsabilidades en tres capas:

- **`logic/` (Capa de Lógica de Negocio / Dominio):** Funciones puras en Go (`ValidateZapatilla`, `ValidateEntrenamiento`, `CalcularRitmo`, `CalcularCalorias`) sin acoplamiento a HTTP ni a la base de datos.
- **`db/` (Capa de Datos):** Esquema relacional (`db/schema/schema.sql`), consultas parametrizadas (`db/queries/queries.sql`) y código autogenerado en tiempo de compilación mediante `sqlc`.
- **`main.go` (Capa de Presentación / API REST):** Servidor HTTP con `loggingMiddleware` y handlers REST que procesan y retornan JSON utilizando los códigos de estado estándar (`200`, `201`, `204`, `400`, `404`, `405`).

## Requisitos Previos

- **Go** (1.22+)
- **Docker** y **Docker Compose**
- **sqlc** (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- **curl** y **make**

---

## Cómo Probar el Proyecto (Automatizado de punta a punta)

Para probar el proyecto recién clonado cumpliendo con todas las etapas (limpieza previa, generación con `sqlc`, compilación, levantamiento de base de datos, prueba de todos los endpoints REST y borrado posterior de contenedores y volúmenes), ejecute:

```bash
make test-api
```

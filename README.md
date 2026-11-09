# Running Tracker - La Capa de Presentación (TP4)

Aplicación web para el registro y seguimiento de entrenamientos de running y control del desgaste de calzado deportivo[cite: 251].

En esta etapa (**TP4**), el proyecto incorpora una **interfaz de usuario (frontend)** construida con HTML5, CSS3 (Flexbox y CSS Grid) y Vanilla JavaScript que consume de forma asíncrona (`fetch`) la **API REST JSON** desarrollada en Go[cite: 246, 247, 251].

## Arquitectura del Proyecto

El sistema respeta la separación en tres capas (Presentación, Lógica de Negocio y Datos)[cite: 12, 132]:

- **`static/` (Capa de Presentación - Cliente):**
  - `index.html`: Estructura principal con formularios de alta y contenedores dinámicos[cite: 246, 247, 251].
  - `styles.css`: Diseño responsivo estructurado mediante **CSS Grid** y **Flexbox**[cite: 246, 247, 251].
  - `app.js`: Cliente JavaScript que se comunica con el backend mediante `fetch` (`GET`, `POST`, `DELETE`), actualiza el DOM dinámicamente sin recargar la página e implementa delegación de eventos para la eliminación de registros[cite: 247, 251].
- **`main.go` y `handlers*.go` (API REST / Servidor HTTP):** Expone los endpoints JSON (`/zapatillas` y `/entrenamientos`), sirve los archivos estáticos del cliente en `/` e implementa el middleware de logging (`middleware.go`)[cite: 251].
- **`logic/` (Capa de Lógica de Negocio):** Funciones puras en Go (`ValidateZapatilla`, `ValidateEntrenamiento`, `CalcularRitmo`, `CalcularCalorias`) desacopladas de HTTP y de la base de datos[cite: 2, 251].
- **`db/` (Capa de Acceso a Datos):** Esquema relacional (`db/schema/schema.sql`), consultas parametrizadas (`db/queries/queries.sql`) y código autogenerado con `sqlc` para PostgreSQL[cite: 251, 255].

## Requisitos Previos

- **Go** (1.22+)
- **Docker** y **Docker Compose**
- **sqlc** (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)[cite: 30]
- **curl** y **make**

## Ejecución y Uso

### 1. Levantar la aplicación completa (Desarrollo)

El `Makefile` automatiza la generación de código con `sqlc`, la compilación del binario, el inicio del contenedor de PostgreSQL y la puesta en marcha del servidor[cite: 251, 254]:

```bash
make run
```

Una vez iniciado el servidor, abrir en el navegador:
👉 **`http://localhost:8080`**

*(Al presionar `Ctrl + C`, el servidor y los contenedores de Docker se detendrán automáticamente).*

### 2. Ejecutar Pruebas Automatizadas (Unitarias y de Integración)

Levanta una base de datos limpia en Docker, ejecuta todos los tests de Go (`logic`, `db/sqlc`, endpoints REST y servidor de archivos estáticos del TP4) y limpia el entorno al finalizar[cite: 251]:

```bash
make test
```

### 3. Probar la API REST de punta a punta con cURL

Ejecuta el flujo completo de ciclo de vida (preparación de infraestructura, arranque del servidor en segundo plano, peticiones CRUD mediante `requests.bash` y limpieza final)[cite: 4, 251]:

```bash
make test-api
```

### 4. Limpieza Manual del Entorno

Para detener contenedores, eliminar volúmenes de datos y borrar binarios compilados[cite: 251]:

```bash
make clean
```

## Endpoints de la API REST

| Método | Ruta | Descripción | Códigos de Estado |
| :--- | :--- | :--- | :--- |
| `GET` | `/zapatillas` | Lista todas las zapatillas activas | `200 OK` |
| `POST` | `/zapatillas` | Crea una nueva zapatilla | `201 Created`, `400 Bad Request` |
| `GET` | `/zapatillas/{id}` | Obtiene una zapatilla por su ID | `200 OK`, `404 Not Found` |
| `PUT` | `/zapatillas/{id}` | Actualiza los datos de una zapatilla | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `DELETE` | `/zapatillas/{id}` | Elimina lógicamente una zapatilla | `204 No Content`, `404 Not Found` |
| `GET` | `/entrenamientos` | Lista el historial de entrenamientos | `200 OK` |
| `POST` | `/entrenamientos` | Registra un entrenamiento y suma Kms al calzado | `201 Created`, `400 Bad Request` |
| `GET` | `/entrenamientos/{id}` | Obtiene un entrenamiento por su ID | `200 OK`, `404 Not Found` |
| `PUT` | `/entrenamientos/{id}` | Actualiza un entrenamiento existente | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `DELETE` | `/entrenamientos/{id}` | Elimina un entrenamiento | `204 No Content`, `404 Not Found` |

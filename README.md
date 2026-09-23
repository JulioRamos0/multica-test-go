# multica-test-go

Este es un proyecto backend en Go que sirve como gestor de tareas (To-Do) implementando un API REST protegida.

## Tabla de Contenido

- [1. Estructura y Metodología](#1-estructura-y-metodología)
- [2. Requisitos](#2-requisitos)
- [3. Cómo ejecutarlo](#3-cómo-ejecutarlo)
- [4. Cómo probarlo](#4-cómo-probarlo)
- [5. Documentación de los Endpoints](#5-documentación-de-los-endpoints)

---

## 1. Estructura y Metodología

El desarrollo de este proyecto se guía por **SDD (Specification Driven Development)** y **TDD (Test Driven Development)**, manteniendo siempre un historial de Git limpio y organizado.

El proyecto está estructurado de la siguiente manera:
- `auth/`: Contiene la lógica de autenticación (Middleware por API Key) y sus pruebas.
- `docs/`: Documentación general y especificaciones funcionales.
- `health/`: Endpoints e información sobre el estado de salud y versión de la aplicación (`/health`, `/health/version`).
- `specs/`: Especificaciones detalladas de las funcionalidades.
- `todo/`: Lógica principal del gestor de tareas y sus manejadores HTTP.

## 2. Requisitos

- [Go](https://golang.org/dl/) 1.20 o superior.

## 3. Cómo ejecutarlo

Actualmente, el proyecto se provee como un conjunto de paquetes modulares. Para integrarlo en un servidor principal y ejecutarlo, debes proveer la clave de API requerida por las rutas protegidas a través de variables de entorno:

```bash
export APP_X_API_KEY="tu-clave-secreta"
go run main.go # (Una vez configurado el punto de entrada principal)
```

## 4. Cómo probarlo

Para ejecutar la suite completa de pruebas unitarias y de integración del proyecto (TDD), utiliza el siguiente comando en la raíz del proyecto:

```bash
go test ./... -v
```

## 5. Documentación de los Endpoints

La aplicación expone dos grupos principales de endpoints: los de estado (públicos) y los de la API (protegidos).

### Endpoints Públicos (Health)

- **Liveness**
  - **URL:** `/health`
  - **Método:** `GET`
  - **Descripción:** Verifica que el servidor está levantado y aceptando conexiones.
  - **Respuesta (200 OK):** `{"status": "UP", "timestamp": "2023-10-01T12:00:00Z"}`

- **Versión**
  - **URL:** `/health/version`
  - **Método:** `GET`
  - **Descripción:** Devuelve la versión actual del API.
  - **Respuesta (200 OK):** `{"version": "1.0.0"}`

### Endpoints Protegidos (Gestor de Tareas)

*Nota: Todos los endpoints bajo `/api/todo` requieren el encabezado HTTP `x-api-key` con un valor válido.*

- **Obtener todas las tareas**
  - **URL:** `/api/todo`
  - **Método:** `GET`
  - **Descripción:** Retorna un arreglo con todas las tareas existentes.

- **Obtener una tarea por ID**
  - **URL:** `/api/todo?id={id}`
  - **Método:** `GET`
  - **Descripción:** Retorna los detalles de la tarea especificada.

- **Crear una nueva tarea**
  - **URL:** `/api/todo`
  - **Método:** `POST`
  - **Cuerpo (JSON):**
    ```json
    {
      "id": "opcional",
      "title": "Título de la tarea",
      "description": "Descripción detallada",
      "date": "2023-10-01T12:00:00Z"
    }
    ```
  - **Respuestas:** `201 Created` en caso de éxito.

- **Actualizar una tarea**
  - **URL:** `/api/todo?id={id}` (El ID también puede ir en el cuerpo)
  - **Método:** `PUT`
  - **Cuerpo (JSON):**
    ```json
    {
      "title": "Nuevo título",
      "description": "Nueva descripción",
      "date": "2023-10-01T12:00:00Z"
    }
    ```
  - **Respuestas:** `200 OK` en caso de éxito.

- **Eliminar una tarea**
  - **URL:** `/api/todo?id={id}`
  - **Método:** `DELETE`
  - **Respuestas:** `204 No Content` en caso de éxito.

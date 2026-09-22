# Especificación Técnica: API REST de Todo (/api/todo)

## 1. Resumen del Objetivo
El objetivo de este documento es definir la arquitectura técnica, los contratos de comunicación, las estructuras de datos en Go y los criterios de aceptación para el endpoint `/api/todo`. Este servicio permite la gestión integral (creación, lectura, actualización y eliminación - CRUD) de tareas pendientes ("Todos"), garantizando consistencia en el manejo de errores y compatibilidad con el ecosistema en Go.

---

## 2. Endpoints

El servicio expone la ruta base `/api/todo` soportando los verbos HTTP descriptivos estándar:

### 2.1 Obtener Todos / Detalle: `GET /api/todo`
Permite listar todas las tareas registradas o consultar una tarea específica mediante el parámetro de consulta `id`.

- **Método HTTP**: `GET`
- **Ruta**: `/api/todo`
- **Query Parameters**:
  - `id` (string, opcional): Identificador único del todo. Si se omite, retorna la lista completa.
- **Headers de Solicitud**:
  - `Accept`: `application/json` (recomendado)
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Códigos de Estado HTTP**:
  - `200 OK`: Lista de tareas retornada o tarea específica encontrada exitosamente.
  - `404 Not Found`: Cuando se especifica un `id` inexistente.

### 2.2 Crear Todo: `POST /api/todo`
Permite crear un nuevo elemento Todo en el sistema.

- **Método HTTP**: `POST`
- **Ruta**: `/api/todo`
- **Headers de Solicitud**:
  - `Content-Type`: `application/json`
  - `Accept`: `application/json`
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Códigos de Estado HTTP**:
  - `201 Created`: Tarea creada exitosamente.
  - `400 Bad Request`: Payload JSON inválido o campos requeridos faltantes (`title`).
  - `409 Conflict`: Si se provee un `id` que ya existe en el sistema.

### 2.3 Actualizar Todo: `PUT /api/todo`
Permite actualizar los datos de una tarea existente. El identificador puede ser provisto vía query param `?id=<id>` o dentro del cuerpo JSON.

- **Método HTTP**: `PUT`
- **Ruta**: `/api/todo`
- **Query Parameters**:
  - `id` (string, opcional si viene en el body): Identificador único del todo a modificar.
- **Headers de Solicitud**:
  - `Content-Type`: `application/json`
  - `Accept`: `application/json`
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Códigos de Estado HTTP**:
  - `200 OK`: Tarea actualizada exitosamente.
  - `400 Bad Request`: Payload inválido o identificador no provisto.
  - `404 Not Found`: Tarea no encontrada para el identificador dado.

### 2.4 Eliminar Todo: `DELETE /api/todo`
Elimina una tarea existente del sistema a partir de su identificador.

- **Método HTTP**: `DELETE`
- **Ruta**: `/api/todo`
- **Query Parameters**:
  - `id` (string, requerido): Identificador único del todo a eliminar.
- **Headers de Solicitud**:
  - Ninguno obligatorio.
- **Headers de Respuesta**:
  - Ninguno obligatorio (cuerpo vacío).
- **Códigos de Estado HTTP**:
  - `204 No Content`: Tarea eliminada exitosamente.
  - `400 Bad Request`: Parámetro `id` no provisto.
  - `404 Not Found`: Tarea no encontrada para el identificador indicado.

---

## 3. Contratos JSON (Request y Response Schemas)

### 3.1 `GET /api/todo` (Listado general)

#### Request Body
- Ninguno (vacío).

#### Response Body (200 OK)
```json
[
  {
    "id": "1",
    "title": "Configurar pipeline CI/CD",
    "description": "Definir GitHub Actions para validación de pruebas unitarias",
    "date": "2026-09-22T10:00:00Z"
  },
  {
    "id": "2",
    "title": "Escribir documentación",
    "description": "Especificar endpoints de la API",
    "date": "2026-09-22T12:00:00Z"
  }
]
```

### 3.2 `GET /api/todo?id=1` (Consulta individual)

#### Request Body
- Ninguno (vacío).

#### Response Body (200 OK)
```json
{
  "id": "1",
  "title": "Configurar pipeline CI/CD",
  "description": "Definir GitHub Actions para validación de pruebas unitarias",
  "date": "2026-09-22T10:00:00Z"
}
```

#### Response Body (404 Not Found)
```json
{
  "error": "todo not found"
}
```

### 3.3 `POST /api/todo` (Creación)

#### Request Body
```json
{
  "title": "Comprar insumos",
  "description": "Comprar café y fruta para la oficina",
  "date": "2026-09-22T15:30:00Z"
}
```
*Nota: El campo `id` es opcional; si no se envía, el servidor generará un identificador único.*

#### Response Body (201 Created)
```json
{
  "id": "20260922153000-abc123",
  "title": "Comprar insumos",
  "description": "Comprar café y fruta para la oficina",
  "date": "2026-09-22T15:30:00Z"
}
```

#### Response Body (400 Bad Request)
```json
{
  "error": "invalid request body: title is required"
}
```

### 3.4 `PUT /api/todo?id=1` (Actualización)

#### Request Body
```json
{
  "title": "Comprar insumos de oficina",
  "description": "Comprar café, té y fruta fresca",
  "date": "2026-09-22T16:00:00Z"
}
```

#### Response Body (200 OK)
```json
{
  "id": "1",
  "title": "Comprar insumos de oficina",
  "description": "Comprar café, té y fruta fresca",
  "date": "2026-09-22T16:00:00Z"
}
```

### 3.5 `DELETE /api/todo?id=1` (Eliminación)

#### Request Body
- Ninguno (vacío).

#### Response Body (204 No Content)
- Sin cuerpo de respuesta.

#### Response Body (404 Not Found)
```json
{
  "error": "todo not found"
}
```

---

## 4. Tipos y Structs en Go

A continuación se definen los tipos e interfaces requeridos para la implementación:

```go
package todo

import (
	"context"
	"net/http"
	"time"
)

// Todo representa la entidad principal de una tarea pendiente.
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// CreateTodoRequest define el payload para la creación de un Todo.
type CreateTodoRequest struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// UpdateTodoRequest define el payload para la actualización de un Todo.
type UpdateTodoRequest struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// ErrorResponse estructura estándar para respuestas de error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Repository define el contrato de almacenamiento para las tareas Todo.
type Repository interface {
	GetAll(ctx context.Context) ([]Todo, error)
	GetByID(ctx context.Context, id string) (Todo, bool, error)
	Create(ctx context.Context, todo Todo) (Todo, error)
	Update(ctx context.Context, todo Todo) (Todo, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// Handler expone las operaciones HTTP para el endpoint /api/todo.
type Handler struct {
	repo Repository
}

// NewHandler inicializa un nuevo Handler con el repositorio especificado.
func NewHandler(repo Repository) *Handler

// ServeHTTP implementa http.Handler enrutando según el método HTTP (GET, POST, PUT, DELETE).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

---

## 5. Criterios de Aceptación (Acceptance Criteria)

### CA-01: Listado de Todos (GET)
- **Dado** que existen elementos Todo en el sistema.
- **Cuando** se realiza un `GET /api/todo` sin parámetros de consulta.
- **Entonces** el servidor retorna código `200 OK`, `Content-Type: application/json`, y un arreglo con todos los elementos almacenados.
- **Y** si no hay elementos almacenados, retorna código `200 OK` con un arreglo vacío `[]`.

### CA-02: Consulta de Todo por ID (GET)
- **Dado** un Todo existente con ID `"todo-123"`.
- **Cuando** se realiza un `GET /api/todo?id=todo-123`.
- **Entonces** el servidor retorna código `200 OK` y el objeto JSON correspondiente.
- **Dado** que el ID no existe en el sistema.
- **Cuando** se realiza un `GET /api/todo?id=inexistente`.
- **Entonces** el servidor retorna código `404 Not Found`.

### CA-03: Creación de Todo (POST)
- **Dado** un cuerpo JSON válido con al menos `title` y opcionalmente `description` y `date`.
- **Cuando** se envía un `POST /api/todo`.
- **Entonces** el servidor retorna código `201 Created` con el objeto creado conteniendo un `id` asignado.
- **Dado** un JSON malformado o sin el campo obligatorio `title`.
- **Cuando** se envía un `POST /api/todo`.
- **Entonces** el servidor retorna código `400 Bad Request`.

### CA-04: Actualización de Todo (PUT)
- **Dado** un Todo existente con ID `"todo-123"`.
- **Cuando** se envía un `PUT /api/todo?id=todo-123` (o con `id` en el body) y los datos a actualizar.
- **Entonces** el servidor actualiza la información y retorna código `200 OK` con el Todo actualizado.
- **Dado** un intento de actualización de un ID no existente.
- **Cuando** se envía un `PUT /api/todo?id=no-existe`.
- **Entonces** el servidor retorna código `404 Not Found`.

### CA-05: Eliminación de Todo (DELETE)
- **Dado** un Todo existente con ID `"todo-123"`.
- **Cuando** se envía un `DELETE /api/todo?id=todo-123`.
- **Entonces** el servidor remueve el elemento y retorna código `204 No Content`.
- **Dado** un intento de eliminación sin proveer el parámetro `id`.
- **Cuando** se envía un `DELETE /api/todo`.
- **Entonces** el servidor retorna código `400 Bad Request`.
- **Dado** un intento de eliminación de un ID inexistente.
- **Cuando** se envía un `DELETE /api/todo?id=inexistente`.
- **Entonces** el servidor retorna código `404 Not Found`.

### CA-06: Manejo de Métodos No Soportados
- **Cuando** se envía una petición con un método HTTP no contemplado (por ejemplo, `PATCH` o `HEAD`).
- **Entonces** el servidor retorna código `405 Method Not Allowed`.

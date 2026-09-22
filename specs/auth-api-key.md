# Especificación Técnica: Autenticación por X-API-Key en `/api/todo`

## 1. Resumen del Objetivo
El objetivo de este documento es definir la especificación técnica, contratos de comunicación, estructuras de datos en Go y criterios de aceptación para incorporar autenticación basada en API Key en los endpoints bajo la ruta `/api/todo`.

La clave secreta esperada se configura mediante la variable de entorno `APP_X_API_KEY`. En cada solicitud hacia `/api/todo` (o cualquier método HTTP asociado a dicho recurso), el cliente debe proporcionar dicha clave en el encabezado HTTP `x-api-key`. Si el encabezado está ausente, vacío o no coincide exactamente con el valor configurado, el servicio deniega el acceso retornando un código de estado `401 Unauthorized`.

---

## 2. Endpoints y Autenticación

La autenticación por API Key aplica a todos los métodos HTTP expuestos por el recurso `/api/todo` (`GET`, `POST`, `PUT`, `DELETE`, etc.), actuando como capa de protección (middleware o interceptor).

### 2.1 Encabezados de Autenticación
- **Header esperado**: `x-api-key` (o normalizado insensible a mayúsculas/minúsculas según el estándar HTTP `X-Api-Key` / `X-API-Key`).
- **Variable de entorno**: `APP_X_API_KEY` define la clave requerida por la aplicación.
  - Si la variable de entorno `APP_X_API_KEY` está configurada, cada petición a `/api/todo` debe validar que el valor recibido en el header coincida con dicho secreto.
  - Si el header no se envía o el valor provisto no es idéntico al configurado en `APP_X_API_KEY`, se rechaza inmediatamente la petición antes de procesar el handler de Todo.

### 2.2 Respuestas de Autenticación

#### Caso No Autorizado (401 Unauthorized)
- **Código de Estado HTTP**: `401 Unauthorized`
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Cuerpo de Respuesta**:
  ```json
  {
    "error": "unauthorized"
  }
  ```

#### Caso Autorizado
- La petición continúa hacia el handler correspondiente de `/api/todo` (`GET`, `POST`, `PUT`, `DELETE`), retornando los códigos habituales (`200 OK`, `201 Created`, `204 No Content`, `400 Bad Request`, `404 Not Found`, etc.) definidos en la especificación base de Todo.

---

## 3. Contratos JSON (Request y Response Schemas)

### 3.1 Respuesta de Error de Autenticación (401 Unauthorized)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "UnauthorizedErrorResponse",
  "type": "object",
  "properties": {
    "error": {
      "type": "string",
      "example": "unauthorized"
    }
  },
  "required": ["error"],
  "additionalProperties": false
}
```

---

## 4. Tipos y Structs en Go

A continuación se detallan las interfaces, structs y funciones recomendadas para implementar el middleware/autenticador en Go:

```go
package auth

import (
	"net/http"
)

// Constantes para autenticación por API Key.
const (
	// HeaderXAPIKey nombre canónico del encabezado HTTP para la API key.
	HeaderXAPIKey = "x-api-key"
	// EnvAppXAPIKey nombre de la variable de entorno que define la clave autorizada.
	EnvAppXAPIKey = "APP_X_API_KEY"
	// ErrMsgUnauthorized mensaje devuelto en el cuerpo JSON en caso de credenciales inválidas.
	ErrMsgUnauthorized = "unauthorized"
)

// AuthConfig contiene la configuración de autenticación.
type AuthConfig struct {
	APIKey string
}

// LoadConfigFromEnv carga la clave API desde la variable de entorno APP_X_API_KEY.
func LoadConfigFromEnv() AuthConfig

// Middleware provee un middleware HTTP para proteger endpoints mediante x-api-key.
type Middleware struct {
	expectedKey string
}

// NewMiddleware construye una nueva instancia de Middleware con la clave esperada.
func NewMiddleware(apiKey string) *Middleware

// Wrap envuelve un http.Handler existente exigiendo autenticación válida.
func (m *Middleware) Wrap(next http.Handler) http.Handler

// HandlerFunc envuelve un http.HandlerFunc exigiendo autenticación válida.
func (m *Middleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc
```

---

## 5. Criterios de Aceptación (Acceptance Criteria)

### CA-AUTH-01: Petición sin encabezado `x-api-key`
- **Dado** que el servicio está configurado con la variable `APP_X_API_KEY="secret-token-123"`.
- **Cuando** se realiza cualquier solicitud HTTP (`GET`, `POST`, `PUT`, `DELETE`) a `/api/todo` sin incluir el encabezado `x-api-key`.
- **Entonces** el servidor responde inmediatamente con código de estado `401 Unauthorized`.
- **Y** el encabezado `Content-Type` es `application/json; charset=utf-8`.
- **Y** el cuerpo de la respuesta contiene `{"error": "unauthorized"}`.

### CA-AUTH-02: Petición con valor incorrecto o vacío en `x-api-key`
- **Dado** que el servicio está configurado con `APP_X_API_KEY="secret-token-123"`.
- **Cuando** se realiza una solicitud HTTP a `/api/todo` con encabezado `x-api-key: wrong-token` o `x-api-key: ""`.
- **Entonces** el servidor responde con código de estado `401 Unauthorized` y payload `{"error": "unauthorized"}`.
- **Y** ninguna operación sobre los Todos (lectura, inserción o modificación) se ejecuta.

### CA-AUTH-03: Petición con valor válido en `x-api-key`
- **Dado** que el servicio está configurado con `APP_X_API_KEY="secret-token-123"`.
- **Cuando** se realiza una solicitud HTTP a `/api/todo` con encabezado `x-api-key: secret-token-123`.
- **Entonces** la autenticación es exitosa y la solicitud se delega al handler de `/api/todo`.
- **Y** la respuesta es la correspondiente a la operación solicitada (por ejemplo `200 OK`, `201 Created`, etc.).

### CA-AUTH-04: Insensibilidad a mayúsculas/minúsculas en el header
- **Dado** que el cliente envía `X-API-KEY: secret-token-123` o `X-Api-Key: secret-token-123`.
- **Cuando** la clave provista coincide exactamente con `APP_X_API_KEY`.
- **Entonces** la solicitud se autentica satisfactoriamente respetando el estándar HTTP de headers.

### CA-AUTH-05: Exclusión de otros endpoints (ej. `/health`)
- **Dado** que el endpoint `/health` o similares no están bajo `/api/todo`.
- **Cuando** se realiza una petición a `/health` sin header `x-api-key`.
- **Entonces** la solicitud responde `200 OK` (la autenticación `x-api-key` solo aplica al ámbito de `/api/todo`).

# Especificación Técnica: Microservicio HTTP de Health Check

## 1. Resumen del Objetivo
Definir la arquitectura técnica, endpoints, contratos de datos y criterios de aceptación para un microservicio HTTP ligero en Go cuyo propósito es exponer el estado de salud operativa de la aplicación (liveness y readiness checks), adecuado para orquestadores (Kubernetes, Docker Swarm) y sistemas de monitoreo/observabilidad.

---

## 2. Definición de Endpoints

### 2.1 Endpoint Liveness: `GET /healthz`
Verifica si el proceso del servicio está vivo y respondiendo peticiones HTTP básicas.

- **Método HTTP**: `GET`
- **Ruta**: `/healthz`
- **Headers de Solicitud**:
  - `Accept`: `application/json` (opcional)
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Códigos de Estado HTTP**:
  - `200 OK`: El servicio está activo y saludable.

### 2.2 Endpoint Readiness: `GET /ready`
Verifica si el servicio está listo para recibir tráfico externo (incluyendo comprobación de dependencias o subsistemas internos si aplican).

- **Método HTTP**: `GET`
- **Ruta**: `/ready`
- **Headers de Solicitud**:
  - `Accept`: `application/json` (opcional)
- **Headers de Respuesta**:
  - `Content-Type`: `application/json; charset=utf-8`
- **Códigos de Estado HTTP**:
  - `200 OK`: Todas las comprobaciones requeridas están operativas (`UP`).
  - `503 Service Unavailable`: Al menos una dependencia crítica no está disponible (`DOWN`).

---

## 3. Esquemas JSON de Request y Response

### 3.1 `GET /healthz`

#### Request Body
- Vacío (sin cuerpo de solicitud).

#### Response Body (200 OK)
```json
{
  "status": "UP",
  "timestamp": "2026-09-22T01:42:55Z"
}
```

### 3.2 `GET /ready`

#### Request Body
- Vacío (sin cuerpo de solicitud).

#### Response Body (200 OK - Servicio Listo)
```json
{
  "status": "UP",
  "timestamp": "2026-09-22T01:42:55Z",
  "checks": [
    {
      "name": "system",
      "status": "UP"
    }
  ]
}
```

#### Response Body (503 Service Unavailable - Falla de Subsistema)
```json
{
  "status": "DOWN",
  "timestamp": "2026-09-22T01:42:55Z",
  "checks": [
    {
      "name": "system",
      "status": "DOWN",
      "error": "system degraded or dependency unreachable"
    }
  ]
}
```

---

## 4. Definición de Tipos y Structs en Go

A continuación se detallan las estructuras Go requeridas para el modelo de datos y serialización JSON:

```go
package health

import "time"

// Status representa el estado de salud ("UP", "DOWN").
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// CheckResult contiene el resultado de una comprobación individual de preparación o dependencia.
type CheckResult struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthResponse representa la respuesta del endpoint liveness (/healthz).
type HealthResponse struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ReadinessResponse representa la respuesta del endpoint readiness (/ready).
type ReadinessResponse struct {
	Status    Status        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Checks    []CheckResult `json:"checks"`
}

// Checker define la interfaz para realizar verificaciones de estado en el endpoint readiness.
type Checker interface {
	Check() CheckResult
}
```

---

## 5. Casos de Prueba Mínimos y Criterios de Aceptación (Acceptance Criteria)

### 5.1 Criterios de Aceptación (AC)

1. **AC-01 (Liveness OK)**:
   - Toda petición `GET` a `/healthz` debe retornar código `200 OK`.
   - El header `Content-Type` debe contener `application/json`.
   - El cuerpo JSON debe contener `status: "UP"` y un campo `timestamp` en formato RFC3339 válido.

2. **AC-02 (Readiness OK)**:
   - Toda petición `GET` a `/ready` cuando todos los checkers reportan `UP` debe retornar código `200 OK`.
   - El campo `status` general debe ser `"UP"`.
   - La lista `checks` debe contener el desglose de cada comprobación efectuada.

3. **AC-03 (Readiness Falla/Degradado)**:
   - Si al menos una verificación reporta `DOWN`, el endpoint `GET /ready` debe retornar código `503 Service Unavailable`.
   - El campo `status` general debe ser `"DOWN"`.
   - El elemento en `checks` con fallo debe incluir el mensaje descriptivo en `error`.

4. **AC-04 (Métodos HTTP no permitidos)**:
   - Peticiones con métodos distintos a `GET` (e.g., `POST`, `PUT`, `DELETE`) en `/healthz` o `/ready` deben responder con código HTTP `405 Method Not Allowed`.

### 5.2 Casos de Prueba Mínimos Requeridos (Table-Driven Tests)

| ID | Endpoint | Método | Estado Dependencias | Código HTTP Esperado | `status` Esperado | Validaciones Adicionales |
|---|---|---|---|---|---|---|
| TC-01 | `/healthz` | `GET` | N/A | `200 OK` | `"UP"` | Validar parsing RFC3339 de `timestamp` |
| TC-02 | `/healthz` | `POST` | N/A | `405 Method Not Allowed` | N/A | Endpoint no acepta mutaciones |
| TC-03 | `/ready` | `GET` | Chequeos saludables | `200 OK` | `"UP"` | Longitud de `checks` >= 1 |
| TC-04 | `/ready` | `GET` | Chequeo con fallo | `503 Service Unavailable` | `"DOWN"` | Objeto en `checks` contiene string `error` |
| TC-05 | `/ready` | `DELETE` | N/A | `405 Method Not Allowed` | N/A | Endpoint solo acepta lecturas |

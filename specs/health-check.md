# Health Check Version Endpoint

## Resumen del objetivo
Proporcionar un endpoint de diagnóstico (health check) que exponga la versión actual del servicio desplegado. Esto permite a los monitores, orquestadores de infraestructura y clientes verificar rápidamente qué versión del código se está ejecutando.

## Endpoints
- **Método HTTP:** `GET`
- **Ruta:** `/health/version`
- **Headers Requeridos:** Ninguno.
- **Códigos de Respuesta:**
  - `200 OK`: La solicitud fue exitosa y se retorna la versión.

## Contratos JSON

### Request
El endpoint no recibe un cuerpo (body).

### Response (`200 OK`)
```json
{
  "version": "1.0.0"
}
```

## Tipos y Structs en Go

```go
package health

// VersionResponse define el contrato de la respuesta para el endpoint de versión.
type VersionResponse struct {
	Version string `json:"version"`
}

// Handler define la interfaz para los controladores de health check.
type Handler interface {
	GetVersion() VersionResponse
}
```

## Criterios de Aceptación (Acceptance Criteria)
1. **Petición Exitosa (Happy Path):**
   - Dado que el servicio está en ejecución,
   - Cuando se realiza una petición `GET` a `/health/version`,
   - Entonces el servicio debe responder con un código HTTP `200 OK`.
   - Y el cuerpo de la respuesta debe ser un JSON válido que contenga la propiedad `"version"` con la versión actual (ejemplo: `{"version": "1.0.0"}`).

2. **Método no permitido:**
   - Dado que el servicio está en ejecución,
   - Cuando se realiza una petición con un método distinto a `GET` (por ejemplo, `POST` o `PUT`) a `/health/version`,
   - Entonces el servicio debe responder con un código HTTP `405 Method Not Allowed`.

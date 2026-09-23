# Global Rate Limiting Middleware

## Resumen del objetivo
Implementar un Middleware de Rate Limiting global en Go para restringir la cantidad de peticiones que un usuario puede realizar. El requerimiento especifica que un usuario, identificado mediante el token en el Auth Header, no podrá hacer más de 60 peticiones por minuto. Si el límite es excedido, la API debe rechazar la petición devolviendo un código de estado HTTP 429 (Too Many Requests).

## Endpoints
Este middleware es global y aplica a todos los endpoints de la API (ej. `/*`).

- **Método HTTP:** Cualquier método (GET, POST, PUT, DELETE, PATCH, etc.).
- **Ruta:** Todas las rutas expuestas o bajo protección.
- **Headers Relevantes:**
  - `Authorization`: `Bearer <token>` u otra convención usada por la API, utilizado para extraer el identificador único del usuario.
- **Códigos de Respuesta Intervenidos:**
  - `429 Too Many Requests`: Devuelto cuando el usuario excedió el límite de 60 peticiones en 1 minuto.

## Contratos JSON
Para la respuesta cuando se excede el límite (429), el esquema debe ser un JSON claro.

**Response schema (429 Too Many Requests):**
```json
{
  "error": "too_many_requests",
  "message": "Rate limit exceeded. Try again later."
}
```

## Tipos y Structs en Go

```go
package middleware

import "net/http"

// RateLimiter define el contrato para el almacenamiento y validación de límites.
type RateLimiter interface {
	// Allow evalúa si una nueva petición para un identificador dado está permitida.
	// Devuelve true si el límite no ha sido excedido; de lo contrario, devuelve false.
	Allow(identifier string) bool
}

// GlobalRateLimiter devuelve el middleware que aplica las restricciones
// usando una implementación de RateLimiter.
func GlobalRateLimiter(limiter RateLimiter) func(http.Handler) http.Handler
```

## Criterios de Aceptación (Acceptance Criteria)

1. **Permitir peticiones dentro del límite:**
   - Un usuario realiza de 1 a 60 peticiones en el lapso de un minuto.
   - **Resultado:** Todas las peticiones proceden hacia los handlers reales y responden normalmente.

2. **Rechazar la petición 61+:**
   - Un usuario realiza su petición número 61 dentro de la misma ventana de 1 minuto.
   - **Resultado:** El middleware intercepta la petición, no llama al handler subyacente y responde con HTTP 429 (Too Many Requests) junto al payload JSON de error.

3. **Independencia por usuario:**
   - El Usuario A realiza 60 peticiones.
   - El Usuario B realiza 1 petición.
   - **Resultado:** El Usuario A comienza a recibir errores 429, mientras que el Usuario B puede seguir operando con normalidad (aislamiento del rate limit por token).

4. **Reinicio del límite (Window Reset):**
   - El Usuario A agota su cuota de 60 peticiones en el minuto 1 y es bloqueado.
   - Transcurrido el minuto, el Usuario A vuelve a enviar peticiones.
   - **Resultado:** Las nuevas peticiones en la siguiente ventana de tiempo son aceptadas (HTTP 200).

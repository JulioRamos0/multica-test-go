# agents.md

## Reglas de Git

- Escribe mensajes de commit limpios y concisos (ej. Conventional Commits).
- NUNCA agregues la firma "Co-authored-by:" ni menciones a la IA en el mensaje del commit. Todo el código debe parecer escrito por un humano.

## Estructura del Proyecto

El proyecto está organizado de la siguiente manera:
- `auth/`: Contiene la lógica de autenticación y sus pruebas.
- `docs/`: Documentación general (ej. SPEC.md).
- `health/`: Endpoints e información sobre el estado de salud y versión de la aplicación.
- `specs/`: Especificaciones detalladas de las funcionalidades.
- `todo/`: Lógica principal del gestor de tareas y su API.

## Metodología

Este proyecto está enfocado en utilizar **SDD (Specification Driven Development)** y **TDD (Test Driven Development)**.

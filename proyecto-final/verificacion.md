# Fase 4 — Verificación del proyecto final: préstamo de laptops

**Curso:** Arquitectura de Software  
**Integrantes:** Christian Javier Benitez Castro — T00082581; Sean Paul Martinez Roa — T00083011  
**Fecha de las pruebas:** 25 de noviembre de 2026  
**Entorno:** Windows, Docker Desktop, API Go + Gin, MySQL 8 y Adminer  
**Estado:** verificación funcional realizada en la máquina de desarrollo; quedan casos complementarios sin ejecutar.

## Propósito y alcance

Comprobar que el módulo nuevo de laptops funciona con autenticación JWT y roles, impide préstamos simultáneos, registra devoluciones y conserva el funcionamiento del módulo de libros. Las afirmaciones siguientes se sustentan en **respuestas reales** mostradas en las capturas adjuntas. No constituyen pruebas automáticas de concurrencia ni sustituyen una auditoría de seguridad.

## Resultados generales

- `docker compose up --build -d` construyó la API y arrancó MySQL, Adminer y la API. `docker compose ps` confirmó MySQL como **healthy**.
- `GET /health` y `GET /ping` respondieron 200; la API de laptops sin JWT respondió 401.
- Se registraron dos usuarios de prueba. El usuario administrador fue habilitado mediante actualización de rol en MySQL y volvió a iniciar sesión para obtener un nuevo JWT.
- Se creó la laptop `UTB-LAP-001` (Lenovo ThinkPad E14); su estado cambió de `available` a `borrowed` al prestar y regresó a `available` al devolver.
- Un segundo préstamo sobre el mismo equipo fue rechazado con **409 Conflict**, y una segunda devolución también con **409 Conflict**.
- El estudiante no pudo crear una laptop: respuesta **403 Forbidden**.
- Se realizó un préstamo y devolución de *Cien años de soledad*; las copias disponibles pasaron de **3 a 2** y luego volvieron a **3**.
- `go build ./...` y `go vet ./...` terminaron sin errores visibles. Se creó el commit `788c9d6` para el módulo de laptops.

## Matriz de verificación — resultados observados

| # | Caso | Solicitud o comando | Esperado | Resultado real | Evidencia |
|---|---|---|---|---|---|
| 01 | Docker y contenedores | `docker compose up --build -d` / `docker compose ps` | 3 servicios en marcha, MySQL `healthy` | **Aprobado** — API, MySQL y Adminer activos | [`01_Docker_Contenedores.png`](evidencias/01_Docker_Contenedores.png) |
| 02 | Health | `GET /health` | 200 + `{"status":"ok"}` | **Aprobado** — HTTP 200 | [`02_Verificacion_Endpoints.png`](evidencias/02_Verificacion_Endpoints.png) |
| 03 | Ping | `GET /ping` | 200 + `{"message":"pong"}` | **Aprobado** — HTTP 200 | [`02_Verificacion_Endpoints.png`](evidencias/02_Verificacion_Endpoints.png) |
| 04 | Acceso anónimo a inventario | `GET /api/laptops` sin token | 401 | **Aprobado** — HTTP 401 | [`02_Verificacion_Endpoints.png`](evidencias/02_Verificacion_Endpoints.png) |
| 05 | Registro estudiante | `POST /api/auth/register` | 201, rol `student` | **Aprobado** — HTTP 201, id 1 | [`03_Registro_Usuario.png`](evidencias/03_Registro_Usuario.png) |
| 06 | Login estudiante | `POST /api/auth/login` | 200, token JWT | **Aprobado** — HTTP 200 | [`04_Login_Usuario.png`](evidencias/04_Login_Usuario.png) |
| 07 | Inventario inicial laptops | `GET /api/laptops` con token | 200, lista vacía | **Aprobado** — HTTP 200, `[]` | [`05_Consulta_Laptops.png`](evidencias/05_Consulta_Laptops.png) |
| 08 | Alta y promoción de administrador | `POST /api/auth/register` + SQL `UPDATE users` | Usuario con rol admin | **Aprobado** — HTTP 201 y rol `admin` en MySQL | [`06_Creacion_Administrador.png`](evidencias/06_Creacion_Administrador.png) |
| 09 | Login administrador | `POST /api/auth/login` | 200, JWT rol admin | **Aprobado** — HTTP 200, rol `admin` | [`07_Login_Administrador.png`](evidencias/07_Login_Administrador.png) |
| 10 | Crear laptop (admin) | `POST /api/laptops` | 201, equipo disponible | **Aprobado** — HTTP 201, serial `UTB-LAP-001`, id 1 | [`08_Registro_Laptop.png`](evidencias/08_Registro_Laptop.png) |
| 11 | Prestar laptop | `POST /api/laptop-loans` | 201, préstamo activo | **Aprobado** — HTTP 201, préstamo 1, 7 días | [`09_Prestamo_Laptop.png`](evidencias/09_Prestamo_Laptop.png) |
| 12 | Estado de equipo prestado | `GET /api/laptops` | 200, estado `borrowed` | **Aprobado** — HTTP 200, estado `borrowed` | [`10_Laptop_Estado_Prestada.png`](evidencias/10_Laptop_Estado_Prestada.png) |
| 13 | Bloqueo de préstamo duplicado | `POST /api/laptop-loans` repetido | Error 409 | **Aprobado** — HTTP 409, laptop no disponible | [`11_Prestamo_Duplicado_Rechazado.png`](evidencias/11_Prestamo_Duplicado_Rechazado.png) |
| 14 | Mis préstamos de laptop | `GET /api/laptop-loans/me` | 200, préstamo activo | **Aprobado** — HTTP 200, préstamo 1 `active` | [`12_Mis_Prestamos_Laptops.png`](evidencias/12_Mis_Prestamos_Laptops.png) |
| 15 | Devolución de laptop | `PUT /api/laptop-loans/1/return` | 200, devuelve equipo | **Aprobado** — HTTP 200 | [`13_Devolucion_Laptop.png`](evidencias/13_Devolucion_Laptop.png) |
| 16 | Inventario restaurado laptop | `GET /api/laptops` | 200, equipo disponible | **Aprobado** — HTTP 200, estado `available` | [`14_Laptop_Disponible.png`](evidencias/14_Laptop_Disponible.png) |
| 17 | Catálogo original de libros | `GET /api/books` | 200, catálogo | **Aprobado** — HTTP 200, 5 libros | [`15_Catalogo_Libros.png`](evidencias/15_Catalogo_Libros.png) |
| 18 | Mis préstamos de libros inicialmente | `GET /api/loans/me` | 200, sin préstamos | **Aprobado** — HTTP 200, `null` | [`16_Prestamos_Libros.png`](evidencias/16_Prestamos_Libros.png) |
| 19 | Prestar libro original | `POST /api/loans` | 201, plazo 14 días | **Aprobado** — HTTP 201, libro 1, préstamo 1 | [`17_Prestamo_Libro.png`](evidencias/17_Prestamo_Libro.png) |
| 20 | Inventario de libros actualizado | `GET /api/books` | 200, menos 1 copia | **Aprobado** — copias disponibles 3 → 2 | [`18_Copias_Libro_Actualizadas.png`](evidencias/18_Copias_Libro_Actualizadas.png) |
| 21 | Mis préstamos de libros | `GET /api/loans/me` | 200, préstamo activo | **Aprobado** — HTTP 200, préstamo 1 `active` | [`19_Mis_Prestamos_Libros.png`](evidencias/19_Mis_Prestamos_Libros.png) |
| 22 | Devolución y restitución de libro | `PUT /api/loans/1/return` + `GET /api/books` | 200, restaura copias | **Aprobado** — copias 2 → 3 | [`20_Devolucion_y_Restauracion_Libro.png`](evidencias/20_Devolucion_y_Restauracion_Libro.png) |
| 23 | Rol insuficiente para crear laptop | `POST /api/laptops` con JWT estudiante | 403 | **Aprobado** — HTTP 403 | [`21_Estudiante_Sin_Permisos.png`](evidencias/21_Estudiante_Sin_Permisos.png) |
| 24 | Rechazo de devolución duplicada | `PUT /api/laptop-loans/1/return` nuevamente | 409 | **Aprobado** — HTTP 409 | [`22_Devolucion_Duplicada_Rechazada.png`](evidencias/22_Devolucion_Duplicada_Rechazada.png) |
| 25 | Validación estática Go | `go build ./...` y `go vet ./...` | Sin errores | **Aprobado** — sin errores reportados | [`23_Verificacion_Codigo_Git.png`](evidencias/23_Verificacion_Codigo_Git.png) |
| 26 | Commit del módulo | `git commit -m "feat: implementar modulo de prestamos de laptops"` | Commit creado | **Aprobado** — hash `788c9d6` | [`24_Commit_Codigo.png`](evidencias/24_Commit_Codigo.png) |

## Casos adicionales por ejecutar (no se presentan como aprobados)

| Caso | Esperado | Estado |
|---|---|---|
| `POST /api/laptops` sin encabezado `Authorization` | 401 Unauthorized | **Pendiente** (solo se probó `GET /api/laptops` sin token) |
| Préstamo de laptop con ID inexistente | 404 Not Found | **Pendiente** |
| Devolver un préstamo de laptop inexistente | 404 Not Found | **Pendiente** |
| Devolver la laptop de un estudiante distinto | 403 Forbidden | **Pendiente** |
| Crear laptop con el mismo serial | 409 Conflict | **Pendiente** |
| Dos solicitudes de préstamo concurrentes (condición de carrera) | A lo sumo un préstamo aceptado | **Pendiente** |
| `GET /api/loans/me` después de la devolución del libro | Préstamo en estado cerrado | **Pendiente** (se verificó el inventario restaurado) |
| Renderización de los 4 diagramas Mermaid en GitHub | Diagramas visibles y coherentes | **Pendiente de verificar al subir el fork** |

## Análisis de seguridad y reglas de negocio

**Autenticación:** la ruta `GET /api/laptops` sin token devolvió 401 y con JWT de estudiante devolvió 200. **Autorización:** una petición `POST /api/laptops` con token de estudiante devolvió 403; el administrador pudo registrar el equipo con 201.

**Integridad de préstamos:** el estado de la laptop pasó de disponible a prestada; un segundo intento devolvió 409 y evitó un préstamo duplicado. Tras la devolución, la laptop reapareció disponible. El intento de devolver de nuevo el préstamo cerrado produjo 409. Estas pruebas comprueban el comportamiento secuencial observado, no una prueba de simultaneidad real.

**Regresión del sistema original:** se registró un préstamo de un libro durante 14 días y se constató que sus copias disponibles disminuyeron de 3 a 2. La devolución respondió 200 y el inventario volvió a 3. Los endpoints `/health` y `/ping` conservaron su respuesta original.

## Observaciones y límites

1. En la respuesta inicial del registro apareció `created_at` con un valor cero (`0001-01-01`), mientras que en el inicio de sesión posterior se observó la fecha real. Debe revisarse esa diferencia; no se declara corregida.
2. La primera respuesta sin préstamos del módulo de libros fue `null` y no `[]`. Respondió HTTP 200 y el flujo funcional posterior funcionó, pero sería conveniente normalizar la serialización de listas vacías.
3. Las capturas contienen tokens y credenciales **de demostración**. Antes de publicar el repositorio, se recomienda invalidar los tokens, cambiar contraseñas de prueba y no subir el archivo `.env`. Si se conservan las capturas originales sin ocultar, quedan visibles a cualquiera que tenga acceso al repositorio.
4. Las referencias anteriores apuntan a capturas reales dentro de `proyecto-final/evidencias/`; deben conservarse junto al documento al realizar el commit de documentación.

## Conclusión

Las pruebas observadas respaldan que **el ciclo funcional de préstamo y devolución de laptops opera correctamente** en el entorno local y que las operaciones básicas de libros, autenticación, health y ping continúan respondiendo. Las validaciones de préstamo duplicado y devolución repetida fueron rechazadas con HTTP 409; el control de roles diferenció correctamente a estudiante y administrador. Quedan identificados casos negativos complementarios y la comprobación de diagramas en GitHub, que no se presentan como ejecutados.

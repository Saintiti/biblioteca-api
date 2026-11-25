# Fase 1 — Análisis del problema

**Arquitectura de Software — Proyecto final**  
**Integrantes:**
- Christian Javier Benitez Castro — T00082581
- Sean Paul Martinez Roa — T00083011  
**Sistema base:** `github.com/carrenoutb/biblioteca-api`

## 1. Problema y estado actual

La biblioteca ya permite prestar libros, pero ahora necesita incorporar laptops. No basta con cambiar el nombre de un campo: un libro se presta según una cantidad de copias, mientras que una laptop física tiene un serial y no debe asignarse a dos personas al mismo tiempo. El objetivo es agregar la función sin romper autenticación, catálogo y préstamos de libros.

## 2. Mapa del código y puntos de acoplamiento

| Archivo existente | Dependencia actual de libros | Implicación para laptops |
|---|---|---|
| `migrations/001_init.sql` | `loans.book_id` es FK `NOT NULL` a `books` | Un préstamo existente no puede representar directamente una laptop. |
| `internal/models/loan.go` | `Loan.BookID` y `LoanRequest.BookID` | Los contratos JSON aceptan `book_id`, no un serial de laptop. |
| `internal/repository/loan_repository.go` | `INSERT` usa `book_id`; listas hacen `JOIN books` | La capa SQL no puede resolver laptops con esos JOIN. |
| `internal/repository/book_repository.go` | Incrementa/decrementa `available_copies` | Las laptops necesitan estado por unidad, no contador de ejemplares. |
| `internal/service/loan_service.go` | `Borrow` busca libro, revisa copias y vence a 14 días; `Return` incrementa copias | Las reglas son propias de libros. |
| `internal/handler/loan_handler.go` | Usa `LoanRequest` y `LoanService` | Los endpoints actuales hablan de libros. |
| `internal/handler/routes.go` | `/api/loans` usa el handler de libros | Harían falta nuevas rutas o generalizar las actuales. |

**Flujo del préstamo actual:** `POST /api/loans` → middleware JWT → handler → `LoanService.Borrow` → `BookRepository.FindByID` y `LoanRepository.HasActiveLoan` → `INSERT loans` y decremento de `books.available_copies`. El servicio toma la decisión de disponibilidad y fecha; el repositorio interactúa con SQL.

## 3. Opciones arquitectónicas

**Opción A — Módulo paralelo especializado:** crear `laptops` y `laptop_loans`, con modelo, repository, service y handler propios. Mantener `books` y `loans` tal como estaban. Reutilizar Gin, MySQL, JWT y los middleware.

**Opción B — Generalizar todos los préstamos:** crear entidad `resources`, hacer que libros/laptops sean tipos de recurso, y migrar `loans` para que señale a un recurso genérico. Mayor flexibilidad futura, pero requiere migrar las referencias y endpoints existentes.

**Opción C — Agregar `laptop_id` opcional a `loans`:** mantener `book_id` o `laptop_id`, agregando reglas para que exactamente uno tenga valor. Es una modificación menor que B, pero aumenta condicionales y complica cada nuevo recurso.

## 4. Tabla de trade-offs

| Opción | Modificación BD | Qué reutiliza / duplica | Si llega otro recurso | Riesgo sobre libros | Complejidad actual |
|---|---|---|---|---|---|
| **A. Paralelo** | Dos tablas nuevas; no altera `loans` | JWT y estructura por capas; duplica parte del flujo | Otro módulo o refactor posterior | **Bajo** | **Baja-media** |
| B. Generalizar | Cambios en tablas y relaciones existentes; migración de datos | Reutiliza motor de préstamos | Extender el modelo de recursos | **Alto** | Alta |
| C. Columna opcional | Cambia `loans` y sus consultas/validaciones | Reutiliza tabla `loans` | Nuevas columnas y más condiciones | Medio-alto | Media |

## 5. Selección preliminar

Para el tamaño y plazo de este proyecto es prioritario **proteger el comportamiento existente y demostrarlo**. Por eso se selecciona la opción A: inventario de laptops por serial, préstamos independientes, misma autenticación y misma estructura por capas. El costo asumido es duplicar algunas operaciones CRUD de préstamos y que un recurso nuevo necesite otra implementación. La decisión formal, con consecuencias, está en `adr.md`.

## 6. Escenario futuro

Si mañana llegan tablets o proyectores, la alternativa A implicaría otra tabla/módulo, o plantear un refactor gradual hacia una entidad general de recursos. Se pospone esa complejidad hasta que el requisito sea real y exista evidencia de que los recursos comparten suficiente comportamiento.

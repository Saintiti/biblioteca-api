# Fase 5 — Diagramas de arquitectura (Mermaid)

**Integrantes:**
- Christian Javier Benitez Castro — T00082581
- Sean Paul Martinez Roa — T00083011  
Decisión ADR-001: módulo de laptops paralelo, misma autenticación y misma base de datos física.

## 1. C4 Nivel 1 — Contexto

**Pregunta:** ¿Quién usa el sistema y con qué objetivo?

```mermaid
flowchart LR
    E[Estudiante] -->|Consulta y solicita préstamo| S[Sistema Biblioteca API]
    A[Administrador] -->|Registra equipos y gestiona catálogo| S
    S -->|Registra usuarios, libros, laptops y préstamos| D[(Información de biblioteca)]
```

## 2. C4 Nivel 2 — Contenedores

**Pregunta:** ¿Qué ejecutables se comunican? El cliente es el navegador/REST Client del estudiante; no se afirma que exista un frontend React desplegado.

```mermaid
flowchart LR
    C[Cliente HTTP / Swagger / REST Client] -->|HTTP y JWT| API[Contenedor API Go + Gin :8080]
    API -->|TCP 3306 interno| DB[(Contenedor MySQL 8)]
    ADMIN[Administrador desde Adminer :8081] -->|SQL| DB
    API -.->|Autenticación JWT| API
```

## 3. Componentes internos de la API

**Pregunta:** ¿Dónde vive la lógica de laptops sin afectar la de libros?

```mermaid
flowchart TB
    Router[Router Gin y middleware JWT / rol] --> LH[LaptopHandler]
    Router --> LLH[LaptopLoanHandler]
    Router --> BH[BookHandler / LoanHandler existentes]
    LH --> LS[LaptopService]
    LLH --> LLS[LaptopLoanService]
    LS --> LR[LaptopRepository]
    LLS --> LR
    LLS --> LLR[LaptopLoanRepository]
    LR --> L[(Tabla laptops)]
    LLR --> L
    LLR --> LL[(Tabla laptop_loans)]
    BH --> BS[Servicios de libros existentes]
    BS --> BR[Repositorios de libros existentes]
    BR --> B[(Tablas books y loans)]
```

## 4. Diagrama de secuencia — Prestar laptop

**Pregunta:** ¿Cuándo se valida y cómo se evita prestarla dos veces?

```mermaid
sequenceDiagram
    actor U as Estudiante
    participant J as Middleware JWT
    participant H as LaptopLoanHandler
    participant S as LaptopLoanService
    participant R as LaptopLoanRepository
    participant DB as MySQL
    U->>J: POST /api/laptop-loans + token
    alt Token ausente o inválido
        J-->>U: 401 Unauthorized
    else Token válido
        J->>H: user_id validado
        H->>S: Borrow(user_id, laptop_id)
        S->>DB: Consultar laptop (revisión inicial)
        alt Laptop inexistente
            S-->>U: 404 Not Found
        else Laptop no disponible
            S-->>U: 409 Conflict
        else Disponible
            S->>R: BorrowAtomic(user_id, laptop_id, due_at=+7 días)
            R->>DB: BEGIN + SELECT status FOR UPDATE
            alt Otra petición la tomó primero
                R->>DB: ROLLBACK
                R-->>U: 409 Conflict
            else Bloqueo obtenido y disponible
                R->>DB: UPDATE laptops SET status=borrowed
                R->>DB: INSERT laptop_loans(active, due_at)
                R->>DB: COMMIT
                R-->>S: Préstamo creado
                S-->>H: Préstamo
                H-->>U: 201 Created
            end
        end
    end
```

## 5. Secuencia complementaria — Devolver y liberar inventario

```mermaid
sequenceDiagram
    actor U as Usuario
    participant H as LaptopLoanHandler
    participant S as LaptopLoanService
    participant R as LaptopLoanRepository
    participant DB as MySQL
    U->>H: PUT /api/laptop-loans/:id/return + JWT
    H->>S: Return(id, user_id, role)
    S->>R: Buscar préstamo
    R->>DB: SELECT préstamo
    alt No existe
        S-->>U: 404
    else No es titular ni admin
        S-->>U: 403
    else Ya devuelto
        S-->>U: 409
    else Préstamo activo y autorizado
        S->>R: ReturnAtomic(id)
        R->>DB: BEGIN + SELECT préstamo FOR UPDATE
        R->>DB: UPDATE laptop_loans status=returned
        R->>DB: UPDATE laptops status=available
        R->>DB: COMMIT
        S-->>H: OK
        H-->>U: 200, laptop liberada
    end
```

## 6. Coherencia con la implementación

- `models/laptop.go`: estructuras de inventario, request y préstamo.
- `repository/laptop_loan_repository.go`: transacción, bloqueo y SQL.
- `service/laptop_service.go`: duración, estado y autorización de devolución.
- `handler/laptop_handler.go`: validación HTTP y códigos de respuesta.
- `handler/routes.go`: JWT, rutas protegidas y ruta admin.
- `migrations/002_laptops.sql`: tablas `laptops` y `laptop_loans`.

Los diagramas son propuesta de arquitectura del código adjunto y deben contrastarse con el fork efectivamente publicado.

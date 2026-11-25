# ADR-001 — Inventario y préstamos independientes para laptops

- **Estado:** Aceptado
- **Fecha:** 25 de noviembre de 2026
- **Responsables:** Christian Javier Benitez Castro — T00082581; Sean Paul Martinez Roa — T00083011
- **Ámbito:** evolución de `biblioteca-api` para préstamo de laptops

## Contexto

La biblioteca hoy registra libros por cantidad de copias y préstamos asociados obligatoriamente a `book_id`. El nuevo requisito introduce laptops, unidades físicas identificadas por serial. Es indispensable entregar el catálogo, el préstamo, la devolución y el historial de cada estudiante, con JWT y roles, **sin romper los endpoints de libros**. El cambio debe ser explicable y verificable durante la sustentación.

### Criterios (orden de importancia)

1. **Compatibilidad y menor riesgo sobre libros:** el profesor exige que todo lo existente funcione.
2. **Consistencia del inventario:** una laptop no puede estar prestada a dos usuarios simultáneamente, incluso con solicitudes concurrentes.
3. **Simplicidad y mantenibilidad:** conservar el estilo `handler → service → repository`.
4. **Extensibilidad futura:** reconocer el costo de introducir otros recursos.

### Alternativas consideradas

| Alternativa | Ventaja | Costo o riesgo |
|---|---|---|
| **A. Módulo paralelo** | No altera `books` ni `loans`; bajo riesgo de regresión | Duplica parte de las operaciones y no es un modelo universal |
| B. Préstamo genérico de recursos | Una lógica común para libros, laptops y futuros recursos | Refactor de tablas, datos, endpoints y reglas; alto riesgo |
| C. `laptop_id` opcional en `loans` | Conserva una sola tabla de historial | Condiciones para garantizar exclusividad y acoplamiento creciente |

## Decisión

**Implementar el módulo de laptops de manera especializada y paralela al de libros**, con dos tablas nuevas (`laptops`, `laptop_loans`), dos repositorios, servicios y handlers dedicados. Se comparten la conexión a MySQL, el router Gin, el middleware JWT y el control de roles. No se modifica la lógica de `LoanService` de libros.

Las laptops son **equipos individuales identificados por `serial_number` único**, no modelos con copias agregadas. El préstamo tiene una duración de **siete días**, distinta de los 14 días de libros; estudiantes y administradores autenticados pueden pedir prestado, y **solo un admin puede registrar laptops**. El titular o un admin puede devolver la laptop. Son decisiones de negocio del proyecto, no reglas ya existentes en el repositorio.

La regla de no duplicar un préstamo se protege en dos niveles:

- En el **servicio**, se verifica que el equipo exista y esté disponible antes de solicitar el préstamo.
- En el **repositorio**, se usa una transacción InnoDB y `SELECT ... FOR UPDATE`, se marca el equipo como `borrowed` y se inserta el préstamo antes de `COMMIT`. Además, una columna generada con índice único permite solo **un préstamo activo por laptop**. El chequeo transaccional evita carreras entre peticiones.

Los estados del inventario son `available` y `borrowed`; los del préstamo `active` y `returned`. La devolución cambia ambos estados en la misma transacción.

## Consecuencias positivas

- **Bajo acoplamiento con libros:** las tablas y endpoints antiguos conservan su contrato.
- **Alta cohesión:** las reglas de laptops están en su propio módulo.
- **Integridad de inventario:** se evita el doble préstamo por transacción y restricción de BD.
- **Reutilización de infraestructura:** no duplicamos autenticación, JWT, Docker, conexión ni API web.
- **Pruebas aislables:** se pueden verificar por separado los préstamos de libros y laptops.

## Consecuencias negativas y mitigaciones

- **Duplicación moderada:** los servicios de préstamos de libros/laptops tienen operaciones parecidas. Se acepta por el riesgo del refactor; si surgen más recursos, se evaluará extraer una abstracción común.
- **Extensibilidad futura limitada:** incorporar proyectores puede requerir otro módulo; el modelo genérico es una evolución posible, no un requisito actual.
- **Nueva migración y nueva consistencia transaccional:** hay que crear y actualizar tablas. Se documenta una migración `002` y el procedimiento para BD ya inicializada.
- **Swagger existente no se regenera automáticamente:** las pruebas nuevas están documentadas en `PRUEBAS_API.http`; si el profesor requiere Swagger actualizado se debe ejecutar `swag init` tras agregar anotaciones y publicar el código generado.

## Alternativas descartadas

**B** queda descartada por el impacto sobre el módulo actual, cuya estabilidad tiene valor explícito en la rúbrica. **C** queda descartada porque cada futuro recurso exigiría más columnas opcionales, condicionantes y cambios en consultas existentes. A no es “perfecta”: es la opción con mejor relación entre simplicidad, compatibilidad y plazo del proyecto.

## Comprobación de la decisión

La decisión arquitectónica fue comprobada mediante pruebas reales. Verificamos el registro de laptops por parte del administrador (201), la consulta del inventario (200), el préstamo (201), el rechazo de préstamos duplicados (409), la devolución (200) y el rechazo de devoluciones repetidas (409).

También comprobamos la seguridad mediante el rechazo de accesos sin token (401) y de creación de laptops por estudiantes (403). Finalmente, verificamos que el módulo original de libros continuara funcionando correctamente, incluyendo sus préstamos, devoluciones y actualización del inventario.**

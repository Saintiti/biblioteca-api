# Guion de sustentación — 10 minutos + 5 de preguntas

**Integrantes:**
- Christian Javier Benitez Castro — T00082581
- Sean Paul Martinez Roa — T00083011  
**Nota para la defensa:** El proyecto se presenta en pareja; cada integrante debe poder explicar todo el diseño de forma individual. El siguiente guion está en primera persona para que cualquiera de los dos pueda utilizarlo.

**Tema:** Evolución de biblioteca-api: préstamo de laptops sin afectar libros

## Minutos 0:00–2:00 — Problema

«El sistema de biblioteca ya permite registrarse, iniciar sesión, consultar libros y solicitar préstamos. El problema es que está diseñado específicamente para libros. Por ejemplo, la tabla `loans` guarda `book_id` y el servicio usa `available_copies`. Eso no sirve directamente para una laptop que tiene serial individual. El reto era incorporar el préstamo de laptops sin dañar lo que ya funciona. Primero revisé la base de datos y las capas del código, en vez de empezar a copiar archivos.»

**Señalar:** `analisis.md` y código de `LoanService.Borrow` original.

## Minutos 2:00–5:00 — Alternativas y ADR (la parte más importante)

«Evalué tres opciones. La primera, crear un módulo paralelo de laptops; la segunda, convertir todos los préstamos en un sistema genérico de recursos; la tercera, añadir un `laptop_id` opcional a la tabla actual de préstamos. La generalización sería mejor si tuviéramos decenas de tipos de recursos, pero obligaría a migrar el módulo de libros y aumentaría el riesgo de romperlo. La tercera parece rápida, pero cada recurso nuevo agregaría más columnas y condiciones. Elegí la primera porque el profesor nos exige compatibilidad y porque hoy solo hay dos tipos de recurso.»

«Esta decisión **gana simplicidad, cohesión y bajo riesgo**, pero **pierde reutilización de la lógica**: algunas operaciones se duplican, y para agregar proyectores habría que crear otro módulo o rediseñar. No digo que sea perfecta; digo que es una decisión razonable según nuestras prioridades.»

**Señalar:** tabla de trade-offs y ADR (contexto, decisión, consecuencias positivas y negativas).

## Minutos 5:00–8:00 — Implementación y diagramas

«Las laptops se registran por serial individual en `laptops` y sus préstamos en `laptop_loans`. El inventario tiene estado `available` o `borrowed`. Cuando alguien solicita un préstamo, el middleware JWT identifica al usuario, el handler recibe la solicitud y el servicio decide si se puede prestar y calcula siete días de vencimiento. El repositorio ejecuta una transacción: bloquea el equipo, cambia el estado y crea el préstamo. Si falla una operación, revierte todo. Además hay una restricción única que impide dos préstamos activos de la misma laptop.»

«No modifiqué la lógica de libros. Reutilicé las mismas capas: handler, service y repository. Solo admin registra equipos. Estudiantes y administradores autenticados pueden prestar. El propietario del préstamo o un admin puede devolver.»

**Señalar:** C4 Nivel 1 y 2, componentes, secuencia, migración SQL y `LaptopLoanRepository.BorrowAtomic`.

## Minutos 8:00–10:00 — Verificación real

«La validación debe comprobar tanto las nuevas funciones como las antiguas. En el módulo nuevo debemos registrar una laptop, listarla, prestarla, intentar prestarla nuevamente y ver un error 409. Después devolverla y verificar que vuelve a estar disponible. También hay pruebas de seguridad: una solicitud sin token debe devolver 401, y un estudiante no puede registrar laptops: 403. Finalmente debo mostrar que `/health`, `/ping`, el login, el catálogo y los préstamos de libros siguen funcionando.»

**IMPORTANTE:** en la defensa, di «probé» únicamente en los casos que efectivamente hayas ejecutado y tengas capturas. No inventar resultados.

## Preguntas posibles (respuestas cortas)

**1. ¿Por qué no generalizaste `loans`?** Porque exigía modificar tablas, servicios y endpoints existentes. Para el plazo y el requisito de no romper libros, el riesgo era mayor que el beneficio inmediato.

**2. ¿Qué sacrificaste?** Duplico parte de la lógica de préstamos y un tercer tipo de recurso costaría más que con un modelo genérico.

**3. ¿Qué pasaría si mañana prestaran proyectores?** Analizaría si conviene otro módulo o extraer un módulo genérico; sería una nueva decisión ADR, no improvisar nuevas columnas.

**4. ¿Cómo evitas prestar la misma laptop dos veces?** Primero compruebo disponibilidad en service y en la transacción bloqueo la fila `FOR UPDATE`; la base de datos también tiene un índice único para los préstamos activos.

**5. ¿Qué pasaría si falla INSERT después de cambiar el estado?** Un `ROLLBACK` revierte ambos cambios; no queda el inventario marcado sin préstamo.

**6. ¿Para qué sirven las capas?** Handler recibe HTTP; service decide reglas; repository ejecuta SQL. Facilita mantenimiento y reduce acoplamiento.

**7. ¿Por qué serial y no copias?** Una laptop es un bien físico identificable. Si una Lenovo tiene un daño o está prestada, necesitamos saber cuál es exactamente.

**8. ¿Por qué siete días?** Es una regla propuesta para este proyecto; un equipo costoso se rota con mayor frecuencia que un libro de 14 días. Puede parametrizarse después.

**9. ¿Cómo sabes que libros sigue funcionando?** Con evidencia de registro/login, listado, préstamo y devolución de libros antes y después del cambio, no solo revisando código.

**10. ¿Por qué no basta con probar el caso exitoso?** También se requieren casos de error: doble préstamo, devolución repetida, falta de token y falta de permiso.

**11. ¿Qué es un ADR?** Un registro de decisión arquitectónica: contexto, alternativas, decisión y consecuencias. Sirve para justificar por qué elegimos una solución.

**12. ¿Qué significa extensibilidad?** Cuánto esfuerzo cuesta incorporar nuevas funciones sin tener que modificar demasiadas partes del sistema.

## Demostración sugerida en vivo

Abrir tres ventanas: (1) `docker compose ps`, (2) `proyecto-final/adr.md` y sus diagramas, y (3) `PRUEBAS_API.http`. Mostrar catálogo, préstamo, rechazo del duplicado y devolución. Conservar capturas por si falla la conexión durante la sustentación.

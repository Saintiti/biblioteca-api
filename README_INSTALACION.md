# Proyecto final — Módulo de préstamo de laptops

**Integrantes:**
- Christian Javier Benitez Castro — T00082581
- Sean Paul Martinez Roa — T00083011  
**Modalidad:** Trabajo en pareja  
**Curso:** Arquitectura de Software  
**Repositorio base:** https://github.com/carrenoutb/biblioteca-api

Este paquete contiene **los archivos nuevos y los archivos que deben sustituirse** en un fork del repositorio del profesor. **No es un repositorio independiente:** los archivos de libros, autenticación, Dockerfile, docs y demás código de origen siguen proviniendo del fork original.

## A. Contenido entregado

```text
cmd/api/main.go                            [actualizado: cableado de laptops]
internal/handler/routes.go                 [actualizado: rutas nuevas sin quitar las antiguas]
internal/handler/laptop_handler.go         [nuevo]
internal/models/laptop.go                  [nuevo]
internal/repository/laptop_repository.go   [nuevo]
internal/repository/laptop_loan_repository.go [nuevo]
internal/service/laptop_service.go         [nuevo]
migrations/001_init.sql                    [corregido SOLO el INSERT inicial del repo público]
migrations/002_laptops.sql                 [nuevo]
docker-compose.yml                         [actualizado: 3 servicios, puerto host 3307]
PRUEBAS_API.http                           [nuevo: pruebas y regresión]
proyecto-final/analisis.md                 [Fase 1]
proyecto-final/adr.md                      [Fase 2]
proyecto-final/diagramas.md                [Fase 5]
proyecto-final/verificacion.md             [Fase 4: plantilla pendiente de evidencias]
proyecto-final/GUIA_SUSTENTACION.md        [guion de 10 min + preguntas]
```

## B. Integrar en el FORK (no en el repositorio del docente)

1. En GitHub abre https://github.com/carrenoutb/biblioteca-api y haz clic en **Fork** para crear tu copia.
2. Desde CMD (cambia `TU_USUARIO` por tu usuario real):

```cmd
cd C:\Users\Lenovo\Downloads
git clone https://github.com/TU_USUARIO/biblioteca-api.git
cd biblioteca-api
```

3. Extrae el ZIP recibido. **Copia el contenido interno de la carpeta extraída** sobre esta carpeta `biblioteca-api` del fork, **combinando carpetas y reemplazando** `cmd/api/main.go`, `internal/handler/routes.go`, `docker-compose.yml` y `migrations/001_init.sql` cuando Windows lo pregunte. No copies el ZIP como una subcarpeta adicional: `proyecto-final/` debe quedar en la raíz.
4. Desde la raíz verifica los cambios:

```cmd
git status
```

**Atención:** este paquete se preparó con la rama `main` pública consultada el 25/11/2026. Si tu fork tiene cambios propios en `main.go`, `routes.go`, `docker-compose.yml` o `001_init.sql`, compara antes de sobreescribirlos.

## C. Configuración de Windows + Docker

En el CMD situado en `biblioteca-api`:

```cmd
copy .env.example .env
docker compose config
```

El Compose incluido usa **3307 de Windows** hacia **3306 de MySQL**. La API se conecta por `DB_HOST=mysql` y `DB_PORT=3306` DENTRO de Docker; no cambies `DB_PORT` a `3307`. Adminer usa `http://localhost:8081` y la API `http://localhost:8080`.

Si antes tenías contenedores antiguos llamados `biblioteca_mysql`, `biblioteca_api` o `biblioteca_adminer`, este compose ya no fija `container_name`, para evitar colisiones de nombre. Aun así, si algún proceso usa `8080`, `8081` o `3307`, detén ese proceso o cambia el **primer** número del puerto correspondiente en Compose.

### Una base de datos NUEVA

```cmd
docker compose up --build -d
docker compose ps
```

El contenedor MySQL inicializará `001_init.sql` y `002_laptops.sql` en orden; no necesitas importar `002` manualmente si el volumen recién se crea.

### Una base de datos que YA tenía usuarios y libros

Los scripts de `docker-entrypoint-initdb.d` **NO se reejecutan con un volumen existente**. Con Docker arriba, ejecuta una sola vez:

```cmd
docker compose exec -T mysql mysql -u biblioteca_user -pbiblioteca_pass biblioteca < migrations\002_laptops.sql
```

La contraseña usada aquí es la del `docker-compose.yml` de ejemplo; si la cambiaste, usa tus credenciales de ese entorno. Para comprobar:

```cmd
docker compose exec mysql mysql -u biblioteca_user -pbiblioteca_pass biblioteca -e "SHOW TABLES;"
```

Deben aparecer `users`, `books`, `loans`, `laptops` y `laptop_loans`. **No ejecutes `docker compose down -v` para arreglar algo sin antes considerar que borra el volumen de los datos anteriores.**

## D. Importante: defecto preexistente de 001_init.sql

Al leer la versión pública del repositorio, el `INSERT INTO books` terminaba después de *The Pragmatic Programmer* con `;` y luego aparecía otra fila suelta para *The Go Programming Language*. En una base de datos nueva eso es SQL inválido. El archivo `001_init.sql` incluido únicamente corrige **ese separador a coma**, sin cambiar tablas ni lógica de los préstamos de libros. Si en tu fork el profesor ya corrigió este defecto, conserva su versión correcta y no necesitas reemplazarla.

## E. Compilar y verificar

```cmd
go build ./...
go vet ./...
docker compose logs --tail=70 api
docker compose ps
```

En un navegador abre:

- http://localhost:8080/health → `{"status":"ok"}`
- http://localhost:8080/ping → `{"message":"pong"}`
- http://localhost:8080/swagger/index.html → interfaz Swagger del proyecto anterior

**Aviso sobre Swagger:** el Swagger generado que venía en `docs/` refleja el proyecto anterior. Los endpoints nuevos se prueban mediante `PRUEBAS_API.http`. Si necesitas añadirlos al Swagger interactivo, deben agregarse anotaciones y regenerar los archivos con la CLI `swag`, no editar a mano los archivos generados.

## F. Preparar admin y estudiante para probar

En VS Code instala **REST Client** (Huachao Mao), abre `PRUEBAS_API.http` y ejecuta primero los registros de **admin** y **estudiante**.

El endpoint `/api/auth/register` crea siempre un estudiante. Para convertir el usuario de demostración en administrador, usa este comando DESPUÉS de registrarlo:

```cmd
docker compose exec mysql mysql -u biblioteca_user -pbiblioteca_pass biblioteca -e "UPDATE users SET role='admin' WHERE email='admin.laptops@example.com'; SELECT id,email,role FROM users WHERE email='admin.laptops@example.com';"
```

**Vuelve a iniciar sesión como admin después de la actualización**, porque el JWT contiene el rol al momento del login. Ejecuta después las demás peticiones del archivo `.http`, en orden: listar, crear, prestar, intentar duplicado, listar préstamos y devolver. Comprueba también 401 y 403.

Si registraste antes esas cuentas y responde que el correo ya existe, puedes usar correos de prueba nuevos en las solicitudes o trabajar con el usuario existente.

## G. Criterios de evidencia y entrega

1. Completa `proyecto-final/verificacion.md` con **capturas reales**; actualmente contiene `Pendiente` honestamente.
2. Guarda las imágenes en `proyecto-final/evidencias/` con los nombres del checklist.
3. Verifica el ciclo original de libros (listar, prestar, devolver) y las rutas `/health`, `/ping`, auth.
4. Visualiza los Mermaid de `proyecto-final/diagramas.md` en GitHub y confirma que se dibujan.
5. Revisa el ADR para poder defender por qué se eligió el módulo paralelo y qué desventajas tiene.

### Commits por hitos (en lugar de un solo commit)

Si aún NO hiciste commits, desde el fork puedes hacerlos por conjuntos:

```cmd
git add migrations/001_init.sql migrations/002_laptops.sql proyecto-final/analisis.md proyecto-final/adr.md
git commit -m "docs: analisis ADR y migracion para laptops"

git add cmd/api/main.go internal docker-compose.yml PRUEBAS_API.http
git commit -m "feat: catalogo y prestamos de laptops por capas"

git add proyecto-final README_INSTALACION.md
git commit -m "docs: diagramas verificacion y defensa"

git push origin main
```

Si tu fork usa una rama distinta, sustituye `main` por esa rama. Antes de `push`, ejecutar las pruebas y añadir capturas. Nunca subas `.env` ni contraseñas reales.

**Entrega oficial:** comparte la URL de TU fork con `proyecto-final/` en la raíz; el plazo indicado por el docente es el 25/11/2026 a las 5:00 p. m.

## H. Lo que falta verificar en la máquina del estudiante

Este paquete fue preparado con base en el código público del repositorio y comprobado sintácticamente en sus archivos Go. **No se pudo ejecutar Docker ni MySQL en este entorno**, por lo que no se afirma que estén verificadas la compilación integral, las migraciones o las pruebas de API. Es obligatorio ejecutar el checklist y guardar evidencia antes de sustentar.

# ADR-0003: Usar Git para control de versiones

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante solicitud explícita
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El handbook y el futuro sistema necesitan historial, revisión y trazabilidad entre
decisiones, documentación, código y pruebas.

## Contexto y restricciones

El directorio del proyecto todavía no tenía control de versiones. El usuario ha
pedido crear el repositorio Git. El flujo de hosting, ramas y pull requests aún no
se ha decidido.

## Alternativas

### Git

Es el estándar de facto del ecosistema objetivo, permite trabajo local,
ramificación, revisión y amplia integración con CI/CD. Su coste es aprender un
modelo distribuido y mantener higiene de commits y ramas.

### Otro sistema de control de versiones

Puede resolver el historial, pero ofrece menos integración con el stack y no
aporta una ventaja identificada para este proyecto.

### Mantener archivos sin VCS

No añade proceso, pero impide auditoría fiable, colaboración y recuperación
granular.

## Decisión del usuario

Inicializar un repositorio Git local para TournamentsManager.

## Consecuencias

- La rama inicial se denomina `main`.
- Se versionan handbook, ADR, infraestructura y código.
- Secretos, dependencias descargadas, artefactos de build y estado local se
  excluyen mediante `.gitignore`.
- Los finales de línea y archivos binarios se normalizan mediante
  `.gitattributes`.
- El proveedor remoto y el flujo de ramas permanecen pendientes.

## Validación

Git debe mostrar un repositorio válido en `main`, sin archivos sensibles o
artefactos locales preparados para commit.

## Disparadores de revisión

- Elección de proveedor remoto.
- Incorporación de más colaboradores.
- Necesidad de proteger ramas, firmar commits o adoptar trunk-based development.

## Documentación afectada

- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [DEVELOPMENT.md](../../DEVELOPMENT.md)
- [DECISIONS.md](../../DECISIONS.md)

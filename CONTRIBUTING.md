# Contribuir

## Antes de cambiar

1. Lee el manifiesto y el documento especializado.
2. Comprueba [DECISIONS.md](docs/governance/DECISIONS.md) y los ADR vigentes.
3. Formula el problema y los criterios de aceptación.
4. Si es una decisión importante, detén la implementación hasta que el usuario la
   acepte.

## Forma de una propuesta

Una propuesta debe incluir:

- problema y evidencia;
- alcance y no alcance;
- alternativas;
- trade-offs y coste de mantenimiento;
- recomendación, claramente etiquetada;
- riesgos y validación;
- documentos que cambiarán.

## Forma de un cambio

Mantén el cambio pequeño y coherente. Incluye:

- implementación o contenido;
- pruebas o comprobaciones;
- documentación;
- ADR cuando corresponda;
- entrada en `CHANGELOG.md` si es relevante para usuarios u operación.

## Comprobaciones de Go

Para cambios Go:

```bash
make check
```

Antes de subir una rama:

```bash
make verify
```

`make format`, `make tidy`, `make tidy-tools` y `make tidy-all` modifican
archivos. Revisa su diff antes de incluirlo en un commit. Los targets `check` y
`verify` solo comprueban y no corrigen silenciosamente.

## Comprobaciones TypeScript y cliente Expo

El runtime y el package manager están fijados por el repositorio. Tras clonar:

```bash
pnpm install
```

Durante un cambio:

```bash
pnpm run check
```

`make check` ejecuta las comprobaciones aplicables de Go y TypeScript desde una
única entrada. `make verify` añade la exportación web del cliente Expo a un
directorio temporal, además de las comprobaciones de generación, build y
vulnerabilidades; no deja un artefacto versionado en el repositorio.

## Revisión

La revisión pregunta:

- ¿respeta el manifiesto?
- ¿la decisión pertenece al usuario y está registrada?
- ¿es la solución más simple suficiente?
- ¿las dependencias apuntan en la dirección acordada?
- ¿se puede probar, observar, operar y revertir?
- ¿documentación y comportamiento coinciden?

## Commits y ramas

`main` es la rama estable y `develop` la rama de integración diaria. El flujo
aceptado está registrado en
[ADR-0013](docs/adr/0013-use-develop-as-integration-branch.md).

- el trabajo ordinario se hace y se commitea en `develop`;
- mientras el trabajo sea individual, no se crean ramas por feature como norma;
- un bloque coherente y verificado se integra de `develop` a `main`;
- después de la integración, el trabajo continúa en `develop`;
- hotfixes, experimentos arriesgados o trabajo paralelo pueden usar ramas
  temporales cuando aporten aislamiento real;
- cada commit debe representar un cambio coherente;
- el mensaje explica la intención, no solo los archivos modificados;
- no se reescribe el historial compartido de `main` ni `develop`;
- no se versionan secretos ni artefactos generados;
- una decisión importante enlaza su ADR.

Cuando haya equipo, trabajo paralelo o necesidad real de seleccionar
funcionalidades por entrega, [ADR-0024](docs/adr/0024-use-ecr-and-digest-based-image-promotion.md)
habilita ramas de feature y `release/<versión>` temporales. Una release nace de
`main` con las features seleccionadas, se valida en `staging` y, si se acepta,
se integra en `main`; después `main` se sincroniza hacia `develop`. No existe
una rama permanente `staging`.

CI ejecuta `make verify` en cada push a `develop` y `main`, y en pull requests
cuando existan. Es una señal adicional, no un check obligatorio: mientras haya
un solo desarrollador, una auto-revisión no añade independencia. Antes de
promover un bloque a `main`, ejecuta `make verify` localmente y revisa el
resultado remoto disponible. No se permiten force-pushes ni borrado de `main` o
`develop`. Véase [ADR-0021](docs/adr/0021-use-advisory-ci-with-local-quality-gate.md).

Al ser un repositorio público, ninguna contribución externa se ejecuta con
secretos o acceso de despliegue. Los cambios en `.github/workflows`, permisos y
rutas de producción requieren revisión específica.

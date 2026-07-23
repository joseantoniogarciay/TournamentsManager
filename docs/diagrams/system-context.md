# Contexto inicial del sistema

> Estado: conceptual; no define tecnología ni despliegue.
>
> Fuente funcional: [PRODUCT.md](../../PRODUCT.md)

```mermaid
flowchart LR
    Guest["Invitado"] -->|"Consulta torneos públicos"| Web["Aplicación web"]
    User["Usuario autenticado"] -->|"Crea, se une y consulta"| Web
    User -->|"Crea, se une y consulta"| Mobile["Aplicación mobile"]

    Web -->|"Contrato API"| API["TournamentsManager API"]
    Mobile -->|"Contrato API"| API

    API --> Identity["Identidad / sesiones"]
    API --> Tournament["Torneos y participación"]
    Tournament --> Data["Persistencia"]
    Identity --> Mail["Verificación / recuperación"]
```

## Lectura

- Web y mobile son clientes del mismo producto, no fuentes de reglas de negocio.
- El invitado solo necesita la aplicación web para el primer vertical slice; el
  acceso invitado mobile se conserva como capacidad futura.
- Identidad demuestra quién es la persona.
- La API decide qué puede hacer sobre cada torneo.
- Persistencia, email y proveedores son detalles externos.

Los límites exactos se decidirán tras concretar el MVP y la estrategia de
identidad.

# ADR-0063: Persistir el locale de cuenta desde el registro para localizar emails

- **Estado:** Aceptado
- **Fecha:** 2026-07-31
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El correo de verificación está escrito fijo en español, aunque el cliente
universal ya soporta español, inglés, italiano y francés. El backend necesita
conocer una preferencia de idioma de la cuenta para localizar los emails sin
crear una configuración separada exclusivamente para correo.

## Contexto y restricciones

- ADR-0056 fija los locales soportados: `es`, `en`, `it` y `fr`; un locale no
  soportado usa inglés como fallback en el cliente.
- El alta local ya crea una cuenta pendiente antes de enviar el email de
  verificación; el contrato OpenAPI es la fuente de verdad del límite HTTP.
- El locale aportado por el cliente es una preferencia de presentación, nunca
  una afirmación de identidad ni una autorización.
- ADR-0053 permite reescribir el esquema inicial y resetear la base local. No
  hay cuentas existentes que migrar ni compatibilidad que preservar.
- No se añade en este corte una pantalla ni una preferencia independiente para
  el idioma de correo.

## Criterios de decisión

1. entregar el email de alta en el idioma efectivo del cliente;
2. disponer de una única preferencia reutilizable para futuros emails;
3. mantener la validación y la fuente de verdad en el backend;
4. evitar endpoints, sincronización periódica y ajustes de producto antes de
   que exista una necesidad demostrada.

## Alternativas

### A — Persistir el locale validado durante el registro

El cliente incluye su locale efectivo en el alta. El backend acepta únicamente
los cuatro valores soportados, lo persiste en `accounts` y el mailer selecciona
el asunto y las partes de texto/HTML correspondientes.

- **Ventajas:** la preferencia queda disponible para el email inicial y los
  posteriores; no duplica ajustes; el backend no depende del estado actual del
  dispositivo al entregar correo.
- **Inconvenientes:** amplía el contrato, el modelo de cuenta y las plantillas
  localizadas.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo; cada nuevo email debe tener contenido para
  los locales admitidos.

### B — Enviar el locale solo para el email de registro

El backend localiza el primer email con el valor recibido, pero no guarda la
preferencia.

- **Ventajas:** menor cambio de datos inicial.
- **Inconvenientes:** los emails posteriores no tienen una preferencia fiable y
  fuerzan volver a decidir el mecanismo de selección.
- **Coste de mantenimiento:** medio por la solución temporal y su sustitución.

### C — Configurar por separado el idioma de emails

La cuenta mantiene una preferencia específica para correo, distinta del idioma
efectivo de la aplicación.

- **Ventajas:** permite que interfaz y correo usen idiomas distintos.
- **Inconvenientes:** introduce una decisión y una interfaz sin necesidad
  actual, además de posibles discrepancias difíciles de explicar.
- **Coste de mantenimiento:** medio.

### D — Sincronizar el locale al abrir Home desde el primer corte

El cliente comprueba periódicamente el idioma efectivo y actualiza la cuenta en
una operación autenticada.

- **Ventajas:** los emails pueden seguir automáticamente cambios de idioma del
  sistema o navegador.
- **Inconvenientes:** añade un endpoint, reglas de frecuencia y sincronización
  de estado sin ser necesario para localizar el alta.
- **Coste de mantenimiento:** medio.

### No cambiar

El correo de verificación se mantiene solo en español, sin atender al idioma de
la persona.

## Recomendación

**Opinión/recomendación:** alternativa A. Resuelve la necesidad actual con una
preferencia de cuenta única y pospone la sincronización automática hasta que
aporte valor real.

## Decisión del usuario

**Aceptada el 2026-07-31:** elegir la alternativa A.

- `RegisterRequest` incluirá un locale de los valores `es`, `en`, `it` o `fr`.
- El backend lo validará y lo persistirá en la cuenta pendiente desde su
  creación.
- El email de verificación, incluido asunto, texto plano, HTML y atributo
  `lang`, se localizará con esa preferencia.
- La preferencia pertenece a la cuenta y será la fuente para futuros emails.
- No se crea una configuración de idioma de correo ni se implementa aún la
  reconciliación automática desde Home. Si más adelante se necesita, se
  diseñará una actualización autenticada de la misma preferencia, no otra
  configuración.
- Se reescribirá el único esquema inicial y se reseteará PostgreSQL local; no
  habrá backfill ni tratamiento de cuentas anteriores.

## Consecuencias

### Positivas

- La primera comunicación llega en el idioma de la aplicación al registrarse.
- Las entregas posteriores reutilizan una preferencia persistente y explícita.
- La regla queda acotada en el dominio de identidad, sin acoplarla al proveedor
  SMTP.

### Negativas y deuda aceptada

- Cada plantilla de email debe mantenerse en los cuatro locales.
- Si una persona cambia después el idioma de su sistema o navegador, los emails
  conservarán el idioma registrado hasta que se decida e implemente una futura
  actualización autenticada.

## Validación

- El contrato rechaza un locale ausente o fuera de la lista permitida.
- Un registro válido persiste el locale y entrega asunto, texto plano y HTML en
  el idioma correspondiente.
- La plantilla HTML declara el atributo `lang` del locale usado.
- Tras un reset local, el esquema inicial contiene la columna y restricción de
  locale sin migraciones incrementales adicionales.

## Disparadores de revisión

- Se añade o elimina un idioma soportado por el producto.
- Soporte demuestra una necesidad frecuente de cambiar el idioma de los emails.
- Se implementa la lectura de preferencias de cuenta o se confirma que el
  idioma efectivo del cliente debe sincronizarse automáticamente.

## Documentación afectada

- [Modelo inicial](../engineering/INITIAL_DATA_MODEL.md)
- [Identidad y acceso](../engineering/IDENTITY.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)

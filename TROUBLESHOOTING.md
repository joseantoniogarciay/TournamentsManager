# Troubleshooting

> Este documento es el índice de diagnóstico. Los procedimientos ejecutables
> viven en `docs/runbooks`.

## Método

1. Definir el síntoma desde el punto de vista del usuario.
2. Confirmar alcance, inicio y último cambio conocido.
3. Recoger evidencia sin modificar el sistema.
4. Seguir el flujo de la petición y comprobar dependencias.
5. Mitigar primero si existe impacto.
6. Verificar recuperación.
7. Registrar causa, evidencia y prevención.

## Antes de que exista código

| Síntoma | Comprobación |
|---|---|
| Una decisión parece final pero no tiene ADR | Revisar [DECISIONS.md](DECISIONS.md) |
| Dos documentos se contradicen | Aplicar la precedencia de [README.md](README.md) |
| Un enlace está roto | Ejecutar la validación documental de Fase 0 |
| Una tecnología aparece sin comparación | Abrir propuesta con el playbook de decisión |
| Una fase avanza sin retrospectiva | Completar la plantilla de retrospectiva |

## Formato de una entrada futura

- **Síntoma**
- **Impacto**
- **Diagnóstico seguro**
- **Causa probable**
- **Mitigación**
- **Recuperación**
- **Verificación**
- **Escalado**
- **Prevención**

Una colección de comandos sin contexto no es un runbook.

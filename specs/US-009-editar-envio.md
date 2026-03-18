# US-009 — Editar envío

**Como** Supervisor
**Quiero** editar un envío confirmado
**Para** corregir datos ingresados incorrectamente

---

## Criterios de aceptación

1. Solo se pueden editar envíos en estado `in_progress`.
2. Al guardar, los cambios se persisten y se registra en auditoría quién y cuándo editó.

> **Estado de implementación**
> Esta funcionalidad **no está implementada**. Actualmente solo los borradores (`pending`) pueden editarse via `PATCH /shipments/:id/draft`. No existe un endpoint para modificar los datos de un envío confirmado (`in_progress`). La spec define el comportamiento esperado para cuando se implemente.

---

## Diferencia con la edición de borradores (US-002)

| Aspecto             | Borrador (`pending`)               | Envío confirmado (`in_progress`)         |
|---------------------|------------------------------------|------------------------------------------|
| Endpoint            | `PATCH /shipments/:id/draft`       | `PATCH /shipments/:id` *(a implementar)* |
| Validación          | Sin validación de campos requeridos | Misma validación que al crear            |
| Tracking ID         | `DRAFT-XXXXXXXX`, no cambia        | `LT-XXXXXXXX`, no cambia                 |
| Registro de auditoría | No aplica (borrador sin historial) | Requerido: evento con usuario y timestamp |
| Estados permitidos  | Solo `pending`                     | Solo `in_progress`                        |

---

## Reglas de negocio

1. Solo se pueden editar envíos en estado `in_progress` — en cualquier otro estado la edición es rechazada.
2. Los campos editables son los mismos que en el alta: datos del remitente, destinatario, direcciones, peso, tipo de paquete, instrucciones especiales y sucursal receptora. El `tracking_id` y el `status` no son editables.
3. Los campos requeridos (validados en el alta) siguen siendo requeridos al guardar — no se pueden dejar en blanco.
4. El servidor registra un evento de auditoría con: tipo `"edited"`, usuario responsable (`changed_by`) y timestamp.
5. Solo Supervisor y Admin pueden editar un envío confirmado.
6. Una vez que el envío avanza a `in_transit` u otro estado posterior, ya no puede editarse.

---

## Diseño del endpoint

```
PATCH /api/v1/shipments/:tracking_id
```

| Campo              | Notas                                          |
|--------------------|------------------------------------------------|
| Body               | Mismos campos que `CreateShipmentPayload`      |
| `changed_by`       | Requerido — usuario que realiza la edición     |
| Respuesta exitosa  | `200 OK` con el envío actualizado              |
| Rol no autorizado  | `403 Forbidden`                                |
| Estado inválido    | `400 Bad Request` — "only in_progress shipments can be edited" |

---

## Comportamiento del frontend esperado

1. En `ShipmentDetail`, cuando el envío está en `in_progress` y el usuario tiene rol `supervisor` o `admin`, mostrar un botón **"Editar"** o un formulario inline similar al de borradores.
2. Al guardar, llamar a `PATCH /api/v1/shipments/:id` con los datos actualizados.
3. Mostrar confirmación de éxito y refrescar los datos del detalle.
4. Si el servidor rechaza (campo faltante, estado inválido), mostrar el error inline.

---

## Escenarios

### CA1 — Editar envío en in_progress exitosamente

- **Dado** que el envío `LT-XXXXXXXX` está en `in_progress`
- **Y** el usuario tiene rol `supervisor` o `admin`
- **Cuando** hace `PATCH /api/v1/shipments/LT-XXXXXXXX` con datos corregidos y `changed_by`
- **Entonces** el servidor responde `200 OK` con el envío actualizado
- **Y** se registra un evento de auditoría con el usuario y timestamp de la edición

### CA2 — Intento de editar envío en estado posterior a in_progress

- **Dado** que el envío está en `in_transit`, `at_branch` o cualquier estado posterior
- **Cuando** se intenta `PATCH /api/v1/shipments/:id`
- **Entonces** el servidor responde `400 Bad Request` con `"only in_progress shipments can be edited"`
- **Y** los datos del envío no cambian

### CA3 — Edición con campo requerido vacío

- **Dado** que el envío está en `in_progress`
- **Cuando** se intenta guardar con `sender_name` vacío
- **Entonces** el servidor responde `400 Bad Request`
- **Y** los datos del envío no cambian

### CA4 — La edición queda registrada en auditoría

- **Dado** que se editó un envío en `in_progress`
- **Cuando** cualquier usuario consulta `GET /shipments/:id/events`
- **Entonces** aparece un evento con `changed_by` igual al usuario que editó y el timestamp de la operación

### CA5 — Operador no puede editar un envío confirmado

- **Dado** que el envío está en `in_progress`
- **Cuando** el operador intenta hacer `PATCH /api/v1/shipments/:id`
- **Entonces** el servidor responde `403 Forbidden`

### CA6 — Driver no puede editar envíos

- **Dado** que el driver tiene un token válido
- **Cuando** hace `PATCH /api/v1/shipments/:id`
- **Entonces** el servidor responde `403 Forbidden`

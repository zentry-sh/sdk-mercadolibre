# Roadmap de Implementacion - Proximas Fases

## Pre-requisitos: API Keys de Mercado Libre / Mercado Pago

### Que necesitas obtener

Para desarrollo y testing necesitas **dos conjuntos de credenciales**:

#### 1. Credenciales de Test (inmediato, sin aprobacion)
- **Donde**: https://www.mercadopago.com.pe/developers/panel/app
- **Crear una aplicacion** en el panel de desarrollador
- Obtendras automaticamente:
  - `Access Token` de prueba (TEST-xxxx)
  - `Public Key` de prueba
  - `Client ID`
  - `Client Secret`
- Estas credenciales permiten operar en **sandbox** sin dinero real

#### 2. Credenciales de Produccion (requiere aprobacion)
- Se activan desde el mismo panel despues de completar requisitos
- Incluyen los mismos 4 campos pero para operaciones reales
- **No necesitas estas hasta fase de integracion real**

### Que credenciales usa el SDK

| Campo SDK | Credencial MP | Uso |
|-----------|---------------|-----|
| `AccessToken` | Access Token | Autorizacion de todas las requests API |
| `ClientID` | Client ID | OAuth2 flow (para apps multi-usuario) |
| `ClientSecret` | Client Secret | OAuth2 flow (para apps multi-usuario) |
| `WebhookSecret` | Secret key | Validacion HMAC de webhooks |

### Pasos para empezar a testear

1. Ir a https://www.mercadopago.com.pe/developers/panel/app
2. Crear nueva aplicacion (seleccionar Peru como pais)
3. En la seccion de credenciales de test, copiar `Access Token`
4. Para webhooks: ir a Webhooks > Configure notification > copiar/revelar la secret key
5. Usar esas credenciales en:
   ```go
   sdk.New(sdk.Config{
       AccessToken:   "TEST-xxxx-xxxx",
       Country:       "PE",
       WebhookSecret: "tu-secret-key",
   })
   ```

### Notas importantes
- Cada **pais** que quieras testear requiere una cuenta de Mercado Pago en ese pais
- Para Peru: https://www.mercadopago.com.pe/developers
- Para Mexico: https://www.mercadopago.com.mx/developers
- Las credenciales de test permiten crear pagos simulados, no se cobra dinero real
- El `user_id` (para QR/Stores) se obtiene del Access Token: `GET /users/me`

---

## Fase 2: Implementacion de Shipments

### Objetivo
Implementar el adapter de Mercado Libre Shipments con CRUD completo, tracking y etiquetas.

### API Endpoints Confirmados (api.mercadolibre.com)

| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Obtener envio | GET | `/shipments/{id}` |
| Historial de estados | GET | `/shipments/{id}/history` |
| Lead time / tracking | GET | `/shipments/{id}/lead_time` |
| Items del envio | GET | `/shipments/{id}/items` |
| Costos | GET | `/shipments/{id}/costs` |
| Carrier info | GET | `/shipments/{id}/carrier` |
| Etiqueta PDF | GET | `/shipments/{id}/labels` |
| Delays | GET | `/shipments/{id}/delays` |

**Header requerido**: `x-format-new: true` (para JSON actualizado)

**Nota**: La API de ML para shipments es mayormente de **lectura**. Los envios se crean
automaticamente cuando un comprador paga una orden en Mercado Libre. No hay `POST /shipments`
directo - el envio se asocia a una orden existente.

### Archivos a crear/modificar

```
providers/mercadolibre/shipment/
  models.go         <- Structs de request/response ML
  mapper.go         <- ML models <-> domain models
  adapter.go        <- Implementar todos los metodos (hoy son stubs)
```

### Tareas
1. Crear `models.go` con structs que mapean la respuesta JSON de ML
2. Crear `mapper.go` con conversion bidireccional
3. Implementar `GetShipment` - GET `/shipments/{id}` con header x-format-new
4. Implementar `GetShipmentByOrder` - GET via filtro por order_id
5. Implementar `ListShipments` - GET con filtros (query params)
6. Implementar `GetTracking` - GET `/shipments/{id}/history`
7. Implementar `GetLabel` - GET `/shipments/{id}/labels` (retorna bytes PDF)
8. Implementar `UpdateShipment` - PUT `/shipments/{id}` (limitado)
9. Implementar `CancelShipment` - POST o PUT cancel
10. Conectar en `sdk.go` - instanciar ShipmentAPI con provider real
11. Agregar sanitizacion en `ShipmentService` (similar a PaymentService)
12. Tests unitarios para mapper + service

### Estimacion: ~400 lineas de codigo nuevo

---

## Fase 3: Implementacion de QR / Instore

### Objetivo
Implementar el adapter de Mercado Pago QR para pagos presenciales con punto de venta.

### API Endpoints Confirmados (api.mercadopago.com)

**Ordenes QR (nueva API v1/orders):**

| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Crear orden | POST | `/v1/orders` |
| Obtener orden | GET | `/v1/orders/{id}` |
| Cancelar orden | POST | `/v1/orders/{id}/cancel` |
| Reembolsar orden | POST | `/v1/orders/{id}/refund` |

**POS (Puntos de Venta):**

| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Crear POS | POST | `/pos` |
| Buscar POS | GET | `/pos` |
| Obtener POS | GET | `/pos/{id}` |
| Actualizar POS | PUT | `/pos/{id}` |
| Eliminar POS | DELETE | `/pos/{id}` |

**Stores (Sucursales):**

| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Crear sucursal | POST | `/users/{user_id}/stores` |
| Obtener sucursal | GET | `/stores/{id}` |
| Buscar sucursales | GET | `/users/{user_id}/stores/search` |
| Actualizar sucursal | PUT | `/users/{user_id}/stores/{id}` |
| Eliminar sucursal | DELETE | `/users/{user_id}/stores/{id}` |

**Headers requeridos**: `X-Idempotency-Key` para operaciones POST

### Archivos a crear/modificar

```
providers/mercadolibre/qr/
  models.go         <- Structs de request/response ML
  mapper.go         <- ML models <-> domain models
  adapter.go        <- Implementar todos los metodos (hoy son stubs)
```

### Tareas
1. Crear `models.go` con structs para Orders, POS, Stores
2. Crear `mapper.go` con conversion bidireccional
3. Implementar `CreateQR` - POST `/v1/orders` (genera QR para la orden)
4. Implementar `GetQR` / `GetQRByExternalReference`
5. Implementar `DeleteQR` - cancelar orden
6. Implementar `GetQRPayment` - obtener pago asociado a orden QR
7. Implementar CRUD de POS (RegisterPOS, GetPOS, ListPOS, DeletePOS)
8. Implementar CRUD de Stores (RegisterStore, GetStore, ListStores)
9. Agregar `X-Idempotency-Key` header a operaciones POST
10. Conectar en `sdk.go` - instanciar QRAPI con provider real
11. Agregar sanitizacion en `QRService`
12. Tests unitarios para mapper + service

### Nota sobre user_id
Las operaciones de Stores requieren `user_id`. Este se obtiene del Access Token
via `GET /users/me`. El SDK deberia obtenerlo automaticamente al inicializarse
y almacenarlo en el Client.

### Estimacion: ~600 lineas de codigo nuevo

---

## Fase 4: Webhook Handler con HMAC

### Objetivo
Implementar validacion de firma y parsing de webhooks de Mercado Pago.

### Flujo de Validacion HMAC (confirmado de docs oficiales)

```
1. Extraer del header `x-signature`: ts=TIMESTAMP,v1=HASH
2. Obtener `x-request-id` del header
3. Obtener `data.id` de los query params (convertir a lowercase)
4. Construir template: "id:{data_id};request-id:{request_id};ts:{ts};"
5. Calcular HMAC-SHA256(template, secret_key) en hex
6. Comparar resultado con v1 del header
7. Opcionalmente validar que ts no sea muy antiguo
8. Responder HTTP 200/201 dentro de 22 segundos
```

### Archivos a crear/modificar

```
providers/mercadolibre/webhook/
  handler.go        <- Implementa ports.WebhookHandler
  models.go         <- Structs del payload de webhook
  hmac.go           <- Logica de validacion HMAC-SHA256
```

### Tareas
1. Crear `hmac.go` con validacion de firma HMAC-SHA256
2. Crear `models.go` con structs del payload de notificacion
3. Implementar `ValidateSignature` - extraer ts/v1, construir template, comparar
4. Implementar `ParsePaymentWebhook` - deserializar payload de pago
5. Implementar `ParseShipmentWebhook` - deserializar payload de envio
6. Implementar `ParseQRWebhook` - deserializar payload de QR
7. Exponer en SDK una forma de procesar webhooks entrantes
8. Tests unitarios con payloads de ejemplo

### Estimacion: ~300 lineas de codigo nuevo

---

## Fase 5: Tests de Integracion + Hardening

### Objetivo
Tests contra la API real (sandbox), manejo de edge cases, documentacion.

### Tareas
1. Tests de integracion para Payment (create, get, list, refund)
2. Tests de integracion para Shipment (get, tracking, label)
3. Tests de integracion para QR (create order, POS, stores)
4. Tests de integracion para Webhooks (payload de ejemplo)
5. Manejo de rate limiting (respetar `RateLimits` de capabilities)
6. Manejo de paginacion en listados
7. Revisar edge cases: tokens expirados, reintentos, timeouts
8. Documentacion de uso con ejemplos reales

### Requiere: API keys de test (ver seccion de Pre-requisitos)

---

## Orden de Ejecucion Recomendado

```
Fase 2 (Shipments)  ──┐
                       ├──> Fase 4 (Webhooks) ──> Fase 5 (Integration Tests)
Fase 3 (QR/Instore) ──┘
```

Fases 2 y 3 son independientes y se pueden trabajar en paralelo.
Fase 4 depende de tener al menos Payments funcional (ya lo tenemos).
Fase 5 requiere API keys y que las fases 2-4 esten completas.

---

## Metricas de Completitud

| Modulo | Fase 1 (actual) | Fase 2 | Fase 3 | Fase 4 | Fase 5 |
|--------|-----------------|--------|--------|--------|--------|
| Payment | 100% | 100% | 100% | 100% | 100% |
| Shipment | 5% (stubs) | 90% | 90% | 90% | 100% |
| QR | 5% (stubs) | 5% | 90% | 90% | 100% |
| Webhooks | 0% (interface) | 0% | 0% | 90% | 100% |
| Tests | 30% | 50% | 70% | 85% | 100% |

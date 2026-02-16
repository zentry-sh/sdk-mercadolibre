# Arquitectura del SDK - Contexto y Normas

## Investigacion Realizada

### Charm Logger (charmbracelet/log)
Estudiamos el paquete `log` de Charm para entender como una libreria Go maneja logging sin imponer dependencias al usuario.

**Hallazgos clave:**
- Expone una interfaz minima, no un logger completo
- El consumidor inyecta su implementacion o recibe un no-op por defecto
- Solo expone `Debug` para librerias (la libreria no decide que es importante para el usuario)
- Pattern `Func` adapter: una funcion puede satisfacer la interfaz sin structs

**Aplicacion en nuestro SDK:**
```go
// pkg/logger/logger.go
type Logger interface {
    Debug(msg string, keyvals ...any)
}
func Nop() Logger       // no-op por defecto
type Func func(...)     // adapter para funciones
```

### Wapikit Architecture
Estudiamos la arquitectura de wapikit.com para entender como estructuran un SDK multi-proveedor.

**Hallazgos clave:**
- Interfaces minimas por dominio (ports), no interfaces "god"
- Cada proveedor implementa los mismos ports
- Configuracion por capacidades (feature discovery), no por condicionales de pais
- Loader de configuracion con cache + embed para archivos YAML

**Aplicacion en nuestro SDK:**
- `core/ports/` define contratos: `PaymentProvider`, `ShipmentProvider`, `QRProvider`, `CapabilitiesProvider`
- `providers/mercadolibre/` implementa esos contratos
- `providers/mercadolibre/config/capabilities/*.yaml` define features por region
- Agregar un pais = agregar un YAML, no codigo

---

## Normas Arquitectonicas Vigentes

### 1. Libreria, no Aplicacion
El SDK es una libreria publica. Las implicaciones son:

| Aspecto | Norma |
|---------|-------|
| Logging | Solo `Debug`. El usuario controla su logging. No forzamos dependencia |
| Errores | Retornar `*errors.SDKError` estructurado. Nunca logear errores internamente |
| Defaults | `Nop()` logger, 30s timeout, pais PE. Todo overrideable |
| Panic | Prohibido. Todo error se retorna, nunca `panic` |

### 2. Seguridad de Strings (Post-Refactoring)
Ninguna concatenacion manual de strings para URLs, queries o headers.

| Contexto | Mecanismo |
|----------|-----------|
| URL paths | `fmt.Sprintf("/v1/payments/%s", url.PathEscape(id))` |
| Query params | `url.Values{}` + `.Encode()` |
| URL base + path | `url.JoinPath(baseURL, path)` |
| Headers | `fmt.Sprintf("Bearer %s", token)` |
| Nombres completos | `p.FirstName + " " + p.LastName` (unico caso seguro, datos propios) |

### 3. Sanitizacion de Inputs
Todo input del usuario se sanitiza en la capa de `usecases` antes de llegar al proveedor.

```
Usuario -> sdk.Payment.Create(req)
         -> PaymentService.CreatePayment(req)  // sanitize + validate aqui
           -> PaymentProvider.CreatePayment(req)  // datos ya limpios
```

Paquete `pkg/sanitize`:
- `String(s)` - trim + elimina null bytes
- `ID(s)` - solo alfanumerico, `-`, `_`
- `Email(s)` - lowercase + trim
- `CountryCode(s)` - 2 chars uppercase
- `CurrencyCode(s)` - 3 chars uppercase

### 4. Seguridad de Memoria
- `io.LimitReader(resp.Body, 10<<20)` en HTTP client (max 10 MiB por respuesta)
- `bytes.NewReader(body)` reutilizado en reintentos (no crea readers nuevos)
- `sync.RWMutex` en cache de capabilities (concurrency safe)
- `any` en lugar de `interface{}` (Go 1.18+)

### 5. Patron de Capabilities (Multi-Region)
El pais es configuracion, no codigo. Nunca `if country == "PE"`.

```
providers/mercadolibre/config/capabilities/
  pe.yaml   -> Peru: PEN, Yape, Olva, QR si
  mx.yaml   -> Mexico: MXN, SPEI, OXXO, Estafeta, QR si
  ar.yaml   -> Argentina: ARS, Rapipago, Correo Argentino
  br.yaml   -> Brasil: BRL, PIX
  cl.yaml   -> Chile: CLP
  co.yaml   -> Colombia: COP
```

Validacion dinamica:
```
amount > provider.MaxAmount() -> error
method not in capabilities   -> error
carrier not in region        -> error
```

### 6. Clean Architecture (Regla de Dependencia)
```
core/domain/     <- Entidades puras. No importa nada externo
core/ports/      <- Interfaces. Solo importa domain/
core/usecases/   <- Logica de negocio. Importa domain/ + ports/
core/errors/     <- Errores del SDK. No importa nada externo

providers/mercadolibre/  <- Implementa ports/. Importa core/ + pkg/
pkg/httputil/            <- HTTP client generico
pkg/logger/              <- Logger interface
pkg/sanitize/            <- Sanitizacion

sdk.go + config.go       <- API publica. Orquesta todo
```

**Prohibiciones:**
- `core/` NUNCA importa `providers/`
- `core/domain/` NUNCA importa `core/usecases/`
- `sdk.go` NUNCA expone tipos de `providers/`

---

## Estado Actual del Proyecto

### Completado (Funcional)
| Modulo | Estado | Detalle |
|--------|--------|---------|
| `core/domain/` | Completo | Payment, Shipment, QR, Capabilities, enums, common types |
| `core/ports/` | Completo | 5 interfaces: Payment, Shipment, QR, Capabilities, Webhook |
| `core/usecases/` | Completo | 4 services con sanitizacion y validacion |
| `core/errors/` | Completo | 21 codigos de error, constructores, helpers |
| `pkg/logger/` | Completo | Interface minimal + Nop + Func adapter |
| `pkg/sanitize/` | Completo | String, ID, Email, CountryCode, CurrencyCode |
| `pkg/httputil/` | Completo | Client con retry, backoff, LimitReader, error mapping |
| `providers/ml/auth` | Completo | OAuth2 code exchange, refresh, TokenManager |
| `providers/ml/payment` | Completo | Adapter + Mapper + Models. CRUD completo contra API real |
| `providers/ml/config` | Completo | Loader YAML con cache, 6 paises |
| `providers/ml/capabilities` | Completo | Validacion por capacidades |
| `sdk.go` | Parcial | Payment API funcional. Shipment/QR instanciados pero sin provider |
| Tests | Parcial | Unit tests para payment service + domain. Faltan shipment/QR |

### Pendiente (Stubs)
| Modulo | Estado | Que falta |
|--------|--------|-----------|
| `providers/ml/shipment/adapter.go` | Stub | Todos los metodos retornan `nil, nil` |
| `providers/ml/qr/adapter.go` | Stub | Todos los metodos retornan `nil, nil` |
| Shipment models + mapper | No existe | Falta crear `models.go` y `mapper.go` en shipment/ |
| QR models + mapper | No existe | Falta crear `models.go` y `mapper.go` en qr/ |
| Webhook handler | Solo interface | Falta implementacion HMAC + parsing |
| `sdk.go` wiring | Parcial | ShipmentAPI y QRAPI no conectados a providers |
| Tests shipment | No existe | Unit tests para ShipmentService |
| Tests QR | No existe | Unit tests para QRService |
| Tests capabilities | Parcial | Solo test basico |
| Integration tests | No existe | Tests contra API real (requiere API keys) |

---

## Endpoints de Mercado Libre Utilizados

### Payments (api.mercadopago.com)
| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Crear pago | POST | `/v1/payments` |
| Obtener pago | GET | `/v1/payments/{id}` |
| Buscar pagos | GET | `/v1/payments/search?...` |
| Cancelar pago | PUT | `/v1/payments/{id}` (status=cancelled) |
| Crear reembolso | POST | `/v1/payments/{id}/refunds` |
| Obtener reembolso | GET | `/v1/payments/{id}/refunds/{refund_id}` |
| Listar reembolsos | GET | `/v1/payments/{id}/refunds` |

### Shipments (api.mercadolibre.com) - Por implementar
| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Obtener envio | GET | `/shipments/{id}` |
| Obtener por orden | GET | `/orders/{order_id}/shipments` |
| Buscar envios | GET | `/shipments/search?...` |
| Actualizar envio | PUT | `/shipments/{id}` |
| Cancelar envio | POST | `/shipments/{id}/cancel` |
| Obtener tracking | GET | `/shipments/{id}/tracking` |
| Obtener etiqueta | GET | `/shipments/{id}/labels` |

### QR / Instore (api.mercadopago.com) - Por implementar
| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Crear QR | POST | `/instore/orders/qr/seller/collectors/{user_id}/pos/{pos_id}/qrs` |
| Obtener orden | GET | `/instore/orders/{order_id}` |
| Eliminar QR | DELETE | `/instore/qr/{qr_id}` |
| Registrar POS | POST | `/pos` |
| Obtener POS | GET | `/pos/{pos_id}` |
| Listar POS | GET | `/pos?store_id={store_id}` |
| Eliminar POS | DELETE | `/pos/{pos_id}` |
| Registrar sucursal | POST | `/stores` |
| Obtener sucursal | GET | `/stores/{store_id}` |
| Listar sucursales | GET | `/stores` |

### OAuth2
| Operacion | Metodo | Endpoint |
|-----------|--------|----------|
| Token por code | POST | `/oauth/token` (grant_type=authorization_code) |
| Refresh token | POST | `/oauth/token` (grant_type=refresh_token) |
| URL autorizacion | GET | `{base}/authorization?response_type=code&...` |

---

## Dependencias Externas

```go
// go.mod
require gopkg.in/yaml.v3 v3.0.1  // unica dependencia externa (para config YAML)
```

Sin frameworks HTTP, sin loggers externos, sin ORMs. Stdlib + 1 dependencia.

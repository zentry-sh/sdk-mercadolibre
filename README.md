# SDK Mercado Libre para Go

SDK en Go para integrar pagos, envíos, códigos QR y webhooks con las APIs de Mercado Libre / Mercado Pago. Soporte multi-región automático para 6 países de LATAM.

## Características

- **Pagos** — Crear, consultar, cancelar, reembolsar pagos con múltiples métodos por país
- **Envíos** — Consultar envíos, tracking en tiempo real, descarga de etiquetas PDF
- **QR / Instore** — Órdenes QR dinámico/estático, gestión de POS y sucursales
- **Webhooks** — Validación HMAC-SHA256, parsing de eventos, HTTP handler listo para montar
- **Multi-Región** — 6 países (PE, MX, AR, BR, CL, CO) con validación automática de capacidades
- **Seguridad** — Sanitización de inputs, `url.PathEscape` en paths, `io.LimitReader` en responses
- **Zero dependencies** — Solo `gopkg.in/yaml.v3` para configuración regional

## Instalación

```bash
go get github.com/zentry/sdk-mercadolibre
```

Requiere Go 1.21+.

## Inicio Rápido

```go
package main

import (
    "context"
    "log"

    sdk "github.com/zentry/sdk-mercadolibre"
    "github.com/zentry/sdk-mercadolibre/core/domain"
)

func main() {
    client, err := sdk.New(sdk.Config{
        AccessToken: "YOUR_ACCESS_TOKEN",
        Country:     "PE",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
        ExternalReference: "order-12345",
        Amount:            domain.Money{Amount: 100.00, Currency: "PEN"},
        Description:       "Compra de productos",
        Payer:             domain.Payer{Email: "customer@example.com"},
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Pago creado: %s (Estado: %s)", payment.ID, payment.Status.String())
}
```

## API Reference

### Pagos

```go
client.Payment.Create(ctx, req)                     // Crear pago
client.Payment.Get(ctx, id)                         // Obtener pago por ID
client.Payment.List(ctx, filters)                   // Buscar pagos con filtros
client.Payment.Cancel(ctx, id)                      // Cancelar pago
client.Payment.Refund(ctx, id, amount)              // Reembolso total o parcial
client.Payment.GetRefund(ctx, paymentID, refundID)  // Obtener reembolso
client.Payment.ListRefunds(ctx, paymentID)          // Listar reembolsos
```

### Envíos

```go
client.Shipment.Get(ctx, id)                  // Obtener envío
client.Shipment.GetByOrder(ctx, orderID)      // Envío por orden
client.Shipment.List(ctx, filters)            // Buscar envíos
client.Shipment.Update(ctx, id, req)          // Actualizar envío
client.Shipment.Cancel(ctx, id)               // Cancelar envío
client.Shipment.GetTracking(ctx, shipmentID)  // Historial de tracking
client.Shipment.GetLabel(ctx, shipmentID)     // Descargar etiqueta PDF ([]byte)
```

### QR / Instore

```go
client.QR.Create(ctx, req)                       // Crear orden QR
client.QR.Get(ctx, qrID)                         // Obtener orden
client.QR.GetByExternalReference(ctx, ref)        // Buscar por referencia externa
client.QR.Delete(ctx, qrID)                      // Cancelar orden QR
client.QR.GetPayment(ctx, qrID)                  // Obtener pago asociado

client.QR.RegisterPOS(ctx, req)                   // Registrar punto de venta
client.QR.GetPOS(ctx, posID)                      // Obtener POS
client.QR.ListPOS(ctx, storeID)                   // Listar POS por sucursal
client.QR.DeletePOS(ctx, posID)                   // Eliminar POS

client.QR.RegisterStore(ctx, req)                 // Registrar sucursal
client.QR.GetStore(ctx, storeID)                  // Obtener sucursal
client.QR.ListStores(ctx)                         // Listar sucursales
```

### Webhooks

El SDK incluye soporte completo para webhooks de Mercado Libre/Pago con validación HMAC-SHA256 integrada.

```go
client.Webhook.Process(ctx, req)      // Validar firma HMAC + parsear evento
client.Webhook.Validate(ctx, req)     // Solo validar firma
client.Webhook.Parse(ctx, payload)    // Solo parsear sin validar
client.Webhook.HTTPHandler(fn)        // Handler net/http listo para montar
```

#### Tipos de Eventos

Los webhooks pueden ser de tres tipos:

| Tipo | Descripción | Event ID |
|------|-------------|----------|
| `payment` | Eventos de pago (creado, actualizado, aprobado, rechazado) | `payment_id` |
| `shipment` | Eventos de envío (creado, actualizado, entregado) | `shipment_id` |
| `qr` | Eventos QR (orden creada, pagada, cancelada) | `qr_id` |

#### Ejemplo de Payload

```json
{
  "id": 1234567890,
  "type": "payment",
  "action": "payment.created",
  "date_created": "2024-01-15T10:30:00Z",
  "user_id": 123456789,
  "api_version": "v2"
}
```

#### Validación de Firma

El SDK valida automáticamente la firma HMAC-SHA256. Configura el secret en la inicialización:

```go
client, err := sdk.New(sdk.Config{
    AccessToken:   "YOUR_ACCESS_TOKEN",
    WebhookSecret: "YOUR_WEBHOOK_SECRET",  // Secret del portal de Mercado Pago
    Country:       "PE",
})
```

#### Ejemplo de HTTP Handler

```go
http.Handle("/webhooks", client.Webhook.HTTPHandler(
    func(ctx context.Context, event *domain.WebhookEvent) error {
        switch {
        case event.IsPaymentEvent():
            log.Printf("Pago %s: %s", event.DataID, event.Type)
        case event.IsShipmentEvent():
            log.Printf("Envío %s: %s", event.DataID, event.Type)
        case event.IsQREvent():
            log.Printf("QR %s: %s", event.DataID, event.Type)
        }
        return nil
    },
))
```

#### Reintentos

Mercado Libre reintenta entregas fallidas hasta 3 días con backoff exponencial. El handler retorna 2xx inmediatamente; el procesamiento pesado debe hacerse asynchronously.

#### Idempotencia

Usa el campo `id` del evento como clave única para evitar procesamiento duplicado:

```go
func (h *WebhookHandler) Handle(ctx context.Context, event *domain.WebhookEvent) error {
    // Verificar si ya fue procesado
    exists, err := h.cache.Exists(ctx, fmt.Sprintf("webhook:%d", event.ID))
    if err == nil && exists {
        return nil // Ya procesado
    }
    
    // Procesar evento...
    
    // Marcar como procesado (TTL: 7 días)
    h.cache.Set(ctx, fmt.Sprintf("webhook:%d", event.ID), "1", 7*24*time.Hour)
    return nil
}
```

#### Testing Local

Usa [ngrok](https://ngrok.com) o [Stripe CLI](https://docs.stripe.com/webhooks/test-local) para probar webhooks localmente:

```bash
# Con ngrok
ngrok http 8080

# Registra la URL en el portal de Mercado Pago
# https://your-ngrok.io/webhooks
```

### Capacidades por Región

```go
client.Capabilities.Get(ctx)                  // Capacidades del país configurado
client.Capabilities.GetForCountry(ctx, "MX")  // Capacidades de otro país
client.Capabilities.GetPaymentMethods(ctx)    // Métodos de pago disponibles
client.Capabilities.GetCarriers(ctx)          // Carriers de logística
client.Capabilities.IsQRSupported(ctx)        // Soporte QR en el país
client.Capabilities.GetCurrency(ctx)          // Moneda del país
```

## Multi-Región

El SDK valida automáticamente cada operación contra las capacidades del país configurado.

| País | Código | Moneda | QR | Métodos de pago |
|------|--------|--------|-----|-----------------|
| Perú | PE | PEN | Sí | Yape, tarjetas, transferencia |
| México | MX | MXN | Sí | SPEI, OXXO, tarjetas |
| Argentina | AR | ARS | Sí | Rapipago, tarjetas |
| Brasil | BR | BRL | Sí | PIX, boleto, tarjetas |
| Chile | CL | CLP | Sí | tarjetas, transferencia |
| Colombia | CO | COP | Sí | PSE, Efecty, tarjetas |

Cambiar de país en runtime:

```go
peClient, _ := sdk.New(sdk.Config{Country: "PE", AccessToken: "..."})
mxClient, _ := peClient.ForCountry("MX")
```

## Arquitectura

El SDK sigue Clean Architecture con regla de dependencia estricta: `core/` nunca importa `providers/`.

```
[sdk.go](sdk.go)              API pública, orquesta todo
[config.go](config.go)           Configuración del SDK

core/
  domain/           Entidades puras (Payment, Shipment, QR, Webhook)
  ports/            Interfaces: PaymentProvider, ShipmentProvider, QRProvider, WebhookHandler
  usecases/         Servicios con sanitización y validación
  errors/           Sistema de errores unificado (21 códigos)

providers/
  mercadolibre/
    payment/        Adapter + Mapper + Models
    shipment/       Adapter + Mapper + Models
    qr/             Adapter + Mapper + Models
    webhook/        Handler HMAC-SHA256 + Parser
    config/         Capabilities por país (YAML embebido)
    [auth.go](providers/mercadolibre/auth.go)         OAuth2 (code exchange, refresh)
    [client.go](providers/mercadolibre/client.go)       HTTP clients por servicio
    [endpoints.go](providers/mercadolibre/endpoints.go)    URLs por región

pkg/
  httputil/         HTTP client con retry, backoff, LimitReader, RequestOption
  logger/           Interface minimal (Debug only) + Nop + Func adapter
  sanitize/         String, ID, Email, CountryCode, CurrencyCode
  idempotency/      UUID v4 para X-Idempotency-Key
```

### Principios de Diseño

- **Librería, no aplicación** — Sin panics, sin logging forzado, errores estructurados
- **Seguridad de strings** — `fmt.Sprintf` + `url.PathEscape` para URLs, `url.Values` para queries
- **Sanitización** — Todo input se sanitiza en la capa de usecases antes de llegar al proveedor
- **Memoria** — `io.LimitReader` (10 MiB max), `bytes.NewReader` reutilizado en reintentos
- **Concurrencia** — `sync.RWMutex` en cache de capabilities
- **Extensibilidad** — Agregar un país = agregar un YAML, no código

### Seguridad

| Capa | Mecanismo |
|------|-----------|
| URLs | `url.PathEscape(id)` dentro de `fmt.Sprintf` |
| Query params | `url.Values{}` + `.Encode()` |
| Headers | `fmt.Sprintf("Bearer %s", token)` |
| Inputs | Sanitización en usecases (trim, null bytes, regex) |
| Webhooks | HMAC-SHA256 con comparación timing-safe |
| HTTP responses | `io.LimitReader(resp.Body, 10<<20)` |
| Idempotencia | UUID v4 via `crypto/rand` para `X-Idempotency-Key` |

## Configuración

```go
client, err := sdk.New(sdk.Config{
    AccessToken:   "YOUR_ACCESS_TOKEN",     // Token de Mercado Pago
    ClientID:      "YOUR_CLIENT_ID",        // OAuth2 (opcional)
    ClientSecret:  "YOUR_CLIENT_SECRET",    // OAuth2 (opcional)
    Country:       "PE",                    // Default: PE
    Timeout:       30 * time.Second,        // Default: 30s
    WebhookSecret: "YOUR_WEBHOOK_SECRET",   // Para validación HMAC
    Logger:        logger.Func(func(msg string, kv ...any) {
        slog.Debug(msg, kv...)
    }),
})
```

### Logger Personalizado

El SDK usa una interfaz minimal de logging compatible con cualquier logger:

```go
type Logger interface {
    Debug(msg string, keyvals ...any)
}
```

Si no se proporciona, se usa un no-op logger. Para integrar con `slog`, `zap`, o `logrus`, solo implementa la interfaz o usa el adapter `logger.Func`.

## Manejo de Errores

Todos los errores retornan `*errors.SDKError` con código estructurado:

```go
payment, err := client.Payment.Get(ctx, "invalid-id")
if err != nil {
    var sdkErr *errors.SDKError
    if stderrors.As(err, &sdkErr) {
        switch sdkErr.Code {
        case errors.ErrCodeNotFound:
            log.Println("Pago no encontrado")
        case errors.ErrCodeUnauthorized:
            log.Println("Token inválido")
        case errors.ErrCodeRateLimited:
            log.Println("Rate limit, reintentar")
        }
    }
}

if errors.IsNotFound(err) {
    log.Println("No existe")
}
```

## Testing

El SDK incluye ~60 tests unitarios con cobertura para servicios, mappers y HMAC:

```bash
go test ./... -v
```

Los tests usan mocks (sin API keys) y son ejecutables offline.

## Dependencias

```
gopkg.in/yaml.v3    Única dependencia externa (configuración regional YAML)
```

Sin frameworks HTTP, sin loggers externos, sin ORMs. Standard library + 1 dependencia.

## Licencia

MIT License

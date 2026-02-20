# SDK Mercado Libre para Go

SDK en Go para integrar pagos, envios, codigos QR y webhooks con las APIs de Mercado Libre / Mercado Pago. Soporte multi-region automatico para 6 paises de LATAM.

## Caracteristicas

- **Pagos** — Crear, consultar, cancelar, reembolsar pagos con multiples metodos por pais
- **Envios** — Consultar envios, tracking en tiempo real, descarga de etiquetas PDF
- **QR / Instore** — Ordenes QR dinamico/estatico, gestion de POS y sucursales
- **Webhooks** — Validacion HMAC-SHA256, parsing de eventos, HTTP handler listo para montar
- **Multi-Region** — 6 paises (PE, MX, AR, BR, CL, CO) con validacion automatica de capacidades
- **Seguridad** — Sanitizacion de inputs, `url.PathEscape` en paths, `io.LimitReader` en responses
- **Zero dependencies** — Solo `gopkg.in/yaml.v3` para configuracion regional

## Instalacion

```bash
go get github.com/zentry/sdk-mercadolibre
```

Requiere Go 1.21+.

## Inicio Rapido

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

### Envios

```go
client.Shipment.Get(ctx, id)                  // Obtener envio
client.Shipment.GetByOrder(ctx, orderID)      // Envio por orden
client.Shipment.List(ctx, filters)            // Buscar envios
client.Shipment.Update(ctx, id, req)          // Actualizar envio
client.Shipment.Cancel(ctx, id)               // Cancelar envio
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

```go
client.Webhook.Process(ctx, req)      // Validar firma HMAC + parsear evento
client.Webhook.Validate(ctx, req)     // Solo validar firma
client.Webhook.Parse(ctx, payload)    // Solo parsear sin validar
client.Webhook.HTTPHandler(fn)        // Handler net/http listo para montar
```

Ejemplo de webhook HTTP handler:

```go
http.Handle("/webhooks", client.Webhook.HTTPHandler(
    func(ctx context.Context, event *domain.WebhookEvent) error {
        switch {
        case event.IsPaymentEvent():
            log.Printf("Pago %s: %s", event.DataID, event.Type)
        case event.IsShipmentEvent():
            log.Printf("Envio %s: %s", event.DataID, event.Type)
        case event.IsQREvent():
            log.Printf("QR %s: %s", event.DataID, event.Type)
        }
        return nil
    },
))
```

### Capacidades por Region

```go
client.Capabilities.Get(ctx)                  // Capacidades del pais configurado
client.Capabilities.GetForCountry(ctx, "MX")  // Capacidades de otro pais
client.Capabilities.GetPaymentMethods(ctx)    // Metodos de pago disponibles
client.Capabilities.GetCarriers(ctx)          // Carriers de logistica
client.Capabilities.IsQRSupported(ctx)        // Soporte QR en el pais
client.Capabilities.GetCurrency(ctx)          // Moneda del pais
```

## Multi-Region

El SDK valida automaticamente cada operacion contra las capacidades del pais configurado.

| Pais | Codigo | Moneda | QR | Metodos de pago |
|------|--------|--------|-----|-----------------|
| Peru | PE | PEN | Si | Yape, tarjetas, transferencia |
| Mexico | MX | MXN | Si | SPEI, OXXO, tarjetas |
| Argentina | AR | ARS | Si | Rapipago, tarjetas |
| Brasil | BR | BRL | Si | PIX, boleto, tarjetas |
| Chile | CL | CLP | Si | tarjetas, transferencia |
| Colombia | CO | COP | Si | PSE, Efecty, tarjetas |

Cambiar de pais en runtime:

```go
peClient, _ := sdk.New(sdk.Config{Country: "PE", AccessToken: "..."})
mxClient, _ := peClient.ForCountry("MX")
```

## Arquitectura

El SDK sigue Clean Architecture con regla de dependencia estricta: `core/` nunca importa `providers/`.

```
sdk.go              API publica, orquesta todo
config.go           Configuracion del SDK

core/
  domain/           Entidades puras (Payment, Shipment, QR, Webhook)
  ports/            Interfaces: PaymentProvider, ShipmentProvider, QRProvider, WebhookHandler
  usecases/         Servicios con sanitizacion y validacion
  errors/           Sistema de errores unificado (21 codigos)

providers/
  mercadolibre/
    payment/        Adapter + Mapper + Models
    shipment/       Adapter + Mapper + Models
    qr/             Adapter + Mapper + Models
    webhook/        Handler HMAC-SHA256 + Parser
    config/         Capabilities por pais (YAML embebido)
    auth.go         OAuth2 (code exchange, refresh)
    client.go       HTTP clients por servicio
    endpoints.go    URLs por region

pkg/
  httputil/         HTTP client con retry, backoff, LimitReader, RequestOption
  logger/           Interface minimal (Debug only) + Nop + Func adapter
  sanitize/         String, ID, Email, CountryCode, CurrencyCode
  idempotency/      UUID v4 para X-Idempotency-Key
```

### Principios de Diseno

- **Libreria, no aplicacion** — Sin panics, sin logging forzado, errores estructurados
- **Seguridad de strings** — `fmt.Sprintf` + `url.PathEscape` para URLs, `url.Values` para queries
- **Sanitizacion** — Todo input se sanitiza en la capa de usecases antes de llegar al proveedor
- **Memoria** — `io.LimitReader` (10 MiB max), `bytes.NewReader` reutilizado en reintentos
- **Concurrencia** — `sync.RWMutex` en cache de capabilities
- **Extensibilidad** — Agregar un pais = agregar un YAML, no codigo

### Seguridad

| Capa | Mecanismo |
|------|-----------|
| URLs | `url.PathEscape(id)` dentro de `fmt.Sprintf` |
| Query params | `url.Values{}` + `.Encode()` |
| Headers | `fmt.Sprintf("Bearer %s", token)` |
| Inputs | Sanitizacion en usecases (trim, null bytes, regex) |
| Webhooks | HMAC-SHA256 con comparacion timing-safe |
| HTTP responses | `io.LimitReader(resp.Body, 10<<20)` |
| Idempotencia | UUID v4 via `crypto/rand` para `X-Idempotency-Key` |

## Configuracion

```go
client, err := sdk.New(sdk.Config{
    AccessToken:   "YOUR_ACCESS_TOKEN",     // Token de Mercado Pago
    ClientID:      "YOUR_CLIENT_ID",        // OAuth2 (opcional)
    ClientSecret:  "YOUR_CLIENT_SECRET",    // OAuth2 (opcional)
    Country:       "PE",                    // Default: PE
    Timeout:       30 * time.Second,        // Default: 30s
    WebhookSecret: "YOUR_WEBHOOK_SECRET",   // Para validacion HMAC
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

Todos los errores retornan `*errors.SDKError` con codigo estructurado:

```go
payment, err := client.Payment.Get(ctx, "invalid-id")
if err != nil {
    var sdkErr *errors.SDKError
    if stderrors.As(err, &sdkErr) {
        switch sdkErr.Code {
        case errors.ErrCodeNotFound:
            log.Println("Pago no encontrado")
        case errors.ErrCodeUnauthorized:
            log.Println("Token invalido")
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
gopkg.in/yaml.v3    Unica dependencia externa (configuracion regional YAML)
```

Sin frameworks HTTP, sin loggers externos, sin ORMs. Standard library + 1 dependencia.

## Licencia

MIT License

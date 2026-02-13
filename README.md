# SDK Mercado Libre para Go

SDK en Go para integración con las APIs de Mercado Libre, proporcionando acceso a pagos, envíos y códigos QR con soporte multi-región automático.

## Descripción

Este SDK abstrae las complejidades de la API de Mercado Libre mediante Clean Architecture, permitiendo trabajar con una interfaz consistente independientemente del país o proveedor. El código está estructurado en capas independientes para facilitar mantenimiento, testing y evolución hacia múltiples proveedores de pago y logística.

### Características Principales

**Pagos**
- Crear y consultar pagos
- Soporte para múltiples métodos de pago por país
- Reembolsos totales y parciales
- Cancelación de transacciones
- Búsqueda y filtrado de pagos

**Envíos**
- Crear y consultar envíos
- Seguimiento en tiempo real
- Descarga de etiquetas de envío
- Gestión de cancelaciones
- Soporte para múltiples carriers

**Códigos QR**
- Generación de QR dinámico y estático
- Registro y gestión de puntos de venta
- Webhooks para notificaciones
- Conciliación de transacciones

**Multi-Región**
- Soporte para 6 países: Perú, México, Argentina, Brasil, Chile, Colombia
- Validación automática según capacidades del país
- Métodos de pago específicos por región
- Monedas y limites configurables por país

## Instalación

```bash
go get github.com/zentry/sdk-mercadolibre
```

## Uso Básico

```go
package main

import (
    "context"
    "log"

    sdk "github.com/zentry/sdk-mercadolibre"
    "github.com/zentry/sdk-mercadolibre/core/domain"
)

func main() {
    ctx := context.Background()

    // Inicializar cliente para Perú
    client, err := sdk.New(sdk.Config{
        AccessToken: "YOUR_ACCESS_TOKEN",
        Country:     "PE",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Crear un pago
    payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
        ExternalReference: "order-12345",
        Amount: domain.Money{
            Amount:   100.00,
            Currency: "PEN",
        },
        Description: "Compra de productos",
        Payer: domain.Payer{
            Email: "customer@example.com",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Pago creado: %s (Estado: %s)", payment.ID, payment.Status.String())
}
```

## Arquitectura

El SDK sigue Clean Architecture con separación clara de capas:

- **core/domain**: Entidades y lógica de negocio independiente del proveedor
- **core/ports**: Interfaces que definen contratos con proveedores
- **core/usecases**: Servicios de aplicación que orquestan la lógica
- **providers**: Adaptadores específicos de cada proveedor (actualmente Mercado Libre)
- **pkg**: Utilidades compartidas (HTTP client, logging)

```
core/
├── domain/           Entidades (Payment, Shipment, QR, Capabilities)
├── errors/           Sistema de errores unificado
├── ports/            Interfaces para proveedores
└── usecases/         Servicios de aplicación

providers/
└── mercadolibre/     Implementación Mercado Libre
    ├── payment/      Adaptador de pagos
    ├── shipment/     Adaptador de envíos
    ├── qr/           Adaptador de QR
    └── config/       Configuración por país

pkg/
├── httputil/         Cliente HTTP con reintentos
└── logger/           Interface de logging
```

## Capacidades por País

Cada país tiene configuración específica que define:
- Métodos de pago soportados
- Límites de monto mínimo y máximo
- Moneda de transacción
- Carriers de logística disponibles
- Soporte para QR dinámico/estático

Las capacidades se validan automáticamente en cada operación, asegurando que las transacciones sean válidas para el país configurado.

## Cambio de País en Tiempo de Ejecución

```go
peClient, _ := sdk.New(sdk.Config{Country: "PE"})
mxClient, _ := peClient.ForCountry("MX")

// Ahora mxClient opera en México con validaciones mexicanas
```

## Testing

El SDK incluye más de 30 tests unitarios. Ejecutar:

```bash
go test ./...
```

## Ejemplos

Consultar el directorio `examples/` para ver casos de uso completos:
- `examples/payment/`: Ejemplo básico de creación de pagos
- `examples/multi_region/`: Ejemplo de operaciones multi-país

## Licencia

MIT License - Ver archivo [LICENSE](LICENSE)

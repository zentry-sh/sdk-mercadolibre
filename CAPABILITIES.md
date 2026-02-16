# SDK Mercado Libre - Multi-Region Capabilities

Este SDK implementa un sistema de **Capacidades Multi-Región** que permite trabajar con Mercado Libre en múltiples países sin cambios de código.

## 🌍 Países Soportados

- **Perú (PE)** - PEN
- **México (MX)** - MXN
- **Argentina (AR)** - ARS
- **Brasil (BR)** - BRL
- **Chile (CL)** - CLP
- **Colombia (CO)** - COP

## 🏗️ Arquitectura de Capacidades

### Principio Fundamental

> El país es **configuración**, no **lógica**

Cada país tiene:
- Métodos de pago específicos
- Límites de montos
- Carriers logísticos
- Reglas de validación
- Limits de rate limiting

Todo esto se define en YAML sin cambiar el código.

## 📋 Estructura de Configuración

### Ejemplo: Peru (pe.yaml)

```yaml
region:
  country_code: "PE"
  currency_code: "PEN"
  locale: "es-PE"
  timezone_iana: "America/Lima"

payment:
  supported_methods:
    - id: "yape"
      type: "transfer"
      name: "Yape"
      min_amount: 1.0
      max_amount: 2000.0
    - id: "plin"
      type: "transfer"
  min_amount: 1.0
  max_amount: 50000.0
  supports_refunds: true
  supports_installments: true
  max_installments: 12
  supported_currencies: ["PEN"]
  requires_kyc: true

shipment:
  supported_carriers:
    - id: "olva"
      name: "Olva Courier"
      service_types: ["standard", "express"]

qr:
  supported: true
  supports_dynamic_qr: true
  max_amount: 2000.0

rate_limits:
  requests_per_second: 10
  requests_per_minute: 300
```

## 💻 Uso del SDK

### 1. Crear Cliente para un País

```go
import sdk "github.com/zentry/sdk-mercadolibre"

client, err := sdk.New(sdk.Config{
    AccessToken: "YOUR_ACCESS_TOKEN",
    Country:     "PE",
})
```

### 2. Acceder a Capacidades

```go
caps, err := client.Capabilities.Get(ctx)

// Métodos de pago disponibles
for _, method := range caps.Payment.SupportedMethods {
    fmt.Printf("%s: %.2f - %.2f\n", 
        method.Name, 
        method.MinAmount.Amount, 
        method.MaxAmount.Amount)
}

// Moneda del país
currency := caps.Region.CurrencyCode // "PEN"

// Máximo de cuotas
maxInstallments := caps.Payment.MaxInstallments // 12
```

### 3. Validación Automática en Pagos

El SDK valida automáticamente cada pago contra las capacidades del país:

```go
payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
    Amount: domain.Money{
        Amount:   100.00,
        Currency: "PEN",  // Validado: ¿Acepta PEN?
    },
    MethodID: "yape",  // Validado: ¿Soporta Yape?
    Payer: domain.Payer{
        Email: "customer@example.com",
    },
})

// Si el monto es > 2000 PEN para Yape, rechaza automáticamente
// Si la moneda no es PEN, rechaza automáticamente
```

### 4. Cambiar de País en Tiempo de Ejecución

```go
peClient, _ := sdk.New(sdk.Config{Country: "PE"})
mxClient, _ := peClient.ForCountry("MX")

// Ahora mxClient está configurado para México
```

### 5. Listar Todas las Regiones

```go
regions, err := client.Capabilities.ListRegions(ctx)

for _, region := range regions {
    fmt.Printf("%s (%s) - %s\n",
        region.CountryCode,
        region.Locale,
        region.CurrencyCode)
}
```

## 🔍 Ejemplos de Uso

### Ejemplo 1: Obtener Métodos de Pago por País

```go
// Para Perú
peCaps, _ := client.Capabilities.Get(ctx)
fmt.Println(peCaps.Payment.SupportedMethods)
// Output: [yape, plin, pagoefectivo, credit_card, debit_card]

// Para México
mxClient, _ := client.ForCountry("MX")
mxCaps, _ := mxClient.Capabilities.Get(ctx)
fmt.Println(mxCaps.Payment.SupportedMethods)
// Output: [spei, oxxo, paycash, credit_card, debit_card, mercado_credito]
```

### Ejemplo 2: Crear Pago con Validación

```go
// Esta solicitud será rechazada
payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
    Amount: domain.Money{
        Amount:   100000.00,  // Excede máximo de 50000
        Currency: "PEN",
    },
})

// Error: "amount 100000.00 exceeds maximum 50000.00 for PE"
```

### Ejemplo 3: Comparar Capacidades entre Países

```go
countries := []string{"PE", "MX", "AR", "BR", "CL", "CO"}

for _, country := range countries {
    c, _ := client.Capabilities.GetForCountry(ctx, country)
    
    fmt.Printf("%s:\n", country)
    fmt.Printf("  Currency: %s\n", c.Region.CurrencyCode)
    fmt.Printf("  Max Payment: %.2f\n", c.Payment.MaxAmount.Amount)
    fmt.Printf("  Max Installments: %d\n", c.Payment.MaxInstallments)
    fmt.Printf("  QR Supported: %v\n", c.QR.Supported)
}
```

## 🎯 Validaciones Automáticas

El SDK valida automáticamente:

1. **Monto**: ¿Está dentro del rango min/max del país?
2. **Moneda**: ¿Usa la moneda correcta?
3. **Método de Pago**: ¿Está disponible en este país?
4. **Cuotas**: ¿Excede el máximo de instalaciones?
5. **Envíos**: ¿El carrier está disponible?
6. **QR**: ¿QR es soportado en este país?

## 🔄 Agregar un Nuevo País

Para agregar soporte para un nuevo país:

1. Crear archivo `providers/mercadolibre/config/capabilities/XX.yaml` (donde XX es el código del país)
2. Definir métodos de pago, carriers, límites
3. ¡Listo! El SDK lo detecta automáticamente

No requiere cambios de código en la aplicación.

## 📊 Estructura de Archivos

```
providers/mercadolibre/config/
├── capabilities/
│   ├── pe.yaml
│   ├── mx.yaml
│   ├── ar.yaml
│   ├── br.yaml
│   ├── cl.yaml
│   └── co.yaml
└── loader.go      # Carga los YAML automáticamente
```

## 🔐 Seguridad

- Las capacidades se cachean en memoria
- No hay I/O en cada validación (solo lookup en cache)
- Las configuraciones se validan al cargar
- Los errores se normalizan

## 📈 Performance

- Carga LAZY: Las capacidades se cargan solo cuando se necesitan
- Caché en memoria: O(1) lookups después del primer acceso
- Embed FS: Los YAML se incluyen en el binario (sin I/O en runtime)

## 🧪 Testing

```bash
go test ./tests/unit/core/capabilities_test.go -v
```

Pruebas incluidas:
- Carga de capabilidades por país
- Validación de pagos
- Validación de QR
- Validación de envíos
- Listado de regiones soportadas

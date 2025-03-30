# Integration Guides

This section provides detailed guides for integrating the Asian Cryptocurrency Payment System with different platforms.

## Overview

The Asian Cryptocurrency Payment System offers SDKs for three main platforms:

1. [Web Integration](#web-integration) - For e-commerce websites and web applications
2. [POS Integration](#pos-integration) - For point-of-sale terminals and retail systems
3. [Kiosk Integration](#kiosk-integration) - For self-service kiosks and terminals

Each SDK provides a simple interface to interact with our payment system, handling the complexities of cryptocurrency transactions, exchange rates, and payment platform integrations.

## Web Integration

Our Web SDK allows you to easily integrate cryptocurrency payments into your website or web application. It supports all major browsers and frameworks.

### Installation

#### Using NPM

```bash
npm install asian-crypto-payment-web-sdk
```

#### Using CDN

```html
<script src="https://cdn.asiancryptopayment.com/sdk/web/v1/asian-crypto-payment.min.js"></script>
```

### Basic Usage

```javascript
// Initialize the SDK
const paymentSDK = new AsianCryptoPayment.initialize({
  apiKey: 'your_api_key',
  environment: 'production', // or 'sandbox' for testing
});

// Create a payment
const payment = await paymentSDK.createPayment({
  amount: 100.50,
  currency: 'MYR',
  paymentMethod: 'ewallet',
  paymentPlatform: 'malaysia_grabpay',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+60123456789',
  metadata: {
    productId: 'PROD123',
    customField: 'custom_value'
  }
});

// Redirect to payment URL
window.location.href = payment.paymentUrl;
```

### Payment Button

Our SDK also provides a pre-built payment button that you can easily add to your website:

```html
<div id="payment-button-container"></div>

<script>
  paymentSDK.renderPaymentButton('#payment-button-container', {
    amount: 100.50,
    currency: 'MYR',
    orderId: 'ORDER123456',
    description: 'Payment for Order #123456',
    onSuccess: function(payment) {
      console.log('Payment successful!', payment);
      // Redirect to success page
      window.location.href = '/success?order_id=' + payment.orderId;
    },
    onFailure: function(error) {
      console.error('Payment failed:', error);
      // Handle error
    }
  });
</script>
```

### Handling Callbacks

When a payment is completed, the user will be redirected to your redirect URL. You can also set up a webhook to receive payment notifications:

```javascript
// Set up webhook handler (server-side code)
app.post('/webhook', (req, res) => {
  const signature = req.headers['x-signature'];
  const payload = req.body;
  
  // Verify signature
  if (paymentSDK.verifyWebhookSignature(payload, signature)) {
    // Process webhook event
    if (payload.type === 'payment.completed') {
      // Update order status
      updateOrderStatus(payload.data.orderId, 'paid');
    }
    
    res.status(200).send('Webhook received');
  } else {
    res.status(400).send('Invalid signature');
  }
});
```

### Advanced Configuration

The SDK supports various configuration options:

```javascript
const paymentSDK = new AsianCryptoPayment.initialize({
  apiKey: 'your_api_key',
  environment: 'production',
  defaultCurrency: 'SGD',
  defaultPaymentPlatform: 'singapore_paynow',
  supportedPaymentMethods: ['ewallet', 'qrcode', 'bank_transfer'],
  supportedCryptocurrencies: ['BTC', 'ETH'],
  language: 'en',
  theme: {
    primaryColor: '#3498db',
    secondaryColor: '#2ecc71',
    fontFamily: 'Arial, sans-serif'
  },
  onPaymentCreated: function(payment) {
    console.log('Payment created:', payment);
  },
  onPaymentCompleted: function(payment) {
    console.log('Payment completed:', payment);
  },
  onPaymentFailed: function(payment, error) {
    console.error('Payment failed:', error);
  }
});
```

For more details, see the [Web SDK Reference](web.md).

## POS Integration

Our POS SDK allows you to integrate cryptocurrency payments into your point-of-sale system. It supports Java-based POS systems.

### Installation

#### Using Gradle

```gradle
dependencies {
    implementation 'com.asiancryptopayment:pos-sdk:1.0.0'
}
```

#### Using Maven

```xml
<dependency>
    <groupId>com.asiancryptopayment</groupId>
    <artifactId>pos-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Basic Usage

```java
// Initialize the SDK
AsianCryptoPayment paymentSDK = new AsianCryptoPayment.Builder()
    .setApiKey("your_api_key")
    .setEnvironment(Environment.PRODUCTION)
    .build();

// Create a payment
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100.50)
    .setCurrency("SGD")
    .setPaymentMethod(PaymentMethod.QR_CODE)
    .setPaymentPlatform("singapore_nets")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+6591234567")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
}

// Check payment status
PaymentStatus status = paymentSDK.getPaymentStatus(payment.getPaymentId());
if (status.getStatus() == Status.COMPLETED) {
    // Payment successful
    printReceipt(payment);
}
```

### Handling Payment Events

```java
// Set up payment listener
paymentSDK.setPaymentListener(new PaymentListener() {
    @Override
    public void onPaymentStatusChanged(Payment payment) {
        if (payment.getStatus() == Status.COMPLETED) {
            // Payment successful
            displaySuccessScreen();
            printReceipt(payment);
        } else if (payment.getStatus() == Status.FAILED) {
            // Payment failed
            displayErrorScreen("Payment failed: " + payment.getErrorMessage());
        }
    }
});
```

For more details, see the [POS SDK Reference](pos.md).

## Kiosk Integration

Our Kiosk SDK allows you to integrate cryptocurrency payments into self-service kiosks and terminals. It supports C/C++ based systems.

### Installation

#### Using CMake

```cmake
# Add the Asian Crypto Payment repository
add_subdirectory(asian_crypto_payment)

# Link against the library
target_link_libraries(your_app asian_crypto_payment)
```

### Basic Usage

```cpp
#include "asian_crypto_payment.h"

// Initialize the SDK
AsianCryptoPayment* payment_sdk = asian_crypto_payment_init("your_api_key", ENVIRONMENT_PRODUCTION);

// Create a payment
PaymentRequest request = {
    .amount = 100.50,
    .currency = "THB",
    .payment_method = PAYMENT_METHOD_QR_CODE,
    .payment_platform = "thailand_promptpay",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+66812345678"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
}

// Check payment status
while (1) {
    PaymentStatus* status = asian_crypto_payment_get_payment_status(payment_sdk, payment->payment_id);
    
    if (status->status == PAYMENT_STATUS_COMPLETED) {
        // Payment successful
        display_success_screen();
        print_receipt(payment);
        break;
    } else if (status->status == PAYMENT_STATUS_FAILED) {
        // Payment failed
        display_error_screen("Payment failed");
        break;
    }
    
    // Wait for 2 seconds before checking again
    sleep(2);
    
    // Free status
    asian_crypto_payment_free_payment_status(status);
}

// Clean up
asian_crypto_payment_free_payment(payment);
asian_crypto_payment_cleanup(payment_sdk);
```

For more details, see the [Kiosk SDK Reference](kiosk.md).

## Country-Specific Integration

For country-specific integration details, please refer to the following guides:

- [Malaysia](countries/malaysia.md)
- [Singapore](countries/singapore.md)
- [Indonesia](countries/indonesia.md)
- [Thailand](countries/thailand.md)
- [Vietnam](countries/vietnam.md)
- [Cambodia](countries/cambodia.md)
- [Laos](countries/laos.md)

## Testing

We provide a sandbox environment for testing your integration. To use the sandbox:

1. Set the environment to 'sandbox' when initializing the SDK
2. Use the test API keys provided in your developer dashboard
3. Use the test credit card numbers and account details provided in the [Testing Guide](testing.md)

## Support

If you encounter any issues with integration, please refer to the [Troubleshooting Guide](../installation/troubleshooting.md) or contact our support team.

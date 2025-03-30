# Indonesia Payment Integration

This guide provides specific details for integrating with Indonesian payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Indonesia:

1. **GoPay** - E-wallet payment system by Gojek
2. **OVO** - E-wallet payment system

## Currency

The official currency for Indonesia is the Indonesian Rupiah (IDR).

## GoPay Integration

### Configuration

To integrate with GoPay, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  indonesia_gopay:
    client_id: "your_client_id"
    client_secret: "your_client_secret"
    merchant_id: "your_merchant_id"
    api_endpoint: "https://api.midtrans.com"
    callback_url: "https://your-domain.com/callback/indonesia/gopay"
    redirect_url: "https://your-domain.com/redirect/indonesia/gopay"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using GoPay
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // IDR 100,000
  currency: 'IDR',
  paymentMethod: 'ewallet',
  paymentPlatform: 'indonesia_gopay',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+6281234567890'
});

// Redirect to GoPay payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using GoPay
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // IDR 100,000
    .setCurrency("IDR")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("indonesia_gopay")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+6281234567890")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using GoPay
PaymentRequest request = {
    .amount = 100000.00,  // IDR 100,000
    .currency = "IDR",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "indonesia_gopay",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+6281234567890"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
}
```

## OVO Integration

### Configuration

To integrate with OVO, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  indonesia_ovo:
    app_id: "your_app_id"
    app_key: "your_app_key"
    merchant_id: "your_merchant_id"
    api_endpoint: "https://api.ovo.id"
    callback_url: "https://your-domain.com/callback/indonesia/ovo"
    redirect_url: "https://your-domain.com/redirect/indonesia/ovo"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using OVO
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // IDR 100,000
  currency: 'IDR',
  paymentMethod: 'ewallet',
  paymentPlatform: 'indonesia_ovo',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+6281234567890'
});

// Redirect to OVO payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using OVO
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // IDR 100,000
    .setCurrency("IDR")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("indonesia_ovo")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+6281234567890")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using OVO
PaymentRequest request = {
    .amount = 100000.00,  // IDR 100,000
    .currency = "IDR",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "indonesia_ovo",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+6281234567890"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Indonesia, you must comply with the following regulations:

1. **Bank Indonesia (BI)** - The central bank of Indonesia regulates payment systems.
2. **Financial Services Authority (OJK)** - Regulates and supervises financial services sector.
3. **PPATK (Indonesian Financial Transaction Reports and Analysis Center)** - Handles anti-money laundering and counter-terrorism financing.

Ensure that you:

1. Register with BI if you're providing payment services
2. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
3. Maintain proper records of all transactions
4. Report suspicious transactions to PPATK
5. Comply with data protection regulations

## Testing

For testing in the sandbox environment, use the following test credentials:

### GoPay Test Credentials

```
Client ID: test_client_id_gopay
Client Secret: test_client_secret_gopay
Merchant ID: test_merchant_id_gopay
```

Test User:
- Phone Number: +6281234567890
- OTP: 123456

### OVO Test Credentials

```
App ID: test_app_id_ovo
App Key: test_app_key_ovo
Merchant ID: test_merchant_id_ovo
```

Test User:
- Phone Number: +6281234567890
- OTP: 123456

## Troubleshooting

### Common Issues with GoPay

1. **QR Code Not Generating**
   - Check if your client ID and secret are correct
   - Ensure the API endpoint is accessible

2. **Payment Timeout**
   - The default payment timeout is 15 minutes
   - Ensure the customer completes the payment within this timeframe

### Common Issues with OVO

1. **Payment Failed with Error Code 400**
   - Verify that the customer's phone number is registered with OVO
   - Check if the amount is within the allowed limits

2. **Callback Not Received**
   - Ensure your callback URL is publicly accessible
   - Check if the callback URL is correctly configured

## Support

For Indonesia-specific support, contact our support team at indonesia-support@asiancryptopayment.com.

# Thailand Payment Integration

This guide provides specific details for integrating with Thai payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Thailand:

1. **PromptPay** - QR code-based payment system
2. **TrueMoney** - E-wallet payment system

## Currency

The official currency for Thailand is the Thai Baht (THB).

## PromptPay Integration

### Configuration

To integrate with PromptPay, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  thailand_promptpay:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.scb.co.th"
    callback_url: "https://your-domain.com/callback/thailand/promptpay"
    redirect_url: "https://your-domain.com/redirect/thailand/promptpay"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using PromptPay
const payment = await paymentSDK.createPayment({
  amount: 1000.00,  // THB 1,000
  currency: 'THB',
  paymentMethod: 'qrcode',
  paymentPlatform: 'thailand_promptpay',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+66812345678'
});

// Display PromptPay QR code
displayQRCode(payment.qrCodeUrl);
```

#### POS SDK

```java
// Create a payment using PromptPay
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(1000.00)  // THB 1,000
    .setCurrency("THB")
    .setPaymentMethod(PaymentMethod.QR_CODE)
    .setPaymentPlatform("thailand_promptpay")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+66812345678")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using PromptPay
PaymentRequest request = {
    .amount = 1000.00,  // THB 1,000
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
```

## TrueMoney Integration

### Configuration

To integrate with TrueMoney, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  thailand_truemoney:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.truemoney.com"
    callback_url: "https://your-domain.com/callback/thailand/truemoney"
    redirect_url: "https://your-domain.com/redirect/thailand/truemoney"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using TrueMoney
const payment = await paymentSDK.createPayment({
  amount: 1000.00,  // THB 1,000
  currency: 'THB',
  paymentMethod: 'ewallet',
  paymentPlatform: 'thailand_truemoney',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+66812345678'
});

// Redirect to TrueMoney payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using TrueMoney
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(1000.00)  // THB 1,000
    .setCurrency("THB")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("thailand_truemoney")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+66812345678")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
} else if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using TrueMoney
PaymentRequest request = {
    .amount = 1000.00,  // THB 1,000
    .currency = "THB",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "thailand_truemoney",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+66812345678"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
} else if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Thailand, you must comply with the following regulations:

1. **Bank of Thailand (BOT)** - The central bank of Thailand regulates payment systems and financial services.
2. **Securities and Exchange Commission (SEC)** - Regulates digital assets and cryptocurrencies.
3. **Anti-Money Laundering Office (AMLO)** - Handles anti-money laundering and counter-terrorism financing.

Ensure that you:

1. Register with the SEC if you're dealing with digital assets
2. Obtain necessary licenses from the BOT for payment services
3. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
4. Maintain proper records of all transactions
5. Report suspicious transactions to AMLO
6. Comply with the Personal Data Protection Act (PDPA) for handling customer data

## Testing

For testing in the sandbox environment, use the following test credentials:

### PromptPay Test Credentials

```
Merchant ID: test_merchant_promptpay
Merchant Key: test_merchant_key_promptpay
```

Test User:
- Phone Number: +66812345678
- National ID: 1234567890123

### TrueMoney Test Credentials

```
Merchant ID: test_merchant_truemoney
Merchant Key: test_merchant_key_truemoney
```

Test User:
- Phone Number: +66812345678
- OTP: 123456

## Troubleshooting

### Common Issues with PromptPay

1. **QR Code Not Generating**
   - Check if your merchant ID and key are correct
   - Ensure the API endpoint is accessible

2. **Payment Not Received**
   - Verify that the customer has scanned the QR code correctly
   - Check if the customer's bank supports PromptPay

### Common Issues with TrueMoney

1. **Payment Failed with Error Code 400**
   - Verify that the customer's phone number is registered with TrueMoney
   - Check if the amount is within the allowed limits

2. **Callback Not Received**
   - Ensure your callback URL is publicly accessible
   - Check if the callback URL is correctly configured

## Support

For Thailand-specific support, contact our support team at thailand-support@asiancryptopayment.com.

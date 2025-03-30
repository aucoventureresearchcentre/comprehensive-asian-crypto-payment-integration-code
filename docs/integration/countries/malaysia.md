# Malaysia Payment Integration

This guide provides specific details for integrating with Malaysian payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Malaysia:

1. **FPX (Financial Process Exchange)** - Bank transfer payment system
2. **GrabPay** - E-wallet payment system

## Currency

The official currency for Malaysia is the Malaysian Ringgit (MYR).

## FPX Integration

### Configuration

To integrate with FPX, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  malaysia_fpx:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.example.com"
    callback_url: "https://your-domain.com/callback/malaysia/fpx"
    redirect_url: "https://your-domain.com/redirect/malaysia/fpx"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using FPX
const payment = await paymentSDK.createPayment({
  amount: 100.50,
  currency: 'MYR',
  paymentMethod: 'bank_transfer',
  paymentPlatform: 'malaysia_fpx',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+60123456789'
});

// Redirect to FPX payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using FPX
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100.50)
    .setCurrency("MYR")
    .setPaymentMethod(PaymentMethod.BANK_TRANSFER)
    .setPaymentPlatform("malaysia_fpx")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+60123456789")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display payment instructions
displayPaymentInstructions(payment.getPaymentUrl());
```

#### Kiosk SDK

```cpp
// Create a payment using FPX
PaymentRequest request = {
    .amount = 100.50,
    .currency = "MYR",
    .payment_method = PAYMENT_METHOD_BANK_TRANSFER,
    .payment_platform = "malaysia_fpx",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+60123456789"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display payment instructions
display_payment_instructions(payment->payment_url);
```

### FPX Bank List

FPX supports the following banks in Malaysia:

- Maybank
- CIMB Bank
- Public Bank
- RHB Bank
- Hong Leong Bank
- AmBank
- Bank Islam
- Bank Rakyat
- Alliance Bank
- Standard Chartered Bank
- OCBC Bank
- HSBC Bank
- UOB Bank
- Bank Muamalat
- Affin Bank

## GrabPay Integration

### Configuration

To integrate with GrabPay, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  malaysia_grabpay:
    client_id: "your_client_id"
    client_secret: "your_client_secret"
    merchant_id: "your_merchant_id"
    api_endpoint: "https://partner-api.grab.com"
    callback_url: "https://your-domain.com/callback/malaysia/grabpay"
    redirect_url: "https://your-domain.com/redirect/malaysia/grabpay"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using GrabPay
const payment = await paymentSDK.createPayment({
  amount: 100.50,
  currency: 'MYR',
  paymentMethod: 'ewallet',
  paymentPlatform: 'malaysia_grabpay',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+60123456789'
});

// Redirect to GrabPay payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using GrabPay
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100.50)
    .setCurrency("MYR")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("malaysia_grabpay")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+60123456789")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using GrabPay
PaymentRequest request = {
    .amount = 100.50,
    .currency = "MYR",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "malaysia_grabpay",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+60123456789"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
}
```

## Regulatory Compliance

When operating in Malaysia, you must comply with the following regulations:

1. **Bank Negara Malaysia (BNM)** - The central bank of Malaysia regulates payment systems and currency matters.
2. **Securities Commission Malaysia (SC)** - Regulates digital assets and cryptocurrencies.

Ensure that you:

1. Register with the SC if you're dealing with digital assets
2. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
3. Maintain proper records of all transactions
4. Report suspicious transactions to the relevant authorities

## Testing

For testing in the sandbox environment, use the following test credentials:

### FPX Test Credentials

```
Merchant ID: test_merchant_fpx
Merchant Key: test_merchant_key_fpx
```

Test Bank Account:
- Bank: Test Bank
- Account Number: 1234567890
- Username: testuser
- Password: testpassword

### GrabPay Test Credentials

```
Client ID: test_client_id_grabpay
Client Secret: test_client_secret_grabpay
Merchant ID: test_merchant_id_grabpay
```

Test User:
- Phone Number: +60123456789
- OTP: 123456

## Troubleshooting

### Common Issues with FPX

1. **Bank Selection Page Not Loading**
   - Check if your merchant ID and key are correct
   - Ensure the API endpoint is accessible

2. **Payment Failed with Error Code 400**
   - Verify that the amount is within the allowed limits (MYR 1 - MYR 30,000)
   - Check if the customer's bank is supported

### Common Issues with GrabPay

1. **QR Code Not Generating**
   - Verify your client ID and secret
   - Check if the API endpoint is accessible

2. **Payment Timeout**
   - The default payment timeout is 5 minutes
   - Ensure the customer completes the payment within this timeframe

## Support

For Malaysia-specific support, contact our support team at malaysia-support@asiancryptopayment.com.

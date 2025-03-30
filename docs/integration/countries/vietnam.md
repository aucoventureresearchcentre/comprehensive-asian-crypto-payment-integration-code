# Vietnam Payment Integration

This guide provides specific details for integrating with Vietnamese payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Vietnam:

1. **MoMo** - E-wallet payment system
2. **VNPay** - Payment gateway and QR code-based payment system

## Currency

The official currency for Vietnam is the Vietnamese Dong (VND).

## MoMo Integration

### Configuration

To integrate with MoMo, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  vietnam_momo:
    partner_code: "your_partner_code"
    access_key: "your_access_key"
    secret_key: "your_secret_key"
    api_endpoint: "https://payment.momo.vn"
    callback_url: "https://your-domain.com/callback/vietnam/momo"
    redirect_url: "https://your-domain.com/redirect/vietnam/momo"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using MoMo
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // VND 100,000
  currency: 'VND',
  paymentMethod: 'ewallet',
  paymentPlatform: 'vietnam_momo',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+84901234567'
});

// Redirect to MoMo payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using MoMo
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // VND 100,000
    .setCurrency("VND")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("vietnam_momo")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+84901234567")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
} else if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using MoMo
PaymentRequest request = {
    .amount = 100000.00,  // VND 100,000
    .currency = "VND",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "vietnam_momo",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+84901234567"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
} else if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## VNPay Integration

### Configuration

To integrate with VNPay, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  vietnam_vnpay:
    merchant_id: "your_merchant_id"
    secure_hash: "your_secure_hash"
    api_endpoint: "https://vnpayment.vn"
    callback_url: "https://your-domain.com/callback/vietnam/vnpay"
    redirect_url: "https://your-domain.com/redirect/vietnam/vnpay"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using VNPay
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // VND 100,000
  currency: 'VND',
  paymentMethod: 'bank_transfer',  // or 'credit_card', 'qrcode'
  paymentPlatform: 'vietnam_vnpay',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+84901234567'
});

// Redirect to VNPay payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using VNPay
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // VND 100,000
    .setCurrency("VND")
    .setPaymentMethod(PaymentMethod.BANK_TRANSFER)  // or CREDIT_CARD, QR_CODE
    .setPaymentPlatform("vietnam_vnpay")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+84901234567")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using VNPay
PaymentRequest request = {
    .amount = 100000.00,  // VND 100,000
    .currency = "VND",
    .payment_method = PAYMENT_METHOD_BANK_TRANSFER,  // or CREDIT_CARD, QR_CODE
    .payment_platform = "vietnam_vnpay",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+84901234567"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Vietnam, you must comply with the following regulations:

1. **State Bank of Vietnam (SBV)** - The central bank of Vietnam regulates payment systems and financial services.
2. **Ministry of Industry and Trade (MOIT)** - Regulates e-commerce activities.
3. **Ministry of Information and Communications (MIC)** - Regulates information technology and telecommunications.

Ensure that you:

1. Obtain necessary licenses from the SBV for payment services
2. Register with MOIT for e-commerce activities
3. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
4. Maintain proper records of all transactions
5. Report suspicious transactions to the relevant authorities
6. Comply with data protection regulations

## Testing

For testing in the sandbox environment, use the following test credentials:

### MoMo Test Credentials

```
Partner Code: test_partner_code_momo
Access Key: test_access_key_momo
Secret Key: test_secret_key_momo
```

Test User:
- Phone Number: +84901234567
- OTP: 123456

### VNPay Test Credentials

```
Merchant ID: test_merchant_id_vnpay
Secure Hash: test_secure_hash_vnpay
```

Test Card:
- Card Number: 9704198526191432198
- Card Name: NGUYEN VAN A
- Expiry Date: 07/15
- OTP: 123456

## Troubleshooting

### Common Issues with MoMo

1. **QR Code Not Generating**
   - Check if your partner code, access key, and secret key are correct
   - Ensure the API endpoint is accessible

2. **Payment Timeout**
   - The default payment timeout is 15 minutes
   - Ensure the customer completes the payment within this timeframe

### Common Issues with VNPay

1. **Payment Page Not Loading**
   - Verify your merchant ID and secure hash
   - Check if the API endpoint is accessible

2. **Transaction Declined**
   - Ensure the test card details are entered correctly
   - Check if the amount is within the allowed limits

## Support

For Vietnam-specific support, contact our support team at vietnam-support@asiancryptopayment.com.

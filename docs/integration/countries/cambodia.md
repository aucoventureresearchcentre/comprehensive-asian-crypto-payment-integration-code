# Cambodia Payment Integration

This guide provides specific details for integrating with Cambodian payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Cambodia:

1. **Wing** - Mobile money and payment service
2. **ABA** - Bank payment system

## Currency

The official currencies for Cambodia are the Cambodian Riel (KHR) and US Dollar (USD). Both currencies are widely accepted.

## Wing Integration

### Configuration

To integrate with Wing, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  cambodia_wing:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.wingmoney.com"
    callback_url: "https://your-domain.com/callback/cambodia/wing"
    redirect_url: "https://your-domain.com/redirect/cambodia/wing"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using Wing
const payment = await paymentSDK.createPayment({
  amount: 10.00,  // USD 10.00 or KHR 40,000
  currency: 'USD',  // or 'KHR'
  paymentMethod: 'ewallet',
  paymentPlatform: 'cambodia_wing',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+85512345678'
});

// Redirect to Wing payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using Wing
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(10.00)  // USD 10.00 or KHR 40,000
    .setCurrency("USD")  // or "KHR"
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("cambodia_wing")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+85512345678")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code or payment instructions
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
} else if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using Wing
PaymentRequest request = {
    .amount = 10.00,  // USD 10.00 or KHR 40,000
    .currency = "USD",  // or "KHR"
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "cambodia_wing",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+85512345678"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code or payment instructions
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
} else if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## ABA Integration

### Configuration

To integrate with ABA, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  cambodia_aba:
    merchant_id: "your_merchant_id"
    merchant_api_key: "your_merchant_api_key"
    merchant_secret: "your_merchant_secret"
    api_endpoint: "https://checkout.payway.com.kh"
    callback_url: "https://your-domain.com/callback/cambodia/aba"
    redirect_url: "https://your-domain.com/redirect/cambodia/aba"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using ABA
const payment = await paymentSDK.createPayment({
  amount: 10.00,  // USD 10.00 or KHR 40,000
  currency: 'USD',  // or 'KHR'
  paymentMethod: 'credit_card',  // or 'ewallet', 'qrcode'
  paymentPlatform: 'cambodia_aba',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+85512345678'
});

// Redirect to ABA payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using ABA
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(10.00)  // USD 10.00 or KHR 40,000
    .setCurrency("USD")  // or "KHR"
    .setPaymentMethod(PaymentMethod.CREDIT_CARD)  // or EWALLET, QR_CODE
    .setPaymentPlatform("cambodia_aba")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+85512345678")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using ABA
PaymentRequest request = {
    .amount = 10.00,  // USD 10.00 or KHR 40,000
    .currency = "USD",  // or "KHR"
    .payment_method = PAYMENT_METHOD_CREDIT_CARD,  // or EWALLET, QR_CODE
    .payment_platform = "cambodia_aba",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+85512345678"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Cambodia, you must comply with the following regulations:

1. **National Bank of Cambodia (NBC)** - The central bank of Cambodia regulates payment systems and financial services.
2. **Financial Intelligence Unit (FIU)** - Handles anti-money laundering and counter-terrorism financing.

Ensure that you:

1. Obtain necessary licenses from the NBC for payment services
2. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
3. Maintain proper records of all transactions
4. Report suspicious transactions to the FIU
5. Comply with data protection regulations

## Testing

For testing in the sandbox environment, use the following test credentials:

### Wing Test Credentials

```
Merchant ID: test_merchant_wing
Merchant Key: test_merchant_key_wing
```

Test User:
- Phone Number: +85512345678
- PIN: 123456

### ABA Test Credentials

```
Merchant ID: test_merchant_aba
Merchant API Key: test_merchant_api_key_aba
Merchant Secret: test_merchant_secret_aba
```

Test Card:
- Card Number: 4111 1111 1111 1111
- Expiry Date: 12/25
- CVV: 123

## Troubleshooting

### Common Issues with Wing

1. **Payment Failed with Error Code 400**
   - Verify that the customer's phone number is registered with Wing
   - Check if the amount is within the allowed limits

2. **Callback Not Received**
   - Ensure your callback URL is publicly accessible
   - Check if the callback URL is correctly configured

### Common Issues with ABA

1. **Payment Page Not Loading**
   - Verify your merchant ID, API key, and secret
   - Check if the API endpoint is accessible

2. **Transaction Declined**
   - Ensure the test card details are entered correctly
   - Check if the amount is within the allowed limits

## Support

For Cambodia-specific support, contact our support team at cambodia-support@asiancryptopayment.com.

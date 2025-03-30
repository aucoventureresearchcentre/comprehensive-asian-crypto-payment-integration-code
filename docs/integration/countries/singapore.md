# Singapore Payment Integration

This guide provides specific details for integrating with Singaporean payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Singapore:

1. **PayNow** - QR code-based payment system
2. **NETS** - Electronic payment system

## Currency

The official currency for Singapore is the Singapore Dollar (SGD).

## PayNow Integration

### Configuration

To integrate with PayNow, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  singapore_paynow:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.example.com"
    callback_url: "https://your-domain.com/callback/singapore/paynow"
    redirect_url: "https://your-domain.com/redirect/singapore/paynow"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using PayNow
const payment = await paymentSDK.createPayment({
  amount: 100.50,
  currency: 'SGD',
  paymentMethod: 'qrcode',
  paymentPlatform: 'singapore_paynow',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+6591234567'
});

// Display PayNow QR code
displayQRCode(payment.qrCodeUrl);
```

#### POS SDK

```java
// Create a payment using PayNow
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100.50)
    .setCurrency("SGD")
    .setPaymentMethod(PaymentMethod.QR_CODE)
    .setPaymentPlatform("singapore_paynow")
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
```

#### Kiosk SDK

```cpp
// Create a payment using PayNow
PaymentRequest request = {
    .amount = 100.50,
    .currency = "SGD",
    .payment_method = PAYMENT_METHOD_QR_CODE,
    .payment_platform = "singapore_paynow",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+6591234567"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
}
```

## NETS Integration

### Configuration

To integrate with NETS, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  singapore_nets:
    merchant_id: "your_merchant_id"
    terminal_id: "your_terminal_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.nets.com.sg"
    callback_url: "https://your-domain.com/callback/singapore/nets"
    redirect_url: "https://your-domain.com/redirect/singapore/nets"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using NETS
const payment = await paymentSDK.createPayment({
  amount: 100.50,
  currency: 'SGD',
  paymentMethod: 'credit_card',
  paymentPlatform: 'singapore_nets',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+6591234567'
});

// Redirect to NETS payment page
window.location.href = payment.paymentUrl;
```

#### POS SDK

```java
// Create a payment using NETS
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100.50)
    .setCurrency("SGD")
    .setPaymentMethod(PaymentMethod.CREDIT_CARD)
    .setPaymentPlatform("singapore_nets")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+6591234567")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using NETS
PaymentRequest request = {
    .amount = 100.50,
    .currency = "SGD",
    .payment_method = PAYMENT_METHOD_CREDIT_CARD,
    .payment_platform = "singapore_nets",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+6591234567"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Singapore, you must comply with the following regulations:

1. **Monetary Authority of Singapore (MAS)** - The central bank and financial regulatory authority of Singapore.
2. **Payment Services Act (PSA)** - Regulates payment systems and payment service providers.

Ensure that you:

1. Register with MAS if you're providing digital payment token services
2. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
3. Maintain proper records of all transactions
4. Report suspicious transactions to the relevant authorities
5. Comply with the Personal Data Protection Act (PDPA) for handling customer data

## Testing

For testing in the sandbox environment, use the following test credentials:

### PayNow Test Credentials

```
Merchant ID: test_merchant_paynow
Merchant Key: test_merchant_key_paynow
```

Test User:
- Phone Number: +6591234567
- UEN: 123456789A

### NETS Test Credentials

```
Merchant ID: test_merchant_nets
Terminal ID: test_terminal_nets
Merchant Key: test_merchant_key_nets
```

Test Card:
- Card Number: 4111 1111 1111 1111
- Expiry Date: 12/25
- CVV: 123

## Troubleshooting

### Common Issues with PayNow

1. **QR Code Not Generating**
   - Check if your merchant ID and key are correct
   - Ensure the API endpoint is accessible

2. **Payment Not Received**
   - Verify that the customer has scanned the QR code correctly
   - Check if the customer's bank supports PayNow

### Common Issues with NETS

1. **Payment Page Not Loading**
   - Verify your merchant ID and terminal ID
   - Check if the API endpoint is accessible

2. **Transaction Declined**
   - Ensure the test card details are entered correctly
   - Check if the amount is within the allowed limits

## Support

For Singapore-specific support, contact our support team at singapore-support@asiancryptopayment.com.

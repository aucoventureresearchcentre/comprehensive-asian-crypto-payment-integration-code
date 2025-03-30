# Country-Specific Integration Guides

This directory contains integration guides for each supported country in the Asian Cryptocurrency Payment System.

## Supported Countries

The system currently supports the following countries:

1. [Malaysia](malaysia.md) - FPX, GrabPay
2. [Singapore](singapore.md) - PayNow, NETS
3. [Indonesia](indonesia.md) - GoPay, OVO
4. [Thailand](thailand.md) - PromptPay, TrueMoney
5. [Vietnam](vietnam.md) - MoMo, VNPay
6. [Cambodia](cambodia.md) - Wing, ABA
7. [Laos](laos.md) - U-Money, LDB

## Integration Overview

Each country guide provides:

- Supported payment platforms
- Configuration details
- Usage examples for Web, POS, and Kiosk integration
- Regulatory compliance information
- Testing credentials
- Troubleshooting tips

## Currency Support

The system supports the following currencies:

- Malaysian Ringgit (MYR)
- Singapore Dollar (SGD)
- Indonesian Rupiah (IDR)
- Thai Baht (THB)
- Vietnamese Dong (VND)
- Cambodian Riel (KHR)
- US Dollar (USD) - for Cambodia
- Lao Kip (LAK)

## Common Integration Steps

Regardless of the country, the general integration process follows these steps:

1. Configure the payment platform in your `config.yaml` file
2. Initialize the SDK with your API key
3. Create a payment request with the appropriate payment platform and method
4. Process the payment response (redirect URL or QR code)
5. Handle the callback or check payment status
6. Update your system when payment is completed

For detailed instructions specific to each country, please refer to the individual country guides.

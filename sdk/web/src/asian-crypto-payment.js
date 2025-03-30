/**
 * Asian Cryptocurrency Payment System - Web SDK
 * Version: 1.0.0
 * 
 * This SDK provides a comprehensive interface for integrating cryptocurrency
 * payments into websites across Asian markets, with support for regulatory
 * compliance in Malaysia, Singapore, Indonesia, Thailand, Brunei, Cambodia,
 * Vietnam, and Laos.
 */

class AsianCryptoPayment {
  /**
   * Initialize the SDK with merchant credentials and configuration
   * 
   * @param {Object} config - Configuration object
   * @param {string} config.apiKey - Merchant API key
   * @param {string} config.merchantId - Merchant ID
   * @param {string} config.countryCode - Two-letter country code (MY, SG, ID, TH, BN, KH, VN, LA)
   * @param {boolean} [config.testMode=false] - Whether to use test mode
   * @param {string} [config.apiEndpoint] - Custom API endpoint (optional)
   * @param {Object} [config.webhookConfig] - Webhook configuration (optional)
   * @param {string[]} [config.supportedCryptocurrencies] - List of supported cryptocurrencies (optional)
   */
  constructor(config) {
    // Validate required parameters
    if (!config.apiKey) throw new Error('API key is required');
    if (!config.merchantId) throw new Error('Merchant ID is required');
    if (!config.countryCode) throw new Error('Country code is required');
    
    // Validate country code
    const validCountryCodes = ['MY', 'SG', 'ID', 'TH', 'BN', 'KH', 'VN', 'LA'];
    if (!validCountryCodes.includes(config.countryCode)) {
      throw new Error(`Invalid country code. Must be one of: ${validCountryCodes.join(', ')}`);
    }
    
    // Initialize configuration
    this.config = {
      apiKey: config.apiKey,
      merchantId: config.merchantId,
      countryCode: config.countryCode,
      testMode: config.testMode || false,
      apiEndpoint: config.apiEndpoint || 'https://api.asiancryptopay.com',
      webhookConfig: config.webhookConfig || null,
      supportedCryptocurrencies: config.supportedCryptocurrencies || [
        'BTC', 'ETH', 'USDT', 'USDC', 'BNB'
      ]
    };
    
    // Initialize state
    this.state = {
      initialized: true,
      lastResponse: null,
      activePayments: new Map()
    };
    
    // Initialize country-specific module
    this._initializeCountryModule();
    
    // Initialize security module
    this._initializeSecurity();
    
    console.log(`Asian Cryptocurrency Payment SDK initialized for ${this.config.countryCode}`);
  }
  
  /**
   * Initialize country-specific regulatory compliance module
   * @private
   */
  _initializeCountryModule() {
    // Map country codes to regulatory modules
    const countryModules = {
      'MY': MalaysiaComplianceModule,
      'SG': SingaporeComplianceModule,
      'ID': IndonesiaComplianceModule,
      'TH': ThailandComplianceModule,
      'BN': BruneiComplianceModule,
      'KH': CambodiaComplianceModule,
      'VN': VietnamComplianceModule,
      'LA': LaosComplianceModule
    };
    
    // Initialize the appropriate module
    const CountryModule = countryModules[this.config.countryCode];
    this.countryModule = new CountryModule(this.config);
    
    // Apply country-specific configurations
    this.countryModule.applyRegulations();
  }
  
  /**
   * Initialize security features
   * @private
   */
  _initializeSecurity() {
    this.security = {
      encryptData: (data) => {
        // Implementation of AES-256 encryption
        return btoa(JSON.stringify(data)); // Simplified for example
      },
      decryptData: (encryptedData) => {
        // Implementation of AES-256 decryption
        return JSON.parse(atob(encryptedData)); // Simplified for example
      },
      generateSignature: (payload, timestamp) => {
        // Implementation of HMAC-SHA256 signature
        const message = `${timestamp}.${JSON.stringify(payload)}`;
        // In a real implementation, use a proper HMAC library
        return this._sha256(this.config.apiKey + message);
      },
      verifySignature: (signature, payload, timestamp) => {
        const expectedSignature = this.security.generateSignature(payload, timestamp);
        return signature === expectedSignature;
      }
    };
  }
  
  /**
   * Simple SHA-256 implementation for example purposes
   * In production, use a proper crypto library
   * @private
   */
  _sha256(message) {
    // This is a placeholder - in production use a proper crypto library
    let hash = 0;
    for (let i = 0; i < message.length; i++) {
      const char = message.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return hash.toString(16);
  }
  
  /**
   * Make an API request to the backend
   * @private
   * @param {string} endpoint - API endpoint
   * @param {string} method - HTTP method
   * @param {Object} [data] - Request payload
   * @returns {Promise<Object>} - API response
   */
  async _apiRequest(endpoint, method, data = null) {
    const url = `${this.config.apiEndpoint}/${endpoint}`;
    const timestamp = Date.now().toString();
    
    const headers = {
      'Content-Type': 'application/json',
      'X-Merchant-ID': this.config.merchantId,
      'X-Timestamp': timestamp,
      'X-Test-Mode': this.config.testMode ? 'true' : 'false'
    };
    
    if (data) {
      headers['X-Signature'] = this.security.generateSignature(data, timestamp);
    }
    
    const requestOptions = {
      method,
      headers,
      body: data ? JSON.stringify(data) : undefined
    };
    
    try {
      const response = await fetch(url, requestOptions);
      const responseData = await response.json();
      
      // Store last response for debugging
      this.state.lastResponse = responseData;
      
      if (!response.ok) {
        throw new Error(responseData.message || 'API request failed');
      }
      
      return responseData;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }
  
  /**
   * Create a new cryptocurrency payment
   * 
   * @param {Object} paymentDetails - Payment details
   * @param {number} paymentDetails.amount - Payment amount
   * @param {string} paymentDetails.currency - Fiat currency code (e.g., MYR, SGD)
   * @param {string} paymentDetails.cryptoCurrency - Cryptocurrency code (e.g., BTC, ETH)
   * @param {string} [paymentDetails.description] - Payment description
   * @param {string} [paymentDetails.orderId] - Merchant order ID
   * @param {string} [paymentDetails.customerEmail] - Customer email
   * @param {string} [paymentDetails.customerName] - Customer name
   * @param {string} [paymentDetails.callbackUrl] - Callback URL for payment updates
   * @param {string} [paymentDetails.successUrl] - Redirect URL on successful payment
   * @param {string} [paymentDetails.cancelUrl] - Redirect URL on cancelled payment
   * @param {Object} [paymentDetails.metadata] - Additional metadata
   * @returns {Promise<Object>} - Payment object
   */
  async createPayment(paymentDetails) {
    // Validate required parameters
    if (!paymentDetails.amount) throw new Error('Payment amount is required');
    if (!paymentDetails.currency) throw new Error('Currency is required');
    if (!paymentDetails.cryptoCurrency) throw new Error('Cryptocurrency is required');
    
    // Validate amount
    if (paymentDetails.amount <= 0) {
      throw new Error('Payment amount must be greater than zero');
    }
    
    // Validate cryptocurrency
    if (!this.config.supportedCryptocurrencies.includes(paymentDetails.cryptoCurrency)) {
      throw new Error(`Unsupported cryptocurrency. Must be one of: ${this.config.supportedCryptocurrencies.join(', ')}`);
    }
    
    // Apply country-specific validations
    this.countryModule.validatePayment(paymentDetails);
    
    // Prepare payment data
    const paymentData = {
      merchant_id: this.config.merchantId,
      amount: paymentDetails.amount,
      currency: paymentDetails.currency,
      crypto_currency: paymentDetails.cryptoCurrency,
      description: paymentDetails.description || '',
      order_id: paymentDetails.orderId || `order-${Date.now()}`,
      customer_email: paymentDetails.customerEmail || '',
      customer_name: paymentDetails.customerName || '',
      callback_url: paymentDetails.callbackUrl || '',
      success_url: paymentDetails.successUrl || '',
      cancel_url: paymentDetails.cancelUrl || '',
      metadata: paymentDetails.metadata || {},
      country_code: this.config.countryCode,
      test_mode: this.config.testMode
    };
    
    // Make API request
    const payment = await this._apiRequest('payments', 'POST', paymentData);
    
    // Store active payment
    this.state.activePayments.set(payment.id, payment);
    
    return payment;
  }
  
  /**
   * Get payment details by ID
   * 
   * @param {string} paymentId - Payment ID
   * @returns {Promise<Object>} - Payment object
   */
  async getPayment(paymentId) {
    if (!paymentId) throw new Error('Payment ID is required');
    
    const payment = await this._apiRequest(`payments/${paymentId}`, 'GET');
    
    // Update active payment if exists
    if (this.state.activePayments.has(paymentId)) {
      this.state.activePayments.set(paymentId, payment);
    }
    
    return payment;
  }
  
  /**
   * Get list of payments
   * 
   * @param {Object} [filters] - Filter parameters
   * @param {string} [filters.status] - Filter by status
   * @param {string} [filters.fromDate] - Filter by date range start
   * @param {string} [filters.toDate] - Filter by date range end
   * @param {number} [filters.limit=20] - Number of results to return
   * @param {number} [filters.offset=0] - Offset for pagination
   * @returns {Promise<Object>} - List of payments
   */
  async getPayments(filters = {}) {
    const queryParams = new URLSearchParams();
    
    if (filters.status) queryParams.append('status', filters.status);
    if (filters.fromDate) queryParams.append('from_date', filters.fromDate);
    if (filters.toDate) queryParams.append('to_date', filters.toDate);
    if (filters.limit) queryParams.append('limit', filters.limit.toString());
    if (filters.offset) queryParams.append('offset', filters.offset.toString());
    
    const endpoint = `payments?${queryParams.toString()}`;
    return await this._apiRequest(endpoint, 'GET');
  }
  
  /**
   * Cancel a payment
   * 
   * @param {string} paymentId - Payment ID
   * @returns {Promise<Object>} - Cancelled payment object
   */
  async cancelPayment(paymentId) {
    if (!paymentId) throw new Error('Payment ID is required');
    
    const payment = await this._apiRequest(`payments/${paymentId}/cancel`, 'POST');
    
    // Update active payment if exists
    if (this.state.activePayments.has(paymentId)) {
      this.state.activePayments.set(paymentId, payment);
    }
    
    return payment;
  }
  
  /**
   * Get current exchange rates
   * 
   * @param {string} [baseCurrency='USD'] - Base currency
   * @param {string[]} [cryptoCurrencies] - List of cryptocurrencies to get rates for
   * @returns {Promise<Object>} - Exchange rates
   */
  async getExchangeRates(baseCurrency = 'USD', cryptoCurrencies = null) {
    const currencies = cryptoCurrencies || this.config.supportedCryptocurrencies;
    
    const queryParams = new URLSearchParams();
    queryParams.append('base_currency', baseCurrency);
    queryParams.append('currencies', currencies.join(','));
    
    const endpoint = `exchange-rates?${queryParams.toString()}`;
    return await this._apiRequest(endpoint, 'GET');
  }
  
  /**
   * Create a payment button that can be added to a webpage
   * 
   * @param {string|HTMLElement} container - Container element or selector
   * @param {Object} paymentDetails - Payment details
   * @param {Object} [options] - Button options
   * @param {string} [options.buttonText='Pay with Cryptocurrency'] - Button text
   * @param {string} [options.buttonClass='asian-crypto-pay-button'] - Button CSS class
   * @param {string} [options.theme='light'] - Button theme (light or dark)
   * @param {Function} [options.onSuccess] - Success callback
   * @param {Function} [options.onCancel] - Cancel callback
   * @param {Function} [options.onError] - Error callback
   * @returns {HTMLElement} - Button element
   */
  createPaymentButton(container, paymentDetails, options = {}) {
    // Default options
    const defaultOptions = {
      buttonText: 'Pay with Cryptocurrency',
      buttonClass: 'asian-crypto-pay-button',
      theme: 'light',
      onSuccess: () => {},
      onCancel: () => {},
      onError: () => {}
    };
    
    // Merge options
    const buttonOptions = { ...defaultOptions, ...options };
    
    // Get container element
    const containerElement = typeof container === 'string' 
      ? document.querySelector(container) 
      : container;
    
    if (!containerElement) {
      throw new Error('Container element not found');
    }
    
    // Create button element
    const button = document.createElement('button');
    button.textContent = buttonOptions.buttonText;
    button.className = buttonOptions.buttonClass;
    button.dataset.theme = buttonOptions.theme;
    
    // Add click event listener
    button.addEventListener('click', async () => {
      try {
        // Create payment
        const payment = await this.createPayment(paymentDetails);
        
        // Open payment modal
        this._openPaymentModal(payment, buttonOptions);
      } catch (error) {
        console.error('Payment creation failed:', error);
        buttonOptions.onError(error);
      }
    });
    
    // Add button to container
    containerElement.appendChild(button);
    
    return button;
  }
  
  /**
   * Open payment modal
   * @private
   * @param {Object} payment - Payment object
   * @param {Object} options - Modal options
   */
  _openPaymentModal(payment, options) {
    // Create modal container
    const modal = document.createElement('div');
    modal.className = 'asian-crypto-pay-modal';
    modal.dataset.theme = options.theme;
    
    // Create modal content
    const modalContent = document.createElement('div');
    modalContent.className = 'asian-crypto-pay-modal-content';
    
    // Create modal header
    const modalHeader = document.createElement('div');
    modalHeader.className = 'asian-crypto-pay-modal-header';
    
    const modalTitle = document.createElement('h3');
    modalTitle.textContent = 'Cryptocurrency Payment';
    
    const closeButton = document.createElement('button');
    closeButton.className = 'asian-crypto-pay-modal-close';
    closeButton.textContent = '×';
    closeButton.addEventListener('click', () => {
      document.body.removeChild(modal);
      options.onCancel();
    });
    
    modalHeader.appendChild(modalTitle);
    modalHeader.appendChild(closeButton);
    
    // Create modal body
    const modalBody = document.createElement('div');
    modalBody.className = 'asian-crypto-pay-modal-body';
    
    const paymentInfo = document.createElement('div');
    paymentInfo.className = 'asian-crypto-pay-payment-info';
    
    const amountInfo = document.createElement('p');
    amountInfo.innerHTML = `<strong>Amount:</strong> ${payment.crypto_amount} ${payment.crypto_currency}`;
    
    const addressInfo = document.createElement('p');
    addressInfo.innerHTML = `<strong>Send to address:</strong>`;
    
    const addressValue = document.createElement('div');
    addressValue.className = 'asian-crypto-pay-address';
    addressValue.textContent = payment.address;
    
    const copyButton = document.createElement('button');
    copyButton.className = 'asian-crypto-pay-copy-button';
    copyButton.textContent = 'Copy';
    copyButton.addEventListener('click', () => {
      navigator.clipboard.writeText(payment.address);
      copyButton.textContent = 'Copied!';
      setTimeout(() => {
        copyButton.textContent = 'Copy';
      }, 2000);
    });
    
    addressInfo.appendChild(addressValue);
    addressInfo.appendChild(copyButton);
    
    const qrCodeContai<response clipped><NOTE>To save on context only part of this file has been shown to you. You should retry this tool after you have searched inside the file with `grep -n` in order to find the line numbers of what you are looking for.</NOTE>
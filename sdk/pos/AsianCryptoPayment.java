/**
 * Asian Cryptocurrency Payment System - POS Terminal SDK
 * Version: 1.0.0
 * 
 * This SDK provides a comprehensive interface for integrating cryptocurrency
 * payments into Point-of-Sale terminals across Asian markets, with support for 
 * regulatory compliance in Malaysia, Singapore, Indonesia, Thailand, Brunei, 
 * Cambodia, Vietnam, and Laos.
 */

package com.asiancryptopay.sdk;

import android.content.Context;
import android.graphics.Bitmap;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.math.BigDecimal;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.TimeZone;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

import okhttp3.Call;
import okhttp3.Callback;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

/**
 * Main SDK class for Asian Cryptocurrency Payment System
 */
public class AsianCryptoPayment {
    private static final String TAG = "AsianCryptoPayment";
    private static final String SDK_VERSION = "1.0.0";
    private static final MediaType JSON = MediaType.parse("application/json; charset=utf-8");
    private static final String DEFAULT_API_ENDPOINT = "https://api.asiancryptopay.com";
    
    // Configuration
    private final String apiKey;
    private final String merchantId;
    private final String countryCode;
    private final boolean testMode;
    private final String apiEndpoint;
    private final List<String> supportedCryptocurrencies;
    private final Map<String, Object> webhookConfig;
    
    // HTTP Client
    private final OkHttpClient httpClient;
    
    // Country-specific compliance module
    private final CountryComplianceModule countryModule;
    
    // Security module
    private final SecurityModule securityModule;
    
    // Android context
    private final Context context;
    
    // Handler for UI thread callbacks
    private final Handler mainHandler;
    
    /**
     * Builder class for AsianCryptoPayment
     */
    public static class Builder {
        private final String apiKey;
        private final String merchantId;
        private final String countryCode;
        private final Context context;
        
        private boolean testMode = false;
        private String apiEndpoint = DEFAULT_API_ENDPOINT;
        private List<String> supportedCryptocurrencies = new ArrayList<>();
        private Map<String, Object> webhookConfig = null;
        
        /**
         * Initialize builder with required parameters
         * 
         * @param context Android context
         * @param apiKey Merchant API key
         * @param merchantId Merchant ID
         * @param countryCode Two-letter country code (MY, SG, ID, TH, BN, KH, VN, LA)
         */
        public Builder(@NonNull Context context, @NonNull String apiKey, @NonNull String merchantId, @NonNull String countryCode) {
            this.context = context.getApplicationContext();
            this.apiKey = apiKey;
            this.merchantId = merchantId;
            this.countryCode = countryCode;
            
            // Default supported cryptocurrencies
            this.supportedCryptocurrencies.add("BTC");
            this.supportedCryptocurrencies.add("ETH");
            this.supportedCryptocurrencies.add("USDT");
            this.supportedCryptocurrencies.add("USDC");
            this.supportedCryptocurrencies.add("BNB");
        }
        
        /**
         * Set test mode
         * 
         * @param testMode Whether to use test mode
         * @return Builder instance
         */
        public Builder setTestMode(boolean testMode) {
            this.testMode = testMode;
            return this;
        }
        
        /**
         * Set custom API endpoint
         * 
         * @param apiEndpoint Custom API endpoint
         * @return Builder instance
         */
        public Builder setApiEndpoint(@NonNull String apiEndpoint) {
            this.apiEndpoint = apiEndpoint;
            return this;
        }
        
        /**
         * Set supported cryptocurrencies
         * 
         * @param supportedCryptocurrencies List of supported cryptocurrencies
         * @return Builder instance
         */
        public Builder setSupportedCryptocurrencies(@NonNull List<String> supportedCryptocurrencies) {
            this.supportedCryptocurrencies = new ArrayList<>(supportedCryptocurrencies);
            return this;
        }
        
        /**
         * Set webhook configuration
         * 
         * @param webhookEndpoint Webhook endpoint URL
         * @param webhookSecret Webhook secret for signature verification
         * @return Builder instance
         */
        public Builder setWebhookConfig(@NonNull String webhookEndpoint, @NonNull String webhookSecret) {
            this.webhookConfig = new HashMap<>();
            this.webhookConfig.put("endpoint", webhookEndpoint);
            this.webhookConfig.put("secret", webhookSecret);
            return this;
        }
        
        /**
         * Build AsianCryptoPayment instance
         * 
         * @return AsianCryptoPayment instance
         */
        public AsianCryptoPayment build() {
            return new AsianCryptoPayment(this);
        }
    }
    
    /**
     * Private constructor, use Builder instead
     * 
     * @param builder Builder instance
     */
    private AsianCryptoPayment(Builder builder) {
        this.context = builder.context;
        this.apiKey = builder.apiKey;
        this.merchantId = builder.merchantId;
        this.countryCode = builder.countryCode;
        this.testMode = builder.testMode;
        this.apiEndpoint = builder.apiEndpoint;
        this.supportedCryptocurrencies = builder.supportedCryptocurrencies;
        this.webhookConfig = builder.webhookConfig;
        
        // Initialize HTTP client
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build();
        
        // Initialize main thread handler
        this.mainHandler = new Handler(Looper.getMainLooper());
        
        // Initialize security module
        this.securityModule = new SecurityModule(this.apiKey);
        
        // Initialize country-specific module
        this.countryModule = createCountryModule(this.countryCode);
        
        Log.i(TAG, "SDK initialized for country: " + this.countryCode);
    }
    
    /**
     * Create country-specific compliance module
     * 
     * @param countryCode Two-letter country code
     * @return CountryComplianceModule instance
     */
    private CountryComplianceModule createCountryModule(String countryCode) {
        switch (countryCode) {
            case "MY":
                return new MalaysiaComplianceModule();
            case "SG":
                return new SingaporeComplianceModule();
            case "ID":
                return new IndonesiaComplianceModule();
            case "TH":
                return new ThailandComplianceModule();
            case "BN":
                return new BruneiComplianceModule();
            case "KH":
                return new CambodiaComplianceModule();
            case "VN":
                return new VietnamComplianceModule();
            case "LA":
                return new LaosComplianceModule();
            default:
                throw new IllegalArgumentException("Unsupported country code: " + countryCode);
        }
    }
    
    /**
     * Create a new cryptocurrency payment
     * 
     * @param paymentDetails Payment details
     * @param callback Callback for payment result
     */
    public void createPayment(@NonNull PaymentDetails paymentDetails, @NonNull PaymentCallback callback) {
        // Validate payment details
        try {
            validatePaymentDetails(paymentDetails);
        } catch (IllegalArgumentException e) {
            callback.onError(e);
            return;
        }
        
        // Apply country-specific validations
        try {
            countryModule.validatePayment(paymentDetails);
        } catch (IllegalArgumentException e) {
            callback.onError(e);
            return;
        }
        
        // Prepare payment data
        JSONObject paymentData = new JSONObject();
        try {
            paymentData.put("merchant_id", merchantId);
            paymentData.put("amount", paymentDetails.getAmount().toString());
            paymentData.put("currency", paymentDetails.getCurrency());
            paymentData.put("crypto_currency", paymentDetails.getCryptoCurrency());
            paymentData.put("description", paymentDetails.getDescription());
            paymentData.put("order_id", paymentDetails.getOrderId() != null ? paymentDetails.getOrderId() : "order-" + System.currentTimeMillis());
            paymentData.put("customer_email", paymentDetails.getCustomerEmail());
            paymentData.put("customer_name", paymentDetails.getCustomerName());
            paymentData.put("callback_url", paymentDetails.getCallbackUrl());
            paymentData.put("success_url", paymentDetails.getSuccessUrl());
            paymentData.put("cancel_url", paymentDetails.getCancelUrl());
            paymentData.put("country_code", countryCode);
            paymentData.put("test_mode", testMode);
            
            // Add metadata if available
            if (paymentDetails.getMetadata() != null) {
                paymentData.put("metadata", new JSONObject(paymentDetails.getMetadata()));
            }
        } catch (JSONException e) {
            callback.onError(new RuntimeException("Failed to create payment data", e));
            return;
        }
        
        // Make API request
        String endpoint = "payments";
        makeApiRequest(endpoint, "POST", paymentData, new ApiCallback() {
            @Override
            public void onSuccess(JSONObject response) {
                try {
                    Payment payment = Payment.fromJson(response);
                    callback.onSuccess(payment);
                } catch (JSONException e) {
                    callback.onError(new RuntimeException("Failed to parse payment response", e));
                }
            }
            
            @Override
            public void onError(Exception e) {
                callback.onError(e);
            }
        });
    }
    
    /**
     * Get payment details by ID
     * 
     * @param paymentId Payment ID
     * @param callback Callback for payment result
     */
    public void getPayment(@NonNull String paymentId, @NonNull PaymentCallback callback) {
        if (paymentId.isEmpty()) {
            callback.onError(new IllegalArgumentException("Payment ID is required"));
            return;
        }
        
        String endpoint = "payments/" + paymentId;
        makeApiRequest(endpoint, "GET", null, new ApiCallback() {
            @Override
            public void onSuccess(JSONObject response) {
                try {
                    Payment payment = Payment.fromJson(response);
                    callback.onSuccess(payment);
                } catch (JSONException e) {
                    callback.onError(new RuntimeException("Failed to parse payment response", e));
                }
            }
            
            @Override
            public void onError(Exception e) {
                callback.onError(e);
            }
        });
    }
    
    /**
     * Get list of payments
     * 
     * @param filters Filter parameters
     * @param callback Callback for payments result
     */
    public void getPayments(@Nullable PaymentFilters filters, @NonNull PaymentsListCallback callback) {
        StringBuilder endpoint = new StringBuilder("payments");
        
        // Add query parameters if filters are provided
        if (filters != null) {
            endpoint.append("?");
            
            if (filters.getStatus() != null) {
                endpoint.append("status=").append(filters.getStatus()).append("&");
            }
            
            if (filters.getFromDate() != null) {
                SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd", Locale.US);
                sdf.setTimeZone(TimeZone.getTimeZone("UTC"));
                endpoint.append("from_date=").append(sdf.format(filters.getFromDate())).append("&");
            }
            
            if (filters.getToDate() != null) {
                SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd", Locale.US);
                sdf.setTimeZone(TimeZone.getTimeZone("UTC"));
                endpoint.append("to_date=").append(sdf.format(filters.getToDate())).append("&");
            }
            
            if (filters.getLimit() > 0) {
                endpoint.append("limit=").append(filters.getLimit()).append("&");
            }
            
            if (filters.getOffset() > 0) {
                endpoint.append("offset=").append(filters.getOffset()).append("&");
            }
            
            // Remove trailing '&' if present
            if (endpoint.charAt(endpoint.length() - 1) == '&') {
                endpoint.deleteCharAt(endpoint.length() - 1);
            }
        }
        
        makeApiRequest(endpoint.toString(), "GET", null, new ApiCallback() {
            @Override
            public void onSuccess(JSONObject response) {
                try {
                    List<Payment> payments = new ArrayList<>();
                    JSONArray paymentsArray = response.getJSONArray("payments");
                    
                    for (int i = 0; i < paymentsArray.length(); i++) {
                        JSONObject paymentJson = paymentsArray.getJSONObject(i);
                        payments.add(Payment.fromJson(paymentJson));
                    }
                    
                    int total = response.getInt("total");
                    callback.onSuccess(payments, total);
                } catch (JSONException e) {
                    callback.onError(new RuntimeException("Failed to parse payments response", e));
                }
            }
            
            @Override
            public void onError(Exception e) {
                callback.onError(e);
            }
        });
    }
    
    /**
     * Cancel a payment
     * 
     * @param paymentId Payment ID
     * @param callback Callback for payment result
     */
    public void cancelPayment(@NonNull String paymentId, @NonNull PaymentCallback callback) {
        if (paymentId.isEmpty()) {
            callback.onError(new IllegalArgumentException("Payment ID is required"));
            return;
        }
        
        String endpoint = "payments/" + paymentId + "/cancel";
        makeApiRequest(endpoint, "POST", null, new ApiCallback() {
            @Override
            public void onSuccess(JSONObject response) {
                try {
                    Payment payment = Payment.fromJson(response);
                    callback.onSuccess(payment);
                } catch (JSONException e) {
                    callback.onError(new RuntimeException("Failed to parse payment response", e));
                }
            }
            
            @Override
            public void onError(Exception e) {
                callback.onError(e);
            }
        });
    }
    
    /**
     * Ge<response clipped><NOTE>To save on context only part of this file has been shown to you. You should retry this tool after you have searched inside the file with `grep -n` in order to find the line numbers of what you are looking for.</NOTE>
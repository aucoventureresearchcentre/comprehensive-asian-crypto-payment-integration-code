/**
 * Asian Cryptocurrency Payment System - Kiosk SDK
 * Version: 1.0.0
 * 
 * This SDK provides a comprehensive interface for integrating cryptocurrency
 * payments into self-service kiosks across Asian markets, with support for 
 * regulatory compliance in Malaysia, Singapore, Indonesia, Thailand, Brunei, 
 * Cambodia, Vietnam, and Laos.
 */

#ifndef ASIAN_CRYPTO_PAYMENT_H
#define ASIAN_CRYPTO_PAYMENT_H

#include <QObject>
#include <QString>
#include <QVariantMap>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QUrl>
#include <QDateTime>
#include <QCryptographicHash>
#include <QMessageAuthenticationCode>
#include <QUuid>
#include <QTimer>
#include <QPixmap>
#include <QQmlEngine>
#include <QJSEngine>
#include <QDebug>
#include <memory>

namespace AsianCryptoPay {

/**
 * @brief Payment status enumeration
 */
enum class PaymentStatus {
    Created,
    Pending,
    Completed,
    Cancelled,
    Expired
};

/**
 * @brief Country code enumeration
 */
enum class CountryCode {
    Malaysia,    // MY
    Singapore,   // SG
    Indonesia,   // ID
    Thailand,    // TH
    Brunei,      // BN
    Cambodia,    // KH
    Vietnam,     // VN
    Laos         // LA
};

/**
 * @brief Convert CountryCode to string
 * @param code Country code
 * @return Two-letter country code string
 */
inline QString countryCodeToString(CountryCode code) {
    switch (code) {
        case CountryCode::Malaysia: return "MY";
        case CountryCode::Singapore: return "SG";
        case CountryCode::Indonesia: return "ID";
        case CountryCode::Thailand: return "TH";
        case CountryCode::Brunei: return "BN";
        case CountryCode::Cambodia: return "KH";
        case CountryCode::Vietnam: return "VN";
        case CountryCode::Laos: return "LA";
        default: return "MY";
    }
}

/**
 * @brief Convert string to CountryCode
 * @param code Two-letter country code string
 * @return CountryCode enum value
 */
inline CountryCode stringToCountryCode(const QString& code) {
    if (code == "MY") return CountryCode::Malaysia;
    if (code == "SG") return CountryCode::Singapore;
    if (code == "ID") return CountryCode::Indonesia;
    if (code == "TH") return CountryCode::Thailand;
    if (code == "BN") return CountryCode::Brunei;
    if (code == "KH") return CountryCode::Cambodia;
    if (code == "VN") return CountryCode::Vietnam;
    if (code == "LA") return CountryCode::Laos;
    return CountryCode::Malaysia;
}

/**
 * @brief Convert PaymentStatus to string
 * @param status Payment status
 * @return Status string
 */
inline QString paymentStatusToString(PaymentStatus status) {
    switch (status) {
        case PaymentStatus::Created: return "created";
        case PaymentStatus::Pending: return "pending";
        case PaymentStatus::Completed: return "completed";
        case PaymentStatus::Cancelled: return "cancelled";
        case PaymentStatus::Expired: return "expired";
        default: return "unknown";
    }
}

/**
 * @brief Convert string to PaymentStatus
 * @param status Status string
 * @return PaymentStatus enum value
 */
inline PaymentStatus stringToPaymentStatus(const QString& status) {
    if (status == "created") return PaymentStatus::Created;
    if (status == "pending") return PaymentStatus::Pending;
    if (status == "completed") return PaymentStatus::Completed;
    if (status == "cancelled") return PaymentStatus::Cancelled;
    if (status == "expired") return PaymentStatus::Expired;
    return PaymentStatus::Created;
}

/**
 * @brief Payment details class
 */
class PaymentDetails {
public:
    /**
     * @brief Constructor
     */
    PaymentDetails() {}
    
    /**
     * @brief Set payment amount
     * @param amount Payment amount
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setAmount(double amount) {
        m_amount = amount;
        return *this;
    }
    
    /**
     * @brief Set fiat currency
     * @param currency Fiat currency code (e.g., MYR, SGD)
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCurrency(const QString& currency) {
        m_currency = currency;
        return *this;
    }
    
    /**
     * @brief Set cryptocurrency
     * @param cryptoCurrency Cryptocurrency code (e.g., BTC, ETH)
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCryptoCurrency(const QString& cryptoCurrency) {
        m_cryptoCurrency = cryptoCurrency;
        return *this;
    }
    
    /**
     * @brief Set payment description
     * @param description Payment description
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setDescription(const QString& description) {
        m_description = description;
        return *this;
    }
    
    /**
     * @brief Set merchant order ID
     * @param orderId Merchant order ID
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setOrderId(const QString& orderId) {
        m_orderId = orderId;
        return *this;
    }
    
    /**
     * @brief Set customer email
     * @param email Customer email
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCustomerEmail(const QString& email) {
        m_customerEmail = email;
        return *this;
    }
    
    /**
     * @brief Set customer name
     * @param name Customer name
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCustomerName(const QString& name) {
        m_customerName = name;
        return *this;
    }
    
    /**
     * @brief Set callback URL
     * @param url Callback URL for payment updates
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCallbackUrl(const QString& url) {
        m_callbackUrl = url;
        return *this;
    }
    
    /**
     * @brief Set success URL
     * @param url Redirect URL on successful payment
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setSuccessUrl(const QString& url) {
        m_successUrl = url;
        return *this;
    }
    
    /**
     * @brief Set cancel URL
     * @param url Redirect URL on cancelled payment
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setCancelUrl(const QString& url) {
        m_cancelUrl = url;
        return *this;
    }
    
    /**
     * @brief Set metadata
     * @param metadata Additional metadata
     * @return Reference to this object for method chaining
     */
    PaymentDetails& setMetadata(const QVariantMap& metadata) {
        m_metadata = metadata;
        return *this;
    }
    
    /**
     * @brief Get payment amount
     * @return Payment amount
     */
    double amount() const { return m_amount; }
    
    /**
     * @brief Get fiat currency
     * @return Fiat currency code
     */
    QString currency() const { return m_currency; }
    
    /**
     * @brief Get cryptocurrency
     * @return Cryptocurrency code
     */
    QString cryptoCurrency() const { return m_cryptoCurrency; }
    
    /**
     * @brief Get payment description
     * @return Payment description
     */
    QString description() const { return m_description; }
    
    /**
     * @brief Get merchant order ID
     * @return Merchant order ID
     */
    QString orderId() const { return m_orderId; }
    
    /**
     * @brief Get customer email
     * @return Customer email
     */
    QString customerEmail() const { return m_customerEmail; }
    
    /**
     * @brief Get customer name
     * @return Customer name
     */
    QString customerName() const { return m_customerName; }
    
    /**
     * @brief Get callback URL
     * @return Callback URL
     */
    QString callbackUrl() const { return m_callbackUrl; }
    
    /**
     * @brief Get success URL
     * @return Success URL
     */
    QString successUrl() const { return m_successUrl; }
    
    /**
     * @brief Get cancel URL
     * @return Cancel URL
     */
    QString cancelUrl() const { return m_cancelUrl; }
    
    /**
     * @brief Get metadata
     * @return Metadata
     */
    QVariantMap metadata() const { return m_metadata; }
    
    /**
     * @brief Convert to JSON object
     * @return JSON object
     */
    QJsonObject toJson() const {
        QJsonObject json;
        json["amount"] = QString::number(m_amount, 'f', 8);
        json["currency"] = m_currency;
        json["crypto_currency"] = m_cryptoCurrency;
        json["description"] = m_description;
        
        if (!m_orderId.isEmpty()) {
            json["order_id"] = m_orderId;
        }
        
        if (!m_customerEmail.isEmpty()) {
            json["customer_email"] = m_customerEmail;
        }
        
        if (!m_customerName.isEmpty()) {
            json["customer_name"] = m_customerName;
        }
        
        if (!m_callbackUrl.isEmpty()) {
            json["callback_url"] = m_callbackUrl;
        }
        
        if (!m_successUrl.isEmpty()) {
            json["success_url"] = m_successUrl;
        }
        
        if (!m_cancelUrl.isEmpty()) {
            json["cancel_url"] = m_cancelUrl;
        }
        
        if (!m_metadata.isEmpty()) {
            json["metadata"] = QJsonObject::fromVariantMap(m_metadata);
        }
        
        return json;
    }
    
private:
    double m_amount = 0.0;
    QString m_currency;
    QString m_cryptoCurrency;
    QString m_description;
    QString m_orderId;
    QString m_customerEmail;
    QString m_customerName;
    QString m_callbackUrl;
    QString m_successUrl;
    QString m_cancelUrl;
    QVariantMap m_metadata;
};

/**
 * @brief Payment class
 */
class Payment {
public:
    /**
     * @brief Constructor
     */
    Payment() {}
    
    /**
     * @brief Create from JSON object
     * @param json JSON object
     * @return Payment object
     */
    static Payment fromJson(const QJsonObject& json) {
        Payment payment;
        
        payment.m_id = json["id"].toString();
        payment.m_merchantId = json["merchant_id"].toString();
        payment.m_amount = json["amount"].toString().toDouble();
        payment.m_currency = json["currency"].toString();
        payment.m_cryptoAmount = json["crypto_amount"].toString().toDouble();
        payment.m_cryptoCurrency = json["crypto_currency"].toString();
        payment.m_description = json["description"].toString();
        payment.m_orderId = json["order_id"].toString();
        payment.m_customerEmail = json["customer_email"].toString();
        payment.m_customerName = json["customer_name"].toString();
        payment.m_address = json["address"].toString();
        payment.m_qrCodeUrl = json["qr_code_url"].toString();
        payment.m_status = stringToPaymentStatus(json["status"].toString());
        
        payment.m_createdAt = QDateTime::fromString(json["created_at"].toString(), Qt::ISODate);
        payment.m_updatedAt = QDateTime::fromString(json["updated_at"].toString(), Qt::ISODate);
        payment.m_expiresAt = QDateTime::fromString(json["expires_at"].toString(), Qt::ISODate);
        
        if (json.contains("metadata") && json["metadata"].isObject()) {
            payment.m_metadata = json["metadata"].toObject().toVariantMap();
        }
        
        return payment;
    }
    
    /**
     * @brief Get payment ID
     * @return Payment ID
     */
    QString id() const { return m_id; }
    
    /**
     * @brief Get merchant ID
     * @return Merchant ID
     */
    QString merchantId() const { return m_merchantId; }
    
    /**
     * @brief Get payment amount
     * @return Payment amount
     */
    double amount() const { return m_amount; }
    
    /**
     * @brief Get fiat currency
     * @return Fiat currency code
     */
    QString currency() const { return m_currency; }
    
    /**
     * @brief Get cryptocurrency amount
     * @return Cryptocurrency amount
     */
    double cryptoAmount() const { return m_cryptoAmount; }
    
    /**
     * @brief Get cryptocurrency
     * @return Cryptocurrency code
     */
    QString cryptoCurrency() const { return m_cryptoCurrency; }
    
    /**
     * @brief Get payment description
     * @return Payment description
     */
    QString description() const { return m_description; }
    
    /**
     * @brief Get merchant order ID
     * @return Merchant order ID
     */
    QString orderId() const { return m_orderId; }
    
    /**
     * @brief Get customer email
     * @return Customer email
     */
    QString customerEmail() const { return m_customerEmail; }
    
    /**
     * @brief Get customer name
     * @return Customer name
     */
    QString customerName() const { return m_customerName; }
    
    /**
     * @brief Get cryptocurrency address
     * @return Cryptocurrency address
     */
    QString address() const { return m_address; }
    
    /**
     * @brief Get QR code URL
     * @return QR code URL
     */
    QString qrCodeUrl() const { return m_qrCodeUrl; }
    
    /**
     * @brief Get payment status
     * @return Payment status
     */
    PaymentStatus status() const { return m_status; }
    
    /**
     * @brief Get payment status as string
     * @return Payment status string
     */
    QString statusString() const { return paymentStatusToString(m_status); }
    
    /**
     * @brief Get creation time
     * @return Creation time
     */
    QDateTime createdAt() const { return m_createdAt; }
    
    /**
     * @brief Get last update time
     * @return Last update time
     */
    QDateTime updatedAt() const { return m_updatedAt; }
    
    /**
     * @brief Get expiration time
     * @return Expiration time
     */
    QDateTime expiresAt() const { return m_expiresAt; }
    
    /**
     * @brief Get metadata
     * @return Metadata
     */
    QVariantMap metadata() const { return m_metadata; }
    
    /**
     * @brief Check if payment is completed
     * @return Whether payment is completed
     */
    bool isCompleted() const { return m_status == PaymentStatus::Completed; }
    
    /**
     * @brief Check if payment is pending
     * @return Whether payment is pending
     */
    bool isPending() const { return m_status == PaymentStatus::Pending; }
    
    /**
     * @brief Check if payment is expired
     * @return Whether payment is expired
     */
    bool isExpired() const { return m_status == PaymentStatus::Expired; }
    
    /**
     * @brief Check if payment is cancelled
     * @return Whether payment is cancelled
     */
    bool isCancelled() const { return m_status == PaymentStatus::Cancelled; }
    
    /**
     * @brief Convert to JSON object
     * @return JSON object
     */
    QJsonObject toJson() const {
        QJsonObject json;
        json["id"] = m_id;
        json["merchant_id"] = m_merchantId;
        json["amount"] = QString::number(m_amount, 'f', 8);
        json["currency"] = m_currency;
        json["crypto_amount"] = QString::number(m_cryptoAmount, 'f', 8);
        json["crypto_currency"] = m_cryptoCurrency;
        json["description"] = m_description;
        json["order_id"] = m_orderId;
        json["customer_email"] = m_customerEmail;
        json["customer_name"] = m_customerName;
        json["address"] = m_address;
        json["qr_code_url"] = m_qrCodeUrl;
        json["status"] = paymentStatusToString(m_status);
        json["created_at"] = m_createdAt.toString(Qt::ISODate);
        json["updated_at"] = m_updatedAt.toString(Qt::ISODate);
        json["expires_at"] = m_expiresAt.toString(Qt::ISODate);
        
        if (!m_metadata.isEmpty()) {
            json["metadata"] = QJsonObject::fromVariantMap(m_metadata);
        }
        
        return json;
    }
    
private:
    QString m_id;
    QString m_merchantId;
    double m_amount = 0.0;
    QString m_currency;
    double m_cryptoAmount = 0.0;
    QString m_cryptoCurre<response clipped><NOTE>To save on context only part of this file has been shown to you. You should retry this tool after you have searched inside the file with `grep -n` in order to find the line numbers of what you are looking for.</NOTE>
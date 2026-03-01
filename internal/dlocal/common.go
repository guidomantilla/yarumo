package dlocal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var singleton atomic.Value

func instance() string {
	value := singleton.Load()
	if cutils.IsEmpty(value) {
		return load()
	}
	return value.(string)
}

func load() string {
	url := "https://api.dlocal.com"
	if cutils.In(os.Getenv("DLOCAL_ENV"), "", "sandbox") {
		url = "https://sandbox.dlocal.com"
	}
	singleton.Store(url)
	return url
}

//

func DefaultConfigOptions() chttp.Options {
	return chttp.OptionsBuilder().
		WithHeaderHandler(HeaderHandler).
		WithOkResponsesWithBody([]int{http.StatusOK}).
		WithErrorResponsesWithBody([]int{http.StatusForbidden, http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusGatewayTimeout}).
		Build()
}

// Sign generates a signature for the request using V2-HMAC-SHA256
func Sign(xLogin string, xDate string, secretKey string, body string) string {

	var buffer bytes.Buffer
	buffer.WriteString(xLogin)
	buffer.WriteString(xDate)
	buffer.WriteString(body)

	h := hmac.New(sha256.New, []byte(secretKey))
	if _, err := h.Write(buffer.Bytes()); err != nil {
		return err.Error()
	}

	signature := h.Sum(nil)

	return fmt.Sprintf("V2-HMAC-SHA256, Signature: %s", hex.EncodeToString(signature))
}

// HeaderHandler generates the headers for the request
func HeaderHandler(requestId string, body string) http.Header {

	xLogin, xDate, secretKey, xTransactionKey := os.Getenv("DLOCAL_X_LOGIN"), cutils.ToISO8601(time.Now()), os.Getenv("DLOCAL_SECRET_KEY"), os.Getenv("DLOCAL_X_TRANSACTION_KEY")
	assert.NotEmpty(xLogin, "DLOCAL_X_LOGIN is empty")
	assert.NotEmpty(secretKey, "DLOCAL_SECRET_KEY is empty")
	assert.NotEmpty(xTransactionKey, "DLOCAL_X_TRANSACTION_KEY is empty")

	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("User-Agent", "MerchantTest / 1.0")
	header.Add("X-Version", "2.1")
	header.Add("X-Date", xDate)
	header.Add("X-Login", xLogin)
	header.Add("X-Trans-Key", xTransactionKey)
	header.Add("X-Idempotency-Key", requestId)
	header.Add("Authorization", Sign(xLogin, xDate, secretKey, body))
	xPaymentSource := os.Getenv("DLOCAL_PAYMENT_SOURCE")
	if xPaymentSource != "" {
		header.Add("X-Payment-Source", xPaymentSource)
	}
	return header
}

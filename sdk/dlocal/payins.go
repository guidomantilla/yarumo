package dlocal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xorcare/pointer"

	chttp "github.com/guidomantilla/go-feather-lib/pkg/common/http"
	cutils "github.com/guidomantilla/go-feather-lib/pkg/common/utils"
)

type payins struct {
	url     string
	options []chttp.Options
}

func NewPayIns(options ...chttp.Options) PayIns {
	return &payins{
		url:     instance(),
		options: append([]chttp.Options{DefaultConfigOptions()}, options...),
	}
}

func (p *payins) opts(options ...chttp.Options) []chttp.Options {
	opts := p.options
	if !cutils.IsEmpty(options) {
		opts = append(opts, options...)
	}
	return opts
}

//PaymentMethods

func (p *payins) SearchPaymentMethodsByCountry(ctx context.Context, requestId string, country string, options ...chttp.Options) ([]PaymentMethod, error) {

	var err error
	var response *[]PaymentMethod
	if response, err = chttp.Call[[]PaymentMethod](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("payments-methods?country=%s", country), nil, p.opts(options...)...); err != nil {
		return nil, err
	}

	return *response, nil
}

//Payments

func (p *payins) CreatePayment(ctx context.Context, requestId string, payment *Payment, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodPost, "secure_payments", payment, p.opts(options...)...)
}

func (p *payins) GetPayment(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("payments/%s", paymentId), nil, p.opts(options...)...)
}

func (p *payins) GetPaymentStatus(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("payments/%s/status", paymentId), nil, p.opts(options...)...)
}

func (p *payins) CancelPayment(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodPost, fmt.Sprintf("payments/%s/cancel", paymentId), nil, p.opts(options...)...)
}

//SavingCards

func (p *payins) CreateCard(ctx context.Context, requestId string, card *SecureCard, options ...chttp.Options) (*Card, error) {
	return chttp.Call[Card](ctx, requestId, p.url, http.MethodPost, "secure_cards", card, p.opts(options...)...)
}

func (p *payins) RetrieveCard(ctx context.Context, requestId string, cardId string, options ...chttp.Options) (*Card, error) {
	return chttp.Call[Card](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("cards/%s", cardId), nil, p.opts(options...)...)
}

func (p *payins) DeleteCard(ctx context.Context, requestId string, cardId string, options ...chttp.Options) (*Card, error) {
	return chttp.Call[Card](ctx, requestId, p.url, http.MethodDelete, fmt.Sprintf("secure_cards/%s", cardId), nil, p.opts(options...)...)
}

//Authorizations

func (p *payins) CreateAuthorization(ctx context.Context, requestId string, payment *Payment, options ...chttp.Options) (*Payment, error) {

	if payment != nil && payment.Card != nil {
		payment.Card.Capture = pointer.Bool(false)
	}

	response, err := p.CreatePayment(ctx, requestId, payment, options...)
	if err != nil {
		return nil, err
	}

	response.AuthorizationId = response.Id
	return response, nil
}

func (p *payins) CaptureAuthorization(ctx context.Context, requestId string, capture *Payment, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodPost, "payments", capture, p.opts(options...)...)
}

func (p *payins) CancelAuthorization(ctx context.Context, requestId string, authorizationId string, options ...chttp.Options) (*Payment, error) {
	return chttp.Call[Payment](ctx, requestId, p.url, http.MethodPost, fmt.Sprintf("payments/%s/cancel", authorizationId), nil, p.opts(options...)...)
}

// Installments

func (p *payins) CreateInstallmentPlan(ctx context.Context, requestId string, installmentPlan *InstallmentPlan, options ...chttp.Options) (*InstallmentPlan, error) {
	return chttp.Call[InstallmentPlan](ctx, requestId, p.url, http.MethodPost, "installments-plans", installmentPlan, p.opts(options...)...)
}

//VirtualAccounts

func (p *payins) CreateUniqueReference(ctx context.Context, requestId string, uniqueReference *UniqueReference, options ...chttp.Options) (*UniqueReference, error) {
	return chttp.Call[UniqueReference](ctx, requestId, p.url, http.MethodPost, "virtual-accounts", uniqueReference, p.opts(options...)...)
}

//Refunds

func (p *payins) MakeRefund(ctx context.Context, requestId string, refund *Refund, options ...chttp.Options) (*Refund, error) {
	return chttp.Call[Refund](ctx, requestId, p.url, http.MethodPost, "refunds", refund, p.opts(options...)...)
}

func (p *payins) RetrieveRefund(ctx context.Context, requestId string, refundId string, options ...chttp.Options) (*Refund, error) {
	return chttp.Call[Refund](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("refunds/%s", refundId), nil, p.opts(options...)...)
}

func (p *payins) RetrieveRefundOrder(ctx context.Context, requestId string, refundOrderId string, options ...chttp.Options) (*Refund, error) {
	return chttp.Call[Refund](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("orders/refunds/%s", refundOrderId), nil, p.opts(options...)...)
}

func (p *payins) CheckRefundStatus(ctx context.Context, requestId string, refundId string, options ...chttp.Options) (*Refund, error) {
	return chttp.Call[Refund](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("refunds/%s/status", refundId), nil, p.opts(options...)...)
}

//Chargebacks

func (p *payins) RetrieveChargeback(ctx context.Context, requestId string, chargebackId string, options ...chttp.Options) (*Notification, error) {
	return chttp.Call[Notification](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("chargebacks/%s", chargebackId), nil, p.opts(options...)...)
}

func (p *payins) RetrieveChargebackStatus(ctx context.Context, requestId string, chargebackId string, options ...chttp.Options) (*Notification, error) {
	return chttp.Call[Notification](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("chargebacks/%s/status", chargebackId), nil, p.opts(options...)...)
}

//Orders

func (p *payins) RetrieveOrder(ctx context.Context, requestId string, orderId string, options ...chttp.Options) (*Order, error) {
	return chttp.Call[Order](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("orders/%s", orderId), nil, p.opts(options...)...)
}

//ExchangeRates

func (p *payins) RetrieveExchangeRates(ctx context.Context, requestId string, from string, to string, options ...chttp.Options) (*CurrencyExchange, error) {
	return chttp.Call[CurrencyExchange](ctx, requestId, p.url, http.MethodGet, fmt.Sprintf("currency-exchanges?from=%s&to=%s", from, to), nil, p.opts(options...)...)
}

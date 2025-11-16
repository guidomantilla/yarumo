package dlocal

import (
	"context"

	chttp "github.com/guidomantilla/go-feather-lib/pkg/common/http"
)

type PaymentMethods interface {
	SearchPaymentMethodsByCountry(ctx context.Context, requestId string, country string, options ...chttp.Options) ([]PaymentMethod, error)
}

type Payments interface {
	CreatePayment(ctx context.Context, requestId string, payment *Payment, options ...chttp.Options) (*Payment, error)
	GetPayment(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error)
	GetPaymentStatus(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error)
	CancelPayment(ctx context.Context, requestId string, paymentId string, options ...chttp.Options) (*Payment, error)
}

type SavingCards interface {
	CreateCard(ctx context.Context, requestId string, card *SecureCard, options ...chttp.Options) (*Card, error)
	RetrieveCard(ctx context.Context, requestId string, cardId string, options ...chttp.Options) (*Card, error)
	DeleteCard(ctx context.Context, requestId string, cardId string, options ...chttp.Options) (*Card, error)
}

type Authorizations interface {
	CreateAuthorization(ctx context.Context, requestId string, payment *Payment, options ...chttp.Options) (*Payment, error)
	CaptureAuthorization(ctx context.Context, requestId string, capture *Payment, options ...chttp.Options) (*Payment, error)
	CancelAuthorization(ctx context.Context, requestId string, authorizationId string, options ...chttp.Options) (*Payment, error)
}

type Installments interface {
	CreateInstallmentPlan(ctx context.Context, requestId string, installmentPlan *InstallmentPlan, options ...chttp.Options) (*InstallmentPlan, error)
}

type VirtualAccounts interface {
	CreateUniqueReference(ctx context.Context, requestId string, uniqueReference *UniqueReference, options ...chttp.Options) (*UniqueReference, error)
}

type Refunds interface {
	MakeRefund(ctx context.Context, requestId string, refund *Refund, options ...chttp.Options) (*Refund, error)
	RetrieveRefund(ctx context.Context, requestId string, refundId string, options ...chttp.Options) (*Refund, error)
	RetrieveRefundOrder(ctx context.Context, requestId string, refundOrderId string, options ...chttp.Options) (*Refund, error)
	CheckRefundStatus(ctx context.Context, requestId string, refundId string, options ...chttp.Options) (*Refund, error)
}

type Chargebacks interface {
	RetrieveChargeback(ctx context.Context, requestId string, chargebackId string, options ...chttp.Options) (*Notification, error)
	RetrieveChargebackStatus(ctx context.Context, requestId string, chargebackId string, options ...chttp.Options) (*Notification, error)
}

type Orders interface {
	RetrieveOrder(ctx context.Context, requestId string, orderId string, options ...chttp.Options) (*Order, error)
}

type ExchangeRates interface {
	RetrieveExchangeRates(ctx context.Context, requestId string, from string, to string, options ...chttp.Options) (*CurrencyExchange, error)
}

type PayIns interface {
	PaymentMethods
	Payments
	SavingCards
	Authorizations
	Installments
	VirtualAccounts
	Refunds
	Chargebacks
	Orders
	ExchangeRates
}

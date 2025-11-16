package dlocal

type AllowedFlow string

const (
	AllowedFlowDirect   AllowedFlow = "DIRECT"
	AllowedFlowRedirect AllowedFlow = "REDIRECT"
)

type PaymentMethodType string

const (
	PaymentMethodTypeCard         PaymentMethodType = "CARD"
	PaymentMethodTypeTicket       PaymentMethodType = "TICKET"
	PaymentMethodTypeWallet       PaymentMethodType = "WALLET"
	PaymentMethodTypeDirectDebit  PaymentMethodType = "DIRECT_DEBIT"
	PaymentMethodTypeBankTransfer PaymentMethodType = "BANK_TRANSFER"
)

type StoredCredentialType string

const (
	StoredCredentialTypeCardOnFile      StoredCredentialType = "CARD_ON_FILE"
	StoredCredentialTypeSubscription    StoredCredentialType = "SUBSCRIPTION"
	StoredCredentialTypeUnscheduled     StoredCredentialType = "UNSCHEDULED_CARD_ON_FILE"
	StoredCredentialTypeInstallments    StoredCredentialType = "INSTALLMENTS"
	StoredCredentialTypeNoShow          StoredCredentialType = "NO_SHOW"
	StoredCredentialTypeDelayedCharges  StoredCredentialType = "DELAYED_CHARGES"
	StoredCredentialTypeReauthorization StoredCredentialType = "REAUTHORIZATION"
	StoredCredentialTypeResubmission    StoredCredentialType = "RESUBMISSION"
)

type StoredCredentialUsage string

const (
	StoredCredentialUsageFirst StoredCredentialUsage = "FIRST"
	StoredCredentialUsageUsed  StoredCredentialUsage = "USED"
)

type TicketType string

const (
	TicketNumeric TicketType = "NUMERIC"
	TicketBarcode TicketType = "BARCODE"
	TicketCustom  TicketType = "CUSTOM"
	TicketRefCode TicketType = "REFERENCE_CODE"
)

type BankAccountType string

const (
	BankAccountTypeChecking BankAccountType = "CHECKING"
	BankAccountTypeSaving   BankAccountType = "SAVING"
)

type BankAccountType2 string

const (
	BankAccountType2CurrentAccounts       BankAccountType2 = "C"
	BankAccountType2SavingsAccounts       BankAccountType2 = "S"
	BankAccountType2InternationalAccounts BankAccountType2 = "I"
)

type ThreeDSecureVersion string

const (
	ThreeDSecureVersion10  ThreeDSecureVersion = "1.0"
	ThreeDSecureVersion20  ThreeDSecureVersion = "2.0"
	ThreeDSecureVersion210 ThreeDSecureVersion = "2.1.0"
	ThreeDSecureVersion220 ThreeDSecureVersion = "2.2.0"
)

type ECI string

const (
	ECI00 ECI = "00"
	ECI01 ECI = "01"
	ECI02 ECI = "02"
	ECI03 ECI = "03"
	ECI04 ECI = "04"
	ECI05 ECI = "05"
	ECI06 ECI = "06"
	ECI07 ECI = "07"
)

type EnrollmentResponse string

const (
	EnrollmentResponseAuthenticationAvailable             EnrollmentResponse = "Y"
	EnrollmentResponseCardholderNotParticipating          EnrollmentResponse = "N"
	EnrollmentResponseUnableToAuthenticateCardNotEligible EnrollmentResponse = "U"
)

type AuthenticationResponse string

const (
	AuthenticationResponseAuthenticationSuccessful          AuthenticationResponse = "Y"
	AuthenticationResponseAttemptsProcessingPerformed       AuthenticationResponse = "A"
	AuthenticationResponseAuthenticationFailed              AuthenticationResponse = "N"
	AuthenticationResponseAuthenticationCouldNotBePerformed AuthenticationResponse = "U"
)

type PaymentStatusCode string

const (
	PaymentStatusCodePending                                   PaymentStatusCode = "100"
	PaymentStatusCodePending3DSecure                           PaymentStatusCode = "101"
	PaymentStatusCodePaid                                      PaymentStatusCode = "200"
	PaymentStatusCodeRejected                                  PaymentStatusCode = "300"
	PaymentStatusCodeRejectedByBank                            PaymentStatusCode = "301"
	PaymentStatusCodeRejectedInsufficientAmount                PaymentStatusCode = "302"
	PaymentStatusCodeRejectedCardBlacklisted                   PaymentStatusCode = "303"
	PaymentStatusCodeRejectedScoreValidation                   PaymentStatusCode = "304"
	PaymentStatusCodeRejectedMaxAttemptsReached                PaymentStatusCode = "305"
	PaymentStatusCodeRejectedCallBankForAuthorize              PaymentStatusCode = "306"
	PaymentStatusCodeRejectedDuplicatedPayment                 PaymentStatusCode = "307"
	PaymentStatusCodeRejectedCardDisabled                      PaymentStatusCode = "308"
	PaymentStatusCodeRejectedCardExpired                       PaymentStatusCode = "309"
	PaymentStatusCodeRejectedCardReportedLost                  PaymentStatusCode = "310"
	PaymentStatusCodeRejectedCardRequestedByBank               PaymentStatusCode = "311"
	PaymentStatusCodeRejectedCardRestrictedByBank              PaymentStatusCode = "312"
	PaymentStatusCodeRejectedCardReportedStolen                PaymentStatusCode = "313"
	PaymentStatusCodeRejectedInvalidCardNumber                 PaymentStatusCode = "314"
	PaymentStatusCodeRejectedInvalidSecurityCode               PaymentStatusCode = "315"
	PaymentStatusCodeRejectedUnsupportedOperation              PaymentStatusCode = "316"
	PaymentStatusCodeRejectedDueHighRisk                       PaymentStatusCode = "317"
	PaymentStatusCodeRejectedInvalidTransaction                PaymentStatusCode = "318"
	PaymentStatusCodeRejectedAmountExceeded                    PaymentStatusCode = "319"
	PaymentStatusCodeRejected3DSecureRequired                  PaymentStatusCode = "320"
	PaymentStatusCodeRejectedErrorInAcquirer                   PaymentStatusCode = "321"
	PaymentStatusCodeRejectedInvalidUserAccount                PaymentStatusCode = "323"
	PaymentStatusCodeRejectedInvalidPin                        PaymentStatusCode = "324"
	PaymentStatusCodeRejectedTransactionFrequencyLimitExceeded PaymentStatusCode = "325"
	PaymentStatusCodeRejectedUserAccountDisabledOrExpired      PaymentStatusCode = "326"
	PaymentStatusCodeRejectedErrorInExternalNetwork            PaymentStatusCode = "327"
	PaymentStatusCodeRejectedInvalidPayerToken                 PaymentStatusCode = "328"
	PaymentStatusCodeRejectedCardNotEnrolledFor3DSecure        PaymentStatusCode = "330"
	PaymentStatusCodeRejectedUserCancelledPayment              PaymentStatusCode = "331"
	PaymentStatusCodeRejectedExpiredPayment                    PaymentStatusCode = "340"
	PaymentStatusCodeRejected3DSecureChallengeNotCompleted     PaymentStatusCode = "341"
	PaymentStatusCodeCancelled                                 PaymentStatusCode = "400"
	PaymentStatusCodeAuthorized                                PaymentStatusCode = "600"
	PaymentStatusCodeVerified                                  PaymentStatusCode = "700"
)

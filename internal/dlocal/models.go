package dlocal

// PaymentMethod Object that identifies each payment method accepted by dLocal.
type PaymentMethod struct {
	Id           *string            `json:"id,omitempty"`      // Payment method ID
	Name         *string            `json:"name,omitempty"`    // Payment type name
	Country      *[]string          `json:"country,omitempty"` // Countries where the payment method is available
	Logo         *string            `json:"logo,omitempty"`    // Payment method image URL
	Type         *PaymentMethodType `json:"type,omitempty"`
	AllowedFlows *[]AllowedFlow     `json:"allowed_flows,omitempty"`
}

// Address Object that identifies the address of the payer.
type Address struct {
	State   *string `json:"state,omitempty"`    // length:"10"  User's address state. Required in South Africa
	City    *string `json:"city,omitempty"`     // length:"100" User’s address city. Required in India and South Africa
	ZipCode *string `json:"zip_code,omitempty"` // length:"10"  User’s address ZIP Code. Required in India and South Africa
	Street  *string `json:"street,omitempty"`   // length:"100" User’s address street. Required in India and South Africa
	Number  *string `json:"number,omitempty"`   // length:"20"  User’s address number. Required in India and South Africa
}

// Payer Object that identifies the payer.
type Payer struct {
	Name          *string  `json:"name,omitempty"`           // length:"100" User's full name
	Email         *string  `json:"email,omitempty"`          // length:"100" User’s email address
	BirthDate     *string  `json:"birth_date,omitempty"`     // length:"10"  User’s birthdate (DD-MM-YYYY)
	Phone         *string  `json:"phone,omitempty"`          // length:"20"  User’s phone. Mandatory in India. Also required for fraud prevention
	Document      *string  `json:"document,omitempty"`       // length:"30"  User’s personal identification number. To see the document code list per country, go to the Country Reference page
	Document2     *string  `json:"document2,omitempty"`      // length:"100" Additional personal identification
	UserReference *string  `json:"user_reference,omitempty"` // length:"125" Unique user ID at the merchant side. Required for fraud prevention
	Ip            *string  `json:"ip,omitempty"`             // length:"39"  User's IP address. Required for fraud prevention
	DeviceId      *string  `json:"device_id,omitempty"`      // length:"25"  User's unique device identifier, for information on how to obtain the device_id see the Device ID documentation. Required for fraud prevention
	Address       *Address `json:"address,omitempty"`        // 			 User's address. Required in India and South Africa for fraud prevention
}

// Card Object that identifies the card.
//
// If you are making a payment with credit card information, you need to use the following endpoint instead:
// https://api.dlocal.com/secure_payments
//
// Card payments with a card_id or token should use the endpoint: https://api.dlocal.com/payments
type Card struct {
	HolderName            *string                `json:"holder_name,omitempty"`             // Required if token or card_id not present
	ExpireMonth           *int                   `json:"expiration_month,omitempty"`        // Required if token or card_id not present
	ExpirationYear        *int                   `json:"expiration_year,omitempty"`         // Required if token or card_id not present
	Number                *string                `json:"number,omitempty"`                  // Required if encrypted_data, token or card_id not present
	CCV                   *string                `json:"cvv,omitempty"`                     // Optional. Required for India
	EncryptedData         *string                `json:"encrypted_data,omitempty"`          // JWE encrypted params
	Token                 *string                `json:"token,omitempty"`                   // Temporary credit card token securely created using Smart Fields
	CVVToken              *string                `json:"cvv_token,omitempty"`               // Temporary CVV token securely created using the CVV-only Smart Field
	CardId                *string                `json:"card_id,omitempty"`                 // Credit card ID returned by the Create a Card endpoint
	Installments          *int                   `json:"installments,omitempty"`            // Number of installments. Default 1
	InstallmentsID        *string                `json:"installments_id,omitempty"`         // Installments ID of an installment plan
	Descriptor            *string                `json:"descriptor,omitempty"`              // Dynamic Descriptor
	Verify                *bool                  `json:"verify,omitempty"`                  // Validates the user’s card without initiating a payment transaction. To enable this functionality, set verify=true and amount=0
	Capture               *bool                  `json:"capture,omitempty"`                 // Whether or not to immediately capture the charge. When false, the charge issues an authorization and will need to be captured later. Default TRUE
	Save                  *bool                  `json:"save,omitempty"`                    // Whether or not to save the card for future payments. The response will include a card_id
	StoredCredentialType  *StoredCredentialType  `json:"stored_credential_type,omitempty"`  // Use this field to store credentials for future payments. See all the information in the Merchant Initiated Transactions page
	StoredCredentialUsage *StoredCredentialUsage `json:"stored_credential_usage,omitempty"` // Indicates if this is the first time the token is used (an initial payment) or if the token has already been used for a previous payment (subsequent payment). See all the information in the Merchant Initiated Transactions page
	Brand                 *string                `json:"brand,omitempty"`                   // Card brand
	Last4                 *string                `json:"last4,omitempty"`                   // Last 4 digits of the card
	Country               *string                `json:"country,omitempty"`
	Deleted               *bool                  `json:"deleted,omitempty"`
}

// Installment Object that identifies the installment.
type Installment struct {
	Id                *string  `json:"id,omitempty"`                 // Installment ID
	InstallmentAmount *float64 `json:"installment_amount,omitempty"` // Installment amount. Includes interests associated to the installment
	TotalAmount       *float64 `json:"total_amount,omitempty"`       // Installments total amount. Includes interests associated to the installment
	Installments      *int     `json:"installments,omitempty"`       // Number of installments
}

// InstallmentPlan Object that identifies the installment plan.
type InstallmentPlan struct {
	Id                 *string        `json:"id,omitempty"`                   // The installments plan ID
	Country            *string        `json:"country,omitempty"`              // The country of the installments plan
	Currency           *string        `json:"currency,omitempty"`             // The currency code
	Bin                *string        `json:"bin,omitempty"`                  // The credit card bin
	Amount             *float64       `json:"amount,omitempty"`               // The amount of the installments plan
	Installments       *[]Installment `json:"installments,omitempty"`         // The installments plan information
	InstallmentsByBank *bool          `json:"installments_by_bank,omitempty"` // If false: the installment interest is known beforehand and can be shown to the buyer. If true: The installment interest is not known beforehand and will be determined by the issuer
}

// Ticket Object that identifies the ticket.
type Ticket struct {
	Type           *TicketType `json:"type,omitempty"`            // Type of ticket
	Number         *string     `json:"number,omitempty"`          // Numeric code of the NUMERIC or CUSTOM ticket
	Barcode        *string     `json:"barcode,omitempty"`         // Code to be included in the barcode of the BARCODE or CUSTOM ticket
	Format         *int        `json:"format,omitempty"`          // Format of the barcode of the BARCODE or CUSTOM ticket. Example CODE_128, or ITF
	Id             *string     `json:"id,omitempty"`              // Reference code of the ticket
	ExpirationDate *string     `json:"expiration_date,omitempty"` // The expiration date of the ticket. ISO-8601
	CompanyName    *string     `json:"company_name,omitempty"`    // Name of the company that acts as the beneficiary of the payment
	CompanyId      *string     `json:"company_id,omitempty"`      // Identifier of the company
	ProviderName   *string     `json:"provider_name,omitempty"`   // Name of the company/bank that is creating the ticket
	ProviderLogo   *string     `json:"provider_logo,omitempty"`   // URL of the logo of the company/bank that is creating the ticket
	ImageUrl       *string     `json:"image_url,omitempty"`       // URL of the full version of the ticket
}

// BankTransfer Object that identifies the bank transfer.
type BankTransfer struct {
	BankAccountType         *BankAccountType `json:"bank_account_type,omitempty"`         // Type of ticket, can be CHECKING or SAVING
	BankName                *string          `json:"bank_name,omitempty"`                 // Name of the bank
	BankCode                *string          `json:"bank_code,omitempty"`                 // Code of the bank
	BeneficiaryName         *string          `json:"beneficiary_name,omitempty"`          // Name of the account holder
	BankAccount             *string          `json:"bank_account,omitempty"`              // Bank account number
	BankAccountLabel        *string          `json:"bank_account_label,omitempty"`        // Label to be displayed related to bank_account
	BankAccount2            *string          `json:"bank_account2,omitempty"`             // Secondary bank account number
	BankAccount2Label       *string          `json:"bank_account2_label,omitempty"`       // Label to be displayed related to bank_account2
	BeneficiaryDocumentType *string          `json:"beneficiary_document_type,omitempty"` // Type of document of the account holder
	Reference               *string          `json:"reference,omitempty"`                 // Reference code for the payer to add on payment
	RedirectUrl             *string          `json:"redirect_url,omitempty"`              // URL of the full version of the ticket. In case you want to redirect
	UserPaymentAmount       *float64         `json:"user_payment_amount,omitempty"`       // Amount the user needs to pay
	PaymentInstruction      *string          `json:"payment_instruction,omitempty"`       // Instructions for making the payment
}

// Wallet Object that identifies the wallet.
type Wallet struct {
	Name       *string `json:"name,omitempty"`           // Name of wallet
	Save       *bool   `json:"save,omitempty"`           // Whether or not a token will be included in the response in order to make future payments. Optional, default FALSE
	Token      *string `json:"token,omitempty"`          // Token used to make recurring payments to a previously saved wallet. Required for Direct Wallet Payments
	Expiration *string `json:"expiration,omitempty"`     // Expiration (in days) of a saved wallet. Optional
	Username   *string `json:"username,omitempty"`       // User's username in the merchant's website. Required for Webpay OneClick
	Email      *string `json:"email,omitempty"`          // User’s email. Required for Webpay OneClick
	Recurring  *string `json:"recurring_info,omitempty"` // Info of recurring information for the user. Optional (for MercadoPago)
	Verify     *bool   `json:"verify,omitempty"`         // If using the request just to verify the wallet, and not creating a payment. Mandatory as TRUE if amount=0. Default FALSE
	Capture    *bool   `json:"capture,omitempty"`        // Whether or not to immediately capture the charge. When FALSE, the charge issues an authorization and will need to be captured later. Default TRUE. Optional
	DeviceId   *string `json:"deviceId,omitempty"`       // Device information. Check relevant section. Required for Mercado Pago
}

// Refund Object that identifies the refund.
type Refund struct {
	Id                  *string           `json:"id,omitempty"`                   // The refund ID
	RefundId            *string           `json:"refund_id,omitempty"`            // The refund ID
	PaymentId           *string           `json:"payment_id,omitempty"`           // The payment ID
	Amount              *string           `json:"amount,omitempty"`               // The amount of the refund. Always in local currency
	AmountRefunded      *string           `json:"amount_refunded,omitempty"`      // Amount received by the end user. Always in local currency
	Currency            *string           `json:"currency,omitempty"`             // The currency code of the refund
	Status              *string           `json:"status,omitempty"`               // The status of the refund
	StatusCode          *string           `json:"status_code,omitempty"`          // The status code of the refund
	StatusDetail        *string           `json:"status_detail,omitempty"`        // The status detail
	CreatedDate         *string           `json:"created_date,omitempty"`         // The date of when the refund was executed
	Notification        *string           `json:"notification_url,omitempty"`     // URL where dLocal will send notifications associated to changes in this refund
	Description         *string           `json:"description,omitempty"`          // Description of the refund
	Bank                *string           `json:"bank,omitempty"`                 // User's bank name
	BankCode            *string           `json:"bank_code,omitempty"`            // User's bank code
	BankAccount         *string           `json:"bank_account,omitempty"`         // User's bank account number
	BankAccountType     *BankAccountType2 `json:"bank_account_type,omitempty"`    // Type of bank account. C: for Current accounts; S: for Savings accounts; I: International accounts
	BankBranchCode      *string           `json:"bank_branch,omitempty"`          // User's bank branch code
	BankBranchName      *string           `json:"bank_branch_name,omitempty"`     // User's bank branch name
	RefundOrderId       *string           `json:"order_refund_id,omitempty"`      // ID of the payment, given by the merchant in their system
	CustomMerchantName  *string           `json:"custom_merchant_name,omitempty"` // Custom merchant name
	BeneficiaryName     *string           `json:"beneficiary_name,omitempty"`     // Name of the account holder
	BeneficiaryLastname *string           `json:"beneficiary_lastname,omitempty"` // Last name of the account holder
	DocumentType        *string           `json:"document_type,omitempty"`        // User's document type
	DocumentId          *string           `json:"document_id,omitempty"`          // User's document number
	Phone               *string           `json:"phone,omitempty"`
	Email               *string           `json:"email,omitempty"`
	Address             *string           `json:"address,omitempty"`
	City                *string           `json:"city,omitempty"`
	Splits              *[]Split          `json:"splits,omitempty"`
}

// Order Object that identifies the payment.
type Order struct {
	OrderId      *string `json:"order_id,omitempty"`      // ID given by the merchant in their system
	PaymentId    *string `json:"payment_id,omitempty"`    // The payment ID at dLocal
	Amount       *string `json:"amount,omitempty"`        // Transaction amount (in the currency entered in the field currency)
	Currency     *string `json:"currency,omitempty"`      // Three-letter ISO currency code, in uppercase
	CreatedDate  *string `json:"created_date,omitempty"`  // Payment’s creation date
	ApprovedDate *string `json:"approved_date,omitempty"` // Payment’s approval date
	Status       *string `json:"status,omitempty"`        // Payment status
	StatusCode   *string `json:"status_code,omitempty"`   // The payment status code
	StatusDetail *string `json:"status_detail,omitempty"` // Payment status detail
}

// CurrencyExchange Object that identifies the currency exchange.
type CurrencyExchange struct {
	From *string  `json:"from,omitempty"` // Origin currency code (ISO-4217)
	To   *string  `json:"to,omitempty"`   // Destination currency code (ISO-4217)
	Rate *float64 `json:"rate,omitempty"` // Ratio of conversion from from currency to to currency
}

// ThreeDSecure Object that identifies the 3D Secure.
type ThreeDSecure struct {
	Mpi                    *bool                   `json:"mpi,omitempty"`                     // TRUE if you are going to use a 3rd-party 3D Secure provider. If null, then mpi = FALSE
	ThreeDSecureVersion    *ThreeDSecureVersion    `json:"three_dsecure_version,omitempty"`   // If null, then three_dsecure_version = "1.0"
	Cavv                   *string                 `json:"cavv,omitempty"`                    // The cardholder authentication value for the 3D Secure authentication session. The returned value is a base64-encoded 20-byte array. Required if mpi = TRUE
	Eci                    *ECI                    `json:"eci,omitempty"`                     // The electronic commerce indicator. Required if mpi = TRUE.
	Xid                    *string                 `json:"xid,omitempty"`                     // The transaction identifier assigned by the Directory Server for v1 authentication (base64 encoded, 20 bytes in a decoded form). Required if mpi = TRUE and three_dsecure_version = "1.0"
	DsTransactionId        *string                 `json:"ds_transaction_id,omitempty"`       // The transaction identifier assigned by the 3DS Server for v2 authentication (36 characters, commonly in UUID format). Required if mpi = TRUE and three_dsecure_version = 2.x
	EnrollmentResponse     *EnrollmentResponse     `json:"enrollment_response,omitempty"`     // The enrollment response from the VERes message from the Directory Server. Options EnrollmentResponse
	AuthenticationResponse *AuthenticationResponse `json:"authentication_response,omitempty"` // From the PARes from the issuer's Access Control System. Options AuthenticationResponse
}

type Payment struct {
	Id                 *string       `json:"id,omitempty"`
	AuthorizationId    *string       `json:"authorization_id,omitempty"`
	Amount             *float64      `json:"amount,omitempty"`
	Currency           *string       `json:"currency,omitempty"`
	PaymentMethodId    *string       `json:"payment_method_id,omitempty"`
	PaymentMethodType  *string       `json:"payment_method_type"`
	PaymentMethodFlow  *string       `json:"payment_method_flow,omitempty"`
	Country            *string       `json:"country,omitempty"`
	OrderId            *string       `json:"order_id,omitempty"`
	OriginalOrderId    *string       `json:"original_order_id,omitempty"`
	Description        *string       `json:"description,omitempty"`
	NotificationUrl    *string       `json:"notification_url,omitempty"`
	CallbackUrl        *string       `json:"callback_url,omitempty"`
	RedirectUrl        *string       `json:"redirect_url,omitempty"`
	AdditionalRiskData *string       `json:"additional_risk_data,omitempty"`
	CreatedDate        *string       `json:"created_date,omitempty"`
	ApprovedDate       *string       `json:"approved_date,omitempty"`
	Status             *string       `json:"status,omitempty"`
	StatusDetail       *string       `json:"status_detail,omitempty"`
	StatusCode         *string       `json:"status_code,omitempty"`
	Payer              *Payer        `json:"payer,omitempty"`
	Card               *Card         `json:"card,omitempty"`
	Ticket             *Ticket       `json:"ticket,omitempty"`
	ThreeDsecure       *ThreeDSecure `json:"three_dsecure,omitempty"`
	Splits             *[]Split      `json:"splits,omitempty"`
}

type Split struct {
	AccountId *string `json:"account_id,omitempty"`
	Amount    *string `json:"amount,omitempty"`
}

type SecureCard struct {
	Country *string `json:"country,omitempty"`
	Card    *Card   `json:"card,omitempty"`
	Payer   *Payer  `json:"payer,omitempty"`
}

type UniqueReference struct {
	Id                *string         `json:"id,omitempty"`
	Country           *string         `json:"country,omitempty"`
	PaymentMethodId   *string         `json:"payment_method_id,omitempty"`
	BankCode          *string         `json:"bank_code,omitempty"`
	BranchCode        *string         `json:"branch_code,omitempty"`
	BankAccount       *string         `json:"bank_account,omitempty"`
	ExternalReference *string         `json:"external_reference,omitempty"`
	CreatedDate       *string         `json:"created_date,omitempty"`
	Status            *string         `json:"status,omitempty"`
	StatusDetail      *string         `json:"status_detail,omitempty"`
	StatusCode        *string         `json:"status_code,omitempty"`
	Payer             *Payer          `json:"payer,omitempty"`
	VirtualAccount    *VirtualAccount `json:"virtual_account,omitempty"`
}

type VirtualAccount struct {
	BankCode    *string `json:"bank_code,omitempty"`
	BranchCode  *string `json:"branch_code,omitempty"`
	BankAccount *string `json:"bank_account,omitempty"`
}

type Notification struct {
	Id              *string  `json:"id,omitempty"`
	PaymentId       *string  `json:"payment_id,omitempty"`
	Amount          *float64 `json:"amount,omitempty"`
	Currency        *string  `json:"currency,omitempty"`
	Status          *string  `json:"status,omitempty"`
	StatusCode      *string  `json:"status_code,omitempty"`
	StatusDetail    *string  `json:"status_detail,omitempty"`
	CreatedDate     *string  `json:"created_date,omitempty"`
	NotificationUrl *string  `json:"notification_url,omitempty"`
	OrderId         *string  `json:"order_id,omitempty"`
}

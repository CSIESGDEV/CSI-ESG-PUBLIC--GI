package errcode

import "sync"

// Code :
type Code string

// Error codes :
const (
	TransactionNotExist   Code = "transaction_not_exist"
	MerchantNotExist      Code = "merchant_not_exist"
	MerchantNotActive     Code = "merchant_not_active"
	StoreNotExist         Code = "store_not_exist"
	OAuthAppNotExist      Code = "oauth_app_not_exist"
	OAuthAppCannotPublish Code = "oauth_app_cannot_publish"
	InvalidCurrencyType   Code = "invalid_currency_type"

	// Authentication
	InvalidAccess        Code = "invalid_access"
	MissingAccessToken   Code = "missing_access_token"
	InvalidAuthCode      Code = "invalid_auth_code"
	AuthCodeExpired      Code = "auth_code_expired"
	InviteCodeExpired    Code = "invite_code_expired"
	InvalidRefreshToken  Code = "invalid_refresh_token"
	InvalidAccessToken   Code = "invalid_access_token"
	InvalidSignature     Code = "invalid_signed_signature"
	GoogleSessionExpired Code = "google_session_expired"
	AdminNotFound        Code = "admin_not_found"
	AdminIsFound         Code = "admin_is_found"
	AdminAlreadyDeleted  Code = "admin_already_deleted"
	AdminNotActive       Code = "admin_not_active"
	UserNotFound         Code = "user_not_found"
	ExceedOTPRequest     Code = "otp_request_exceeded"
	EmailFound           Code = "email_found"
	EmailNotFound        Code = "email_not_found"
	SSMNumberFound       Code = "ssm_number_found"
	SSMNumberNotFound    Code = "ssm_number_not_found"
	CompanyNameFound     Code = "company_name_found"
	CompanyNameNotFound  Code = "company_name_not_found"
	RecordFound          Code = "record_found"
	AuthenticationError  Code = "authentication_error"
	InvalidToken         Code = "invalid_token"

	// Process error
	InvalidRequest               Code = "invalid_request"
	InvalidFile                  Code = "invalid_file"
	ValidationError              Code = "validation_error"
	ServerError                  Code = "server_error"
	UniqueIDAlreadyExist         Code = "unique_id_already_exist"
	InvalidQueryString           Code = "invalid_query_string"
	InvalidPhoneNumber           Code = "invalid_phone_number"
	CustomerEntryNotFound        Code = "customer_entry_not_found"
	CustomerEntryAlreadyExist    Code = "customer_entry_already_exist"
	NotAllowLogin                Code = "not_allow_login"
	CustomerEntryAlreadyVerified Code = "customer_entry_already_verified"
	FunctionNotFound             Code = "function_not_found"
	StatusInActive               Code = "status_inactive"
	ExceedUserLimit              Code = "exceed_user_limit"
	MismatchSME                  Code = "mismatch_sme_id"
	EmailError                   Code = "email_error"
	ResourceNotFound             Code = "resource_not_found"

	// API Key
	InvalidAPIKey     Code = "invalid_api_key"
	InvalidAPIScope   Code = "invalid_api_scope"
	NoPermittedScope  Code = "no_permitted_scope"
	ScopeNotPermitted Code = "scope_not_permitted"

	// Mykad
	RecordNotFound Code = "record_not_found"

	// OpenAPI
	InvalidOpenAPIToken Code = "TOKEN_INVALID"
)

// Message :
var Message sync.Map

func init() {
	Message.Store(ValidationError, "Validation error")
	Message.Store(ExceedOTPRequest, "OTP request too many times")
}

package entity

import "time"

type Collection = string

// Collection :
const (
	CollectionAdmin            Collection = "admin"
	CollectionSMEUser          Collection = "smeUser"
	CollectionSME              Collection = "sme"
	CollectionCorporate        Collection = "corporate"
	CollectionCorporateUser    Collection = "corporateUser"
	CollectionLearningResource Collection = "learningResource"
	CollectionQuestionSet      Collection = "questionSet"
	CollectionQuestion         Collection = "question"
	CollectionAssessment       Collection = "assessment"
	CollectionAssessmentEntry  Collection = "assessmentEntry"
	CollectionNews             Collection = "news"
	CollectionSubscription     Collection = "subscription"
	CollectionConnection       Collection = "connection"
)

// Model :
type Model struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	DeletedAt time.Time `bson:"deletedAt" json:"deletedAt"`
}

// UserStatus :
type UserStatus string

var (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusInActive  UserStatus = "INACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusInvited   UserStatus = "INVITED"
	UserStatusExpired   UserStatus = "EXPIRED"
	UserStatusDeleted   UserStatus = "DELETED"
)

// Status :
type Status string

var (
	StatusActive     Status = "ACTIVE"
	StatusInActive   Status = "INACTIVE"
	StatusVerified   Status = "VERIFIED"
	StatusUnverified Status = "UNVERIFIED"
	StatusSuspended  Status = "SUSPENDED"
	StatusInvited    Status = "INVITED"
	StatusExpired    Status = "EXPIRED"
	StatusDeleted    Status = "DELETED"
	StatusRejected   Status = "REJECTED"
	StatusPending    Status = "PENDING"
	StatusReceived   Status = "RECEIVED"
)

// RespondStatus :
type RespondStatus string

var (
	ResponseStatusToStart     RespondStatus = "TO_START"
	ResponseStatusInProgress  RespondStatus = "IN_PROGRESS"
	ResponseStatusToReview    RespondStatus = "TO_REVIEW"
	ResponseStatusToSubmitted RespondStatus = "SUBMITTED"
	ResponseStatusDeleted     RespondStatus = "DELETED"
)

// UserRole :
type UserRole string

var (
	UserRoleAdmin           UserRole = "ADMIN"
	UserRoleUser            UserRole = "USER"
	UserRoleSMEVendor       UserRole = "SME_VENDOR"
	UserRoleCorporateVendor UserRole = "CORPORATE_VENDOR"
)

// Designation :
type UserDesignation string

var (
	DesignationManager UserDesignation = "MANAGER"
	DesignationCEO     UserDesignation = "CEO"
)

// Scope :
type Scope string

var (
	ScopeAdminSME           Scope = "ADMIN_SME"
	ScopeAdminCorporate     Scope = "ADMIN_CORPORATE"
	ScopeSMEAnalytics       Scope = "SME_ANALYTICS"
	ScopeCorporateAnalytics Scope = "CORPORATE_ANALYTICS"
)

// UserTitle :
type UserTitle string

var (
	UserTitleDato     UserTitle = "DATO"
	UserTitleDatoSri  UserTitle = "DATO_SRI"
	UserTitleDatuk    UserTitle = "DATUK"
	UserTitleDatukSri UserTitle = "DATUK_SRI"
	UserTitleDr       UserTitle = "DR."
	UserTitleHaji     UserTitle = "HAJI"
	UserTitleMr       UserTitle = "MR."
	UserTitleMs       UserTitle = "MS."
	UserTitleProf     UserTitle = "PROF."
	UserTitleTanSri   UserTitle = "TAN_SRI"
	UserTitleTun      UserTitle = "TUN"
	UserTitleTunDr    UserTitle = "TUN_DR."
)

// BusinessEntity :
type BusinessEntity string

var (
	BusinessEntitySoleProprietorship           BusinessEntity = "SOLE_PROPRIETORSHIP"
	BusinessEntityPartnership                  BusinessEntity = "PARTNERSHIP"
	BusinessEntityLimitedLiabilityPartnership  BusinessEntity = "LIMITED_LIABILITY_PARTNERSHIP"
	BusinessEntitySendirianBerhad              BusinessEntity = "SENDIRIAN_BERHAD"
	BusinessEntityPublicLimited                BusinessEntity = "PUBLIC_LIMITED"
	BusinessEntityProfessionalServiceProviders BusinessEntity = "PROFESSIONAL_SERVICE_PROVIDERS"
	BusinessEntityEducationalInstitution       BusinessEntity = "EDUCATIONAL_INSTITUTION"
	BusinessEntityNGO                          BusinessEntity = "NGO"
	BusinessEntityGovernment                   BusinessEntity = "GOVERNMENT"
)

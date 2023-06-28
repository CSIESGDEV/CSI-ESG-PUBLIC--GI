package constant

import (
	"sme-api/app/entity"
	"sme-api/app/env"
)

// Constants
var (
	CORSDomain = []string{
		env.Config.App.SystemPath,
		"*"}
)

var UserDesignations = []entity.UserDesignation{
	entity.DesignationManager,
	entity.DesignationCEO,
}

var UserScopes = []entity.Scope{
	entity.ScopeAdminSME,
	entity.ScopeAdminCorporate,
	entity.ScopeSMEAnalytics,
	entity.ScopeCorporateAnalytics,
}

var UserMapScopes = map[string]string{
	"ADMINSME":           "AdminSME",
	"ADMINCORPORATE":     "AdminCorporate",
	"SMEANALYTICS":       "SMEAnalytics",
	"CORPORATEANALYTICS": "CorporateAnalytics",
}

var UserRoles = []entity.UserRole{
	entity.UserRoleAdmin,
	entity.UserRoleUser,
}

// Industries code :
var MSIC = []string{
	"D 35",
	"E 36-39",
	"G 45",
	"G 46",
	"G 47",
	"H 49-53",
	"I 55-56",
	"J 58-63",
	"K 64-66",
	"L 68",
	"M 69-75",
	"N 77-82",
	"O 84",
	"P 85",
	"Q 86-88",
	"R 90-93",
	"S 94-96",
	"C 10",
	"C 11",
	"C 12",
	"C 13",
	"C 14",
	"C 15",
	"C 16",
	"C 17",
	"C 18",
	"C 19",
	"C 20",
	"C 21",
	"C 22",
	"C 23",
	"C 24",
	"C 25",
	"C 26",
	"C 27",
	"C 28",
	"C 29",
	"C 30",
	"C 31",
	"C 32",
	"C 33",
	"F 41",
	"F 42",
	"F 43",
	"A 01",
	"A 02",
	"A 03",
	"B 05",
	"B 06",
	"B 07",
	"B 08",
	"B 09",
}

// UserTitles :
var UserTitles = []entity.UserTitle{
	entity.UserTitleDato,
	entity.UserTitleDatoSri,
	entity.UserTitleDatuk,
	entity.UserTitleDatukSri,
	entity.UserTitleDr,
	entity.UserTitleHaji,
	entity.UserTitleMr,
	entity.UserTitleMs,
	entity.UserTitleProf,
	entity.UserTitleTanSri,
	entity.UserTitleTun,
	entity.UserTitleTunDr,
}

// BusinessEntities :
var BusinessEntities = map[entity.BusinessEntity]string{
	entity.BusinessEntitySoleProprietorship:           "Sole Proprietorship",
	entity.BusinessEntityPartnership:                  "Partnership",
	entity.BusinessEntityLimitedLiabilityPartnership:  "Limited Liability Partnership",
	entity.BusinessEntitySendirianBerhad:              "Sendirian Berhad",
	entity.BusinessEntityProfessionalServiceProviders: "Professional Service Providers",
}

// UserMapTitles :
var UserMapTitles = map[string]entity.UserTitle{
	// entity.UserTitleDato:     "Dato",
	// entity.UserTitleDatoSri:  "Dato Sri",
	// entity.UserTitleDatuk:    "Datuk",
	// entity.UserTitleDatukSri: "Datuk Sri",
	// entity.UserTitleDr:       "Dr.",
	// entity.UserTitleHaji:     "Haji",
	"Mr.": entity.UserTitleMr,
	// entity.UserTitleMs:       "Ms.",
	// entity.UserTitleProf:     "Prof.",
	// entity.UserTitleTanSri:   "Tan Sri",
	// entity.UserTitleTun:      "Tun",
	// entity.UserTitleTunDr:    "Tun Dr.",
}

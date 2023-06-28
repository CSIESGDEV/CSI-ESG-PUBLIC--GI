package router

import (
	"sme-api/app/handler"

	"github.com/labstack/echo/v4"
)

// V1 :
func V1(e *echo.Echo, h *handler.Handler) {

	v1 := e.Group("/api/v1")
	v1.GET("/health", h.APIHealthCheckVersion1)
	// v1.POST("/contactUs", h.SendContactUsEmail)

	v1.POST("/signedURL", h.GetSignedUrl)

	v1.POST("/vendor/login", h.VendorLogin)
	v1.POST("/admin/login", h.AdminLogin)

	admin := v1.Group("/admin")
	admin.POST("", h.CreateAdmin)
	admin.GET("", h.GetAdminByToken)
	admin.GET("s", h.GetAdmins)
	admin.PUT("", h.UpdateAdmin)
	// admin.GET("/corporateUsers", h.GetCorporateUsers)
	admin.GET("/smeUsers", h.GetSMEUsers)
	admin.POST("/smeUser", h.CreateSMEUser)
	admin.PUT("/smeUser", h.UpdateSMEUser)
	// admin.POST("/corporateUser", h.CreateCorporateUser)
	// admin.PUT("/corporateUser", h.UpdateCorporateUser)

	v1.POST("/smeRegister", h.CreateSME)
	v1.POST("/smeSubscription", h.CreateSubscription)
	// SME profile
	sme := v1.Group("/sme")
	sme.POST("", h.CreateSME)
	sme.GET("s", h.GetSMEs)
	sme.PUT("", h.UpdateSME)
	sme.DELETE("", h.DeleteSME)
	// sme.GET("/corporateUsers", h.GetCorporateUsers)

	// SME user
	v1.POST("/smeUser/login", h.SMEUserLogin)
	v1.POST("/smeUser/register", h.CreateSMEUser)
	// v1.GET("/smeUser", h.GetSMEUserByToken)
	smeUser := v1.Group("/smeUser")
	smeUser.POST("", h.CreateSMEUser)
	smeUser.PUT("", h.UpdateSMEUser)
	smeUser.GET("", h.GetSMEUserByToken)
	smeUser.GET("s", h.GetSMEUsers)

	// v1.POST("/corporateRegister", h.CreateCorporate)
	// // Corporate
	// corporate := v1.Group("/corporate")
	// corporate.POST("", h.CreateCorporate)
	// corporate.GET("s", h.GetCorporates)
	// corporate.PUT("", h.UpdateCorporate)
	// corporate.DELETE("", h.DeleteCorporate)
	// corporate.GET("/smeUsers", h.GetSMEUsers)

	// subscription
	subscription := v1.Group("/subscription")
	subscription.POST("", h.CreateSubscription)
	subscription.GET("s", h.GetSubscriptions)
	subscription.PUT("", h.UpdateSubscription)
	subscription.DELETE("", h.DeleteSubscription)

	// Corporate user
	// v1.POST("/corporateUser/login", h.CorporateUserLogin)
	// v1.POST("/corporateUser/register", h.CreateCorporateUser)
	// // v1.GET("/corporateUser", h.GetCorporateUserByToken)
	// corporateUser := v1.Group("/corporateUser")
	// corporateUser.POST("", h.CreateCorporateUser)
	// corporateUser.PUT("", h.UpdateCorporateUser)
	// corporateUser.GET("", h.GetCorporateUserByToken)
	// corporateUser.GET("s", h.GetCorporateUsers)

	// Learning resource
	learningResource := v1.Group("/learningResource")
	learningResource.POST("", h.CreateLearningResource)
	learningResource.PUT("", h.UpdateLearningResource)
	learningResource.GET("s", h.GetLearningResources)
	learningResource.DELETE("", h.DeleteLearningResource)

	// News
	news := v1.Group("/news")
	news.POST("", h.CreateNews)
	news.PUT("", h.UpdateNews)
	news.GET("s", h.GetNews)
	news.DELETE("", h.DeleteNews)

	// Question set
	questionSet := v1.Group("/questionSet")
	questionSet.POST("", h.CreateQuestionSet)
	questionSet.PUT("", h.UpdateQuestionSet)
	questionSet.GET("s", h.GetQuestionSets)
	questionSet.DELETE("", h.DeleteQuestionSet)

	// Question
	question := v1.Group("/question")
	question.POST("", h.CreateQuestion)
	question.PUT("", h.UpdateQuestion)
	question.GET("s", h.GetQuestions)
	question.DELETE("", h.DeleteQuestion)

	// Assessment
	assessment := v1.Group("/assessment")
	assessment.POST("", h.CreateAssessment)
	assessment.PUT("", h.UpdateAssessment)
	assessment.GET("s", h.GetAssessments)
	assessment.DELETE("", h.DeleteAssessment)

	// Assessment Entry
	v1.GET("/assessmentEntries", h.GetAssessmentEntries)
	assessmentEntry := v1.Group("/assessmentEntry")
	assessmentEntry.PUT("/responses", h.SubmitResponse)
	assessmentEntry.PUT("/submit", h.SubmitAssessment)

	// Connection
	connection := v1.Group("/connection")
	connection.POST("", h.CreateConnection)
	connection.PUT("", h.UpdateConnection)
	connection.GET("s", h.GetConnections)
}

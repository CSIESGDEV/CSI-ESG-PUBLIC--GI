package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sme-api/app/constant"
	"sme-api/app/handler"
	"sme-api/app/router"
	"text/template"
	"time"

	bs "sme-api/app/bootstrap"

	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	casbinmw "github.com/reedom/echo-middleware-casbin"
)

// Template struct
type Template struct {
	templates *template.Template
}

// Render Template
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// casbinDataSource is a datasource for the middleware.
type casbinDataSource struct{}

// Introduce another middleware which extracts the user's role and set
// it to "role" in echo.Context.
func setUserMiddleware(h *handler.Handler) echo.MiddlewareFunc {
	return h.AuthenticatedUser
}

// GetSubject gets a subject from echo.Context.
// In this sample, it expects other middleware has set a user's role at "role".
func (r *casbinDataSource) GetSubject(c echo.Context) string {
	return c.Get("role").(string)
}

// Start :
func Start(port string) {
	bs := bs.New()
	h := handler.New(bs)
	e := echo.New()

	e.Validator = bs.Validator
	ce, err := casbin.NewEnforcer("app/casbin/rbac_model.conf", "app/casbin/rbac_policy.csv")
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Use(setUserMiddleware(h))
	e.Use(casbinmw.Middleware(ce, &casbinDataSource{}))

	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.BodyLimit("5M"),
		middleware.Secure(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(100)),
		middleware.RequestIDWithConfig(middleware.RequestIDConfig{
			Generator: func() string {
				return fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(100000))
			},
		}))

	// CORS : Allow cross site domain
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     constant.CORSDomain,
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.PATCH, echo.OPTIONS, echo.HEAD},
		AllowCredentials: true,
	}))

	// Adding template
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("app/resource/view/*.html")),
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "info", "SME API WORKING FINE")
	})

	router.V1(e, h)

	e.Logger.Fatal(e.Start(":" + port))
}

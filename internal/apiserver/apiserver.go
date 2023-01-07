package apiserver

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/armantarkhanian/jwt"
	"github.com/armantarkhanian/websocket"
	"github.com/centrifugal/centrifuge"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/renju24/backend/internal/pkg/apierror"
	"github.com/renju24/backend/internal/pkg/config"
	"github.com/rs/zerolog"
)

// APIServer is the main object of programm.
type APIServer struct {
	runMode string
	addr    string

	router         *gin.Engine
	logger         *zerolog.Logger
	config         *config.Config
	jwt            *jwt.EncodeDecoder
	centrifugeNode *centrifuge.Node

	// Dependecies.
	db Database
	ConfigReader
}

func redirectToTls(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}

// Run runs the HTTP server.
func (a *APIServer) Run() error {
	if a.runMode == "prod" {
		go func() {
			if err := http.ListenAndServe(":80", http.HandlerFunc(redirectToTls)); err != nil {
				a.logger.Fatal().Err(err).Send()
			}
		}()
		return a.router.RunTLS(a.addr, "./cert.crt", "./private.key")
	} else {
		return a.router.Run(a.addr)
	}
}

// APIServer should be a singleton, so make it global.
var singleton *APIServer

// NewAPIServer creates a singleton APIServer object.
func NewAPIServer(runMode string, db Database, router *gin.Engine, logger *zerolog.Logger, configReader ConfigReader) *APIServer {
	if singleton == nil {
		singleton = initApi(runMode, db, router, logger, configReader)
	}
	return singleton
}

func initApi(runMode string, db Database, router *gin.Engine, logger *zerolog.Logger, configReader ConfigReader) *APIServer {
	// Read config from database.
	logger.Info().Msg("Reading config from configReader.")
	config, err := configReader.ReadConfig()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	logger.Info().Interface("config", config).Send()

	jwtEncodeDecoder, err := jwt.New(config.Server.Token.SigningKey)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	a := &APIServer{
		runMode:      runMode,
		addr:         ":8008",
		router:       router,
		logger:       logger,
		config:       config,
		jwt:          jwtEncodeDecoder,
		db:           db,
		ConfigReader: configReader,
	}

	if a.runMode == "prod" {
		a.addr = ":443"
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowWebSockets = true

	a.router.Use(
		sameSiteMiddleware(a),
		cors.New(corsConfig),
		loggerMiddleware(a),
	)

	a.router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	})

	// TODO: just for test. Remove it in production.
	a.router.SetHTMLTemplate(template.Must(template.New("").Parse(`
	{{ define "unauthorized.html" }}
		<b>Вы не авторизованы!</b><br><br>

		<a href="/api/v1/oauth2/web/google" style="outline: none;text-decoration: none;">
			<img src="assets/images/logos/google.svg" width="30px" style="margin:10px;">
		</a>
		<a href="/api/v1/oauth2/web/yandex" style="outline: none;text-decoration: none;">
			<img src="assets/images/logos/yandex.svg" width="30px" style="margin:10px;">
		</a>
		<a href="/api/v1/oauth2/web/vk" style="outline: none;text-decoration: none;">
			<img src="assets/images/logos/vk.svg" width="30px" style="margin:10px;">
		</a>
	{{ end }}

	{{ define "authorized.html" }}
		<b>Вы авторизованы!</b><br><br>

		<b>ID</b>: {{ .ID }}<br>
		<b>Username</b>: {{ .Username }}<br>
		<b>Email</b>: {{ .Email }}<br><br>

		<a href="/logout">Логаут</a>
	{{ end }}
	`)))

	// Static assets directory.
	a.router.Static("/assets", "./assets")

	// Main page.
	a.router.GET("/", func(c *gin.Context) {
		token, _ := c.Cookie(a.config.Server.Token.Cookie.Name)
		var payload jwt.Payload
		if err := a.jwt.Decode(token, &payload); err != nil { // If user is not authorized.
			c.HTML(http.StatusOK, "unauthorized.html", nil)
			return
		}
		userID, err := strconv.ParseInt(payload.Subject, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, &apierror.Error{
				Error: apierror.ErrorInternal,
			})
			return
		}
		user, err := a.db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, &apierror.Error{
				Error: apierror.ErrorInternal,
			})
			return
		}
		c.HTML(http.StatusOK, "authorized.html", map[string]interface{}{
			"ID":       user.ID,
			"Username": user.Username,
			"Email":    user.Email,
		})
	})

	a.router.GET("/logout", func(c *gin.Context) {
		c.SetCookie(
			a.config.Server.Token.Cookie.Name,
			"",
			-1,
			a.config.Server.Token.Cookie.Path,
			a.config.Server.Token.Cookie.Domain,
			false,
			false,
		)
		c.Redirect(http.StatusFound, "/")
	})

	// POST /api/v1/*
	apiRoutes := a.router.Group("/api/v1")
	{
		apiRoutes.POST("/sign_up", signUp(a))
		apiRoutes.POST("/sign_in", signIn(a))
		apiRoutes.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "PONG") })
		apiRoutes.GET("/oauth2/:platform/services", oauth2Services(a))
		apiRoutes.GET("/oauth2/:platform/:service", oauth2Login(a))
		apiRoutes.GET("/oauth2/:platform/:service/callback", oauth2Callback(a))
	}

	// Initialize WebSocket server.
	node, handler, err := websocket.New(websocket.Config{
		Engine:        &websocket.MemoryEngine{},
		ClientHandler: a,
		NodeHandler:   a,
		TokenLookup: websocket.TokenLookup{
			Header:       config.Server.Token.Header.Name,
			Cookie:       config.Server.Token.Cookie.Name,
			HeaderPrefix: "Bearer",
		},
		WebsocketConfig: centrifuge.WebsocketConfig{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			UseWriteBufferPool: true,
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	a.centrifugeNode = node

	// GET /connection/websocket
	a.router.GET("/connection/websocket", gin.WrapH(handler))

	return a
}

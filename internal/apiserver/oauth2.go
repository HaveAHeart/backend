package apiserver

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/armantarkhanian/jwt"
	"github.com/gin-gonic/gin"
	oauth "github.com/renju24/backend/internal/pkg/oauth2"
)

// TODO: hard code.
func oauth2Services(api *APIServer) gin.HandlerFunc {
	type service struct {
		Name  oauth.Service `json:"name"`
		Image string        `json:"image"`
		URL   string        `json:"url"`
	}
	type response []service
	host := "http://localhost" + api.addr
	if api.runMode == "prod" {
		host = filepath.Join("https://", api.config.Server.Token.Cookie.Domain)
	}
	return func(c *gin.Context) {
		platform, err := oauth.ParsePlatform(c.Param("platform"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		var (
			googleLogo = filepath.Join(host, "assets/images/logos/google.svg")
			yandexLogo = filepath.Join(host, "assets/images/logos/yandex.svg")
			vkLogo     = filepath.Join(host, "assets/images/logos/vk.svg")
		)
		switch platform {
		case oauth.Web:
			c.JSON(http.StatusOK, &response{
				{
					Name:  oauth.Google,
					Image: googleLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Google.Callbacks.Web, "/callback"),
				},
				{
					Name:  oauth.Yandex,
					Image: yandexLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Yandex.Callbacks.Web, "/callback"),
				},
				{
					Name:  oauth.VK,
					Image: vkLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Yandex.Callbacks.Android, "/callback"),
				},
			})
		case oauth.Android:
			c.JSON(http.StatusOK, &response{
				{
					Name:  oauth.Google,
					Image: googleLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Google.Callbacks.Android, "/callback"),
				},
				{
					Name:  oauth.Yandex,
					Image: yandexLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Yandex.Callbacks.Android, "/callback"),
				},
				{
					Name:  oauth.VK,
					Image: vkLogo,
					URL:   strings.TrimSuffix(api.config.Oauth2.Yandex.Callbacks.Android, "/callback"),
				},
			})
		default:
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
}

func oauth2Login(api *APIServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := oauth.ParseService(c.Param("service"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		platform, err := oauth.ParsePlatform(c.Param("platform"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if platform == oauth.Web {
			cookieValue, _ := c.Cookie(api.config.Server.Token.Cookie.Name)
			var payload jwt.Payload
			if err = api.jwt.Decode(cookieValue, &payload); err == nil {
				// If user is already authorized, then redirect him to the main page.
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
				return
			}
		}
		oauthCfg, err := oauth.OauthConfig(api.config.Oauth2, service, platform)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		authPage := oauthCfg.AuthCodeURL("c5ae9e1e63e14761a2933110db39fb3a")
		c.Redirect(http.StatusFound, authPage)
	}
}

func oauth2Callback(api *APIServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := oauth.ParseService(c.Param("service"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		platform, err := oauth.ParsePlatform(c.Param("platform"))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		code := c.Request.FormValue("code")
		if code == "" {
			switch platform {
			case oauth.Android:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Android)
			default:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
			}
			return
		}
		var oauthUser *oauth.User
		switch service {
		case oauth.Google:
			oauthUser, err = oauth.GoogleOauth(api.config.Oauth2, code, service, platform)
		case oauth.Yandex:
			oauthUser, err = oauth.YandexOauth(api.config.Oauth2, code, service, platform)
		case oauth.VK:
			oauthUser, err = oauth.VKOauth(api.config.Oauth2, code, service, platform)
		}
		if err != nil {
			api.logger.Error().Err(err).Send()
			switch platform {
			case oauth.Android:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Android)
			default:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
			}
			return
		}
		user, err := api.db.CreateUserOauth(oauthUser.Username, oauthUser.Email, oauthUser.ID, service)
		if err != nil {
			api.logger.Error().Err(err).Send()
			switch platform {
			case oauth.Android:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Android)
			default:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
			}
			return
		}
		jwtToken, err := api.jwt.Encode(jwt.Payload{
			Subject:        strconv.FormatInt(user.ID, 10),
			ExpirationTime: int64(api.config.Server.Token.Cookie.MaxAge),
		})
		if err != nil {
			api.logger.Error().Err(err).Send()
			switch platform {
			case oauth.Android:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Android)
			default:
				c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
			}
			return
		}
		if platform == oauth.Android {
			deepLink := api.config.Oauth2.DeepLinks.Android + "?token=" + jwtToken
			c.Redirect(http.StatusFound, deepLink)
			return
		}
		c.SetCookie(
			api.config.Server.Token.Cookie.Name,
			jwtToken,
			api.config.Server.Token.Cookie.MaxAge,
			api.config.Server.Token.Cookie.Path,
			api.config.Server.Token.Cookie.Domain,
			api.config.Server.Token.Cookie.Secure,
			api.config.Server.Token.Cookie.HttpOnly,
		)
		c.Redirect(http.StatusFound, api.config.Oauth2.DeepLinks.Web)
	}
}

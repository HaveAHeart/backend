package apiserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/armantarkhanian/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/renju24/backend/internal/pkg/apierror"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

type signupResponse struct {
	Token string `json:"token"`
}

func signUp(api *APIServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req signupRequest
		if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
			c.JSON(http.StatusBadRequest, &APIError{
				Error: apierror.ErrorInvalidBody,
			})
			return
		}
		if err := req.Validate(); err != nil {
			if apiErr, ok := err.(*apierror.Error); ok {
				c.JSON(http.StatusBadRequest, &APIError{
					Error: apiErr,
				})
				return
			}
			c.JSON(http.StatusBadRequest, &APIError{
				Error: apierror.ErrorInvalidBody,
			})
			return
		}
		passwordBcrypt, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &APIError{
				Error: apierror.ErrorInternal,
			})
			return
		}
		user, err := api.db.CreateUser(req.Username, req.Email, string(passwordBcrypt))
		if err != nil {
			if errors.Is(err, apierror.ErrorUsernameIsTaken) {
				c.JSON(http.StatusBadRequest, &APIError{
					Error: apierror.ErrorUsernameIsTaken,
				})
				return
			}
			if errors.Is(err, apierror.ErrorEmailIsTaken) {
				c.JSON(http.StatusBadRequest, &APIError{
					Error: apierror.ErrorEmailIsTaken,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, &APIError{
				Error: apierror.ErrorInternal,
			})
			return
		}

		jwtToken, err := api.jwt.Encode(jwt.Payload{
			Subject:        strconv.FormatInt(user.ID, 10),
			ExpirationTime: int64(api.config.Server.Token.Cookie.MaxAge),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &APIError{
				Error: apierror.ErrorInternal,
			})
			return
		}

		resp := signupResponse{
			Token: jwtToken,
		}

		c.SetCookie(
			api.config.Server.Token.Cookie.Name,
			resp.Token,
			api.config.Server.Token.Cookie.MaxAge,
			api.config.Server.Token.Cookie.Path,
			api.config.Server.Token.Cookie.Domain,
			api.config.Server.Token.Cookie.Secure,
			api.config.Server.Token.Cookie.HttpOnly,
		)

		c.JSON(200, &resp)
	}
}

func (req *signupRequest) Validate() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" {
		return apierror.ErrorUsernameIsRequired
	}
	if req.Email == "" {
		return apierror.ErrorEmailIsRequired
	}
	if req.Password == "" {
		return apierror.ErrorPasswordIsRequired
	}
	if req.RepeatedPassword == "" {
		return apierror.ErrorRepeatedPasswordIsRequired
	}
	usernameLength := utf8.RuneCountInString(req.Username)
	if usernameLength < 4 || usernameLength > 32 {
		return apierror.ErrorInvalidUsernameLength
	}
	emailLength := utf8.RuneCountInString(req.Email)
	if emailLength < 5 || emailLength > 84 {
		return apierror.ErrorInvalidEmailLength
	}
	if strings.Count(req.Email, "@") != 1 {
		return apierror.ErrorInvalidEmail
	}
	passwordLength := utf8.RuneCountInString(req.Password)
	if passwordLength < 8 || passwordLength > 64 {
		return apierror.ErrorInvalidPasswordLength
	}

	var (
		hasLetter bool
		hasDigit  bool
	)

	for _, char := range req.Password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		default:
			return apierror.ErrorInvalidPasswordCharacter
		}
	}

	if !hasLetter {
		return apierror.ErrorMissingUpperInPassword
	}
	if !hasDigit {
		return apierror.ErrorMissingDigitInPassword
	}

	if req.Password != req.RepeatedPassword {
		return apierror.ErrorPasswordsNotEqual
	}

	return nil
}

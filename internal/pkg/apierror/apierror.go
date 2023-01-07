package apierror

import (
	"encoding/json"

	"github.com/centrifugal/centrifuge"
)

// Error is the JSON-object that server will return when an error occurs.
type Error struct {
	Error *centrifuge.Error `json:"error"`
}

type apiErrorJSON struct {
	Error errorJSON `json:"error"`
}

type errorJSON struct {
	Code      uint32 `json:"code"`
	Message   string `json:"message"`
	Temporary bool   `json:"temporary"`
}

// MarshalJSON ...
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&apiErrorJSON{
		Error: errorJSON{
			Code:      e.Error.Code,
			Message:   e.Error.Message,
			Temporary: e.Error.Temporary,
		},
	})
}

// MarshalJSON ...
func (e *Error) UnmarshalJSON(data []byte) error {
	var apiError apiErrorJSON
	if err := json.Unmarshal(data, &apiError); err != nil {
		return err
	}
	e.Error = &centrifuge.Error{
		Code:      apiError.Error.Code,
		Message:   apiError.Error.Message,
		Temporary: apiError.Error.Temporary,
	}
	return nil
}

var (
	ErrorInternal                   = centrifuge.ErrorInternal
	ErrorUnauthorized               = centrifuge.ErrorUnauthorized
	ErrorPermissionDenied           = centrifuge.ErrorPermissionDenied
	ErrorMethodNotFound             = centrifuge.ErrorMethodNotFound
	ErrorAlreadySubscribed          = centrifuge.ErrorAlreadySubscribed
	ErrorBadRequest                 = centrifuge.ErrorBadRequest
	ErrorUsernameIsRequired         = &centrifuge.Error{401, "username is required", false}
	ErrorEmailIsRequired            = &centrifuge.Error{402, "email is required", false}
	ErrorPasswordIsRequired         = &centrifuge.Error{403, "password is required", false}
	ErrorRepeatedPasswordIsRequired = &centrifuge.Error{404, "repeated_password is required", false}
	ErrorInvalidUsernameLength      = &centrifuge.Error{405, "invalid username length", false}
	ErrorInvalidEmail               = &centrifuge.Error{406, "invalid email", false}
	ErrorInvalidEmailLength         = &centrifuge.Error{407, "invalid email length", false}
	ErrorInvalidPasswordLength      = &centrifuge.Error{408, "invalid password length", false}
	ErrorInvalidPasswordCharacter   = &centrifuge.Error{409, "invalid password character", false}
	ErrorMissingLetterInPassword    = &centrifuge.Error{410, "missing letter character", false}
	ErrorMissingDigitInPassword     = &centrifuge.Error{411, "missing digit character", false}
	ErrorPasswordsNotEqual          = &centrifuge.Error{412, "passwords are not equal", false}
	ErrorUsernameIsTaken            = &centrifuge.Error{413, "username is already taken", false}
	ErrorEmailIsTaken               = &centrifuge.Error{414, "email is already taken", false}
	ErrorInvalidCredentials         = &centrifuge.Error{415, "invalid credentials", false}
	ErrorUserNotFound               = &centrifuge.Error{416, "user not found", false}
	ErrorCallingYourselfForGame     = &centrifuge.Error{417, "can't call yourself for a game", false}
	ErrorInvalidUsernameCharacter   = &centrifuge.Error{418, "invalid username character", false}
	ErrorInviterAlreadyPlaying      = &centrifuge.Error{419, "inviter already playing a game", false}
	ErrorOpponentAlreadyPlaying     = &centrifuge.Error{420, "opponent already playing a game", false}
	ErrorGameNotFound               = &centrifuge.Error{421, "game not found", false}
	ErrorGameIsNotActive            = &centrifuge.Error{422, "game is not active", false}
	ErrFirstMoveShouldBeBlack       = &centrifuge.Error{423, "first move should be made by black user", false}
	ErrFirstMoveShouldBeInCenter    = &centrifuge.Error{424, "first move should be in board's center", false}
	ErrCoordinatesOutside           = &centrifuge.Error{425, "coordinates outside the board", false}
	ErrFieldAlreadyTaken            = &centrifuge.Error{426, "field is already taken", false}
	ErrInvalidTurn                  = &centrifuge.Error{427, "invalid turn", false}
)

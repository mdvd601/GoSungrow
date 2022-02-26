package login

import (
	"GoSungrow/Only"
	"GoSungrow/iSolarCloud/api"
	"GoSungrow/iSolarCloud/api/output"
	"errors"
	"fmt"
	"time"
)

const (
	DateTimeFormat       = "2006-01-02T15:04:05"
	DefaultAuthTokenFile = "AuthToken.json"
	TokenValidHours      = 24
	LastLoginDateFormat  = "2006-01-02 15:04:05"
)

type SunGrowAuth struct {
	AppKey       string
	UserAccount  string
	UserPassword string
	TokenFile    string
	Token        string
	Force        bool

	lastLogin time.Time
	newToken  bool
	retry     int
	err       error
}

func (a *SunGrowAuth) Verify() error {
	var err error
	for range Only.Once {
		if a == nil {
			err = errors.New("no auth details")
			break
		}

		if a.AppKey == "" {
			err = errors.New("API AppKey")
			a.err = err
			break
		}
		if a.UserAccount == "" {
			err = errors.New("empty API Username")
			a.err = err
			break
		}
		if a.UserPassword == "" {
			err = errors.New("empty API Password")
			a.err = err
			break
		}
	}

	return err
}

func (e *EndPoint) Login(auth *SunGrowAuth) error {
	for range Only.Once {
		e.Auth = auth
		e.Request.RequestData = RequestData{
			UserAccount:  auth.UserAccount,
			UserPassword: auth.UserPassword,
		}
		e.Request.RequestCommon = api.RequestCommon{
			Appkey:  auth.AppKey,
			SysCode: "900",
		}

		e.Error = e.Auth.Verify()
		if e.Error != nil {
			break
		}

		e.Error = e.readTokenFile()
		if e.Error != nil {
			break
		}

		if auth.Force {
			e.Auth.Token = ""
			e.Response.ResultData.Token = ""
		}

		if e.IsTokenValid() {
			break
		}

		foo := Assert(e.Call())
		e.Error = foo.GetError()
		if e.Error != nil {
			break
		}
		e.Request = foo.Request
		e.Response = foo.Response

		// e.Auth.UserAccount = e.Response.ResultData.UserAccount
		e.Auth.Token = e.Response.ResultData.Token
		e.Auth.lastLogin, _ = time.Parse(LastLoginDateFormat, e.Response.ResultData.LoginLastDate)

		e.Error = e.saveToken()
		if e.Error != nil {
			break
		}
	}

	return e.Error
}

func (e *EndPoint) IsTokenInvalid() bool {
	for range Only.Once {
		if e.Token() == "" {
			e.Auth.newToken = true
			break
		}
		if e.HoursFromLastLogin() > TokenValidHours {
			e.Auth.newToken = true
			break
		}
	}

	return e.Auth.newToken
}

func (e *EndPoint) IsTokenValid() bool {
	var ok bool
	for range Only.Once {
		if e.Token() == "" {
			break
		}
		if e.HoursFromLastLogin() > TokenValidHours {
			break
		}
		ok = true
	}

	return ok
}

func (e *EndPoint) HoursFromLastLogin() time.Duration {
	return time.Now().Sub(e.Auth.lastLogin)
}

func (e *EndPoint) HasTokenChanged() bool {
	ok := e.Auth.newToken
	if e.Auth.newToken {
		e.Auth.newToken = false
	}
	return ok
}

func (e *EndPoint) LastLogin() time.Time {
	return e.Auth.lastLogin
}

func (e *EndPoint) Print() {
	fmt.Printf("Email:\t%s\n", e.Email())
	fmt.Printf("Create Date:\t%s\n", e.CreateDate())
	fmt.Printf("Login Last Date:\t%s\n", e.LoginLastDate())
	fmt.Printf("Login Last IP:\t%s\n", e.LoginLastIP())
	fmt.Printf("Login State:\t%s\n", e.LoginState())
	fmt.Printf("User Account:\t%s\n", e.UserAccount())
	fmt.Printf("User Id:\t%s\n", e.UserID())
	fmt.Printf("User Name:\t%s\n", e.UserName())
	fmt.Printf("Is Online:\t%s\n", e.IsOnline())
	fmt.Printf("Token:\t%s\n", e.Token())
	fmt.Printf("Token File:\t%s\n", e.Auth.TokenFile)
}

// Retrieves a token from a local file.
func (e *EndPoint) readTokenFile() error {
	for range Only.Once {
		e.Auth.TokenFile = e.GetFilePath()

		// e.Error = e.ApiReadDataFile(e.Auth.TokenFile, &e.Response.ResultData)
		e.Error = output.FileRead(e.Auth.TokenFile, &e.Response.ResultData)
		if e.Error != nil {
			break
		}

		e.Response.ResultData.Msg = ""
		e.Response.ResultMsg = ""
		e.Response.ResultCode = ""

		e.Auth.Token = e.Token()
		if e.Auth.Token == "" {
			e.Auth.newToken = true
			break
		}

		// 2022-02-17 14:27:03
		e.Auth.lastLogin, e.Error = time.Parse(LastLoginDateFormat, e.Response.ResultData.LoginLastDate)
		if e.Error != nil {
			e.Auth.newToken = true
			e.Error = nil
			break
		}
	}

	return e.Error
}

// Saves a token to a file path.
func (e *EndPoint) saveToken() error {
	for range Only.Once {
		e.Auth.TokenFile = e.GetFilePath()

		e.Error = output.FileWrite(e.Auth.TokenFile, e.Response.ResultData, output.DefaultFileMode)
		// e.Error = e.ApiWriteDataFile(e.Auth.TokenFile, e.Response.ResultData, 0644)
		if e.Error != nil {
			break
		}
	}

	return e.Error
}

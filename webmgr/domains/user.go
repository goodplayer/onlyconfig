package domains

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/goodplayer/onlyconfig/webmgr/config"
	"github.com/goodplayer/onlyconfig/webmgr/tools"
)

type UserJwt struct {
	Username string `json:"un"`
	UserId   string `json:"uid"`

	jwt.RegisteredClaims
}

type User struct {
	UserId   string
	UserName string
	Password string

	Name  string
	Email string
}

func (u *User) ValidatePassword(password string) bool {
	return tools.ValidatePassword(u.Password, password)
}

func (u *User) GenerateJwtToken(key []byte) (string, error) {
	now := time.Now()
	claims := &UserJwt{
		Username: u.UserName,
		UserId:   u.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour * 30)), // 30天
			IssuedAt:  jwt.NewNumericDate(now),                          // 签发时间
			NotBefore: jwt.NewNumericDate(now),                          // 生效时间
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	return s, nil
}

type UserHandler struct {
	Config    *atomic.Value
	UserStore UserStore
}

func (uh *UserHandler) getConfig() *config.WebManagerConfig {
	return uh.Config.Load().(*config.WebManagerConfig)
}

func (uh *UserHandler) LoginUser(ctx context.Context, username string, password string) (*User, error) {
	u, err := uh.UserStore.QueryUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}

	if !u.ValidatePassword(password) {
		return nil, errors.New("invalid password")
	}

	return u, nil
}

func (uh *UserHandler) ValidateUserJwtToken(token string) (bool, error) {
	claims := new(UserJwt)
	cfg := uh.getConfig()

	var errs []error
	for _, key := range cfg.JwtSecrets {
		t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if _, ok := t.Claims.(*UserJwt); ok && t.Valid {
			return true, nil
		} else {
			return false, err
		}
	}
	return false, errors.Join(errs...)
}

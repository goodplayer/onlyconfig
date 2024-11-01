package domains

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/goodplayer/onlyconfig/webmgr/config"
	"github.com/goodplayer/onlyconfig/webmgr/tools"
)

type UserJwt struct {
	Username string `json:"un"`
	UserId   string `json:"uid"`
	UUID     string `json:"uuid"`

	jwt.RegisteredClaims
}

type User struct {
	UserId   string
	UserName string
	Password string

	Name  string
	Email string
}

func (u *User) EncryptPassword() (err error) {
	u.Password, err = tools.HashPassword(u.Password)
	return
}

func (u *User) ValidatePassword(password string) bool {
	return tools.ValidatePassword(u.Password, password)
}

func (u *User) GenerateJwtToken(key []byte) (string, string, error) {
	now := time.Now()
	claims := &UserJwt{
		Username: u.UserName,
		UserId:   u.UserId,
		UUID:     uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24hr
			IssuedAt:  jwt.NewNumericDate(now),                     // 签发时间
			NotBefore: jwt.NewNumericDate(now),                     // 生效时间
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		return "", "", err
	}
	return s, claims.UUID, nil
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

func (uh *UserHandler) GetClaimsFromJwtToken(token string) (*UserJwt, error) {
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

		if val, ok := t.Claims.(*UserJwt); ok && t.Valid {
			return val, nil
		} else {
			return nil, err
		}
	}
	return nil, errors.Join(errs...)
}

func (uh *UserHandler) QueryOrganizationsByUserId(ctx context.Context, userId string) ([]*Org, error) {
	return uh.UserStore.QueryOrganizationsByUserId(ctx, userId)
}

func (uh *UserHandler) LoadOrganizationByOrgId(ctx context.Context, orgId string) (*Org, error) {
	return uh.UserStore.QueryOrganizationByOrgId(ctx, orgId)
}

func (uh *UserHandler) CreateOrganization(ctx context.Context, orgName string, username string) error {
	user, err := uh.UserStore.QueryUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if has, err := uh.UserStore.ExistsOrganizationByName(ctx, orgName); err != nil {
		return err
	} else if has {
		return errors.New("organization already exists")
	}
	now := time.Now()
	org := &Org{
		OrgId:       orgName,
		OrgName:     orgName,
		TimeCreated: now.UnixMilli(),
		TimeUpdated: now.UnixMilli(),
	}
	return uh.UserStore.SaveNewOrganization(ctx, org, user)
}

func (uh *UserHandler) AddUserToOrg(ctx context.Context, username, orgName string, role int) error {
	user, err := uh.UserStore.QueryUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	org, err := uh.UserStore.QueryOrganizationByName(ctx, orgName)
	if err != nil {
		return err
	}
	if has, err := uh.UserStore.ExistsUserOrgLink(ctx, user, org); err != nil {
		return err
	} else if has {
		return errors.New("the user has joint to the org")
	}
	if role == 1 || role == 2 {
		// 1 - owner
		// 2 - user
		if err := uh.UserStore.LinkUserOrg(ctx, user, org, role); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("the role is invalid:" + strconv.Itoa(role))
	}
}

type UserRegisterReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (uh *UserHandler) UserRegister(ctx context.Context, req *UserRegisterReq) error {
	user := &User{
		UserId:   req.Username,
		UserName: req.Username,
		Password: req.Password,
		Name:     req.DisplayName,
		Email:    req.Email,
	}
	if err := user.EncryptPassword(); err != nil {
		return err
	}
	if err := uh.UserStore.SaveNewUser(ctx, user); err != nil {
		return err
	}
	return nil
}

type ChangePasswordReq struct {
	Username    string
	OldPassword string `json:"old"`
	NewPassword string `json:"new"`
}

func (uh *UserHandler) ChangePassword(ctx context.Context, req *ChangePasswordReq) error {
	user, err := uh.UserStore.QueryUserByUsername(ctx, req.Username)
	if err != nil {
		return err
	}
	if !user.ValidatePassword(req.OldPassword) {
		return errors.New("invalid old password")
	}
	user.Password = req.NewPassword
	if err := user.EncryptPassword(); err != nil {
		return err
	}
	if err := uh.UserStore.UpdateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

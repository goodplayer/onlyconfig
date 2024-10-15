package postgres

import (
	"context"

	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/dbtxn"
)

type User struct {
	UserId         string `xorm:"'user_id' pk"`
	Username       string `xorm:"'username' unique"`
	Password       string `xorm:"'password'"`
	DisplayName    string `xorm:"'display_name'"`
	Email          string `xorm:"'email'"`
	UserStatus     int    `xorm:"'user_status'"`
	ExternalType   string `xorm:"'external_type'"`
	ExternalUserId string `xorm:"'external_user_id'"`
	TimeCreated    int64  `xorm:"'time_created'"`
	TimeUpdated    int64  `xorm:"'time_updated'"`
}

func (u *User) TableName() string {
	return "onlyconfig_user"
}

type UserStoreImpl struct {
}

func (u *UserStoreImpl) QueryUserByUsername(ctx context.Context, username string) (*domains.User, error) {
	sess := dbtxn.GetTxn(ctx)

	user := new(User)
	has, err := sess.Where("username = ? and user_status = 0", username).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	} else {
		return &domains.User{
			UserId:   user.UserId,
			UserName: user.Username,
			Password: user.Password,
			Name:     user.DisplayName,
			Email:    user.Email,
		}, nil
	}
}

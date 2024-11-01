package postgres

import (
	"context"
	"errors"
	"time"

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

type Org struct {
	OrgId       string `xorm:"'org_id' pk"`
	OrgName     string `xorm:"'org_name'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *Org) TableName() string {
	return "onlyconfig_org"
}

type UserOrgMapping struct {
	MappingId   string `xorm:"'user_org_mapping_id' pk autoincr"`
	OrgId       string `xorm:"'org_id'"`
	UserId      string `xorm:"'user_id'"`
	RoleType    int    `xorm:"'role_type'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *UserOrgMapping) TableName() string {
	return "onlyconfig_user_org_mapping"
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
		return nil, errors.New("user not exists")
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

func (u *UserStoreImpl) QueryOrganizationsByUserId(ctx context.Context, userId string) (result []*domains.Org, rerr error) {
	sess := dbtxn.GetTxn(ctx)

	var mappings []*UserOrgMapping
	if err := sess.Where("user_id = ?", userId).Find(&mappings); err != nil {
		return nil, err
	}
	for _, mapping := range mappings {
		var org = new(Org)
		if has, err := sess.Where("org_id = ?", mapping.OrgId).Get(org); err != nil {
			return nil, err
		} else if !has {
			return nil, errors.New("org id not found: " + mapping.OrgId)
		} else {
			ownerList, userList, err := u.queryOrgUsersByOrgId(ctx, org.OrgId)
			if err != nil {
				return nil, err
			}
			result = append(result, &domains.Org{
				OrgId:       org.OrgId,
				OrgName:     org.OrgName,
				TimeCreated: org.TimeCreated,
				TimeUpdated: org.TimeUpdated,

				UserList:  userList,
				OwnerList: ownerList,
			})
		}
	}

	return
}

func (u *UserStoreImpl) QueryOrganizationByOrgId(ctx context.Context, orgId string) (*domains.Org, error) {
	r := new(Org)

	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("org_id = ?", orgId).Get(r); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("org id not found: " + orgId)
	} else {
		return &domains.Org{
			OrgId:       r.OrgId,
			OrgName:     r.OrgName,
			TimeCreated: r.TimeCreated,
			TimeUpdated: r.TimeUpdated,
		}, nil
	}
}

func (u *UserStoreImpl) queryOrgUsersByOrgId(ctx context.Context, orgId string) (ownerList, userList []*domains.User, rerr error) {
	var userMapping []*UserOrgMapping
	sess := dbtxn.GetTxn(ctx)
	if err := sess.Where("org_id = ?", orgId).Find(&userMapping); err != nil {
		return nil, nil, err
	}
	for _, mapping := range userMapping {
		var user = new(User)
		if has, err := sess.Where("user_id = ?", mapping.UserId).Get(user); err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errors.New("user not found: " + mapping.UserId)
		} else {
			if mapping.RoleType == 1 {
				ownerList = append(ownerList, &domains.User{
					UserId:   user.UserId,
					UserName: user.Username,
					Password: user.Password,
					Name:     user.DisplayName,
					Email:    user.Email,
				})
			} else if mapping.RoleType == 2 {
				userList = append(userList, &domains.User{
					UserId:   user.UserId,
					UserName: user.Username,
					Password: user.Password,
					Name:     user.DisplayName,
					Email:    user.Email,
				})
			} else {
				return nil, nil, errors.New("unknown role type for userId: " + mapping.UserId)
			}
		}
	}
	return
}

func (u *UserStoreImpl) SaveNewOrganization(ctx context.Context, org *domains.Org, owner *domains.User) error {
	organization := &Org{
		OrgId:       org.OrgId,
		OrgName:     org.OrgName,
		TimeCreated: org.TimeCreated,
		TimeUpdated: org.TimeUpdated,
	}
	mapping := &UserOrgMapping{
		MappingId:   "",
		OrgId:       organization.OrgId,
		UserId:      owner.UserId,
		RoleType:    1, // role_type = owner
		TimeCreated: organization.TimeCreated,
		TimeUpdated: organization.TimeUpdated,
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(organization); err != nil {
		return err
	}
	if _, err := sess.Insert(mapping); err != nil {
		return err
	}
	return nil
}

func (u *UserStoreImpl) ExistsOrganizationByName(ctx context.Context, orgName string) (bool, error) {
	org := new(Org)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("org_name = ?", orgName).Get(org); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (u *UserStoreImpl) LinkUserOrg(ctx context.Context, user *domains.User, org *domains.Org, role int) error {
	now := time.Now()
	mapping := &UserOrgMapping{
		OrgId:       org.OrgId,
		UserId:      user.UserId,
		RoleType:    role,
		TimeCreated: now.UnixMilli(),
		TimeUpdated: now.UnixMilli(),
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(mapping); err != nil {
		return err
	}
	return nil
}

func (u *UserStoreImpl) QueryOrganizationByName(ctx context.Context, orgName string) (*domains.Org, error) {
	organization := new(Org)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("org_name = ?", orgName).Get(organization); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("organization not found: " + orgName)
	}
	ownerList, userList, err := u.queryOrgUsersByOrgId(ctx, organization.OrgId)
	if err != nil {
		return nil, err
	}
	return &domains.Org{
		OrgId:       organization.OrgId,
		OrgName:     organization.OrgName,
		TimeCreated: organization.TimeCreated,
		TimeUpdated: organization.TimeUpdated,
		OwnerList:   ownerList,
		UserList:    userList,
	}, nil
}

func (u *UserStoreImpl) ExistsUserOrgLink(ctx context.Context, user *domains.User, org *domains.Org) (bool, error) {
	mapping := new(UserOrgMapping)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("org_id = ? and user_id = ?", org.OrgId, user.UserId).Get(mapping); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (u *UserStoreImpl) SaveNewUser(ctx context.Context, user *domains.User) error {
	now := time.Now()
	newUser := &User{
		UserId:         user.UserId,
		Username:       user.UserName,
		Password:       user.Password,
		DisplayName:    user.Name,
		Email:          user.Email,
		UserStatus:     0,
		ExternalType:   "",
		ExternalUserId: "",
		TimeCreated:    now.UnixMilli(),
		TimeUpdated:    now.UnixMilli(),
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(newUser); err != nil {
		return err
	}
	return nil
}

func (u *UserStoreImpl) UpdateUser(ctx context.Context, user *domains.User) error {
	sess := dbtxn.GetTxn(ctx)

	userEntity := new(User)
	has, err := sess.Where("user_id = ? and user_status = 0", user.UserId).Get(userEntity)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("user not exists")
	}

	userEntity.Username = user.UserName
	userEntity.Password = user.Password
	userEntity.DisplayName = user.Name
	userEntity.Email = user.Email
	now := time.Now()
	userEntity.TimeUpdated = now.UnixMilli()

	if _, err := sess.ID(userEntity.UserId).Update(userEntity); err != nil {
		return err
	}
	return nil
}

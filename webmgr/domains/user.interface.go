package domains

import "context"

type UserStore interface {
	SaveNewOrganization(ctx context.Context, org *Org, owner *User) error
	LinkUserOrg(ctx context.Context, user *User, org *Org, role int) error
	SaveNewUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error

	QueryUserByUsername(ctx context.Context, username string) (*User, error)
	QueryOrganizationsByUserId(ctx context.Context, userId string) ([]*Org, error)
	QueryOrganizationByOrgId(ctx context.Context, orgId string) (*Org, error)
	ExistsOrganizationByName(ctx context.Context, orgName string) (bool, error)
	QueryOrganizationByName(ctx context.Context, orgName string) (*Org, error)
	ExistsUserOrgLink(ctx context.Context, user *User, org *Org) (bool, error)
}

package domains

type Org struct {
	OrgId       string
	OrgName     string
	TimeCreated int64
	TimeUpdated int64

	OwnerList []*User
	UserList  []*User
}

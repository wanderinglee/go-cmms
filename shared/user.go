package shared

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Name     string `db:"name"`
	Passwd   string `db:"passwd"`
	Email    string `db:"email"`
	Role     string `db:"role"`
	SMS      string `db:"sms"`
	Sites    []Site `db:"site"`
}

type UserUpdateData struct {
	Channel int
	User    *User
}

type UserUpdate struct {
	Channel  int    `db:"channel"`
	ID       int    `db:"id"`
	Username string `db:"username"`
	Name     string `db:"name"`
	Passwd   string `db:"passwd"`
	Email    string `db:"email"`
	SMS      string `db:"sms"`
}

type UserSite struct {
	SiteID   int    `db:"site_id"`
	SiteName string `db:"site_name"`
	Count    int    `db:"count"`
}

type SiteUser struct {
	UserID   int    `db:"user_id"`
	Username string `db:"username"`
	Count    int    `db:"count"`
}

type UserSiteRequest struct {
	Channel int
	User    *User
	Site    *Site
}

type UserSiteSetRequest struct {
	Channel int
	UserID  int
	SiteID  int
	Role    string
	IsSet   bool
}
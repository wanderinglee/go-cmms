package main

import (
	"fmt"
	"log"
	"time"

	"github.com/steveoc64/go-cmms/shared"
)

type UserRPC struct{}

///////////////////////////////////////////////////////////
// SQL
const UserGetQuery = `select 
u.id,u.username,u.passwd,u.email,u.role,u.sms,u.name
	from users u
	where id=$1`

const UserListQuery = `select 
u.id,u.username,u.passwd,u.email,u.role,u.sms,u.name
	from users u
	order by u.username`

///////////////////////////////////////////////////////////
// Code

// Get all users
func (u *UserRPC) List(channel int, profs *[]shared.User) error {
	start := time.Now()

	conn := Connections.Get(channel)

	DB.SQL(UserListQuery, conn.UserID).QueryStructs(profs)

	logger(start, "User.List",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%d Users", len(*profs)))

	return nil
}

// Get the user for the given channel
func (u *UserRPC) Me(channel int, prof *shared.User) error {
	start := time.Now()

	conn := Connections.Get(channel)

	DB.SQL(UserGetQuery, conn.UserID).QueryStruct(prof)

	logger(start, "User.Me",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%s %s %s", prof.Email, prof.SMS, prof.Name))

	return nil
}

// Get the user for the given id
func (u *UserRPC) Get(id int, prof *shared.User) error {
	start := time.Now()

	DB.SQL(UserGetQuery, id).QueryStruct(prof)

	logger(start, "User.Get",
		fmt.Sprintf("%d", id),
		fmt.Sprintf("%s %s %s", prof.Email, prof.SMS, prof.Name))

	return nil
}

// Set the user profile from the popdown list at the top
func (u *UserRPC) Set(req *shared.UserUpdate, done *bool) error {
	start := time.Now()

	conn := Connections.Get(req.Channel)

	DB.Update("users").
		SetWhitelist(req, "name", "passwd", "email", "sms").
		Where("id = $1", req.ID).
		Exec()

	logger(start, "User.Set",
		fmt.Sprintf("Channel %d, User %d %s %s",
			req.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%s %s %s %s", req.Email, req.SMS, req.Name, req.Passwd))

	// *done = true

	return nil
}

// Full update of user record, including username
func (u *UserRPC) Update(data *shared.UserUpdateData, done *bool) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	DB.Update("users").
		SetWhitelist(data.User, "username", "name", "passwd", "email", "sms").
		Where("id = $1", data.User.ID).
		Exec()

	logger(start, "User.Save",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%d %s %s %s %s",
			data.User.ID, data.User.Username, data.User.Email, data.User.SMS, data.User.Name, data.User.Passwd))

	*done = true

	return nil
}

// Add a new user record
func (u *UserRPC) Insert(data *shared.UserUpdateData, id *int) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	DB.InsertInto("users").
		Whitelist("username", "name", "passwd", "email", "sms").
		Record(data.User).
		Returning("id").
		QueryScalar(id)

	logger(start, "User.Insert",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%d %s %s %s %s %s",
			*id, data.User.Username, data.User.Email, data.User.SMS, data.User.Name, data.User.Passwd))

	return nil
}

// Delete a user
func (u *UserRPC) Delete(data *shared.UserUpdateData, ok *bool) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	*ok = false
	id := data.User.ID
	DB.DeleteFrom("users").
		Where("id=$1", id).
		Exec()

	logger(start, "User.Delete",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("%d %s %s %s %s %s",
			id, data.User.Username, data.User.Email, data.User.SMS, data.User.Name, data.User.Passwd))

	return nil
}

// Get an array of Sites for this user
func (u *UserRPC) GetSites(data shared.UserSiteRequest, userSites *[]shared.UserSite) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	DB.SQL(`select 
		s.id as site_id,s.name as site_name,count(u.*)
		from site s
		left join user_site u
			on u.site_id=s.id
			and u.user_id=$1
		group by s.id
		order by s.name`, data.User.ID).QueryStructs(userSites)

	logger(start, "User.GetSites",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("User %d - %d Sites",
			data.User.ID, len(*userSites)))

	return nil
}

// Get an array of Users for this site
func (u *UserRPC) GetSiteUsers(data shared.UserSiteRequest, siteUsers *[]shared.SiteUser) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	DB.SQL(`select 
		u.id as user_id,u.username as username,count(s.*)
		from users u
		left join user_site s
			on s.user_id=u.id
			and s.site_id=$1
		group by u.id
		order by u.username`, data.Site.ID).QueryStructs(siteUsers)

	logger(start, "User.GetSiteUsers",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("Site %d - %d Users",
			data.Site.ID, len(*siteUsers)))

	return nil
}

// Set the user site relationship
func (u *UserRPC) SetSite(data shared.UserSiteSetRequest, done *bool) error {
	start := time.Now()

	conn := Connections.Get(data.Channel)

	// delete any existing relationship
	DB.DeleteFrom("user_site").
		Where("user_id=$1 and site_id=$2", data.UserID, data.SiteID).
		Exec()

	if data.IsSet {
		// if the role is undefined, then read it from the user
		if data.Role == "" {
			DB.SQL(`select role from users where id=$1`, data.UserID).QueryScalar(&data.Role)
			log.Println("fetched user role", data.Role)
		}

		DB.SQL(`insert into 
			user_site (user_id,site_id,role)
			values    ($1, $2, $3)`, data.UserID, data.SiteID, data.Role).
			Exec()
	}

	logger(start, "User.SetSite",
		fmt.Sprintf("Channel %d, User %d %s %s",
			data.Channel, conn.UserID, conn.Username, conn.UserRole),
		fmt.Sprintf("User %d Site %d Role %s %t",
			data.UserID, data.SiteID, data.Role, data.IsSet))

	*done = true
	return nil
}
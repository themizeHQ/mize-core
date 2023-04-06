package user

type UserStatusType string

var (
	AVAILABLE UserStatusType = "available"
	AWAY      UserStatusType = "away"
	BUSY      UserStatusType = "busy"
	MEETING   UserStatusType = "meeting"
	CLASS     UserStatusType = "class"
)

var AvailableUserLanguage = []interface{}{"english"}

// User Discoverability
//
//	set methods their profile can be found
type UserDiscoverability string

var (
	DISCOVERABILITY_EMAIL    UserDiscoverability = "email"
	DISCOVERABILITY_USERNAME UserDiscoverability = "username"
	DISCOVERABILITY_PHONE    UserDiscoverability = "phone"
)

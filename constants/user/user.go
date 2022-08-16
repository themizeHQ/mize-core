package user

type UserStatusType string

var (
	AVAILABLE UserStatusType = "available"
	AWAY      UserStatusType = "away"
	BUSY      UserStatusType = "busy"
	MEETING   UserStatusType = "meeting"
)

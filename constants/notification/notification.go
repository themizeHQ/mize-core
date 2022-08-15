package notification

// NOTIFICATION CONSTANT
//
// different type of notifications that created.
//
// eg. workspace invite, added to channel, tagged in message.
type NotificationType string

var (
	WORKSPACE_INVITE NotificationType = "workspace_invite"
)

// NOTIFICATION CONSTANT
//
// represents who gets to see the notification
type NotificationScope string

var (
	WORKSPACE_NOTIFICATION NotificationScope = "workspace_notification"
	APP_WIDE_NOTIFICATION  NotificationScope = "app_wide_notification"
	USER_NOTIFICATION      NotificationScope = "user_notification"
)

// NOTIFICATION CONSTANT
//
//  represents how important a notification is
type NotificationImportanceLevel string

var (
	NOTIFICATION_VERY_IMPORTANT NotificationImportanceLevel = "very_important"
	NOTIFICATION_IMPORTANT      NotificationImportanceLevel = "important"
	NOTIFICATION_NORMAL         NotificationImportanceLevel = "normal"
	NOTIFICATION_NOT_IMPORTANT  NotificationImportanceLevel = "not_important"
)

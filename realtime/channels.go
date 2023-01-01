package realtime

type DefaultChannelsType struct {
	APP_WIDE_NOTIFICATION_CHANNEL string
}

var DefaultChannels = DefaultChannelsType{
	APP_WIDE_NOTIFICATION_CHANNEL: "14mr1362-17b2-4211-6p9q-1qec60cm207v",
}

type MessageScopeType struct {
	NOTIFICATION string
	ALERT        string
	CONVERSATION string
}

var MessageScope = MessageScopeType{
	NOTIFICATION: "notification",
	ALERT:        "alert",
	CONVERSATION: "conversation",
}

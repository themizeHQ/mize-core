package channel

type ChannelAdminAccess string

const (
	CHANNEL_FULL_ACCESS       ChannelAdminAccess = "full_access"
	CHANNEL_DELETE_ACCESS     ChannelAdminAccess = "delete_access"
	CHANNEL_MEMBERSHIP_ACCESS ChannelAdminAccess = "channel_membership_access"
	CHANNEL_INFO_EDIT_ACCESS  ChannelAdminAccess = "channel_info_edit_access"
)

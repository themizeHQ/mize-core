package workspace

type WorkspaceType string

type workspaceType struct {
	PROFESSIONAL WorkspaceType
	EDUCATIONAL  WorkspaceType
}

var WorkspaceTypes = workspaceType{
	PROFESSIONAL: "professional",
	EDUCATIONAL:  "educational",
}

type AdminAccessType string

type adminAccess struct {
	FULL_ACCESS AdminAccessType
}

var AdminAccess = adminAccess{
	FULL_ACCESS: "full_access",
}

type MemberActionType string

type memberActions struct {
	SEND_TEXT_MESSAGE  MemberActionType
	SEND_IMAGE         MemberActionType
	SEND_VOICE_MESSAGE MemberActionType
	SEND_VIDEO         MemberActionType
}

var MemberActions = memberActions{
	SEND_TEXT_MESSAGE:  "send_text_message",
	SEND_IMAGE:         "send_image",
	SEND_VOICE_MESSAGE: "send_voice_message",
	SEND_VIDEO:         "send_video",
}

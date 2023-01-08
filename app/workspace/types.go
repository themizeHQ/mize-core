package workspace

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	workspace_member_constants "mize.app/constants/workspace"
)

type AddWorkspacePiviledges struct {
	Permissions []workspace_member_constants.AdminAccessType `bson:"permissions" json:"permissions"`
	Id          string                                       `bson:"id" json:"id"`
}

func (p *AddWorkspacePiviledges) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Id, validation.Required.Error("id is a required field"), is.MongoID.Error("id must be a valid mongodb id")),
		validation.Field(&p.Permissions, validation.Each(validation.In(workspace_member_constants.AdminAccess.FULL_ACCESS,
			workspace_member_constants.AdminAccess.EDIT_CHANNELS_ACCESS, workspace_member_constants.AdminAccess.ALERT_ACCESS, workspace_member_constants.AdminAccess.EDIT_MEMBERS_ACCESS,
			workspace_member_constants.AdminAccess.SCHEDULE_ACCESS, workspace_member_constants.AdminAccess.TEAMS_ACCESS).Error("pass in a valid permission"))),
	)
}

type UpdateChannelType struct {
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Private     bool   `bson:"private,omitempty" json:"private,omitempty"`
	Compulsory  bool   `bson:"compulsory,omitempty" json:"compulsory,omitempty"`
}

func (c *UpdateChannelType) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Description, validation.Length(0, 1000)),
		validation.Field(&c.Name, validation.Length(0, 50)),
	)
}

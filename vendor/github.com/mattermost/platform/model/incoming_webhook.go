// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

const (
	DEFAULT_WEBHOOK_USERNAME = "webhook"
	DEFAULT_WEBHOOK_ICON     = "/static/images/webhook_icon.jpg"
)

type IncomingWebhook struct {
	Id          string `json:"id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	UserId      string `json:"user_id"`
	ChannelId   string `json:"channel_id"`
	TeamId      string `json:"team_id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type IncomingWebhookRequest struct {
	Text        string          `json:"text"`
	Username    string          `json:"username"`
	IconURL     string          `json:"icon_url"`
	ChannelName string          `json:"channel"`
	Props       StringInterface `json:"props"`
	Attachments interface{}     `json:"attachments"`
	Type        string          `json:"type"`
}

func (o *IncomingWebhook) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func IncomingWebhookFromJson(data io.Reader) *IncomingWebhook {
	decoder := json.NewDecoder(data)
	var o IncomingWebhook
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func IncomingWebhookListToJson(l []*IncomingWebhook) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func IncomingWebhookListFromJson(data io.Reader) []*IncomingWebhook {
	decoder := json.NewDecoder(data)
	var o []*IncomingWebhook
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (o *IncomingWebhook) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.id.app_error", nil, "")
	}

	if o.CreateAt == 0 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.create_at.app_error", nil, "id="+o.Id)
	}

	if o.UpdateAt == 0 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.update_at.app_error", nil, "id="+o.Id)
	}

	if len(o.UserId) != 26 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.user_id.app_error", nil, "")
	}

	if len(o.ChannelId) != 26 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.channel_id.app_error", nil, "")
	}

	if len(o.TeamId) != 26 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.team_id.app_error", nil, "")
	}

	if len(o.DisplayName) > 64 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.display_name.app_error", nil, "")
	}

	if len(o.Description) > 128 {
		return NewLocAppError("IncomingWebhook.IsValid", "model.incoming_hook.description.app_error", nil, "")
	}

	return nil
}

func (o *IncomingWebhook) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
}

func (o *IncomingWebhook) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func IncomingWebhookRequestFromJson(data io.Reader) *IncomingWebhookRequest {
	decoder := json.NewDecoder(data)
	var o IncomingWebhookRequest
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

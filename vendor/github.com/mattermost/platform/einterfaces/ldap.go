// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package einterfaces

import (
	"github.com/mattermost/platform/model"
)

type LdapInterface interface {
	DoLogin(id string, password string) (*model.User, *model.AppError)
	GetUser(id string) (*model.User, *model.AppError)
	CheckPassword(id string, password string) *model.AppError
	SwitchToLdap(userId, ldapId, ldapPassword string) *model.AppError
	ValidateFilter(filter string) *model.AppError
	Syncronize() *model.AppError
	StartLdapSyncJob()
}

var theLdapInterface LdapInterface

func RegisterLdapInterface(newInterface LdapInterface) {
	theLdapInterface = newInterface
}

func GetLdapInterface() LdapInterface {
	return theLdapInterface
}

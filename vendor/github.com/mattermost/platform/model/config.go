// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

const (
	CONN_SECURITY_NONE     = ""
	CONN_SECURITY_TLS      = "TLS"
	CONN_SECURITY_STARTTLS = "STARTTLS"

	IMAGE_DRIVER_LOCAL = "local"
	IMAGE_DRIVER_S3    = "amazons3"

	DATABASE_DRIVER_MYSQL    = "mysql"
	DATABASE_DRIVER_POSTGRES = "postgres"

	SERVICE_GITLAB = "gitlab"
	SERVICE_GOOGLE = "google"

	WEBSERVER_MODE_REGULAR  = "regular"
	WEBSERVER_MODE_GZIP     = "gzip"
	WEBSERVER_MODE_DISABLED = "disabled"

	GENERIC_NOTIFICATION = "generic"
	FULL_NOTIFICATION    = "full"

	DIRECT_MESSAGE_ANY  = "any"
	DIRECT_MESSAGE_TEAM = "team"

	FAKE_SETTING = "********************************"
)

type ServiceSettings struct {
	ListenAddress                     string
	MaximumLoginAttempts              int
	SegmentDeveloperKey               string
	GoogleDeveloperKey                string
	EnableOAuthServiceProvider        bool
	EnableIncomingWebhooks            bool
	EnableOutgoingWebhooks            bool
	EnableCommands                    *bool
	EnableOnlyAdminIntegrations       *bool
	EnablePostUsernameOverride        bool
	EnablePostIconOverride            bool
	EnableTesting                     bool
	EnableDeveloper                   *bool
	EnableSecurityFixAlert            *bool
	EnableInsecureOutgoingConnections *bool
	EnableMultifactorAuthentication   *bool
	AllowCorsFrom                     *string
	SessionLengthWebInDays            *int
	SessionLengthMobileInDays         *int
	SessionLengthSSOInDays            *int
	SessionCacheInMinutes             *int
	WebsocketSecurePort               *int
	WebsocketPort                     *int
	WebserverMode                     *string
}

type SSOSettings struct {
	Enable          bool
	Secret          string
	Id              string
	Scope           string
	AuthEndpoint    string
	TokenEndpoint   string
	UserApiEndpoint string
}

type SqlSettings struct {
	DriverName         string
	DataSource         string
	DataSourceReplicas []string
	MaxIdleConns       int
	MaxOpenConns       int
	Trace              bool
	AtRestEncryptKey   string
}

type LogSettings struct {
	EnableConsole bool
	ConsoleLevel  string
	EnableFile    bool
	FileLevel     string
	FileFormat    string
	FileLocation  string
}

type FileSettings struct {
	DriverName                 string
	Directory                  string
	EnablePublicLink           bool
	PublicLinkSalt             string
	ThumbnailWidth             int
	ThumbnailHeight            int
	PreviewWidth               int
	PreviewHeight              int
	ProfileWidth               int
	ProfileHeight              int
	InitialFont                string
	AmazonS3AccessKeyId        string
	AmazonS3SecretAccessKey    string
	AmazonS3Bucket             string
	AmazonS3Region             string
	AmazonS3Endpoint           string
	AmazonS3BucketEndpoint     string
	AmazonS3LocationConstraint *bool
	AmazonS3LowercaseBucket    *bool
}

type EmailSettings struct {
	EnableSignUpWithEmail    bool
	EnableSignInWithEmail    *bool
	EnableSignInWithUsername *bool
	SendEmailNotifications   bool
	RequireEmailVerification bool
	FeedbackName             string
	FeedbackEmail            string
	SMTPUsername             string
	SMTPPassword             string
	SMTPServer               string
	SMTPPort                 string
	ConnectionSecurity       string
	InviteSalt               string
	PasswordResetSalt        string
	SendPushNotifications    *bool
	PushNotificationServer   *string
	PushNotificationContents *string
}

type RateLimitSettings struct {
	EnableRateLimiter bool
	PerSec            int
	MemoryStoreSize   int
	VaryByRemoteAddr  bool
	VaryByHeader      string
}

type PrivacySettings struct {
	ShowEmailAddress bool
	ShowFullName     bool
}

type SupportSettings struct {
	TermsOfServiceLink *string
	PrivacyPolicyLink  *string
	AboutLink          *string
	HelpLink           *string
	ReportAProblemLink *string
	SupportEmail       *string
}

type TeamSettings struct {
	SiteName                  string
	MaxUsersPerTeam           int
	EnableTeamCreation        bool
	EnableUserCreation        bool
	EnableOpenServer          *bool
	RestrictCreationToDomains string
	RestrictTeamNames         *bool
	EnableCustomBrand         *bool
	CustomBrandText           *string
	RestrictDirectMessage     *string
}

type LdapSettings struct {
	// Basic
	Enable             *bool
	LdapServer         *string
	LdapPort           *int
	ConnectionSecurity *string
	BaseDN             *string
	BindUsername       *string
	BindPassword       *string

	// Filtering
	UserFilter *string

	// User Mapping
	FirstNameAttribute *string
	LastNameAttribute  *string
	EmailAttribute     *string
	UsernameAttribute  *string
	NicknameAttribute  *string
	IdAttribute        *string

	// Advanced
	SkipCertificateVerification *bool
	QueryTimeout                *int

	// Customization
	LoginFieldName *string
}

type ComplianceSettings struct {
	Enable      *bool
	Directory   *string
	EnableDaily *bool
}

type Config struct {
	ServiceSettings    ServiceSettings
	TeamSettings       TeamSettings
	SqlSettings        SqlSettings
	LogSettings        LogSettings
	FileSettings       FileSettings
	EmailSettings      EmailSettings
	RateLimitSettings  RateLimitSettings
	PrivacySettings    PrivacySettings
	SupportSettings    SupportSettings
	GitLabSettings     SSOSettings
	GoogleSettings     SSOSettings
	LdapSettings       LdapSettings
	ComplianceSettings ComplianceSettings
}

func (o *Config) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *Config) GetSSOService(service string) *SSOSettings {
	switch service {
	case SERVICE_GITLAB:
		return &o.GitLabSettings
	case SERVICE_GOOGLE:
		return &o.GoogleSettings
	}

	return nil
}

func ConfigFromJson(data io.Reader) *Config {
	decoder := json.NewDecoder(data)
	var o Config
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func (o *Config) SetDefaults() {

	if len(o.SqlSettings.AtRestEncryptKey) == 0 {
		o.SqlSettings.AtRestEncryptKey = NewRandomString(32)
	}

	if len(o.FileSettings.PublicLinkSalt) == 0 {
		o.FileSettings.PublicLinkSalt = NewRandomString(32)
	}

	if o.FileSettings.AmazonS3LocationConstraint == nil {
		o.FileSettings.AmazonS3LocationConstraint = new(bool)
		*o.FileSettings.AmazonS3LocationConstraint = false
	}

	if o.FileSettings.AmazonS3LowercaseBucket == nil {
		o.FileSettings.AmazonS3LowercaseBucket = new(bool)
		*o.FileSettings.AmazonS3LowercaseBucket = false
	}

	if len(o.EmailSettings.InviteSalt) == 0 {
		o.EmailSettings.InviteSalt = NewRandomString(32)
	}

	if len(o.EmailSettings.PasswordResetSalt) == 0 {
		o.EmailSettings.PasswordResetSalt = NewRandomString(32)
	}

	if o.ServiceSettings.EnableDeveloper == nil {
		o.ServiceSettings.EnableDeveloper = new(bool)
		*o.ServiceSettings.EnableDeveloper = false
	}

	if o.ServiceSettings.EnableSecurityFixAlert == nil {
		o.ServiceSettings.EnableSecurityFixAlert = new(bool)
		*o.ServiceSettings.EnableSecurityFixAlert = true
	}

	if o.ServiceSettings.EnableInsecureOutgoingConnections == nil {
		o.ServiceSettings.EnableInsecureOutgoingConnections = new(bool)
		*o.ServiceSettings.EnableInsecureOutgoingConnections = false
	}

	if o.ServiceSettings.EnableMultifactorAuthentication == nil {
		o.ServiceSettings.EnableMultifactorAuthentication = new(bool)
		*o.ServiceSettings.EnableMultifactorAuthentication = false
	}

	if o.TeamSettings.RestrictTeamNames == nil {
		o.TeamSettings.RestrictTeamNames = new(bool)
		*o.TeamSettings.RestrictTeamNames = true
	}

	if o.TeamSettings.EnableCustomBrand == nil {
		o.TeamSettings.EnableCustomBrand = new(bool)
		*o.TeamSettings.EnableCustomBrand = false
	}

	if o.TeamSettings.CustomBrandText == nil {
		o.TeamSettings.CustomBrandText = new(string)
		*o.TeamSettings.CustomBrandText = ""
	}

	if o.TeamSettings.EnableOpenServer == nil {
		o.TeamSettings.EnableOpenServer = new(bool)
		*o.TeamSettings.EnableOpenServer = false
	}

	if o.TeamSettings.RestrictDirectMessage == nil {
		o.TeamSettings.RestrictDirectMessage = new(string)
		*o.TeamSettings.RestrictDirectMessage = DIRECT_MESSAGE_ANY
	}

	if o.EmailSettings.EnableSignInWithEmail == nil {
		o.EmailSettings.EnableSignInWithEmail = new(bool)

		if o.EmailSettings.EnableSignUpWithEmail == true {
			*o.EmailSettings.EnableSignInWithEmail = true
		} else {
			*o.EmailSettings.EnableSignInWithEmail = false
		}
	}

	if o.EmailSettings.EnableSignInWithUsername == nil {
		o.EmailSettings.EnableSignInWithUsername = new(bool)
		*o.EmailSettings.EnableSignInWithUsername = false
	}

	if o.EmailSettings.SendPushNotifications == nil {
		o.EmailSettings.SendPushNotifications = new(bool)
		*o.EmailSettings.SendPushNotifications = false
	}

	if o.EmailSettings.PushNotificationServer == nil {
		o.EmailSettings.PushNotificationServer = new(string)
		*o.EmailSettings.PushNotificationServer = ""
	}

	if o.EmailSettings.PushNotificationContents == nil {
		o.EmailSettings.PushNotificationContents = new(string)
		*o.EmailSettings.PushNotificationContents = GENERIC_NOTIFICATION
	}

	if !IsSafeLink(o.SupportSettings.TermsOfServiceLink) {
		o.SupportSettings.TermsOfServiceLink = nil
	}

	if o.SupportSettings.TermsOfServiceLink == nil {
		o.SupportSettings.TermsOfServiceLink = new(string)
		*o.SupportSettings.TermsOfServiceLink = "/static/help/terms.html"
	}

	if !IsSafeLink(o.SupportSettings.PrivacyPolicyLink) {
		o.SupportSettings.PrivacyPolicyLink = nil
	}

	if o.SupportSettings.PrivacyPolicyLink == nil {
		o.SupportSettings.PrivacyPolicyLink = new(string)
		*o.SupportSettings.PrivacyPolicyLink = "/static/help/privacy.html"
	}

	if !IsSafeLink(o.SupportSettings.AboutLink) {
		o.SupportSettings.AboutLink = nil
	}

	if o.SupportSettings.AboutLink == nil {
		o.SupportSettings.AboutLink = new(string)
		*o.SupportSettings.AboutLink = "/static/help/about.html"
	}

	if !IsSafeLink(o.SupportSettings.HelpLink) {
		o.SupportSettings.HelpLink = nil
	}

	if o.SupportSettings.HelpLink == nil {
		o.SupportSettings.HelpLink = new(string)
		*o.SupportSettings.HelpLink = "/static/help/help.html"
	}

	if !IsSafeLink(o.SupportSettings.ReportAProblemLink) {
		o.SupportSettings.ReportAProblemLink = nil
	}

	if o.SupportSettings.ReportAProblemLink == nil {
		o.SupportSettings.ReportAProblemLink = new(string)
		*o.SupportSettings.ReportAProblemLink = "/static/help/report_problem.html"
	}

	if o.SupportSettings.SupportEmail == nil {
		o.SupportSettings.SupportEmail = new(string)
		*o.SupportSettings.SupportEmail = "feedback@mattermost.com"
	}

	if o.LdapSettings.LdapPort == nil {
		o.LdapSettings.LdapPort = new(int)
		*o.LdapSettings.LdapPort = 389
	}

	if o.LdapSettings.QueryTimeout == nil {
		o.LdapSettings.QueryTimeout = new(int)
		*o.LdapSettings.QueryTimeout = 60
	}

	if o.LdapSettings.Enable == nil {
		o.LdapSettings.Enable = new(bool)
		*o.LdapSettings.Enable = false
	}

	if o.LdapSettings.UserFilter == nil {
		o.LdapSettings.UserFilter = new(string)
		*o.LdapSettings.UserFilter = ""
	}

	if o.LdapSettings.LoginFieldName == nil {
		o.LdapSettings.LoginFieldName = new(string)
		*o.LdapSettings.LoginFieldName = ""
	}

	if o.ServiceSettings.SessionLengthWebInDays == nil {
		o.ServiceSettings.SessionLengthWebInDays = new(int)
		*o.ServiceSettings.SessionLengthWebInDays = 30
	}

	if o.ServiceSettings.SessionLengthMobileInDays == nil {
		o.ServiceSettings.SessionLengthMobileInDays = new(int)
		*o.ServiceSettings.SessionLengthMobileInDays = 30
	}

	if o.ServiceSettings.SessionLengthSSOInDays == nil {
		o.ServiceSettings.SessionLengthSSOInDays = new(int)
		*o.ServiceSettings.SessionLengthSSOInDays = 30
	}

	if o.ServiceSettings.SessionCacheInMinutes == nil {
		o.ServiceSettings.SessionCacheInMinutes = new(int)
		*o.ServiceSettings.SessionCacheInMinutes = 10
	}

	if o.ServiceSettings.EnableCommands == nil {
		o.ServiceSettings.EnableCommands = new(bool)
		*o.ServiceSettings.EnableCommands = false
	}

	if o.ServiceSettings.EnableOnlyAdminIntegrations == nil {
		o.ServiceSettings.EnableOnlyAdminIntegrations = new(bool)
		*o.ServiceSettings.EnableOnlyAdminIntegrations = true
	}

	if o.ServiceSettings.WebsocketPort == nil {
		o.ServiceSettings.WebsocketPort = new(int)
		*o.ServiceSettings.WebsocketPort = 80
	}

	if o.ServiceSettings.WebsocketSecurePort == nil {
		o.ServiceSettings.WebsocketSecurePort = new(int)
		*o.ServiceSettings.WebsocketSecurePort = 443
	}

	if o.ServiceSettings.AllowCorsFrom == nil {
		o.ServiceSettings.AllowCorsFrom = new(string)
		*o.ServiceSettings.AllowCorsFrom = ""
	}

	if o.ServiceSettings.WebserverMode == nil {
		o.ServiceSettings.WebserverMode = new(string)
		*o.ServiceSettings.WebserverMode = "regular"
	}

	if o.ComplianceSettings.Enable == nil {
		o.ComplianceSettings.Enable = new(bool)
		*o.ComplianceSettings.Enable = false
	}

	if o.ComplianceSettings.Directory == nil {
		o.ComplianceSettings.Directory = new(string)
		*o.ComplianceSettings.Directory = "./data/"
	}

	if o.ComplianceSettings.EnableDaily == nil {
		o.ComplianceSettings.EnableDaily = new(bool)
		*o.ComplianceSettings.EnableDaily = false
	}

	if o.LdapSettings.ConnectionSecurity == nil {
		o.LdapSettings.ConnectionSecurity = new(string)
		*o.LdapSettings.ConnectionSecurity = ""
	}

	if o.LdapSettings.SkipCertificateVerification == nil {
		o.LdapSettings.SkipCertificateVerification = new(bool)
		*o.LdapSettings.SkipCertificateVerification = false
	}

	if o.LdapSettings.NicknameAttribute == nil {
		o.LdapSettings.NicknameAttribute = new(string)
		*o.LdapSettings.NicknameAttribute = ""
	}
}

func (o *Config) IsValid() *AppError {

	if o.ServiceSettings.MaximumLoginAttempts <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.login_attempts.app_error", nil, "")
	}

	if len(o.ServiceSettings.ListenAddress) == 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.listen_address.app_error", nil, "")
	}

	if o.TeamSettings.MaxUsersPerTeam <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.max_users.app_error", nil, "")
	}

	if !(*o.TeamSettings.RestrictDirectMessage == DIRECT_MESSAGE_ANY || *o.TeamSettings.RestrictDirectMessage == DIRECT_MESSAGE_TEAM) {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.restrict_direct_message.app_error", nil, "")
	}

	if len(o.SqlSettings.AtRestEncryptKey) < 32 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.encrypt_sql.app_error", nil, "")
	}

	if !(o.SqlSettings.DriverName == DATABASE_DRIVER_MYSQL || o.SqlSettings.DriverName == DATABASE_DRIVER_POSTGRES) {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.sql_driver.app_error", nil, "")
	}

	if o.SqlSettings.MaxIdleConns <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.sql_idle.app_error", nil, "")
	}

	if len(o.SqlSettings.DataSource) == 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.sql_data_src.app_error", nil, "")
	}

	if o.SqlSettings.MaxOpenConns <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.sql_max_conn.app_error", nil, "")
	}

	if !(o.FileSettings.DriverName == IMAGE_DRIVER_LOCAL || o.FileSettings.DriverName == IMAGE_DRIVER_S3) {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_driver.app_error", nil, "")
	}

	if o.FileSettings.PreviewHeight < 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_preview_height.app_error", nil, "")
	}

	if o.FileSettings.PreviewWidth <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_preview_width.app_error", nil, "")
	}

	if o.FileSettings.ProfileHeight <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_profile_height.app_error", nil, "")
	}

	if o.FileSettings.ProfileWidth <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_profile_width.app_error", nil, "")
	}

	if o.FileSettings.ThumbnailHeight <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_thumb_height.app_error", nil, "")
	}

	if o.FileSettings.ThumbnailWidth <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_thumb_width.app_error", nil, "")
	}

	if len(o.FileSettings.PublicLinkSalt) < 32 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.file_salt.app_error", nil, "")
	}

	if !(o.EmailSettings.ConnectionSecurity == CONN_SECURITY_NONE || o.EmailSettings.ConnectionSecurity == CONN_SECURITY_TLS || o.EmailSettings.ConnectionSecurity == CONN_SECURITY_STARTTLS) {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.email_security.app_error", nil, "")
	}

	if len(o.EmailSettings.InviteSalt) < 32 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.email_salt.app_error", nil, "")
	}

	if len(o.EmailSettings.PasswordResetSalt) < 32 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.email_reset_salt.app_error", nil, "")
	}

	if o.RateLimitSettings.MemoryStoreSize <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.rate_mem.app_error", nil, "")
	}

	if o.RateLimitSettings.PerSec <= 0 {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.rate_sec.app_error", nil, "")
	}

	if !(*o.LdapSettings.ConnectionSecurity == CONN_SECURITY_NONE || *o.LdapSettings.ConnectionSecurity == CONN_SECURITY_TLS || *o.LdapSettings.ConnectionSecurity == CONN_SECURITY_STARTTLS) {
		return NewLocAppError("Config.IsValid", "model.config.is_valid.ldap_security.app_error", nil, "")
	}

	return nil
}

func (o *Config) GetSanitizeOptions() map[string]bool {
	options := map[string]bool{}
	options["fullname"] = o.PrivacySettings.ShowFullName
	options["email"] = o.PrivacySettings.ShowEmailAddress

	return options
}

func (o *Config) Sanitize() {
	if &o.LdapSettings != nil && len(*o.LdapSettings.BindPassword) > 0 {
		*o.LdapSettings.BindPassword = FAKE_SETTING
	}

	o.FileSettings.PublicLinkSalt = FAKE_SETTING
	if len(o.FileSettings.AmazonS3SecretAccessKey) > 0 {
		o.FileSettings.AmazonS3SecretAccessKey = FAKE_SETTING
	}

	o.EmailSettings.InviteSalt = FAKE_SETTING
	o.EmailSettings.PasswordResetSalt = FAKE_SETTING
	if len(o.EmailSettings.SMTPPassword) > 0 {
		o.EmailSettings.SMTPPassword = FAKE_SETTING
	}

	if len(o.GitLabSettings.Secret) > 0 {
		o.GitLabSettings.Secret = FAKE_SETTING
	}

	o.SqlSettings.DataSource = FAKE_SETTING
	o.SqlSettings.AtRestEncryptKey = FAKE_SETTING

	for i := range o.SqlSettings.DataSourceReplicas {
		o.SqlSettings.DataSourceReplicas[i] = FAKE_SETTING
	}
}

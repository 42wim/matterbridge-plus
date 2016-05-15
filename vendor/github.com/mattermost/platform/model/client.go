// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"bytes"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	HEADER_REQUEST_ID         = "X-Request-ID"
	HEADER_VERSION_ID         = "X-Version-ID"
	HEADER_ETAG_SERVER        = "ETag"
	HEADER_ETAG_CLIENT        = "If-None-Match"
	HEADER_FORWARDED          = "X-Forwarded-For"
	HEADER_REAL_IP            = "X-Real-IP"
	HEADER_FORWARDED_PROTO    = "X-Forwarded-Proto"
	HEADER_TOKEN              = "token"
	HEADER_BEARER             = "BEARER"
	HEADER_AUTH               = "Authorization"
	HEADER_REQUESTED_WITH     = "X-Requested-With"
	HEADER_REQUESTED_WITH_XML = "XMLHttpRequest"

	API_URL_SUFFIX_V1 = "/api/v1"
	API_URL_SUFFIX_V3 = "/api/v3"
	API_URL_SUFFIX    = API_URL_SUFFIX_V3
)

type Result struct {
	RequestId string
	Etag      string
	Data      interface{}
}

type Client struct {
	Url        string       // The location of the server like "http://localhost:8065"
	ApiUrl     string       // The api location of the server like "http://localhost:8065/api/v3"
	HttpClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	TeamId     string
}

// NewClient constructs a new client with convienence methods for talking to
// the server.
func NewClient(url string) *Client {
	return &Client{url, url + API_URL_SUFFIX, &http.Client{}, "", "", ""}
}

func (c *Client) SetOAuthToken(token string) {
	c.AuthToken = token
	c.AuthType = HEADER_TOKEN
}

func (c *Client) ClearOAuthToken() {
	c.AuthToken = ""
	c.AuthType = HEADER_BEARER
}

func (c *Client) SetTeamId(teamId string) {
	c.TeamId = teamId
}

func (c *Client) GetTeamId() string {
	if len(c.TeamId) == 0 {
		println(`You are trying to use a route that requires a team_id, 
        	but you have not called SetTeamId() in client.go`)
	}

	return c.TeamId
}

func (c *Client) ClearTeamId() {
	c.TeamId = ""
}

func (c *Client) GetTeamRoute() string {
	return fmt.Sprintf("/teams/%v", c.GetTeamId())
}

func (c *Client) GetChannelRoute(channelId string) string {
	return fmt.Sprintf("/teams/%v/channels/%v", c.GetTeamId(), channelId)
}

func (c *Client) GetChannelNameRoute(channelName string) string {
	return fmt.Sprintf("/teams/%v/channels/name/%v", c.GetTeamId(), channelName)
}

func (c *Client) DoPost(url, data, contentType string) (*http.Response, *AppError) {
	rq, _ := http.NewRequest("POST", c.Url+url, strings.NewReader(data))
	rq.Header.Set("Content-Type", contentType)

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return rp, nil
	}
}

func (c *Client) DoApiPost(url string, data string) (*http.Response, *AppError) {
	rq, _ := http.NewRequest("POST", c.ApiUrl+url, strings.NewReader(data))

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return rp, nil
	}
}

func (c *Client) DoApiGet(url string, data string, etag string) (*http.Response, *AppError) {
	rq, _ := http.NewRequest("GET", c.ApiUrl+url, strings.NewReader(data))

	if len(etag) > 0 {
		rq.Header.Set(HEADER_ETAG_CLIENT, etag)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode == 304 {
		return rp, nil
	} else if rp.StatusCode >= 300 {
		return rp, AppErrorFromJson(rp.Body)
	} else {
		return rp, nil
	}
}

func getCookie(name string, resp *http.Response) *http.Cookie {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

func (c *Client) Must(result *Result, err *AppError) *Result {
	if err != nil {
		l4g.Close()
		time.Sleep(time.Second)
		panic(err)
	}

	return result
}

func (c *Client) SignupTeam(email string, displayName string) (*Result, *AppError) {
	m := make(map[string]string)
	m["email"] = email
	m["display_name"] = displayName
	if r, err := c.DoApiPost("/teams/signup", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateTeamFromSignup(teamSignup *TeamSignup) (*Result, *AppError) {
	if r, err := c.DoApiPost("/teams/create_from_signup", teamSignup.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamSignupFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateTeam(team *Team) (*Result, *AppError) {
	if r, err := c.DoApiPost("/teams/create", team.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAllTeams() (*Result, *AppError) {
	if r, err := c.DoApiGet("/teams/all", "", ""); err != nil {
		return nil, err
	} else {

		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamMapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAllTeamListings() (*Result, *AppError) {
	if r, err := c.DoApiGet("/teams/all_team_listings", "", ""); err != nil {
		return nil, err
	} else {

		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamMapFromJson(r.Body)}, nil
	}
}

func (c *Client) FindTeamByName(name string) (*Result, *AppError) {
	m := make(map[string]string)
	m["name"] = name
	if r, err := c.DoApiPost("/teams/find_team_by_name", MapToJson(m)); err != nil {
		return nil, err
	} else {
		val := false
		if body, _ := ioutil.ReadAll(r.Body); string(body) == "true" {
			val = true
		}

		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), val}, nil
	}
}

func (c *Client) AddUserToTeam(userId string) (*Result, *AppError) {
	data := make(map[string]string)
	data["user_id"] = userId
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/add_user_to_team", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) AddUserToTeamFromInvite(hash, dataToHash, inviteId string) (*Result, *AppError) {
	data := make(map[string]string)
	data["hash"] = hash
	data["data"] = dataToHash
	data["invite_id"] = inviteId
	if r, err := c.DoApiPost("/teams/add_user_to_team_from_invite", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamFromJson(r.Body)}, nil
	}
}

func (c *Client) InviteMembers(invites *Invites) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/invite_members", invites.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), InvitesFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateTeam(team *Team) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/update", team.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateUser(user *User, hash string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/create", user.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateUserWithInvite(user *User, hash string, data string, inviteId string) (*Result, *AppError) {

	url := "/users/create?d=" + url.QueryEscape(data) + "&h=" + url.QueryEscape(hash) + "&iid=" + url.QueryEscape(inviteId)

	if r, err := c.DoApiPost(url, user.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateUserFromSignup(user *User, data string, hash string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/create?d="+url.QueryEscape(data)+"&h="+hash, user.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) GetUser(id string, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/"+id+"/get", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) GetMe(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/me", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) GetProfilesForDirectMessageList(teamId string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/profiles_for_dm_list/"+teamId, "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserMapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetProfiles(teamId string, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/profiles/"+teamId, "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserMapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetDirectProfiles(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/direct_profiles", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserMapFromJson(r.Body)}, nil
	}
}

func (c *Client) LoginById(id string, password string) (*Result, *AppError) {
	m := make(map[string]string)
	m["id"] = id
	m["password"] = password
	return c.login(m)
}

func (c *Client) Login(loginId string, password string) (*Result, *AppError) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	return c.login(m)
}

func (c *Client) LoginByLdap(loginId string, password string) (*Result, *AppError) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["ldap_only"] = "true"
	return c.login(m)
}

func (c *Client) LoginWithDevice(loginId string, password string, deviceId string) (*Result, *AppError) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["device_id"] = deviceId
	return c.login(m)
}

func (c *Client) login(m map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/login", MapToJson(m)); err != nil {
		return nil, err
	} else {
		c.AuthToken = r.Header.Get(HEADER_TOKEN)
		c.AuthType = HEADER_BEARER
		sessionToken := getCookie(SESSION_COOKIE_TOKEN, r)

		if c.AuthToken != sessionToken.Value {
			NewLocAppError("/users/login", "model.client.login.app_error", nil, "")
		}

		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) Logout() (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/logout", ""); err != nil {
		return nil, err
	} else {
		c.AuthToken = ""
		c.AuthType = HEADER_BEARER
		c.TeamId = ""

		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) CheckMfa(loginId string) (*Result, *AppError) {
	m := make(map[string]string)
	m["login_id"] = loginId

	if r, err := c.DoApiPost("/users/mfa", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GenerateMfaQrCode() (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/generate_mfa_qr", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), r.Body}, nil
	}
}

func (c *Client) UpdateMfa(activate bool, token string) (*Result, *AppError) {
	m := make(map[string]interface{})
	m["activate"] = activate
	m["token"] = token

	if r, err := c.DoApiPost("/users/update_mfa", StringInterfaceToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) AdminResetMfa(userId string) (*Result, *AppError) {
	m := make(map[string]string)
	m["user_id"] = userId

	if r, err := c.DoApiPost("/admin/reset_mfa", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) RevokeSession(sessionAltId string) (*Result, *AppError) {
	m := make(map[string]string)
	m["id"] = sessionAltId

	if r, err := c.DoApiPost("/users/revoke_session", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetSessions(id string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/"+id+"/sessions", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), SessionsFromJson(r.Body)}, nil
	}
}

func (c *Client) EmailToOAuth(m map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/claim/email_to_sso", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) OAuthToEmail(m map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/claim/oauth_to_email", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) LDAPToEmail(m map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/claim/ldap_to_email", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) EmailToLDAP(m map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/claim/ldap_to_email", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) Command(channelId string, command string, suggest bool) (*Result, *AppError) {
	m := make(map[string]string)
	m["command"] = command
	m["channelId"] = channelId
	m["suggest"] = strconv.FormatBool(suggest)
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/commands/execute", MapToJson(m)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CommandResponseFromJson(r.Body)}, nil
	}
}

func (c *Client) ListCommands() (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/commands/list", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CommandListFromJson(r.Body)}, nil
	}
}

func (c *Client) ListTeamCommands() (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/commands/list_team_commands", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CommandListFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateCommand(cmd *Command) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/commands/create", cmd.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CommandFromJson(r.Body)}, nil
	}
}

func (c *Client) RegenCommandToken(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/commands/regen_token", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CommandFromJson(r.Body)}, nil
	}
}

func (c *Client) DeleteCommand(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/commands/delete", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAudits(id string, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/"+id+"/audits", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), AuditsFromJson(r.Body)}, nil
	}
}

func (c *Client) GetLogs() (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/logs", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ArrayFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAllAudits() (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/audits", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), AuditsFromJson(r.Body)}, nil
	}
}

func (c *Client) GetClientProperties() (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/client_props", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetConfig() (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/config", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ConfigFromJson(r.Body)}, nil
	}
}

func (c *Client) SaveConfig(config *Config) (*Result, *AppError) {
	if r, err := c.DoApiPost("/admin/save_config", config.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) TestEmail(config *Config) (*Result, *AppError) {
	if r, err := c.DoApiPost("/admin/test_email", config.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetComplianceReports() (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/compliance_reports", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), CompliancesFromJson(r.Body)}, nil
	}
}

func (c *Client) SaveComplianceReport(job *Compliance) (*Result, *AppError) {
	if r, err := c.DoApiPost("/admin/save_compliance_report", job.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ComplianceFromJson(r.Body)}, nil
	}
}

func (c *Client) DownloadComplianceReport(id string) (*Result, *AppError) {
	var rq *http.Request
	rq, _ = http.NewRequest("GET", c.ApiUrl+"/admin/download_compliance_report/"+id, nil)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, "BEARER "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError("/admin/download_compliance_report", "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return &Result{rp.Header.Get(HEADER_REQUEST_ID),
			rp.Header.Get(HEADER_ETAG_SERVER), rp.Body}, nil
	}
}

func (c *Client) GetTeamAnalytics(teamId, name string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/analytics/"+teamId+"/"+name, "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), AnalyticsRowsFromJson(r.Body)}, nil
	}
}

func (c *Client) GetSystemAnalytics(name string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/admin/analytics/"+name, "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), AnalyticsRowsFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateChannel(channel *Channel) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/create", channel.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateDirectChannel(userId string) (*Result, *AppError) {
	data := make(map[string]string)
	data["user_id"] = userId
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/create_direct", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateChannel(channel *Channel) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/update", channel.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateChannelHeader(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/update_header", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateChannelPurpose(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/update_purpose", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateNotifyProps(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/channels/update_notify_props", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetChannels(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/channels/", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetChannel(id, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(id)+"/", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelDataFromJson(r.Body)}, nil
	}
}

func (c *Client) GetMoreChannels(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/channels/more", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetChannelCounts(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/channels/counts", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelCountsFromJson(r.Body)}, nil
	}
}

func (c *Client) JoinChannel(id string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(id)+"/join", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) JoinChannelByName(name string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelNameRoute(name)+"/join", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) LeaveChannel(id string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(id)+"/leave", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) DeleteChannel(id string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(id)+"/delete", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) AddChannelMember(id, user_id string) (*Result, *AppError) {
	data := make(map[string]string)
	data["user_id"] = user_id
	if r, err := c.DoApiPost(c.GetChannelRoute(id)+"/add", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) RemoveChannelMember(id, user_id string) (*Result, *AppError) {
	data := make(map[string]string)
	data["user_id"] = user_id
	if r, err := c.DoApiPost(c.GetChannelRoute(id)+"/remove", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) UpdateLastViewedAt(channelId string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(channelId)+"/update_last_viewed_at", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) GetChannelExtraInfo(id string, memberLimit int, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(id)+"/extra_info/"+strconv.FormatInt(int64(memberLimit), 10), "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), ChannelExtraFromJson(r.Body)}, nil
	}
}

func (c *Client) CreatePost(post *Post) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(post.ChannelId)+"/posts/create", post.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdatePost(post *Post) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(post.ChannelId)+"/posts/update", post.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPosts(channelId string, offset int, limit int, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(channelId)+fmt.Sprintf("/posts/page/%v/%v", offset, limit), "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPostsSince(channelId string, time int64) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(channelId)+fmt.Sprintf("/posts/since/%v", time), "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPostsBefore(channelId string, postid string, offset int, limit int, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(channelId)+fmt.Sprintf("/posts/%v/before/%v/%v", postid, offset, limit), "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPostsAfter(channelId string, postid string, offset int, limit int, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(fmt.Sprintf(c.GetChannelRoute(channelId)+"/posts/%v/after/%v/%v", postid, offset, limit), "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPost(channelId string, postId string, etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetChannelRoute(channelId)+fmt.Sprintf("/posts/%v/get", postId), "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) DeletePost(channelId string, postId string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetChannelRoute(channelId)+fmt.Sprintf("/posts/%v/delete", postId), ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) SearchPosts(terms string, isOrSearch bool) (*Result, *AppError) {
	data := map[string]interface{}{}
	data["terms"] = terms
	data["is_or_search"] = isOrSearch
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/posts/search", StringInterfaceToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), PostListFromJson(r.Body)}, nil
	}
}

func (c *Client) UploadProfileFile(data []byte, contentType string) (*Result, *AppError) {
	return c.uploadFile(c.ApiUrl+"/users/newimage", data, contentType)
}

func (c *Client) UploadPostAttachment(data []byte, contentType string) (*Result, *AppError) {
	return c.uploadFile(c.ApiUrl+c.GetTeamRoute()+"/files/upload", data, contentType)
}

func (c *Client) uploadFile(url string, data []byte, contentType string) (*Result, *AppError) {
	rq, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	rq.Header.Set("Content-Type", contentType)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, "BEARER "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return &Result{rp.Header.Get(HEADER_REQUEST_ID),
			rp.Header.Get(HEADER_ETAG_SERVER), FileUploadResponseFromJson(rp.Body)}, nil
	}
}

func (c *Client) GetFile(url string, isFullUrl bool) (*Result, *AppError) {
	var rq *http.Request
	if isFullUrl {
		rq, _ = http.NewRequest("GET", url, nil)
	} else {
		rq, _ = http.NewRequest("GET", c.ApiUrl+c.GetTeamRoute()+"/files/get"+url, nil)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, "BEARER "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return &Result{rp.Header.Get(HEADER_REQUEST_ID),
			rp.Header.Get(HEADER_ETAG_SERVER), rp.Body}, nil
	}
}

func (c *Client) GetFileInfo(url string) (*Result, *AppError) {
	var rq *http.Request
	rq, _ = http.NewRequest("GET", c.ApiUrl+c.GetTeamRoute()+"/files/get_info"+url, nil)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, "BEARER "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil {
		return nil, NewLocAppError(url, "model.client.connecting.app_error", nil, err.Error())
	} else if rp.StatusCode >= 300 {
		return nil, AppErrorFromJson(rp.Body)
	} else {
		return &Result{rp.Header.Get(HEADER_REQUEST_ID),
			rp.Header.Get(HEADER_ETAG_SERVER), FileInfoFromJson(rp.Body)}, nil
	}
}

func (c *Client) GetPublicLink(filename string) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/files/get_public_link", MapToJson(map[string]string{"filename": filename})); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), StringFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateUser(user *User) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/update", user.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateUserRoles(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/update_roles", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) AttachDeviceId(deviceId string) (*Result, *AppError) {
	data := make(map[string]string)
	data["device_id"] = deviceId
	if r, err := c.DoApiPost("/users/attach_device", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateActive(userId string, active bool) (*Result, *AppError) {
	data := make(map[string]string)
	data["user_id"] = userId
	data["active"] = strconv.FormatBool(active)
	if r, err := c.DoApiPost("/users/update_active", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateUserNotify(data map[string]string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/update_notify", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), UserFromJson(r.Body)}, nil
	}
}

func (c *Client) UpdateUserPassword(userId, currentPassword, newPassword string) (*Result, *AppError) {
	data := make(map[string]string)
	data["current_password"] = currentPassword
	data["new_password"] = newPassword
	data["user_id"] = userId

	if r, err := c.DoApiPost("/users/newpassword", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) SendPasswordReset(email string) (*Result, *AppError) {
	data := map[string]string{}
	data["email"] = email
	if r, err := c.DoApiPost("/users/send_password_reset", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) ResetPassword(code, newPassword string) (*Result, *AppError) {
	data := map[string]string{}
	data["code"] = code
	data["new_password"] = newPassword
	if r, err := c.DoApiPost("/users/reset_password", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) AdminResetPassword(userId, newPassword string) (*Result, *AppError) {
	data := map[string]string{}
	data["user_id"] = userId
	data["new_password"] = newPassword
	if r, err := c.DoApiPost("/admin/reset_password", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetStatuses(data []string) (*Result, *AppError) {
	if r, err := c.DoApiPost("/users/status", ArrayToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetMyTeam(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/me", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamFromJson(r.Body)}, nil
	}
}

func (c *Client) GetTeamMembers(teamId string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/teams/members/"+teamId, "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), TeamMembersFromJson(r.Body)}, nil
	}
}

func (c *Client) RegisterApp(app *OAuthApp) (*Result, *AppError) {
	if r, err := c.DoApiPost("/oauth/register", app.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), OAuthAppFromJson(r.Body)}, nil
	}
}

func (c *Client) AllowOAuth(rspType, clientId, redirect, scope, state string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/oauth/allow?response_type="+rspType+"&client_id="+clientId+"&redirect_uri="+url.QueryEscape(redirect)+"&scope="+scope+"&state="+url.QueryEscape(state), "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAccessToken(data url.Values) (*Result, *AppError) {
	if r, err := c.DoApiPost("/oauth/access_token", data.Encode()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), AccessResponseFromJson(r.Body)}, nil
	}
}

func (c *Client) CreateIncomingWebhook(hook *IncomingWebhook) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/hooks/incoming/create", hook.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), IncomingWebhookFromJson(r.Body)}, nil
	}
}

func (c *Client) PostToWebhook(id, payload string) (*Result, *AppError) {
	if r, err := c.DoPost("/hooks/"+id, payload, "application/x-www-form-urlencoded"); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), nil}, nil
	}
}

func (c *Client) DeleteIncomingWebhook(id string) (*Result, *AppError) {
	data := make(map[string]string)
	data["id"] = id
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/hooks/incoming/delete", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) ListIncomingWebhooks() (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/hooks/incoming/list", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), IncomingWebhookListFromJson(r.Body)}, nil
	}
}

func (c *Client) GetAllPreferences() (*Result, *AppError) {
	if r, err := c.DoApiGet("/preferences/", "", ""); err != nil {
		return nil, err
	} else {
		preferences, _ := PreferencesFromJson(r.Body)
		return &Result{r.Header.Get(HEADER_REQUEST_ID), r.Header.Get(HEADER_ETAG_SERVER), preferences}, nil
	}
}

func (c *Client) SetPreferences(preferences *Preferences) (*Result, *AppError) {
	if r, err := c.DoApiPost("/preferences/save", preferences.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), preferences}, nil
	}
}

func (c *Client) GetPreference(category string, name string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/preferences/"+category+"/"+name, "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID), r.Header.Get(HEADER_ETAG_SERVER), PreferenceFromJson(r.Body)}, nil
	}
}

func (c *Client) GetPreferenceCategory(category string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/preferences/"+category, "", ""); err != nil {
		return nil, err
	} else {
		preferences, _ := PreferencesFromJson(r.Body)
		return &Result{r.Header.Get(HEADER_REQUEST_ID), r.Header.Get(HEADER_ETAG_SERVER), preferences}, nil
	}
}

func (c *Client) CreateOutgoingWebhook(hook *OutgoingWebhook) (*Result, *AppError) {
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/hooks/outgoing/create", hook.ToJson()); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), OutgoingWebhookFromJson(r.Body)}, nil
	}
}

func (c *Client) DeleteOutgoingWebhook(id string) (*Result, *AppError) {
	data := make(map[string]string)
	data["id"] = id
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/hooks/outgoing/delete", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) ListOutgoingWebhooks() (*Result, *AppError) {
	if r, err := c.DoApiGet(c.GetTeamRoute()+"/hooks/outgoing/list", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), OutgoingWebhookListFromJson(r.Body)}, nil
	}
}

func (c *Client) RegenOutgoingWebhookToken(id string) (*Result, *AppError) {
	data := make(map[string]string)
	data["id"] = id
	if r, err := c.DoApiPost(c.GetTeamRoute()+"/hooks/outgoing/regen_token", MapToJson(data)); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), OutgoingWebhookFromJson(r.Body)}, nil
	}
}

func (c *Client) MockSession(sessionToken string) {
	c.AuthToken = sessionToken
	c.AuthType = HEADER_BEARER
}

func (c *Client) GetClientLicenceConfig(etag string) (*Result, *AppError) {
	if r, err := c.DoApiGet("/license/client_config", "", etag); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), MapFromJson(r.Body)}, nil
	}
}

func (c *Client) GetInitialLoad() (*Result, *AppError) {
	if r, err := c.DoApiGet("/users/initial_load", "", ""); err != nil {
		return nil, err
	} else {
		return &Result{r.Header.Get(HEADER_REQUEST_ID),
			r.Header.Get(HEADER_ETAG_SERVER), InitialLoadFromJson(r.Body)}, nil
	}
}

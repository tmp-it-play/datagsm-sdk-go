package oauth

import datagsm "github.com/tmp-it-play/datagsm-sdk-go"

const (
	ScopeOpenID      = "openid"
	ScopeAccountRead = "datagsm:account_read"
	ScopeStudentRead = "datagsm:student_read"
	ScopeClubRead    = "datagsm:club_read"
	ScopeProjectRead = "datagsm:project_read"
)

type Token struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int64   `json:"expires_in"`
	RefreshToken *string `json:"refresh_token"`
	Scope        string  `json:"scope"`
	IDToken      *string `json:"id_token"`
}

type UserInfo struct {
	Subject    string                `json:"sub"`
	ID         int64                 `json:"id"`
	Email      string                `json:"email"`
	Role       string                `json:"role"`
	Status     string                `json:"status"`
	ObjectType *string               `json:"objectType"`
	Student    *datagsm.Student      `json:"student"`
	Teacher    *Teacher              `json:"teacher"`
	Clubs      []datagsm.ClubSummary `json:"clubs"`
	Projects   []ProjectSummary      `json:"projects"`
}

type Teacher struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Department  string  `json:"department"`
	Description *string `json:"description"`
}

type ProjectSummary struct {
	ID     int64                 `json:"id"`
	Name   string                `json:"name"`
	Status datagsm.ProjectStatus `json:"status"`
	Club   *datagsm.ClubSummary  `json:"club"`
}

type Discovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

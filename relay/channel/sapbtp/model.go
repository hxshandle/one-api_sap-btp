package sapbtp

import (
	"time"
)

type SAPBTPConfiguration struct {
	URL    string          `json:"url"`
	UAA    SAPBTPUAAConfig `json:"uaa"`
	Vendor string          `json:"vendor"`
}

type SAPBTPUAAConfig struct {
	ClientID        string `json:"clientid"`
	ClientSecret    string `json:"clientsecret"`
	URL             string `json:"url"`
	IdentityZone    string `json:"identityzone"`
	IdentityZoneID  string `json:"identityzoneid"`
	TenantID        string `json:"tenantid"`
	TenantMode      string `json:"tenantmode"`
	SBURL           string `json:"sburl"`
	APIURL          string `json:"apiurl"`
	VerificationKey string `json:"verificationkey"`
	Xsappname       string `json:"xsappname"`
	SubaccountID    string `json:"subaccountid"`
	UAADomain       string `json:"uaadomain"`
	ZoneID          string `json:"zoneid"`
	CredentialType  string `json:"credential-type"`
}

type tokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	JTI         string `json:"jti"`
	ExpiryTime  time.Time
}

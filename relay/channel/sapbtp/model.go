package sapbtp

import (
	"one-api/relay/channel/openai"
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

type ChatRequest struct {
	DeploymentID     string                 `json:"deployment_id"`
	Messages         []openai.Message       `json:"messages"`
	Prompt           any                    `json:"prompt,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	N                int                    `json:"n,omitempty"`
	Input            any                    `json:"input,omitempty"`
	Instruction      string                 `json:"instruction,omitempty"`
	Size             string                 `json:"size,omitempty"`
	Functions        any                    `json:"functions,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64                `json:"presence_penalty,omitempty"`
	ResponseFormat   *openai.ResponseFormat `json:"response_format,omitempty"`
	Seed             float64                `json:"seed,omitempty"`
	Tools            any                    `json:"tools,omitempty"`
	ToolChoice       any                    `json:"tool_choice,omitempty"`
	User             string                 `json:"user,omitempty"`
}

package flashclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var UserAgentVersion string = "development"

var FARestUserAgentBase string = "Dev_Pure_FA_OpenMetrics_exporter"

var FARestUserAgent string = FARestUserAgentBase + "/" + UserAgentVersion

type FAClient struct {
	apiToken   string
	RestClient *resty.Client

	endPoint   string
	apiVersion string
	insecure   bool
	xAuthToken *Token
	xRequestID string
	tokenLock  sync.Mutex
}

type Token struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (t *Token) IsExpired() bool {
	return time.Until(t.Expires) < 0
}

func (t *Token) ExpiresSoon() bool {
	return time.Until(t.Expires) < 30*time.Second
}

func (fa *FAClient) RefreshSession() error {
	fa.tokenLock.Lock()
	defer fa.tokenLock.Unlock()
	if fa.xAuthToken == nil || fa.xAuthToken.ExpiresSoon() {
		// Get new token
		type ResponseBody struct {
			Items []struct {
				Username string `json:"username"`
			} `json:"items"`
		}
		// result := new(ResponseBody)

		res, err := resty.New().
			SetBaseURL(fa.endPoint + "/api/" + fa.apiVersion).
			SetTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: fa.insecure}}).
			R().
			SetHeaders(map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
				"X-Request-ID": fa.xRequestID,
				"api-token":    fa.apiToken,
			}).
			// SetResult(&result).
			Post("/login")
		if err != nil {
			return fmt.Errorf("failed to login, error: %v, response: %s", err, res.String())
		} else if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("failed to login, status code: %d, response: %s", res.StatusCode(), res.String())
		}

		xAuthToken := res.Header().Get("x-auth-token")
		if xAuthToken == "" {
			return fmt.Errorf("login response missing x-auth-token header")
		}

		fa.xAuthToken = &Token{
			Token:   xAuthToken,
			Expires: time.Now().Add(time.Duration(5) * time.Minute),
		}
		fa.RestClient.SetHeader("x-auth-token", fa.xAuthToken.Token)
	}
	return nil
}

func GetAPISupportedVersions(endpoint string, insecure bool) (ApiVersions, error) {
	endpoint = CleanURI(endpoint)

	type ResponseBody struct {
		Versions []string `json:"version"`
	}

	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: insecure})

	result := new(ResponseBody)
	res, err := client.R().
		SetResult(&result).
		Get(endpoint + "/api/api_version")
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("not a valid FlashArray REST API server")
	} else if len(result.Versions) == 0 {
		return nil, fmt.Errorf("not a valid FlashArray REST API version: %v", result.Versions)
	}
	return result.Versions, nil
}

func (fa *FAClient) GetVersions() (ApiVersions, error) {
	return GetAPISupportedVersions(fa.endPoint, fa.insecure)
}

type ClientConfig struct {
	ApiVersion string
	UserAgent  string
	RequestID  string
	Debug      bool
	Insecure   bool
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		ApiVersion: "latest",
		UserAgent:  FARestUserAgent,
		RequestID:  "",
		Debug:      false,
		Insecure:   false,
	}
}

func NewRestClient(endpoint string, apitoken string, config ...ClientConfig) (*FAClient, error) {
	var cfg ClientConfig
	if len(config) == 0 {
		cfg = DefaultClientConfig()
	} else if len(config) > 1 {
		return nil, fmt.Errorf("only one ClientConfig can be provided")
	} else {
		cfg = config[0]
	}

	endpoint = CleanURI(endpoint)

	apiVersion := ""
	if apiVersions, err := GetAPISupportedVersions(endpoint, cfg.Insecure); err != nil {
		return nil, fmt.Errorf("failed to get supported API versions: %w", err)
	} else if strings.EqualFold(cfg.ApiVersion, "latest") {
		apiVersion = apiVersions.Latest()
	} else if !apiVersions.Contains(cfg.ApiVersion) {
		return nil, fmt.Errorf("API version %s is not supported by the FlashArray, supported versions are: %v", cfg.ApiVersion, apiVersions)
	} else {
		apiVersion = cfg.ApiVersion
	}

	if VersionLessThan(apiVersion, "2.0") {
		return nil, fmt.Errorf("API version %s is not supported by the FlashArray, minimum supported version is 2.0", apiVersion)
	}

	client := resty.New()
	client.SetBaseURL(endpoint + "/api/" + apiVersion)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.Insecure})
	client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"X-Request-ID": cfg.RequestID,
		"User-Agent":   FARestUserAgent + " (" + cfg.UserAgent + ")",
	})

	if cfg.Debug {
		client.SetDebug(true)
	}

	fa := &FAClient{
		endPoint:   endpoint,
		apiToken:   apitoken,
		apiVersion: apiVersion,
		xRequestID: cfg.RequestID,
		RestClient: client,
		insecure:   cfg.Insecure,
	}

	return fa, nil
}

func (fa *FAClient) Close() error {
	fa.tokenLock.Lock()
	defer fa.tokenLock.Unlock()
	if fa.xAuthToken == nil {
		return nil
	}
	resp, err := fa.RestClient.R().
		SetHeader("X-Request-ID", fa.xRequestID).
		SetHeader("x-auth-token", fa.xAuthToken.Token).
		Post("/logout")
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to logout, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

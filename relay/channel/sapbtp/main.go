package sapbtp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/common/logger"
	"one-api/model"
	"one-api/relay/channel/openai"
	"one-api/relay/constant"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var btpTokens sync.Map
var btpFullURLs sync.Map

func ConvertRequest(request openai.GeneralOpenAIRequest) *ChatRequest {
	// copy all the attribute from request to ChatRequest

	return &ChatRequest{
		DeploymentID:     request.Model,
		Messages:         request.Messages,
		Prompt:           request.Prompt,
		Stream:           request.Stream,
		MaxTokens:        request.MaxTokens,
		Temperature:      request.Temperature,
		TopP:             request.TopP,
		N:                request.N,
		Input:            request.Input,
		Instruction:      request.Instruction,
		Size:             request.Size,
		Functions:        request.Functions,
		FrequencyPenalty: request.FrequencyPenalty,
		PresencePenalty:  request.PresencePenalty,
		ResponseFormat:   request.ResponseFormat,
		Seed:             request.Seed,
		Tools:            request.Tools,
		ToolChoice:       request.ToolChoice,
		User:             request.User,
	}
}

func buildAuthRequest(channel *model.Channel) (token *tokenData, err error) {
	cfg := GetSAPBTPConfig(channel)
	authUrl := cfg.UAA.URL + "/oauth/token"
	clientId := cfg.UAA.ClientID
	clientSecret := cfg.UAA.ClientSecret
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", "client_credentials")
	auth := base64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))
	req, err := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to create auth request for channel %d, error: %s", channel.Id, err.Error()))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+auth)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to send auth request for channel %d, error: %s", channel.Id, err.Error()))
	}
	defer resp.Body.Close()

	// 解析请求
	var tokenData tokenData
	// err = json.Unmarshal(body, &tokenData)
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to decode auth response for channel %d, error: %s", channel.Id, err.Error()))
	}
	tokenData.ExpiryTime = time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)
	return &tokenData, nil
}

func GetToken(channel *model.Channel) string {
	data, ok := btpTokens.Load(channel.Key)
	if ok {
		tkData := data.(*tokenData)
		if time.Now().Before(tkData.ExpiryTime) {
			return tkData.AccessToken
		}
	}
	// TODO get token from BTP
	tk, err := buildAuthRequest(channel)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to get token for channel %d, error: %s", channel.Id, err.Error()))
	}
	btpTokens.Store(channel.Key, tk)
	return tk.AccessToken
}

func GetSAPBTPConfig(channel *model.Channel) *SAPBTPConfiguration {
	// 这里将channel.Other从String转换成SAPBTPConfiguration对象
	var sapbtpConfig SAPBTPConfiguration
	err := json.Unmarshal([]byte(channel.Other), &sapbtpConfig)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal sapbtp config for channel %d, error: %s", channel.Id, err.Error()))
	}
	return &sapbtpConfig

}

func GetSAPBTPFullRequestURLByChannelId(channelId int) string {
	data, ok := btpFullURLs.Load(channelId)
	if ok {
		return data.(string)
	}
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to get channel by id %d, error: %s", channelId, err.Error()))
	}
	cfg := GetSAPBTPConfig(channel)
	btpFullURLs.Store(channelId, cfg.URL)
	return cfg.URL
}

func StreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*openai.ErrorWithStatusCode, string) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			dataChan <- data
			data = data[6:]
			relayMode = 1
			if !strings.HasPrefix(data, "[DONE]") {
				switch relayMode {
				case constant.RelayModeChatCompletions:
					var streamResponse openai.ChatCompletionsStreamResponse
					err := json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						logger.SysError("error unmarshalling stream response: " + err.Error())
						continue // just ignore the error
					}
					for _, choice := range streamResponse.Choices {
						responseText += choice.Delta.Content
					}
				case constant.RelayModeCompletions:
					var streamResponse openai.CompletionsStreamResponse
					err := json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						logger.SysError("error unmarshalling stream response: " + err.Error())
						continue
					}
					for _, choice := range streamResponse.Choices {
						responseText += choice.Text
					}
				}
			}
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if strings.HasPrefix(data, "data: [DONE]") {
				data = data[:12]
			}
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			c.Render(-1, common.CustomEvent{Data: data})
			return true
		case <-stopChan:
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}
func Handler(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage) {
	var chatResponse openai.TextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	fmt.Println(string(responseBody))
	err = json.Unmarshal(responseBody, &chatResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	jsonResponse, err := json.Marshal(chatResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &chatResponse.Usage
}

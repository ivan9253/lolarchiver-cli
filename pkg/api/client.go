package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	baseURL = "https://api.lolarchiver.com"
)

// Client represents the API client
type Client struct {
	apiKey string
	client *http.Client
}

// NewClient creates a new API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Request represents a generic API request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
}

// Response represents a generic API response
type Response struct {
	StatusCode int
	Body       []byte
}

// Do performs an API request
func (c *Client) Do(req Request) (*Response, error) {
	// Start spinner on stderr
	done := make(chan bool)
	go func() {
		startTime := time.Now()
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Fprintf(os.Stderr, "\rProcessing (%v)...", elapsed)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			done <- true
			fmt.Fprint(os.Stderr, "\r")
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	httpReq, err := http.NewRequest(req.Method, baseURL+req.Path, bodyReader)
	if err != nil {
		done <- true
		fmt.Fprint(os.Stderr, "\r")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", c.apiKey)

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		done <- true
		fmt.Fprint(os.Stderr, "\r")
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		done <- true
		fmt.Fprint(os.Stderr, "\r")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Stop spinner and return
	result := &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
	}
	done <- true
	fmt.Fprint(os.Stderr, "\r")
	return result, nil
}

// CheckCredits checks the remaining API credits
func (c *Client) CheckCredits() (*Response, error) {
	return c.Do(Request{
		Method: "POST",
		Path:   "/credits_left",
	})
}

// YouTubeUserComments retrieves all comments from a YouTube user
func (c *Client) YouTubeUserComments(userID, handle, channelID string, offset int) (*Response, error) {
	body := map[string]interface{}{
		"offset": offset,
	}
	if userID != "" {
		body["user_id"] = userID
	}
	if handle != "" {
		body["handle"] = handle
	}
	if channelID != "" {
		body["channel_id"] = channelID
	}

	return c.Do(Request{
		Method: "POST",
		Path:   "/youtube/user_all_comments",
		Body:   body,
	})
}

// YouTubeCommentReplies retrieves replies for a specific YouTube comment
func (c *Client) YouTubeCommentReplies(commentID string) (*Response, error) {
	return c.Do(Request{
		Method: "POST",
		Path:   "/youtube/comment_replies",
		Body: map[string]string{
			"comment_id": commentID,
		},
	})
}

// ReversePhoneLookup performs a reverse phone lookup
func (c *Client) ReversePhoneLookup(phone string, insecureMode bool) (*Response, error) {
	headers := map[string]string{
		"phone": phone,
	}
	if insecureMode {
		headers["insecuremode"] = "true"
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/reverse_phone_lookup",
		Headers: headers,
	})
}

// ReverseEmailLookup performs a reverse email lookup
func (c *Client) ReverseEmailLookup(email string, insecureMode bool) (*Response, error) {
	headers := map[string]string{
		"email": email,
	}
	if insecureMode {
		headers["insecuremode"] = "true"
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/reverse_email_lookup",
		Headers: headers,
	})
}

// TwitterHistoryLookup retrieves Twitter history
func (c *Client) TwitterHistoryLookup(handle string, id int64, byOld bool) (*Response, error) {
	headers := map[string]string{}
	if handle != "" {
		headers["handle"] = handle
	}
	if id != 0 {
		headers["id"] = fmt.Sprintf("%d", id)
	}
	if byOld {
		headers["byold"] = "true"
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitter_history_lookup",
		Headers: headers,
	})
}

// DatabaseLookup performs a database search
func (c *Client) DatabaseLookup(query string, exact bool) (*Response, error) {
	headers := map[string]string{
		"query": query,
	}
	if exact {
		headers["exact"] = "true"
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/database_lookup",
		Headers: headers,
	})
}

// TwitchUserMessages retrieves all messages from a Twitch user
func (c *Client) TwitchUserMessages(username, server string, offset int) (*Response, error) {
	headers := map[string]string{
		"username": username,
		"server":   server,
		"offset":   fmt.Sprintf("%d", offset),
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitch/user_all_messages",
		Headers: headers,
	})
}

// TwitchUserTimeouts retrieves chat bans/timeouts for a Twitch user
func (c *Client) TwitchUserTimeouts(username string, offset int) (*Response, error) {
	headers := map[string]string{
		"username": username,
		"offset":   fmt.Sprintf("%d", offset),
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitch/user_all_timeouts",
		Headers: headers,
	})
}

// TwitchUserHistory retrieves Twitch user history
func (c *Client) TwitchUserHistory(username, mode string) (*Response, error) {
	headers := map[string]string{
		"username": username,
		"mode":     mode,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitch/user_history",
		Headers: headers,
	})
}

// TwitchFollowage retrieves following list of a Twitch user
func (c *Client) TwitchFollowage(username string) (*Response, error) {
	headers := map[string]string{
		"username": username,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitch/followage",
		Headers: headers,
	})
}

// TwitchFollowers retrieves followers list of a Twitch user
func (c *Client) TwitchFollowers(username string) (*Response, error) {
	headers := map[string]string{
		"username": username,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/twitch/followers",
		Headers: headers,
	})
}

// KickUserMessages retrieves all messages from a Kick user
func (c *Client) KickUserMessages(username string, offset int) (*Response, error) {
	headers := map[string]string{
		"username": username,
		"offset":   fmt.Sprintf("%d", offset),
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/kick/user_all_messages",
		Headers: headers,
	})
}

// KickUserTimeouts retrieves chat bans/timeouts for a Kick user
func (c *Client) KickUserTimeouts(username string) (*Response, error) {
	headers := map[string]string{
		"username": username,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/kick/user_all_timeouts",
		Headers: headers,
	})
}

// KickUserModChannels retrieves channels where a Kick user is a moderator
func (c *Client) KickUserModChannels(username string) (*Response, error) {
	headers := map[string]string{
		"username": username,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/kick/user_channel_mods_in",
		Headers: headers,
	})
}

// KickUserSubscribers retrieves subscribers list of a Kick user
func (c *Client) KickUserSubscribers(username string) (*Response, error) {
	headers := map[string]string{
		"username": username,
	}

	return c.Do(Request{
		Method:  "POST",
		Path:    "/kick/user_subscribers_list",
		Headers: headers,
	})
} 
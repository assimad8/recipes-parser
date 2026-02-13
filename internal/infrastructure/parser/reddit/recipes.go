package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"parser/internal/domain/entities"
	parsers "parser/internal/infrastructure/parser"
	"strings"
	"time"
)

// Typed errors
var (
	ErrNotFound    = errors.New("not found")       // 404
	ErrInvalidURL  = errors.New("invalid url")    // malformed URL
	ErrRateLimited = &RateLimitError{}            // 429
)

// Custom error for rate limiting
type RateLimitError struct{}

func (e *RateLimitError) Error() string { return "rate limited" }

// RedditParser implements Parser interface
type RedditParser struct {
	client *http.Client
}

// NewRedditParser creates a new parser instance
func NewRedditParser() parsers.Parser {
	return &RedditParser{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchRecipes fetches recipes from a subreddit URL
func (r *RedditParser) FetchRecipes(url string, limit int) ([]entities.Recipe, error) {
	if !isValidURL(url) {
		return nil, ErrInvalidURL
	}

	url = fmt.Sprintf("%s.json?limit=%d", url, limit)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Reddit requires a descriptive User-Agent
	req.Header.Set("User-Agent", "MyRecipeParserBot/0.1 by <your_reddit_username>")

	var resp *http.Response
	var lastErr error

	for attempt := range 5 {
		resp, err = r.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("[httpclient] request failed: %w", err)
			// exponential backoff
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200:
			lastErr = nil
			break // success
		case 404:
			return nil, ErrNotFound // permanent failure, do not retry
		case 429:
			wait := time.Duration((attempt+1)*5) * time.Second
			log.Printf("Rate limited by Reddit, sleeping %s\n", wait)
			time.Sleep(wait)
			lastErr = ErrRateLimited
			continue // retry
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty response body")
	}

	return parse(&data)
}

// parse converts Reddit JSON to []Recipe
func parse(data *[]byte) ([]entities.Recipe, error) {
	var listing Listing
	if err := json.Unmarshal(*data, &listing); err != nil {
		return nil, err
	}

	recipes := make([]entities.Recipe, 0)
	for _, post := range listing.Data.Children {
		if post.Kind != "t3" {
			continue
		}
		d := post.Data

		if d.LinkFlairText != "Recipe" {
			continue
		}

		imageURL := d.URLOverride
		if imageURL == "" {
			imageURL = d.URL
		}

		if !isImage(imageURL) {
			continue
		}

		recipes = append(recipes, entities.Recipe{
			ID:        d.ID,
			Title:     d.Title,
			Author:    d.Author,
			CreatedAt: time.Unix(int64(d.CreatedUTC), 0),
			ImageURL:  imageURL,
			Link:      "https://www.reddit.com" + d.Permalink,
		})
	}

	return recipes, nil
}

// isImage checks for valid image extensions
func isImage(url string) bool {
	url = strings.ToLower(url)
	return strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".jpeg") || strings.HasSuffix(url, ".png")
}

// isValidURL checks if the URL is a valid Reddit subreddit or post URL
func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must be http or https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	// Must have a host
	if parsed.Host == "" {
		return false
	}

	// Allow common Reddit hosts
	allowedHosts := []string{"reddit.com", "www.reddit.com", "old.reddit.com"}
	validHost := false
	for _, h := range allowedHosts {
		if strings.EqualFold(parsed.Host, h) {
			validHost = true
			break
		}
	}
	if !validHost {
		return false
	}

	// Path must start with /r/ (subreddit) or /user/ (optional)
	if strings.HasPrefix(parsed.Path, "/r/") || strings.HasPrefix(parsed.Path, "/user/") {
		return true
	}

	return false
}

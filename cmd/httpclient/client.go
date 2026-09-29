package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

// TokenProvider é qualquer coisa que forneça um access token.
// O JWT do seu pacote implementa isso naturalmente.
type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Client é um wrapper do http.Client que injeta o access token
// automaticamente em toda requisição.
type Client struct {
	http     *http.Client
	baseURL  string
	provider TokenProvider
	limiter  *rate.Limiter
}

func (c *Client) Limiter() *rate.Limiter {
	return c.limiter
}

// New cria um Client que injeta o Bearer token automaticamente.
func New(baseURL string, provider TokenProvider, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	// ⚠️ Envolve o Transport DEPOIS de aplicar as options,
	// porque a option pode ter trocado o http.Client.
	c.http.Transport = &authTransport{
		base:     c.http.Transport, // pode ser nil → cai no DefaultTransport
		provider: provider,
	}

	return c
}

// authTransport injeta o Bearer token em toda request.
type authTransport struct {
	base     http.RoundTripper
	provider TokenProvider
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// ✅ Pega o token (renova se necessário).
	// Passa o contexto da request pra que o refresh seja cancelável.
	token, err := t.provider.GetAccessToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// ✅ Clona a request pra não mutar a original.
	// http.RoundTripper não pode modificar a request recebida.
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+token)

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(clone)
}

// internal/httpclient/client.go (continuação)

func (c *Client) Get(ctx context.Context, path string, query url.Values, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, query, headers, nil)
}

func (c *Client) Post(ctx context.Context, path string, query url.Values, headers map[string]string, body any) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, query, headers, body)
}

func (c *Client) Put(ctx context.Context, path string, query url.Values, headers map[string]string, body any) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, path, query, headers, body)
}

func (c *Client) Patch(ctx context.Context, path string, query url.Values, headers map[string]string, body any) (*http.Response, error) {
	return c.do(ctx, http.MethodPatch, path, query, headers, body)
}

func (c *Client) Delete(ctx context.Context, path string, query url.Values, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, path, query, headers, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, query url.Values, headers map[string]string, body any) (*http.Response, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// Aplica headers customizados
	for k, v := range headers {
		if k == "Authorization" {
			continue
		}
		req.Header.Set(k, v)
	}

	if body != nil && headers["Content-Type"] == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait: %w", err)
		}
	}

	return c.http.Do(req)
}

// DoJSON faz a request e decodifica a resposta em `out`.
func (c *Client) DoJSON(ctx context.Context, method string, path string, query url.Values, headers map[string]string, body, out any) error {
	resp, err := c.do(ctx, method, path, query, headers, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

package mcptypebuilder_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	mcptypebuilder "github.com/djpiper28/openapi-to-mcp-server/v1/internal/mcp_type_builder"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestMcpServerBuilder(t *testing.T) {
	var client ClientInterface = &Client{Server: "Testing 123"}
	b := mcptypebuilder.New("test", "v1.0.0", client)
	server, err := b.Build()

	require.NoError(t, err)
	require.NotNil(t, server)
}

type RequestEditorFn func()

type PutApiV1AccessControlConfigJSONRequestBody struct {
	Id         string
	Data       string
	OtherStuff int
}

// An mocked extract of output from oapi-codegen go file
type ClientInterface interface {
	GetApiV1About(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetApiV1AccessControlConfig(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
	PutApiV1AccessControlConfigWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)
	PutApiV1AccessControlConfig(ctx context.Context, body PutApiV1AccessControlConfigJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetApiV1AccessControlPointLives(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetApiV1AccessControlPointLivesId(ctx context.Context, id openapi_types.UUID, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetApiV1AccessControlPoints(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
}

type Client struct {
	Server string
	ClientInterface
}

func (c *Client) GetApiV1About(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) GetApiV1AccessControlConfig(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) PutApiV1AccessControlConfigWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) PutApiV1AccessControlConfig(ctx context.Context, body PutApiV1AccessControlConfigJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) GetApiV1AccessControlPointLives(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) GetApiV1AccessControlPointLivesId(ctx context.Context, id openapi_types.UUID, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (c *Client) GetApiV1AccessControlPoints(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

package mcptypebuilder_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	mcptypebuilder "github.com/djpiper28/openapi-to-mcp-server/v1/lib/mcp_type_builder"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestMcpServerBuilder(t *testing.T) {
	var client ClientInterface = &Client{Server: "Testing 123"}
	b := mcptypebuilder.New("test", "v1.0.0", client, &Mapper{})
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

// Mocked struct mapper, this will be codegened usually
type Mapper struct{}

func (m *Mapper) StructType(key string) (reflect.Type, error) {
	switch key {
	case reflect.ValueOf(PutApiV1AccessControlConfigJSONRequestBody{}).Type().Name():
		return reflect.TypeOf(PutApiV1AccessControlConfigJSONRequestBody{}), nil
	case reflect.ValueOf(openapi_types.UUID{}).Type().Name():
		return reflect.TypeOf(openapi_types.UUID{}), nil
	default:
		return nil, errors.New("Unused type")
	}
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

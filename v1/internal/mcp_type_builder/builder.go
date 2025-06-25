package mcptypebuilder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type client any

type Builder[T client] struct {
	Client  T
	Name    string
	Version string
}

func New[T client](Name, Version string, Client T) *Builder[T] {
	return &Builder[T]{
		Client:  Client,
		Name:    Name,
		Version: Version,
	}
}

var (
	// An HTTP call with no arguments/body
	simpleCall = regexp.MustCompile("func\\([^,]+\\.[^,]+, context\\.Context, \\.\\.\\.[^,]+\\.RequestEditorFn\\) \\(.+, error\\)")
	// An HTTP call with arguments/body
	normalCall = regexp.MustCompile("func\\([^,]+\\.[^,]+, context\\.Context, [^,]+\\.[^,]+, \\.\\.\\.[^,]+\\.RequestEditorFn\\) \\(.+, error\\)")
)

func (b *Builder[T]) simpleCall(method reflect.Method) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := log.With("method", method.Name, "type", "simpleCall")

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		doAction := func() (*mcp.CallToolResult, error) {
			defer func() {
				if reason := recover(); reason != nil {
					log.Error("Paniced whilst doing API call", "reason", reason)
				}
			}()

			resp := method.Func.Call([]reflect.Value{
				reflect.ValueOf(b.Client),
				reflect.ValueOf(ctx),
			})

			httpResp := resp[0].Interface().(*http.Response)
			err := resp[1].Interface().(error)

			if err != nil {
				err = errors.Join(errors.New("Cannot call API"), err)
				log.Warn("API call failed", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				err = errors.Join(errors.New("Cannot read response from API"), err)
				log.Warn("API response read failed", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		}

		resp, err := doAction()
		if resp == nil && err == nil {
			return nil, errors.New("The tool paniced, see logs")
		} else {
			return resp, err
		}
	}
}

func (b *Builder[T]) advancedCall(method reflect.Method) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := log.With("method", method.Name, "type", "advancedCall")

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		doAction := func() (*mcp.CallToolResult, error) {
			defer func() {
				if reason := recover(); reason != nil {
					log.Error("Paniced whilst doing API call", "reason", reason)
				}
			}()

			resp := method.Func.Call([]reflect.Value{
				reflect.ValueOf(b.Client),
				// TODO: argument 1
				reflect.ValueOf(ctx),
			})

			httpResp := resp[0].Interface().(*http.Response)
			err := resp[1].Interface().(error)

			if err != nil {
				err = errors.Join(errors.New("Cannot call API"), err)
				log.Warn("API call failed", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				err = errors.Join(errors.New("Cannot read response from API"), err)
				log.Warn("API response read failed", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		}

		resp, err := doAction()
		if resp == nil && err == nil {
			return nil, errors.New("The tool paniced, see logs")
		} else {
			return resp, err
		}
	}
}

func (b *Builder[T]) addTool(method reflect.Method, server *server.MCPServer) error {
	// Cannot call this damn method so get rid of it
	if !method.IsExported() {
		return nil
	}

	// These are method that let you override the body, they are probably useless
	if strings.HasSuffix(method.Name, "WithBody") {
		return nil
	}

	signature := method.Type.String()
	log := log.With("method", method.Name, "signature", signature)
	log.Info("Looking at method")

	// TODO: use server.ToolHandlerFunc (which is broken for some reason)
	var handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	tool := mcp.NewTool(method.Name, mcp.WithDescription("Wrapped REST API call"))

	if simpleCall.MatchString(signature) {
		log.Info("Found simple call")
		handler = b.simpleCall(method)
	} else if normalCall.MatchString(signature) {
		log.Info("Found normal call")
		handler = b.advancedCall(method)
	} else {
		return nil
	}

	server.AddTool(tool, handler)
	return nil
}

func (b *Builder[T]) addTools(server *server.MCPServer) error {
	r := reflect.TypeOf(b.Client)
	log.Infof("Scanning %d methods", r.NumMethod())

	for i := range r.NumMethod() {
		meth := r.Method(i)
		err := b.addTool(meth, server)
		if err != nil {
			return errors.Join(fmt.Errorf("Cannot process client method (%s)", meth.Name), err)
		}
	}
	return nil
}

func (b *Builder[T]) Build() (*server.MCPServer, error) {
	server := server.NewMCPServer(b.Name, b.Version)
	err := b.addTools(server)
	if err != nil {
		return nil, errors.Join(errors.New("Cannot add tools"), err)
	}
	return server, nil
}

package mcptypebuilder

import (
	"context"
	"encoding/json"
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

type ClientStructMapper interface {
	//  maps a struct name i.e: GetPets to a reflect.Type
	StructType(key string) (reflect.Type, error)
}

type client any

type RequestEditorFn func(ctx context.Context, req *http.Request) error

type Builder[T client] struct {
	client       T
	structMapper ClientStructMapper
	name         string
	version      string
	editorFuncs  []RequestEditorFn
}

func New[T client](Name, Version string, Client T, mapper ClientStructMapper) *Builder[T] {
	return &Builder[T]{
		client:       Client,
		structMapper: mapper,
		name:         Name,
		version:      Version,
		editorFuncs:  make([]RequestEditorFn, 0),
	}
}

func (b *Builder[T]) AddRequestEditorFn(fn RequestEditorFn) {
	b.editorFuncs = append(b.editorFuncs, fn)
}

var (
	// An HTTP call with no arguments/body
	simpleCall = regexp.MustCompile("func\\([^,]+\\.[^,]+, context\\.Context, \\.\\.\\.[^,]+\\.RequestEditorFn\\) \\(.+, error\\)")
	// An HTTP call with arguments/body
	normalCall = regexp.MustCompile("func\\([^,]+\\.[^,]+, context\\.Context, [^,]+\\.(?P<type>[^,]+), \\.\\.\\.[^,]+\\.RequestEditorFn\\) \\(.+, error\\)")
)

func (b *Builder[T]) typeOfNormalCall(signature string) (reflect.Type, error) {
	keys := normalCall.FindStringSubmatch(signature)
	if len(keys) != 2 {
		return nil, fmt.Errorf("Cannot find type name in normal call (%s)\n\tmatches (%s)", signature, keys)
	}

	key := keys[1]
	return b.structMapper.StructType(key)
}

// func asAiPrompt(t reflect.Type) string {
//
// }

func (b *Builder[T]) simpleCall(method reflect.Method) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := log.With("method", method.Name, "type", "simpleCall")

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		doAction := func() (*mcp.CallToolResult, error) {
			defer func() {
				if reason := recover(); reason != nil {
					log.Error("Paniced whilst doing API call", "reason", reason)
				}
			}()

			args := []reflect.Value{
				reflect.ValueOf(b.client),
				reflect.ValueOf(ctx),
			}

			for _, editorFn := range b.editorFuncs {
				args = append(args, reflect.ValueOf(editorFn))
			}
			resp := method.Func.Call(args)

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

func (b *Builder[T]) advancedCall(method reflect.Method) (func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), error) {
	log := log.With("method", method.Name, "type", "advancedCall")
	signature := method.Type.String()
	argType, err := b.typeOfNormalCall(signature)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		doAction := func() (*mcp.CallToolResult, error) {
			defer func() {
				if reason := recover(); reason != nil {
					log.Error("Paniced whilst doing API call", "reason", reason)
				}
			}()

			req := reflect.New(argType.Elem()).Interface()
			err := json.Unmarshal([]byte(request.GetString("data", "{}")), &req)
			if err != nil {
				err = errors.Join(errors.New("Invalid body for call"), err)
				log.Warn("API call has wrong body", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			args := []reflect.Value{
				reflect.ValueOf(b.client),
				reflect.ValueOf(req),
				reflect.ValueOf(ctx),
			}

			for _, editorFn := range b.editorFuncs {
				args = append(args, reflect.ValueOf(editorFn))
			}
			resp := method.Func.Call(args)

			httpResp := resp[0].Interface().(*http.Response)
			err = resp[1].Interface().(error)

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
	}, nil
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

		var err error
		handler, err = b.advancedCall(method)
		if err != nil {
			err = errors.Join(fmt.Errorf("Cannot create tool for method (%s)", method.Name), err)
			log.Error("Cannot generate request type", "error", err)
			return errors.Join(err)
		}
	} else {
		return nil
	}

	server.AddTool(tool, handler)
	return nil
}

func (b *Builder[T]) addTools(server *server.MCPServer) error {
	r := reflect.TypeOf(b.client)
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
	server := server.NewMCPServer(b.name, b.version)
	err := b.addTools(server)
	if err != nil {
		return nil, errors.Join(errors.New("Cannot add tools"), err)
	}
	return server, nil
}

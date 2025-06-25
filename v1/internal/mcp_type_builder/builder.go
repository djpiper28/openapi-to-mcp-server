package mcptypebuilder

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
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
	log.Info("Looking at method", "signature", signature, "method", method.Name)

	if simpleCall.MatchString(signature) {
		log.Info("Found simple call", "method", method.Name)
	} else if normalCall.MatchString(signature) {
		log.Info("Found normal call", "method", method.Name)
	}

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

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	structtypemapgenerator "github.com/djpiper28/openapi-to-mcp-server/v1/cmd/struct_type_map_generator"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	InputFile       string `short:"i" long:"input" description:"Input file, i.e: specification.json" required:"true"`
	OutputDirectory string `short:"o" long:"output" description:"Output directory" required:"true"`
	PackageName     string `short:"p" long:"package" description:"Root package for the MCP server, i.e: " required:"true"`
}

func main() {
	log.Info("Generating MCP server...")

	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	rest, err := parser.Parse()
	if err != nil {
		log.Fatal("Cannot parse command line arguments", "error", err)
	}

	if len(rest) > 0 {
		log.Fatal("There are unused command line arguments", "unused", rest)
	}

	err = touchFolder(opts.OutputDirectory)
	if err != nil {
		log.Fatal("Cannot create output directory", "error", err)
	}

	err = opts.GenerateApiClient()
	if err != nil {
		log.Fatal("Cannot create API client", "error", err)
	}

	err = opts.GenerateTypeMapper()
	if err != nil {
		log.Fatal("Cannot create type mapper", "error", err)
	}

	err = opts.GenerateServer()
	if err != nil {
		log.Fatal("Cannot create server", "error", err)
	}
}

func touchFolder(name string) error {
	f, err := os.Stat(name)
	if err != nil {
		log.Debug("Cannot stat folder", "error", err)

		err = os.MkdirAll(name, 0777)
		if err != nil {
			return errors.Join(errors.New("Cannot create folder"), err)
		}
		return nil
	}

	if !f.IsDir() {
		return errors.New("A file with same name already exists")
	}

	return nil
}

const apiClientPackage = "apiclient"

func (o *Options) GenerateApiClient() error {
	clientPath := filepath.Join(o.OutputDirectory, apiClientPackage)

	log.Info("Creating API client...", "directory", clientPath)

	err := touchFolder(clientPath)
	if err != nil {
		return errors.Join(fmt.Errorf("Cannot create output folder (%s) for API client", clientPath), err)
	}

	const configFile = "config.yaml"
	configData := fmt.Sprintf(`package: %s
output: client.gen.go
generate:
  models: true
  client: true`, apiClientPackage)

	err = os.WriteFile(filepath.Join(clientPath, configFile), []byte(configData), 0666)
	if err != nil {
		return errors.Join(errors.New("Cannot create config for oapi-codegen"), err)
	}

	goPath, err := exec.LookPath("go")
	if err != nil {
		return errors.Join(errors.New("Cannot find go in your path"), err)
	}

	cmd := exec.Cmd{
		Path: goPath,
		Args: []string{
			goPath,
			"tool",
			"oapi-codegen",
			"-config",
			configFile,
			o.InputFile,
		},
		Dir:    clientPath,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	log.Debug("Executing", "command", cmd)
	err = cmd.Run()
	if err != nil {
		return errors.Join(errors.New("Cannot start oapi-codegen"), err)
	}

	log.Info("Completed API client generation")
	return nil
}

func (o *Options) GenerateTypeMapper() error {
	log.Info("Creating type map for API client...")
	inputPackage := o.PackageName + "/" + apiClientPackage
	clientPath := filepath.Join(o.OutputDirectory, apiClientPackage, "client.gen.go")

	stats, err := structtypemapgenerator.Generate(structtypemapgenerator.Args{
		InputPackageName:  inputPackage,
		InputFile:         clientPath,
		OutputPackageName: o.PackageName,
		OutputFile:        filepath.Join(o.OutputDirectory, "mapper.go"),
	})

	if err != nil {
		return errors.Join(errors.New("Cannot create struct type map"), err)
	}

	log.Infof("Created %d types in type mapper", len(stats.TypesFound))
	log.Info("Completed type mapper generation")
	return nil
}

func (o *Options) GenerateServer() error {
	log.Info("Creating server.go...")

	parts := strings.Split(o.PackageName, "/")
	packageNameShort := parts[len(parts)-1]

	buffer := fmt.Sprintf(`package %s

import (
	"errors"

  "%s/%s"
	builder "github.com/djpiper28/openapi-to-mcp-server/v1/lib/mcp_type_builder"
	"github.com/mark3labs/mcp-go/server"
)
`, packageNameShort, o.PackageName, apiClientPackage)

	buffer += fmt.Sprintf(`
func New(baseUrl, name, version string, requestEditorFns ...builder.RequestEditorFn) (*server.MCPServer, error) {
	client, err := %s.NewClient(baseUrl)
	if err != nil {
		return nil, errors.Join(errors.New("Cannot create client"), err)
	}

	b := builder.New(name, version, client, newMapper())

	for _, fn := range requestEditorFns {
		b.AddRequestEditorFn(fn)
	}

	server, err := b.Build()
	if err != nil {
		return nil, errors.Join(errors.New("Cannot create server"), err)
	}

	return server, nil
}`, apiClientPackage)

	err := os.WriteFile(filepath.Join(o.OutputDirectory, "server.go"), []byte(buffer), 0666)
	if err != nil {
		return errors.Join(errors.New("Cannot save server.go"), err)
	}

	log.Info("Created server.go...")
	return nil
}

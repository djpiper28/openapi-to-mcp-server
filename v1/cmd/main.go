package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
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

func (o *Options) GenerateApiClient() error {
	const clientPackage = "apiclient"
	clientPath := filepath.Join(o.OutputDirectory, clientPackage)

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
  client: true`, clientPackage)

	err = os.WriteFile(filepath.Join(clientPath, configFile), []byte(configData), 0777)
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

	return nil
}

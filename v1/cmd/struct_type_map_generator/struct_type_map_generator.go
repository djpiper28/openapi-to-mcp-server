package structtypemapgenerator

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Contains exposed stats about the output, and the internal state (hidden)
type StructTypeMapGenerator struct {
	TypesFound []string
	buffer     string
}

// All are required to be non-empty
type Args struct {
	InputPackageName  string // Complete package name of source i.e: github.com/djpiper28/openapi-to-mcp-server/test-output/test/apiclient
	InputFile         string // i.e: ../../../test-output/test/apiclient/client.gen.go
	OutputFile        string // i.e: ../../../test-output/test/mapper.go
	OutputPackageName string // Complete package name of the output i.e: github.com/djpiper28/openapi-to-mcp-server/test-output/test
}

func Generate(args Args) (*StructTypeMapGenerator, error) {
	data, err := os.ReadFile(args.InputFile)
	if err != nil {
		err = errors.Join(errors.New("Cannot read input file"), err)
		return nil, err
	}

	state := &StructTypeMapGenerator{
		TypesFound: make([]string, 0),
		buffer: `// AUTO GENERATED DO NOT MODIFY BY HAND
// generated by github.com/djpiper28/openapi-to-mcp-server
`,
	}

	parts := strings.Split(args.OutputPackageName, "/")
	outputPackageName := parts[len(parts)-1]
	state.buffer += fmt.Sprintf("package %s\n\n", outputPackageName)
	state.buffer += fmt.Sprintf(`import (
  "fmt"
  "reflect"

	openapi_types "github.com/oapi-codegen/runtime/types"
  target "%s"
)

`, args.InputPackageName)
	state.buffer += `// Maps struct names to structs
type structMapper struct {}

func newMapper() *structMapper {
  return &structMapper{}
}

func (s *structMapper) StructType(key string) (reflect.Type, error) {
  switch key {
`

	state.generateMapEntries(string(data))
	state.buffer += `
  case "UUID":
    return reflect.TypeOf(openapi_types.UUID{}), nil
  // case "Date":
  //   return reflect.TypeOf(openapi_types.Date{}), nil
  // case "Email":
  //   return reflect.TypeOf(openapi_types.Email{}), nil
  default:
    return nil, fmt.Errorf("Cannot find type %s", key)
  }
}`

	err = os.WriteFile(args.OutputFile, []byte(state.buffer), 0666)
	if err != nil {
		err = errors.Join(errors.New("Cannot create output file"), err)
		return nil, err
	}

	return state, nil
}

var (
	typeAliasRegex = regexp.MustCompile("^type\\s*(?P<alias>[^\\s]+)\\s*[^\\s]+\\s*")
	structRegex    = regexp.MustCompile("^type\\s*(?P<alias>[^\\s]+)\\s*struct\\s*\\{\\s*")
)

// Scans for:
//
//	`type xAlias x`
//
// and
//
//	`type x struct`
func (s *StructTypeMapGenerator) generateMapEntries(data string) {
	lines := strings.SplitSeq(data, "\n")
	for line := range lines {
		var key string
		if typeAliasRegex.MatchString(line) {
			key = typeAliasRegex.FindStringSubmatch(line)[1]
		} else if structRegex.MatchString(line) {
			key = structRegex.FindStringSubmatch(line)[1]
		}

		if key == "" {
			continue
		}

		s.buffer += fmt.Sprintf(`  case "%s":
    return reflect.TypeOf(*new(target.%s)), nil
`, key, key)

		s.TypesFound = append(s.TypesFound, key)
	}
}

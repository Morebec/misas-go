package builtin

import (
	"fmt"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
	"strings"
)

const HTTPEndpointType spec.Type = "http_endpoint"

type HTTPEndpointFailureResponse struct {
	StatusCode  int    `yaml:"statusCode"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
	ErrorType   string `yaml:"errorType"`
}

type HTTPEndpointSuccessResponse struct {
	StatusCode  int    `yaml:"statusCode"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
	// The DataType returned by this response
	Type typesystem.DataType `yaml:"type"`
}

type HTTPEndpointResponses struct {
	Success  HTTPEndpointSuccessResponse   `yaml:"success"`
	Failures []HTTPEndpointFailureResponse `yaml:"failures"`
}

type HTTPEndpointSpecProperties struct {
	Path      string                `yaml:"path"`
	Method    string                `yaml:"method"`
	Request   typesystem.DataType   `yaml:"request"`
	Responses HTTPEndpointResponses `yaml:"responses"`
}

func (c HTTPEndpointSpecProperties) IsSpecProperties() {}

func HTTPEndpointDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[HTTPEndpointSpecProperties](HTTPEndpointType)
	return processing.NewTypeBasedDeserializer(HTTPEndpointType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		properties := s.Properties.(HTTPEndpointSpecProperties)
		if properties.Request == "" {
			properties.Request = typesystem.Null
		}

		s.Properties = properties

		return s, nil
	})
}

func HTTPEndpointDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		endpoints := specs.SelectType(HTTPEndpointType)

		var nodes []processing.DependencyNode

		for _, ep := range endpoints {
			props := ep.Properties.(HTTPEndpointSpecProperties)
			var deps []spec.TypeName

			requestType := props.Request.ExtractUserDefined()
			if requestType != "" {
				deps = append(deps, spec.TypeName(requestType))
			}

			// success response
			if props.Responses.Success.Type.IsUserDefined() {
				deps = append(deps, spec.TypeName(props.Responses.Success.Type.ExtractUserDefined()))
			}
			nodes = append(nodes, processing.NewDependencyNode(ep, deps...))
		}

		return nodes, nil
	}
}

func HTTPEndpointsShouldFollowNamingConvention() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		endpoints := specs.SelectType(HTTPEndpointType)
		var lintingWarnings processing.LintingWarnings
		for _, c := range endpoints {
			if len(c.TypeName.Parts()) < 3 {
				lintingWarnings = append(lintingWarnings, fmt.Sprintf("endpoint %s does not follow the module.aggregate.entity.action naming convention at %s", c.TypeName, c.Source.Location))
			}
		}

		return lintingWarnings, nil
	}
}

func HTTPEndpointPathsShouldStartWithForwardSlash() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)
		for _, c := range endpoints {
			if !strings.HasPrefix(c.Properties.(HTTPEndpointSpecProperties).Path, "/") {
				lintingErrors = append(lintingErrors, errors.Errorf("endpoint %s does not have a path that starts with \"/\" at %s", c.TypeName, c.Source))
			}
		}
		return nil, lintingErrors
	}
}

func HTTPEndpointPathsShouldNotEndWithForwardSlash() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)
		for _, c := range endpoints {
			if strings.HasSuffix(c.Properties.(HTTPEndpointSpecProperties).Path, "/") {
				lintingErrors = append(lintingErrors, errors.Errorf("endpoint %s should not have a path that ends with \"/\" at %s", c.TypeName, c.Source))
			}
		}
		return nil, lintingErrors
	}
}

func HTTPEndpointPathsShouldBeUnique() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)

		endpointPaths := map[string]spec.Group{}
		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			if _, found := endpointPaths[properties.Path]; found {
				endpointPaths[properties.Path] = spec.Group{}
			}
			endpointPaths[properties.Path] = append(endpointPaths[properties.Path], ep)
		}

		for p, v := range endpointPaths {
			if len(v) > 1 {
				specsWithPath := spec.MapSpecGroup[string](v, func(s spec.Spec) string { return fmt.Sprintf("%s at %s", string(s.TypeName), s.Source.Location) })
				lintingErrors = append(lintingErrors, errors.Errorf("duplicate endpoint paths found for %s in %s", p, strings.Join(specsWithPath, ", ")))
			}
		}

		return nil, lintingErrors
	}
}

func HTTPEndpointPathShouldBeLowercase() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)

		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			if properties.Path != strings.ToLower(properties.Path) {
				lintingErrors = append(lintingErrors, errors.Errorf("endpoint %s does not have a lowercase path at %s", ep.TypeName, ep.Source.Location))
			}
		}
		return nil, lintingErrors
	}
}

func HTTPEndpointsShouldHaveEitherGETorPOSTMethod() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var LintingWarnings processing.LintingWarnings
		endpoints := specs.SelectType(HTTPEndpointType)

		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			if properties.Method != "GET" && properties.Method != "POST" {
				LintingWarnings = append(LintingWarnings, fmt.Sprintf(
					"endpoint %s should have either GET or POST as method, got \"%s\" at %s",
					ep.TypeName,
					properties.Method,
					ep.Source.Location,
				))
			}
		}
		return LintingWarnings, nil
	}
}

func HTTPEndpointResponseShouldHaveValidStatusCode() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var LintingWarnings processing.LintingWarnings
		endpoints := specs.SelectType(HTTPEndpointType)

		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			if properties.Responses.Success.StatusCode < 200 || properties.Responses.Success.StatusCode >= 300 {
				LintingWarnings = append(LintingWarnings, fmt.Sprintf(
					"endpoint %s success response should have a valid status code, got \"%d\" at %s",
					ep.TypeName,
					properties.Responses.Success.StatusCode,
					ep.Source.Location,
				))
			}

			for _, res := range properties.Responses.Failures {
				if res.StatusCode < 200 || res.StatusCode >= 600 {
					LintingWarnings = append(LintingWarnings, fmt.Sprintf(
						"endpoint %s failure response should have a valid status code, got \"%d\" at %s",
						ep.TypeName,
						res.StatusCode,
						ep.Source.Location,
					))
				}
			}
		}
		return LintingWarnings, nil
	}
}

func HTTPEndpointsWithCommandRequestTypeMustHaveMethodPOST() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)

		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			requestTypeUserDef := properties.Request.ExtractUserDefined()
			if requestTypeUserDef == "" {
				continue
			}
			reqTypeName := spec.TypeName(requestTypeUserDef)
			req := specs.SelectTypeName(reqTypeName)

			if req.TypeName == reqTypeName {
				continue // Not found will be caught when resolving dependencies.
			}

			if req.Type != CommandType {
				continue
			}

			if properties.Method != "POST" {
				lintingErrors = append(lintingErrors, errors.Errorf(
					"endpoint %s uses a command as its request type, but does not have method POST at %s",
					ep.TypeName,
					ep.Source.Location,
				))
			}
		}
		return nil, lintingErrors
	}
}

func HTTPEndpointsWithQueryRequestTypeMustHaveMethodGET() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var lintingErrors processing.LintingErrors
		endpoints := specs.SelectType(HTTPEndpointType)

		for _, ep := range endpoints {
			properties := ep.Properties.(HTTPEndpointSpecProperties)
			requestTypeUserDef := properties.Request.ExtractUserDefined()
			if requestTypeUserDef == "" {
				continue
			}
			reqTypeName := spec.TypeName(requestTypeUserDef)
			req := specs.SelectTypeName(reqTypeName)

			if req.TypeName == reqTypeName {
				continue // Not found will be caught when resolving dependencies.
			}

			if req.Type != QueryType {
				continue
			}

			if properties.Method != "GET" {
				lintingErrors = append(lintingErrors, errors.Errorf(
					"endpoint %s uses a command as its request type, but does not have method POST at %s",
					ep.TypeName,
					ep.Source.Location,
				))
			}
		}
		return nil, lintingErrors
	}
}

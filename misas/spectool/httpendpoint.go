package spectool

import "github.com/morebec/specter"

type HTTPEndpointFailureResponse struct {
	StatusCode  int    `hcl:"statusCode"`
	Description string `hcl:"description"`
	Example     string `hcl:"example,optional"`
	ErrorType   string `hcl:"errorType"`
}

type HTTPEndpointSuccessResponse struct {
	StatusCode  int    `hcl:"statusCode"`
	Description string `hcl:"description"`
	Example     string `hcl:"example"`
	// The DataType returned by this response
	Type DataType `hcl:"type"`
}

type HTTPEndpointResponses struct {
	Success  HTTPEndpointSuccessResponse   `hcl:"success,optional"`
	Failures []HTTPEndpointFailureResponse `hcl:"failures,optional"`
}

type HTTPEndpoint struct {
	Nam    string `hcl:"name,label"`
	Method string `hcl:"method,label"`
	Path   string `hcl:"path,label"`
	Desc   string `hcl:"description"`

	Request   DataType              `hcl:"request,block"`
	Responses HTTPEndpointResponses `hcl:"responses,block"`

	Annots Annotations `hcl:"annotations,block,optional"`
	Src    specter.Source
}

func (he *HTTPEndpoint) Name() specter.SpecificationName {
	return specter.SpecificationName(he.Nam)
}

func (he *HTTPEndpoint) Type() specter.SpecificationType {
	return "http_endpoint"
}

func (he *HTTPEndpoint) Description() string {
	return he.Desc
}

func (he *HTTPEndpoint) Source() specter.Source {
	return he.Src
}

func (he *HTTPEndpoint) SetSource(s specter.Source) {
	he.Src = s
}

func (he *HTTPEndpoint) Annotations() Annotations {
	return he.Annots
}

func (he *HTTPEndpoint) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName

	requestType := he.Request.ExtractUserDefined()
	if requestType != "" {
		deps = append(deps, specter.SpecificationName(requestType))
	}

	// success response
	if he.Responses.Success.Type.IsUserDefined() {
		deps = append(deps, specter.SpecificationName(he.Responses.Success.Type.ExtractUserDefined()))
	}

	return deps
}

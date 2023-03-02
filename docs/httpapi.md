# HTTP API

## Creating an HTTP API Web Server
The library offers a default implementation of an HTTP API Web server that aids with implementing
web servers.

```go
httpapi.NewServer()
```

It supports passing it a list of ServerOption to configure it:

```go
crosOptions := cors.New(cors.Options{
    AllowedOrigins:     []string{"*"},
    AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:     []string{"Link"},
    AllowCredentials:   true,
    OptionsPassthrough: false,
    MaxAge:             3599, // Maximum value not ignored by any of major browsers
})

s := httpapi.NewServer(
    httpapi.WithPort(config.Port),

    // Middleware
    httpapi.WithMiddleware(crosOptions.Handler),

    // Public Endpoints
    httpapi.WithHeartbeat("/health"),
    httpapi.WithPostEndpoint("/login", LoginEndpoint(cb, tokenManager)),

    // Data Sources

    // Protected Endpoints
    httpapi.WithGroup(
        httpapi.UseGroupMiddleware(
            tokenManager.VerifierHandler(),
            tokenManager.AuthenticatorHandler(),
        ),

        // Auth
        httpapi.WithPostGroupEndpoint("/auth", AuthEndpoint()),
		
        // Label definitions
        httpapi.WithGetGroupEndpoint("/labels", GetLabelsEndpoint(qb)),
    ),
)
```

## Creating an Endpoint
```go
func LoginEndpoint(cb command.Bus, tm *JWTTokenAuthManager) httpapi.EndpointFunc {
	return func(r *httpapi.EndpointRequest) httpapi.EndpointResponse {
		cmd := users.LoginCommandPayload{}

		if err := r.Unmarshal(&cmd); err != nil {
			return httpapi.NewErrorResponse(err)
		}

		resp, err := cb.Send(r.Context(), command.Command{
			Payload:  cmd,
			Metadata: nil,
		})
		if err != nil {
			return httpapi.NewErrorResponse(err)
		}

		user := resp.(*users.User)
		encodedToken, err := tm.Encode(*user)
		if err != nil {
			return httpapi.NewErrorResponse(err)
		}

		return httpapi.NewSuccessResponse(encodedToken).
			WithCookie(&http.Cookie{Name: "jwt", Value: encodedToken, HttpOnly: true})
	}
}
```

> Note: The `httpapi.EndpointRequest` struct embeds a `http.Request` struct.

## Creating a Response

### Creating a Success Response
```go
httpapi.NewSuccessResponse(
	map[string]any{
		"hello": "world"
    }, 
)
```
The data parameter passed to this function, will be marshalled to JSON, it is therefore important
to ensure when passing a struct it contains the necessary elements for JSON marshalling.


### Creating a failure Response
Returns a response that indicates a failure when processing a request:
```go
httpapi.NewFailureResponse(
	"unauthorized_access", 
	"you are not authorized to access this resource.",
	nil, 
)
```


### Creating an Error Response
Returns a failure response based on a go error object. Internally, it makes a call to the
`httpapi.NewFailureResponse` method to generate a failure response. It tries to find
an error code wrapped in the error (if the error was created using `morebec/go-errors`) 
or defaults to`httpapi.NewInternalErrorResponse` if none could be found.

```go
httpapi.NewErrorResponse(err)
```


### Creating an Internal Server Error Response
Returns an internal error response that should correspond to a 500 HTTP response.
```go
httpapi.NewInternalErrorResponse(err)
```
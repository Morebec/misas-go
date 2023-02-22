# HTTP API

## Creating an HTTP API Web Server
The library offers a default implementation of an HTTP API Web server that aids with implementing
web servers.

```go
httpapi.NewServer()
```

## Creating a Response

### Creating a Success Response
```go
httpapi.NewSuccessResponse(
	"unauthorized_access", 
	"you are not authorized to access this resource.",
	nil, 
)
```


### Creating a failure Response
Returns a response that indicates a failure when processing a request, internally
it makes a call to the
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
an error code wrapped in the error or defaults to`httpapi.NewInternalErrorResponse` 
if none could be found.

```go
httpapi.NewErrorResponse(err)
```


### Creating an Internal Server Error Response
Returns an internal error response that should correspond to a 500 HTTP response.
```go
httpapi.NewInternalErrorResponse(err)
```
# Clients Testing Helpers Package Examples

The clients testing package provides helper functions that can be used for HTTP client requirements in testing scenarios.

## Sending a Request

To send a HTTP request, a usable client must first be created. Further instruction on setting up clients can be found in the [e2e clients package](https://github.ibm.com/automation-base-pak/iaf-shared-helpers/blob/main/e2e/clients/examples.md).

The `SendRequest` function takes the method type (GET, POST etc), the path to send the request to and an optional request body as parameters. For compatible clients it handles adding authentication if provided. 

Example usage with a simple HTTP client, no authentication:

```go
tlsHost := e2e.CreateRoute(k8sClient, tlsRoute, tlsRouteName)

// Create a tls no auth client request, to prove that auth is on
client = pkgclients.NewHTTPClientBuilder(fmt.Sprintf("https://%s", tlsHost)).WithTLS(nil, true).Build()
Eventually(func() (int, error) {
    resp, err := client.SendRequest("GET", "", nil)
    return resp.StatusCode, err
}, 10, 1).Should(Equal(401)) // No credentials provided
```

Example usage with an authenticated client and checking of response code:

```go
resp, err := elasticClient.SendRequest(http.MethodPost, "_security/users", bytes.NewBuffer(requestBody))
Expect(err).NotTo(HaveOccurred(), "failed to add a new security user")
Expect(resp.StatusCode).To(Equal(http.StatusCreated), "201 Created status code not received from request")
```

The `RequestParams` struct specified a method type (GET, POST etc), the path to send the request to, an optional request body and headers. This struct can be used to call `CustomSendRequest` which will use that configuration to perform an HTTP request.

```go
tlsHost := e2e.CreateRoute(k8sClient, tlsRoute, tlsRouteName)

// Create a tls no auth client request, to prove that auth is on
client = pkgclients.NewHTTPClientBuilder(fmt.Sprintf("https://%s", tlsHost)).WithTLS(nil, true).Build()
Eventually(func() (int, error) {
    resp, err := client.CustomSendRequest(pkgclient.RequestParams{method: "GET", path: ""})
    return resp.StatusCode, err
}, 10, 1).Should(Equal(401)) // No credentials provided
```



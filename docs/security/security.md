# Security

Flyte is designed to be used across different security zones. 
All that is required is that the packs are able to access the Flyte-API. For example, the packs can run in the most secure (e.g. production) environment, monitoring—or even being embedded within—production apps, 
while the Flyte-API runs in a less secure environment, or even the internal network.


### TLS Mode

TLS is supported by flyte

To run flyte with TLS mode enabled, filenames of a certificate and matching private key for the server must be
provided using the environment variables `FLYTE_TLS_CERT_PATH` and `FLYTE_TLS_KEY_PATH`. 

More information regarding the use of TLS in Golang can be found in the [Golang documentation](https://golang.org/pkg/net/http/#ListenAndServeTLS).

## Authentication and Authorization

Authentication and Authorization is handled using [JWT](https://jwt.io/).
Flyte implements OpenID Connect (OIDC), an authentication layer on top of OAuth 2.0, an authorization framework.

More information regarding how Flyte API and an OIDC provider interact with Users / Flyte Packs can be found [here](AuthenticationAuthorization.md).

To enable auth you must set the following env variables:

- `FLYTE_OIDC_ISSUER_URL` - this is the URL to the OpenID Connect issuer.
- `FLYTE_OIDC_ISSUER_CLIENT_ID` - this is the Client ID of the OpenID Connect issuer.
- `FLYTE_AUTH_POLICY_PATH` - this is a filepath to an auth policy yaml file (see below for details)

#### jwt token

Protected resources are accessed by including a valid JWT bearer token in HTTP requests to flyte.
This token must be issued by the same provider as the referenced OIDC issuer in flyte configuration (`FLYTE_OIDC_ISSUER_URL` & `FLYTE_OIDC_ISSUER_CLIENT_ID`)

#### auth policy yaml 

An example of an auth policy yaml file can be found in [auth/testdata/policy_config.yaml](../../auth/testdata/policy_config.yaml) 

The file routes incoming requests (based on the path and http method) to a claims based authorisation policy. 
This routing is completely independent of any routing specified in the app. 
The file consists of an array of path policies. Each path policy can contain the following elements:

- `path` specifies a valid [vestigo](https://github.com/husobee/vestigo) path. These can be:
  - static e.g. `/packs` - will only match this exact path 
  - wildcarded e.g. `/packs/*` - will match multiple levels below e.g. /packs/foo, /packs/foo/bar will all match. 
  The wildcard can appear mid url as well, so for example `/packs/*/id` would match /packs/foo/bar/id etc.
  - dynamic/templated e.g. `/packs/:pack` - will match a single level below e.g. /packs/foo-pack. 
  Additionally the template value can be used in the claims (either as a key or value) e.g. claims: { "pack" : ":pack" } 
  would be dynamically resolved to be { "pack" : "foo-pack" } for the above request 
- `method` a list of valid http methods that are used in conjunction with the path to match incoming events. 
If no methods are specified then matching is done purely on the path i.e. all valid http methods are accepted.
- `claims` a map[string][]string of claims that an incoming token must match at least one of to be successfully authenticated. 
If no claims are specified then any request that matches the routing will be allowed through (regardless of whether it contains a valid token or not)  

If the request does not match a path & method in the policy file then 401 unauthorized will be returned.

If the request matches a path & method but does not satisfy the claims then 401 unauthorized will be returned.

If the request matches a path & method and satisfies it's claims then the request will be allowed to continue to the app.

A path policy may be as follows:

```yaml
path: /packs/:pack
methods: [DELETE]
claims:
    groups:
    - admin
    - 1
    pack:
    - :pack
```

This means that a request to e.g. `DELETE /packs/foo-pack` must include a jwt token including claims that satisfy at least one of above claims.
The following types are supported for claims in the token:
 - string
 - bool
 - int
 - []string
 
For example a token containing any of the following claims would succeed (this is not an exhaustive list):

```
- "groups" : "admin"
- "groups" : [..., "admin", ...]
- "groups" : "1"
- "groups" : 1
- "groups" : [..., "1", ...]
- "pack" : "foo-pack"  # the is a dynamic claim where the value of ':pack' in the url ('foo-pack') is then used in the claim
```

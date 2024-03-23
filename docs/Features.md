# Token distribution
Auth service was designed for the [phantom token](https://curity.io/resources/learn/phantom-token-pattern/) approach. The service can generate opaque tokens for the clients and translate them to JWTs for the backend.

## Opaque tokens 
Opaque tokens on their own contain no information about the owner's identity, roles or any other information.

Tokens are generated from random 16 char long strings generated from this charset: `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`.
Each token's random string is used to lookup token's related JWT and is effectively its `jti` (JWT ID).

Each token contains a CRC32 checksum to reject poorly forged fakes without making additional DB lookups.
The checksum is appended to the source string and the result is encoded using unpadded Base64URL.

A prefix is added for readability. `df` stands for dev_forum.

#### Opaque token generation example:
A token ID of `aaaaaaaaaaaa` will result in:
```shell
dfa_YWFhYWFhYWFhYWFhXzlhNWVhMWZh # access token
# or
dfr_YWFhYWFhYWFhYWFhXzlhNWVhMWZh # refresh token
```

## JWTs
Each opaque token has to be translated to a JWT before it can be used by any of the backend services.

Data needed to recreate a JWT from a token ID is stored in MongoDB in an unencrypted form.

Every translated JWT is also given:
- a `kid` which points to the private key used to sign the token. It's used to retrieve a correct public JWK from the JWKS (JWK Set) to verify the token with,
- a `jti` which is set to the opaque token source string,
- a `type` claim which describes whether a JWT is an access token or a refresh token.

# Token Validator package written in Go
A `JWTValidator` compatible with the auth-service is available in package `github.com/krixlion/dev_forum-auth/pkg/tokens/validator`.

Example:
```Go
// refreshFunc is used to retrieve a fresh keyset for the
// validator to search for keys with given `kid`.
func refreshFunc(ctx context.Context) (validator.[]Key, error) {
    // You can implement your own or use `DefaultRefreshFunc` included in the package.
    panic("not implemented")
}

func example(encodedJWT string) {
    validator := validator.NewValidator("auth-service", refreshFunc)
    
    // Run starts up the validator to refresh the keySet automatically using its `refreshFunc`.
    go validator.Run()

    // JWTValidator implements `github.com/krixlion/dev_forum-auth/pkg/tokens.Validator`.
    if err := validator.VerifyToken(encodedJWT); err != nil {
        panic(fmt.Sprintf("JWT validation failed: %s", err))
    }
}
```

# Signing keys rotation
JWK Set used to sign and verify JWTs is regularly rotated to mitigate the risk of any of the keys being compromised and used to perform unauthorized operations. Once the keyset is rotated it needs to be fetched by each service in the backend again.\
If the `JWTValidator` is used then the keyset will be refetched automatically.

It's planned to eventually add option to configure the duration between rotation cycles. 
Currently it's set to 24 hours.

# Telemetry
Auth service is exposing a HTTP `/metrics` endpoint for Prometheus to scrap.

It's also sending traces to an [OpenTelemetry-Collector](https://opentelemetry.io/docs/collector/). OpenTelemetry-Collector URL is configurable through `OTEL_EXPORTER_OTLP_ENDPOINT` env variable.\
You can integrate OtelCollector with any backend supported by it, e.g [Jaeger](https://www.jaegertracing.io/docs/).

# Token distribution
Auth service was designed for the [phantom token](https://curity.io/resources/learn/phantom-token-pattern/) approach. The service can generate opaque tokens for the clients and translate them to JWTs for the backend.


## Opaque tokens 
Opaque tokens on their own contain no information about the owner's identity, roles or any other information.

Tokens are generated from random 16 char long strings generated from this charset: `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`.
Each token's random string is used to lookup token's related JWT and is effectively it's `jti` (JWT ID).

Each token contains a CRC32 checksum to reject poorly forged fakes without making additional DB lookups.

#### Opaque token generation example:
a token ID of `aaaaaaaaaaaa` will result in:
```shell
dfa_YWFhYWFhYWFhYWFhXzlhNWVhMWZh # access token
dfr_YWFhYWFhYWFhYWFhXzlhNWVhMWZh # refresh token
```

## JWTs
Each opaque token has to be translated to a JWT before it can be used by any of the backend services.

Data needed to recreate a JWT is stored in MongoDB in an unencrypted form.

Every translated JWT is also given:
- a `kid` which points to the private key used to sign the token. It's used to retrieve a correct public JWK from the JWKS (JWK Set) to verify the token with,
- a `jti` which is set to the opaque token source,
- a `type` claim which describes whether a JWT is an access token or a refresh token.
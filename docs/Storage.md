## Token storage

Auth service stores issued JWT tokens in a MongoDB database.

Collection: `tokens`

```jsonc
// Token schema
{
    "_id": "string",
    "user_id": "string",
    "type": "string",
    "expires_at": "Date",
    "issued_at": "Date"
}
```

## Private key storage

Auth service stores JWK Set in a HashiCorp Vault.

Keys can be found on paths equal to their `kid`.
Each key contains fields:

- `private` - PEM encoded private key,
- `algorithm` - e.g. RS256,
- `keyType` - e.g. RSA.

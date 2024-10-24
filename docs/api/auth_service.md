# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [auth_service.proto](#auth_service-proto)
    - [GetAccessTokenRequest](#auth-GetAccessTokenRequest)
    - [GetAccessTokenResponse](#auth-GetAccessTokenResponse)
    - [Jwk](#auth-Jwk)
    - [SignInRequest](#auth-SignInRequest)
    - [SignInResponse](#auth-SignInResponse)
    - [SignOutRequest](#auth-SignOutRequest)
    - [TranslateAccessTokenRequest](#auth-TranslateAccessTokenRequest)
    - [TranslateAccessTokenRequest.MetadataEntry](#auth-TranslateAccessTokenRequest-MetadataEntry)
    - [TranslateAccessTokenResponse](#auth-TranslateAccessTokenResponse)
    - [TranslateAccessTokenResponse.MetadataEntry](#auth-TranslateAccessTokenResponse-MetadataEntry)
  
    - [AuthService](#auth-AuthService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="auth_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## auth_service.proto



<a name="auth-GetAccessTokenRequest"></a>

### GetAccessTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  | Encoded JWT refresh token |






<a name="auth-GetAccessTokenResponse"></a>

### GetAccessTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  | Encoded JWT access token |






<a name="auth-Jwk"></a>

### Jwk



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kid | [string](#string) |  | Key ID |
| kty | [string](#string) |  | Key Type |
| alg | [string](#string) |  | Key Signature Algorithm |
| key | [google.protobuf.Any](#google-protobuf-Any) |  | Field for key-specific data. Eg. {n, e} for RSA or {crv, x, y} for EC. |






<a name="auth-SignInRequest"></a>

### SignInRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| password | [string](#string) |  |  |
| email | [string](#string) |  |  |






<a name="auth-SignInResponse"></a>

### SignInResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  | Encoded JWT refresh token |






<a name="auth-SignOutRequest"></a>

### SignOutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  | Encoded JWT refresh token |






<a name="auth-TranslateAccessTokenRequest"></a>

### TranslateAccessTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| opaque_access_token | [string](#string) |  |  |
| metadata | [TranslateAccessTokenRequest.MetadataEntry](#auth-TranslateAccessTokenRequest-MetadataEntry) | repeated |  |






<a name="auth-TranslateAccessTokenRequest-MetadataEntry"></a>

### TranslateAccessTokenRequest.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="auth-TranslateAccessTokenResponse"></a>

### TranslateAccessTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  | Encoded JWT access token |
| metadata | [TranslateAccessTokenResponse.MetadataEntry](#auth-TranslateAccessTokenResponse-MetadataEntry) | repeated | Trace ID and etc. |






<a name="auth-TranslateAccessTokenResponse-MetadataEntry"></a>

### TranslateAccessTokenResponse.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="auth-AuthService"></a>

### AuthService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SignIn | [SignInRequest](#auth-SignInRequest) | [SignInResponse](#auth-SignInResponse) | Upon successful login user receives a refresh_token. When it expires or is revoked user has to login again. |
| SignOut | [SignOutRequest](#auth-SignOutRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | SignOut revokes user&#39;s active refresh_token. |
| GetAccessToken | [GetAccessTokenRequest](#auth-GetAccessTokenRequest) | [GetAccessTokenResponse](#auth-GetAccessTokenResponse) | Creates a new access token from a given refresh token. |
| GetValidationKeySet | [.google.protobuf.Empty](#google-protobuf-Empty) | [Jwk](#auth-Jwk) stream | Returns a list of public JWKs to use to verify incoming JWTs. |
| TranslateAccessToken | [TranslateAccessTokenRequest](#auth-TranslateAccessTokenRequest) stream | [TranslateAccessTokenResponse](#auth-TranslateAccessTokenResponse) stream | Requires mTLS client cert to be provided. Responds with a JWT related to given opaque token. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |


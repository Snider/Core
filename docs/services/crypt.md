---
title: internal
---
# Service: `internal`






## Types

### `type HashType`
```go
type HashType      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }

```
HashType defines the supported hashing algorithms.





### `type Options`
```go
type Options      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: nil
     5  .  .  Closing: -
     6  .  }
     7  .  Incomplete: false
     8  }

```
Options holds configuration for the crypt service.





### `type Service`
```go
type Service      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 1) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: nil
     8  .  .  .  .  Type: *ast.StarExpr {
     9  .  .  .  .  .  Star: -
    10  .  .  .  .  .  X: *ast.IndexExpr {
    11  .  .  .  .  .  .  X: *ast.SelectorExpr {
    12  .  .  .  .  .  .  .  X: *ast.Ident {
    13  .  .  .  .  .  .  .  .  NamePos: -
    14  .  .  .  .  .  .  .  .  Name: "core"
    15  .  .  .  .  .  .  .  .  Obj: nil
    16  .  .  .  .  .  .  .  }
    17  .  .  .  .  .  .  .  Sel: *ast.Ident {
    18  .  .  .  .  .  .  .  .  NamePos: -
    19  .  .  .  .  .  .  .  .  Name: "Runtime"
    20  .  .  .  .  .  .  .  .  Obj: nil
    21  .  .  .  .  .  .  .  }
    22  .  .  .  .  .  .  }
    23  .  .  .  .  .  .  Lbrack: -
    24  .  .  .  .  .  .  Index: *ast.Ident {
    25  .  .  .  .  .  .  .  NamePos: -
    26  .  .  .  .  .  .  .  Name: "Options"
    27  .  .  .  .  .  .  .  Obj: *ast.Object {
    28  .  .  .  .  .  .  .  .  Kind: type
    29  .  .  .  .  .  .  .  .  Name: "Options"
    30  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    31  .  .  .  .  .  .  .  .  .  Doc: nil
    32  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    33  .  .  .  .  .  .  .  .  .  .  NamePos: -
    34  .  .  .  .  .  .  .  .  .  .  Name: "Options"
    35  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 27)
    36  .  .  .  .  .  .  .  .  .  }
    37  .  .  .  .  .  .  .  .  .  TypeParams: nil
    38  .  .  .  .  .  .  .  .  .  Assign: -
    39  .  .  .  .  .  .  .  .  .  Type: *ast.StructType {
    40  .  .  .  .  .  .  .  .  .  .  Struct: -
    41  .  .  .  .  .  .  .  .  .  .  Fields: *ast.FieldList {
    42  .  .  .  .  .  .  .  .  .  .  .  Opening: -
    43  .  .  .  .  .  .  .  .  .  .  .  List: nil
    44  .  .  .  .  .  .  .  .  .  .  .  Closing: -
    45  .  .  .  .  .  .  .  .  .  .  }
    46  .  .  .  .  .  .  .  .  .  .  Incomplete: false
    47  .  .  .  .  .  .  .  .  .  }
    48  .  .  .  .  .  .  .  .  .  Comment: nil
    49  .  .  .  .  .  .  .  .  }
    50  .  .  .  .  .  .  .  .  Data: nil
    51  .  .  .  .  .  .  .  .  Type: nil
    52  .  .  .  .  .  .  .  }
    53  .  .  .  .  .  .  }
    54  .  .  .  .  .  .  Rbrack: -
    55  .  .  .  .  .  }
    56  .  .  .  .  }
    57  .  .  .  .  Tag: nil
    58  .  .  .  .  Comment: nil
    59  .  .  .  }
    60  .  .  }
    61  .  .  Closing: -
    62  .  }
    63  .  Incomplete: false
    64  }

```
Service provides cryptographic functions to the application.



#### Methods

- `DecryptPGP(recipientPath, message, passphrase      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, signerPath      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "string"
     5  .  .  Obj: nil
     6  .  }
     7  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: DecryptPGP decrypts a PGP message, optionally verifying the signature.

- `EncryptPGP(writer      0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "io"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "Writer"
     9  .  .  Obj: nil
    10  .  }
    11  }
, recipientPath, data      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, signerPath, signerPassphrase      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "string"
     5  .  .  Obj: nil
     6  .  }
     7  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: EncryptPGP encrypts data for a recipient, optionally signing it.

- `Fletcher16(payload      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "uint16"
     3  .  Obj: nil
     4  }
`: Fletcher16 computes the Fletcher-16 checksum.

- `Fletcher32(payload      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "uint32"
     3  .  Obj: nil
     4  }
`: Fletcher32 computes the Fletcher-32 checksum.

- `Fletcher64(payload      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "uint64"
     3  .  Obj: nil
     4  }
`: Fletcher64 computes the Fletcher-64 checksum.

- `Hash(lib      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "HashType"
     3  .  Obj: *ast.Object {
     4  .  .  Kind: type
     5  .  .  Name: "HashType"
     6  .  .  Decl: *ast.TypeSpec {
     7  .  .  .  Doc: nil
     8  .  .  .  Name: *ast.Ident {
     9  .  .  .  .  NamePos: -
    10  .  .  .  .  Name: "HashType"
    11  .  .  .  .  Obj: *(obj @ 3)
    12  .  .  .  }
    13  .  .  .  TypeParams: nil
    14  .  .  .  Assign: -
    15  .  .  .  Type: *ast.Ident {
    16  .  .  .  .  NamePos: -
    17  .  .  .  .  Name: "string"
    18  .  .  .  .  Obj: nil
    19  .  .  .  }
    20  .  .  .  Comment: nil
    21  .  .  }
    22  .  .  Data: nil
    23  .  .  Type: nil
    24  .  }
    25  }
, payload      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
`: Hash computes a hash of the payload using the specified algorithm.

- `Luhn(payload      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "bool"
     3  .  Obj: nil
     4  }
`: Luhn validates a number using the Luhn algorithm.





## Functions

- `Register(c      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "core"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "Core"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "any"
     3  .  Obj: nil
     4  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Register is the constructor for dynamic dependency injection (used with core.WithService). It creates a Service instance and initializes its core.Runtime field.

- `TestHash(t      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "testing"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "T"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
) `:

- `TestLuhn(t      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "testing"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "T"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
) `:

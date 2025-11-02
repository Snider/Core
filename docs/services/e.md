---
title: e
---
# Service: `e`

Package e provides a standardized error handling mechanism for the Core library.
It allows for wrapping errors with contextual information, making it easier to
trace the origin of an error and provide meaningful feedback.

The design of this package is influenced by the need for a simple, yet powerful
way to handle errors that can occur in different layers of the application,
from low-level file operations to high-level service interactions.

The key features of this package are:
  - Error wrapping: The Op and an optional Msg field provide context about
    where and why an error occurred.
  - Stack traces: By wrapping errors, we can build a logical stack trace
    that is more informative than a raw stack trace.
  - Consistent error handling: Encourages a uniform approach to error
    handling across the entire codebase.





## Types

### `type Error`
```go
type Error      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 3) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: *ast.CommentGroup {
     7  .  .  .  .  .  List: []*ast.Comment (len = 1) {
     8  .  .  .  .  .  .  0: *ast.Comment {
     9  .  .  .  .  .  .  .  Slash: -
    10  .  .  .  .  .  .  .  Text: "// Op is the operation being performed, e.g., \"config.Load\"."
    11  .  .  .  .  .  .  }
    12  .  .  .  .  .  }
    13  .  .  .  .  }
    14  .  .  .  .  Names: []*ast.Ident (len = 1) {
    15  .  .  .  .  .  0: *ast.Ident {
    16  .  .  .  .  .  .  NamePos: -
    17  .  .  .  .  .  .  Name: "Op"
    18  .  .  .  .  .  .  Obj: *ast.Object {
    19  .  .  .  .  .  .  .  Kind: var
    20  .  .  .  .  .  .  .  Name: "Op"
    21  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    22  .  .  .  .  .  .  .  Data: nil
    23  .  .  .  .  .  .  .  Type: nil
    24  .  .  .  .  .  .  }
    25  .  .  .  .  .  }
    26  .  .  .  .  }
    27  .  .  .  .  Type: *ast.Ident {
    28  .  .  .  .  .  NamePos: -
    29  .  .  .  .  .  Name: "string"
    30  .  .  .  .  .  Obj: nil
    31  .  .  .  .  }
    32  .  .  .  .  Tag: nil
    33  .  .  .  .  Comment: nil
    34  .  .  .  }
    35  .  .  .  1: *ast.Field {
    36  .  .  .  .  Doc: *ast.CommentGroup {
    37  .  .  .  .  .  List: []*ast.Comment (len = 1) {
    38  .  .  .  .  .  .  0: *ast.Comment {
    39  .  .  .  .  .  .  .  Slash: -
    40  .  .  .  .  .  .  .  Text: "// Msg is a human-readable message explaining the error."
    41  .  .  .  .  .  .  }
    42  .  .  .  .  .  }
    43  .  .  .  .  }
    44  .  .  .  .  Names: []*ast.Ident (len = 1) {
    45  .  .  .  .  .  0: *ast.Ident {
    46  .  .  .  .  .  .  NamePos: -
    47  .  .  .  .  .  .  Name: "Msg"
    48  .  .  .  .  .  .  Obj: *ast.Object {
    49  .  .  .  .  .  .  .  Kind: var
    50  .  .  .  .  .  .  .  Name: "Msg"
    51  .  .  .  .  .  .  .  Decl: *(obj @ 35)
    52  .  .  .  .  .  .  .  Data: nil
    53  .  .  .  .  .  .  .  Type: nil
    54  .  .  .  .  .  .  }
    55  .  .  .  .  .  }
    56  .  .  .  .  }
    57  .  .  .  .  Type: *ast.Ident {
    58  .  .  .  .  .  NamePos: -
    59  .  .  .  .  .  Name: "string"
    60  .  .  .  .  .  Obj: nil
    61  .  .  .  .  }
    62  .  .  .  .  Tag: nil
    63  .  .  .  .  Comment: nil
    64  .  .  .  }
    65  .  .  .  2: *ast.Field {
    66  .  .  .  .  Doc: *ast.CommentGroup {
    67  .  .  .  .  .  List: []*ast.Comment (len = 1) {
    68  .  .  .  .  .  .  0: *ast.Comment {
    69  .  .  .  .  .  .  .  Slash: -
    70  .  .  .  .  .  .  .  Text: "// Err is the underlying error that was wrapped."
    71  .  .  .  .  .  .  }
    72  .  .  .  .  .  }
    73  .  .  .  .  }
    74  .  .  .  .  Names: []*ast.Ident (len = 1) {
    75  .  .  .  .  .  0: *ast.Ident {
    76  .  .  .  .  .  .  NamePos: -
    77  .  .  .  .  .  .  Name: "Err"
    78  .  .  .  .  .  .  Obj: *ast.Object {
    79  .  .  .  .  .  .  .  Kind: var
    80  .  .  .  .  .  .  .  Name: "Err"
    81  .  .  .  .  .  .  .  Decl: *(obj @ 65)
    82  .  .  .  .  .  .  .  Data: nil
    83  .  .  .  .  .  .  .  Type: nil
    84  .  .  .  .  .  .  }
    85  .  .  .  .  .  }
    86  .  .  .  .  }
    87  .  .  .  .  Type: *ast.Ident {
    88  .  .  .  .  .  NamePos: -
    89  .  .  .  .  .  Name: "error"
    90  .  .  .  .  .  Obj: nil
    91  .  .  .  .  }
    92  .  .  .  .  Tag: nil
    93  .  .  .  .  Comment: nil
    94  .  .  .  }
    95  .  .  }
    96  .  .  Closing: -
    97  .  }
    98  .  Incomplete: false
    99  }

```
Error represents a standardized error with operational context.



#### Methods

- `Error()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
`: Error returns the string representation of the error.

- `Unwrap()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Unwrap provides compatibility for Go's errors.Is and errors.As functions.





## Functions

- `E(op, msg      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, err      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: E is a helper function to create a new Error. This is the primary way to create errors that will be consumed by the system. For example:  	return e.E("config.Load", "failed to load config file", err)  The 'op' parameter should be in the format of 'package.function' or 'service.method'. The 'msg' parameter should be a human-readable message that can be displayed to the user. The 'err' parameter is the underlying error that is being wrapped.

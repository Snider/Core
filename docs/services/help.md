---
title: help
---
# Service: `help`






## Types

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
Options holds configuration for the help service.





### `type Service`
```go
type Service      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 4) {
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
    60  .  .  .  1: *ast.Field {
    61  .  .  .  .  Doc: nil
    62  .  .  .  .  Names: []*ast.Ident (len = 1) {
    63  .  .  .  .  .  0: *ast.Ident {
    64  .  .  .  .  .  .  NamePos: -
    65  .  .  .  .  .  .  Name: "config"
    66  .  .  .  .  .  .  Obj: *ast.Object {
    67  .  .  .  .  .  .  .  Kind: var
    68  .  .  .  .  .  .  .  Name: "config"
    69  .  .  .  .  .  .  .  Decl: *(obj @ 60)
    70  .  .  .  .  .  .  .  Data: nil
    71  .  .  .  .  .  .  .  Type: nil
    72  .  .  .  .  .  .  }
    73  .  .  .  .  .  }
    74  .  .  .  .  }
    75  .  .  .  .  Type: *ast.SelectorExpr {
    76  .  .  .  .  .  X: *ast.Ident {
    77  .  .  .  .  .  .  NamePos: -
    78  .  .  .  .  .  .  Name: "core"
    79  .  .  .  .  .  .  Obj: nil
    80  .  .  .  .  .  }
    81  .  .  .  .  .  Sel: *ast.Ident {
    82  .  .  .  .  .  .  NamePos: -
    83  .  .  .  .  .  .  Name: "Config"
    84  .  .  .  .  .  .  Obj: nil
    85  .  .  .  .  .  }
    86  .  .  .  .  }
    87  .  .  .  .  Tag: nil
    88  .  .  .  .  Comment: nil
    89  .  .  .  }
    90  .  .  .  2: *ast.Field {
    91  .  .  .  .  Doc: nil
    92  .  .  .  .  Names: []*ast.Ident (len = 1) {
    93  .  .  .  .  .  0: *ast.Ident {
    94  .  .  .  .  .  .  NamePos: -
    95  .  .  .  .  .  .  Name: "display"
    96  .  .  .  .  .  .  Obj: *ast.Object {
    97  .  .  .  .  .  .  .  Kind: var
    98  .  .  .  .  .  .  .  Name: "display"
    99  .  .  .  .  .  .  .  Decl: *(obj @ 90)
   100  .  .  .  .  .  .  .  Data: nil
   101  .  .  .  .  .  .  .  Type: nil
   102  .  .  .  .  .  .  }
   103  .  .  .  .  .  }
   104  .  .  .  .  }
   105  .  .  .  .  Type: *ast.SelectorExpr {
   106  .  .  .  .  .  X: *ast.Ident {
   107  .  .  .  .  .  .  NamePos: -
   108  .  .  .  .  .  .  Name: "core"
   109  .  .  .  .  .  .  Obj: nil
   110  .  .  .  .  .  }
   111  .  .  .  .  .  Sel: *ast.Ident {
   112  .  .  .  .  .  .  NamePos: -
   113  .  .  .  .  .  .  Name: "Display"
   114  .  .  .  .  .  .  Obj: nil
   115  .  .  .  .  .  }
   116  .  .  .  .  }
   117  .  .  .  .  Tag: nil
   118  .  .  .  .  Comment: nil
   119  .  .  .  }
   120  .  .  .  3: *ast.Field {
   121  .  .  .  .  Doc: nil
   122  .  .  .  .  Names: []*ast.Ident (len = 1) {
   123  .  .  .  .  .  0: *ast.Ident {
   124  .  .  .  .  .  .  NamePos: -
   125  .  .  .  .  .  .  Name: "assets"
   126  .  .  .  .  .  .  Obj: *ast.Object {
   127  .  .  .  .  .  .  .  Kind: var
   128  .  .  .  .  .  .  .  Name: "assets"
   129  .  .  .  .  .  .  .  Decl: *(obj @ 120)
   130  .  .  .  .  .  .  .  Data: nil
   131  .  .  .  .  .  .  .  Type: nil
   132  .  .  .  .  .  .  }
   133  .  .  .  .  .  }
   134  .  .  .  .  }
   135  .  .  .  .  Type: *ast.SelectorExpr {
   136  .  .  .  .  .  X: *ast.Ident {
   137  .  .  .  .  .  .  NamePos: -
   138  .  .  .  .  .  .  Name: "embed"
   139  .  .  .  .  .  .  Obj: nil
   140  .  .  .  .  .  }
   141  .  .  .  .  .  Sel: *ast.Ident {
   142  .  .  .  .  .  .  NamePos: -
   143  .  .  .  .  .  .  Name: "FS"
   144  .  .  .  .  .  .  Obj: nil
   145  .  .  .  .  .  }
   146  .  .  .  .  }
   147  .  .  .  .  Tag: nil
   148  .  .  .  .  Comment: nil
   149  .  .  .  }
   150  .  .  }
   151  .  .  Closing: -
   152  .  }
   153  .  Incomplete: false
   154  }

```
Service manages the in-app help system.



#### Methods

- `HandleIPCEvents(c      0  *ast.StarExpr {
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
, msg      0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "core"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "Message"
     9  .  .  Obj: nil
    10  .  }
    11  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: HandleIPCEvents processes IPC messages, including injecting dependencies on startup.

- `ServiceStartup(     0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "context"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "Context"
     9  .  .  Obj: nil
    10  .  }
    11  }
,      0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "application"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "ServiceOptions"
     9  .  .  Obj: nil
    10  .  }
    11  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: ServiceStartup is called when the app starts, after dependencies are injected.

- `Show()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Show displays the help window.

- `ShowAt(anchor      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: ShowAt displays a specific section of the help documentation.





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
`: Register is the constructor for dynamic dependency injection (used with core.WithService). It creates a Service instance and initialises its core.Runtime field. Dependencies are injected during ServiceStartup.

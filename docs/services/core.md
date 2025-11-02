---
title: core
---
# Service: `core`






## Types

### `type ActionDisplayOpenWindow`
```go
type ActionDisplayOpenWindow      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Name"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "Name"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.Ident {
    21  .  .  .  .  .  NamePos: -
    22  .  .  .  .  .  Name: "string"
    23  .  .  .  .  .  Obj: nil
    24  .  .  .  .  }
    25  .  .  .  .  Tag: nil
    26  .  .  .  .  Comment: nil
    27  .  .  .  }
    28  .  .  .  1: *ast.Field {
    29  .  .  .  .  Doc: nil
    30  .  .  .  .  Names: []*ast.Ident (len = 1) {
    31  .  .  .  .  .  0: *ast.Ident {
    32  .  .  .  .  .  .  NamePos: -
    33  .  .  .  .  .  .  Name: "Options"
    34  .  .  .  .  .  .  Obj: *ast.Object {
    35  .  .  .  .  .  .  .  Kind: var
    36  .  .  .  .  .  .  .  Name: "Options"
    37  .  .  .  .  .  .  .  Decl: *(obj @ 28)
    38  .  .  .  .  .  .  .  Data: nil
    39  .  .  .  .  .  .  .  Type: nil
    40  .  .  .  .  .  .  }
    41  .  .  .  .  .  }
    42  .  .  .  .  }
    43  .  .  .  .  Type: *ast.SelectorExpr {
    44  .  .  .  .  .  X: *ast.Ident {
    45  .  .  .  .  .  .  NamePos: -
    46  .  .  .  .  .  .  Name: "application"
    47  .  .  .  .  .  .  Obj: nil
    48  .  .  .  .  .  }
    49  .  .  .  .  .  Sel: *ast.Ident {
    50  .  .  .  .  .  .  NamePos: -
    51  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    52  .  .  .  .  .  .  Obj: nil
    53  .  .  .  .  .  }
    54  .  .  .  .  }
    55  .  .  .  .  Tag: nil
    56  .  .  .  .  Comment: nil
    57  .  .  .  }
    58  .  .  }
    59  .  .  Closing: -
    60  .  }
    61  .  Incomplete: false
    62  }

```
ActionDisplayOpenWindow is a structured message for requesting a new window.





### `type ActionServiceShutdown`
```go
type ActionServiceShutdown      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: nil
     5  .  .  Closing: -
     6  .  }
     7  .  Incomplete: false
     8  }

```





### `type ActionServiceStartup`
```go
type ActionServiceStartup      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: nil
     5  .  .  Closing: -
     6  .  }
     7  .  Incomplete: false
     8  }

```





### `type Config`
```go
type Config      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Get"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "Get"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
    26  .  .  .  .  .  .  .  0: *ast.Field {
    27  .  .  .  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    29  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    30  .  .  .  .  .  .  .  .  .  .  NamePos: -
    31  .  .  .  .  .  .  .  .  .  .  Name: "key"
    32  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    33  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    34  .  .  .  .  .  .  .  .  .  .  .  Name: "key"
    35  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 26)
    36  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    37  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    38  .  .  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  .  }
    40  .  .  .  .  .  .  .  .  }
    41  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    42  .  .  .  .  .  .  .  .  .  NamePos: -
    43  .  .  .  .  .  .  .  .  .  Name: "string"
    44  .  .  .  .  .  .  .  .  .  Obj: nil
    45  .  .  .  .  .  .  .  .  }
    46  .  .  .  .  .  .  .  .  Tag: nil
    47  .  .  .  .  .  .  .  .  Comment: nil
    48  .  .  .  .  .  .  .  }
    49  .  .  .  .  .  .  .  1: *ast.Field {
    50  .  .  .  .  .  .  .  .  Doc: nil
    51  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    52  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    53  .  .  .  .  .  .  .  .  .  .  NamePos: -
    54  .  .  .  .  .  .  .  .  .  .  Name: "out"
    55  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    56  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    57  .  .  .  .  .  .  .  .  .  .  .  Name: "out"
    58  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 49)
    59  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    60  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    61  .  .  .  .  .  .  .  .  .  .  }
    62  .  .  .  .  .  .  .  .  .  }
    63  .  .  .  .  .  .  .  .  }
    64  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    65  .  .  .  .  .  .  .  .  .  NamePos: -
    66  .  .  .  .  .  .  .  .  .  Name: "any"
    67  .  .  .  .  .  .  .  .  .  Obj: nil
    68  .  .  .  .  .  .  .  .  }
    69  .  .  .  .  .  .  .  .  Tag: nil
    70  .  .  .  .  .  .  .  .  Comment: nil
    71  .  .  .  .  .  .  .  }
    72  .  .  .  .  .  .  }
    73  .  .  .  .  .  .  Closing: -
    74  .  .  .  .  .  }
    75  .  .  .  .  .  Results: *ast.FieldList {
    76  .  .  .  .  .  .  Opening: -
    77  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    78  .  .  .  .  .  .  .  0: *ast.Field {
    79  .  .  .  .  .  .  .  .  Doc: nil
    80  .  .  .  .  .  .  .  .  Names: nil
    81  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    82  .  .  .  .  .  .  .  .  .  NamePos: -
    83  .  .  .  .  .  .  .  .  .  Name: "error"
    84  .  .  .  .  .  .  .  .  .  Obj: nil
    85  .  .  .  .  .  .  .  .  }
    86  .  .  .  .  .  .  .  .  Tag: nil
    87  .  .  .  .  .  .  .  .  Comment: nil
    88  .  .  .  .  .  .  .  }
    89  .  .  .  .  .  .  }
    90  .  .  .  .  .  .  Closing: -
    91  .  .  .  .  .  }
    92  .  .  .  .  }
    93  .  .  .  .  Tag: nil
    94  .  .  .  .  Comment: nil
    95  .  .  .  }
    96  .  .  .  1: *ast.Field {
    97  .  .  .  .  Doc: nil
    98  .  .  .  .  Names: []*ast.Ident (len = 1) {
    99  .  .  .  .  .  0: *ast.Ident {
   100  .  .  .  .  .  .  NamePos: -
   101  .  .  .  .  .  .  Name: "Set"
   102  .  .  .  .  .  .  Obj: *ast.Object {
   103  .  .  .  .  .  .  .  Kind: func
   104  .  .  .  .  .  .  .  Name: "Set"
   105  .  .  .  .  .  .  .  Decl: *(obj @ 96)
   106  .  .  .  .  .  .  .  Data: nil
   107  .  .  .  .  .  .  .  Type: nil
   108  .  .  .  .  .  .  }
   109  .  .  .  .  .  }
   110  .  .  .  .  }
   111  .  .  .  .  Type: *ast.FuncType {
   112  .  .  .  .  .  Func: -
   113  .  .  .  .  .  TypeParams: nil
   114  .  .  .  .  .  Params: *ast.FieldList {
   115  .  .  .  .  .  .  Opening: -
   116  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   117  .  .  .  .  .  .  .  0: *ast.Field {
   118  .  .  .  .  .  .  .  .  Doc: nil
   119  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   120  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   121  .  .  .  .  .  .  .  .  .  .  NamePos: -
   122  .  .  .  .  .  .  .  .  .  .  Name: "key"
   123  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   124  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   125  .  .  .  .  .  .  .  .  .  .  .  Name: "key"
   126  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 117)
   127  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   128  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   129  .  .  .  .  .  .  .  .  .  .  }
   130  .  .  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   133  .  .  .  .  .  .  .  .  .  NamePos: -
   134  .  .  .  .  .  .  .  .  .  Name: "string"
   135  .  .  .  .  .  .  .  .  .  Obj: nil
   136  .  .  .  .  .  .  .  .  }
   137  .  .  .  .  .  .  .  .  Tag: nil
   138  .  .  .  .  .  .  .  .  Comment: nil
   139  .  .  .  .  .  .  .  }
   140  .  .  .  .  .  .  .  1: *ast.Field {
   141  .  .  .  .  .  .  .  .  Doc: nil
   142  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   143  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   144  .  .  .  .  .  .  .  .  .  .  NamePos: -
   145  .  .  .  .  .  .  .  .  .  .  Name: "v"
   146  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   147  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   148  .  .  .  .  .  .  .  .  .  .  .  Name: "v"
   149  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 140)
   150  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   151  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   152  .  .  .  .  .  .  .  .  .  .  }
   153  .  .  .  .  .  .  .  .  .  }
   154  .  .  .  .  .  .  .  .  }
   155  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   156  .  .  .  .  .  .  .  .  .  NamePos: -
   157  .  .  .  .  .  .  .  .  .  Name: "any"
   158  .  .  .  .  .  .  .  .  .  Obj: nil
   159  .  .  .  .  .  .  .  .  }
   160  .  .  .  .  .  .  .  .  Tag: nil
   161  .  .  .  .  .  .  .  .  Comment: nil
   162  .  .  .  .  .  .  .  }
   163  .  .  .  .  .  .  }
   164  .  .  .  .  .  .  Closing: -
   165  .  .  .  .  .  }
   166  .  .  .  .  .  Results: *ast.FieldList {
   167  .  .  .  .  .  .  Opening: -
   168  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   169  .  .  .  .  .  .  .  0: *ast.Field {
   170  .  .  .  .  .  .  .  .  Doc: nil
   171  .  .  .  .  .  .  .  .  Names: nil
   172  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   173  .  .  .  .  .  .  .  .  .  NamePos: -
   174  .  .  .  .  .  .  .  .  .  Name: "error"
   175  .  .  .  .  .  .  .  .  .  Obj: nil
   176  .  .  .  .  .  .  .  .  }
   177  .  .  .  .  .  .  .  .  Tag: nil
   178  .  .  .  .  .  .  .  .  Comment: nil
   179  .  .  .  .  .  .  .  }
   180  .  .  .  .  .  .  }
   181  .  .  .  .  .  .  Closing: -
   182  .  .  .  .  .  }
   183  .  .  .  .  }
   184  .  .  .  .  Tag: nil
   185  .  .  .  .  Comment: nil
   186  .  .  .  }
   187  .  .  }
   188  .  .  Closing: -
   189  .  }
   190  .  Incomplete: false
   191  }

```
Config provides access to application configuration.





### `type Contract`
```go
type Contract      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "DontPanic"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "DontPanic"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.Ident {
    21  .  .  .  .  .  NamePos: -
    22  .  .  .  .  .  Name: "bool"
    23  .  .  .  .  .  Obj: nil
    24  .  .  .  .  }
    25  .  .  .  .  Tag: nil
    26  .  .  .  .  Comment: nil
    27  .  .  .  }
    28  .  .  .  1: *ast.Field {
    29  .  .  .  .  Doc: nil
    30  .  .  .  .  Names: []*ast.Ident (len = 1) {
    31  .  .  .  .  .  0: *ast.Ident {
    32  .  .  .  .  .  .  NamePos: -
    33  .  .  .  .  .  .  Name: "DisableLogging"
    34  .  .  .  .  .  .  Obj: *ast.Object {
    35  .  .  .  .  .  .  .  Kind: var
    36  .  .  .  .  .  .  .  Name: "DisableLogging"
    37  .  .  .  .  .  .  .  Decl: *(obj @ 28)
    38  .  .  .  .  .  .  .  Data: nil
    39  .  .  .  .  .  .  .  Type: nil
    40  .  .  .  .  .  .  }
    41  .  .  .  .  .  }
    42  .  .  .  .  }
    43  .  .  .  .  Type: *ast.Ident {
    44  .  .  .  .  .  NamePos: -
    45  .  .  .  .  .  Name: "bool"
    46  .  .  .  .  .  Obj: nil
    47  .  .  .  .  }
    48  .  .  .  .  Tag: nil
    49  .  .  .  .  Comment: nil
    50  .  .  .  }
    51  .  .  }
    52  .  .  Closing: -
    53  .  }
    54  .  Incomplete: false
    55  }

```





### `type Core`
```go
type Core      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 10) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "once"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "once"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.SelectorExpr {
    21  .  .  .  .  .  X: *ast.Ident {
    22  .  .  .  .  .  .  NamePos: -
    23  .  .  .  .  .  .  Name: "sync"
    24  .  .  .  .  .  .  Obj: nil
    25  .  .  .  .  .  }
    26  .  .  .  .  .  Sel: *ast.Ident {
    27  .  .  .  .  .  .  NamePos: -
    28  .  .  .  .  .  .  Name: "Once"
    29  .  .  .  .  .  .  Obj: nil
    30  .  .  .  .  .  }
    31  .  .  .  .  }
    32  .  .  .  .  Tag: nil
    33  .  .  .  .  Comment: nil
    34  .  .  .  }
    35  .  .  .  1: *ast.Field {
    36  .  .  .  .  Doc: nil
    37  .  .  .  .  Names: []*ast.Ident (len = 1) {
    38  .  .  .  .  .  0: *ast.Ident {
    39  .  .  .  .  .  .  NamePos: -
    40  .  .  .  .  .  .  Name: "initErr"
    41  .  .  .  .  .  .  Obj: *ast.Object {
    42  .  .  .  .  .  .  .  Kind: var
    43  .  .  .  .  .  .  .  Name: "initErr"
    44  .  .  .  .  .  .  .  Decl: *(obj @ 35)
    45  .  .  .  .  .  .  .  Data: nil
    46  .  .  .  .  .  .  .  Type: nil
    47  .  .  .  .  .  .  }
    48  .  .  .  .  .  }
    49  .  .  .  .  }
    50  .  .  .  .  Type: *ast.Ident {
    51  .  .  .  .  .  NamePos: -
    52  .  .  .  .  .  Name: "error"
    53  .  .  .  .  .  Obj: nil
    54  .  .  .  .  }
    55  .  .  .  .  Tag: nil
    56  .  .  .  .  Comment: nil
    57  .  .  .  }
    58  .  .  .  2: *ast.Field {
    59  .  .  .  .  Doc: nil
    60  .  .  .  .  Names: []*ast.Ident (len = 1) {
    61  .  .  .  .  .  0: *ast.Ident {
    62  .  .  .  .  .  .  NamePos: -
    63  .  .  .  .  .  .  Name: "App"
    64  .  .  .  .  .  .  Obj: *ast.Object {
    65  .  .  .  .  .  .  .  Kind: var
    66  .  .  .  .  .  .  .  Name: "App"
    67  .  .  .  .  .  .  .  Decl: *(obj @ 58)
    68  .  .  .  .  .  .  .  Data: nil
    69  .  .  .  .  .  .  .  Type: nil
    70  .  .  .  .  .  .  }
    71  .  .  .  .  .  }
    72  .  .  .  .  }
    73  .  .  .  .  Type: *ast.StarExpr {
    74  .  .  .  .  .  Star: -
    75  .  .  .  .  .  X: *ast.SelectorExpr {
    76  .  .  .  .  .  .  X: *ast.Ident {
    77  .  .  .  .  .  .  .  NamePos: -
    78  .  .  .  .  .  .  .  Name: "application"
    79  .  .  .  .  .  .  .  Obj: nil
    80  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  Sel: *ast.Ident {
    82  .  .  .  .  .  .  .  NamePos: -
    83  .  .  .  .  .  .  .  Name: "App"
    84  .  .  .  .  .  .  .  Obj: nil
    85  .  .  .  .  .  .  }
    86  .  .  .  .  .  }
    87  .  .  .  .  }
    88  .  .  .  .  Tag: nil
    89  .  .  .  .  Comment: nil
    90  .  .  .  }
    91  .  .  .  3: *ast.Field {
    92  .  .  .  .  Doc: nil
    93  .  .  .  .  Names: []*ast.Ident (len = 1) {
    94  .  .  .  .  .  0: *ast.Ident {
    95  .  .  .  .  .  .  NamePos: -
    96  .  .  .  .  .  .  Name: "assets"
    97  .  .  .  .  .  .  Obj: *ast.Object {
    98  .  .  .  .  .  .  .  Kind: var
    99  .  .  .  .  .  .  .  Name: "assets"
   100  .  .  .  .  .  .  .  Decl: *(obj @ 91)
   101  .  .  .  .  .  .  .  Data: nil
   102  .  .  .  .  .  .  .  Type: nil
   103  .  .  .  .  .  .  }
   104  .  .  .  .  .  }
   105  .  .  .  .  }
   106  .  .  .  .  Type: *ast.SelectorExpr {
   107  .  .  .  .  .  X: *ast.Ident {
   108  .  .  .  .  .  .  NamePos: -
   109  .  .  .  .  .  .  Name: "embed"
   110  .  .  .  .  .  .  Obj: nil
   111  .  .  .  .  .  }
   112  .  .  .  .  .  Sel: *ast.Ident {
   113  .  .  .  .  .  .  NamePos: -
   114  .  .  .  .  .  .  Name: "FS"
   115  .  .  .  .  .  .  Obj: nil
   116  .  .  .  .  .  }
   117  .  .  .  .  }
   118  .  .  .  .  Tag: nil
   119  .  .  .  .  Comment: nil
   120  .  .  .  }
   121  .  .  .  4: *ast.Field {
   122  .  .  .  .  Doc: nil
   123  .  .  .  .  Names: []*ast.Ident (len = 1) {
   124  .  .  .  .  .  0: *ast.Ident {
   125  .  .  .  .  .  .  NamePos: -
   126  .  .  .  .  .  .  Name: "serviceLock"
   127  .  .  .  .  .  .  Obj: *ast.Object {
   128  .  .  .  .  .  .  .  Kind: var
   129  .  .  .  .  .  .  .  Name: "serviceLock"
   130  .  .  .  .  .  .  .  Decl: *(obj @ 121)
   131  .  .  .  .  .  .  .  Data: nil
   132  .  .  .  .  .  .  .  Type: nil
   133  .  .  .  .  .  .  }
   134  .  .  .  .  .  }
   135  .  .  .  .  }
   136  .  .  .  .  Type: *ast.Ident {
   137  .  .  .  .  .  NamePos: -
   138  .  .  .  .  .  Name: "bool"
   139  .  .  .  .  .  Obj: nil
   140  .  .  .  .  }
   141  .  .  .  .  Tag: nil
   142  .  .  .  .  Comment: nil
   143  .  .  .  }
   144  .  .  .  5: *ast.Field {
   145  .  .  .  .  Doc: nil
   146  .  .  .  .  Names: []*ast.Ident (len = 1) {
   147  .  .  .  .  .  0: *ast.Ident {
   148  .  .  .  .  .  .  NamePos: -
   149  .  .  .  .  .  .  Name: "ipcMu"
   150  .  .  .  .  .  .  Obj: *ast.Object {
   151  .  .  .  .  .  .  .  Kind: var
   152  .  .  .  .  .  .  .  Name: "ipcMu"
   153  .  .  .  .  .  .  .  Decl: *(obj @ 144)
   154  .  .  .  .  .  .  .  Data: nil
   155  .  .  .  .  .  .  .  Type: nil
   156  .  .  .  .  .  .  }
   157  .  .  .  .  .  }
   158  .  .  .  .  }
   159  .  .  .  .  Type: *ast.SelectorExpr {
   160  .  .  .  .  .  X: *ast.Ident {
   161  .  .  .  .  .  .  NamePos: -
   162  .  .  .  .  .  .  Name: "sync"
   163  .  .  .  .  .  .  Obj: nil
   164  .  .  .  .  .  }
   165  .  .  .  .  .  Sel: *ast.Ident {
   166  .  .  .  .  .  .  NamePos: -
   167  .  .  .  .  .  .  Name: "RWMutex"
   168  .  .  .  .  .  .  Obj: nil
   169  .  .  .  .  .  }
   170  .  .  .  .  }
   171  .  .  .  .  Tag: nil
   172  .  .  .  .  Comment: nil
   173  .  .  .  }
   174  .  .  .  6: *ast.Field {
   175  .  .  .  .  Doc: nil
   176  .  .  .  .  Names: []*ast.Ident (len = 1) {
   177  .  .  .  .  .  0: *ast.Ident {
   178  .  .  .  .  .  .  NamePos: -
   179  .  .  .  .  .  .  Name: "ipcHandlers"
   180  .  .  .  .  .  .  Obj: *ast.Object {
   181  .  .  .  .  .  .  .  Kind: var
   182  .  .  .  .  .  .  .  Name: "ipcHandlers"
   183  .  .  .  .  .  .  .  Decl: *(obj @ 174)
   184  .  .  .  .  .  .  .  Data: nil
   185  .  .  .  .  .  .  .  Type: nil
   186  .  .  .  .  .  .  }
   187  .  .  .  .  .  }
   188  .  .  .  .  }
   189  .  .  .  .  Type: *ast.ArrayType {
   190  .  .  .  .  .  Lbrack: -
   191  .  .  .  .  .  Len: nil
   192  .  .  .  .  .  Elt: *ast.FuncType {
   193  .  .  .  .  .  .  Func: -
   194  .  .  .  .  .  .  TypeParams: nil
   195  .  .  .  .  .  .  Params: *ast.FieldList {
   196  .  .  .  .  .  .  .  Opening: -
   197  .  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   198  .  .  .  .  .  .  .  .  0: *ast.Field {
   199  .  .  .  .  .  .  .  .  .  Doc: nil
   200  .  .  .  .  .  .  .  .  .  Names: nil
   201  .  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
   202  .  .  .  .  .  .  .  .  .  .  Star: -
   203  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   204  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   205  .  .  .  .  .  .  .  .  .  .  .  Name: "Core"
   206  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   207  .  .  .  .  .  .  .  .  .  .  .  .  Kind: type
   208  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Core"
   209  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
   210  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   211  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
   212  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   213  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Core"
   214  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 206)
   215  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   216  .  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
   217  .  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
   218  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *(obj @ 0)
   219  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   220  .  .  .  .  .  .  .  .  .  .  .  .  }
   221  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   222  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   223  .  .  .  .  .  .  .  .  .  .  .  }
   224  .  .  .  .  .  .  .  .  .  .  }
   225  .  .  .  .  .  .  .  .  .  }
   226  .  .  .  .  .  .  .  .  .  Tag: nil
   227  .  .  .  .  .  .  .  .  .  Comment: nil
   228  .  .  .  .  .  .  .  .  }
   229  .  .  .  .  .  .  .  .  1: *ast.Field {
   230  .  .  .  .  .  .  .  .  .  Doc: nil
   231  .  .  .  .  .  .  .  .  .  Names: nil
   232  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   233  .  .  .  .  .  .  .  .  .  .  NamePos: -
   234  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   235  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   236  .  .  .  .  .  .  .  .  .  .  .  Kind: type
   237  .  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   238  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
   239  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   240  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
   241  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   242  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   243  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 235)
   244  .  .  .  .  .  .  .  .  .  .  .  .  }
   245  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
   246  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
   247  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.InterfaceType {
   248  .  .  .  .  .  .  .  .  .  .  .  .  .  Interface: -
   249  .  .  .  .  .  .  .  .  .  .  .  .  .  Methods: *ast.FieldList {
   250  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
   251  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: nil
   252  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   253  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   254  .  .  .  .  .  .  .  .  .  .  .  .  .  Incomplete: false
   255  .  .  .  .  .  .  .  .  .  .  .  .  }
   256  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   257  .  .  .  .  .  .  .  .  .  .  .  }
   258  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   259  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   260  .  .  .  .  .  .  .  .  .  .  }
   261  .  .  .  .  .  .  .  .  .  }
   262  .  .  .  .  .  .  .  .  .  Tag: nil
   263  .  .  .  .  .  .  .  .  .  Comment: nil
   264  .  .  .  .  .  .  .  .  }
   265  .  .  .  .  .  .  .  }
   266  .  .  .  .  .  .  .  Closing: -
   267  .  .  .  .  .  .  }
   268  .  .  .  .  .  .  Results: *ast.FieldList {
   269  .  .  .  .  .  .  .  Opening: -
   270  .  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   271  .  .  .  .  .  .  .  .  0: *ast.Field {
   272  .  .  .  .  .  .  .  .  .  Doc: nil
   273  .  .  .  .  .  .  .  .  .  Names: nil
   274  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   275  .  .  .  .  .  .  .  .  .  .  NamePos: -
   276  .  .  .  .  .  .  .  .  .  .  Name: "error"
   277  .  .  .  .  .  .  .  .  .  .  Obj: nil
   278  .  .  .  .  .  .  .  .  .  }
   279  .  .  .  .  .  .  .  .  .  Tag: nil
   280  .  .  .  .  .  .  .  .  .  Comment: nil
   281  .  .  .  .  .  .  .  .  }
   282  .  .  .  .  .  .  .  }
   283  .  .  .  .  .  .  .  Closing: -
   284  .  .  .  .  .  .  }
   285  .  .  .  .  .  }
   286  .  .  .  .  }
   287  .  .  .  .  Tag: nil
   288  .  .  .  .  Comment: nil
   289  .  .  .  }
   290  .  .  .  7: *ast.Field {
   291  .  .  .  .  Doc: nil
   292  .  .  .  .  Names: []*ast.Ident (len = 1) {
   293  .  .  .  .  .  0: *ast.Ident {
   294  .  .  .  .  .  .  NamePos: -
   295  .  .  .  .  .  .  Name: "serviceMu"
   296  .  .  .  .  .  .  Obj: *ast.Object {
   297  .  .  .  .  .  .  .  Kind: var
   298  .  .  .  .  .  .  .  Name: "serviceMu"
   299  .  .  .  .  .  .  .  Decl: *(obj @ 290)
   300  .  .  .  .  .  .  .  Data: nil
   301  .  .  .  .  .  .  .  Type: nil
   302  .  .  .  .  .  .  }
   303  .  .  .  .  .  }
   304  .  .  .  .  }
   305  .  .  .  .  Type: *ast.SelectorExpr {
   306  .  .  .  .  .  X: *ast.Ident {
   307  .  .  .  .  .  .  NamePos: -
   308  .  .  .  .  .  .  Name: "sync"
   309  .  .  .  .  .  .  Obj: nil
   310  .  .  .  .  .  }
   311  .  .  .  .  .  Sel: *ast.Ident {
   312  .  .  .  .  .  .  NamePos: -
   313  .  .  .  .  .  .  Name: "RWMutex"
   314  .  .  .  .  .  .  Obj: nil
   315  .  .  .  .  .  }
   316  .  .  .  .  }
   317  .  .  .  .  Tag: nil
   318  .  .  .  .  Comment: nil
   319  .  .  .  }
   320  .  .  .  8: *ast.Field {
   321  .  .  .  .  Doc: nil
   322  .  .  .  .  Names: []*ast.Ident (len = 1) {
   323  .  .  .  .  .  0: *ast.Ident {
   324  .  .  .  .  .  .  NamePos: -
   325  .  .  .  .  .  .  Name: "services"
   326  .  .  .  .  .  .  Obj: *ast.Object {
   327  .  .  .  .  .  .  .  Kind: var
   328  .  .  .  .  .  .  .  Name: "services"
   329  .  .  .  .  .  .  .  Decl: *(obj @ 320)
   330  .  .  .  .  .  .  .  Data: nil
   331  .  .  .  .  .  .  .  Type: nil
   332  .  .  .  .  .  .  }
   333  .  .  .  .  .  }
   334  .  .  .  .  }
   335  .  .  .  .  Type: *ast.MapType {
   336  .  .  .  .  .  Map: -
   337  .  .  .  .  .  Key: *ast.Ident {
   338  .  .  .  .  .  .  NamePos: -
   339  .  .  .  .  .  .  Name: "string"
   340  .  .  .  .  .  .  Obj: nil
   341  .  .  .  .  .  }
   342  .  .  .  .  .  Value: *ast.Ident {
   343  .  .  .  .  .  .  NamePos: -
   344  .  .  .  .  .  .  Name: "any"
   345  .  .  .  .  .  .  Obj: nil
   346  .  .  .  .  .  }
   347  .  .  .  .  }
   348  .  .  .  .  Tag: nil
   349  .  .  .  .  Comment: nil
   350  .  .  .  }
   351  .  .  .  9: *ast.Field {
   352  .  .  .  .  Doc: nil
   353  .  .  .  .  Names: []*ast.Ident (len = 1) {
   354  .  .  .  .  .  0: *ast.Ident {
   355  .  .  .  .  .  .  NamePos: -
   356  .  .  .  .  .  .  Name: "servicesLocked"
   357  .  .  .  .  .  .  Obj: *ast.Object {
   358  .  .  .  .  .  .  .  Kind: var
   359  .  .  .  .  .  .  .  Name: "servicesLocked"
   360  .  .  .  .  .  .  .  Decl: *(obj @ 351)
   361  .  .  .  .  .  .  .  Data: nil
   362  .  .  .  .  .  .  .  Type: nil
   363  .  .  .  .  .  .  }
   364  .  .  .  .  .  }
   365  .  .  .  .  }
   366  .  .  .  .  Type: *ast.Ident {
   367  .  .  .  .  .  NamePos: -
   368  .  .  .  .  .  Name: "bool"
   369  .  .  .  .  .  Obj: nil
   370  .  .  .  .  }
   371  .  .  .  .  Tag: nil
   372  .  .  .  .  Comment: nil
   373  .  .  .  }
   374  .  .  }
   375  .  .  Closing: -
   376  .  }
   377  .  Incomplete: false
   378  }

```



#### Methods

- `ACTION(msg      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Message"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`:

- `Config()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Config"
     3  .  Obj: nil
     4  }
`: Config returns the registered Config service.

- `Core()      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Core"
     5  .  .  Obj: nil
     6  .  }
     7  }
`:

- `Display()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Display"
     3  .  Obj: nil
     4  }
`: Display returns the registered Display service.

- `RegisterAction(handler      0  *ast.FuncType {
     1  .  Func: -
     2  .  TypeParams: nil
     3  .  Params: *ast.FieldList {
     4  .  .  Opening: -
     5  .  .  List: []*ast.Field (len = 2) {
     6  .  .  .  0: *ast.Field {
     7  .  .  .  .  Doc: nil
     8  .  .  .  .  Names: nil
     9  .  .  .  .  Type: *ast.StarExpr {
    10  .  .  .  .  .  Star: -
    11  .  .  .  .  .  X: *ast.Ident {
    12  .  .  .  .  .  .  NamePos: -
    13  .  .  .  .  .  .  Name: "Core"
    14  .  .  .  .  .  .  Obj: nil
    15  .  .  .  .  .  }
    16  .  .  .  .  }
    17  .  .  .  .  Tag: nil
    18  .  .  .  .  Comment: nil
    19  .  .  .  }
    20  .  .  .  1: *ast.Field {
    21  .  .  .  .  Doc: nil
    22  .  .  .  .  Names: nil
    23  .  .  .  .  Type: *ast.Ident {
    24  .  .  .  .  .  NamePos: -
    25  .  .  .  .  .  Name: "Message"
    26  .  .  .  .  .  Obj: nil
    27  .  .  .  .  }
    28  .  .  .  .  Tag: nil
    29  .  .  .  .  Comment: nil
    30  .  .  .  }
    31  .  .  }
    32  .  .  Closing: -
    33  .  }
    34  .  Results: *ast.FieldList {
    35  .  .  Opening: -
    36  .  .  List: []*ast.Field (len = 1) {
    37  .  .  .  0: *ast.Field {
    38  .  .  .  .  Doc: nil
    39  .  .  .  .  Names: nil
    40  .  .  .  .  Type: *ast.Ident {
    41  .  .  .  .  .  NamePos: -
    42  .  .  .  .  .  Name: "error"
    43  .  .  .  .  .  Obj: nil
    44  .  .  .  .  }
    45  .  .  .  .  Tag: nil
    46  .  .  .  .  Comment: nil
    47  .  .  .  }
    48  .  .  }
    49  .  .  Closing: -
    50  .  }
    51  }
) `:

- `RegisterActions(handlers      0  *ast.Ellipsis {
     1  .  Ellipsis: -
     2  .  Elt: *ast.FuncType {
     3  .  .  Func: -
     4  .  .  TypeParams: nil
     5  .  .  Params: *ast.FieldList {
     6  .  .  .  Opening: -
     7  .  .  .  List: []*ast.Field (len = 2) {
     8  .  .  .  .  0: *ast.Field {
     9  .  .  .  .  .  Doc: nil
    10  .  .  .  .  .  Names: nil
    11  .  .  .  .  .  Type: *ast.StarExpr {
    12  .  .  .  .  .  .  Star: -
    13  .  .  .  .  .  .  X: *ast.Ident {
    14  .  .  .  .  .  .  .  NamePos: -
    15  .  .  .  .  .  .  .  Name: "Core"
    16  .  .  .  .  .  .  .  Obj: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  .  Tag: nil
    20  .  .  .  .  .  Comment: nil
    21  .  .  .  .  }
    22  .  .  .  .  1: *ast.Field {
    23  .  .  .  .  .  Doc: nil
    24  .  .  .  .  .  Names: nil
    25  .  .  .  .  .  Type: *ast.Ident {
    26  .  .  .  .  .  .  NamePos: -
    27  .  .  .  .  .  .  Name: "Message"
    28  .  .  .  .  .  .  Obj: nil
    29  .  .  .  .  .  }
    30  .  .  .  .  .  Tag: nil
    31  .  .  .  .  .  Comment: nil
    32  .  .  .  .  }
    33  .  .  .  }
    34  .  .  .  Closing: -
    35  .  .  }
    36  .  .  Results: *ast.FieldList {
    37  .  .  .  Opening: -
    38  .  .  .  List: []*ast.Field (len = 1) {
    39  .  .  .  .  0: *ast.Field {
    40  .  .  .  .  .  Doc: nil
    41  .  .  .  .  .  Names: nil
    42  .  .  .  .  .  Type: *ast.Ident {
    43  .  .  .  .  .  .  NamePos: -
    44  .  .  .  .  .  .  Name: "error"
    45  .  .  .  .  .  .  Obj: nil
    46  .  .  .  .  .  }
    47  .  .  .  .  .  Tag: nil
    48  .  .  .  .  .  Comment: nil
    49  .  .  .  .  }
    50  .  .  .  }
    51  .  .  .  Closing: -
    52  .  .  }
    53  .  }
    54  }
) `:

- `RegisterService(name      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, api      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "any"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`:

- `Service(name      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "any"
     3  .  Obj: nil
     4  }
`:

- `ServiceShutdown(ctx      0  *ast.SelectorExpr {
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
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`:

- `ServiceStartup(ctx      0  *ast.SelectorExpr {
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
, options      0  *ast.SelectorExpr {
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
`:



### `type Crypt`
```go
type Crypt      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "EncryptPGP"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "EncryptPGP"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: []*ast.Field (len = 3) {
    26  .  .  .  .  .  .  .  0: *ast.Field {
    27  .  .  .  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    29  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    30  .  .  .  .  .  .  .  .  .  .  NamePos: -
    31  .  .  .  .  .  .  .  .  .  .  Name: "writer"
    32  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    33  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    34  .  .  .  .  .  .  .  .  .  .  .  Name: "writer"
    35  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 26)
    36  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    37  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    38  .  .  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  .  }
    40  .  .  .  .  .  .  .  .  }
    41  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
    42  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    43  .  .  .  .  .  .  .  .  .  .  NamePos: -
    44  .  .  .  .  .  .  .  .  .  .  Name: "io"
    45  .  .  .  .  .  .  .  .  .  .  Obj: nil
    46  .  .  .  .  .  .  .  .  .  }
    47  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    48  .  .  .  .  .  .  .  .  .  .  NamePos: -
    49  .  .  .  .  .  .  .  .  .  .  Name: "Writer"
    50  .  .  .  .  .  .  .  .  .  .  Obj: nil
    51  .  .  .  .  .  .  .  .  .  }
    52  .  .  .  .  .  .  .  .  }
    53  .  .  .  .  .  .  .  .  Tag: nil
    54  .  .  .  .  .  .  .  .  Comment: nil
    55  .  .  .  .  .  .  .  }
    56  .  .  .  .  .  .  .  1: *ast.Field {
    57  .  .  .  .  .  .  .  .  Doc: nil
    58  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
    59  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    60  .  .  .  .  .  .  .  .  .  .  NamePos: -
    61  .  .  .  .  .  .  .  .  .  .  Name: "recipientPath"
    62  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    63  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    64  .  .  .  .  .  .  .  .  .  .  .  Name: "recipientPath"
    65  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 56)
    66  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    67  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    68  .  .  .  .  .  .  .  .  .  .  }
    69  .  .  .  .  .  .  .  .  .  }
    70  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
    71  .  .  .  .  .  .  .  .  .  .  NamePos: -
    72  .  .  .  .  .  .  .  .  .  .  Name: "data"
    73  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    74  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    75  .  .  .  .  .  .  .  .  .  .  .  Name: "data"
    76  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 56)
    77  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    78  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    79  .  .  .  .  .  .  .  .  .  .  }
    80  .  .  .  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  .  .  }
    82  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    83  .  .  .  .  .  .  .  .  .  NamePos: -
    84  .  .  .  .  .  .  .  .  .  Name: "string"
    85  .  .  .  .  .  .  .  .  .  Obj: nil
    86  .  .  .  .  .  .  .  .  }
    87  .  .  .  .  .  .  .  .  Tag: nil
    88  .  .  .  .  .  .  .  .  Comment: nil
    89  .  .  .  .  .  .  .  }
    90  .  .  .  .  .  .  .  2: *ast.Field {
    91  .  .  .  .  .  .  .  .  Doc: nil
    92  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
    93  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    94  .  .  .  .  .  .  .  .  .  .  NamePos: -
    95  .  .  .  .  .  .  .  .  .  .  Name: "signerPath"
    96  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    97  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    98  .  .  .  .  .  .  .  .  .  .  .  Name: "signerPath"
    99  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 90)
   100  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   101  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   102  .  .  .  .  .  .  .  .  .  .  }
   103  .  .  .  .  .  .  .  .  .  }
   104  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
   105  .  .  .  .  .  .  .  .  .  .  NamePos: -
   106  .  .  .  .  .  .  .  .  .  .  Name: "signerPassphrase"
   107  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   108  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   109  .  .  .  .  .  .  .  .  .  .  .  Name: "signerPassphrase"
   110  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 90)
   111  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   112  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   113  .  .  .  .  .  .  .  .  .  .  }
   114  .  .  .  .  .  .  .  .  .  }
   115  .  .  .  .  .  .  .  .  }
   116  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
   117  .  .  .  .  .  .  .  .  .  Star: -
   118  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   119  .  .  .  .  .  .  .  .  .  .  NamePos: -
   120  .  .  .  .  .  .  .  .  .  .  Name: "string"
   121  .  .  .  .  .  .  .  .  .  .  Obj: nil
   122  .  .  .  .  .  .  .  .  .  }
   123  .  .  .  .  .  .  .  .  }
   124  .  .  .  .  .  .  .  .  Tag: nil
   125  .  .  .  .  .  .  .  .  Comment: nil
   126  .  .  .  .  .  .  .  }
   127  .  .  .  .  .  .  }
   128  .  .  .  .  .  .  Closing: -
   129  .  .  .  .  .  }
   130  .  .  .  .  .  Results: *ast.FieldList {
   131  .  .  .  .  .  .  Opening: -
   132  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   133  .  .  .  .  .  .  .  0: *ast.Field {
   134  .  .  .  .  .  .  .  .  Doc: nil
   135  .  .  .  .  .  .  .  .  Names: nil
   136  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   137  .  .  .  .  .  .  .  .  .  NamePos: -
   138  .  .  .  .  .  .  .  .  .  Name: "string"
   139  .  .  .  .  .  .  .  .  .  Obj: nil
   140  .  .  .  .  .  .  .  .  }
   141  .  .  .  .  .  .  .  .  Tag: nil
   142  .  .  .  .  .  .  .  .  Comment: nil
   143  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  1: *ast.Field {
   145  .  .  .  .  .  .  .  .  Doc: nil
   146  .  .  .  .  .  .  .  .  Names: nil
   147  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   148  .  .  .  .  .  .  .  .  .  NamePos: -
   149  .  .  .  .  .  .  .  .  .  Name: "error"
   150  .  .  .  .  .  .  .  .  .  Obj: nil
   151  .  .  .  .  .  .  .  .  }
   152  .  .  .  .  .  .  .  .  Tag: nil
   153  .  .  .  .  .  .  .  .  Comment: nil
   154  .  .  .  .  .  .  .  }
   155  .  .  .  .  .  .  }
   156  .  .  .  .  .  .  Closing: -
   157  .  .  .  .  .  }
   158  .  .  .  .  }
   159  .  .  .  .  Tag: nil
   160  .  .  .  .  Comment: nil
   161  .  .  .  }
   162  .  .  .  1: *ast.Field {
   163  .  .  .  .  Doc: nil
   164  .  .  .  .  Names: []*ast.Ident (len = 1) {
   165  .  .  .  .  .  0: *ast.Ident {
   166  .  .  .  .  .  .  NamePos: -
   167  .  .  .  .  .  .  Name: "DecryptPGP"
   168  .  .  .  .  .  .  Obj: *ast.Object {
   169  .  .  .  .  .  .  .  Kind: func
   170  .  .  .  .  .  .  .  Name: "DecryptPGP"
   171  .  .  .  .  .  .  .  Decl: *(obj @ 162)
   172  .  .  .  .  .  .  .  Data: nil
   173  .  .  .  .  .  .  .  Type: nil
   174  .  .  .  .  .  .  }
   175  .  .  .  .  .  }
   176  .  .  .  .  }
   177  .  .  .  .  Type: *ast.FuncType {
   178  .  .  .  .  .  Func: -
   179  .  .  .  .  .  TypeParams: nil
   180  .  .  .  .  .  Params: *ast.FieldList {
   181  .  .  .  .  .  .  Opening: -
   182  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   183  .  .  .  .  .  .  .  0: *ast.Field {
   184  .  .  .  .  .  .  .  .  Doc: nil
   185  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 3) {
   186  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   187  .  .  .  .  .  .  .  .  .  .  NamePos: -
   188  .  .  .  .  .  .  .  .  .  .  Name: "recipientPath"
   189  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   190  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   191  .  .  .  .  .  .  .  .  .  .  .  Name: "recipientPath"
   192  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 183)
   193  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   194  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   195  .  .  .  .  .  .  .  .  .  .  }
   196  .  .  .  .  .  .  .  .  .  }
   197  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
   198  .  .  .  .  .  .  .  .  .  .  NamePos: -
   199  .  .  .  .  .  .  .  .  .  .  Name: "message"
   200  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   201  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   202  .  .  .  .  .  .  .  .  .  .  .  Name: "message"
   203  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 183)
   204  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   205  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   206  .  .  .  .  .  .  .  .  .  .  }
   207  .  .  .  .  .  .  .  .  .  }
   208  .  .  .  .  .  .  .  .  .  2: *ast.Ident {
   209  .  .  .  .  .  .  .  .  .  .  NamePos: -
   210  .  .  .  .  .  .  .  .  .  .  Name: "passphrase"
   211  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   212  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   213  .  .  .  .  .  .  .  .  .  .  .  Name: "passphrase"
   214  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 183)
   215  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   216  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   217  .  .  .  .  .  .  .  .  .  .  }
   218  .  .  .  .  .  .  .  .  .  }
   219  .  .  .  .  .  .  .  .  }
   220  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   221  .  .  .  .  .  .  .  .  .  NamePos: -
   222  .  .  .  .  .  .  .  .  .  Name: "string"
   223  .  .  .  .  .  .  .  .  .  Obj: nil
   224  .  .  .  .  .  .  .  .  }
   225  .  .  .  .  .  .  .  .  Tag: nil
   226  .  .  .  .  .  .  .  .  Comment: nil
   227  .  .  .  .  .  .  .  }
   228  .  .  .  .  .  .  .  1: *ast.Field {
   229  .  .  .  .  .  .  .  .  Doc: nil
   230  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   231  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   232  .  .  .  .  .  .  .  .  .  .  NamePos: -
   233  .  .  .  .  .  .  .  .  .  .  Name: "signerPath"
   234  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   235  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   236  .  .  .  .  .  .  .  .  .  .  .  Name: "signerPath"
   237  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 228)
   238  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   239  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   240  .  .  .  .  .  .  .  .  .  .  }
   241  .  .  .  .  .  .  .  .  .  }
   242  .  .  .  .  .  .  .  .  }
   243  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
   244  .  .  .  .  .  .  .  .  .  Star: -
   245  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   246  .  .  .  .  .  .  .  .  .  .  NamePos: -
   247  .  .  .  .  .  .  .  .  .  .  Name: "string"
   248  .  .  .  .  .  .  .  .  .  .  Obj: nil
   249  .  .  .  .  .  .  .  .  .  }
   250  .  .  .  .  .  .  .  .  }
   251  .  .  .  .  .  .  .  .  Tag: nil
   252  .  .  .  .  .  .  .  .  Comment: nil
   253  .  .  .  .  .  .  .  }
   254  .  .  .  .  .  .  }
   255  .  .  .  .  .  .  Closing: -
   256  .  .  .  .  .  }
   257  .  .  .  .  .  Results: *ast.FieldList {
   258  .  .  .  .  .  .  Opening: -
   259  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   260  .  .  .  .  .  .  .  0: *ast.Field {
   261  .  .  .  .  .  .  .  .  Doc: nil
   262  .  .  .  .  .  .  .  .  Names: nil
   263  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   264  .  .  .  .  .  .  .  .  .  NamePos: -
   265  .  .  .  .  .  .  .  .  .  Name: "string"
   266  .  .  .  .  .  .  .  .  .  Obj: nil
   267  .  .  .  .  .  .  .  .  }
   268  .  .  .  .  .  .  .  .  Tag: nil
   269  .  .  .  .  .  .  .  .  Comment: nil
   270  .  .  .  .  .  .  .  }
   271  .  .  .  .  .  .  .  1: *ast.Field {
   272  .  .  .  .  .  .  .  .  Doc: nil
   273  .  .  .  .  .  .  .  .  Names: nil
   274  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   275  .  .  .  .  .  .  .  .  .  NamePos: -
   276  .  .  .  .  .  .  .  .  .  Name: "error"
   277  .  .  .  .  .  .  .  .  .  Obj: nil
   278  .  .  .  .  .  .  .  .  }
   279  .  .  .  .  .  .  .  .  Tag: nil
   280  .  .  .  .  .  .  .  .  Comment: nil
   281  .  .  .  .  .  .  .  }
   282  .  .  .  .  .  .  }
   283  .  .  .  .  .  .  Closing: -
   284  .  .  .  .  .  }
   285  .  .  .  .  }
   286  .  .  .  .  Tag: nil
   287  .  .  .  .  Comment: nil
   288  .  .  .  }
   289  .  .  }
   290  .  .  Closing: -
   291  .  }
   292  .  Incomplete: false
   293  }

```
Crypt provides cryptographic functions.





### `type Display`
```go
type Display      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 1) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "OpenWindow"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "OpenWindow"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    26  .  .  .  .  .  .  .  0: *ast.Field {
    27  .  .  .  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    29  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    30  .  .  .  .  .  .  .  .  .  .  NamePos: -
    31  .  .  .  .  .  .  .  .  .  .  Name: "opts"
    32  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    33  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    34  .  .  .  .  .  .  .  .  .  .  .  Name: "opts"
    35  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 26)
    36  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    37  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    38  .  .  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  .  }
    40  .  .  .  .  .  .  .  .  }
    41  .  .  .  .  .  .  .  .  Type: *ast.Ellipsis {
    42  .  .  .  .  .  .  .  .  .  Ellipsis: -
    43  .  .  .  .  .  .  .  .  .  Elt: *ast.Ident {
    44  .  .  .  .  .  .  .  .  .  .  NamePos: -
    45  .  .  .  .  .  .  .  .  .  .  Name: "WindowOption"
    46  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    47  .  .  .  .  .  .  .  .  .  .  .  Kind: type
    48  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowOption"
    49  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    50  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    51  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    52  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    53  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowOption"
    54  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 46)
    55  .  .  .  .  .  .  .  .  .  .  .  .  }
    56  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
    57  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
    58  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.InterfaceType {
    59  .  .  .  .  .  .  .  .  .  .  .  .  .  Interface: -
    60  .  .  .  .  .  .  .  .  .  .  .  .  .  Methods: *ast.FieldList {
    61  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
    62  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    63  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
    64  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    65  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    66  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    67  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    68  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Apply"
    69  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    70  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: func
    71  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Apply"
    72  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 63)
    73  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    74  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    75  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    76  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    77  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    78  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.FuncType {
    79  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Func: -
    80  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
    81  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Params: *ast.FieldList {
    82  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
    83  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    84  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
    85  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    86  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: nil
    87  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
    88  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Star: -
    89  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    90  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    91  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
    92  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    93  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: type
    94  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
    95  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    96  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    97  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    98  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    99  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
   100  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 92)
   101  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   102  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
   103  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
   104  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.StructType {
   105  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Struct: -
   106  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Fields: *ast.FieldList {
   107  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
   108  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 5) {
   109  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
   110  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   111  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   112  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   113  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   114  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
   115  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   116  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   117  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
   118  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 109)
   119  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   120  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   121  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   122  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   123  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   124  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   125  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   126  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   127  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   128  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   129  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   130  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   131  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  1: *ast.Field {
   133  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   134  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   135  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   136  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   137  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Title"
   138  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   139  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   140  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Title"
   141  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 132)
   142  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   143  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   144  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   145  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   146  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   147  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   148  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   149  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   150  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   151  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   152  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   153  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   154  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   155  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  2: *ast.Field {
   156  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   157  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   158  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   159  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   160  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "URL"
   161  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   162  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   163  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "URL"
   164  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 155)
   165  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   166  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   167  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   168  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   169  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   170  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   171  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   172  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   173  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   174  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   175  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   176  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   177  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   178  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  3: *ast.Field {
   179  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   180  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   181  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   182  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   183  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Width"
   184  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   185  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   186  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Width"
   187  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 178)
   188  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   189  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   190  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   191  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   192  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   193  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   194  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   195  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "int"
   196  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   197  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   198  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   199  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   200  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   201  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  4: *ast.Field {
   202  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   203  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   204  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   205  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   206  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Height"
   207  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   208  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   209  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Height"
   210  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 201)
   211  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   212  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   213  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   214  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   215  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   216  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   217  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   218  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "int"
   219  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   220  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   221  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   222  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: *ast.CommentGroup {
   223  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   224  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Comment {
   225  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Slash: -
   226  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Text: "// Add other common window options here as needed"
   227  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   228  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   229  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   230  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   231  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   232  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   233  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   234  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Incomplete: false
   235  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   236  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   237  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   238  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   239  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   240  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   241  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   242  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   243  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   244  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   245  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   246  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   247  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   248  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   249  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Results: nil
   250  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   251  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   252  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   253  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   254  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   255  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   256  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   257  .  .  .  .  .  .  .  .  .  .  .  .  .  Incomplete: false
   258  .  .  .  .  .  .  .  .  .  .  .  .  }
   259  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   260  .  .  .  .  .  .  .  .  .  .  .  }
   261  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   262  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   263  .  .  .  .  .  .  .  .  .  .  }
   264  .  .  .  .  .  .  .  .  .  }
   265  .  .  .  .  .  .  .  .  }
   266  .  .  .  .  .  .  .  .  Tag: nil
   267  .  .  .  .  .  .  .  .  Comment: nil
   268  .  .  .  .  .  .  .  }
   269  .  .  .  .  .  .  }
   270  .  .  .  .  .  .  Closing: -
   271  .  .  .  .  .  }
   272  .  .  .  .  .  Results: *ast.FieldList {
   273  .  .  .  .  .  .  Opening: -
   274  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   275  .  .  .  .  .  .  .  0: *ast.Field {
   276  .  .  .  .  .  .  .  .  Doc: nil
   277  .  .  .  .  .  .  .  .  Names: nil
   278  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   279  .  .  .  .  .  .  .  .  .  NamePos: -
   280  .  .  .  .  .  .  .  .  .  Name: "error"
   281  .  .  .  .  .  .  .  .  .  Obj: nil
   282  .  .  .  .  .  .  .  .  }
   283  .  .  .  .  .  .  .  .  Tag: nil
   284  .  .  .  .  .  .  .  .  Comment: nil
   285  .  .  .  .  .  .  .  }
   286  .  .  .  .  .  .  }
   287  .  .  .  .  .  .  Closing: -
   288  .  .  .  .  .  }
   289  .  .  .  .  }
   290  .  .  .  .  Tag: nil
   291  .  .  .  .  Comment: nil
   292  .  .  .  }
   293  .  .  }
   294  .  .  Closing: -
   295  .  }
   296  .  Incomplete: false
   297  }

```
Display manages windows and UI.





### `type Help`
```go
type Help      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Show"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "Show"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: nil
    26  .  .  .  .  .  .  Closing: -
    27  .  .  .  .  .  }
    28  .  .  .  .  .  Results: *ast.FieldList {
    29  .  .  .  .  .  .  Opening: -
    30  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    31  .  .  .  .  .  .  .  0: *ast.Field {
    32  .  .  .  .  .  .  .  .  Doc: nil
    33  .  .  .  .  .  .  .  .  Names: nil
    34  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    35  .  .  .  .  .  .  .  .  .  NamePos: -
    36  .  .  .  .  .  .  .  .  .  Name: "error"
    37  .  .  .  .  .  .  .  .  .  Obj: nil
    38  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  Tag: nil
    40  .  .  .  .  .  .  .  .  Comment: nil
    41  .  .  .  .  .  .  .  }
    42  .  .  .  .  .  .  }
    43  .  .  .  .  .  .  Closing: -
    44  .  .  .  .  .  }
    45  .  .  .  .  }
    46  .  .  .  .  Tag: nil
    47  .  .  .  .  Comment: nil
    48  .  .  .  }
    49  .  .  .  1: *ast.Field {
    50  .  .  .  .  Doc: nil
    51  .  .  .  .  Names: []*ast.Ident (len = 1) {
    52  .  .  .  .  .  0: *ast.Ident {
    53  .  .  .  .  .  .  NamePos: -
    54  .  .  .  .  .  .  Name: "ShowAt"
    55  .  .  .  .  .  .  Obj: *ast.Object {
    56  .  .  .  .  .  .  .  Kind: func
    57  .  .  .  .  .  .  .  Name: "ShowAt"
    58  .  .  .  .  .  .  .  Decl: *(obj @ 49)
    59  .  .  .  .  .  .  .  Data: nil
    60  .  .  .  .  .  .  .  Type: nil
    61  .  .  .  .  .  .  }
    62  .  .  .  .  .  }
    63  .  .  .  .  }
    64  .  .  .  .  Type: *ast.FuncType {
    65  .  .  .  .  .  Func: -
    66  .  .  .  .  .  TypeParams: nil
    67  .  .  .  .  .  Params: *ast.FieldList {
    68  .  .  .  .  .  .  Opening: -
    69  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    70  .  .  .  .  .  .  .  0: *ast.Field {
    71  .  .  .  .  .  .  .  .  Doc: nil
    72  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    73  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    74  .  .  .  .  .  .  .  .  .  .  NamePos: -
    75  .  .  .  .  .  .  .  .  .  .  Name: "anchor"
    76  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    77  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    78  .  .  .  .  .  .  .  .  .  .  .  Name: "anchor"
    79  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 70)
    80  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    81  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    82  .  .  .  .  .  .  .  .  .  .  }
    83  .  .  .  .  .  .  .  .  .  }
    84  .  .  .  .  .  .  .  .  }
    85  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    86  .  .  .  .  .  .  .  .  .  NamePos: -
    87  .  .  .  .  .  .  .  .  .  Name: "string"
    88  .  .  .  .  .  .  .  .  .  Obj: nil
    89  .  .  .  .  .  .  .  .  }
    90  .  .  .  .  .  .  .  .  Tag: nil
    91  .  .  .  .  .  .  .  .  Comment: nil
    92  .  .  .  .  .  .  .  }
    93  .  .  .  .  .  .  }
    94  .  .  .  .  .  .  Closing: -
    95  .  .  .  .  .  }
    96  .  .  .  .  .  Results: *ast.FieldList {
    97  .  .  .  .  .  .  Opening: -
    98  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    99  .  .  .  .  .  .  .  0: *ast.Field {
   100  .  .  .  .  .  .  .  .  Doc: nil
   101  .  .  .  .  .  .  .  .  Names: nil
   102  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   103  .  .  .  .  .  .  .  .  .  NamePos: -
   104  .  .  .  .  .  .  .  .  .  Name: "error"
   105  .  .  .  .  .  .  .  .  .  Obj: nil
   106  .  .  .  .  .  .  .  .  }
   107  .  .  .  .  .  .  .  .  Tag: nil
   108  .  .  .  .  .  .  .  .  Comment: nil
   109  .  .  .  .  .  .  .  }
   110  .  .  .  .  .  .  }
   111  .  .  .  .  .  .  Closing: -
   112  .  .  .  .  .  }
   113  .  .  .  .  }
   114  .  .  .  .  Tag: nil
   115  .  .  .  .  Comment: nil
   116  .  .  .  }
   117  .  .  }
   118  .  .  Closing: -
   119  .  }
   120  .  Incomplete: false
   121  }

```
Help manages the in-app documentation and help system.





### `type I18n`
```go
type I18n      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: *ast.CommentGroup {
     7  .  .  .  .  .  List: []*ast.Comment (len = 1) {
     8  .  .  .  .  .  .  0: *ast.Comment {
     9  .  .  .  .  .  .  .  Slash: -
    10  .  .  .  .  .  .  .  Text: "// Translate returns the translated string for the given key."
    11  .  .  .  .  .  .  }
    12  .  .  .  .  .  }
    13  .  .  .  .  }
    14  .  .  .  .  Names: []*ast.Ident (len = 1) {
    15  .  .  .  .  .  0: *ast.Ident {
    16  .  .  .  .  .  .  NamePos: -
    17  .  .  .  .  .  .  Name: "Translate"
    18  .  .  .  .  .  .  Obj: *ast.Object {
    19  .  .  .  .  .  .  .  Kind: func
    20  .  .  .  .  .  .  .  Name: "Translate"
    21  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    22  .  .  .  .  .  .  .  Data: nil
    23  .  .  .  .  .  .  .  Type: nil
    24  .  .  .  .  .  .  }
    25  .  .  .  .  .  }
    26  .  .  .  .  }
    27  .  .  .  .  Type: *ast.FuncType {
    28  .  .  .  .  .  Func: -
    29  .  .  .  .  .  TypeParams: nil
    30  .  .  .  .  .  Params: *ast.FieldList {
    31  .  .  .  .  .  .  Opening: -
    32  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    33  .  .  .  .  .  .  .  0: *ast.Field {
    34  .  .  .  .  .  .  .  .  Doc: nil
    35  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    36  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    37  .  .  .  .  .  .  .  .  .  .  NamePos: -
    38  .  .  .  .  .  .  .  .  .  .  Name: "key"
    39  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    40  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    41  .  .  .  .  .  .  .  .  .  .  .  Name: "key"
    42  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 33)
    43  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    44  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    45  .  .  .  .  .  .  .  .  .  .  }
    46  .  .  .  .  .  .  .  .  .  }
    47  .  .  .  .  .  .  .  .  }
    48  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    49  .  .  .  .  .  .  .  .  .  NamePos: -
    50  .  .  .  .  .  .  .  .  .  Name: "string"
    51  .  .  .  .  .  .  .  .  .  Obj: nil
    52  .  .  .  .  .  .  .  .  }
    53  .  .  .  .  .  .  .  .  Tag: nil
    54  .  .  .  .  .  .  .  .  Comment: nil
    55  .  .  .  .  .  .  .  }
    56  .  .  .  .  .  .  }
    57  .  .  .  .  .  .  Closing: -
    58  .  .  .  .  .  }
    59  .  .  .  .  .  Results: *ast.FieldList {
    60  .  .  .  .  .  .  Opening: -
    61  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    62  .  .  .  .  .  .  .  0: *ast.Field {
    63  .  .  .  .  .  .  .  .  Doc: nil
    64  .  .  .  .  .  .  .  .  Names: nil
    65  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    66  .  .  .  .  .  .  .  .  .  NamePos: -
    67  .  .  .  .  .  .  .  .  .  Name: "string"
    68  .  .  .  .  .  .  .  .  .  Obj: nil
    69  .  .  .  .  .  .  .  .  }
    70  .  .  .  .  .  .  .  .  Tag: nil
    71  .  .  .  .  .  .  .  .  Comment: nil
    72  .  .  .  .  .  .  .  }
    73  .  .  .  .  .  .  }
    74  .  .  .  .  .  .  Closing: -
    75  .  .  .  .  .  }
    76  .  .  .  .  }
    77  .  .  .  .  Tag: nil
    78  .  .  .  .  Comment: nil
    79  .  .  .  }
    80  .  .  .  1: *ast.Field {
    81  .  .  .  .  Doc: *ast.CommentGroup {
    82  .  .  .  .  .  List: []*ast.Comment (len = 1) {
    83  .  .  .  .  .  .  0: *ast.Comment {
    84  .  .  .  .  .  .  .  Slash: -
    85  .  .  .  .  .  .  .  Text: "// SetLanguage changes the active language."
    86  .  .  .  .  .  .  }
    87  .  .  .  .  .  }
    88  .  .  .  .  }
    89  .  .  .  .  Names: []*ast.Ident (len = 1) {
    90  .  .  .  .  .  0: *ast.Ident {
    91  .  .  .  .  .  .  NamePos: -
    92  .  .  .  .  .  .  Name: "SetLanguage"
    93  .  .  .  .  .  .  Obj: *ast.Object {
    94  .  .  .  .  .  .  .  Kind: func
    95  .  .  .  .  .  .  .  Name: "SetLanguage"
    96  .  .  .  .  .  .  .  Decl: *(obj @ 80)
    97  .  .  .  .  .  .  .  Data: nil
    98  .  .  .  .  .  .  .  Type: nil
    99  .  .  .  .  .  .  }
   100  .  .  .  .  .  }
   101  .  .  .  .  }
   102  .  .  .  .  Type: *ast.FuncType {
   103  .  .  .  .  .  Func: -
   104  .  .  .  .  .  TypeParams: nil
   105  .  .  .  .  .  Params: *ast.FieldList {
   106  .  .  .  .  .  .  Opening: -
   107  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   108  .  .  .  .  .  .  .  0: *ast.Field {
   109  .  .  .  .  .  .  .  .  Doc: nil
   110  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   111  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   112  .  .  .  .  .  .  .  .  .  .  NamePos: -
   113  .  .  .  .  .  .  .  .  .  .  Name: "lang"
   114  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   115  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   116  .  .  .  .  .  .  .  .  .  .  .  Name: "lang"
   117  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 108)
   118  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   119  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   120  .  .  .  .  .  .  .  .  .  .  }
   121  .  .  .  .  .  .  .  .  .  }
   122  .  .  .  .  .  .  .  .  }
   123  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   124  .  .  .  .  .  .  .  .  .  NamePos: -
   125  .  .  .  .  .  .  .  .  .  Name: "string"
   126  .  .  .  .  .  .  .  .  .  Obj: nil
   127  .  .  .  .  .  .  .  .  }
   128  .  .  .  .  .  .  .  .  Tag: nil
   129  .  .  .  .  .  .  .  .  Comment: nil
   130  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  Closing: -
   133  .  .  .  .  .  }
   134  .  .  .  .  .  Results: *ast.FieldList {
   135  .  .  .  .  .  .  Opening: -
   136  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   137  .  .  .  .  .  .  .  0: *ast.Field {
   138  .  .  .  .  .  .  .  .  Doc: nil
   139  .  .  .  .  .  .  .  .  Names: nil
   140  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   141  .  .  .  .  .  .  .  .  .  NamePos: -
   142  .  .  .  .  .  .  .  .  .  Name: "error"
   143  .  .  .  .  .  .  .  .  .  Obj: nil
   144  .  .  .  .  .  .  .  .  }
   145  .  .  .  .  .  .  .  .  Tag: nil
   146  .  .  .  .  .  .  .  .  Comment: nil
   147  .  .  .  .  .  .  .  }
   148  .  .  .  .  .  .  }
   149  .  .  .  .  .  .  Closing: -
   150  .  .  .  .  .  }
   151  .  .  .  .  }
   152  .  .  .  .  Tag: nil
   153  .  .  .  .  Comment: nil
   154  .  .  .  }
   155  .  .  }
   156  .  .  Closing: -
   157  .  }
   158  .  Incomplete: false
   159  }

```
I18n provides internationalization and localization services.





### `type Message`
```go
type Message      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: nil
     5  .  .  Closing: -
     6  .  }
     7  .  Incomplete: false
     8  }

```





### `type Option`
```go
type Option      0  *ast.FuncType {
     1  .  Func: -
     2  .  TypeParams: nil
     3  .  Params: *ast.FieldList {
     4  .  .  Opening: -
     5  .  .  List: []*ast.Field (len = 1) {
     6  .  .  .  0: *ast.Field {
     7  .  .  .  .  Doc: nil
     8  .  .  .  .  Names: nil
     9  .  .  .  .  Type: *ast.StarExpr {
    10  .  .  .  .  .  Star: -
    11  .  .  .  .  .  X: *ast.Ident {
    12  .  .  .  .  .  .  NamePos: -
    13  .  .  .  .  .  .  Name: "Core"
    14  .  .  .  .  .  .  Obj: *ast.Object {
    15  .  .  .  .  .  .  .  Kind: type
    16  .  .  .  .  .  .  .  Name: "Core"
    17  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    18  .  .  .  .  .  .  .  .  Doc: nil
    19  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    20  .  .  .  .  .  .  .  .  .  NamePos: -
    21  .  .  .  .  .  .  .  .  .  Name: "Core"
    22  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 14)
    23  .  .  .  .  .  .  .  .  }
    24  .  .  .  .  .  .  .  .  TypeParams: nil
    25  .  .  .  .  .  .  .  .  Assign: -
    26  .  .  .  .  .  .  .  .  Type: *ast.StructType {
    27  .  .  .  .  .  .  .  .  .  Struct: -
    28  .  .  .  .  .  .  .  .  .  Fields: *ast.FieldList {
    29  .  .  .  .  .  .  .  .  .  .  Opening: -
    30  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 10) {
    31  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
    32  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    33  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    34  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    35  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    36  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "once"
    37  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    38  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    39  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "once"
    40  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 31)
    41  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    42  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    43  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    44  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    45  .  .  .  .  .  .  .  .  .  .  .  .  }
    46  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
    47  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    48  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    49  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "sync"
    50  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    51  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    52  .  .  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    53  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    54  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Once"
    55  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    56  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    57  .  .  .  .  .  .  .  .  .  .  .  .  }
    58  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
    59  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
    60  .  .  .  .  .  .  .  .  .  .  .  }
    61  .  .  .  .  .  .  .  .  .  .  .  1: *ast.Field {
    62  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    63  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    64  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    65  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    66  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "initErr"
    67  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    68  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    69  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "initErr"
    70  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 61)
    71  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    72  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    73  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    74  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    75  .  .  .  .  .  .  .  .  .  .  .  .  }
    76  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    77  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    78  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "error"
    79  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    80  .  .  .  .  .  .  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
    82  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
    83  .  .  .  .  .  .  .  .  .  .  .  }
    84  .  .  .  .  .  .  .  .  .  .  .  2: *ast.Field {
    85  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    86  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    87  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    88  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    89  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "App"
    90  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    91  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    92  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "App"
    93  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 84)
    94  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    95  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    96  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    97  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    98  .  .  .  .  .  .  .  .  .  .  .  .  }
    99  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
   100  .  .  .  .  .  .  .  .  .  .  .  .  .  Star: -
   101  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.SelectorExpr {
   102  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   103  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   104  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "application"
   105  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   106  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   107  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   108  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   109  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "App"
   110  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   111  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   112  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   113  .  .  .  .  .  .  .  .  .  .  .  .  }
   114  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   115  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   116  .  .  .  .  .  .  .  .  .  .  .  }
   117  .  .  .  .  .  .  .  .  .  .  .  3: *ast.Field {
   118  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   119  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   120  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   121  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   122  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "assets"
   123  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   124  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   125  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "assets"
   126  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 117)
   127  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   128  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   129  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   130  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
   133  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   134  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   135  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "embed"
   136  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   137  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   138  .  .  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   139  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   140  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "FS"
   141  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   142  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   143  .  .  .  .  .  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   145  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   146  .  .  .  .  .  .  .  .  .  .  .  }
   147  .  .  .  .  .  .  .  .  .  .  .  4: *ast.Field {
   148  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   149  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   150  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   151  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   152  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "serviceLock"
   153  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   154  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   155  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "serviceLock"
   156  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 147)
   157  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   158  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   159  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   160  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   161  .  .  .  .  .  .  .  .  .  .  .  .  }
   162  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   163  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   164  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "bool"
   165  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   166  .  .  .  .  .  .  .  .  .  .  .  .  }
   167  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   168  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   169  .  .  .  .  .  .  .  .  .  .  .  }
   170  .  .  .  .  .  .  .  .  .  .  .  5: *ast.Field {
   171  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   172  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   173  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   174  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   175  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "ipcMu"
   176  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   177  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   178  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "ipcMu"
   179  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 170)
   180  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   181  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   182  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   183  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   184  .  .  .  .  .  .  .  .  .  .  .  .  }
   185  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
   186  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   187  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   188  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "sync"
   189  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   190  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   191  .  .  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   192  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   193  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "RWMutex"
   194  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   195  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   196  .  .  .  .  .  .  .  .  .  .  .  .  }
   197  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   198  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   199  .  .  .  .  .  .  .  .  .  .  .  }
   200  .  .  .  .  .  .  .  .  .  .  .  6: *ast.Field {
   201  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   202  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   203  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   204  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   205  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "ipcHandlers"
   206  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   207  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   208  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "ipcHandlers"
   209  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 200)
   210  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   211  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   212  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   213  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   214  .  .  .  .  .  .  .  .  .  .  .  .  }
   215  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.ArrayType {
   216  .  .  .  .  .  .  .  .  .  .  .  .  .  Lbrack: -
   217  .  .  .  .  .  .  .  .  .  .  .  .  .  Len: nil
   218  .  .  .  .  .  .  .  .  .  .  .  .  .  Elt: *ast.FuncType {
   219  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Func: -
   220  .  .  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
   221  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Params: *ast.FieldList {
   222  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
   223  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   224  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
   225  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   226  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: nil
   227  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
   228  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Star: -
   229  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   230  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   231  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Core"
   232  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 14)
   233  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   234  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   235  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   236  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   237  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   238  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  1: *ast.Field {
   239  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   240  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: nil
   241  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   242  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   243  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   244  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   245  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: type
   246  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   247  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
   248  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   249  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
   250  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   251  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Message"
   252  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 244)
   253  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   254  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
   255  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
   256  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.InterfaceType {
   257  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Interface: -
   258  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Methods: *ast.FieldList {
   259  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
   260  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: nil
   261  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   262  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   263  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Incomplete: false
   264  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   265  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   266  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   267  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   268  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   269  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   270  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   271  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   272  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   273  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   274  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   275  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   276  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   277  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Results: *ast.FieldList {
   278  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
   279  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   280  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
   281  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   282  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: nil
   283  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   284  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   285  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "error"
   286  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   287  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   288  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   289  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   290  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   291  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   292  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   293  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   294  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   295  .  .  .  .  .  .  .  .  .  .  .  .  }
   296  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   297  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   298  .  .  .  .  .  .  .  .  .  .  .  }
   299  .  .  .  .  .  .  .  .  .  .  .  7: *ast.Field {
   300  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   301  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   302  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   303  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   304  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "serviceMu"
   305  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   306  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   307  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "serviceMu"
   308  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 299)
   309  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   310  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   311  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   312  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   313  .  .  .  .  .  .  .  .  .  .  .  .  }
   314  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.SelectorExpr {
   315  .  .  .  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
   316  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   317  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "sync"
   318  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   319  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   320  .  .  .  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
   321  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   322  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "RWMutex"
   323  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   324  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   325  .  .  .  .  .  .  .  .  .  .  .  .  }
   326  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   327  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   328  .  .  .  .  .  .  .  .  .  .  .  }
   329  .  .  .  .  .  .  .  .  .  .  .  8: *ast.Field {
   330  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   331  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   332  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   333  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   334  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "services"
   335  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   336  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   337  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "services"
   338  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 329)
   339  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   340  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   341  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   342  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   343  .  .  .  .  .  .  .  .  .  .  .  .  }
   344  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.MapType {
   345  .  .  .  .  .  .  .  .  .  .  .  .  .  Map: -
   346  .  .  .  .  .  .  .  .  .  .  .  .  .  Key: *ast.Ident {
   347  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   348  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   349  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   350  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   351  .  .  .  .  .  .  .  .  .  .  .  .  .  Value: *ast.Ident {
   352  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   353  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "any"
   354  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   355  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   356  .  .  .  .  .  .  .  .  .  .  .  .  }
   357  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   358  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   359  .  .  .  .  .  .  .  .  .  .  .  }
   360  .  .  .  .  .  .  .  .  .  .  .  9: *ast.Field {
   361  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   362  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   363  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   364  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   365  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "servicesLocked"
   366  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   367  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   368  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "servicesLocked"
   369  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 360)
   370  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   371  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   372  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   373  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   374  .  .  .  .  .  .  .  .  .  .  .  .  }
   375  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   376  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   377  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "bool"
   378  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   379  .  .  .  .  .  .  .  .  .  .  .  .  }
   380  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   381  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   382  .  .  .  .  .  .  .  .  .  .  .  }
   383  .  .  .  .  .  .  .  .  .  .  }
   384  .  .  .  .  .  .  .  .  .  .  Closing: -
   385  .  .  .  .  .  .  .  .  .  }
   386  .  .  .  .  .  .  .  .  .  Incomplete: false
   387  .  .  .  .  .  .  .  .  }
   388  .  .  .  .  .  .  .  .  Comment: nil
   389  .  .  .  .  .  .  .  }
   390  .  .  .  .  .  .  .  Data: nil
   391  .  .  .  .  .  .  .  Type: nil
   392  .  .  .  .  .  .  }
   393  .  .  .  .  .  }
   394  .  .  .  .  }
   395  .  .  .  .  Tag: nil
   396  .  .  .  .  Comment: nil
   397  .  .  .  }
   398  .  .  }
   399  .  .  Closing: -
   400  .  }
   401  .  Results: *ast.FieldList {
   402  .  .  Opening: -
   403  .  .  List: []*ast.Field (len = 1) {
   404  .  .  .  0: *ast.Field {
   405  .  .  .  .  Doc: nil
   406  .  .  .  .  Names: nil
   407  .  .  .  .  Type: *ast.Ident {
   408  .  .  .  .  .  NamePos: -
   409  .  .  .  .  .  Name: "error"
   410  .  .  .  .  .  Obj: nil
   411  .  .  .  .  }
   412  .  .  .  .  Tag: nil
   413  .  .  .  .  Comment: nil
   414  .  .  .  }
   415  .  .  }
   416  .  .  Closing: -
   417  .  }
   418  }

```





### `type Runtime`
```go
type Runtime      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "core"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "core"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.StarExpr {
    21  .  .  .  .  .  Star: -
    22  .  .  .  .  .  X: *ast.Ident {
    23  .  .  .  .  .  .  NamePos: -
    24  .  .  .  .  .  .  Name: "Core"
    25  .  .  .  .  .  .  Obj: nil
    26  .  .  .  .  .  }
    27  .  .  .  .  }
    28  .  .  .  .  Tag: nil
    29  .  .  .  .  Comment: nil
    30  .  .  .  }
    31  .  .  .  1: *ast.Field {
    32  .  .  .  .  Doc: nil
    33  .  .  .  .  Names: []*ast.Ident (len = 1) {
    34  .  .  .  .  .  0: *ast.Ident {
    35  .  .  .  .  .  .  NamePos: -
    36  .  .  .  .  .  .  Name: "opts"
    37  .  .  .  .  .  .  Obj: *ast.Object {
    38  .  .  .  .  .  .  .  Kind: var
    39  .  .  .  .  .  .  .  Name: "opts"
    40  .  .  .  .  .  .  .  Decl: *(obj @ 31)
    41  .  .  .  .  .  .  .  Data: nil
    42  .  .  .  .  .  .  .  Type: nil
    43  .  .  .  .  .  .  }
    44  .  .  .  .  .  }
    45  .  .  .  .  }
    46  .  .  .  .  Type: *ast.Ident {
    47  .  .  .  .  .  NamePos: -
    48  .  .  .  .  .  Name: "T"
    49  .  .  .  .  .  Obj: *ast.Object {
    50  .  .  .  .  .  .  Kind: type
    51  .  .  .  .  .  .  Name: "T"
    52  .  .  .  .  .  .  Decl: *ast.Field {
    53  .  .  .  .  .  .  .  Doc: nil
    54  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    55  .  .  .  .  .  .  .  .  0: *ast.Ident {
    56  .  .  .  .  .  .  .  .  .  NamePos: -
    57  .  .  .  .  .  .  .  .  .  Name: "T"
    58  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 49)
    59  .  .  .  .  .  .  .  .  }
    60  .  .  .  .  .  .  .  }
    61  .  .  .  .  .  .  .  Type: *ast.Ident {
    62  .  .  .  .  .  .  .  .  NamePos: -
    63  .  .  .  .  .  .  .  .  Name: "any"
    64  .  .  .  .  .  .  .  .  Obj: nil
    65  .  .  .  .  .  .  .  }
    66  .  .  .  .  .  .  .  Tag: nil
    67  .  .  .  .  .  .  .  Comment: nil
    68  .  .  .  .  .  .  }
    69  .  .  .  .  .  .  Data: nil
    70  .  .  .  .  .  .  Type: nil
    71  .  .  .  .  .  }
    72  .  .  .  .  }
    73  .  .  .  .  Tag: nil
    74  .  .  .  .  Comment: nil
    75  .  .  .  }
    76  .  .  }
    77  .  .  Closing: -
    78  .  }
    79  .  Incomplete: false
    80  }

```
Runtime is a helper struct embedded in services to provide access to the core application.



#### Methods

- `Config()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Config"
     3  .  Obj: nil
     4  }
`: Config returns the registered Config service from the core application.

- `Core()      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Core"
     5  .  .  Obj: nil
     6  .  }
     7  }
`: Core returns the central core instance.



### `type WindowConfig`
```go
type WindowConfig      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 5) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Name"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "Name"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.Ident {
    21  .  .  .  .  .  NamePos: -
    22  .  .  .  .  .  Name: "string"
    23  .  .  .  .  .  Obj: nil
    24  .  .  .  .  }
    25  .  .  .  .  Tag: nil
    26  .  .  .  .  Comment: nil
    27  .  .  .  }
    28  .  .  .  1: *ast.Field {
    29  .  .  .  .  Doc: nil
    30  .  .  .  .  Names: []*ast.Ident (len = 1) {
    31  .  .  .  .  .  0: *ast.Ident {
    32  .  .  .  .  .  .  NamePos: -
    33  .  .  .  .  .  .  Name: "Title"
    34  .  .  .  .  .  .  Obj: *ast.Object {
    35  .  .  .  .  .  .  .  Kind: var
    36  .  .  .  .  .  .  .  Name: "Title"
    37  .  .  .  .  .  .  .  Decl: *(obj @ 28)
    38  .  .  .  .  .  .  .  Data: nil
    39  .  .  .  .  .  .  .  Type: nil
    40  .  .  .  .  .  .  }
    41  .  .  .  .  .  }
    42  .  .  .  .  }
    43  .  .  .  .  Type: *ast.Ident {
    44  .  .  .  .  .  NamePos: -
    45  .  .  .  .  .  Name: "string"
    46  .  .  .  .  .  Obj: nil
    47  .  .  .  .  }
    48  .  .  .  .  Tag: nil
    49  .  .  .  .  Comment: nil
    50  .  .  .  }
    51  .  .  .  2: *ast.Field {
    52  .  .  .  .  Doc: nil
    53  .  .  .  .  Names: []*ast.Ident (len = 1) {
    54  .  .  .  .  .  0: *ast.Ident {
    55  .  .  .  .  .  .  NamePos: -
    56  .  .  .  .  .  .  Name: "URL"
    57  .  .  .  .  .  .  Obj: *ast.Object {
    58  .  .  .  .  .  .  .  Kind: var
    59  .  .  .  .  .  .  .  Name: "URL"
    60  .  .  .  .  .  .  .  Decl: *(obj @ 51)
    61  .  .  .  .  .  .  .  Data: nil
    62  .  .  .  .  .  .  .  Type: nil
    63  .  .  .  .  .  .  }
    64  .  .  .  .  .  }
    65  .  .  .  .  }
    66  .  .  .  .  Type: *ast.Ident {
    67  .  .  .  .  .  NamePos: -
    68  .  .  .  .  .  Name: "string"
    69  .  .  .  .  .  Obj: nil
    70  .  .  .  .  }
    71  .  .  .  .  Tag: nil
    72  .  .  .  .  Comment: nil
    73  .  .  .  }
    74  .  .  .  3: *ast.Field {
    75  .  .  .  .  Doc: nil
    76  .  .  .  .  Names: []*ast.Ident (len = 1) {
    77  .  .  .  .  .  0: *ast.Ident {
    78  .  .  .  .  .  .  NamePos: -
    79  .  .  .  .  .  .  Name: "Width"
    80  .  .  .  .  .  .  Obj: *ast.Object {
    81  .  .  .  .  .  .  .  Kind: var
    82  .  .  .  .  .  .  .  Name: "Width"
    83  .  .  .  .  .  .  .  Decl: *(obj @ 74)
    84  .  .  .  .  .  .  .  Data: nil
    85  .  .  .  .  .  .  .  Type: nil
    86  .  .  .  .  .  .  }
    87  .  .  .  .  .  }
    88  .  .  .  .  }
    89  .  .  .  .  Type: *ast.Ident {
    90  .  .  .  .  .  NamePos: -
    91  .  .  .  .  .  Name: "int"
    92  .  .  .  .  .  Obj: nil
    93  .  .  .  .  }
    94  .  .  .  .  Tag: nil
    95  .  .  .  .  Comment: nil
    96  .  .  .  }
    97  .  .  .  4: *ast.Field {
    98  .  .  .  .  Doc: nil
    99  .  .  .  .  Names: []*ast.Ident (len = 1) {
   100  .  .  .  .  .  0: *ast.Ident {
   101  .  .  .  .  .  .  NamePos: -
   102  .  .  .  .  .  .  Name: "Height"
   103  .  .  .  .  .  .  Obj: *ast.Object {
   104  .  .  .  .  .  .  .  Kind: var
   105  .  .  .  .  .  .  .  Name: "Height"
   106  .  .  .  .  .  .  .  Decl: *(obj @ 97)
   107  .  .  .  .  .  .  .  Data: nil
   108  .  .  .  .  .  .  .  Type: nil
   109  .  .  .  .  .  .  }
   110  .  .  .  .  .  }
   111  .  .  .  .  }
   112  .  .  .  .  Type: *ast.Ident {
   113  .  .  .  .  .  NamePos: -
   114  .  .  .  .  .  Name: "int"
   115  .  .  .  .  .  Obj: nil
   116  .  .  .  .  }
   117  .  .  .  .  Tag: nil
   118  .  .  .  .  Comment: *ast.CommentGroup {
   119  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   120  .  .  .  .  .  .  0: *ast.Comment {
   121  .  .  .  .  .  .  .  Slash: -
   122  .  .  .  .  .  .  .  Text: "// Add other common window options here as needed"
   123  .  .  .  .  .  .  }
   124  .  .  .  .  .  }
   125  .  .  .  .  }
   126  .  .  .  }
   127  .  .  }
   128  .  .  Closing: -
   129  .  }
   130  .  Incomplete: false
   131  }

```
WindowConfig represents the configuration for a window.





### `type WindowOption`
```go
type WindowOption      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 1) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Apply"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "Apply"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    26  .  .  .  .  .  .  .  0: *ast.Field {
    27  .  .  .  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  .  .  .  Names: nil
    29  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
    30  .  .  .  .  .  .  .  .  .  Star: -
    31  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    32  .  .  .  .  .  .  .  .  .  .  NamePos: -
    33  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
    34  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    35  .  .  .  .  .  .  .  .  .  .  .  Kind: type
    36  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
    37  .  .  .  .  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    38  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    39  .  .  .  .  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    40  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    41  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "WindowConfig"
    42  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 34)
    43  .  .  .  .  .  .  .  .  .  .  .  .  }
    44  .  .  .  .  .  .  .  .  .  .  .  .  TypeParams: nil
    45  .  .  .  .  .  .  .  .  .  .  .  .  Assign: -
    46  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.StructType {
    47  .  .  .  .  .  .  .  .  .  .  .  .  .  Struct: -
    48  .  .  .  .  .  .  .  .  .  .  .  .  .  Fields: *ast.FieldList {
    49  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Opening: -
    50  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 5) {
    51  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
    52  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    53  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    54  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    55  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    56  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
    57  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    58  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    59  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
    60  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 51)
    61  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    62  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    63  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    64  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    65  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    66  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    67  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    68  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
    69  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    70  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    71  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
    72  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
    73  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    74  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  1: *ast.Field {
    75  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    76  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    77  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    78  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    79  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Title"
    80  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    81  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    82  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Title"
    83  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 74)
    84  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    85  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    86  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    87  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    88  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    89  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    90  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    91  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
    92  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    93  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    94  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
    95  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
    96  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
    97  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  2: *ast.Field {
    98  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    99  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   100  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   101  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   102  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "URL"
   103  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   104  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   105  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "URL"
   106  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 97)
   107  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   108  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   109  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   110  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   111  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   112  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   113  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   114  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   115  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   116  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   117  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   118  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   119  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   120  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  3: *ast.Field {
   121  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   122  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   123  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   124  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   125  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Width"
   126  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   127  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   128  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Width"
   129  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 120)
   130  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   131  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   132  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   133  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   134  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   135  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   136  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   137  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "int"
   138  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   139  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   140  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   141  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   142  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   143  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  4: *ast.Field {
   144  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   145  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   146  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   147  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   148  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Height"
   149  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   150  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   151  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Height"
   152  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 143)
   153  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   154  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   155  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   156  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   157  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   158  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   159  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   160  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "int"
   161  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   162  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   163  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   164  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Comment: *ast.CommentGroup {
   165  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   166  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Comment {
   167  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Slash: -
   168  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Text: "// Add other common window options here as needed"
   169  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   170  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   171  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   172  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   173  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   174  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Closing: -
   175  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   176  .  .  .  .  .  .  .  .  .  .  .  .  .  Incomplete: false
   177  .  .  .  .  .  .  .  .  .  .  .  .  }
   178  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   179  .  .  .  .  .  .  .  .  .  .  .  }
   180  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   181  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   182  .  .  .  .  .  .  .  .  .  .  }
   183  .  .  .  .  .  .  .  .  .  }
   184  .  .  .  .  .  .  .  .  }
   185  .  .  .  .  .  .  .  .  Tag: nil
   186  .  .  .  .  .  .  .  .  Comment: nil
   187  .  .  .  .  .  .  .  }
   188  .  .  .  .  .  .  }
   189  .  .  .  .  .  .  Closing: -
   190  .  .  .  .  .  }
   191  .  .  .  .  .  Results: nil
   192  .  .  .  .  }
   193  .  .  .  .  Tag: nil
   194  .  .  .  .  Comment: nil
   195  .  .  .  }
   196  .  .  }
   197  .  .  Closing: -
   198  .  }
   199  .  Incomplete: false
   200  }

```
WindowOption configures window creation.





### `type Workspace`
```go
type Workspace      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 4) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "CreateWorkspace"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: func
    13  .  .  .  .  .  .  .  Name: "CreateWorkspace"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.FuncType {
    21  .  .  .  .  .  Func: -
    22  .  .  .  .  .  TypeParams: nil
    23  .  .  .  .  .  Params: *ast.FieldList {
    24  .  .  .  .  .  .  Opening: -
    25  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    26  .  .  .  .  .  .  .  0: *ast.Field {
    27  .  .  .  .  .  .  .  .  Doc: nil
    28  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
    29  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    30  .  .  .  .  .  .  .  .  .  .  NamePos: -
    31  .  .  .  .  .  .  .  .  .  .  Name: "identifier"
    32  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    33  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    34  .  .  .  .  .  .  .  .  .  .  .  Name: "identifier"
    35  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 26)
    36  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    37  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    38  .  .  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  .  }
    40  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
    41  .  .  .  .  .  .  .  .  .  .  NamePos: -
    42  .  .  .  .  .  .  .  .  .  .  Name: "password"
    43  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    44  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    45  .  .  .  .  .  .  .  .  .  .  .  Name: "password"
    46  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 26)
    47  .  .  .  .  .  .  .  .  .  .  .  Data: nil
    48  .  .  .  .  .  .  .  .  .  .  .  Type: nil
    49  .  .  .  .  .  .  .  .  .  .  }
    50  .  .  .  .  .  .  .  .  .  }
    51  .  .  .  .  .  .  .  .  }
    52  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    53  .  .  .  .  .  .  .  .  .  NamePos: -
    54  .  .  .  .  .  .  .  .  .  Name: "string"
    55  .  .  .  .  .  .  .  .  .  Obj: nil
    56  .  .  .  .  .  .  .  .  }
    57  .  .  .  .  .  .  .  .  Tag: nil
    58  .  .  .  .  .  .  .  .  Comment: nil
    59  .  .  .  .  .  .  .  }
    60  .  .  .  .  .  .  }
    61  .  .  .  .  .  .  Closing: -
    62  .  .  .  .  .  }
    63  .  .  .  .  .  Results: *ast.FieldList {
    64  .  .  .  .  .  .  Opening: -
    65  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
    66  .  .  .  .  .  .  .  0: *ast.Field {
    67  .  .  .  .  .  .  .  .  Doc: nil
    68  .  .  .  .  .  .  .  .  Names: nil
    69  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    70  .  .  .  .  .  .  .  .  .  NamePos: -
    71  .  .  .  .  .  .  .  .  .  Name: "string"
    72  .  .  .  .  .  .  .  .  .  Obj: nil
    73  .  .  .  .  .  .  .  .  }
    74  .  .  .  .  .  .  .  .  Tag: nil
    75  .  .  .  .  .  .  .  .  Comment: nil
    76  .  .  .  .  .  .  .  }
    77  .  .  .  .  .  .  .  1: *ast.Field {
    78  .  .  .  .  .  .  .  .  Doc: nil
    79  .  .  .  .  .  .  .  .  Names: nil
    80  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    81  .  .  .  .  .  .  .  .  .  NamePos: -
    82  .  .  .  .  .  .  .  .  .  Name: "error"
    83  .  .  .  .  .  .  .  .  .  Obj: nil
    84  .  .  .  .  .  .  .  .  }
    85  .  .  .  .  .  .  .  .  Tag: nil
    86  .  .  .  .  .  .  .  .  Comment: nil
    87  .  .  .  .  .  .  .  }
    88  .  .  .  .  .  .  }
    89  .  .  .  .  .  .  Closing: -
    90  .  .  .  .  .  }
    91  .  .  .  .  }
    92  .  .  .  .  Tag: nil
    93  .  .  .  .  Comment: nil
    94  .  .  .  }
    95  .  .  .  1: *ast.Field {
    96  .  .  .  .  Doc: nil
    97  .  .  .  .  Names: []*ast.Ident (len = 1) {
    98  .  .  .  .  .  0: *ast.Ident {
    99  .  .  .  .  .  .  NamePos: -
   100  .  .  .  .  .  .  Name: "SwitchWorkspace"
   101  .  .  .  .  .  .  Obj: *ast.Object {
   102  .  .  .  .  .  .  .  Kind: func
   103  .  .  .  .  .  .  .  Name: "SwitchWorkspace"
   104  .  .  .  .  .  .  .  Decl: *(obj @ 95)
   105  .  .  .  .  .  .  .  Data: nil
   106  .  .  .  .  .  .  .  Type: nil
   107  .  .  .  .  .  .  }
   108  .  .  .  .  .  }
   109  .  .  .  .  }
   110  .  .  .  .  Type: *ast.FuncType {
   111  .  .  .  .  .  Func: -
   112  .  .  .  .  .  TypeParams: nil
   113  .  .  .  .  .  Params: *ast.FieldList {
   114  .  .  .  .  .  .  Opening: -
   115  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   116  .  .  .  .  .  .  .  0: *ast.Field {
   117  .  .  .  .  .  .  .  .  Doc: nil
   118  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   119  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   120  .  .  .  .  .  .  .  .  .  .  NamePos: -
   121  .  .  .  .  .  .  .  .  .  .  Name: "name"
   122  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   123  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   124  .  .  .  .  .  .  .  .  .  .  .  Name: "name"
   125  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 116)
   126  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   127  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   128  .  .  .  .  .  .  .  .  .  .  }
   129  .  .  .  .  .  .  .  .  .  }
   130  .  .  .  .  .  .  .  .  }
   131  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   132  .  .  .  .  .  .  .  .  .  NamePos: -
   133  .  .  .  .  .  .  .  .  .  Name: "string"
   134  .  .  .  .  .  .  .  .  .  Obj: nil
   135  .  .  .  .  .  .  .  .  }
   136  .  .  .  .  .  .  .  .  Tag: nil
   137  .  .  .  .  .  .  .  .  Comment: nil
   138  .  .  .  .  .  .  .  }
   139  .  .  .  .  .  .  }
   140  .  .  .  .  .  .  Closing: -
   141  .  .  .  .  .  }
   142  .  .  .  .  .  Results: *ast.FieldList {
   143  .  .  .  .  .  .  Opening: -
   144  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   145  .  .  .  .  .  .  .  0: *ast.Field {
   146  .  .  .  .  .  .  .  .  Doc: nil
   147  .  .  .  .  .  .  .  .  Names: nil
   148  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   149  .  .  .  .  .  .  .  .  .  NamePos: -
   150  .  .  .  .  .  .  .  .  .  Name: "error"
   151  .  .  .  .  .  .  .  .  .  Obj: nil
   152  .  .  .  .  .  .  .  .  }
   153  .  .  .  .  .  .  .  .  Tag: nil
   154  .  .  .  .  .  .  .  .  Comment: nil
   155  .  .  .  .  .  .  .  }
   156  .  .  .  .  .  .  }
   157  .  .  .  .  .  .  Closing: -
   158  .  .  .  .  .  }
   159  .  .  .  .  }
   160  .  .  .  .  Tag: nil
   161  .  .  .  .  Comment: nil
   162  .  .  .  }
   163  .  .  .  2: *ast.Field {
   164  .  .  .  .  Doc: nil
   165  .  .  .  .  Names: []*ast.Ident (len = 1) {
   166  .  .  .  .  .  0: *ast.Ident {
   167  .  .  .  .  .  .  NamePos: -
   168  .  .  .  .  .  .  Name: "WorkspaceFileGet"
   169  .  .  .  .  .  .  Obj: *ast.Object {
   170  .  .  .  .  .  .  .  Kind: func
   171  .  .  .  .  .  .  .  Name: "WorkspaceFileGet"
   172  .  .  .  .  .  .  .  Decl: *(obj @ 163)
   173  .  .  .  .  .  .  .  Data: nil
   174  .  .  .  .  .  .  .  Type: nil
   175  .  .  .  .  .  .  }
   176  .  .  .  .  .  }
   177  .  .  .  .  }
   178  .  .  .  .  Type: *ast.FuncType {
   179  .  .  .  .  .  Func: -
   180  .  .  .  .  .  TypeParams: nil
   181  .  .  .  .  .  Params: *ast.FieldList {
   182  .  .  .  .  .  .  Opening: -
   183  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   184  .  .  .  .  .  .  .  0: *ast.Field {
   185  .  .  .  .  .  .  .  .  Doc: nil
   186  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   187  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   188  .  .  .  .  .  .  .  .  .  .  NamePos: -
   189  .  .  .  .  .  .  .  .  .  .  Name: "filename"
   190  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   191  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   192  .  .  .  .  .  .  .  .  .  .  .  Name: "filename"
   193  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 184)
   194  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   195  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   196  .  .  .  .  .  .  .  .  .  .  }
   197  .  .  .  .  .  .  .  .  .  }
   198  .  .  .  .  .  .  .  .  }
   199  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   200  .  .  .  .  .  .  .  .  .  NamePos: -
   201  .  .  .  .  .  .  .  .  .  Name: "string"
   202  .  .  .  .  .  .  .  .  .  Obj: nil
   203  .  .  .  .  .  .  .  .  }
   204  .  .  .  .  .  .  .  .  Tag: nil
   205  .  .  .  .  .  .  .  .  Comment: nil
   206  .  .  .  .  .  .  .  }
   207  .  .  .  .  .  .  }
   208  .  .  .  .  .  .  Closing: -
   209  .  .  .  .  .  }
   210  .  .  .  .  .  Results: *ast.FieldList {
   211  .  .  .  .  .  .  Opening: -
   212  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   213  .  .  .  .  .  .  .  0: *ast.Field {
   214  .  .  .  .  .  .  .  .  Doc: nil
   215  .  .  .  .  .  .  .  .  Names: nil
   216  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   217  .  .  .  .  .  .  .  .  .  NamePos: -
   218  .  .  .  .  .  .  .  .  .  Name: "string"
   219  .  .  .  .  .  .  .  .  .  Obj: nil
   220  .  .  .  .  .  .  .  .  }
   221  .  .  .  .  .  .  .  .  Tag: nil
   222  .  .  .  .  .  .  .  .  Comment: nil
   223  .  .  .  .  .  .  .  }
   224  .  .  .  .  .  .  .  1: *ast.Field {
   225  .  .  .  .  .  .  .  .  Doc: nil
   226  .  .  .  .  .  .  .  .  Names: nil
   227  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   228  .  .  .  .  .  .  .  .  .  NamePos: -
   229  .  .  .  .  .  .  .  .  .  Name: "error"
   230  .  .  .  .  .  .  .  .  .  Obj: nil
   231  .  .  .  .  .  .  .  .  }
   232  .  .  .  .  .  .  .  .  Tag: nil
   233  .  .  .  .  .  .  .  .  Comment: nil
   234  .  .  .  .  .  .  .  }
   235  .  .  .  .  .  .  }
   236  .  .  .  .  .  .  Closing: -
   237  .  .  .  .  .  }
   238  .  .  .  .  }
   239  .  .  .  .  Tag: nil
   240  .  .  .  .  Comment: nil
   241  .  .  .  }
   242  .  .  .  3: *ast.Field {
   243  .  .  .  .  Doc: nil
   244  .  .  .  .  Names: []*ast.Ident (len = 1) {
   245  .  .  .  .  .  0: *ast.Ident {
   246  .  .  .  .  .  .  NamePos: -
   247  .  .  .  .  .  .  Name: "WorkspaceFileSet"
   248  .  .  .  .  .  .  Obj: *ast.Object {
   249  .  .  .  .  .  .  .  Kind: func
   250  .  .  .  .  .  .  .  Name: "WorkspaceFileSet"
   251  .  .  .  .  .  .  .  Decl: *(obj @ 242)
   252  .  .  .  .  .  .  .  Data: nil
   253  .  .  .  .  .  .  .  Type: nil
   254  .  .  .  .  .  .  }
   255  .  .  .  .  .  }
   256  .  .  .  .  }
   257  .  .  .  .  Type: *ast.FuncType {
   258  .  .  .  .  .  Func: -
   259  .  .  .  .  .  TypeParams: nil
   260  .  .  .  .  .  Params: *ast.FieldList {
   261  .  .  .  .  .  .  Opening: -
   262  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   263  .  .  .  .  .  .  .  0: *ast.Field {
   264  .  .  .  .  .  .  .  .  Doc: nil
   265  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
   266  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   267  .  .  .  .  .  .  .  .  .  .  NamePos: -
   268  .  .  .  .  .  .  .  .  .  .  Name: "filename"
   269  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   270  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   271  .  .  .  .  .  .  .  .  .  .  .  Name: "filename"
   272  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 263)
   273  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   274  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   275  .  .  .  .  .  .  .  .  .  .  }
   276  .  .  .  .  .  .  .  .  .  }
   277  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
   278  .  .  .  .  .  .  .  .  .  .  NamePos: -
   279  .  .  .  .  .  .  .  .  .  .  Name: "content"
   280  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   281  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   282  .  .  .  .  .  .  .  .  .  .  .  Name: "content"
   283  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 263)
   284  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   285  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   286  .  .  .  .  .  .  .  .  .  .  }
   287  .  .  .  .  .  .  .  .  .  }
   288  .  .  .  .  .  .  .  .  }
   289  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   290  .  .  .  .  .  .  .  .  .  NamePos: -
   291  .  .  .  .  .  .  .  .  .  Name: "string"
   292  .  .  .  .  .  .  .  .  .  Obj: nil
   293  .  .  .  .  .  .  .  .  }
   294  .  .  .  .  .  .  .  .  Tag: nil
   295  .  .  .  .  .  .  .  .  Comment: nil
   296  .  .  .  .  .  .  .  }
   297  .  .  .  .  .  .  }
   298  .  .  .  .  .  .  Closing: -
   299  .  .  .  .  .  }
   300  .  .  .  .  .  Results: *ast.FieldList {
   301  .  .  .  .  .  .  Opening: -
   302  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   303  .  .  .  .  .  .  .  0: *ast.Field {
   304  .  .  .  .  .  .  .  .  Doc: nil
   305  .  .  .  .  .  .  .  .  Names: nil
   306  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   307  .  .  .  .  .  .  .  .  .  NamePos: -
   308  .  .  .  .  .  .  .  .  .  Name: "error"
   309  .  .  .  .  .  .  .  .  .  Obj: nil
   310  .  .  .  .  .  .  .  .  }
   311  .  .  .  .  .  .  .  .  Tag: nil
   312  .  .  .  .  .  .  .  .  Comment: nil
   313  .  .  .  .  .  .  .  }
   314  .  .  .  .  .  .  }
   315  .  .  .  .  .  .  Closing: -
   316  .  .  .  .  .  }
   317  .  .  .  .  }
   318  .  .  .  .  Tag: nil
   319  .  .  .  .  Comment: nil
   320  .  .  .  }
   321  .  .  }
   322  .  .  Closing: -
   323  .  }
   324  .  Incomplete: false
   325  }

```
Workspace manages user workspaces.







## Functions

- `App()      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "application"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "App"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
`: App returns the global application instance.

- `MustServiceFor(c      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Core"
     5  .  .  Obj: nil
     6  .  }
     7  }
, name      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "T"
     3  .  Obj: *ast.Object {
     4  .  .  Kind: type
     5  .  .  Name: "T"
     6  .  .  Decl: *ast.Field {
     7  .  .  .  Doc: nil
     8  .  .  .  Names: []*ast.Ident (len = 1) {
     9  .  .  .  .  0: *ast.Ident {
    10  .  .  .  .  .  NamePos: -
    11  .  .  .  .  .  Name: "T"
    12  .  .  .  .  .  Obj: *(obj @ 3)
    13  .  .  .  .  }
    14  .  .  .  }
    15  .  .  .  Type: *ast.Ident {
    16  .  .  .  .  NamePos: -
    17  .  .  .  .  Name: "any"
    18  .  .  .  .  Obj: nil
    19  .  .  .  }
    20  .  .  .  Tag: nil
    21  .  .  .  Comment: nil
    22  .  .  }
    23  .  .  Data: nil
    24  .  .  Type: nil
    25  .  }
    26  }
`: MustServiceFor retrieves a registered service by name and asserts its type to the given interface T. It panics if the service is not found or cannot be cast to T.

- `ServiceFor(c      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Core"
     5  .  .  Obj: nil
     6  .  }
     7  }
, name      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "T"
     3  .  Obj: *ast.Object {
     4  .  .  Kind: type
     5  .  .  Name: "T"
     6  .  .  Decl: *ast.Field {
     7  .  .  .  Doc: nil
     8  .  .  .  Names: []*ast.Ident (len = 1) {
     9  .  .  .  .  0: *ast.Ident {
    10  .  .  .  .  .  NamePos: -
    11  .  .  .  .  .  Name: "T"
    12  .  .  .  .  .  Obj: *(obj @ 3)
    13  .  .  .  .  }
    14  .  .  .  }
    15  .  .  .  Type: *ast.Ident {
    16  .  .  .  .  NamePos: -
    17  .  .  .  .  Name: "any"
    18  .  .  .  .  Obj: nil
    19  .  .  .  }
    20  .  .  .  Tag: nil
    21  .  .  .  Comment: nil
    22  .  .  }
    23  .  .  Data: nil
    24  .  .  Type: nil
    25  .  }
    26  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: ServiceFor retrieves a registered service by name and asserts its type to the given interface T.

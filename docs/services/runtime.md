---
title: runtime
---
# Service: `runtime`






## Types

### `type Runtime`
```go
type Runtime      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 8) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "app"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "app"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.StarExpr {
    21  .  .  .  .  .  Star: -
    22  .  .  .  .  .  X: *ast.SelectorExpr {
    23  .  .  .  .  .  .  X: *ast.Ident {
    24  .  .  .  .  .  .  .  NamePos: -
    25  .  .  .  .  .  .  .  Name: "application"
    26  .  .  .  .  .  .  .  Obj: nil
    27  .  .  .  .  .  .  }
    28  .  .  .  .  .  .  Sel: *ast.Ident {
    29  .  .  .  .  .  .  .  NamePos: -
    30  .  .  .  .  .  .  .  Name: "App"
    31  .  .  .  .  .  .  .  Obj: nil
    32  .  .  .  .  .  .  }
    33  .  .  .  .  .  }
    34  .  .  .  .  }
    35  .  .  .  .  Tag: nil
    36  .  .  .  .  Comment: nil
    37  .  .  .  }
    38  .  .  .  1: *ast.Field {
    39  .  .  .  .  Doc: nil
    40  .  .  .  .  Names: []*ast.Ident (len = 1) {
    41  .  .  .  .  .  0: *ast.Ident {
    42  .  .  .  .  .  .  NamePos: -
    43  .  .  .  .  .  .  Name: "Core"
    44  .  .  .  .  .  .  Obj: *ast.Object {
    45  .  .  .  .  .  .  .  Kind: var
    46  .  .  .  .  .  .  .  Name: "Core"
    47  .  .  .  .  .  .  .  Decl: *(obj @ 38)
    48  .  .  .  .  .  .  .  Data: nil
    49  .  .  .  .  .  .  .  Type: nil
    50  .  .  .  .  .  .  }
    51  .  .  .  .  .  }
    52  .  .  .  .  }
    53  .  .  .  .  Type: *ast.StarExpr {
    54  .  .  .  .  .  Star: -
    55  .  .  .  .  .  X: *ast.SelectorExpr {
    56  .  .  .  .  .  .  X: *ast.Ident {
    57  .  .  .  .  .  .  .  NamePos: -
    58  .  .  .  .  .  .  .  Name: "core"
    59  .  .  .  .  .  .  .  Obj: nil
    60  .  .  .  .  .  .  }
    61  .  .  .  .  .  .  Sel: *ast.Ident {
    62  .  .  .  .  .  .  .  NamePos: -
    63  .  .  .  .  .  .  .  Name: "Core"
    64  .  .  .  .  .  .  .  Obj: nil
    65  .  .  .  .  .  .  }
    66  .  .  .  .  .  }
    67  .  .  .  .  }
    68  .  .  .  .  Tag: nil
    69  .  .  .  .  Comment: nil
    70  .  .  .  }
    71  .  .  .  2: *ast.Field {
    72  .  .  .  .  Doc: nil
    73  .  .  .  .  Names: []*ast.Ident (len = 1) {
    74  .  .  .  .  .  0: *ast.Ident {
    75  .  .  .  .  .  .  NamePos: -
    76  .  .  .  .  .  .  Name: "Config"
    77  .  .  .  .  .  .  Obj: *ast.Object {
    78  .  .  .  .  .  .  .  Kind: var
    79  .  .  .  .  .  .  .  Name: "Config"
    80  .  .  .  .  .  .  .  Decl: *(obj @ 71)
    81  .  .  .  .  .  .  .  Data: nil
    82  .  .  .  .  .  .  .  Type: nil
    83  .  .  .  .  .  .  }
    84  .  .  .  .  .  }
    85  .  .  .  .  }
    86  .  .  .  .  Type: *ast.StarExpr {
    87  .  .  .  .  .  Star: -
    88  .  .  .  .  .  X: *ast.SelectorExpr {
    89  .  .  .  .  .  .  X: *ast.Ident {
    90  .  .  .  .  .  .  .  NamePos: -
    91  .  .  .  .  .  .  .  Name: "config"
    92  .  .  .  .  .  .  .  Obj: nil
    93  .  .  .  .  .  .  }
    94  .  .  .  .  .  .  Sel: *ast.Ident {
    95  .  .  .  .  .  .  .  NamePos: -
    96  .  .  .  .  .  .  .  Name: "Service"
    97  .  .  .  .  .  .  .  Obj: nil
    98  .  .  .  .  .  .  }
    99  .  .  .  .  .  }
   100  .  .  .  .  }
   101  .  .  .  .  Tag: nil
   102  .  .  .  .  Comment: nil
   103  .  .  .  }
   104  .  .  .  3: *ast.Field {
   105  .  .  .  .  Doc: nil
   106  .  .  .  .  Names: []*ast.Ident (len = 1) {
   107  .  .  .  .  .  0: *ast.Ident {
   108  .  .  .  .  .  .  NamePos: -
   109  .  .  .  .  .  .  Name: "Display"
   110  .  .  .  .  .  .  Obj: *ast.Object {
   111  .  .  .  .  .  .  .  Kind: var
   112  .  .  .  .  .  .  .  Name: "Display"
   113  .  .  .  .  .  .  .  Decl: *(obj @ 104)
   114  .  .  .  .  .  .  .  Data: nil
   115  .  .  .  .  .  .  .  Type: nil
   116  .  .  .  .  .  .  }
   117  .  .  .  .  .  }
   118  .  .  .  .  }
   119  .  .  .  .  Type: *ast.StarExpr {
   120  .  .  .  .  .  Star: -
   121  .  .  .  .  .  X: *ast.SelectorExpr {
   122  .  .  .  .  .  .  X: *ast.Ident {
   123  .  .  .  .  .  .  .  NamePos: -
   124  .  .  .  .  .  .  .  Name: "display"
   125  .  .  .  .  .  .  .  Obj: nil
   126  .  .  .  .  .  .  }
   127  .  .  .  .  .  .  Sel: *ast.Ident {
   128  .  .  .  .  .  .  .  NamePos: -
   129  .  .  .  .  .  .  .  Name: "Service"
   130  .  .  .  .  .  .  .  Obj: nil
   131  .  .  .  .  .  .  }
   132  .  .  .  .  .  }
   133  .  .  .  .  }
   134  .  .  .  .  Tag: nil
   135  .  .  .  .  Comment: nil
   136  .  .  .  }
   137  .  .  .  4: *ast.Field {
   138  .  .  .  .  Doc: nil
   139  .  .  .  .  Names: []*ast.Ident (len = 1) {
   140  .  .  .  .  .  0: *ast.Ident {
   141  .  .  .  .  .  .  NamePos: -
   142  .  .  .  .  .  .  Name: "Help"
   143  .  .  .  .  .  .  Obj: *ast.Object {
   144  .  .  .  .  .  .  .  Kind: var
   145  .  .  .  .  .  .  .  Name: "Help"
   146  .  .  .  .  .  .  .  Decl: *(obj @ 137)
   147  .  .  .  .  .  .  .  Data: nil
   148  .  .  .  .  .  .  .  Type: nil
   149  .  .  .  .  .  .  }
   150  .  .  .  .  .  }
   151  .  .  .  .  }
   152  .  .  .  .  Type: *ast.StarExpr {
   153  .  .  .  .  .  Star: -
   154  .  .  .  .  .  X: *ast.SelectorExpr {
   155  .  .  .  .  .  .  X: *ast.Ident {
   156  .  .  .  .  .  .  .  NamePos: -
   157  .  .  .  .  .  .  .  Name: "help"
   158  .  .  .  .  .  .  .  Obj: nil
   159  .  .  .  .  .  .  }
   160  .  .  .  .  .  .  Sel: *ast.Ident {
   161  .  .  .  .  .  .  .  NamePos: -
   162  .  .  .  .  .  .  .  Name: "Service"
   163  .  .  .  .  .  .  .  Obj: nil
   164  .  .  .  .  .  .  }
   165  .  .  .  .  .  }
   166  .  .  .  .  }
   167  .  .  .  .  Tag: nil
   168  .  .  .  .  Comment: nil
   169  .  .  .  }
   170  .  .  .  5: *ast.Field {
   171  .  .  .  .  Doc: nil
   172  .  .  .  .  Names: []*ast.Ident (len = 1) {
   173  .  .  .  .  .  0: *ast.Ident {
   174  .  .  .  .  .  .  NamePos: -
   175  .  .  .  .  .  .  Name: "Crypt"
   176  .  .  .  .  .  .  Obj: *ast.Object {
   177  .  .  .  .  .  .  .  Kind: var
   178  .  .  .  .  .  .  .  Name: "Crypt"
   179  .  .  .  .  .  .  .  Decl: *(obj @ 170)
   180  .  .  .  .  .  .  .  Data: nil
   181  .  .  .  .  .  .  .  Type: nil
   182  .  .  .  .  .  .  }
   183  .  .  .  .  .  }
   184  .  .  .  .  }
   185  .  .  .  .  Type: *ast.StarExpr {
   186  .  .  .  .  .  Star: -
   187  .  .  .  .  .  X: *ast.SelectorExpr {
   188  .  .  .  .  .  .  X: *ast.Ident {
   189  .  .  .  .  .  .  .  NamePos: -
   190  .  .  .  .  .  .  .  Name: "crypt"
   191  .  .  .  .  .  .  .  Obj: nil
   192  .  .  .  .  .  .  }
   193  .  .  .  .  .  .  Sel: *ast.Ident {
   194  .  .  .  .  .  .  .  NamePos: -
   195  .  .  .  .  .  .  .  Name: "Service"
   196  .  .  .  .  .  .  .  Obj: nil
   197  .  .  .  .  .  .  }
   198  .  .  .  .  .  }
   199  .  .  .  .  }
   200  .  .  .  .  Tag: nil
   201  .  .  .  .  Comment: nil
   202  .  .  .  }
   203  .  .  .  6: *ast.Field {
   204  .  .  .  .  Doc: nil
   205  .  .  .  .  Names: []*ast.Ident (len = 1) {
   206  .  .  .  .  .  0: *ast.Ident {
   207  .  .  .  .  .  .  NamePos: -
   208  .  .  .  .  .  .  Name: "I18n"
   209  .  .  .  .  .  .  Obj: *ast.Object {
   210  .  .  .  .  .  .  .  Kind: var
   211  .  .  .  .  .  .  .  Name: "I18n"
   212  .  .  .  .  .  .  .  Decl: *(obj @ 203)
   213  .  .  .  .  .  .  .  Data: nil
   214  .  .  .  .  .  .  .  Type: nil
   215  .  .  .  .  .  .  }
   216  .  .  .  .  .  }
   217  .  .  .  .  }
   218  .  .  .  .  Type: *ast.StarExpr {
   219  .  .  .  .  .  Star: -
   220  .  .  .  .  .  X: *ast.SelectorExpr {
   221  .  .  .  .  .  .  X: *ast.Ident {
   222  .  .  .  .  .  .  .  NamePos: -
   223  .  .  .  .  .  .  .  Name: "i18n"
   224  .  .  .  .  .  .  .  Obj: nil
   225  .  .  .  .  .  .  }
   226  .  .  .  .  .  .  Sel: *ast.Ident {
   227  .  .  .  .  .  .  .  NamePos: -
   228  .  .  .  .  .  .  .  Name: "Service"
   229  .  .  .  .  .  .  .  Obj: nil
   230  .  .  .  .  .  .  }
   231  .  .  .  .  .  }
   232  .  .  .  .  }
   233  .  .  .  .  Tag: nil
   234  .  .  .  .  Comment: nil
   235  .  .  .  }
   236  .  .  .  7: *ast.Field {
   237  .  .  .  .  Doc: nil
   238  .  .  .  .  Names: []*ast.Ident (len = 1) {
   239  .  .  .  .  .  0: *ast.Ident {
   240  .  .  .  .  .  .  NamePos: -
   241  .  .  .  .  .  .  Name: "Workspace"
   242  .  .  .  .  .  .  Obj: *ast.Object {
   243  .  .  .  .  .  .  .  Kind: var
   244  .  .  .  .  .  .  .  Name: "Workspace"
   245  .  .  .  .  .  .  .  Decl: *(obj @ 236)
   246  .  .  .  .  .  .  .  Data: nil
   247  .  .  .  .  .  .  .  Type: nil
   248  .  .  .  .  .  .  }
   249  .  .  .  .  .  }
   250  .  .  .  .  }
   251  .  .  .  .  Type: *ast.StarExpr {
   252  .  .  .  .  .  Star: -
   253  .  .  .  .  .  X: *ast.SelectorExpr {
   254  .  .  .  .  .  .  X: *ast.Ident {
   255  .  .  .  .  .  .  .  NamePos: -
   256  .  .  .  .  .  .  .  Name: "workspace"
   257  .  .  .  .  .  .  .  Obj: nil
   258  .  .  .  .  .  .  }
   259  .  .  .  .  .  .  Sel: *ast.Ident {
   260  .  .  .  .  .  .  .  NamePos: -
   261  .  .  .  .  .  .  .  Name: "Service"
   262  .  .  .  .  .  .  .  Obj: nil
   263  .  .  .  .  .  .  }
   264  .  .  .  .  .  }
   265  .  .  .  .  }
   266  .  .  .  .  Tag: nil
   267  .  .  .  .  Comment: nil
   268  .  .  .  }
   269  .  .  }
   270  .  .  Closing: -
   271  .  }
   272  .  Incomplete: false
   273  }

```
Runtime is the container that holds all instantiated services.
Its fields are the concrete types, allowing Wails to bind them directly.



#### Methods

- `ServiceName()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
`: ServiceName returns the name of the service. This is used by Wails to identify the service.

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
) `: ServiceShutdown is called by Wails at application shutdown.

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
) `: ServiceStartup is called by Wails at application startup.



### `type ServiceFactory`
```go
type ServiceFactory      0  *ast.FuncType {
     1  .  Func: -
     2  .  TypeParams: nil
     3  .  Params: *ast.FieldList {
     4  .  .  Opening: -
     5  .  .  List: nil
     6  .  .  Closing: -
     7  .  }
     8  .  Results: *ast.FieldList {
     9  .  .  Opening: -
    10  .  .  List: []*ast.Field (len = 2) {
    11  .  .  .  0: *ast.Field {
    12  .  .  .  .  Doc: nil
    13  .  .  .  .  Names: nil
    14  .  .  .  .  Type: *ast.Ident {
    15  .  .  .  .  .  NamePos: -
    16  .  .  .  .  .  Name: "any"
    17  .  .  .  .  .  Obj: nil
    18  .  .  .  .  }
    19  .  .  .  .  Tag: nil
    20  .  .  .  .  Comment: nil
    21  .  .  .  }
    22  .  .  .  1: *ast.Field {
    23  .  .  .  .  Doc: nil
    24  .  .  .  .  Names: nil
    25  .  .  .  .  Type: *ast.Ident {
    26  .  .  .  .  .  NamePos: -
    27  .  .  .  .  .  Name: "error"
    28  .  .  .  .  .  Obj: nil
    29  .  .  .  .  }
    30  .  .  .  .  Tag: nil
    31  .  .  .  .  Comment: nil
    32  .  .  .  }
    33  .  .  }
    34  .  .  Closing: -
    35  .  }
    36  }

```
ServiceFactory defines a function that creates a service instance.

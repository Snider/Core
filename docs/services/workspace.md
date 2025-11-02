---
title: workspace
---
# Service: `workspace`




## Constants

```godefaultWorkspacelistFile
```




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
Options holds configuration for the workspace service.





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
    65  .  .  .  .  .  .  Name: "activeWorkspace"
    66  .  .  .  .  .  .  Obj: *ast.Object {
    67  .  .  .  .  .  .  .  Kind: var
    68  .  .  .  .  .  .  .  Name: "activeWorkspace"
    69  .  .  .  .  .  .  .  Decl: *(obj @ 60)
    70  .  .  .  .  .  .  .  Data: nil
    71  .  .  .  .  .  .  .  Type: nil
    72  .  .  .  .  .  .  }
    73  .  .  .  .  .  }
    74  .  .  .  .  }
    75  .  .  .  .  Type: *ast.StarExpr {
    76  .  .  .  .  .  Star: -
    77  .  .  .  .  .  X: *ast.Ident {
    78  .  .  .  .  .  .  NamePos: -
    79  .  .  .  .  .  .  Name: "Workspace"
    80  .  .  .  .  .  .  Obj: *ast.Object {
    81  .  .  .  .  .  .  .  Kind: type
    82  .  .  .  .  .  .  .  Name: "Workspace"
    83  .  .  .  .  .  .  .  Decl: *ast.TypeSpec {
    84  .  .  .  .  .  .  .  .  Doc: nil
    85  .  .  .  .  .  .  .  .  Name: *ast.Ident {
    86  .  .  .  .  .  .  .  .  .  NamePos: -
    87  .  .  .  .  .  .  .  .  .  Name: "Workspace"
    88  .  .  .  .  .  .  .  .  .  Obj: *(obj @ 80)
    89  .  .  .  .  .  .  .  .  }
    90  .  .  .  .  .  .  .  .  TypeParams: nil
    91  .  .  .  .  .  .  .  .  Assign: -
    92  .  .  .  .  .  .  .  .  Type: *ast.StructType {
    93  .  .  .  .  .  .  .  .  .  Struct: -
    94  .  .  .  .  .  .  .  .  .  Fields: *ast.FieldList {
    95  .  .  .  .  .  .  .  .  .  .  Opening: -
    96  .  .  .  .  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
    97  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Field {
    98  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
    99  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   100  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   101  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   102  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
   103  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   104  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   105  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Name"
   106  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 97)
   107  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   108  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   109  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   110  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   111  .  .  .  .  .  .  .  .  .  .  .  .  }
   112  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   113  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   114  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   115  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   116  .  .  .  .  .  .  .  .  .  .  .  .  }
   117  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   118  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   119  .  .  .  .  .  .  .  .  .  .  .  }
   120  .  .  .  .  .  .  .  .  .  .  .  1: *ast.Field {
   121  .  .  .  .  .  .  .  .  .  .  .  .  Doc: nil
   122  .  .  .  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   123  .  .  .  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   124  .  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   125  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Path"
   126  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   127  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   128  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "Path"
   129  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 120)
   130  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   131  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   132  .  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   133  .  .  .  .  .  .  .  .  .  .  .  .  .  }
   134  .  .  .  .  .  .  .  .  .  .  .  .  }
   135  .  .  .  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   136  .  .  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
   137  .  .  .  .  .  .  .  .  .  .  .  .  .  Name: "string"
   138  .  .  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
   139  .  .  .  .  .  .  .  .  .  .  .  .  }
   140  .  .  .  .  .  .  .  .  .  .  .  .  Tag: nil
   141  .  .  .  .  .  .  .  .  .  .  .  .  Comment: nil
   142  .  .  .  .  .  .  .  .  .  .  .  }
   143  .  .  .  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  .  .  .  Closing: -
   145  .  .  .  .  .  .  .  .  .  }
   146  .  .  .  .  .  .  .  .  .  Incomplete: false
   147  .  .  .  .  .  .  .  .  }
   148  .  .  .  .  .  .  .  .  Comment: nil
   149  .  .  .  .  .  .  .  }
   150  .  .  .  .  .  .  .  Data: nil
   151  .  .  .  .  .  .  .  Type: nil
   152  .  .  .  .  .  .  }
   153  .  .  .  .  .  }
   154  .  .  .  .  }
   155  .  .  .  .  Tag: nil
   156  .  .  .  .  Comment: nil
   157  .  .  .  }
   158  .  .  .  2: *ast.Field {
   159  .  .  .  .  Doc: nil
   160  .  .  .  .  Names: []*ast.Ident (len = 1) {
   161  .  .  .  .  .  0: *ast.Ident {
   162  .  .  .  .  .  .  NamePos: -
   163  .  .  .  .  .  .  Name: "workspaceList"
   164  .  .  .  .  .  .  Obj: *ast.Object {
   165  .  .  .  .  .  .  .  Kind: var
   166  .  .  .  .  .  .  .  Name: "workspaceList"
   167  .  .  .  .  .  .  .  Decl: *(obj @ 158)
   168  .  .  .  .  .  .  .  Data: nil
   169  .  .  .  .  .  .  .  Type: nil
   170  .  .  .  .  .  .  }
   171  .  .  .  .  .  }
   172  .  .  .  .  }
   173  .  .  .  .  Type: *ast.MapType {
   174  .  .  .  .  .  Map: -
   175  .  .  .  .  .  Key: *ast.Ident {
   176  .  .  .  .  .  .  NamePos: -
   177  .  .  .  .  .  .  Name: "string"
   178  .  .  .  .  .  .  Obj: nil
   179  .  .  .  .  .  }
   180  .  .  .  .  .  Value: *ast.Ident {
   181  .  .  .  .  .  .  NamePos: -
   182  .  .  .  .  .  .  Name: "string"
   183  .  .  .  .  .  .  Obj: nil
   184  .  .  .  .  .  }
   185  .  .  .  .  }
   186  .  .  .  .  Tag: nil
   187  .  .  .  .  Comment: *ast.CommentGroup {
   188  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   189  .  .  .  .  .  .  0: *ast.Comment {
   190  .  .  .  .  .  .  .  Slash: -
   191  .  .  .  .  .  .  .  Text: "// Maps Workspace ID to Public Key"
   192  .  .  .  .  .  .  }
   193  .  .  .  .  .  }
   194  .  .  .  .  }
   195  .  .  .  }
   196  .  .  .  3: *ast.Field {
   197  .  .  .  .  Doc: nil
   198  .  .  .  .  Names: []*ast.Ident (len = 1) {
   199  .  .  .  .  .  0: *ast.Ident {
   200  .  .  .  .  .  .  NamePos: -
   201  .  .  .  .  .  .  Name: "medium"
   202  .  .  .  .  .  .  Obj: *ast.Object {
   203  .  .  .  .  .  .  .  Kind: var
   204  .  .  .  .  .  .  .  Name: "medium"
   205  .  .  .  .  .  .  .  Decl: *(obj @ 196)
   206  .  .  .  .  .  .  .  Data: nil
   207  .  .  .  .  .  .  .  Type: nil
   208  .  .  .  .  .  .  }
   209  .  .  .  .  .  }
   210  .  .  .  .  }
   211  .  .  .  .  Type: *ast.SelectorExpr {
   212  .  .  .  .  .  X: *ast.Ident {
   213  .  .  .  .  .  .  NamePos: -
   214  .  .  .  .  .  .  Name: "io"
   215  .  .  .  .  .  .  Obj: nil
   216  .  .  .  .  .  }
   217  .  .  .  .  .  Sel: *ast.Ident {
   218  .  .  .  .  .  .  NamePos: -
   219  .  .  .  .  .  .  Name: "Medium"
   220  .  .  .  .  .  .  Obj: nil
   221  .  .  .  .  .  }
   222  .  .  .  .  }
   223  .  .  .  .  Tag: nil
   224  .  .  .  .  Comment: nil
   225  .  .  .  }
   226  .  .  }
   227  .  .  Closing: -
   228  .  }
   229  .  Incomplete: false
   230  }

```
Service manages user workspaces.



#### Methods

- `CreateWorkspace(identifier, password      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
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
`: CreateWorkspace creates a new, obfuscated workspace on the local medium.

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
`: ServiceStartup initializes the service, loading the workspace list.

- `SwitchWorkspace(name      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: SwitchWorkspace changes the active workspace.

- `WorkspaceFileGet(filename      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
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
`: WorkspaceFileGet retrieves a file from the active workspace.

- `WorkspaceFileSet(filename, content      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: WorkspaceFileSet writes a file to the active workspace.

- `getWorkspaceDir()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: getWorkspaceDir retrieves the WorkspaceDir from the config service.



### `type Workspace`
```go
type Workspace      0  *ast.StructType {
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
    33  .  .  .  .  .  .  Name: "Path"
    34  .  .  .  .  .  .  Obj: *ast.Object {
    35  .  .  .  .  .  .  .  Kind: var
    36  .  .  .  .  .  .  .  Name: "Path"
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
    51  .  .  }
    52  .  .  Closing: -
    53  .  }
    54  .  Incomplete: false
    55  }

```
Workspace represents a user's workspace.





### `type localMedium`
```go
type localMedium      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: nil
     5  .  .  Closing: -
     6  .  }
     7  .  Incomplete: false
     8  }

```
localMedium implements the Medium interface for the local disk.



#### Methods

- `EnsureDir(path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: EnsureDir creates a directory on the local disk.

- `FileGet(path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
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
`: FileGet reads a file from the local disk.

- `FileSet(path, content      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: FileSet writes a file to the local disk.

- `IsFile(path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "bool"
     3  .  Obj: nil
     4  }
`: IsFile checks if a path exists and is a file on the local disk.

- `Read(path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
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
`: Read reads a file from the local disk.

- `Write(path, content      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Write writes a file to the local disk.





## Functions

- `NewLocalMedium()      0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "io"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "Medium"
     9  .  .  Obj: nil
    10  .  }
    11  }
`: NewLocalMedium creates a new instance of the local storage medium.

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
`: Register is the constructor for dynamic dependency injection (used with core.WithService). It creates a Service instance and initializes its core.Runtime field. Dependencies are injected during ServiceStartup.

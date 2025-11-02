---
title: internal
---
# Service: `internal`




## Constants

```goappName
```


```goconfigFileName
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
Options holds configuration for the config service.





### `type Service`
```go
type Service      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 11) {
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
    57  .  .  .  .  Tag: *ast.BasicLit {
    58  .  .  .  .  .  ValuePos: -
    59  .  .  .  .  .  Kind: STRING
    60  .  .  .  .  .  Value: "`json:\"-\"`"
    61  .  .  .  .  }
    62  .  .  .  .  Comment: nil
    63  .  .  .  }
    64  .  .  .  1: *ast.Field {
    65  .  .  .  .  Doc: *ast.CommentGroup {
    66  .  .  .  .  .  List: []*ast.Comment (len = 1) {
    67  .  .  .  .  .  .  0: *ast.Comment {
    68  .  .  .  .  .  .  .  Slash: -
    69  .  .  .  .  .  .  .  Text: "// Persistent fields, saved to config.json."
    70  .  .  .  .  .  .  }
    71  .  .  .  .  .  }
    72  .  .  .  .  }
    73  .  .  .  .  Names: []*ast.Ident (len = 1) {
    74  .  .  .  .  .  0: *ast.Ident {
    75  .  .  .  .  .  .  NamePos: -
    76  .  .  .  .  .  .  Name: "ConfigPath"
    77  .  .  .  .  .  .  Obj: *ast.Object {
    78  .  .  .  .  .  .  .  Kind: var
    79  .  .  .  .  .  .  .  Name: "ConfigPath"
    80  .  .  .  .  .  .  .  Decl: *(obj @ 64)
    81  .  .  .  .  .  .  .  Data: nil
    82  .  .  .  .  .  .  .  Type: nil
    83  .  .  .  .  .  .  }
    84  .  .  .  .  .  }
    85  .  .  .  .  }
    86  .  .  .  .  Type: *ast.Ident {
    87  .  .  .  .  .  NamePos: -
    88  .  .  .  .  .  Name: "string"
    89  .  .  .  .  .  Obj: nil
    90  .  .  .  .  }
    91  .  .  .  .  Tag: *ast.BasicLit {
    92  .  .  .  .  .  ValuePos: -
    93  .  .  .  .  .  Kind: STRING
    94  .  .  .  .  .  Value: "`json:\"configPath,omitempty\"`"
    95  .  .  .  .  }
    96  .  .  .  .  Comment: nil
    97  .  .  .  }
    98  .  .  .  2: *ast.Field {
    99  .  .  .  .  Doc: nil
   100  .  .  .  .  Names: []*ast.Ident (len = 1) {
   101  .  .  .  .  .  0: *ast.Ident {
   102  .  .  .  .  .  .  NamePos: -
   103  .  .  .  .  .  .  Name: "UserHomeDir"
   104  .  .  .  .  .  .  Obj: *ast.Object {
   105  .  .  .  .  .  .  .  Kind: var
   106  .  .  .  .  .  .  .  Name: "UserHomeDir"
   107  .  .  .  .  .  .  .  Decl: *(obj @ 98)
   108  .  .  .  .  .  .  .  Data: nil
   109  .  .  .  .  .  .  .  Type: nil
   110  .  .  .  .  .  .  }
   111  .  .  .  .  .  }
   112  .  .  .  .  }
   113  .  .  .  .  Type: *ast.Ident {
   114  .  .  .  .  .  NamePos: -
   115  .  .  .  .  .  Name: "string"
   116  .  .  .  .  .  Obj: nil
   117  .  .  .  .  }
   118  .  .  .  .  Tag: *ast.BasicLit {
   119  .  .  .  .  .  ValuePos: -
   120  .  .  .  .  .  Kind: STRING
   121  .  .  .  .  .  Value: "`json:\"userHomeDir,omitempty\"`"
   122  .  .  .  .  }
   123  .  .  .  .  Comment: nil
   124  .  .  .  }
   125  .  .  .  3: *ast.Field {
   126  .  .  .  .  Doc: nil
   127  .  .  .  .  Names: []*ast.Ident (len = 1) {
   128  .  .  .  .  .  0: *ast.Ident {
   129  .  .  .  .  .  .  NamePos: -
   130  .  .  .  .  .  .  Name: "RootDir"
   131  .  .  .  .  .  .  Obj: *ast.Object {
   132  .  .  .  .  .  .  .  Kind: var
   133  .  .  .  .  .  .  .  Name: "RootDir"
   134  .  .  .  .  .  .  .  Decl: *(obj @ 125)
   135  .  .  .  .  .  .  .  Data: nil
   136  .  .  .  .  .  .  .  Type: nil
   137  .  .  .  .  .  .  }
   138  .  .  .  .  .  }
   139  .  .  .  .  }
   140  .  .  .  .  Type: *ast.Ident {
   141  .  .  .  .  .  NamePos: -
   142  .  .  .  .  .  Name: "string"
   143  .  .  .  .  .  Obj: nil
   144  .  .  .  .  }
   145  .  .  .  .  Tag: *ast.BasicLit {
   146  .  .  .  .  .  ValuePos: -
   147  .  .  .  .  .  Kind: STRING
   148  .  .  .  .  .  Value: "`json:\"rootDir,omitempty\"`"
   149  .  .  .  .  }
   150  .  .  .  .  Comment: nil
   151  .  .  .  }
   152  .  .  .  4: *ast.Field {
   153  .  .  .  .  Doc: nil
   154  .  .  .  .  Names: []*ast.Ident (len = 1) {
   155  .  .  .  .  .  0: *ast.Ident {
   156  .  .  .  .  .  .  NamePos: -
   157  .  .  .  .  .  .  Name: "CacheDir"
   158  .  .  .  .  .  .  Obj: *ast.Object {
   159  .  .  .  .  .  .  .  Kind: var
   160  .  .  .  .  .  .  .  Name: "CacheDir"
   161  .  .  .  .  .  .  .  Decl: *(obj @ 152)
   162  .  .  .  .  .  .  .  Data: nil
   163  .  .  .  .  .  .  .  Type: nil
   164  .  .  .  .  .  .  }
   165  .  .  .  .  .  }
   166  .  .  .  .  }
   167  .  .  .  .  Type: *ast.Ident {
   168  .  .  .  .  .  NamePos: -
   169  .  .  .  .  .  Name: "string"
   170  .  .  .  .  .  Obj: nil
   171  .  .  .  .  }
   172  .  .  .  .  Tag: *ast.BasicLit {
   173  .  .  .  .  .  ValuePos: -
   174  .  .  .  .  .  Kind: STRING
   175  .  .  .  .  .  Value: "`json:\"cacheDir,omitempty\"`"
   176  .  .  .  .  }
   177  .  .  .  .  Comment: nil
   178  .  .  .  }
   179  .  .  .  5: *ast.Field {
   180  .  .  .  .  Doc: nil
   181  .  .  .  .  Names: []*ast.Ident (len = 1) {
   182  .  .  .  .  .  0: *ast.Ident {
   183  .  .  .  .  .  .  NamePos: -
   184  .  .  .  .  .  .  Name: "ConfigDir"
   185  .  .  .  .  .  .  Obj: *ast.Object {
   186  .  .  .  .  .  .  .  Kind: var
   187  .  .  .  .  .  .  .  Name: "ConfigDir"
   188  .  .  .  .  .  .  .  Decl: *(obj @ 179)
   189  .  .  .  .  .  .  .  Data: nil
   190  .  .  .  .  .  .  .  Type: nil
   191  .  .  .  .  .  .  }
   192  .  .  .  .  .  }
   193  .  .  .  .  }
   194  .  .  .  .  Type: *ast.Ident {
   195  .  .  .  .  .  NamePos: -
   196  .  .  .  .  .  Name: "string"
   197  .  .  .  .  .  Obj: nil
   198  .  .  .  .  }
   199  .  .  .  .  Tag: *ast.BasicLit {
   200  .  .  .  .  .  ValuePos: -
   201  .  .  .  .  .  Kind: STRING
   202  .  .  .  .  .  Value: "`json:\"configDir,omitempty\"`"
   203  .  .  .  .  }
   204  .  .  .  .  Comment: nil
   205  .  .  .  }
   206  .  .  .  6: *ast.Field {
   207  .  .  .  .  Doc: nil
   208  .  .  .  .  Names: []*ast.Ident (len = 1) {
   209  .  .  .  .  .  0: *ast.Ident {
   210  .  .  .  .  .  .  NamePos: -
   211  .  .  .  .  .  .  Name: "DataDir"
   212  .  .  .  .  .  .  Obj: *ast.Object {
   213  .  .  .  .  .  .  .  Kind: var
   214  .  .  .  .  .  .  .  Name: "DataDir"
   215  .  .  .  .  .  .  .  Decl: *(obj @ 206)
   216  .  .  .  .  .  .  .  Data: nil
   217  .  .  .  .  .  .  .  Type: nil
   218  .  .  .  .  .  .  }
   219  .  .  .  .  .  }
   220  .  .  .  .  }
   221  .  .  .  .  Type: *ast.Ident {
   222  .  .  .  .  .  NamePos: -
   223  .  .  .  .  .  Name: "string"
   224  .  .  .  .  .  Obj: nil
   225  .  .  .  .  }
   226  .  .  .  .  Tag: *ast.BasicLit {
   227  .  .  .  .  .  ValuePos: -
   228  .  .  .  .  .  Kind: STRING
   229  .  .  .  .  .  Value: "`json:\"dataDir,omitempty\"`"
   230  .  .  .  .  }
   231  .  .  .  .  Comment: nil
   232  .  .  .  }
   233  .  .  .  7: *ast.Field {
   234  .  .  .  .  Doc: nil
   235  .  .  .  .  Names: []*ast.Ident (len = 1) {
   236  .  .  .  .  .  0: *ast.Ident {
   237  .  .  .  .  .  .  NamePos: -
   238  .  .  .  .  .  .  Name: "WorkspaceDir"
   239  .  .  .  .  .  .  Obj: *ast.Object {
   240  .  .  .  .  .  .  .  Kind: var
   241  .  .  .  .  .  .  .  Name: "WorkspaceDir"
   242  .  .  .  .  .  .  .  Decl: *(obj @ 233)
   243  .  .  .  .  .  .  .  Data: nil
   244  .  .  .  .  .  .  .  Type: nil
   245  .  .  .  .  .  .  }
   246  .  .  .  .  .  }
   247  .  .  .  .  }
   248  .  .  .  .  Type: *ast.Ident {
   249  .  .  .  .  .  NamePos: -
   250  .  .  .  .  .  Name: "string"
   251  .  .  .  .  .  Obj: nil
   252  .  .  .  .  }
   253  .  .  .  .  Tag: *ast.BasicLit {
   254  .  .  .  .  .  ValuePos: -
   255  .  .  .  .  .  Kind: STRING
   256  .  .  .  .  .  Value: "`json:\"workspaceDir,omitempty\"`"
   257  .  .  .  .  }
   258  .  .  .  .  Comment: nil
   259  .  .  .  }
   260  .  .  .  8: *ast.Field {
   261  .  .  .  .  Doc: nil
   262  .  .  .  .  Names: []*ast.Ident (len = 1) {
   263  .  .  .  .  .  0: *ast.Ident {
   264  .  .  .  .  .  .  NamePos: -
   265  .  .  .  .  .  .  Name: "DefaultRoute"
   266  .  .  .  .  .  .  Obj: *ast.Object {
   267  .  .  .  .  .  .  .  Kind: var
   268  .  .  .  .  .  .  .  Name: "DefaultRoute"
   269  .  .  .  .  .  .  .  Decl: *(obj @ 260)
   270  .  .  .  .  .  .  .  Data: nil
   271  .  .  .  .  .  .  .  Type: nil
   272  .  .  .  .  .  .  }
   273  .  .  .  .  .  }
   274  .  .  .  .  }
   275  .  .  .  .  Type: *ast.Ident {
   276  .  .  .  .  .  NamePos: -
   277  .  .  .  .  .  Name: "string"
   278  .  .  .  .  .  Obj: nil
   279  .  .  .  .  }
   280  .  .  .  .  Tag: *ast.BasicLit {
   281  .  .  .  .  .  ValuePos: -
   282  .  .  .  .  .  Kind: STRING
   283  .  .  .  .  .  Value: "`json:\"default_route\"`"
   284  .  .  .  .  }
   285  .  .  .  .  Comment: nil
   286  .  .  .  }
   287  .  .  .  9: *ast.Field {
   288  .  .  .  .  Doc: nil
   289  .  .  .  .  Names: []*ast.Ident (len = 1) {
   290  .  .  .  .  .  0: *ast.Ident {
   291  .  .  .  .  .  .  NamePos: -
   292  .  .  .  .  .  .  Name: "Features"
   293  .  .  .  .  .  .  Obj: *ast.Object {
   294  .  .  .  .  .  .  .  Kind: var
   295  .  .  .  .  .  .  .  Name: "Features"
   296  .  .  .  .  .  .  .  Decl: *(obj @ 287)
   297  .  .  .  .  .  .  .  Data: nil
   298  .  .  .  .  .  .  .  Type: nil
   299  .  .  .  .  .  .  }
   300  .  .  .  .  .  }
   301  .  .  .  .  }
   302  .  .  .  .  Type: *ast.ArrayType {
   303  .  .  .  .  .  Lbrack: -
   304  .  .  .  .  .  Len: nil
   305  .  .  .  .  .  Elt: *ast.Ident {
   306  .  .  .  .  .  .  NamePos: -
   307  .  .  .  .  .  .  Name: "string"
   308  .  .  .  .  .  .  Obj: nil
   309  .  .  .  .  .  }
   310  .  .  .  .  }
   311  .  .  .  .  Tag: *ast.BasicLit {
   312  .  .  .  .  .  ValuePos: -
   313  .  .  .  .  .  Kind: STRING
   314  .  .  .  .  .  Value: "`json:\"features\"`"
   315  .  .  .  .  }
   316  .  .  .  .  Comment: nil
   317  .  .  .  }
   318  .  .  .  10: *ast.Field {
   319  .  .  .  .  Doc: nil
   320  .  .  .  .  Names: []*ast.Ident (len = 1) {
   321  .  .  .  .  .  0: *ast.Ident {
   322  .  .  .  .  .  .  NamePos: -
   323  .  .  .  .  .  .  Name: "Language"
   324  .  .  .  .  .  .  Obj: *ast.Object {
   325  .  .  .  .  .  .  .  Kind: var
   326  .  .  .  .  .  .  .  Name: "Language"
   327  .  .  .  .  .  .  .  Decl: *(obj @ 318)
   328  .  .  .  .  .  .  .  Data: nil
   329  .  .  .  .  .  .  .  Type: nil
   330  .  .  .  .  .  .  }
   331  .  .  .  .  .  }
   332  .  .  .  .  }
   333  .  .  .  .  Type: *ast.Ident {
   334  .  .  .  .  .  NamePos: -
   335  .  .  .  .  .  Name: "string"
   336  .  .  .  .  .  Obj: nil
   337  .  .  .  .  }
   338  .  .  .  .  Tag: *ast.BasicLit {
   339  .  .  .  .  .  ValuePos: -
   340  .  .  .  .  .  Kind: STRING
   341  .  .  .  .  .  Value: "`json:\"language\"`"
   342  .  .  .  .  }
   343  .  .  .  .  Comment: nil
   344  .  .  .  }
   345  .  .  }
   346  .  .  Closing: -
   347  .  }
   348  .  Incomplete: false
   349  }

```
Service provides access to the application's configuration.
It handles loading, saving, and providing access to configuration values.



#### Methods

- `Get(key      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, out      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "any"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Get retrieves a configuration value by its key.

- `IsFeatureEnabled(feature      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "bool"
     3  .  Obj: nil
     4  }
`: IsFeatureEnabled checks if a specific feature is enabled in the config.

- `Save()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Save writes the current configuration to config.json.

- `Set(key      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, v      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "any"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Set updates a configuration value and saves the config.





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

- `TestConfigService(t      0  *ast.StarExpr {
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

- `TestIsFeatureEnabled(t      0  *ast.StarExpr {
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

- `TestSet_Bad(t      0  *ast.StarExpr {
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

- `TestSet_Good(t      0  *ast.StarExpr {
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

- `TestSet_Ugly(t      0  *ast.StarExpr {
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

- `setupTestEnv(t      0  *ast.StarExpr {
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
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
,      0  *ast.FuncType {
     1  .  Func: -
     2  .  TypeParams: nil
     3  .  Params: *ast.FieldList {
     4  .  .  Opening: -
     5  .  .  List: nil
     6  .  .  Closing: -
     7  .  }
     8  .  Results: nil
     9  }
`: setupTestEnv creates a temporary home directory for testing and ensures a clean environment.

---
title: io
---
# Service: `io`






## Types

### `type Medium`
```go
type Medium      0  *ast.InterfaceType {
     1  .  Interface: -
     2  .  Methods: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 6) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: *ast.CommentGroup {
     7  .  .  .  .  .  List: []*ast.Comment (len = 1) {
     8  .  .  .  .  .  .  0: *ast.Comment {
     9  .  .  .  .  .  .  .  Slash: -
    10  .  .  .  .  .  .  .  Text: "// Read retrieves the content of a file as a string."
    11  .  .  .  .  .  .  }
    12  .  .  .  .  .  }
    13  .  .  .  .  }
    14  .  .  .  .  Names: []*ast.Ident (len = 1) {
    15  .  .  .  .  .  0: *ast.Ident {
    16  .  .  .  .  .  .  NamePos: -
    17  .  .  .  .  .  .  Name: "Read"
    18  .  .  .  .  .  .  Obj: *ast.Object {
    19  .  .  .  .  .  .  .  Kind: func
    20  .  .  .  .  .  .  .  Name: "Read"
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
    38  .  .  .  .  .  .  .  .  .  .  Name: "path"
    39  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    40  .  .  .  .  .  .  .  .  .  .  .  Kind: var
    41  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
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
    61  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
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
    73  .  .  .  .  .  .  .  1: *ast.Field {
    74  .  .  .  .  .  .  .  .  Doc: nil
    75  .  .  .  .  .  .  .  .  Names: nil
    76  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    77  .  .  .  .  .  .  .  .  .  NamePos: -
    78  .  .  .  .  .  .  .  .  .  Name: "error"
    79  .  .  .  .  .  .  .  .  .  Obj: nil
    80  .  .  .  .  .  .  .  .  }
    81  .  .  .  .  .  .  .  .  Tag: nil
    82  .  .  .  .  .  .  .  .  Comment: nil
    83  .  .  .  .  .  .  .  }
    84  .  .  .  .  .  .  }
    85  .  .  .  .  .  .  Closing: -
    86  .  .  .  .  .  }
    87  .  .  .  .  }
    88  .  .  .  .  Tag: nil
    89  .  .  .  .  Comment: nil
    90  .  .  .  }
    91  .  .  .  1: *ast.Field {
    92  .  .  .  .  Doc: *ast.CommentGroup {
    93  .  .  .  .  .  List: []*ast.Comment (len = 1) {
    94  .  .  .  .  .  .  0: *ast.Comment {
    95  .  .  .  .  .  .  .  Slash: -
    96  .  .  .  .  .  .  .  Text: "// Write saves the given content to a file, overwriting it if it exists."
    97  .  .  .  .  .  .  }
    98  .  .  .  .  .  }
    99  .  .  .  .  }
   100  .  .  .  .  Names: []*ast.Ident (len = 1) {
   101  .  .  .  .  .  0: *ast.Ident {
   102  .  .  .  .  .  .  NamePos: -
   103  .  .  .  .  .  .  Name: "Write"
   104  .  .  .  .  .  .  Obj: *ast.Object {
   105  .  .  .  .  .  .  .  Kind: func
   106  .  .  .  .  .  .  .  Name: "Write"
   107  .  .  .  .  .  .  .  Decl: *(obj @ 91)
   108  .  .  .  .  .  .  .  Data: nil
   109  .  .  .  .  .  .  .  Type: nil
   110  .  .  .  .  .  .  }
   111  .  .  .  .  .  }
   112  .  .  .  .  }
   113  .  .  .  .  Type: *ast.FuncType {
   114  .  .  .  .  .  Func: -
   115  .  .  .  .  .  TypeParams: nil
   116  .  .  .  .  .  Params: *ast.FieldList {
   117  .  .  .  .  .  .  Opening: -
   118  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   119  .  .  .  .  .  .  .  0: *ast.Field {
   120  .  .  .  .  .  .  .  .  Doc: nil
   121  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
   122  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   123  .  .  .  .  .  .  .  .  .  .  NamePos: -
   124  .  .  .  .  .  .  .  .  .  .  Name: "path"
   125  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   126  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   127  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
   128  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 119)
   129  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   130  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   131  .  .  .  .  .  .  .  .  .  .  }
   132  .  .  .  .  .  .  .  .  .  }
   133  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
   134  .  .  .  .  .  .  .  .  .  .  NamePos: -
   135  .  .  .  .  .  .  .  .  .  .  Name: "content"
   136  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   137  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   138  .  .  .  .  .  .  .  .  .  .  .  Name: "content"
   139  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 119)
   140  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   141  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   142  .  .  .  .  .  .  .  .  .  .  }
   143  .  .  .  .  .  .  .  .  .  }
   144  .  .  .  .  .  .  .  .  }
   145  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   146  .  .  .  .  .  .  .  .  .  NamePos: -
   147  .  .  .  .  .  .  .  .  .  Name: "string"
   148  .  .  .  .  .  .  .  .  .  Obj: nil
   149  .  .  .  .  .  .  .  .  }
   150  .  .  .  .  .  .  .  .  Tag: nil
   151  .  .  .  .  .  .  .  .  Comment: nil
   152  .  .  .  .  .  .  .  }
   153  .  .  .  .  .  .  }
   154  .  .  .  .  .  .  Closing: -
   155  .  .  .  .  .  }
   156  .  .  .  .  .  Results: *ast.FieldList {
   157  .  .  .  .  .  .  Opening: -
   158  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   159  .  .  .  .  .  .  .  0: *ast.Field {
   160  .  .  .  .  .  .  .  .  Doc: nil
   161  .  .  .  .  .  .  .  .  Names: nil
   162  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   163  .  .  .  .  .  .  .  .  .  NamePos: -
   164  .  .  .  .  .  .  .  .  .  Name: "error"
   165  .  .  .  .  .  .  .  .  .  Obj: nil
   166  .  .  .  .  .  .  .  .  }
   167  .  .  .  .  .  .  .  .  Tag: nil
   168  .  .  .  .  .  .  .  .  Comment: nil
   169  .  .  .  .  .  .  .  }
   170  .  .  .  .  .  .  }
   171  .  .  .  .  .  .  Closing: -
   172  .  .  .  .  .  }
   173  .  .  .  .  }
   174  .  .  .  .  Tag: nil
   175  .  .  .  .  Comment: nil
   176  .  .  .  }
   177  .  .  .  2: *ast.Field {
   178  .  .  .  .  Doc: *ast.CommentGroup {
   179  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   180  .  .  .  .  .  .  0: *ast.Comment {
   181  .  .  .  .  .  .  .  Slash: -
   182  .  .  .  .  .  .  .  Text: "// EnsureDir makes sure a directory exists, creating it if necessary."
   183  .  .  .  .  .  .  }
   184  .  .  .  .  .  }
   185  .  .  .  .  }
   186  .  .  .  .  Names: []*ast.Ident (len = 1) {
   187  .  .  .  .  .  0: *ast.Ident {
   188  .  .  .  .  .  .  NamePos: -
   189  .  .  .  .  .  .  Name: "EnsureDir"
   190  .  .  .  .  .  .  Obj: *ast.Object {
   191  .  .  .  .  .  .  .  Kind: func
   192  .  .  .  .  .  .  .  Name: "EnsureDir"
   193  .  .  .  .  .  .  .  Decl: *(obj @ 177)
   194  .  .  .  .  .  .  .  Data: nil
   195  .  .  .  .  .  .  .  Type: nil
   196  .  .  .  .  .  .  }
   197  .  .  .  .  .  }
   198  .  .  .  .  }
   199  .  .  .  .  Type: *ast.FuncType {
   200  .  .  .  .  .  Func: -
   201  .  .  .  .  .  TypeParams: nil
   202  .  .  .  .  .  Params: *ast.FieldList {
   203  .  .  .  .  .  .  Opening: -
   204  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   205  .  .  .  .  .  .  .  0: *ast.Field {
   206  .  .  .  .  .  .  .  .  Doc: nil
   207  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   208  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   209  .  .  .  .  .  .  .  .  .  .  NamePos: -
   210  .  .  .  .  .  .  .  .  .  .  Name: "path"
   211  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   212  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   213  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
   214  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 205)
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
   228  .  .  .  .  .  .  }
   229  .  .  .  .  .  .  Closing: -
   230  .  .  .  .  .  }
   231  .  .  .  .  .  Results: *ast.FieldList {
   232  .  .  .  .  .  .  Opening: -
   233  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   234  .  .  .  .  .  .  .  0: *ast.Field {
   235  .  .  .  .  .  .  .  .  Doc: nil
   236  .  .  .  .  .  .  .  .  Names: nil
   237  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   238  .  .  .  .  .  .  .  .  .  NamePos: -
   239  .  .  .  .  .  .  .  .  .  Name: "error"
   240  .  .  .  .  .  .  .  .  .  Obj: nil
   241  .  .  .  .  .  .  .  .  }
   242  .  .  .  .  .  .  .  .  Tag: nil
   243  .  .  .  .  .  .  .  .  Comment: nil
   244  .  .  .  .  .  .  .  }
   245  .  .  .  .  .  .  }
   246  .  .  .  .  .  .  Closing: -
   247  .  .  .  .  .  }
   248  .  .  .  .  }
   249  .  .  .  .  Tag: nil
   250  .  .  .  .  Comment: nil
   251  .  .  .  }
   252  .  .  .  3: *ast.Field {
   253  .  .  .  .  Doc: *ast.CommentGroup {
   254  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   255  .  .  .  .  .  .  0: *ast.Comment {
   256  .  .  .  .  .  .  .  Slash: -
   257  .  .  .  .  .  .  .  Text: "// IsFile checks if a path exists and is a regular file."
   258  .  .  .  .  .  .  }
   259  .  .  .  .  .  }
   260  .  .  .  .  }
   261  .  .  .  .  Names: []*ast.Ident (len = 1) {
   262  .  .  .  .  .  0: *ast.Ident {
   263  .  .  .  .  .  .  NamePos: -
   264  .  .  .  .  .  .  Name: "IsFile"
   265  .  .  .  .  .  .  Obj: *ast.Object {
   266  .  .  .  .  .  .  .  Kind: func
   267  .  .  .  .  .  .  .  Name: "IsFile"
   268  .  .  .  .  .  .  .  Decl: *(obj @ 252)
   269  .  .  .  .  .  .  .  Data: nil
   270  .  .  .  .  .  .  .  Type: nil
   271  .  .  .  .  .  .  }
   272  .  .  .  .  .  }
   273  .  .  .  .  }
   274  .  .  .  .  Type: *ast.FuncType {
   275  .  .  .  .  .  Func: -
   276  .  .  .  .  .  TypeParams: nil
   277  .  .  .  .  .  Params: *ast.FieldList {
   278  .  .  .  .  .  .  Opening: -
   279  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   280  .  .  .  .  .  .  .  0: *ast.Field {
   281  .  .  .  .  .  .  .  .  Doc: nil
   282  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   283  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   284  .  .  .  .  .  .  .  .  .  .  NamePos: -
   285  .  .  .  .  .  .  .  .  .  .  Name: "path"
   286  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   287  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   288  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
   289  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 280)
   290  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   291  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   292  .  .  .  .  .  .  .  .  .  .  }
   293  .  .  .  .  .  .  .  .  .  }
   294  .  .  .  .  .  .  .  .  }
   295  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   296  .  .  .  .  .  .  .  .  .  NamePos: -
   297  .  .  .  .  .  .  .  .  .  Name: "string"
   298  .  .  .  .  .  .  .  .  .  Obj: nil
   299  .  .  .  .  .  .  .  .  }
   300  .  .  .  .  .  .  .  .  Tag: nil
   301  .  .  .  .  .  .  .  .  Comment: nil
   302  .  .  .  .  .  .  .  }
   303  .  .  .  .  .  .  }
   304  .  .  .  .  .  .  Closing: -
   305  .  .  .  .  .  }
   306  .  .  .  .  .  Results: *ast.FieldList {
   307  .  .  .  .  .  .  Opening: -
   308  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   309  .  .  .  .  .  .  .  0: *ast.Field {
   310  .  .  .  .  .  .  .  .  Doc: nil
   311  .  .  .  .  .  .  .  .  Names: nil
   312  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   313  .  .  .  .  .  .  .  .  .  NamePos: -
   314  .  .  .  .  .  .  .  .  .  Name: "bool"
   315  .  .  .  .  .  .  .  .  .  Obj: nil
   316  .  .  .  .  .  .  .  .  }
   317  .  .  .  .  .  .  .  .  Tag: nil
   318  .  .  .  .  .  .  .  .  Comment: nil
   319  .  .  .  .  .  .  .  }
   320  .  .  .  .  .  .  }
   321  .  .  .  .  .  .  Closing: -
   322  .  .  .  .  .  }
   323  .  .  .  .  }
   324  .  .  .  .  Tag: nil
   325  .  .  .  .  Comment: nil
   326  .  .  .  }
   327  .  .  .  4: *ast.Field {
   328  .  .  .  .  Doc: *ast.CommentGroup {
   329  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   330  .  .  .  .  .  .  0: *ast.Comment {
   331  .  .  .  .  .  .  .  Slash: -
   332  .  .  .  .  .  .  .  Text: "// FileGet is a convenience function that reads a file from the medium."
   333  .  .  .  .  .  .  }
   334  .  .  .  .  .  }
   335  .  .  .  .  }
   336  .  .  .  .  Names: []*ast.Ident (len = 1) {
   337  .  .  .  .  .  0: *ast.Ident {
   338  .  .  .  .  .  .  NamePos: -
   339  .  .  .  .  .  .  Name: "FileGet"
   340  .  .  .  .  .  .  Obj: *ast.Object {
   341  .  .  .  .  .  .  .  Kind: func
   342  .  .  .  .  .  .  .  Name: "FileGet"
   343  .  .  .  .  .  .  .  Decl: *(obj @ 327)
   344  .  .  .  .  .  .  .  Data: nil
   345  .  .  .  .  .  .  .  Type: nil
   346  .  .  .  .  .  .  }
   347  .  .  .  .  .  }
   348  .  .  .  .  }
   349  .  .  .  .  Type: *ast.FuncType {
   350  .  .  .  .  .  Func: -
   351  .  .  .  .  .  TypeParams: nil
   352  .  .  .  .  .  Params: *ast.FieldList {
   353  .  .  .  .  .  .  Opening: -
   354  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   355  .  .  .  .  .  .  .  0: *ast.Field {
   356  .  .  .  .  .  .  .  .  Doc: nil
   357  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
   358  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   359  .  .  .  .  .  .  .  .  .  .  NamePos: -
   360  .  .  .  .  .  .  .  .  .  .  Name: "path"
   361  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   362  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   363  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
   364  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 355)
   365  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   366  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   367  .  .  .  .  .  .  .  .  .  .  }
   368  .  .  .  .  .  .  .  .  .  }
   369  .  .  .  .  .  .  .  .  }
   370  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   371  .  .  .  .  .  .  .  .  .  NamePos: -
   372  .  .  .  .  .  .  .  .  .  Name: "string"
   373  .  .  .  .  .  .  .  .  .  Obj: nil
   374  .  .  .  .  .  .  .  .  }
   375  .  .  .  .  .  .  .  .  Tag: nil
   376  .  .  .  .  .  .  .  .  Comment: nil
   377  .  .  .  .  .  .  .  }
   378  .  .  .  .  .  .  }
   379  .  .  .  .  .  .  Closing: -
   380  .  .  .  .  .  }
   381  .  .  .  .  .  Results: *ast.FieldList {
   382  .  .  .  .  .  .  Opening: -
   383  .  .  .  .  .  .  List: []*ast.Field (len = 2) {
   384  .  .  .  .  .  .  .  0: *ast.Field {
   385  .  .  .  .  .  .  .  .  Doc: nil
   386  .  .  .  .  .  .  .  .  Names: nil
   387  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   388  .  .  .  .  .  .  .  .  .  NamePos: -
   389  .  .  .  .  .  .  .  .  .  Name: "string"
   390  .  .  .  .  .  .  .  .  .  Obj: nil
   391  .  .  .  .  .  .  .  .  }
   392  .  .  .  .  .  .  .  .  Tag: nil
   393  .  .  .  .  .  .  .  .  Comment: nil
   394  .  .  .  .  .  .  .  }
   395  .  .  .  .  .  .  .  1: *ast.Field {
   396  .  .  .  .  .  .  .  .  Doc: nil
   397  .  .  .  .  .  .  .  .  Names: nil
   398  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   399  .  .  .  .  .  .  .  .  .  NamePos: -
   400  .  .  .  .  .  .  .  .  .  Name: "error"
   401  .  .  .  .  .  .  .  .  .  Obj: nil
   402  .  .  .  .  .  .  .  .  }
   403  .  .  .  .  .  .  .  .  Tag: nil
   404  .  .  .  .  .  .  .  .  Comment: nil
   405  .  .  .  .  .  .  .  }
   406  .  .  .  .  .  .  }
   407  .  .  .  .  .  .  Closing: -
   408  .  .  .  .  .  }
   409  .  .  .  .  }
   410  .  .  .  .  Tag: nil
   411  .  .  .  .  Comment: nil
   412  .  .  .  }
   413  .  .  .  5: *ast.Field {
   414  .  .  .  .  Doc: *ast.CommentGroup {
   415  .  .  .  .  .  List: []*ast.Comment (len = 1) {
   416  .  .  .  .  .  .  0: *ast.Comment {
   417  .  .  .  .  .  .  .  Slash: -
   418  .  .  .  .  .  .  .  Text: "// FileSet is a convenience function that writes a file to the medium."
   419  .  .  .  .  .  .  }
   420  .  .  .  .  .  }
   421  .  .  .  .  }
   422  .  .  .  .  Names: []*ast.Ident (len = 1) {
   423  .  .  .  .  .  0: *ast.Ident {
   424  .  .  .  .  .  .  NamePos: -
   425  .  .  .  .  .  .  Name: "FileSet"
   426  .  .  .  .  .  .  Obj: *ast.Object {
   427  .  .  .  .  .  .  .  Kind: func
   428  .  .  .  .  .  .  .  Name: "FileSet"
   429  .  .  .  .  .  .  .  Decl: *(obj @ 413)
   430  .  .  .  .  .  .  .  Data: nil
   431  .  .  .  .  .  .  .  Type: nil
   432  .  .  .  .  .  .  }
   433  .  .  .  .  .  }
   434  .  .  .  .  }
   435  .  .  .  .  Type: *ast.FuncType {
   436  .  .  .  .  .  Func: -
   437  .  .  .  .  .  TypeParams: nil
   438  .  .  .  .  .  Params: *ast.FieldList {
   439  .  .  .  .  .  .  Opening: -
   440  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   441  .  .  .  .  .  .  .  0: *ast.Field {
   442  .  .  .  .  .  .  .  .  Doc: nil
   443  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
   444  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
   445  .  .  .  .  .  .  .  .  .  .  NamePos: -
   446  .  .  .  .  .  .  .  .  .  .  Name: "path"
   447  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   448  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   449  .  .  .  .  .  .  .  .  .  .  .  Name: "path"
   450  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 441)
   451  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   452  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   453  .  .  .  .  .  .  .  .  .  .  }
   454  .  .  .  .  .  .  .  .  .  }
   455  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
   456  .  .  .  .  .  .  .  .  .  .  NamePos: -
   457  .  .  .  .  .  .  .  .  .  .  Name: "content"
   458  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
   459  .  .  .  .  .  .  .  .  .  .  .  Kind: var
   460  .  .  .  .  .  .  .  .  .  .  .  Name: "content"
   461  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 441)
   462  .  .  .  .  .  .  .  .  .  .  .  Data: nil
   463  .  .  .  .  .  .  .  .  .  .  .  Type: nil
   464  .  .  .  .  .  .  .  .  .  .  }
   465  .  .  .  .  .  .  .  .  .  }
   466  .  .  .  .  .  .  .  .  }
   467  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   468  .  .  .  .  .  .  .  .  .  NamePos: -
   469  .  .  .  .  .  .  .  .  .  Name: "string"
   470  .  .  .  .  .  .  .  .  .  Obj: nil
   471  .  .  .  .  .  .  .  .  }
   472  .  .  .  .  .  .  .  .  Tag: nil
   473  .  .  .  .  .  .  .  .  Comment: nil
   474  .  .  .  .  .  .  .  }
   475  .  .  .  .  .  .  }
   476  .  .  .  .  .  .  Closing: -
   477  .  .  .  .  .  }
   478  .  .  .  .  .  Results: *ast.FieldList {
   479  .  .  .  .  .  .  Opening: -
   480  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
   481  .  .  .  .  .  .  .  0: *ast.Field {
   482  .  .  .  .  .  .  .  .  Doc: nil
   483  .  .  .  .  .  .  .  .  Names: nil
   484  .  .  .  .  .  .  .  .  Type: *ast.Ident {
   485  .  .  .  .  .  .  .  .  .  NamePos: -
   486  .  .  .  .  .  .  .  .  .  Name: "error"
   487  .  .  .  .  .  .  .  .  .  Obj: nil
   488  .  .  .  .  .  .  .  .  }
   489  .  .  .  .  .  .  .  .  Tag: nil
   490  .  .  .  .  .  .  .  .  Comment: nil
   491  .  .  .  .  .  .  .  }
   492  .  .  .  .  .  .  }
   493  .  .  .  .  .  .  Closing: -
   494  .  .  .  .  .  }
   495  .  .  .  .  }
   496  .  .  .  .  Tag: nil
   497  .  .  .  .  Comment: nil
   498  .  .  .  }
   499  .  .  }
   500  .  .  Closing: -
   501  .  }
   502  .  Incomplete: false
   503  }

```
Medium defines the standard interface for a storage backend.
This allows for different implementations (e.g., local disk, S3, SFTP)
to be used interchangeably.





### `type MockMedium`
```go
type MockMedium      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: []*ast.Ident (len = 1) {
     8  .  .  .  .  .  0: *ast.Ident {
     9  .  .  .  .  .  .  NamePos: -
    10  .  .  .  .  .  .  Name: "Files"
    11  .  .  .  .  .  .  Obj: *ast.Object {
    12  .  .  .  .  .  .  .  Kind: var
    13  .  .  .  .  .  .  .  Name: "Files"
    14  .  .  .  .  .  .  .  Decl: *(obj @ 5)
    15  .  .  .  .  .  .  .  Data: nil
    16  .  .  .  .  .  .  .  Type: nil
    17  .  .  .  .  .  .  }
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Type: *ast.MapType {
    21  .  .  .  .  .  Map: -
    22  .  .  .  .  .  Key: *ast.Ident {
    23  .  .  .  .  .  .  NamePos: -
    24  .  .  .  .  .  .  Name: "string"
    25  .  .  .  .  .  .  Obj: nil
    26  .  .  .  .  .  }
    27  .  .  .  .  .  Value: *ast.Ident {
    28  .  .  .  .  .  .  NamePos: -
    29  .  .  .  .  .  .  Name: "string"
    30  .  .  .  .  .  .  Obj: nil
    31  .  .  .  .  .  }
    32  .  .  .  .  }
    33  .  .  .  .  Tag: nil
    34  .  .  .  .  Comment: nil
    35  .  .  .  }
    36  .  .  .  1: *ast.Field {
    37  .  .  .  .  Doc: nil
    38  .  .  .  .  Names: []*ast.Ident (len = 1) {
    39  .  .  .  .  .  0: *ast.Ident {
    40  .  .  .  .  .  .  NamePos: -
    41  .  .  .  .  .  .  Name: "Dirs"
    42  .  .  .  .  .  .  Obj: *ast.Object {
    43  .  .  .  .  .  .  .  Kind: var
    44  .  .  .  .  .  .  .  Name: "Dirs"
    45  .  .  .  .  .  .  .  Decl: *(obj @ 36)
    46  .  .  .  .  .  .  .  Data: nil
    47  .  .  .  .  .  .  .  Type: nil
    48  .  .  .  .  .  .  }
    49  .  .  .  .  .  }
    50  .  .  .  .  }
    51  .  .  .  .  Type: *ast.MapType {
    52  .  .  .  .  .  Map: -
    53  .  .  .  .  .  Key: *ast.Ident {
    54  .  .  .  .  .  .  NamePos: -
    55  .  .  .  .  .  .  Name: "string"
    56  .  .  .  .  .  .  Obj: nil
    57  .  .  .  .  .  }
    58  .  .  .  .  .  Value: *ast.Ident {
    59  .  .  .  .  .  .  NamePos: -
    60  .  .  .  .  .  .  Name: "bool"
    61  .  .  .  .  .  .  Obj: nil
    62  .  .  .  .  .  }
    63  .  .  .  .  }
    64  .  .  .  .  Tag: nil
    65  .  .  .  .  Comment: nil
    66  .  .  .  }
    67  .  .  }
    68  .  .  Closing: -
    69  .  }
    70  .  Incomplete: false
    71  }

```
MockMedium implements the Medium interface for testing purposes.



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
`:

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
`:

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
`:

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
`:

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
`:

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
`:





## Functions

- `Copy(sourceMedium      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, sourcePath      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
, destMedium      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, destPath      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Copy copies a file from a source medium to a destination medium.

- `EnsureDir(m      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: EnsureDir ensures a directory exists on the given medium.

- `IsFile(m      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, path      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "bool"
     3  .  Obj: nil
     4  }
`: IsFile checks if a path is a file on the given medium.

- `Read(m      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, path      0  *ast.Ident {
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
`: Read retrieves the content of a file from the given medium.

- `Write(m      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "Medium"
     3  .  Obj: nil
     4  }
, path, content      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: Write saves content to a file on the given medium.

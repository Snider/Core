---
title: display
---
# Service: `display`






## Types

### `type ActionOpenWindow`
```go
type ActionOpenWindow      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 1) {
     5  .  .  .  0: *ast.Field {
     6  .  .  .  .  Doc: nil
     7  .  .  .  .  Names: nil
     8  .  .  .  .  Type: *ast.SelectorExpr {
     9  .  .  .  .  .  X: *ast.Ident {
    10  .  .  .  .  .  .  NamePos: -
    11  .  .  .  .  .  .  Name: "application"
    12  .  .  .  .  .  .  Obj: nil
    13  .  .  .  .  .  }
    14  .  .  .  .  .  Sel: *ast.Ident {
    15  .  .  .  .  .  .  NamePos: -
    16  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    17  .  .  .  .  .  .  Obj: nil
    18  .  .  .  .  .  }
    19  .  .  .  .  }
    20  .  .  .  .  Tag: nil
    21  .  .  .  .  Comment: nil
    22  .  .  .  }
    23  .  .  }
    24  .  .  Closing: -
    25  .  }
    26  .  Incomplete: false
    27  }

```
ActionOpenWindow is an IPC message used to request a new window.





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
Options holds configuration for the display service.





### `type Service`
```go
type Service      0  *ast.StructType {
     1  .  Struct: -
     2  .  Fields: *ast.FieldList {
     3  .  .  Opening: -
     4  .  .  List: []*ast.Field (len = 2) {
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
    90  .  .  }
    91  .  .  Closing: -
    92  .  }
    93  .  Incomplete: false
    94  }

```
Service manages windowing, dialogs, and other visual elements.



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
`: HandleIPCEvents processes IPC messages and performs actions such as opening windows or initializing services based on message types.

- `NewWithOptions(opts      0  *ast.Ellipsis {
     1  .  Ellipsis: -
     2  .  Elt: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "WindowOption"
     5  .  .  Obj: *ast.Object {
     6  .  .  .  Kind: type
     7  .  .  .  Name: "WindowOption"
     8  .  .  .  Decl: *ast.TypeSpec {
     9  .  .  .  .  Doc: nil
    10  .  .  .  .  Name: *ast.Ident {
    11  .  .  .  .  .  NamePos: -
    12  .  .  .  .  .  Name: "WindowOption"
    13  .  .  .  .  .  Obj: *(obj @ 5)
    14  .  .  .  .  }
    15  .  .  .  .  TypeParams: nil
    16  .  .  .  .  Assign: -
    17  .  .  .  .  Type: *ast.FuncType {
    18  .  .  .  .  .  Func: -
    19  .  .  .  .  .  TypeParams: nil
    20  .  .  .  .  .  Params: *ast.FieldList {
    21  .  .  .  .  .  .  Opening: -
    22  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    23  .  .  .  .  .  .  .  0: *ast.Field {
    24  .  .  .  .  .  .  .  .  Doc: nil
    25  .  .  .  .  .  .  .  .  Names: nil
    26  .  .  .  .  .  .  .  .  Type: *ast.StarExpr {
    27  .  .  .  .  .  .  .  .  .  Star: -
    28  .  .  .  .  .  .  .  .  .  X: *ast.SelectorExpr {
    29  .  .  .  .  .  .  .  .  .  .  X: *ast.Ident {
    30  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    31  .  .  .  .  .  .  .  .  .  .  .  Name: "application"
    32  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    33  .  .  .  .  .  .  .  .  .  .  }
    34  .  .  .  .  .  .  .  .  .  .  Sel: *ast.Ident {
    35  .  .  .  .  .  .  .  .  .  .  .  NamePos: -
    36  .  .  .  .  .  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    37  .  .  .  .  .  .  .  .  .  .  .  Obj: nil
    38  .  .  .  .  .  .  .  .  .  .  }
    39  .  .  .  .  .  .  .  .  .  }
    40  .  .  .  .  .  .  .  .  }
    41  .  .  .  .  .  .  .  .  Tag: nil
    42  .  .  .  .  .  .  .  .  Comment: nil
    43  .  .  .  .  .  .  .  }
    44  .  .  .  .  .  .  }
    45  .  .  .  .  .  .  Closing: -
    46  .  .  .  .  .  }
    47  .  .  .  .  .  Results: *ast.FieldList {
    48  .  .  .  .  .  .  Opening: -
    49  .  .  .  .  .  .  List: []*ast.Field (len = 1) {
    50  .  .  .  .  .  .  .  0: *ast.Field {
    51  .  .  .  .  .  .  .  .  Doc: nil
    52  .  .  .  .  .  .  .  .  Names: nil
    53  .  .  .  .  .  .  .  .  Type: *ast.Ident {
    54  .  .  .  .  .  .  .  .  .  NamePos: -
    55  .  .  .  .  .  .  .  .  .  Name: "error"
    56  .  .  .  .  .  .  .  .  .  Obj: nil
    57  .  .  .  .  .  .  .  .  }
    58  .  .  .  .  .  .  .  .  Tag: nil
    59  .  .  .  .  .  .  .  .  Comment: nil
    60  .  .  .  .  .  .  .  }
    61  .  .  .  .  .  .  }
    62  .  .  .  .  .  .  Closing: -
    63  .  .  .  .  .  }
    64  .  .  .  .  }
    65  .  .  .  .  Comment: nil
    66  .  .  .  }
    67  .  .  .  Data: nil
    68  .  .  .  Type: nil
    69  .  .  }
    70  .  }
    71  }
)      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "application"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "WebviewWindow"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: NewWithOptions creates a new window by applying a series of options.

- `NewWithStruct(options      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Window"
     5  .  .  Obj: *ast.Object {
     6  .  .  .  Kind: type
     7  .  .  .  Name: "Window"
     8  .  .  .  Decl: *ast.TypeSpec {
     9  .  .  .  .  Doc: nil
    10  .  .  .  .  Name: *ast.Ident {
    11  .  .  .  .  .  NamePos: -
    12  .  .  .  .  .  Name: "Window"
    13  .  .  .  .  .  Obj: *(obj @ 5)
    14  .  .  .  .  }
    15  .  .  .  .  TypeParams: nil
    16  .  .  .  .  Assign: -
    17  .  .  .  .  Type: *ast.SelectorExpr {
    18  .  .  .  .  .  X: *ast.Ident {
    19  .  .  .  .  .  .  NamePos: -
    20  .  .  .  .  .  .  Name: "application"
    21  .  .  .  .  .  .  Obj: nil
    22  .  .  .  .  .  }
    23  .  .  .  .  .  Sel: *ast.Ident {
    24  .  .  .  .  .  .  NamePos: -
    25  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    26  .  .  .  .  .  .  Obj: nil
    27  .  .  .  .  .  }
    28  .  .  .  .  }
    29  .  .  .  .  Comment: nil
    30  .  .  .  }
    31  .  .  .  Data: nil
    32  .  .  .  Type: nil
    33  .  .  }
    34  .  }
    35  }
)      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "application"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "WebviewWindow"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: NewWithStruct creates a new window using the provided options and returns its handle.

- `NewWithURL(url      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
)      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "application"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "WebviewWindow"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: NewWithURL creates a new default window pointing to the specified URL.

- `OpenWindow(opts      0  *ast.Ellipsis {
     1  .  Ellipsis: -
     2  .  Elt: *ast.SelectorExpr {
     3  .  .  X: *ast.Ident {
     4  .  .  .  NamePos: -
     5  .  .  .  Name: "core"
     6  .  .  .  Obj: nil
     7  .  .  }
     8  .  .  Sel: *ast.Ident {
     9  .  .  .  NamePos: -
    10  .  .  .  Name: "WindowOption"
    11  .  .  .  Obj: nil
    12  .  .  }
    13  .  }
    14  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: OpenWindow creates a new window with the default options.

- `SelectDirectory()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
,      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: SelectDirectory opens a directory selection dialog and returns the selected path.

- `ServiceName()      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "string"
     3  .  Obj: nil
     4  }
`:

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
`: ServiceStartup initializes the display service and sets up the main application window and system tray.

- `ShowEnvironmentDialog() `: ShowEnvironmentDialog displays a dialog containing detailed information about the application's runtime environment.

- `Window()      0  *ast.StarExpr {
     1  .  Star: -
     2  .  X: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "Window"
     5  .  .  Obj: *ast.Object {
     6  .  .  .  Kind: type
     7  .  .  .  Name: "Window"
     8  .  .  .  Decl: *ast.TypeSpec {
     9  .  .  .  .  Doc: nil
    10  .  .  .  .  Name: *ast.Ident {
    11  .  .  .  .  .  NamePos: -
    12  .  .  .  .  .  Name: "Window"
    13  .  .  .  .  .  Obj: *(obj @ 5)
    14  .  .  .  .  }
    15  .  .  .  .  TypeParams: nil
    16  .  .  .  .  Assign: -
    17  .  .  .  .  Type: *ast.SelectorExpr {
    18  .  .  .  .  .  X: *ast.Ident {
    19  .  .  .  .  .  .  NamePos: -
    20  .  .  .  .  .  .  Name: "application"
    21  .  .  .  .  .  .  Obj: nil
    22  .  .  .  .  .  }
    23  .  .  .  .  .  Sel: *ast.Ident {
    24  .  .  .  .  .  .  NamePos: -
    25  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    26  .  .  .  .  .  .  Obj: nil
    27  .  .  .  .  .  }
    28  .  .  .  .  }
    29  .  .  .  .  Comment: nil
    30  .  .  .  }
    31  .  .  .  Data: nil
    32  .  .  .  Type: nil
    33  .  .  }
    34  .  }
    35  }
`:

- `buildMenu() `: buildMenu creates and sets the main application menu.

- `handleOpenWindowAction(msg      0  *ast.MapType {
     1  .  Map: -
     2  .  Key: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "string"
     5  .  .  Obj: nil
     6  .  }
     7  .  Value: *ast.Ident {
     8  .  .  NamePos: -
     9  .  .  Name: "any"
    10  .  .  Obj: nil
    11  .  }
    12  }
)      0  *ast.Ident {
     1  .  NamePos: -
     2  .  Name: "error"
     3  .  Obj: nil
     4  }
`: handleOpenWindowAction processes a message to configure and create a new window using specified name and options.

- `monitorScreenChanges() `: monitorScreenChanges listens for theme change events and logs when screen configuration changes occur.

- `systemTray() `: setupTray configures and creates the system tray icon and menu.



### `type Window`
```go
type Window      0  *ast.SelectorExpr {
     1  .  X: *ast.Ident {
     2  .  .  NamePos: -
     3  .  .  Name: "application"
     4  .  .  Obj: nil
     5  .  }
     6  .  Sel: *ast.Ident {
     7  .  .  NamePos: -
     8  .  .  Name: "WebviewWindowOptions"
     9  .  .  Obj: nil
    10  .  }
    11  }

```





### `type WindowOption`
```go
type WindowOption      0  *ast.FuncType {
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
    11  .  .  .  .  .  X: *ast.SelectorExpr {
    12  .  .  .  .  .  .  X: *ast.Ident {
    13  .  .  .  .  .  .  .  NamePos: -
    14  .  .  .  .  .  .  .  Name: "application"
    15  .  .  .  .  .  .  .  Obj: nil
    16  .  .  .  .  .  .  }
    17  .  .  .  .  .  .  Sel: *ast.Ident {
    18  .  .  .  .  .  .  .  NamePos: -
    19  .  .  .  .  .  .  .  Name: "WebviewWindowOptions"
    20  .  .  .  .  .  .  .  Obj: nil
    21  .  .  .  .  .  .  }
    22  .  .  .  .  .  }
    23  .  .  .  .  }
    24  .  .  .  .  Tag: nil
    25  .  .  .  .  Comment: nil
    26  .  .  .  }
    27  .  .  }
    28  .  .  Closing: -
    29  .  }
    30  .  Results: *ast.FieldList {
    31  .  .  Opening: -
    32  .  .  List: []*ast.Field (len = 1) {
    33  .  .  .  0: *ast.Field {
    34  .  .  .  .  Doc: nil
    35  .  .  .  .  Names: nil
    36  .  .  .  .  Type: *ast.Ident {
    37  .  .  .  .  .  NamePos: -
    38  .  .  .  .  .  Name: "error"
    39  .  .  .  .  .  Obj: nil
    40  .  .  .  .  }
    41  .  .  .  .  Tag: nil
    42  .  .  .  .  Comment: nil
    43  .  .  .  }
    44  .  .  }
    45  .  .  Closing: -
    46  .  }
    47  }

```







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

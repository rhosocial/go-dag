# Packages in standard

The `standard` package contains five sub-packages: [`cache`](cache), [`channel`](channel), [`context`](context),
[`transit`](transit), and [`workflow`](workflow).
You can click on the links to view their respective README.md.

## Installation

To install the go-dag package and its sub-packages, use the following command:

```shell
go get github.com/rhosocial/go-dag
```

This command will fetch all the necessary packages, including `cache`, `channel`, `context`, `transit`, and `workflow`.

If you prefer to install a specific sub-package, you can run:

```shell
go get github.com/rhosocial/go-dag/workflow/standard/<sub-package>
```

Replace <sub-package> with the desired sub-package name, such as `cache`, `channel`, `context`, `transit`, or `workflow`.

After installation, you can import the packages in your Go code as needed:

```go
package main

import (
    "github.com/rhosocial/go-dag/workflow/standard/workflow"
    // Other imports as necessary
)
```
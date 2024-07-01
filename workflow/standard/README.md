# Packages in standard

The `standard` package contains seven sub-packages: [`cache`](cache), [`channel`](channel), [`context`](context),
[`logger`](logger), [`transit`](transit), [`worker`](worker) and [`workflow`](workflow). 
Compared to the [`simple`](../simple) package, this package offers more comprehensive features in "transit",
"workflow", and "logger". Additionally, it includes features such as "caching", "execution options", and a "worker pool".
The `workflow` package references the other six packages. However, the other six packages can also be used independently.
You can click on the links to view their respective README.md.

This package has been thoroughly tested with Go version 1.20, 1.21, and 1.22, and will maintain compatibility with subsequent versions.

## Installation

To install the `go-dag` package and its sub-packages, use the following command:

```shell
go get github.com/rhosocial/go-dag
```

This command will fetch all the necessary packages, including `cache`, `channel`, `context`, `logger`, `transit`,
`worker` and `workflow`.

If you prefer to install a specific sub-package, you can run:

```shell
go get github.com/rhosocial/go-dag/workflow/standard/<sub-package>
```

Replace <sub-package> with the desired sub-package name, such as `cache`, `channel`, `context`, `logger`, `transit`,
`worker` or `workflow`.

After installation, you can import the packages in your Go code as needed:

```go
package main

import (
    "github.com/rhosocial/go-dag/workflow/standard/workflow"
    // Other imports as necessary
)
```
# go-dag
A Go-based framework has been developed to oversee the execution of workflows delineated by directed acyclic graphs (DAGs).

[![Go Reference](https://pkg.go.dev/badge/github.com/rhosocial/go-dag.svg)](https://pkg.go.dev/github.com/rhosocial/go-dag)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  
![GitHub Tag](https://img.shields.io/github/v/tag/rhosocial/go-dag)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rhosocial/go-dag)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhosocial/go-dag)](https://goreportcard.com/report/github.com/rhosocial/go-dag)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/rhosocial/go-dag/go.yml?branch=r1.0)
[![Codecov coverage](https://codecov.io/gh/rhosocial/go-dag/branch/r1.0/graph/badge.svg)](https://app.codecov.io/gh/rhosocial/go-dag/tree/r1.0)

Unlike other DAG frameworks, `GO-DAG` does not prioritize concepts like edges and vertices.
Instead, it enables users to concentrate on defining transits (workers along with their inputs and outputs). 
`GO-DAG` automatically constructs workflows based on these user-defined transits and subsequently executes them.

As of now, the project is thoroughly tested in Go versions 1.20, 1.21, and 1.22,
and it continuously adapts to the latest version available.

Furthermore, the project maintains a strict adherence to zero external dependencies,
relying solely on internal code packages (excluding testing requirements).

## Introduction

`GO-DAG` comprises multiple code packages, each tailored to various levels of workflow complexity.

Presently, the released versions encompass:

- [workflow/simple](workflow/simple): Designed for describing straightforward workflows, it supports DAGs of any complexity.
It also features logging capabilities enabling customization of transits and log events. Despite its simplicity, it adequately fulfills the requirements of most scenarios.

> In the future, more complex workflow functionalities will be introduced to cater to diverse needs.

## Reference

If you seek detailed usage instructions and further information about `GO-DAG`, please visit [this](https://docs.go-dag.dev.rho.social/).

If you prefer real-time assistance from a person regarding how to use it, you can:

[![Gitter](https://img.shields.io/gitter/room/rhosocial/go-dag)](https://matrix.to/#/#go-dag.rhosocial:gitter.im)

If you're seeking guidance on leveraging this framework to streamline your complex tasks into highly efficient workflows, you can begin by donating to:

[![Donate](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/vistart/donate)

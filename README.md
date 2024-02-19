# go-dag
A framework developed in Go that manages the execution of workflows described by directed acyclic graphs.

[![Go Reference](https://pkg.go.dev/badge/github.com/rhosocial/go-dag.svg)](https://pkg.go.dev/github.com/rhosocial/go-dag)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  
![GitHub Tag](https://img.shields.io/github/v/tag/rhosocial/go-dag)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rhosocial/go-dag)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhosocial/go-dag)](https://goreportcard.com/report/github.com/rhosocial/go-dag)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/rhosocial/go-dag/go.yml?branch=r1.0)
[![Codecov coverage](https://codecov.io/gh/rhosocial/go-dag/branch/r1.0/graph/badge.svg)](https://app.codecov.io/gh/rhosocial/go-dag/tree/r1.0)

Unlike any other DAG framework, `GO-DAG` does not emphasize concepts such as edges and vertices,
but allows users to focus on defining `transit`(worker and its inputs and outputs).
`GO-DAG` will automatically builds workflows based on user-defined transits and then executes them.

Now this project is fully tested in Go 1.20, 1.21, 1.22, and always be adapted to the latest version.

This project always adheres to zero external dependencies,
that is, it only relies on internal code packages (except for testing).

## Introduction

`GO-DAG` consists of several code packages that differ in the complexity of the described workflows.

Currently, the released versions include:

- [workflow/simple](workflow/simple): used to describe simple workflows, supports DAGs of any complexity,
and has logging capabilities that allow custom transits and log events. Although simple, it can satisfy most scenarios.

> In the future, workflows with more complex functions will be introduced.

## Reference

If you want to know the specific usage of `GO-DAG` and more details, please visit [this](https://docs.go-dag.dev.rho.social/).

If you want real-person online guidance on how to use it, you can:

[![Gitter](https://img.shields.io/gitter/room/rhosocial/go-dag)](https://matrix.to/#/#go-dag.rhosocial:gitter.im)

If you want guidance on using this framework to transform your complex tasks into the most efficient workflows, you can start by donating to:

[![Donate](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/vistart/donate)

# go-dag
A framework developed in Go that manages the execution of workflows described by directed acyclic graphs.

Unlike any other DAG framework, this framework does not emphasize concepts such as edges and vertices,
but allows users to focus on defining tasks and their inputs and outputs.

[![Go Reference](https://pkg.go.dev/badge/github.com/rhosocial/go-dag.svg)](https://pkg.go.dev/github.com/rhosocial/go-dag)
![GitHub Tag](https://img.shields.io/github/v/tag/rhosocial/go-dag)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rhosocial/go-dag)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhosocial/go-dag)](https://goreportcard.com/report/github.com/rhosocial/go-dag)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/rhosocial/go-dag/go.yml?branch=r1.0)
[![Codecov coverage](https://codecov.io/gh/rhosocial/go-dag/branch/r1.0/graph/badge.svg)](https://app.codecov.io/gh/rhosocial/go-dag/tree/r1.0)

## Introduction

`GO-DAG` is expected to differentiate frameworks with different functions based on the complexity of the workflow and tasks.

Currently, the released versions include:

- [workflow/simple](workflow/simple): used to describe simple workflows, supports DAGs of any complexity,
and has logging capabilities that allow custom transits and log events. Although simple, it can satisfy most scenarios.

## Reference

If you want to know the specific usage of `GO-DAG` and more details, please visit [this](https://docs.go-dag.dev.rho.social/).

If you want real-person online guidance on how to use it, you can:

[![Gitter](https://img.shields.io/gitter/room/rhosocial/go-dag)](https://matrix.to/#/#go-dag.rhosocial:gitter.im)

If you want guidance on using this framework to transform your complex tasks into the most efficient workflows, you can start by donating to:

[![Donate](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/vistart/donate)

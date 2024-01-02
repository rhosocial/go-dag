# go-dag

Parallel tasks execution graph developed using Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/rhosocial/go-dag.svg)](https://pkg.go.dev/github.com/rhosocial/go-dag)

## Introduction

GO-DAG is a package developed in Go language that is used to compile task execution graphs and manage the execution process as efficiently as possible according to the execution graph.

The so-called "execution graph" refers to a directed acyclic graph that specifies the execution dependencies of a series of tasks.

After the execution graph is drawn up, GO-DAG will maximize the parallel execution of each task according to the dependencies and sequence of each task.

## Reference

For more details, please visit [this](https://docs.go-dag.dev.rho.social/).
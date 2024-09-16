# Function Type Generator (fnt)

Replacing dependencies with function types allow for more robust testing.

Given the following interface:

    type Client interface {
        DoSomething(ctx context.Context) error
    }

It's usage in code:

    func PerformAction(ctx context.Context, cl Client) error {
        return cl.DoSomething(ctx)
    }

Creating unit tests for ___PerformAction___ would be problematic since there is a dependency on ___Client___ which may
not be available when tests are run. An alternative pattern would be to send in a function type to
___PerformAction___ that can be substituted by tests to mock certain results

    type DoSomething func(context.Context) error

And then in ___PerformAction___:

    func PerformAction(ctx context.Context doSomething DoSomething) {
        return doSomething(ctx)
    }

This would allow ___PerformAction___ to be thoroughly tested by substituting ___DoSomething___ with a function that can
simulate different scenarios.

## Installation

    go get github.com/pkachelhoffer/fnt
    go install github.com/pkachelhoffer/fnt

## Usage

    //go:generate fnt [parameters]

## Parameters

    --interface=[Name of interface to generate, required]
    --inputPath=[Path to folder where interface can be found, optional, default to working dir]
    --outputFile=[Path to output filename, optional, default to inputfile_gen.go]
    --outputPackage=[Name of package of generated file, optional, default to input interface package name]
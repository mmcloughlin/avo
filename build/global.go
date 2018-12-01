package build

import "os"

// ctx provides a global build context.
var ctx = NewContext()

func TEXT(name string) { ctx.TEXT(name) }
func EOF()             { ctx.Generate(os.Stdout, os.Stderr) }

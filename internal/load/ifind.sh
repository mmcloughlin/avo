#!/bin/bash

inst=$1

eval $(go env)

section() {
    echo '------------------------------------------'
    echo $1
    echo '------------------------------------------'
}

section 'stdlib cmd/'
grep -iR ${inst} ${GOROOT}/src/cmd/{asm,internal/obj/x86}

section 'x/arch/x86 repo'
grep -iR ${inst} ${GOPATH}/src/golang.org/x/arch/x86/

section '*_amd64.s files in stdlib'
find ${GOROOT}/src -name '*_amd64.s' | xargs grep -i ${inst}
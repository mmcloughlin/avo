#!/bin/bash -ex

baseurl='https://github.com/dgryski/go-stadtx/raw/3c3d9b328c24a9b5ecd370654cd6e9d60a85752d'

dl() {
    filename=$1
    url="${baseurl}/${filename}"
    {
        echo "// Downloaded from '${url}'. DO NOT EDIT."
        echo
        curl -L ${url}
    } > ${filename}
}

dl stadtx.go
dl stadtx_test.go

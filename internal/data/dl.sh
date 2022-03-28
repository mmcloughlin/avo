#!/bin/bash -ex

datadir=$(dirname "${BASH_SOURCE[0]}")

dl() {
    local url=$1
    local name=${2:-$(basename ${url})}

    mkdir -p ${datadir}
    curl --output ${datadir}/${name} ${url}

    echo "* ${url}"
}

hdr() {
    echo "-----------------------------------------------------------------------------"
    echo $1
    echo "-----------------------------------------------------------------------------"
}

addlicense() {
    local repo=$1
    local file=$2

    tmp=$(mktemp)
    mv ${file} ${tmp}

    # append to LICENSE file
    {
        hdr "${repo} license"
        echo
        cat ${tmp}
        echo
    } >> ${datadir}/LICENSE

    # include in readme
    echo "### License"
    echo '```'
    cat ${tmp}
    echo '```'

    rm ${tmp}
}

{
    echo '# data'
    echo 'Underlying data files for instruction database.'
    echo

    # golang/arch x86 csv
    repo='golang/arch'
    ref='b76863e36670e165c85261bc41fabaf345376022'

    echo "## ${repo}"
    echo 'Files downloaded:'
    echo
    dl https://raw.githubusercontent.com/${repo}/${ref}/x86/x86.v0.2.csv
    dl https://raw.githubusercontent.com/${repo}/${ref}/LICENSE golang-arch-license.txt
    addlicense ${repo} ${datadir}/golang-arch-license.txt

    # golang/go aliases list.
    repo='golang/go'
    ref='go1.17.6'

    echo "## ${repo}"
    echo 'Files downloaded:'
    echo
    dl https://raw.githubusercontent.com/${repo}/${ref}/src/cmd/asm/internal/arch/arch.go arch.go.txt
    dl https://raw.githubusercontent.com/${repo}/${ref}/LICENSE golang-go-license.txt
    addlicense ${repo} ${datadir}/golang-go-license.txt

    # opcodes
    repo='Maratyszcza/Opcodes'
    ref='6e2b0cd9f1403ecaf164dea7019dd54db5aea252'

    echo "## ${repo}"
    echo 'Files downloaded:'
    echo
    dl https://raw.githubusercontent.com/${repo}/${ref}/opcodes/x86_64.xml
    dl https://raw.githubusercontent.com/${repo}/${ref}/license.rst
    addlicense ${repo} ${datadir}/license.rst

} > ${datadir}/README.md

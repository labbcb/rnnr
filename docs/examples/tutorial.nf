#!/usr/bin/env nextflow

params.str = 'Hello world!'

process splitLetters {
    container 'debian:buster-slim'

    output:
    file 'chunk_*' into letters

    """
    printf '${params.str}' | split -b 6 - chunk_
    """
}


process convertToUpper {
    container 'debian:buster-slim'

    input:
    file x from letters.flatten()

    output:
    stdout result

    """
    cat $x | tr '[a-z]' '[A-Z]'
    """
}

result.view { it.trim() }
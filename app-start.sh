#!/bin/sh

while [[ $# -gt 0 ]]; do
    case "$1" in
        -amd64)
            amd64="$2"
            shift 2
            ;;
        -arm64)
            arm64="$2"
            shift 2
            ;;
        -c)
            config="$2"
            shift 2
            ;;
    esac
done

arch=$(uname -m)

if [ "$arch" = "x86_64" ] || [ "$arch" = "amd64" ]; then
    echo "This is an AMD64 system."
    echo "Command is $amd64"
    ./$amd64
elif [ "$arch" = "aarch64" ] || [ "$arch" = "arm64" ]; then
    echo "This is an ARM64 system."
    echo "Command is $arm64"
    ./$arm64
else
    echo "Unsupported Architecture: $arch"
fi
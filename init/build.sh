scriptDir=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")

make -C ${scriptDir} docker-build 
make -C ${scriptDir} docker-push
THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $THIS_DIR/variables.sh

go run $ROOT_DIR/cmd/main.go -c $ROOT_DIR/internal/config/config-local.yaml
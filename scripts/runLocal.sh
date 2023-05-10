THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $THIS_DIR/variables.sh
export CONFIG_FILE="$ROOT_DIR/internal/config/config-local.yaml"
go run $ROOT_DIR/main.go
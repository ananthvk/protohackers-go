#!/bin/bash

REMOTE_DIRECTORY="~/protohackers/bin"

set -e

if [[ $# -lt 1 ]]; then
    echo "Invalid number of arguments" >&2
    echo "Usage: ./deploy.sh <path> <extra args passed to program>" >&2
    echo "Example: ./deploy.sh ./00_smoke_test/cmd/server" >&2
    exit 1
fi
if [[ -n "$SSH_TARGET" ]]; then
    echo "SSH_TARGET set to $SSH_TARGET"
else
    echo "Error: SSH_TARGET not set. Please provide a target in the form user@host" >&2
    exit 1
fi


function on_exit {      
  rm -rf "$TMP_DIR"
  echo "Deleted build directory $TMP_DIR"
}
trap on_exit EXIT

TMP_DIR=$(mktemp -d -t protohackers.XXXX)
PROGRAM_PATH=$1
BINARY_NAME="$(basename $PROGRAM_PATH)"
OUTPUT_PATH="${TMP_DIR}/${BINARY_NAME}"

# Build the binary
GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o $OUTPUT_PATH $1

echo "Built binary: $OUTPUT_PATH"

# Create a directoy on the remote machine
echo "Creating directoy $REMOTE_DIRECTORY on remote"
ssh $SSH_TARGET "mkdir -p $REMOTE_DIRECTORY"

# Copy the binary to the remote machine
echo "Copying binary to $SSH_TARGET:$REMOTE_DIRECTORY"
scp $OUTPUT_PATH "$SSH_TARGET:$REMOTE_DIRECTORY"

# Execute the program on the remote machine, pass extra args to the program
shift
echo "Executing program on remote machine; Extra args: $@"
ssh -t $SSH_TARGET "
    chmod +x $REMOTE_DIRECTORY/$BINARY_NAME;
    $REMOTE_DIRECTORY/$BINARY_NAME $@
"
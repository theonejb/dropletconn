_dropletconn_complete() {
  local word completions
  word="$1"
  completions="$(dropletconn completion "${word}")"
  reply=( "${(ps:\n:)completions}" )
}

compctl -f -K _dropletconn_complete dropletconn

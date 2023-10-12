#!/usr/bin/env bash
# shellcheck disable=SC2094 # Dirty hack avoid runcommand to steal stdout

title="MiSTer NFC Multigame Menu"
script_args=("$@")
#scriptdir="$(dirname "$(readlink -f "${0}")")"
#version="0.1"
#basedir="/media/fat"
[[ -f "/media/fat/Scripts/.dialogrc" ]] && export DIALOGRC="/media/fat/Scripts/.dialogrc"

_depends() {
  if ! [[ -x "$(command -v dialog)" ]]; then
    echo "dialog not installed." >"$(tty)"
    sleep 10
    _exit 1
  fi
}

main() {
  export game

  game="$(_menu -- "${script_args[@]}")"
  for ((i=0; i<${#script_args[@]}; i++)); do
    if [[ "${script_args[$i]}" == "${game}" ]]; then
      index=$((i + 1))
      #workaround, for the shape of the array
      [[ "${script_args[$index]}" == "${game}" ]] && index=$(( index + 1))
      game="${script_args[$index]}"
      break
    fi
  done
}

# Display a menu
# Usage: [msg="message"] _menu [--optional-arguments] -- [ tag item ] ...
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_menu() {
  local menu_items optional_args

  # Separate optional arguments from menu items
  while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--" ]]; then
      shift
      break
    else
      optional_args+=("$1")
      shift
    fi
  done

  # Collect menu items
  while [[ $# -gt 0 ]]; do
    menu_items+=("$1")
    shift
  done

  dialog \
    --backtitle "${title}" \
    "${optional_args[@]}" \
    --menu "${msg:-Launch Game}" \
    22 77 16 "${menu_items[@]}" 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
  return "${?}"
}

main
echo "${game}"
curl --request POST --url "http://localhost:8182/api/launch" --data "{\"path\":\"${game}\"}"

# vim: set expandtab ts=2 sw=2:

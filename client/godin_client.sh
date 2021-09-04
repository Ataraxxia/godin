#!/bin/bash

check_root() {
  if (( "${EUID:-$(id -u)}" != 0 )); then
    printf -- '%s\n' "This must be run as root" >&2
    exit 1
  fi
}

# TODO? add check_sudo()
# If run via sudo or not we just need sudo access for sudo apt-get yum update etc.
# Possibly something like 'sudo -l | grep "$0"'
# Then call like 'check_sudo || check_root'

# Uncomment to enable root check
#check_root

export LC_ALL=C

conf_file=/etc/godin/godin-client.conf
verbose=0
debug=0
update=0
quiet=0

protocol="1"
tags=""
client_hostname="${HOSTNAME}"
server_url="http://godin.example.com/reports/upload"

tmp_pkg_list="/tmp/godin_pkg_list"
tmp_host_info="/tmp/godin_host_info"
tmp_payload="/tmp/godin_payload"
tmp_repo_info="/tmp/godin_repo_info"

usage() {
cat << EOF
${0} [-v] [-d] [-u] [-s SERVER] [-c FILE] [-t tags] [-h HOSTNAME]

  -v VERBOSE output (default is silent)
  -d DEBUG output
  -u refresh repository cache using apt-get update/yum makecache, requires root privileges
  -s SERVER: web server address, e.g. https://godin.example.com/reports/upload
  -c FILE: config file location (default is /etc/godin/godin-client.conf)
  -t TAGS: comma-separated, no whitespace including list of tags, e.g. -t www,dev-vm
  -h HOSTNAME: specify the hostname of the local host
  -q QUIET: Hide any output

Command line options override config file options.
EOF
  exit 0
}

# shellcheck disable=SC1090
parseopts() {
  [[ -s "${conf_file}" ]] && source "${conf_file}"

  while getopts "vduqs:c:t:h:" opt; do
    case ${opt} in
      v) verbose=1 ;;
      d) debug=1; verbose=1 ;;
      q) debug=0; verbose=0; quiet=1 ;;
      u) update=0 ;;
      s) server_url=${OPTARG};;
      c)
        conf_file=${OPTARG}
        if [[ -s "${conf_file}" ]]; then
          source "${conf_file}"
        else
          echo "Specified configuration file does not exist!"
        fi
      ;;
      t) tags="${OPTARG}" ;;
      h) client_hostname=${OPTARG} ;;
      *) usage ;;
    esac
  done
}

check_command_exists(){
  command -v "${1:?No command specified}" >/dev/null 2>&1
}

check_requirements() {
  if ! check_command_exists curl; then
    echo 'Error: curl is not installed.' >&2
    exit 1
  fi
}

is_0() {
  (( "${1:?No parameter supplied}" == 0 ))
}

is_1() {
  (( "${1:?No parameter supplied}" == 1 ))
}

# JSON functions borrowed from
# https://raw.githubusercontent.com/rawiriblundell/jsonprint/master/lib/jsonprint.sh
json_sanitize() {
  if [[ -n "${1}" ]]; then
    _input="${1}"
  else
    read -r _input
  fi

  # Strip any literal double quotes.
  # These will be re-added if required by an output function
  _input="${_input%\"}"
  _input="${_input#\"}"

  # Strip any literal single quotes.
  # These will be re-added if required by an output function
  _input="${_input%\'}"
  _input="${_input#\'}"

  # Strip any trailing instances of ":" or "="
  _input="${_input%%:*}"
  _input="${_input%%=*}"

  # Remove any leading whitespace from 'value'
  _input="${_input#"${_input%%[![:space:]]*}"}"

  # Remove any trailing whitespace from 'key'
  _input="${_input%"${_input##*[![:space:]]}"}"

  # Return the input from whence it came
  printf -- '%s' "${_input}"
  unset -v _input
}

# Format a string keypair
# With '-c' or '--comma', we return '"key": "value",'
# Without either arg, we return '"key": "value"'
# If the value is blank or literally 'null', we return 'null' unquoted
json_str() {
  case "${1}" in
    (-c|--comma) shift 1; _comma="," ;;
    (*)          _comma="" ;;
  esac
  # Clean and assign the _key variable
  _key="$(json_sanitise "${1:-null}")"
  case "${2}" in
    (null|'') printf -- '"%s": %s%s' "${_key}" "null" "${_comma}" ;;
    (*)       shift 1; printf -- '"%s": "%s"%s' "${_key}" "${*}" "${_comma}" ;;
  esac
  unset -v _comma _key
}

get_host_data() {
  kernel_version=$(uname -r)
  architecture=$(uname -m)
  os=""
  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    case "${ID}" in
      (debian|raspbian)       os="Debian $(cat /etc/debian_version)" ;;
      (ubuntu|fedora|*suse*)  os="${PRETTY_NAME}" ;;
      (centos)                os="$(cat /etc/centos-release)" ;;
      (rhel)                  os="$(cat /etc/redhat-release)" ;;
      (arch)                  os="${NAME}" ;;
      (*)                     os="${NAME} ${VERSION}" ;;
    esac
  else
    for release_file in /etc/*{release,version}; do 
      if [[ -f "${release_file}" ]]; then
        case "${release_file}" in
          /etc/SuSE-release)
            os=$(grep -i suse "${release_file}")
            break
          ;;
          /etc/lsb-release)
            tmp_os=$(grep DISTRIB_DESCRIPTION "${release_file}")
            os=$(echo "${tmp_os}" | sed -e 's/DISTRIB_DES="\(.*\)"/\1/')
            if [[ -z "${os}" ]]; then
              tmp_os=$(grep DISTRIB_DESC "${release_file}")
              os=$(echo "${tmp_os}" | sed -e 's/DISTRIB_DESC="\(.*\)"/\1/')
            fi
            [[ -z "${os}" ]] && continue
            break
          ;;
        esac
      fi
    done
  fi

  # Print JSON 
  truncate -s 0 "${tmp_host_info}"
  {
    echo "\"host_info\" : {"
      json_str -c kernel "${kernel_version}"
      json_str -c architecture "${architecture}"
      json_str -c os "${os}"
      json_str hostname "${client_hostname}"
    echo -n "}"
  } >> "${tmp_host_info}"

  is_0 "${quiet}" && echo "Using hostname: ${client_hostname}"
}

get_apt_packages() {
  if is_1 "${update}"; then
    is_0 "${quiet}" && echo "Running apt-get update"
    apt-get update -qq
  fi

  # Get list of installed package with "ii" status
  packages_list=$(dpkg -l | grep "^ii")
  packages_count=$(echo "${packages_list}" | wc -l)

  is_0 "${quiet}" && echo "Found ${packages_count} installed packages"

  truncate -s 0  "${tmp_pkg_list}"
  {
    echo "\"repo_type\" : \"deb"
    echo "\"package_manager\" : \"apt"

    echo "\"packages\" : ["
    i=0
    while read -r package; do
      package_name=$(echo "${package}" | awk '{print $2}')
      package_arch=$(echo "${package}" | awk '{print $4}')

      i=$(( ++i ))
      package_details=$(apt-cache policy "${package_name}")
      installed_ver=$(echo "${package_details}" | grep -i "installed" | awk '{print $2}')
      candidate_ver=$(echo "${package_details}" | grep -i "candidate" | awk '{print $2}')

      # We need to extract line that comes after match for "***"
      tmp_line_no=$(echo "${package_details}" | grep -n "\*\*\*" | awk '{print $1}' FS=":")
      repository_str=$(echo "${package_details}" | awk "NR==$tmp_line_no+1")
      repository_url=$(echo "${repository_str}" | cut -f 2- -d ' ')


      if [[ "$installed_ver" != "$candidate_ver" ]]; then
        # We print package + repo + candidate + candidate_repo
        tmp_line_no=$(echo "${package_details}" | grep -n "$candidate_ver " | awk '{print $1}' FS=":")
        candidate_repository_str=$(echo "${package_details}" | awk "NR==$tmp_line_no+1")
        candidate_repository_url=$(echo "${repository_str}" | cut -f 2- -d ' ')

        echo "{"
          json_str -c name "$package_name"
          json_str -c version "$installed_ver"
          json_str -c architecture "$package_arch"
          json_str repository "$repository_url" 
        echo "},"
        echo "{"
          json_str -c upgrade yes
          json_str -c name "$package_name"
          json_str -c version "$candidate_ver"
          json_str repository "$candidate_repository_url"
        if (( i == packages_count )); then
          echo "}"
        else
          echo "},"
        fi
      else
        # We only print package + repo
        echo "{"
          json_str -c name "$package_name"
          json_str -c version "$installed_ver"
          json_str -c architecture "$package_arch"
          json_str repository "$repository_url"
        if (( i == packages_count )); then
          echo "}"
        else
          echo "},"
        fi
      fi
    done < <(dpkg -l | grep "^ii")
    echo "]"
  } >> "${tmp_pkg_list}"
}

get_yum_packages() {
  is_1 "${update}" && yum makecache --quiet

  # while loop will execute in current shell instead of a sub-shell
  #shopt -s lastpipe 

  truncate -s 0 "${tmp_repo_info}"
  {
  echo "\"repositories\" : ["
  while read -r line; do
    field_id=$(echo "${line}" | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]' | awk -F':' '{print $1}')
    field_val=$(echo "${line}" | cut -d ':' -f 2- )
      case "${field_id}" in
        "repo-id")
          repo_alias=$(echo "${field_val}" | awk -F'/' '{print $1}' | tr -d '[:space:]')
          echo "{"
          json_str -c repository_alias "$repo_alias"
          json_Str -c repository_id "$field_val"
        ;;
        "repo-name")
         json_str repository_name "$field_val"
        ;;
        "repo-baseurl")
          baseurl=$(echo "$field_val" | awk '{print $1}')
          json_str repository_baseurl "$baseurl"
          echo "}," 
        ;;
      esac
    done < <(yum repoinfo)
  } >> "${tmp_repo_info}"
  truncate -s-2 "${tmp_repo_info}" # Removing trailing comma
  echo "]" >> "${tmp_repo_info}"

  packages_count=$(repoquery '*' --queryformat='%{name} %{evr} %{ui_from_repo}' --installed | wc -l)
  #upgrades_count=$(yum check-updates | grep -v ^$ | wc -l)

  is_0 "${quiet}" && echo "Found ${packages_count} installed packages"

  truncate -s 0 "${tmp_pkg_list}"
  {
    echo "\"repo_type\" : \"rpm"
    echo "\"package_manager\" : \"yum"
    echo "\"packages\" : ["
  } >> "${tmp_pkg_list}"

  i=0
  while read -r package; do
    package_name=$(echo "${package}" | awk '{print $1}' | awk -F'.' '{print $1}')
    package_arch=$(echo "${package}" | awk '{print $1}' | awk -F'.' '{print $2}')
    package_version=$(echo "${package}" | awk '{print $2}')
    package_aliasrepo=$(echo "${package}" | awk '{print $3}')

    if is_0 "$i" && [[ -n "$package_name" ]] && [[ "$package_name" == "Loaded" ]]; then
      i=$(( ++i ))
      continue
    fi

    {
      echo "{"
        json_str -c upgrade yes
        json_str -c name "$package_name"
        json_str -c version "$package_version"
        json_str -c architecture "$package_arch"
        json_str repository "$package_aliasrepo"
      echo "},"
    } >> "${tmp_pkg_list}"
  done < <(yum check-updates | grep -v ^$)

  i=0
  # simply listing with yum can possibly break some columns
  {
    while read -r package; do
      i=$(( ++i ))

      package_name=$(echo "${package}" | awk '{print $1}' | awk -F'.' '{print $1}')
      package_arch=$(echo "${package}" | awk '{print $1}' | awk -F'.' '{print $2}')
      package_version=$(echo "${package}" | awk '{print $2}')
      package_aliasrepo=$(echo "${package}" | awk '{print $3}')

      echo "{"
        json_str -c name "$package_name"
        json_str -c version "$package_version"
        json_str -c architecture "$package_arch"
        json_str repository "$package_aliasrepo"
      if (( i == packages_count )); then
        echo "}"
      else
        echo "},"
      fi
    done < <(repoquery '*' --queryformat='%{name} %{evr} %{ui_from_repo}' --installed)

    echo "]"
  } >> "${tmp_pkg_list}"
}

cleanup() {
  rm "${tmp_host_info}"
  rm "${tmp_pkg_list}"
  if [[ -s "${tmp_repo_info}" ]]; then
    rm "${tmp_repo_info}"
  fi
  rm "${tmp_payload}"
}

#################################
#       MAIN                    #
#################################

parseopts "$@"
check_requirements

get_host_data

if check_command_exists apt-get; then
  get_apt_packages
elif check_command_exists yum; then
  get_yum_packages
fi

truncate -s 0 "${tmp_payload}"

{
  echo "{"
  cat "${tmp_host_info}"
  echo ","

  echo -n "\"tags\" : ["
  if [[ -n "$tags" ]]; then
    for tag in ${tags//,/ }; do
      echo -n "$tag"
    done
  fi
} >> "${tmp_payload}"

truncate -s-1 "${tmp_payload}"

{
  echo "],"

  json_str protocol "${protocol}"
  if [[ -s "${tmp_repo_info}" ]]; then
    cat "${tmp_repo_info}"
    echo ","
  fi
  cat "${tmp_pkg_list}"
  echo "}"
} >> "${tmp_payload}"


curl -L -X POST -H "Content-Type: application/json" -d @"${tmp_payload}" "${server_url}"
cleanup

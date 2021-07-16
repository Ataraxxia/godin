#!/bin/bash

# TODO Check if run as root; or not we just need sudo access for sudo apt-get yum update etc.

export LC_ALL=C

CONF_FILE=/etc/godin/godin-client.conf
PROTOCOL=1
VERBOSE=false
DEBUG=false
TAGS=""
CLIENT_HOSTNAME=`echo $HOSTNAME`
UPDATE=true

TMP_PKG_LIST="/tmp/godin_pkg_list"
TMP_HOST_INFO="/tmp/godin_host_info"

usage() {
    echo "${0} [-v] [-d] [-u] [-s SERVER] [-c FILE] [-t TAGS] [-h HOSTNAME]"
    echo "-v: verbose output (default is silent)"
    echo "-d: debug output"
    echo "-u: DO NOT automaticaly perform update using apt/yum"
    echo "-s SERVER: web server address, e.g. https://godin.example.com"
    echo "-c FILE: config file location (default is /etc/patchman/godin-client.conf)"
    echo "-t TAGS: comma-separated list of tags, e.g. -t www,dev"
    echo "-h HOSTNAME: specify the hostname of the local host"
    echo
    echo "Command line options override config file options."
    exit 0
}

parseopts() {
    while getopts "vdus:c:t:h:" opt; do
        case ${opt} in
        v)
            VERBOSE=true
            ;;
        d)
            DEBUG=true
            VERBOSE=true
            ;;
	u)
	    UPDATE=false
	    ;;
        s)
            SERVER_URL=${OPTARG}
            ;;
        c)
            CONF_FILE=${OPTARG}
            ;;
        t)
            TAGS="${OPTARG}"
            ;;
        h)
            CLIENT_HOSTNAME=${OPTARG}
            ;;
        *)
            usage
            ;;
        esac
    done
}

check_command_exists(){
	if [ -x "$(command -v ${1})" ]; then
		return 0
	else
		return 1
	fi
}

check_requirements() {
	if ! check_command_exists curl; then
		echo 'Error: curl is not installed.' >&2
		exit 1
	fi
}

get_host_data() {
	kernel_version=$(uname -r)
	architecture=$(uname -m)
	os=""
	if [ -f /etc/os-release ] ; then
		. /etc/os-release
		if [ "${ID}" == "debian" ] ; then
			os="Debian $(cat /etc/debian_version)"
		elif [ "${ID}" == "raspbian" ] ; then
			os="Raspbian $(cat /etc/debian_version)"
		elif [ "${ID}" == "ubuntu" ] ; then
			os="${PRETTY_NAME}"
		elif [ "${ID}" == "centos" ] ; then
			os="$(cat /etc/centos-release)"
		elif [ "${ID}" == "rhel" ] ; then
			os="$(cat /etc/redhat-release)"
		elif [ "${ID}" == "fedora" ] ; then
			os="${PRETTY_NAME}"
		elif [ "${ID}" == "arch" ] ; then
			os="${NAME}"
		elif [[ "${ID}" =~ "suse" ]] ; then
			os="${PRETTY_NAME}"
		else
			os="${NAME} ${VERSION}"
		fi
	else
	        releases="/etc/SuSE-release /etc/lsb-release /etc/debian_version /etc/fermi-release /etc/redhat-release /etc/fedora-release /etc/centos-release"
		for r in ${releases}; do 
			if [ -f ${r} ]; then
				case "${r}" in
				/etc/SuSE-release)
					os=$(grep -i suse ${r})
					break
					;;
				/etc/lsb-release)
					tmp_os=$(grep DISTRIB_DESCRIPTION ${r})
					os=$(echo ${tmp_os} | sed -e 's/DISTRIB_DES="\(.*\)"/\1/')
					if [ -z "${os}" ]; then
						tmp_os=$(grep DISTRIB_DESC ${r})
						os=$(echo ${tmp_os} | sed -e 's/DISTRIB_DESC="\(.*\)"/\1/')
					fi
					if [ -z "${os}" ]; then
						continue
					fi
					break
					;;
				esac
			fi
		done
	fi

	# Print JSON 
	echo "" > $TMP_HOST_INFO
	echo "{" >> $TMP_HOST_INFO
	echo "	\"kernel\": \"$kernel_version\"," >> $TMP_HOST_INFO
	echo "	\"architecture\": \"$architecture\"," >> $TMP_HOST_INFO
	echo "	\"os\": \"$os\"" >> $TMP_HOST_INFO
	echo "}" >> $TMP_HOST_INFO
}

get_apt_packages() {
	if $UPDATE; then
		apt-get update
	fi

	# Get list of installed package with "ii" status
	packages_list=$(dpkg -l | grep "^ii" | awk '{print $2}')

	echo "" > $TMP_PKG_LIST

	echo "{" >> $TMP_PKG_LIST
	for package in $packages_list; do 
		package_details=$(apt-cache policy ${package})
		installed_ver=$(echo $"$package_details" | grep -i "installed" | awk '{print $2}')
		candidate_ver=$(echo $"$package_details" | grep -i "candidate" | awk '{print $2}')

		# We need to extract line that comes after match for "***"
		tmp_line_no=$(echo $"$package_details" | grep -n "\*\*\*" | awk '{print $1}' FS=":")
		repository_str=$(echo $"$package_details" | awk "NR==$tmp_line_no+1")
		repository_url=$(echo $repository_str | cut -f 2- -d ' ')


		if [ "$installed_ver" != "$candidate_ver" ]; then
			# We print package + repo + candidate + candidate_repo
			tmp_line_no=$(echo $"$package_details" | grep -n "$candidate_ver " | awk '{print $1}' FS=":")
			candidate_repository_str=$(echo $"$package_details" | awk "NR==$tmp_line_no+1")
			candidate_repository_url=$(echo $repository_str | cut -f 2- -d ' ')

			echo "{" >> $TMP_PKG_LIST
			echo "	\"name\": \"$package\"," >> $TMP_PKG_LIST
			echo "	\"version\": \"$installed_ver\"," >> $TMP_PKG_LIST
			echo "	\"repository\": \"$repository_url\"," >> $TMP_PKG_LIST
			echo "	\"upgradable\": \"yes\"," >> $TMP_PKG_LIST
			echo "	\"candidate\": {" >> $TMP_PKG_LIST
			echo "		\"version\": \"$candidate_ver\"," >> $TMP_PKG_LIST
			echo "		\"repository\": \"$candidate_repository_url\"" >> $TMP_PKG_LIST
			echo "	\"}\"" >> $TMP_PKG_LIST
			echo "}," >> $TMP_PKG_LIST
		else
			# We only print package + repo
			echo "{" >> $TMP_PKG_LIST
			echo "	\"name\": \"$package\"," >> $TMP_PKG_LIST
			echo "	\"version\": \"$installed_ver\"," >> $TMP_PKG_LIST
			echo "	\"repository\": \"$repository_url\"," >> $TMP_PKG_LIST
			echo "	\"upgradable\": \"no\"" >> $TMP_PKG_LIST
			echo "}," >> $TMP_PKG_LIST
		fi
	done
	echo "}" >> $TMP_PKG_LIST
}

get_yum_packages() {
}

cleanup() {
	rm $TMP_HOST_INFO
	rm $TMP_PKG_LIST
}


parseopts "$@"
check_requirements

get_host_data

	if check_command_exists apt-get; then
		get_apt_packages
	elif check_command_exists yum; then
		get_yum_packages
	fi

if $VERBOSE; then
	cat $TMP_HOST_INFO $TMP_PKG_LIST
fi

#cleanup


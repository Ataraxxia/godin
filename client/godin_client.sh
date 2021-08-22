#!/bin/bash

# TODO Check if run as root; or not we just need sudo access for sudo apt-get yum update etc.

export LC_ALL=C

CONF_FILE=/etc/godin/godin-client.conf
PROTOCOL=1
VERBOSE=false
DEBUG=false
TAGS=""
CLIENT_HOSTNAME=`echo $HOSTNAME`
UPDATE=false
QUIET=false
SERVER_URL="http://godin.example.com"

TMP_PKG_LIST="/tmp/godin_pkg_list"
TMP_HOST_INFO="/tmp/godin_host_info"
TMP_PAYLOAD="/tmp/godin_payload"
TMP_REPO_INFO="/tmp/godin_repo_info"

usage() {
    echo "${0} [-v] [-d] [-u] [-s SERVER] [-c FILE] [-t TAGS] [-h HOSTNAME]"
    echo "-v: verbose output (default is silent)"
    echo "-d: debug output"
    echo "-u: refresh repository cache using apt-get update/yum makecache, requires root privileges"
    echo "-s SERVER: web server address, e.g. https://godin.example.com/upload"
    echo "-c FILE: config file location (default is /etc/patchman/godin-client.conf)"
    echo "-t TAGS: comma-separated list of tags, e.g. -t www,dev"
    echo "-h HOSTNAME: specify the hostname of the local host"
	echo "-q QUIET: Hide any output"
    echo
    echo "Command line options override config file options."
    exit 0
}

parseopts() {
    while getopts "vduqs:c:t:h:" opt; do
        case ${opt} in
        v)
            VERBOSE=true
            ;;
        d)
            DEBUG=true
            VERBOSE=true
            ;;
		q)
			VERBOSE=false
			DEBUG=false
			QUIET=true
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
	truncate -s 0 $TMP_HOST_INFO
	echo "\"host_info\" : {" >> $TMP_HOST_INFO
	echo "	\"kernel\": \"$kernel_version\"," >> $TMP_HOST_INFO
	echo "	\"architecture\": \"$architecture\"," >> $TMP_HOST_INFO
	echo "	\"os\": \"$os\"," >> $TMP_HOST_INFO
	echo "	\"hostname\": \"$HOSTNAME\"" >> $TMP_HOST_INFO
	echo "}" >> $TMP_HOST_INFO
}

get_apt_packages() {
	if $UPDATE; then
		apt-get update -qq
	fi

	# Get list of installed package with "ii" status
	packages_list=$(dpkg -l | grep "^ii")
	packages_count=$(echo $"$packages_list" | wc -l)

	if [ ! $QUIET ]; then
		echo "Found $packages_count installed packages"
	fi

	truncate -s 0  $TMP_REPO_INFO
	echo "\"repo_type\" : \"deb\"," >> $TMP_PKG_LIST
	echo "\"package_manager\" : \"apt\"," >> $TMP_PKG_LIST
	echo "\"packages\" : [" >> $TMP_PKG_LIST
	i=0
	dpkg -l | grep "^ii" | while read package; do
		package_name=$(echo $package | awk '{print $2}')
		package_arch=$(echo $package | awk '{print $4}')

		i=$(expr $i + 1)
		package_details=$(apt-cache policy ${package_name})
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
			echo "	\"name\": \"$package_name\"," >> $TMP_PKG_LIST
			echo "	\"version\": \"$installed_ver\"," >> $TMP_PKG_LIST
			echo "	\"architecture\": \"$package_arch\"," >> $TMP_PKG_LIST
			echo "	\"repository\": \"$repository_url\"," >> $TMP_PKG_LIST
			echo "}," >> $TMP_PKG_LIST
			echo "{" >> $TMP_PKG_LIST
			echo "	\"upgrade\": \"yes\"," >> $TMP_PKG_LIST
			echo "	\"name\": \"$package_name\"," >> $TMP_PKG_LIST			
			echo "	\"version\": \"$candidate_ver\"," >> $TMP_PKG_LIST
			echo "	\"repository\": \"$candidate_repository_url\"" >> $TMP_PKG_LIST
			if [ $i -eq $packages_count ]; then
				echo "}" >> $TMP_PKG_LIST
			else
				echo "}," >> $TMP_PKG_LIST
			fi
		else
			# We only print package + repo
			echo "{" >> $TMP_PKG_LIST
			echo "	\"name\": \"$package_name\"," >> $TMP_PKG_LIST
			echo "	\"version\": \"$installed_ver\"," >> $TMP_PKG_LIST
			echo "	\"architecture\": \"$package_arch\"," >> $TMP_PKG_LIST
			echo "	\"repository\": \"$repository_url\"," >> $TMP_PKG_LIST
			if [ $i -eq $packages_count ]; then
				echo "}" >> $TMP_PKG_LIST
			else
				echo "}," >> $TMP_PKG_LIST
			fi
		fi
		echo -ne "$i out of $packages_count \r"
	done
	echo "]" >> $TMP_PKG_LIST
}

get_yum_packages() {
	if $UPDATE; then
		yum makecache --quiet
	fi

	# while loop will execute in current shell instead of a sub-shell
	#shopt -s lastpipe 

	truncate -s 0  $TMP_REPO_INFO
	echo "\"repositories\" : [" >> $TMP_REPO_INFO
	yum repoinfo | while read line; do
		field_id=$(echo $line | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]' | awk -F':' '{print $1}')
		field_val=$(echo $line | cut -d ':' -f 2- )
		case "$field_id" in
		    "repo-id")
				repo_alias=$(echo $field_val | awk -F'/' '{print $1}' | tr -d '[:space:]')
				echo "	{" >> $TMP_REPO_INFO
				echo "		\"repository_alias\" : \"$repo_alias\"," >> $TMP_REPO_INFO
				echo "		\"repository_id\" : \"$field_val\"," >> $TMP_REPO_INFO
		       ;;
		   "repo-name")
				echo "		\"repository_name\" : \"$field_val\"," >> $TMP_REPO_INFO
		    	;;
		   "repo-baseurl")
		   		baseurl=$(echo $field_val | awk '{print $1}')
		   		echo "		\"repository_baseurl\" : \"$baseurl\"" >> $TMP_REPO_INFO
				echo "	}," >> $TMP_REPO_INFO
		    	;;
		esac
	done
	echo "]" >> $TMP_REPO_INFO

	packages_count=$(repoquery '*' --queryformat='%{name} %{evr} %{ui_from_repo}' --installed | wc -l)
	#upgrades_count=$(yum check-updates | grep -v ^$ | wc -l)

	if [ ! $QUIET ]; then
		echo "Found $packages_count installed packages"
	fi

	truncate -s 0 $TMP_PKG_LIST
	echo "\"repo_type\" : \"rpm\"," >> $TMP_PKG_LIST
	echo "\"package_manager\" : \"yum\"," >> $TMP_PKG_LIST
	echo "\"packages\" : [" >> $TMP_PKG_LIST

	i=0
	yum check-updates | grep -v ^$ | while read package; do
		package_name=$(echo $package | awk '{print $1}' | awk -F'.' '{print $1}')
		package_arch=$(echo $package | awk '{print $1}' | awk -F'.' '{print $2}')
		package_version=$(echo $package | awk '{print $2}')
		package_aliasrepo=$(echo $package | awk '{print $3}')

		echo "{" >> $TMP_PKG_LIST
		echo "	\"upgrade\": \"yes\"," >> $TMP_PKG_LIST
		echo "	\"name\": \"$package_name\"," >> $TMP_PKG_LIST
		echo "	\"version\": \"$package_version\"," >> $TMP_PKG_LIST
		echo "	\"architecture\": \"$package_arch\"," >> $TMP_PKG_LIST
		echo "	\"repository_alias\": \"$package_aliasrepo\"," >> $TMP_PKG_LIST
		echo "}," >> $TMP_PKG_LIST
	done

	i=0
	# simply listing with yum can possibly break some columns
    repoquery '*' --queryformat='%{name} %{evr} %{ui_from_repo}' --installed | while read package; do
		i=$(expr $i + 1)

        package_name=$(echo $package | awk '{print $1}' | awk -F'.' '{print $1}')
        package_arch=$(echo $package | awk '{print $1}' | awk -F'.' '{print $2}')
        package_version=$(echo $package | awk '{print $2}')
        package_aliasrepo=$(echo $package | awk '{print $3}')

		echo "{" >> $TMP_PKG_LIST
		echo "	\"name\": \"$package_name\"," >> $TMP_PKG_LIST
		echo "	\"version\": \"$package_version\"," >> $TMP_PKG_LIST
		echo "	\"architecture\": \"$package_arch\"," >> $TMP_PKG_LIST
		echo "	\"repository_alias\": \"$package_aliasrepo\"," >> $TMP_PKG_LIST
		if [ $i -eq $packages_count ]; then
			echo "}" >> $TMP_PKG_LIST
		else
			echo "}," >> $TMP_PKG_LIST
		fi
    done

	echo "]" >> $TMP_PKG_LIST

}

cleanup() {
	rm $TMP_HOST_INFO
	rm $TMP_PKG_LIST
	rm $TMP_REPO_INFO
	rm $TMP_PAYLOAD
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

truncate -s 0 $TMP_PAYLOAD

echo "{" >> $TMP_PAYLOAD
cat $TMP_HOST_INFO >> $TMP_PAYLOAD
echo "," >> $TMP_PAYLOAD
echo "\"tags\" : \"$TAGS\"," >> $TMP_PAYLOAD
if [ -s $TMP_REPO_INFO ]; then
	cat $TMP_REPO_INFO >> $TMP_PAYLOAD
	echo "," >> $TMP_PAYLOAD
fi
cat $TMP_PKG_LIST >> $TMP_PAYLOAD
echo "}" >> $TMP_PAYLOAD


curl -X POST -H "Content-Type: application/json" -d @$TMP_PAYLOAD $SERVER_URL
cleanup


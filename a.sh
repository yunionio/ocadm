#!/usr/bin/env bash
rmdb ()
{
    local tmp=$(mktemp);
    local user=${DB_USER:-};
    local pswd=${DB_PSWD:-};
    local host=${DB_HOST:-};
    if [ -f /opt/yunion/upgrade/config.yml ]; then
        # rm -i -f $tmp;
        parse_yaml /opt/yunion/upgrade/config.yml > $tmp;
        source $tmp;
    fi;
    if [ -z "$user" ]; then
        user="$primary_master_node_db_user";
    fi;
    if [ -z "$host" ]; then
        host="$primary_master_node_db_host";
    fi;
    if [ -z "$pswd" ]; then
        pswd="$primary_master_node_db_password";
    fi;
    if ! mysql -u "$user" -p"$pswd" -h "$host" -e ";"; then
        echo "rm db failed: mysql -u $user -p$pswd -h $host -e ';'";
        return;
    fi;
    if [ -z "$user" ]; then
        user=root;
    fi;
    if [ -z "$host" ]; then
        host=localhost;
    fi;
    mysql -u "$user" -p"$pswd" -h "$host" -e "show databases" | grep --color=auto -Pv '^mysql$|_schema|^Database' | while read line; do
        mysql -u "$user" -p"$pswd" -h "$host" -e "drop database $line;" && echo "dropped db $line";
    done
}

parse_yaml ()
{
    local prefix=$2;
    local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034');
    sed -ne "s|^\($s\):|\1|" -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p" $1 | awk -F$fs '{
                indent = length($1)/2;
                vname[indent] = $2;
                for (i in vname) {if (i > indent) {delete vname[i]}}
                if (length($3) > 0) {
                        vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
                        printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
                }
        }'
}

#rmdb


#!/bin/bash -x

# install java 8
wget --no-cookies --no-check-certificate --header "Cookie:oraclelicense=accept-securebackup-cookie" "http://download.oracle.com/otn-pub/java/jdk/8u131-b11/d54c1d3a095b4ff2b6607d096fa80163/jdk-8u131-linux-x64.rpm"
yum -y localinstall jdk-8u131-linux-x64.rpm

# install cassandra 3
cat >/etc/yum.repos.d/cassandra.repo <<EOF
[cassandra]
name=Apache Cassandra
baseurl=https://www.apache.org/dist/cassandra/redhat/311x/
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://www.apache.org/dist/cassandra/KEYS
EOF
yum -y install cassandra nc

# setup and wait for seed
hostname=`hostname`
ip=`hostname -i`
instance=""
ad_index=""
base_hostname=""
while IFS='-' read -ra ADDR; do
  for i in "${ADDR[@]}"; do
    ad_index=$instance
    instance=$i
    # skip the last by using previous
    if [ "$previous_val" != "" ]; then
      base_hostname="$base_hostname$previous_val-"
    fi
    previous_val=$ad_index
  done
done <<< "$hostname"

original_seed="127.0.0.1"
domain=`hostname -d`

# use AD-instance suffix convention, w/ 1-1 for seed
seed_index="1"
domain=${domain//$ad_index/$seed_index}
new_seed="$base_hostname$seed_index-$seed_index.$domain"

if [ $instance -eq 1 ] && [ $ad_index -eq 1 ]; then
  # cannot use hostname for seed itself, but ok for non-seeds
  new_seed=$ip
else
  `nc -w 5 --send-only $new_seed 7000 </dev/null`
  ec=$?
  while [ $ec -gt 0 ]; do
    sleep 5
    echo "checking for $new_seed 7000..."
    `nc -w 5 --send-only $new_seed 7000 </dev/null`
    ec=$?
  done
  sindex=$(expr $instance - 1)
  sleep $(expr $sindex \* 40)
fi

sed -i "s/$original_seed/$new_seed/g" /etc/cassandra/conf/cassandra.yaml
sed -i "s/localhost/$ip/g" /etc/cassandra/conf/cassandra.yaml

systemctl stop firewalld
systemctl disable firewalld

systemctl start cassandra
systemctl enable cassandra

# handle bootstrap exception for 10min
for i in `seq 1 20`; do
  `systemctl status cassandra`
  ec=$?
  if [ $ec -gt 0 ]; then
    systemctl start cassandra
  fi
  sleep 30
done

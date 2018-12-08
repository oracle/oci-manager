#!/bin/bash -x

# install apache httpd 
yum -y install httpd nc

systemctl stop firewalld
systemctl disable firewalld

systemctl start httpd
systemctl enable httpd

echo > /var/www/html/index.html
echo "hello world" >> /var/www/html/index.html

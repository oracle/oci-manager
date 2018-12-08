#!/bin/bash

KEY_FILE=`grep key_file ~/.oci/config | awk -F "=" '{ print $2 }' | head -1 | xargs`
kubectl create secret generic apikey -n oci-system --from-file=apikey=$KEY_FILE
echo "Created secret in namespace: oci-system for oci apikey key_file: $KEY_FILE"

TMP_OCICONFIG_FILE=/tmp/ociconfig

grep -v key_file ~/.oci/config > $TMP_OCICONFIG_FILE
echo "key_file=/etc/oci-apikey/apikey" >> $TMP_OCICONFIG_FILE

kubectl create configmap ociconfig -n oci-system --from-file=config=$TMP_OCICONFIG_FILE
echo "Created configmap for ociconfig based on your ~/.oci/config"
rm $TMP_OCICONFIG_FILE

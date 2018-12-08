# Assuming the file is in your home directory
chmod +x ./get-kubeconfig.sh

# Export ENDPOINT which points to the region's api server
export ENDPOINT=containerengine.us-phoenix-1.oraclecloud.com

# Retrieve and save the kubeconfig
./get-kubeconfig.sh ocid1.cluster.oc1.phx.aaaaaaaaafsgiobvgmzgimbugi2tgnzzmq2gkyzrge3wgnjugctgmmrsgntd > ~/kubeconfig
 
# Use the kubeconfig with kubectl commands
export KUBECONFIG=~/kubeconfig
kubectl get nodes

minikube start --cpus 4 --memory 8192 \
  --apiserver-port=7445 \
  --network-plugin=cni \
  --cni=false

"$(dirname "$0")/install-cilium.sh"

flux bootstrap github \
  --owner=antoncuranz \
  --repository=cloudlab \
  --branch=local \
  --path="./clusters/local" \
  --personal

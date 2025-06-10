#!/bin/bash

# Function to print help message
print_help() {
  echo "Usage: $0 -n <namespace>"
  echo "Options:"
  echo "  -n <namespace>    Specify the namespace"
  echo "  -h, --help        Show this help message"
}

# Function to print error messages in red
print_error() {
  echo -e "\033[31m$1\033[0m"
}

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    -n) NAMESPACE="$2"; shift ;;
    -h|--help|-help) print_help; exit 0 ;;
    *) print_error "Unknown parameter passed: $1"; print_help; exit 1 ;;
  esac
  shift
done

# Check if namespace is provided
if [ -z "$NAMESPACE" ]; then
  print_error "Namespace not specified!"
  print_help
  exit 1
fi

OUTPUT_FILE="SriovT2Card-Admin-logs.txt"

# Check if the namespace exists
if ! kubectl get namespace "$NAMESPACE" > /dev/null 2>&1; then
  print_error "Namespace '$NAMESPACE' does not exist!"
  exit 1
fi

# Function to collect logs from all matching Pods
collect_logs() {
  echo "Collecting logs from namespace: $NAMESPACE, Pods starting with: sriovt2card-admin"
  > "$OUTPUT_FILE"
  while true; do
    pods=$(kubectl get pods -n "$NAMESPACE" | awk '/sriovt2card-admin/ {print $1}')
    if [ -z "$pods" ]; then
      echo "No pods found matching the criteria."
    else
      echo "Found pods: $pods"
      for pod in $pods; do
        echo "Logs from pod: $pod" >> "$OUTPUT_FILE"
        kubectl logs -n "$NAMESPACE" "$pod" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
      done
    fi
    sleep 2
  done
}

# Start collecting logs
collect_logs


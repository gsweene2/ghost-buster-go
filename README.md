# ghost-buster-go

## Run Locally
```bash
go get github.com/gsweene2/ghost-buster-go

go build

# Set token and signing secret

go run events.go
```

## Run with Docker
```bash
# build
docker build --tag ghost-buster-go .

# Tag
docker tag ghost-buster-go:latest ghost-buster-go:v1.0

# Run
docker run --publish 8080:8080 ghost-buster-go:v1.0
```

### Push to ECR
```bash

# Locate ECR Repo and export

# Login
aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin $ECR_REPO

# Build
docker build -tag ghost-buster-go .

# Tag and Push
docker tag ghost-buster-go:latest $ECR_REPO/ghost-buster-go:v1.0
docker push $ECR_REPO/ghost-buster-go:v1.0
```

## Create EKS Cluster

Export
```bash
. .env
```

Create file with your variables
```bash
cat << EOF > eks.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: ghost-buster-go
  region: ${AWS_REGION}
  version: "1.19"

availabilityZones: ["${AZS[0]}", "${AZS[1]}", "${AZS[2]}"]

managedNodeGroups:
- name: nodegroup
  desiredCapacity: 3
  instanceType: t2.small
  ssh:
    enableSsm: true

cloudWatch:
 clusterLogging:
   enableTypes: ["*"]

secretsEncryption:
  keyARN: ${MASTER_ARN}
EOF
```

Use eksctl to create cluster
```bash
eksctl create cluster -f eks.yaml
```

## Create objects
```bash
# Slack App
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Nginx
kubectl apply -f nginx/deployment.yaml
kubectl apply -f nginx/service.yaml
```

## Test nginx
```bash
# Get Load Balancer Address
kubectl get service/nginx-service-loadbalancer |  awk {'print $1" " $2 " " $4 " " $5'} | column -t
# Curl
curl -silent abdcde0b86ec444978dc7aa0e5daad54-295517781.us-east-2.elb.amazonaws.com:80 | grep title
# Expected
<title>Welcome to nginx!</title>
```

## View in Console

To see k8s objects and resources in AWS Console, add your user to k8s auth.

2 ways to do this:

1. Using eksctl
    ```bash
    # Configure Variables specific to you
    export rolearn=arn:aws:iam::111122223333:user/garrett.sweeney
    export username=garrett.sweeney
    eksctl create iamidentitymapping --cluster ghost-buster-go --arn ${rolearn} --group system:masters --username ${username}
    # Confirm user added
    kubectl describe configmap -n kube-system aws-auth
    ```

2. Modify the config map with kubectl
    ```bash
    # View existing map
    kubectl describe configmap -n kube-system aws-auth
    kubectl edit configmap -n kube-system aws-auth

    mapUsers: |
      - userarn: arn:aws:iam::111122223333:user/garrett.sweeney
        username: garrett.sweeney
        groups:
          - system:masters
    ```

## Errors & Issues

Can't see cluster objects in console?
https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html

Unauthorized or Access Denied
https://docs.aws.amazon.com/eks/latest/userguide/troubleshooting.html#unauthorized

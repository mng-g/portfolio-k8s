# Portfolio Kubernetes Deployment

This repository contains the necessary configurations and instructions to deploy a portfolio application with frontend and backend components on Kubernetes (K8s). The architecture involves multiple services, including database, application backend, and frontend, all orchestrated on K8s.

## Table of Contents
- [Build, Tag, and Push Docker Images](#build-tag-and-push-docker-images)
  - [Frontend](#frontend)
  - [Backend](#backend)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Testing Locally](#test-locally)
- [Cloud-Native PostgreSQL Operator](#get-cloudnative-pg-operator-and-create-cluster)
- [Expose Services on the Internet](#expose-on-internet-for-test)
- [TLS Configuration](#tls)
- [Monitoring](#monitoring)
- [Logging](#logging)

---

## Build, Tag, and Push Docker Images

Before deploying your application on Kubernetes, build and push the Docker images for both the frontend and backend.

### Frontend

```bash
echo "<YOUR_GITHUB_PAT>" | docker login ghcr.io -u mng-g --password-stdin
docker build -t ghcr.io/mng-g/go-frontend-app:latest .
docker tag ghcr.io/mng-g/go-frontend-app:latest ghcr.io/mng-g/go-frontend-app:vX.X.X
docker push ghcr.io/mng-g/go-frontend-app:latest
docker push ghcr.io/mng-g/go-frontend-app:vX.X.X
```

### Backend

```bash
echo "<YOUR_GITHUB_PAT>" | docker login ghcr.io -u mng-g --password-stdin
docker build -t ghcr.io/mng-g/go-backend-app:latest .
docker tag ghcr.io/mng-g/go-backend-app:latest ghcr.io/mng-g/go-backend-app:vX.X.X
docker push ghcr.io/mng-g/go-backend-app:latest
docker push ghcr.io/mng-g/go-backend-app:vX.X.X
```

---

## Kubernetes Deployment

### Create Deployments

Deploy the frontend and backend applications on Kubernetes with the following commands:
```bash
kubectl apply -f deploy/helm/namespaces
kubectl apply -f deploy/helm/templates
```
An Ingress will expose the app on [go-app.local](go-app.local). You may need to edit your /etc/hosts file adding the domain and the external IP of the ingress-controller. 

### Set GitHub Container Registry Token

To pull private images, create a Kubernetes secret to store your GitHub token:

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=mng-g \
  --docker-password=<YOUR_GITHUB_PAT> \
  --docker-email=mingazzini.michael@gmail.com -n go-app
```

### Add `imagePullSecrets` to Deployment

Ensure that your deployment configurations reference the secret for pulling images:

```yaml
imagePullSecrets:
  - name: ghcr-secret
```

### Expose Services

Expose the frontend and backend deployments as services:

```bash
kubectl expose deployment -n go-app go-backend --name go-backend-svc --port 9191
kubectl expose deployment -n go-app go-frontend --name go-frontend-svc --port 9090
```

### Port Forward for Local Testing

To test the services locally, use port forwarding:

```bash
kubectl port-forward service/go-backend-svc -n go-app 9191:9191
kubectl port-forward service/go-frontend-svc -n go-app 9090:9090
```

---

## Test Locally

To test the application locally, use the following steps to run the services with Docker.

1. **Build Docker Images Locally:**

   ```bash
   cd go-backend-app/
   docker build -t ghcr.io/mng-g/go-backend-app:latest .
   cd ../go-frontend-app/
   docker build -t ghcr.io/mng-g/go-frontend-app:latest .
   cd ..
   ```

2. **Create a Docker Network:**

   ```bash
   docker network create demo-portfolio-k8s
   ```

3. **Run PostgreSQL:**

   ```bash
   docker run --name go-postgres --network demo-portfolio-k8s \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_DB=mydb \
     -p 5432:5432 -d postgres:latest
   ```

4. **Run Backend:**

   ```bash
   docker run --name go-backend --network demo-portfolio-k8s \
     -e DB_HOST=go-postgres \
     -e DB_PORT=5432 \
     -e DB_USER=postgres \
     -e DB_PASS=password \
     -e DB_NAME=mydb \
     -p 9191:9191 -d ghcr.io/mng-g/go-backend-app:latest
   ```

5. **Run Frontend:**

   ```bash
   docker run -p 9797:9090 -d ghcr.io/mng-g/go-frontend-app:latest
   ```

---

## Get CloudNative-PG Operator and Create Cluster

To deploy PostgreSQL using the CloudNative-PG operator, run the following:

```bash
kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.25/releases/cnpg-1.25.1.yaml
kubectl apply -f deploy/helm/database/postgres-cluster.yaml

curl -sSfL \
  https://github.com/cloudnative-pg/cloudnative-pg/raw/main/hack/install-cnpg-plugin.sh | \
  sudo sh -s -- -b /usr/local/bin

kubectl cnpg status go-postgres -n go-app
```

---

## Expose on Internet for Test

You can expose your services publicly for testing using [ngrok](https://ngrok.com/):

```bash
ngrok http --host-header=go-app.local 172.28.100.0:80 --url=becoming-mutt-forcibly.ngrok-free.app --basic-auth="user:password"
```

---

## TLS

For securing your application with TLS, use [cert-manager](https://cert-manager.io/docs/) to manage certificates.

### Install cert-manager:

```bash
helm repo add jetstack https://charts.jetstack.io --force-update
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --set 'extraArgs={--dns01-recursive-nameservers-only,--dns01-recursive-nameservers=8.8.8.8:53\,1.1.1.1:53}'
```

### Generate and Apply Certificates:

1. **Self-Signed Certificate:**

   ```bash
   kubectl apply -f deploy/helm/certificate/selfsigned-cluster-issuer.yaml
   ```

2. **Signed by Internal CA:**

   ```bash
   openssl req -x509 -new -nodes -keyout ca.key.pem -out ca.cert.pem -days 365 -subj "/CN=MyCA"
   kubectl create secret tls -n cert-manager ca-key-pair --cert=ca.cert.pem --key=ca.key.pem
   kubectl apply -f deploy/helm/certificate/internal-ca-cluster-issuer.yaml
   ```

3. **Configure Ingress for TLS:**

   ```bash
   kubectl replace -f deploy/helm/templates/ingress.yaml
   ```

---

## Monitoring

Install Prometheus and Grafana for monitoring:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace
```

### Access Dashboards:

- **Prometheus Dashboard:** [http://localhost:9090](http://localhost:9090)

  ```bash
  kubectl port-forward svc/prometheus-stack-kube-prom-prometheus -n monitoring 9090
  ```

- **Grafana Dashboard:** [http://localhost:3000](http://localhost:3000)

  ```bash
  export POD_NAME=$(kubectl --namespace monitoring get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=prometheus-stack" -oname)
  kubectl --namespace monitoring port-forward $POD_NAME 3000
  ```

  Retrieve the Grafana admin password:

  ```bash
  kubectl --namespace monitoring get secrets prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo
  ```

### Deploy Service Monitor:

```bash
kubectl apply -f deploy/helm/monitoring/backend-servicemonitor.yaml
```

Add Prometheus as a data source in Grafana and create dashboards.

---

## Logging

Set up centralized logging with Grafana Loki:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install loki grafana/loki-stack \
  --namespace logging \
  --create-namespace \
  --set loki.enabled=true \
  --set promtail.enabled=true \
  --set promtail.config.server.http_listen_port=9080 \
  --set promtail.config.server.grpc_listen_port=0
```

Add Loki as a data source in Grafana and create dashboards for logs.
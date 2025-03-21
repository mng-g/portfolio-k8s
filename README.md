# portfolio-k8s

### Build, tag and push images
#### Frontend
```bash
echo "<YOUR_GITHUB_PAT>" | docker login ghcr.io -u mng-g --password-stdin
docker build -t ghcr.io/mng-g/go-frontend-app:latest .
docker tag ghcr.io/mng-g/go-frontend-app:latest ghcr.io/mng-g/go-frontend-app:vX.X.X
docker push ghcr.io/mng-g/go-frontend-app:latest
docker push ghcr.io/mng-g/go-frontend-app:vX.X.X
```
#### Backend
```bash
echo "<YOUR_GITHUB_PAT>" | docker login ghcr.io -u mng-g --password-stdin
docker build -t ghcr.io/mng-g/go-backend-app:latest .
docker tag ghcr.io/mng-g/go-backend-app:latest ghcr.io/mng-g/go-backend-app:vX.X.X
docker push ghcr.io/mng-g/go-backend-app:latest
docker push ghcr.io/mng-g/go-backend-app:vX.X.X
```
#### Deploy on K8s
```bash
k create deployment go-frontend --image ghcr.io/mng-g/go-frontend-app:vX.X.X -n go-app
k create deployment go-backend --image ghcr.io/mng-g/go-backend-app:vX.X.X -n go-app
```

You need to set ghcr token (PAT) to pull private images:
```bash
kubectl create secret docker-registry ghcr-secret --docker-server=ghcr.io --docker-username=mng-g --docker-password=<YOUR_GITHUB_PAT> --docker-email=mingazzini.michael@gmail.com -n go-app
```
Add to the deployment:
```yaml
      imagePullSecrets:
        - name: ghcr-secret
```
```bash
k expose deployment -n go-app go-backend --name go-backend-svc --port 9191
k expose deployment -n go-app go-frontend --name go-frontend-svc --port 9090
k port-forward service/go-backend-svc -n go-app 9191:9191
k port-forward service/go-frontend-svc -n go-app 9090:9090
```

## Test locally:
```bash
cd go-backend-app/
docker build -t ghcr.io/mng-g/go-backend-app:latest .
cd ../go-frontend-app/
docker build -t ghcr.io/mng-g/go-frontend-app:latest .
cd ..

docker network create demo-portfolio-k8s

docker run --name go-postgres --network demo-portfolio-k8s \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=mydb \
  -p 5432:5432 -d postgres:latest

docker run --name go-backend --network demo-portfolio-k8s \
  -e DB_HOST=go-postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASS=password \
  -e DB_NAME=mydb \
  -p 9191:9191 -d ghcr.io/mng-g/go-backend-app:latest

docker run -p 9797:9090 -d ghcr.io/mng-g/go-frontend-app:latest
```



### Get cloudnative-pg operator and create cluster
```bash
kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.25/releases/cnpg-1.25.1.yaml
k apply -f deploy/helm/templates/postgres-cluster.yaml

curl -sSfL \
  https://github.com/cloudnative-pg/cloudnative-pg/raw/main/hack/install-cnpg-plugin.sh | \
  sudo sh -s -- -b /usr/local/bin

k cnpg status go-postgres -n go-app
```

### Deploy on k8s: 
```bash
k apply -f deploy/helm/templates
```


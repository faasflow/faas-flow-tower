# faas-flow-tower
A monitoring function stack to visualize faas-flow functions and requests in realtime
    
Dashboard provide details for each faas-flow functions incuding graphical representation of dags based on function definition
![alt dashboard](doc/dashboard.png)   

Tower provides realtime timeline for requests for individual nodes of each faas-flow functions
![alt dashboard](doc/monitoring.png)    
   

## Deploy OpenFaaS
FaasFlow Tower requires the OpenFaaS to be deployed and the OpenFaaS Cli. You can either have your OpenFaaS deployed in Kubernets otherwise in Swarm. To deploy OpenFaaS and to install the OpenFaaS cli client follow this guide: [https://docs.openfaas.com/deployment/](https://docs.openfaas.com/deployment/).     

> Note: If you have deployed your OpenFaaS in Kubenets, it is recomanded to deploy FaaSFlow Tower services in same environment to simplify configuration

## Deploy Faas-flow Components with Faas-flow Infra
> Faas-Flow infra provides the kubernets is and swarm deployment resources for faas-flow dependencies 
[https://github.com/s8sg/faas-flow-infra](https://github.com/s8sg/faas-flow-infra)

## Deploy Faas-flow Tower

### Configure 
Configuration are defined in `conf.yml`. Based on your deployment you may need to update the configuration before you use the deployment script.   

#### Kubernets
For components deployed in kubernets   
```yaml
environment:
  gateway_url: "http://gateway.openfaas:8080/" 
  basic_auth: true
  secret_mount_path: "/var/openfaas/secrets"
  trace_url: "http://jaegertracing.faasflow:16686/" 
```
#### Docker Swam
For components deployed in swarm
```yaml
environment:
  gateway_url: "http://gateway:8080/" 
  basic_auth: true
  secret_mount_path: "/var/openfaas/secrets"
  trace_url: "http://jaegertracing:16686/" 
```

### Deploy Functions
Deploy the OpenFaaS functions in the OpenFaaS 
```sh
faas deploy -g localhost:8080
```
Change the `localhost:8080` to your openfaas Gateway URL 

## Access the Dashboard
Once deployed the dashboard will be available as a openfaas function at 
[localhost:8080/function/faas-flow-dashboard](localhost:8080/function/faas-flow-dashboard)     
Change the `localhost:8080` to your openfaas Gateway URL    

## Make your flow visible 
To make flow functions visible in the dashboard add `faas-flow : 1` label in `stack.yml` of each flow functions  
```yaml
annotations:
   faas-flow-desc: "option labels to provide flow descriptions"
labels:
   faas-flow : 1
``` 
   
#### Monitoring
Faasflow fetches the monitoring information from jaeger trace server. To enable tracing for flow function add environment `enable_tracing: true` and set trace server url `trace_server: "jaegertracing:5775"` in `stack.yml`. For kubernets use `trace_server: "jaegertracing.openfaas:5775"`

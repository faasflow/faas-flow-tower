# faas-flow-tower
A monitoring function stack to visualize faas-flow functions and requests in realtime
    
### Dashboard
Dashboard provide details for each faas-flow functions incuding graphical representation of dags based on function definition
![alt dashboard](doc/dashboard.png)
To make flow functions visible in dashboard add the below labels in `stack.yml` of each flow functions  
```
annotations:
   faas-flow-desc: "option labels to provide flow descriptions"
labels:
   faas-flow : 1
``` 
  
   
### Monitoring
Tower provides realtime timeline for requests for individual nodes of each faas-flow functions
![alt dashboard](doc/monitoring.png)
Faasflow fetches the monitoring information from trace server    
For flow function enable tracing `enable_tracing: true` and set trace server url `trace_server: "jaegertracing:5775"`    
Provide the same trace server api url `trace_url: "jaegertracing:16686"` in `conf.yml`   
   

### Getting Started
FaaS-Flow Tower comes with the default `StateStore`, `DataStore` and `EventManager`

 |Item|Implementation|
 |---|---|
 |StateStore|[Consul StateStore](https://github.com/s8sg/faas-flow-consul-statestore)|
 |DataStore|[Minio DataStore](https://github.com/s8sg/faas-flow-minio-datastore)|
 |EventStore|[Jaguar](https://github.com/jaegertracing/jaeger)|


#### Deploy in Swarm 
To deploy in OpenFaaS need to be running. The local host need to have swarm node initialized. 
###### Clone the repo
```
git clone https://github.com/s8sg/faas-flow-tower
cd faas-flow-tower
```
###### Set OpenFaaS Gateway
Update the `stack.yml` with the gateway url
```
provider:
  name: faas
  gateway: http://127.0.0.1:8080
```
###### Deploy with the script
```
./deploy.sh
```
This script will deploy the function in and the other services in the same network as OpenFaaS 


#### Deploy in Kubernets
```

```

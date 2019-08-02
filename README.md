# faas-flow-tower
A monitoring function stack to visualize faas-flow functions and requests in realtime
    
### Dashboard
Dashboard provide details for each faas-flow functions incuding graphical representation of dags based on function definition
![alt dashboard](doc/dashboard.png)
To make flow functions visible in dashboard add the below labels in `stack.yml`   
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
   

### Build and Deploy
```
git clone https://github.com/s8sg/faas-flow-dashboard
cd faas-flow-dashboard
faas build
faas deploy
```

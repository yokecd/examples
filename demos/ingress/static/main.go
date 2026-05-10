package main

import "fmt"

func main() {
	fmt.Println(`
- kind: Deployment
  apiVersion: apps/v1
  metadata:
    name: flight
  spec:
    replicas: 1
    selector:
      matchLabels:
        app.kubernetes.io/name: flight
    template:
      metadata:
        name: flight
        labels:
          app.kubernetes.io/name: flight
      spec:
        containers:
          - name: main
            image: davidmdm/c4ts:latest
            env:
              - name: PORT
                value: "3000"
            resources: {}
    strategy: {}
  status: {}
- kind: Service
  apiVersion: v1
  metadata:
    name: flight
  spec:
    ports:
      - name: http
        protocol: TCP
        port: 80
        targetPort: 3000
    selector:
      app.kubernetes.io/name: flight
  status:
    loadBalancer: {}
- kind: Ingress
  apiVersion: networking.k8s.io/v1
  metadata:
    name: flight
  spec:
    rules:
      - http:
          paths:
            - path: /miaow
              pathType: Prefix
              backend:
                service:
                  name: flight
                  port:
                    name: http
  status:
    loadBalancer: {}`)
}

apiVersion: platform.acme.co/v1alpha1
kind: FleetManager
metadata:
  name: staging-fleet
spec:
  parameters:
    fleetName: api-servers
    region: us-east-1
    instanceCount: 3
    environment: staging
    tags:
      team: backend
      environment: staging 
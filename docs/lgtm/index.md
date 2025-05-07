# LGTM Observability stack

## Invidivual

Each service from the stack can be run individually, with its own docker-compose file.
For simplicity, I will list the commands here:

### Base command

```shell
docker-compose -f .\deployments\docker\lgtm-individual\docker-compose.{name}.yml up -d 
```
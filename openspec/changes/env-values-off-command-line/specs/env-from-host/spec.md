# Spec Delta

## REMOVED Requirements

### Requirement: Values stay off the docker command line
**Reason**: Generalized to all environment variables by the container-assembly requirement "Environment values passed via stdin env file", which also removes the need for the `DOCKER_*` exception.
**Migration**: None; mapped values are still kept off the `docker run` arguments.

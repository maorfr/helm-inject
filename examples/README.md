## Examples

### myExecutable


This is an extremely simple script which makes sure that resources of type 'Deployment' are always applied with an apiVersion of apps/v1.
Custom scripts such as this one should consider more cases.
For example, multiple resources in a single file.

To test this script, copy it to your $PATH, and execute the following command:
```
# Positive test

myExecutable inject /path/to/chart/templates/deployment.yaml

# Negative test

myExecutable inject /path/to/chart/templates/service.yaml
```
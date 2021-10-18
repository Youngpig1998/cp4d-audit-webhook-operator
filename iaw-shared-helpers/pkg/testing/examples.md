# Testing Package Examples

The testing package provides helper functions for unit testing. This document outlines example usage for different testing types and some of the helper functions included in the package. 

## File Utils 

The package contains useful common file access and manipulation functions. 

### Checking Files In a Pod

If a requirement of a test is to check for the existence or content of a particular file within a Kubernetes Pod, the `CopyFromPod` function can facilate this. Below are code snippets of an example use case checking if a file has content. The `IsStringInFile` method can be used to further check for the existence of a particular string within the file:

To define the target pod to access files from:

```go
targetPod = testing.CopyPod{
    Namespace:     namespace,
    Name:          podName,
    ContainerName: containerName,
}
```

Within a test `It` block:

```go
Eventually(func() error {
    err := targetPod.CopyFromPod(fmt.Sprintf(DirectoryPath, fileName), "")
    if err != nil {
        return err
    }
    empty, err := testing.IsFileEmpty(fileName)
    if err != nil {
        return errors.Errorf("failed to check if %s was empty: %s", fileName, err.Error())
    }
    if empty {
        return errors.Errorf("expected %s to have some contents, but it is empty", fileName)
    }
    return nil
}, 30, 5).ShouldNot(HaveOccurred(), fmt.Sprintf("failed to get %s with expected contents", fileName))
```

## Model Tests

The testing shared helper package has methods to help facilitate the writing of Operator Model tests. The following snippets demonstrate example usage. 

The following methods should be included in your suite test file:

replace _typeinstances_ and _typemodel_ with the type you are working with, for example `elasticv1beta1instances` and `elasticModelv1m0p0` would be correct for elasticsearch.

```go
func modelResourceTest(instances []typeinstances.TestInstance, resourceType string, extractor func(typemodel.Model) map[types.NamespacedName]interface{}) {
    for _, instance := range instances {
        instance := instance // See https://onsi.github.io/ginkgo/#patterns-for-dynamically-generating-tests
        It(fmt.Sprintf("Matches the snapshot for %s created by %s", resourceType, instance.Name), func() {
            model, err := typemodel.Create(instance.Instance.Spec.Version, instance.Instance, scheme.Scheme)
            Expect(err).NotTo(HaveOccurred())
            resources := extractor(model)
            iafschtesting.ResourceSnapshotTest(instance.Name, resourceType, resources)
        })
    }
}
```

Add all CR instances you wish to test to the `defaultInstances` array. For example for elastic, one valid test instance is `elasticv1beta1instances.ESMinimal()` and another is	`elasticv1beta1instances.ESWithDefaultStorage()`.

```go
func defaultInstances() []typeinstances.TestInstance {
    return []typeinstances.TestInstance{
        typeinstances.CRInstance(),
        typeinstances.CRInstance2(),
    }
}
```

The following is an example of a test instance to be included in your model tests file:

```go
var _ = Describe("V1m0p0", func() {
	Describe("Create", func() {
		Context("ConfigMap", func() {
			extractor := func(model typemodel.Model) map[types.NamespacedName]interface{} {
				resource, resourceName := model.ConfigMap()
				return map[types.NamespacedName]interface{}{
					resourceName: resource,
				}
			}
			modelResourceTest(defaultInstances(), "ConfigMap", extractor)
		})
	})
})
```

Duplicate the test Context for each of the resources your model creates.

For resource types for which there can be more than one instance of this resource included in the model, a loop can be added as in the following example:

```go
Context("StatefulSets", func() {
    extractor := func(model elasticModelv1m0p0.Model) map[types.NamespacedName]interface{} {
        resources := model.StatefulSets()
        extractedResources := make(map[types.NamespacedName]interface{})
        for namespacedName, resource := range resources {
            extractedResources[namespacedName] = resource
        }
        return extractedResources
    }
    modelResourceTest(defaultInstances(), "StatefulSets", extractor)
})
```

### Model Test Snapshots

Snapshot yaml files are used to facilitate model testing resource comparisons - expected against actual.

Adding a new set of model tests:

- On first addition of a new set of model tests a `snapshots` directory must be created in the same directory as the model and test files.

- When the tests are run for the first time, a `-temp` snapshot yaml file will be created for each resource tested, under a different directory for each instance (these will be created automatically). 

- It is up to the user to remove the `-temp` from the filename if the form of the resource is deemed correct and this will be the yaml to be compared against for future test runs.

Existing model test runs:

- If a test run fails, a `-temp` snapshot file will remain and the user must observe whether the changes are desired and update the non-temp version of the file as necessary, finally deleting the `-temp` file.

- If a test run passes, the `-temp` files will be automatically deleted. 

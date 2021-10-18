# Resources Helpers Examples
The `resources` package provides helper functions that can be shared across operators in IAF

## Method and Examples

### Reconcile

#### Usage

The `Reconcile` shared method handles triggering Kubernetes events for creating, updating and deleting Kubernetes objects - it can be used like so:
```
resources.Reconcile(namespacedName, resource)
```
Where the `namespacedName` is the name of the resource and `resource` is the desired state. The `Reconcile` function returns three values of types: `ctrl.Result{}`, `bool` and `err` (if any). 

1. The first return value returns the result from the function.
2. The second return value will be `true` if an error was encountered and `false` otherwise. 
3. The last return value will be an `err` if an error was encountered or `nill` otherwise.

#### Reconciling new resource types
To reconcile a type of resource not available by default you can create your own resource following the `Reconcileable` interface which requires your struct has the following functions:
```
type Reconcileable interface {
	metav1.Object
	ShouldUpdate(runtime.Object) (bool, runtime.Object)
	GetResource() runtime.Object
	ResourceKind() string
	NewResourceInstance() runtime.Object
	ResourceIsNil() bool
}
```
Your resource can then be used with resources.Reconcile() the same as any other resources.

#### Overriding the behaviour for the return variables
The behaviour for the second return value can be overridden to return only `true` by passing a third `resources.SetExitOnChange` parameter to the `Reconcile` function like so:
```
// this will always return true
resources.Reconcile(namespacedName, desired, resources.SetExitOnChange)
```
It is not recommended to add `resources.SetExitOnChange` as it is very difficult to verify if a change is going to occur. It requires in some cases (for non simple resources such as `statefulsets`) very complex `ShouldUpdate` functions to be written, for no real benefit. It is kept for legacy reasons more than anything.

#### A note on Kubernetes resources
Calling the `Reconcile` function may not have any effect on Kubernetes resources if they are already in their desired state - the actual change of resources is handled inside Kubernetes only if the current resource differs. 

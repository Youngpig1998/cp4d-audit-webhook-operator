## Kubernetes Operator-SKD初实践分享

#### Operator-SDK简介🍺

Operator-SDK是Operator Framework的组件之一，主要用来编写Kubernetes的一些插件，让我们能够更方便地操作Kubernetes。

#### 实践期望🍻

在我们准备部署一个服务的时候，可能会考虑到诸多因素，例如最简单的一个Web应用也需要一个Deployment、一个Service，才能够满足我们的基本需求。抑或是你需要部署后端，涉及数据库，存储卷，再或者是期望Ingress，流量管控，那么部署起来就更复杂了，可能需要写很多YAML文件，即便你有相应的模版，也需要拿出一定的时间才能达到自己觉得满意的状态。

流程越多越容易出现失误，失误越多越耽误时间，时间耽误多了效率就降下来了。所以我们有时候可能会希望，如果有一个Kind，它就叫Nginx或者是MySQL、Redis等，自己输入相应的参数进去就能够达到自己的预期，那将会是一件非常美好的事情。

那么，如果你有这样的想法，Operator-SDK就可以帮助到你！

附：Operator-SDK官网链接：

[Operator-SDK]: https://sdk.operatorframework.io/



#### 环境清单罗列🧾

|      配置项      |     具体配置     |
| :--------------: | :--------------: |
|     操作系统     |      centos      |
|    Golang版本    |       1.15       |
|  Kubernetes版本  | Openshift v4.8.2 |
| Operator-SDK版本 |      v0.1.8      |

#### 环境部署🌲

###### Golang

略

###### Kubernetes/Openshift

https://github.com/Youngpig1998/KuberneteCluster-built

###### Operator-SDK

直接去官方github官网release下载操作系统对应的版本

我们输入以下命令可以查看operator-sdk版本

```bash
mv operator-sdk_linux_amd64 operator-sdk
chmod 744 operator-sdk
mv operator-sdk /usr/bin


operator-sdk version
```

**PS: Golang版本需要与operator-sdk版本中的go版本一致**

#### 预期目标🌟

```yaml
apiVersion: audit.watson.ibm.com/v1beta1
kind: AuditWebhook
metadata:
  labels:
    control-plane: controller-manager
    app.kubernetes.io/instance: ibm-auditwebhook-operator
    app.kubernetes.io/managed-by: ibm-auditwebhook-operator
    app.kubernetes.io/name: ibm-auditwebhook-operator
  name: auditwebhook-sample
spec:
  # Add fields here
  dockerRegistryPrefix: "cp.stg.icr.io/cp"
  imagePullSecrets:
    - name: "cp.stg.icr.io"
```

我们希望通过这么一个简单的yaml文件就能够部署完毕一个简单的AuditWebhook服务

#### 具体操作💦

创建一个目录，在该目录下新建项目

```bash
mkdir cp4d-audit-webhook-operator && cd cp4d-audit-webhook-operator


operator-sdk init --domain watson.ibm.com  --repo github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator
cd mysql
```

###### 注意⚠️：

此处可能因为科学上网的原因，大家在init的时候会卡住，此处推荐大家设置GOPROXY

```bash
export GO111MODULE=on

export GOPROXY=https://goproxy.cn
```

通过以下命令查看是否修改成功

```sh
go env|grep GOPROXY
```

这个过程中会生成很多文件

现在我们需要首先声明我们的自定义Kind的结构模式

```shell
operator-sdk create api --group audit  --version v1beta1 --kind AuditWebhook --resource --controller
```

由于我们上方所期望的yaml结构中包含dockerRegistryPrefix和imagePullSecrets两个属性，所以我们需要对它们进行声明并将逻辑实现

打开api/v1beta1/auditwebhook_types.go

在AuditWebhookSpec中 我们将具体写上我们期望的属性

```go
type AuditWebhookSpec struct {
	// The mirror image corresponding to the business service, including the name: tag
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// The mirror image corresponding to the business service, including the dockerregistryprefix
	DockerRegistryPrefix string `json:"dockerRegistryPrefix"`
}
```

```go
type AuditWebhookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}
```

需要引入的包 此处不给大家罗列 因为Goland会帮大家自行添加 同时也可以参考本次分享的源码

通过以下这个指令 帮我们生成了相应属性所依赖的部分代码 具体细节本次分享中不必特别关注，可以参考本项目中的Makefile文件

```bash
make generate
```

其实上述命令就相当于是现在本地编译一遍，看看缺少哪些相关的依赖包，再自己使用go get命令下载即可

我们新建工具包文件夹internal/operator

```bash
mkdir internal/operator
mkdir iaw-shared-helpers/pkg

touch internal/operator/resources.go
```

在文件夹internal/operator/resources.go中，我们将把对于Deployment、Service、Secret、Issuer、Certificate、NetworkPolicy等等资源的逻辑具体实现。在iaw-shared-helpers/pkg中，我们也创建了相关的辅助函数，具体细节可参考本次分享的源码。



通过执行create api命令，operator sdk会帮我们生成controllers文件夹，我们只需在里面的 [kind]_controller.go中实现我们自己的Reconcile逻辑即可。

controllers/auditwebhook_controller.go的Reconcile函数

```go
func (r *AuditWebhookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {


	log := r.Log.WithValues("auditwebhook", req.NamespacedName)


	log.Info("1. start reconcile logic")
	// Instantialize the data structure
	instance := &auditv1beta1.AuditWebhook{}

	//First,query the webhook instance
	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {
		// If there is no instance, an empty result is returned, so that the Reconcile method will not be called immediately
		if errors.IsNotFound(err) {
			log.Info("Instance not found, maybe removed")
			return reconcile.Result{}, nil
		}
		log.Error(err, "query action happens error")
		// Return error message
		return ctrl.Result{}, err
	}



	//Set the bootstrapClient's owner value as the webhook,so the resources we create then will be set reference to the webhook
	//when the webhook cr is deleted,the resources(such as deployment.configmap,issuer...) we create will be deleted too
	bootstrapClient, err := bootstrap.NewClient(r.Config,r.Scheme, controllerManagerName,instance)
	if err != nil {
		log.Error(err, "failed to initialise bootstrap client")
		return ctrl.Result{}, err
	}


	//operandRequestName, operandRequest := operator.OperandRequest()
	//
	//done := bootstrapClient.InitialiseCommonServices(operandRequestName, operandRequest)
	//err = <-done
	//if err != nil {
	//	log.Error(err, "failed to initialise common services")
	//	return ctrl.Result{}, err
	//}


	//We create networkpolicy first
	networkPolicyName, networkPolicy := operator.NetworkPolicy()
	err = bootstrapClient.CreateResource(networkPolicyName, networkPolicy)
	if err != nil {
		log.Error(err, "failed to create operator NetworkPolicy", "Name", networkPolicyName)
		return ctrl.Result{}, err
	}

	issuerName, issuer := operator.Issuer()
	err = bootstrapClient.CreateResource(issuerName, issuer)
	if err != nil {
		log.Error(err, "failed to create operator issuer", "Name", issuerName)
		return ctrl.Result{}, err
	}

	certificateName, certificate := operator.Certificate(instance)
	err = bootstrapClient.CreateResource(certificateName, certificate)
	if err != nil {
		log.Error(err, "failed to create operator certificate", "Name", certificateName)
		return ctrl.Result{}, err
	}


	secretData,err := r.getSecretData(ctx,req)
	if err != nil {
		log.Error(err, "failed to get  secret's data")
		return ctrl.Result{}, err
	}

	secretName, secret := operator.Secret(secretData)
	err = bootstrapClient.CreateResource(secretName, secret)
	if err != nil {
		log.Error(err, "failed to create operator secret", "Name", secretName)
		return ctrl.Result{}, err
	}


	configMapName, configMap := operator.ConfigMap(instance)
	err = bootstrapClient.CreateResource(configMapName, configMap)
	if err != nil {
		log.Error(err, "failed to create operator configMap", "Name", configMapName)
		return ctrl.Result{}, err
	}


	serviceName, service := operator.Service()
	err = bootstrapClient.CreateResource(serviceName, service)
	if err != nil {
		log.Error(err, "failed to create operator service", "Name", serviceName)
		return ctrl.Result{}, err
	}

	deploymentName, deployment := operator.Deployment(instance)
	err = bootstrapClient.CreateResource(deploymentName, deployment)
	if err != nil {
		log.Error(err, "failed to create operator Deployment", "Name", deploymentName)
		return ctrl.Result{}, err
	}

	mcName, mc := operator.MutatingWebhookConfiguration(instance)
	err = bootstrapClient.CreateResource(mcName, mc)
	if err != nil {
		log.Error(err, "failed to create operator Mc", "Name", mcName)
		return ctrl.Result{}, err
	}


	return ctrl.Result{}, nil
}
```

在编写好所有的代码后我们再执行以下命令让operator sdk生成相关的config文件夹下的yaml文件。

```shell
make manifests
```

### **PS：每次修改了types.go后，需要执行make generate命令，修改了controller.go后需要执行make manifests命令**



执行好上述命令后便可以根据需要修改Dockerfile并且制作镜像，部署operator

```bash
docker build -t {IMAGE_NAME} .
docker push
make deploy
```



#### 源码地址💲

https://github.com/Youngpig1998/cp4d-audit-webhook-operator

#### 参考博客🙏

https://www.qikqiak.com/

https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/

https://xinchen.blog.csdn.net/article/details/113089414

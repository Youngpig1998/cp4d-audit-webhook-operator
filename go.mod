module github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator

go 1.16

require (
	github.com/IBM/operand-deployment-lifecycle-manager v1.7.0
	github.com/go-logr/logr v0.4.0
	github.com/imdario/mergo v0.3.12
	github.com/jetstack/cert-manager v1.3.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/controller-runtime v0.9.2
)

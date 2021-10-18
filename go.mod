module github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator

go 1.15

require (
	github.com/IBM/operand-deployment-lifecycle-manager v1.7.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/imdario/mergo v0.3.10
	github.com/jetstack/cert-manager v1.3.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/r3labs/diff/v2 v2.13.6
	go.uber.org/zap v1.15.0
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/jetstack/cert-manager => github.com/jetstack/cert-manager v0.10.0

replace k8s.io/api => k8s.io/api v0.19.3

replace k8s.io/client-go => k8s.io/client-go v0.19.3

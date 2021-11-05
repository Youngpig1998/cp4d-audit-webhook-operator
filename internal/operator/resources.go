package operator

import (
	odlmv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	auditv1beta1 "github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/api/v1beta1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/certificates"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/configmaps"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/deployments"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/issuers"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/mutatingwebhookconfigurations"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/networkpolicies"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/secrets"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/services"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkpolicy "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"strings"
)



const (
	APP_NAME = "audit-webhook"
	//Request CPU resource of a single pod
	CPU_REQUEST = "300m"
	//Request CPU resource limit of a single pod
	CPU_LIMIT = "500m"
	//Request Memory resource of a single pod
	MEM_REQUEST = "100Mi"
	//Request Memory resource limit of a single pod
	MEM_LIMIT = "200Mi"
)


var (
	operandRequestName = "ibm-certmanager-operators"
	networkPolicyName = "audit-webhook-networkpolicy"
	issuerName = "selfsigned-issuer"
	certificateName = "serving-cert"
	secretName = "audit-webhook-tls-secret"
	configMapName = "audit-webhook-configmap"
	serviceName = "audit-webhook-service"
	mutatingwebhookConfigurationName = "audit-webhook-config"
	deploymentName = "audit-webhook-server"
	commonservices = []string{"ibm-cert-manager-operator"}
)





func OperandRequest() (string, *odlmv1alpha1.OperandRequest) {
	operands := []odlmv1alpha1.Operand{}
	for _, commonService := range commonservices {
		operands = append(operands, odlmv1alpha1.Operand{Name: commonService})
	}
	operandRequest := &odlmv1alpha1.OperandRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:   operandRequestName,
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Spec: odlmv1alpha1.OperandRequestSpec{
			Requests: []odlmv1alpha1.Request{
				{
					Registry:          "common-service",
					RegistryNamespace: "ibm-common-services",
					Operands:          operands,
				},
			},
		},
	}
	return operandRequestName, operandRequest
}


func NetworkPolicy() (string, resources.Reconcileable) {

	netProtocol := corev1.Protocol("TCP")

	networkPolicyIngress := []networkpolicy.NetworkPolicyIngressRule{
		{
			Ports: []networkpolicy.NetworkPolicyPort{
				{
					Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 8081},
					Protocol: &netProtocol,
				},
			},
			From: []networkpolicy.NetworkPolicyPeer{},
		},
	}

	networkPolicy := &networkpolicy.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   networkPolicyName,
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
			//Namespace: webHook.Namespace,
		},
		Spec: networkpolicy.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":   "audit-webhook",
				},
			},
			Ingress: networkPolicyIngress,
			PolicyTypes: []networkpolicy.PolicyType{"Ingress"},
		},
	}

	return networkPolicyName,networkpolicies.From(networkPolicy)
}

func Issuer() (string, resources.Reconcileable){

	issuer := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:       issuerName,
			Labels:		map[string]string{
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Spec:       certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				SelfSigned: &certmanagerv1.SelfSignedIssuer{},
			},
		},
	}

	return  issuerName,issuers.From(issuer)
}




func Certificate(webHook *auditv1beta1.AuditWebhook) (string, resources.Reconcileable){

	const dnsNameFront = "audit-webhook-service."
	const dnsNameBack = ".svc"
	var dnsName = dnsNameFront + webHook.Namespace + dnsNameBack
	certificate := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       certificateName,
			Labels:		map[string]string{
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Spec:       certmanagerv1.CertificateSpec{
			DNSNames:     []string{
				dnsName,
			},
			SecretName:   secretName,
			IssuerRef:    cmmeta.ObjectReference{
				Name:  issuerName,
				Kind:  "Issuer",
			},
		},
	}

	return certificateName,certificates.From(certificate)
}


func Secret(secretData map[string][]byte) (string, resources.Reconcileable){

	secretType := corev1.SecretTypeTLS

	secret := &corev1.Secret{
		Type: secretType,
		ObjectMeta: metav1.ObjectMeta{
			//Namespace: webHook.Namespace,
			Name:      secretName,
			Labels: map[string]string{
				"app": APP_NAME,
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by":  "ibm-auditwebhook-operator",
				"app.kubernetes.io/name": "ibm-auditwebhook-operator",
			},
		},
		Data: secretData,
	}

	return secretName,secrets.From(secret)
}


func ConfigMap(webHook *auditv1beta1.AuditWebhook) (string, resources.Reconcileable) {

	//imageName := "cp.stg.icr.io/cp/opencontent-fluentd:ruby-ubi"
	imageName := "docker.io/youngpig/fluentd:latest"
	if len(strings.TrimSpace(webHook.Spec.DockerRegistryPrefix)) > 0 {
		//imageName = webHook.Spec.DockerRegistryPrefix + "/opencontent-fluentd@sha256:d71c70d59540caead90cfb46c83ebafe55787078f73e48bf12558f73b997b17e"
		imageName = webHook.Spec.DockerRegistryPrefix + "/fluentd:latest"
	}



	volume_patch := "{\"name\":\"internal-tls\",\"secret\":{\"secretName\":\"internal-tls\",\"defaultMode\":420}}"
	//container_patch := "{\"name\": \"sidecar\", 	\"image\": \"" + imageName + "\", 	\"securityContext\": { 		\"runAsNonRoot\": true 	}, 	\"resources\": { 		\"requests\": { 			\"memory\": \"100Mi\", 			\"cpu\": \"100m\" 		}, 		\"limits\": { 			\"memory\": \"250Mi\", 			\"cpu\": \"250m\" 		} 	}, 	\"imagePullPolicy\": \"Always\", 	\"volumeMounts\": [{ 		\"name\": \"varlog\", 		\"mountPath\": \"/var/log\" 	}, { 		\"name\": \"internal-tls\", 		\"mountPath\": \"/etc/internal-tls\" 	}], 	\"env\": [{ 		\"name\": \"NS_DOMAIN\", 		\"value\": \"https://zen-audit-svc." + webHook.Namespace + ":9880/records\" 	}] }"
	container_patch := "{\"name\": \"sidecar\", 	\"image\": \"" + imageName + "\", 	\"securityContext\": { 		\"runAsNonRoot\": true 	}, 	\"resources\": { 		\"requests\": { 			\"memory\": \"100Mi\", 			\"cpu\": \"100m\" 		}, 		\"limits\": { 			\"memory\": \"250Mi\", 			\"cpu\": \"250m\" 		} 	}, 	\"imagePullPolicy\": \"IfNotPresent\", 	\"volumeMounts\": [{ 		\"name\": \"varlog\", 		\"mountPath\": \"/var/log\" 	}], 	\"env\": [{ 		\"name\": \"NS_DOMAIN\", 		\"value\": \"https://zen-audit-svc." + webHook.Namespace + ":9880/records\" 	}] }"
	// Instantialize the data structure
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			//Namespace: webHook.Namespace,
			Name:      configMapName,
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Data: map[string]string{
			"volume_patch":    volume_patch,
			"container_patch": container_patch,
		},
	}

	return  configMapName,configmaps.From(configmap)
}




func Service() (string, resources.Reconcileable) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			//Namespace: webHook.Namespace,
			Name:      serviceName,
			Labels: map[string]string{
				"app":                          APP_NAME,
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port: 443,
				TargetPort: intstr.IntOrString{
					IntVal: 8081,
					StrVal: "8081",
				},
				Protocol: corev1.ProtocolTCP,
			},
			},
			Selector: map[string]string{
				"app": APP_NAME,
			},
		},
	}

	return serviceName,services.From(service)
}

func MutatingWebhookConfiguration(webHook *auditv1beta1.AuditWebhook) (string, resources.Reconcileable) {

	path := "/add-sidecar"
	certpath := webHook.Namespace + "/serving-cert"

	matchPolicy := new(admissionregistrationv1.MatchPolicyType)
	*matchPolicy = admissionregistrationv1.Equivalent

	sideEffects := new(admissionregistrationv1.SideEffectClass)
	*sideEffects = "None"

	failurePolicy := new(admissionregistrationv1.FailurePolicyType)
	*failurePolicy = admissionregistrationv1.Ignore

	scope := new(admissionregistrationv1.ScopeType)
	*scope = admissionregistrationv1.NamespacedScope

	mc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			//Namespace: webHook.Namespace,
			Name:      mutatingwebhookConfigurationName,
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": certpath + "",
			},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:        "audit.watson.org",
			MatchPolicy: matchPolicy,
			SideEffects: sideEffects,
			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			ObjectSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cp4d-audit": "yes",
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{
				Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
					Scope:       scope,
				},
			},
			},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Name:      serviceName,
					Namespace: webHook.Namespace,
					Path:      &path,
					Port:      pointer.Int32Ptr(443),
				},
				//CABundle: stringtoslicebyte(webHook.Spec.CaBundle),
				//CABundle: sDecForCABundle,
			},
			FailurePolicy: failurePolicy,
		},
		},
	}

	return mutatingwebhookConfigurationName,mutatingwebhookconfigurations.From(mc)
}

func Deployment(webHook *auditv1beta1.AuditWebhook) (string, resources.Reconcileable) {

	isRunAsRoot := false
	pIsRunAsRoot := &isRunAsRoot //bool pointer

	isPrivileged := false
	pIsPrivileged := &isPrivileged

	isAllowPrivilegeEscalation := false
	pIsAllowPrivilegeEscalation := &isAllowPrivilegeEscalation

	isReadOnlyRootFilesystem := false
	pIsReadOnlyRootFilesystem := &isReadOnlyRootFilesystem

	//imageName := "cp.stg.icr.io/cp/opencontent-audit-webhook@sha256:0d8c98939b31aa261d09b9f38f834cf524007cf6af1a6e02198bee115d04f918"
	imageName := "docker.io/youngpig/audit-webhook:latest"
	if len(strings.TrimSpace(webHook.Spec.DockerRegistryPrefix)) > 0 {
		//imageName = webHook.Spec.DockerRegistryPrefix + "/opencontent-audit-webhook:v0.1.0"
		imageName = webHook.Spec.DockerRegistryPrefix + "/audit-webhook:latest"
	}


	// Instantialize the data structure
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			//Namespace: webHook.Namespace,
			Name:      deploymentName,
			Labels: map[string]string{
				"app":                          APP_NAME,
				"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
				"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
				"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			// The replica is computed
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": APP_NAME,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                          APP_NAME,
						"app.kubernetes.io/instance":   "ibm-auditwebhook-operator",
						"app.kubernetes.io/managed-by": "ibm-auditwebhook-operator",
						"app.kubernetes.io/name":       "ibm-auditwebhook-operator",
					},
					Annotations: map[string]string{
						"productName":    "ibm-auditwebhook",
						"productID":      "96808888679886798867988679886798",
						"productVersion": "1.0.0",
						"productMetric":  "VIRTUAL_PROCESSOR_CORE",
						"cloudpakId":     "96808888679886798867988679886798",
						"cloudpakName":   "Cloud Pak Open",
					},
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											corev1.NodeSelectorRequirement{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values: []string{
													"amd64",
												},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								corev1.PreferredSchedulingTerm{
									Weight: 3,
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											corev1.NodeSelectorRequirement{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values: []string{
													"amd64",
												},
											},
										},
									},
								},
							},
						},
					},
					HostNetwork:      false,
					HostIPC:          false,
					HostPID:          false,
					ImagePullSecrets: webHook.Spec.ImagePullSecrets,
					Containers: []corev1.Container{{
						Image: imageName,
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 8081,
									},
									Scheme: "HTTPS",
								},
							},
							InitialDelaySeconds: 15,
							PeriodSeconds:       20,
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/readyz",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 8081,
										StrVal: "8081",
									},
									Scheme: "HTTPS",
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       10,
						},
						ImagePullPolicy: "IfNotPresent",
						Name:            APP_NAME,
						Command:         []string{"/audit-webhook"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8081,
						}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse(CPU_REQUEST),
								"memory": resource.MustParse(MEM_REQUEST),
							},
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse(CPU_LIMIT),
								"memory": resource.MustParse(MEM_LIMIT),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
							RunAsNonRoot:             pIsRunAsRoot,
							AllowPrivilegeEscalation: pIsAllowPrivilegeEscalation,
							ReadOnlyRootFilesystem:   pIsReadOnlyRootFilesystem,
							Privileged:               pIsPrivileged,
						},
						Env: []corev1.EnvVar{
							{
								Name: "VOLUME_PATCH",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "audit-webhook-configmap",
										},
										Key: "volume_patch",
									},
								},
							},
							{
								Name: "CONTAINER_PATCH",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "audit-webhook-configmap",
										},
										Key: "container_patch",
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/certs",
								Name:      "certs",
								ReadOnly:  false,
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "audit-webhook-tls-secret",
								},
							},
						},
					},
				},
			},
		},
	}



	return deploymentName,deployments.From(deployment)
}

















































































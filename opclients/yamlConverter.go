package opclients

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	//apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
)

var path = ""

func createK8sObjects(fileNames []string) {
	for _, fn := range fileNames {
		yfile, err := ioutil.ReadFile(path + "/" + fn)

		if err != nil {

			log.Fatal(err)
		}

		yfilestring := string(yfile)

		parts := strings.Split(yfilestring, "\n---\n")

		for _, part := range parts {
			fmt.Println("--------------------------")
			toCodify([]byte(part))
		}
	}
}

func getAllYamlFileNamesInPath(path string, recursive bool) ([]string, error) {
	files := []string{}

	items, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("Error caught while getting files from path ", err)
		return []string{}, fmt.Errorf("Error caught while getting files %v", err)
	}

	for _, f := range items {
		if f.IsDir() && recursive {
			subFiles, err := getAllYamlFileNamesInPath(path+"/"+f.Name(), recursive)
			if err != nil {
				fmt.Println("Error caught while getting files from path ", err)
				return []string{}, fmt.Errorf("Error caught while getting files %v", err)
			}
			files = append(files, subFiles...)
		} else if strings.HasSuffix(f.Name(), ".yaml") {
			files = append(files, f.Name())
		}
	}
	return files, nil
}

func toCodify(raw []byte) {
	serializer := scheme.Codecs.UniversalDeserializer()
	var decoded runtime.Object
	decoded, _, _ = serializer.Decode([]byte(raw), nil, nil)

	switch x := decoded.(type) {
	case *appsv1.Deployment:
		fmt.Printf("Deploymenty found %v.\n", x)
		applyDeployment(x)
	// case *appsv1.DaemonSet:
	// 	fmt.Printf("DeamonSet found %v.\n", x)
	// 	applyDeamonset(x)
	// case *corev1.ConfigMap:
	// 	fmt.Printf("ConfigMap found %v.\n", x)
	// 	applyConfigMap(x)
	case *corev1.Service:
		fmt.Printf("Service found %v.\n", x)
		applyService(x)
	case *rbacv1.Role:
		fmt.Printf("Role found %v.\n", x)
		applyRole(x)
	case *rbacv1.ClusterRole:
		fmt.Printf("Cluster Role found %v.\n", x)
		applyClusterRole(x)
	case *rbacv1.RoleBinding:
		fmt.Printf("Role Binding found %v.\n", x)
		applyRoleBinding(x)
	case *rbacv1.ClusterRoleBinding:
		fmt.Printf("Cluster Role Binding found %v.\n", x)
		applyClusterRoleBinding(x)
	case *corev1.ServiceAccount:
		fmt.Printf("ServiceAccount found %v.\n", x)
		applyServiceAccount(x)
	// case *apiextensionsv1.CustomResourceDefinition:
	// 	fmt.Printf("CRD found %v.\n", x)
	// 	applyCrds(x)
	case *corev1.Namespace:
		fmt.Printf("Namespace found %v.\n", x)
		applyNamespace(x)
		break
	default:
		//return nil, fmt.Errorf("missing support for type: %s", x.GetObjectKind().GroupVersionKind().Kind)
	}
}

// func applyCrds(obj *apiextensionsv1.CustomResourceDefinition) {
// 	clientset := connectToK8s()
// 	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
// 	clientset.apiextensionsv1().CustomResourceDefinition().Create(context.TODO(), obj, metav1.CreateOptions{})
// }

// func applyConfigMap(obj *corev1.ConfigMap) {
// 	clientset := connectToK8s()
// 	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
// 	clientset.CoreV1().ConfigMap().Create(context.TODO(), obj, metav1.CreateOptions{})
// }

// func applyDeamonset(obj *corev1.DaemonSet) {
// 	clientset := connectToK8s()
// 	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
// 	clientset.CoreV1().DaemonSet().Create(context.TODO(), obj, metav1.CreateOptions{})
// }

func applyNamespace(obj *corev1.Namespace) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	clientset.CoreV1().Namespaces().Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyDeployment(obj *appsv1.Deployment) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	obj.Status = appsv1.DeploymentStatus{}
	clientset.AppsV1().Deployments("default").Create(context.TODO(), obj, metav1.CreateOptions{})
}

func cleanObjectMeta(m metav1.ObjectMeta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       m.Name,
		Namespace:                  m.Namespace,
		Labels:                     m.Labels,
		Annotations:                m.Annotations,
		ResourceVersion:            m.ResourceVersion,
		Finalizers:                 m.Finalizers,
		Generation:                 m.Generation,
		GenerateName:               m.GenerateName,
		UID:                        m.UID,
		ManagedFields:              m.ManagedFields,
		OwnerReferences:            m.OwnerReferences,
		DeletionGracePeriodSeconds: m.DeletionGracePeriodSeconds,
	}
}

func applyServiceAccount(obj *corev1.ServiceAccount) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	//obj.Status = corev1.ServiceStatus{}
	clientset.CoreV1().ServiceAccounts("default").Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyService(obj *corev1.Service) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	obj.Status = corev1.ServiceStatus{}
	clientset.CoreV1().Services("default").Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyClusterRole(obj *rbacv1.ClusterRole) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	clientset.RbacV1().ClusterRoles().Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyClusterRoleBinding(obj *rbacv1.ClusterRoleBinding) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyRole(obj *rbacv1.Role) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	clientset.RbacV1().Roles("default").Create(context.TODO(), obj, metav1.CreateOptions{})
}

func applyRoleBinding(obj *rbacv1.RoleBinding) {
	clientset := connectToK8s()
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	clientset.RbacV1().RoleBindings("default").Create(context.TODO(), obj, metav1.CreateOptions{})
}

func connectToK8s() *kubernetes.Clientset {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "/root"
	}

	configPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Panicln("failed to create K8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln("Failed to create K8s clientset")
	}

	return clientset
}

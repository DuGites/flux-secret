package stackoverdrivesecret

import (
	"context"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"os"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	v1beta1 "github.com/flux-secret/pkg/apis/fluxsecret/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

const (
	// StackOverDrive ... The Secret Type Kind
	CustomResourceKind = "StackOverDriveSecret"
)

// Add creates a new StackOverDriveSecret Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStackOverDriveSecret{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("stackoverdrivesecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &v1beta1.StackOverDriveSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1beta1.StackOverDriveSecret{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStackOverDriveSecret{}

// ReconcileStackOverDriveSecret reconciles a StackOverDriveSecret object
type ReconcileStackOverDriveSecret struct {
	client.Client
	scheme *runtime.Scheme
}

// asOwner returns an owner reference set as the Lrs instance
func asOwner(uid types.UID, customResourceInstanceName string) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion:         v1beta1.SchemeGroupVersion.String(),
		Kind:               CustomResourceKind,
		Name:               customResourceInstanceName,
		UID:                uid,
		Controller:         &trueVar,
		BlockOwnerDeletion: &trueVar,
	}
}

func decryptSecretData(region string, secretData string) ([]byte, error) {

	decoded, err := base64.StdEncoding.DecodeString(secretData)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Error(err, "Session creation failed")
		return nil, err
	}
	kmsClient := kms.New(sess)
	decryptRsp, err := kmsClient.Decrypt(&kms.DecryptInput{
		CiphertextBlob: []byte(decoded),
	})
	if err != nil {
		log.Error(err, "KMS Decryption Failed")
		return nil, err
	}
	log.Info("Decoded Secret Data", decryptRsp.Plaintext)
	return decryptRsp.Plaintext, nil
}

// Reconcile ...
// +kubebuilder:rbac:groups="",resources=stackoverdrivesecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fluxsecret.stackoverdrive.io,resources=stackoverdrivesecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fluxsecret.stackoverdrive.io,resources=stackoverdrivesecrets/status,verbs=get;update;patch
func (r *ReconcileStackOverDriveSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the StackOverDriveSecret instance
	instance := &v1beta1.StackOverDriveSecret{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	log.Info("The region is", "Region", region)
	for secretName, secretValue := range instance.Spec.EncryptedData {
		decryptedSecret, err := decryptSecretData(region, secretValue)
		if err != nil {
			log.Error(err, "Decryption of KMS Secret Failed")
		}
		secretObj := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            secretName,
				Namespace:       instance.Namespace,
				OwnerReferences: []metav1.OwnerReference{asOwner(instance.UID, instance.Name)},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{secretName: decryptedSecret},
		}
		if err := controllerutil.SetControllerReference(instance, secretObj, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		// Check if the Secret already exists
		found := &corev1.Secret{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: secretObj.Name, Namespace: secretObj.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating Secret", "namespace", secretObj.Namespace, "name", secretObj.Name)
			err = r.Create(context.TODO(), secretObj)
			return reconcile.Result{}, err
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Update the found object and write the result back if there are any changes
		if !reflect.DeepEqual(secretObj, found) {
			found = secretObj
			log.Info("Updating Secret", "namespace", secretObj.Namespace, "name", secretObj.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}
	return reconcile.Result{}, nil

}

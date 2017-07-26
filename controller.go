package main

import (
	"context"
	"os/exec"

	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// AppController is a controller for application resource.
type AppController struct {
	AppClient *rest.RESTClient
	AppScheme *runtime.Scheme
}

// Run starts an App resource controller
func (app *AppController) Run(ctx context.Context) error {

	// Watch for app objects
	err := app.watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to register watch for App resource: %v", err)
	}

	<-ctx.Done()
	return ctx.Err()
}

// watch watches for App events and responds accordingly
func (app *AppController) watch(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		app.AppClient,
		"apps", // resource name plural
		apiv1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		source,

		&App{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		0,

		// Define custom resource event handlers
		cache.ResourceEventHandlerFuncs{
			AddFunc: app.onAdd,
			// UpdateFunc: app.onUpdate,
			DeleteFunc: app.onDelete,
		},
	)

	go controller.Run(ctx.Done())

	return nil
}

// OnAdd is called when an object is added.
// OnAdd also seems to apply when the controller is first run and
// custom app resources already exist...
func (app *AppController) onAdd(obj interface{}) {
	appItem := obj.(*App)

	// so we don't try to redeploy pre-existing objects
	// when the controller starts...
	if appItem.Status.State == "" {

		// objects from the store are read-only local cache.
		// make a copy so we can do things with it.
		copyObject, err := app.AppScheme.Copy(appItem)
		if err != nil {
			fmt.Printf("failed to copy app object: %v", err)
		}

		appCopy := copyObject.(*App)

		siteName := appItem.Spec.Name
		siteType := appItem.Spec.Type
		var helmChart string
		var siteNameArg string

		switch siteType {
		case "drupal":
			helmChart = "stable/drupal"
			siteNameArg = ""
		case "wordpress":
			helmChart = "stable/wordpress"
			siteNameArg = "wordpressBlogName=" + siteName
		default:
			appCopy.Status = AppStatus{
				State:   AppStateInvalidType,
				Message: "Invalid App Type provided",
			}
		}

		if appCopy.Status.State != AppStateInvalidType {
			out, err := exec.Command(
				"helm",
				"install",
				"--name="+appItem.ObjectMeta.Name,
				"--set",
				siteNameArg,
				helmChart,
			).CombinedOutput()

			fmt.Println(string(out))

			if err != nil {
				fmt.Println(err)
				appCopy.Status = AppStatus{
					State:   AppStateFailed,
					Message: err.Error() + ": " + string(out),
				}
			} else {
				appCopy.Status = AppStatus{
					State:   AppStateProcessed,
					Message: "App processed successfully",
				}
			}
		}

		err = app.AppClient.Put().
			Name(appItem.ObjectMeta.Name).
			Namespace(appItem.ObjectMeta.Namespace).
			Resource("apps").
			Body(appCopy).
			Do().
			Error()

		if err != nil {
			fmt.Printf("failed to update status: %v", err)
		} else {
			fmt.Printf("Updated status: %v\n", appCopy)
		}
	}
}

// OnUpdate is called when an object is modified. Note that oldObj is the
// last known state of the object-- it is possible that several changes
// were combined together, so you can't use this to see every single
// change. OnUpdate is also called when a re-list happens, and it will
// get called even if nothing changed. This is useful for periodically
// evaluating or syncing something.
// The update of status in onAdd also seems to trigger onUpdate...
func (app *AppController) onUpdate(oldObj, newObj interface{}) {
	oldExample := oldObj.(*App)
	newExample := newObj.(*App)
	fmt.Printf("[CONTROLLER] OnUpdate oldObj: %s\n", oldExample.ObjectMeta.SelfLink)
	fmt.Printf("[CONTROLLER] OnUpdate newObj: %s\n", newExample.ObjectMeta.SelfLink)
}

// OnDelete will get the final state of the item if it is known, otherwise
// it will get an object of type DeletedFinalStateUnknown. This can
// happen if the watch is closed and misses the delete event and we don't
// notice the deletion until the subsequent re-list.
func (app *AppController) onDelete(obj interface{}) {
	appItem := obj.(*App)

	out, err := exec.Command(
		"helm",
		"delete",
		appItem.Spec.Name,
	).CombinedOutput()

	fmt.Println(string(out))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("[CONTROLLER] OnDelete %s\n", appItem.ObjectMeta.SelfLink)
}

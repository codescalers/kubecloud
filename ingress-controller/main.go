package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zosbase/pkg/gridtypes/zos"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Route struct {
	Host     string   `json:"host"`
	Path     string   `json:"path"`
	Backends []string `json:"backends"`
}

type Config struct {
	Routes []Route `json:"routes"`
}

type Controller struct {
	clientset       *kubernetes.Clientset
	informerFactory informers.SharedInformerFactory
	stopCh          chan struct{}
	network         string
	mnemonic        string
	lastConfigHash  string
}

func NewController() (*Controller, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Controller{
		clientset:       clientset,
		informerFactory: informers.NewSharedInformerFactory(clientset, 0),
		stopCh:          make(chan struct{}),
		network:         os.Getenv("THREEFOLD_NETWORK"),
		mnemonic:        os.Getenv("THREEFOLD_MNEMONIC"),
	}, nil
}

func (c *Controller) Start() error {
	ingressInformer := c.informerFactory.Networking().V1().Ingresses().Informer()
	serviceInformer := c.informerFactory.Core().V1().Services().Informer()
	endpointSliceInformer := c.informerFactory.Discovery().V1().EndpointSlices().Informer()

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.handleChange() },
		UpdateFunc: func(oldObj, newObj interface{}) { c.handleChange() },
		DeleteFunc: func(obj interface{}) { c.handleChange() },
	}

	ingressInformer.AddEventHandler(handler)
	serviceInformer.AddEventHandler(handler)
	endpointSliceInformer.AddEventHandler(handler)

	c.informerFactory.Start(c.stopCh)
	c.informerFactory.WaitForCacheSync(c.stopCh)

	klog.Info("Ingress controller started")

	c.handleChange()
	return nil
}

func (c *Controller) Stop() {
	close(c.stopCh)
	klog.Info("Ingress controller shutting down")
}

func (c *Controller) handleChange() {
	routes := c.generateRoutes()
	config := Config{Routes: routes}

	configHash := c.calculateConfigHash(config)
	if configHash == c.lastConfigHash {
		klog.V(2).Info("Configuration unchanged, skipping update")
		return
	}

	klog.Info("Configuration changed, updating...")
	c.lastConfigHash = configHash

	if err := c.printConfig(config); err != nil {
		klog.Errorf("Failed to print config: %v", err)
	}

	if err := c.deployGateway(config); err != nil {
		klog.Errorf("Failed to deploy gateway: %v", err)
	}
}

func (c *Controller) printConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (c *Controller) generateRoutes() []Route {
	ingressInformer := c.informerFactory.Networking().V1().Ingresses().Informer()
	var routes []Route

	ingressList := ingressInformer.GetStore().List()
	for _, obj := range ingressList {
		ing := obj.(*networkingv1.Ingress)
		namespace := ing.Namespace

		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				route := Route{
					Host: rule.Host,
					Path: path.Path,
				}

				backends, err := c.getBackends(namespace, &path)
				if err != nil {
					klog.Warningf("Failed to get backends for %s/%s: %v", namespace, path.Backend.Service.Name, err)
					continue
				}

				route.Backends = backends
				routes = append(routes, route)
			}
		}
	}

	return routes
}

func (c *Controller) getBackends(namespace string, path *networkingv1.HTTPIngressPath) ([]string, error) {
	svcName := path.Backend.Service.Name
	portNum := path.Backend.Service.Port.Number

	// Get the service to check its type
	svc, err := c.clientset.CoreV1().Services(namespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get service %s/%s: %w", namespace, svcName, err)
	}

	// Handle ExternalName services differently
	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		externalName := svc.Spec.ExternalName
		if externalName == "" {
			return nil, fmt.Errorf("ExternalName service %s/%s has no external name configured", namespace, svcName)
		}
		// For ExternalName services, use the external name and port
		backend := fmt.Sprintf("%s:%d", externalName, portNum)
		klog.Infof("Found ExternalName service backend: %s", backend)
		return []string{backend}, nil
	}

	// Check if service has external IPs configured
	if len(svc.Spec.ExternalIPs) > 0 {
		var backends []string
		for _, externalIP := range svc.Spec.ExternalIPs {
			// For services with external IPs, use the service port (not target port)
			servicePort := portNum
			if svc.Spec.Type == corev1.ServiceTypeNodePort {
				// For NodePort services, we can use either the service port or node port
				// Since external ingress controller accesses from outside, use service port with external IP
				servicePort = portNum
			}
			backend := fmt.Sprintf("%s:%d", externalIP, servicePort)
			backends = append(backends, backend)
		}
		klog.Infof("Found service with external IPs: %v", backends)
		return backends, nil
	}

	// For regular services, find endpoints via EndpointSlices
	labelSelector := labels.Set{
		"kubernetes.io/service-name": svcName,
	}.AsSelector()

	endpointSliceList, err := c.clientset.DiscoveryV1().EndpointSlices(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get endpointslices for %s/%s: %w", namespace, svcName, err)
	}

	var backends []string
	for _, endpointSlice := range endpointSliceList.Items {
		for _, endpoint := range endpointSlice.Endpoints {
			if endpoint.Conditions.Ready != nil && !*endpoint.Conditions.Ready {
				continue
			}
			for _, address := range endpoint.Addresses {
				for _, port := range endpointSlice.Ports {
					if port.Port != nil && *port.Port == portNum {
						backends = append(backends, fmt.Sprintf("%s:%d", address, portNum))
					}
				}
			}
		}
	}

	if len(backends) == 0 {
		klog.Infof("No ready endpoints found for %s/%s:%d", namespace, svcName, portNum)
	}

	return backends, nil
}

func (c *Controller) deployGateway(config Config) error {
	if len(config.Routes) == 0 {
		return nil
	}

	if c.network == "" || c.mnemonic == "" {
		klog.Warning("ThreeFold network or mnemonic not configured, skipping gateway deployment")
		return nil
	}

	var zosBackends []zos.Backend
	for _, route := range config.Routes {
		for _, backend := range route.Backends {
			backend = fmt.Sprintf("http://%s", backend)
			zosBackends = append(zosBackends, zos.Backend(backend))
		}
	}

	if len(zosBackends) == 0 {
		return nil
	}

	gateway := workloads.GatewayNameProxy{
		Name:         config.Routes[0].Host,
		Backends:     zosBackends,
		SolutionType: config.Routes[0].Host,
	}

	pluginClient, err := deployer.NewTFPluginClient(
		c.mnemonic,
		deployer.WithNetwork(c.network),
		// deployer.WithSubstrateURL("wss://tfchain.dev.grid.tf/ws"),
	)
	if err != nil {
		return fmt.Errorf("failed to create TF plugin client: %w", err)
	}

	node, err := c.selectNode(pluginClient)
	if err != nil {
		return fmt.Errorf("failed to select node: %w", err)
	}

	gateway.NodeID = node
	if err := pluginClient.GatewayNameDeployer.Deploy(context.TODO(), &gateway); err != nil {
		return fmt.Errorf("failed to deploy gateway on node %d: %w", node, err)
	}

	klog.Infof("Gateway deployed successfully on node %d", node)
	return nil
}

func (c *Controller) selectNode(pluginClient deployer.TFPluginClient) (uint32, error) {
	trueVal := true
	nodes, err := deployer.FilterNodes(
		context.Background(),
		pluginClient,
		types.NodeFilter{Domain: &trueVal, Status: []string{"up"}},
		nil, nil, nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to filter nodes: %w", err)
	}

	if len(nodes) == 0 {
		return 0, fmt.Errorf("no available nodes found")
	}

	return uint32(nodes[0].NodeID), nil
}

func (c *Controller) calculateConfigHash(config Config) string {
	data, err := json.Marshal(config)
	if err != nil {
		klog.Errorf("Failed to marshal config for hashing: %v", err)
		return ""
	}
	return fmt.Sprintf("%x", md5.Sum(data))
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	controller, err := NewController()
	if err != nil {
		klog.Fatalf("Failed to create controller: %v", err)
	}

	if err := controller.Start(); err != nil {
		klog.Fatalf("Failed to start controller: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	controller.Stop()
}

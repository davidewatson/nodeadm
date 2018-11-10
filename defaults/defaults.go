package defaults

import (
	"fmt"

	"github.com/platform9/nodeadm/apis"
	"github.com/platform9/nodeadm/constants"
	"github.com/platform9/nodeadm/workarounds"
	corev1 "k8s.io/api/core/v1"
	kubeadmv1alpha2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha2"
)

// SetInitDefaults sets defaults on the configuration used by init
func SetInitDefaults(config *apis.InitConfiguration) {
	// First set Networking defaults
	SetNetworkingDefaults(&config.Networking)
	// Second set MasterConfiguration.Networking defaults
	SetMasterConfigurationNetworkingDefaultsWithNetworking(config)
	// Third use the remainder of MasterConfiguration defaults
	kubeadmv1alpha2.SetDefaults_MasterConfiguration(&config.MasterConfiguration)
	config.MasterConfiguration.Kind = "MasterConfiguration"
	config.MasterConfiguration.APIVersion = "kubeadm.k8s.io/v1alpha2"
	config.MasterConfiguration.KubernetesVersion = constants.KubernetesVersion
	config.MasterConfiguration.NodeRegistration.Taints = []corev1.Taint{} // empty slice denotes no taints
	addOrAppend(&config.MasterConfiguration.APIServerExtraArgs, "feature-gates", constants.FeatureGates)
	addOrAppend(&config.MasterConfiguration.ControllerManagerExtraArgs, "feature-gates", constants.FeatureGates)
	addOrAppend(&config.MasterConfiguration.SchedulerExtraArgs, "feature-gates", constants.FeatureGates)
}

// SetInitDynamicDefaults sets defaults derived at runtime
func SetInitDynamicDefaults(config *apis.InitConfiguration) error {
	nodeName, err := constants.GetHostnameOverride()
	if err != nil {
		return fmt.Errorf("unable to dervice hostname override: %v", err)
	}
	config.MasterConfiguration.NodeRegistration.Name = nodeName
	return nil
}

// SetJoinDefaults sets defaults on the configuration used by join
func SetJoinDefaults(config *apis.JoinConfiguration) {
}

// SetJoinDynamicDefaults sets defaults derived at runtime
func SetJoinDynamicDefaults(config *apis.JoinConfiguration) error {
	nodeName, err := constants.GetHostnameOverride()
	if err != nil {
		return fmt.Errorf("unable to dervice hostname override: %v", err)
	}
	config.NodeConfiguration.NodeRegistration.Name = nodeName
	return nil
}

// SetNetworkingDefaults sets defaults for the network configuration
func SetNetworkingDefaults(netConfig *apis.Networking) {
	if netConfig.ServiceSubnet == "" {
		netConfig.ServiceSubnet = constants.DefaultServiceSubnet
	}
	if netConfig.DNSDomain == "" {
		netConfig.DNSDomain = constants.DefaultDNSDomain
	}
}

// SetMasterConfigurationNetworkingDefaultsWithNetworking sets defaults with
// values from the top-level network configuration
func SetMasterConfigurationNetworkingDefaultsWithNetworking(config *apis.InitConfiguration) {
	if config.MasterConfiguration.Networking.ServiceSubnet == "" {
		config.MasterConfiguration.Networking.ServiceSubnet = config.Networking.ServiceSubnet
	}
	// If MasterConfigurationNetworking.PodSubnet is provided directly, it takes precedence
	if config.MasterConfiguration.Networking.PodSubnet == "" {
		// Set controller manager extra args directly because of the issue
		// https://github.com/kubernetes/kubeadm/issues/724
		workarounds.SetControllerManagerExtraArgs(config)
	}
	if config.MasterConfiguration.Networking.DNSDomain == "" {
		config.MasterConfiguration.Networking.DNSDomain = config.Networking.DNSDomain
	}
}

func addOrAppend(extraArgs *map[string]string, key string, value string) {
	// Create a new map if it doesn't exist.
	if *extraArgs == nil {
		*extraArgs = make(map[string]string)
	}
	// Add the key with the value if it doesn't exist. Otherwise, append the value
	// to the pre-existing values.
	prevFeatureGates := (*extraArgs)[key]
	if prevFeatureGates == "" {
		(*extraArgs)[key] = value
	} else {
		featureGates := prevFeatureGates + "," + value
		(*extraArgs)[key] = featureGates
	}
}

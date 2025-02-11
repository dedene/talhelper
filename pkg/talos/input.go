package talos

import (
	"github.com/budimanjojo/talhelper/pkg/config"
	"github.com/budimanjojo/talhelper/pkg/decrypt"
	tconfig "github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1/generate"
	"gopkg.in/yaml.v3"
)

// NewClusterInput takes `Talhelper` config and path to encrypted `secretFile` and
// returns Talos `generate.Input`. It also returns an error, if any. 
func NewClusterInput(c *config.TalhelperConfig, secretFile string) (*generate.Input, error) {
	kubernetesVersion := c.GetK8sVersion()

	versionContract, err := tconfig.ParseContractFromVersion(c.GetTalosVersion())
	if err != nil {
		return nil, err
	}

	var secrets *generate.SecretsBundle

	if secretFile != "" {
		decrypted, err := decrypt.DecryptYamlWithSops(secretFile)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(decrypted, &secrets)
		if err != nil {
			return nil, err
		}
		secrets.Clock = generate.NewClock()
	} else {
		secrets, err = NewSecretBundle(generate.NewClock(), generate.WithVersionContract(versionContract))
		if err != nil {
			return nil, err
		}
	}

	opts := parseOptions(c, versionContract)

	input, err := generate.NewInput(c.ClusterName, c.Endpoint, kubernetesVersion, secrets, opts...)
	if err != nil {
		return nil, err
	}

	input.PodNet = c.GetClusterPodNets()
	input.ServiceNet = c.GetClusterSvcNets()
	input.AdditionalMachineCertSANs = c.AdditionalMachineCertSans
	input.AdditionalSubjectAltNames = c.AdditionalApiServerCertSans

	return input, nil
}

// parseOptions takes `TalhelperConfig` and returns slice of Talos `generate.GenOption`
// compatible with the specified `versionContract`.
func parseOptions(c *config.TalhelperConfig, versionContract *tconfig.VersionContract) []generate.GenOption {
	opts := []generate.GenOption{}

	opts = append(opts, generate.WithVersionContract(versionContract))
	opts = append(opts, generate.WithInstallImage(c.GetInstallerURL()))

	if c.AllowSchedulingOnMasters || c.AllowSchedulingOnControlPlanes {
		opts = append(opts, generate.WithAllowSchedulingOnControlPlanes(true))
	}

	if c.CNIConfig.Name != "" {
		opts = append(opts, generate.WithClusterCNIConfig(&v1alpha1.CNIConfig{CNIName: c.CNIConfig.Name, CNIUrls: c.CNIConfig.Urls}))
	}

	if c.Domain != "" {
		opts = append(opts, generate.WithDNSDomain(c.Domain))
	}

	return opts
}

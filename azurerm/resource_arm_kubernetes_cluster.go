package azurerm

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/kubernetes"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmKubernetesClusterCreate,
		Read:   resourceArmKubernetesClusterRead,
		Update: resourceArmKubernetesClusterCreate,
		Delete: resourceArmKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: func(diff *schema.ResourceDiff, v interface{}) error {
			if v, exists := diff.GetOk("network_profile"); exists {
				rawProfiles := v.([]interface{})
				if len(rawProfiles) == 0 {
					return nil
				}

				// then ensure the conditionally-required fields are set
				profile := rawProfiles[0].(map[string]interface{})
				networkPlugin := profile["network_plugin"].(string)

				if networkPlugin != "kubenet" && networkPlugin != "azure" {
					return nil
				}

				dockerBridgeCidr := profile["docker_bridge_cidr"].(string)
				dnsServiceIP := profile["dns_service_ip"].(string)
				serviceCidr := profile["service_cidr"].(string)

				// All empty values.
				if dockerBridgeCidr == "" && dnsServiceIP == "" && serviceCidr == "" {
					return nil
				}

				// All set values.
				if dockerBridgeCidr != "" && dnsServiceIP != "" && serviceCidr != "" {
					return nil
				}

				return fmt.Errorf("`docker_bridge_cidr`, `dns_service_ip` and `service_cidr` should all be empty or all should be set.")
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"dns_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kubernetes_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"node_resource_group": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kube_config": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"client_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"kube_config_raw": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"linux_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ssh_key": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,

							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_data": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},

			"agent_pool_profile": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateKubernetesClusterAgentPoolName(),
						},

						"count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 50),
						},

						// TODO: remove this field in the next major version
						"dns_prefix": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "This field has been removed by Azure",
						},

						"fqdn": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: "This field has been deprecated. Use the parent `fqdn` instead",
						},

						"vm_size": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"os_disk_size_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"vnet_subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"os_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  containerservice.Linux,
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Linux),
								string(containerservice.Windows),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"max_pods": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},

			"service_principal": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"client_secret": {
							Type:      schema.TypeString,
							ForceNew:  true,
							Required:  true,
							Sensitive: true,
						},
					},
				},
				Set: resourceAzureRMKubernetesClusterServicePrincipalProfileHash,
			},

			"network_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_plugin": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(containerservice.Azure),
								string(containerservice.Kubenet),
							}, false),
						},

						"dns_service_ip": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"docker_bridge_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"pod_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"service_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},

			"addon_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_application_routing": {
							Type:     schema.TypeList,
							MaxItems: 1,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										ForceNew: true,
										Required: true,
									},
									"http_application_routing_zone_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"oms_agent": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"log_analytics_workspace_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmKubernetesClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	kubernetesClustersClient := client.kubernetesClustersClient

	log.Printf("[INFO] preparing arguments for Azure ARM AKS managed cluster creation.")

	resGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)
	location := azureRMNormalizeLocation(d.Get("location").(string))
	dnsPrefix := d.Get("dns_prefix").(string)
	kubernetesVersion := d.Get("kubernetes_version").(string)

	linuxProfile := expandAzureRmKubernetesClusterLinuxProfile(d)
	agentProfiles := expandAzureRmKubernetesClusterAgentProfiles(d)
	servicePrincipalProfile := expandAzureRmKubernetesClusterServicePrincipal(d)
	networkProfile := expandAzureRmKubernetesClusterNetworkProfile(d)
	addonProfiles := expandAzureRmKubernetesClusterAddonProfiles(d)

	tags := d.Get("tags").(map[string]interface{})

	// we can't do this in the CustomizeDiff since the interpolations aren't evaluated at that point
	if networkProfile != nil {
		// ensure there's a Subnet ID attached
		if networkProfile.NetworkPlugin == containerservice.Azure {
			for _, profile := range agentProfiles {
				if profile.VnetSubnetID == nil {
					return fmt.Errorf("A `vnet_subnet_id` must be specified when the `network_plugin` is set to `azure`.")
				}
			}
		}
	}

	parameters := containerservice.ManagedCluster{
		Name:     &name,
		Location: &location,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			AddonProfiles:           addonProfiles,
			AgentPoolProfiles:       &agentProfiles,
			DNSPrefix:               &dnsPrefix,
			KubernetesVersion:       &kubernetesVersion,
			LinuxProfile:            linuxProfile,
			ServicePrincipalProfile: servicePrincipalProfile,
			NetworkProfile:          networkProfile,
		},
		Tags: expandTags(tags),
	}

	ctx := client.StopContext
	future, err := kubernetesClustersClient.CreateOrUpdate(ctx, resGroup, name, parameters)
	if err != nil {
		return err
	}

	err = future.WaitForCompletionRef(ctx, kubernetesClustersClient.Client)
	if err != nil {
		return err
	}

	read, err := kubernetesClustersClient.Get(ctx, resGroup, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read AKS Managed Cluster %q (Resource Group %q) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmKubernetesClusterRead(d, meta)
}

func resourceArmKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	kubernetesClustersClient := meta.(*ArmClient).kubernetesClustersClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["managedClusters"]

	ctx := client.StopContext
	resp, err := kubernetesClustersClient.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AKS Managed Cluster %q (resource group %q): %+v", name, resGroup, err)
	}

	profile, err := kubernetesClustersClient.GetAccessProfile(ctx, resGroup, name, "clusterUser")
	if err != nil {
		return fmt.Errorf("Error getting access profile while making Read request on AKS Managed Cluster %q (resource group %q): %+v", name, resGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	if props := resp.ManagedClusterProperties; props != nil {
		d.Set("dns_prefix", props.DNSPrefix)
		d.Set("fqdn", props.Fqdn)
		d.Set("kubernetes_version", props.KubernetesVersion)
		d.Set("node_resource_group", props.NodeResourceGroup)

		linuxProfile := flattenAzureRmKubernetesClusterLinuxProfile(props.LinuxProfile)
		if err := d.Set("linux_profile", linuxProfile); err != nil {
			return fmt.Errorf("Error setting `linux_profile`: %+v", err)
		}

		addonProfiles := flattenAzureRmKubernetesClusterAddonProfiles(props.AddonProfiles)
		if err := d.Set("addon_profile", addonProfiles); err != nil {
			return fmt.Errorf("Error setting `addon_profile`: %+v", err)
		}

		agentPoolProfiles := flattenAzureRmKubernetesClusterAgentPoolProfiles(props.AgentPoolProfiles, resp.Fqdn)
		if err := d.Set("agent_pool_profile", agentPoolProfiles); err != nil {
			return fmt.Errorf("Error setting `agent_pool_profile`: %+v", err)
		}

		networkProfile := flattenAzureRmKubernetesClusterNetworkProfile(props.NetworkProfile)
		if err := d.Set("network_profile", networkProfile); err != nil {
			return fmt.Errorf("Error setting `network_profile`: %+v", err)
		}

		servicePrincipal := flattenAzureRmKubernetesClusterServicePrincipalProfile(resp.ManagedClusterProperties.ServicePrincipalProfile)
		if err := d.Set("service_principal", servicePrincipal); err != nil {
			return fmt.Errorf("Error setting `service_principal`: %+v", err)
		}
	}

	kubeConfigRaw, kubeConfig := flattenAzureRmKubernetesClusterAccessProfile(&profile)
	d.Set("kube_config_raw", kubeConfigRaw)

	if err := d.Set("kube_config", kubeConfig); err != nil {
		return fmt.Errorf("Error setting `kube_config`: %+v", err)
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmKubernetesClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	kubernetesClustersClient := client.kubernetesClustersClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["managedClusters"]

	ctx := client.StopContext
	future, err := kubernetesClustersClient.Delete(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error issuing AzureRM delete request of AKS Managed Cluster %q (resource Group %q): %+v", name, resGroup, err)
	}

	return future.WaitForCompletionRef(ctx, kubernetesClustersClient.Client)
}

func flattenAzureRmKubernetesClusterLinuxProfile(profile *containerservice.LinuxProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	values := make(map[string]interface{})
	sshKeys := make([]interface{}, 0)

	if username := profile.AdminUsername; username != nil {
		values["admin_username"] = *username
	}

	if ssh := profile.SSH; ssh != nil {
		if keys := ssh.PublicKeys; keys != nil {
			for _, sshKey := range *keys {
				outputs := make(map[string]interface{}, 0)
				if keyData := sshKey.KeyData; keyData != nil {
					outputs["key_data"] = *keyData
				}
				sshKeys = append(sshKeys, outputs)
			}
		}
	}

	values["ssh_key"] = sshKeys

	return []interface{}{values}
}

func flattenAzureRmKubernetesClusterAgentPoolProfiles(profiles *[]containerservice.ManagedClusterAgentPoolProfile, fqdn *string) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	agentPoolProfiles := make([]interface{}, 0)

	for _, profile := range *profiles {
		agentPoolProfile := make(map[string]interface{})

		if profile.Count != nil {
			agentPoolProfile["count"] = int(*profile.Count)
		}

		if fqdn != nil {
			// temporarily persist the parent FQDN here until `fqdn` is removed from the `agent_pool_profile`
			agentPoolProfile["fqdn"] = *fqdn
		}

		if profile.Name != nil {
			agentPoolProfile["name"] = *profile.Name
		}

		if profile.VMSize != "" {
			agentPoolProfile["vm_size"] = string(profile.VMSize)
		}

		if profile.OsDiskSizeGB != nil {
			agentPoolProfile["os_disk_size_gb"] = int(*profile.OsDiskSizeGB)
		}

		if profile.VnetSubnetID != nil {
			agentPoolProfile["vnet_subnet_id"] = *profile.VnetSubnetID
		}

		if profile.OsType != "" {
			agentPoolProfile["os_type"] = string(profile.OsType)
		}

		if profile.MaxPods != nil {
			agentPoolProfile["max_pods"] = int(*profile.MaxPods)
		}

		agentPoolProfiles = append(agentPoolProfiles, agentPoolProfile)
	}

	return agentPoolProfiles
}

func flattenAzureRmKubernetesClusterServicePrincipalProfile(profile *containerservice.ManagedClusterServicePrincipalProfile) *schema.Set {
	if profile == nil {
		return nil
	}

	servicePrincipalProfiles := &schema.Set{
		F: resourceAzureRMKubernetesClusterServicePrincipalProfileHash,
	}

	values := make(map[string]interface{})

	if clientId := profile.ClientID; clientId != nil {
		values["client_id"] = *clientId
	}
	if secret := profile.Secret; secret != nil {
		values["client_secret"] = *secret
	}

	servicePrincipalProfiles.Add(values)

	return servicePrincipalProfiles
}

func flattenAzureRmKubernetesClusterAccessProfile(profile *containerservice.ManagedClusterAccessProfile) (*string, []interface{}) {
	if profile != nil {
		if accessProfile := profile.AccessProfile; accessProfile != nil {
			if kubeConfigRaw := accessProfile.KubeConfig; kubeConfigRaw != nil {
				rawConfig := string(*kubeConfigRaw)

				kubeConfig, err := kubernetes.ParseKubeConfig(rawConfig)
				if err != nil {
					return utils.String(rawConfig), []interface{}{}
				}

				flattenedKubeConfig := flattenKubernetesClusterKubeConfig(*kubeConfig)
				return utils.String(rawConfig), flattenedKubeConfig
			}
		}
	}
	return nil, []interface{}{}
}

func flattenAzureRmKubernetesClusterNetworkProfile(profile *containerservice.NetworkProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	values := make(map[string]interface{})

	values["network_plugin"] = profile.NetworkPlugin

	if profile.ServiceCidr != nil {
		values["service_cidr"] = *profile.ServiceCidr
	}

	if profile.DNSServiceIP != nil {
		values["dns_service_ip"] = *profile.DNSServiceIP
	}

	if profile.DockerBridgeCidr != nil {
		values["docker_bridge_cidr"] = *profile.DockerBridgeCidr
	}

	if profile.PodCidr != nil {
		values["pod_cidr"] = *profile.PodCidr
	}

	return []interface{}{values}
}

func flattenKubernetesClusterKubeConfig(config kubernetes.KubeConfig) []interface{} {
	values := make(map[string]interface{})

	cluster := config.Clusters[0].Cluster
	user := config.Users[0].User
	name := config.Users[0].Name

	values["host"] = cluster.Server
	values["username"] = name
	values["password"] = user.Token
	values["client_certificate"] = user.ClientCertificteData
	values["client_key"] = user.ClientKeyData
	values["cluster_ca_certificate"] = cluster.ClusterAuthorityData

	return []interface{}{values}
}

func expandAzureRmKubernetesClusterLinuxProfile(d *schema.ResourceData) *containerservice.LinuxProfile {
	profiles := d.Get("linux_profile").([]interface{})

	if len(profiles) == 0 {
		return nil
	}

	config := profiles[0].(map[string]interface{})

	adminUsername := config["admin_username"].(string)
	linuxKeys := config["ssh_key"].([]interface{})

	keyData := ""
	if key, ok := linuxKeys[0].(map[string]interface{}); ok {
		keyData = key["key_data"].(string)
	}
	sshPublicKey := containerservice.SSHPublicKey{
		KeyData: &keyData,
	}

	sshPublicKeys := []containerservice.SSHPublicKey{sshPublicKey}

	profile := containerservice.LinuxProfile{
		AdminUsername: &adminUsername,
		SSH: &containerservice.SSHConfiguration{
			PublicKeys: &sshPublicKeys,
		},
	}

	return &profile
}

func expandAzureRmKubernetesClusterServicePrincipal(d *schema.ResourceData) *containerservice.ManagedClusterServicePrincipalProfile {
	value, exists := d.GetOk("service_principal")
	if !exists {
		return nil
	}

	configs := value.(*schema.Set).List()

	config := configs[0].(map[string]interface{})

	clientId := config["client_id"].(string)
	clientSecret := config["client_secret"].(string)

	principal := containerservice.ManagedClusterServicePrincipalProfile{
		ClientID: &clientId,
		Secret:   &clientSecret,
	}

	return &principal
}

func expandAzureRmKubernetesClusterAgentProfiles(d *schema.ResourceData) []containerservice.ManagedClusterAgentPoolProfile {
	configs := d.Get("agent_pool_profile").([]interface{})
	config := configs[0].(map[string]interface{})
	profiles := make([]containerservice.ManagedClusterAgentPoolProfile, 0, len(configs))

	name := config["name"].(string)
	count := int32(config["count"].(int))
	vmSize := config["vm_size"].(string)
	osDiskSizeGB := int32(config["os_disk_size_gb"].(int))
	osType := config["os_type"].(string)

	profile := containerservice.ManagedClusterAgentPoolProfile{
		Name:           utils.String(name),
		Count:          utils.Int32(count),
		VMSize:         containerservice.VMSizeTypes(vmSize),
		OsDiskSizeGB:   utils.Int32(osDiskSizeGB),
		StorageProfile: containerservice.ManagedDisks,
		OsType:         containerservice.OSType(osType),
	}

	if maxPods := int32(config["max_pods"].(int)); maxPods > 0 {
		profile.MaxPods = utils.Int32(maxPods)
	}

	vnetSubnetID := config["vnet_subnet_id"].(string)
	if vnetSubnetID != "" {
		profile.VnetSubnetID = utils.String(vnetSubnetID)
	}

	profiles = append(profiles, profile)

	return profiles
}

func expandAzureRmKubernetesClusterNetworkProfile(d *schema.ResourceData) *containerservice.NetworkProfile {
	configs := d.Get("network_profile").([]interface{})
	if len(configs) == 0 {
		return nil
	}

	config := configs[0].(map[string]interface{})

	networkPlugin := config["network_plugin"].(string)

	networkProfile := containerservice.NetworkProfile{
		NetworkPlugin: containerservice.NetworkPlugin(networkPlugin),
	}

	if v, ok := config["dns_service_ip"]; ok && v.(string) != "" {
		dnsServiceIP := v.(string)
		networkProfile.DNSServiceIP = utils.String(dnsServiceIP)
	}

	if v, ok := config["pod_cidr"]; ok && v.(string) != "" {
		podCidr := v.(string)
		networkProfile.PodCidr = utils.String(podCidr)
	}

	if v, ok := config["docker_bridge_cidr"]; ok && v.(string) != "" {
		dockerBridgeCidr := v.(string)
		networkProfile.DockerBridgeCidr = utils.String(dockerBridgeCidr)
	}

	if v, ok := config["service_cidr"]; ok && v.(string) != "" {
		serviceCidr := v.(string)
		networkProfile.ServiceCidr = utils.String(serviceCidr)
	}

	return &networkProfile
}

func expandAzureRmKubernetesClusterAddonProfiles(d *schema.ResourceData) map[string]*containerservice.ManagedClusterAddonProfile {
	profiles := d.Get("addon_profile").([]interface{})
	if len(profiles) == 0 {
		return nil
	}

	profile := profiles[0].(map[string]interface{})
	addonProfiles := map[string]*containerservice.ManagedClusterAddonProfile{}

	httpApplicationRouting := profile["http_application_routing"].([]interface{})
	if len(httpApplicationRouting) > 0 {
		value := httpApplicationRouting[0].(map[string]interface{})
		enabled := value["enabled"].(bool)
		addonProfiles["httpApplicationRouting"] = &containerservice.ManagedClusterAddonProfile{
			Enabled: utils.Bool(enabled),
		}
	}

	omsAgent := profile["oms_agent"].([]interface{})
	if len(omsAgent) > 0 {
		value := omsAgent[0].(map[string]interface{})
		config := make(map[string]*string)
		enabled := value["enabled"].(bool)

		if workspaceId, ok := value["log_analytics_workspace_id"]; ok {
			config["logAnalyticsWorkspaceResourceID"] = utils.String(workspaceId.(string))
		}

		addonProfiles["omsagent"] = &containerservice.ManagedClusterAddonProfile{
			Enabled: utils.Bool(enabled),
			Config:  config,
		}
	}

	return addonProfiles
}

func flattenAzureRmKubernetesClusterAddonProfiles(profile map[string]*containerservice.ManagedClusterAddonProfile) []interface{} {
	values := make(map[string]interface{}, 0)

	routes := make([]interface{}, 0)
	if httpApplicationRouting := profile["httpApplicationRouting"]; httpApplicationRouting != nil {
		enabled := false
		if enabledVal := httpApplicationRouting.Enabled; enabledVal != nil {
			enabled = *enabledVal
		}

		zoneName := ""
		if v := httpApplicationRouting.Config["HTTPApplicationRoutingZoneName"]; v != nil {
			zoneName = *v
		}

		output := map[string]interface{}{
			"enabled": enabled,
			"http_application_routing_zone_name": zoneName,
		}
		routes = append(routes, output)
	}
	values["http_application_routing"] = routes

	agents := make([]interface{}, 0)
	if omsAgent := profile["omsagent"]; omsAgent != nil {
		enabled := false
		if enabledVal := omsAgent.Enabled; enabledVal != nil {
			enabled = *enabledVal
		}

		workspaceId := ""
		if workspaceResourceID := omsAgent.Config["logAnalyticsWorkspaceResourceID"]; workspaceResourceID != nil {
			workspaceId = *workspaceResourceID
		}

		output := map[string]interface{}{
			"enabled":                    enabled,
			"log_analytics_workspace_id": workspaceId,
		}
		agents = append(agents, output)
	}
	values["oms_agent"] = agents

	return []interface{}{values}
}

func resourceAzureRMKubernetesClusterServicePrincipalProfileHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["client_id"].(string)))
	}

	return hashcode.String(buf.String())
}

func validateKubernetesClusterAgentPoolName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[a-z]{1}[a-z0-9]{0,11}$"),
		"Agent Pool names must start with a lowercase letter, have max length of 12, and only have characters a-z0-9.",
	)
}

package psm

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// convertToStringSlice converts various input types to a slice of strings
func convertToStringSlice(input interface{}) []string {
	var result []string
	inputSlice, ok := input.([]interface{})
	if !ok {
		return result
	}
	for _, v := range inputSlice {
		str, ok := v.(string)
		if !ok {
			continue
		}
		result = append(result, str)
	}
	return result
}

func getStringOrEmpty(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getBoolOrDefault(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		b, ok := v.(bool)
		if ok {
			return b
		}
	}
	return defaultValue
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				result[i] = fmt.Sprintf("%v", item)
			}
			return result
		}
	}
	return nil
}

func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func ExpandStringSet(set *schema.Set) []string {
	list := set.List()
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func FlattenStringList(list []string) *schema.Set {
	set := schema.NewSet(schema.HashString, []interface{}{})
	for _, v := range list {
		set.Add(v)
	}
	return set
}

func expandSpec(d *schema.ResourceData) TunnelSpec {
	tunnelData := d.Get("tunnel").([]interface{})[0].(map[string]interface{})

	spec := TunnelSpec{
		HAMode:                    getStringOrEmpty(tunnelData, "ha_mode"),
		TunnelEndpoints:           expandTunnelEndpoints(tunnelData["tunnel_endpoints"].([]interface{})),
		PolicyDistributionTargets: convertToStringSlice(tunnelData["policy_distribution_targets"].([]interface{})),
		DisableTCPMSSAdjust:       getBoolOrDefault(tunnelData, "disable_tcp_mss_adjust", false),
	}

	return spec
}

func expandTunnelEndpoints(endpoints []interface{}) []TunnelEndpoint {
	var tunnelEndpoints []TunnelEndpoint

	for _, endpoint := range endpoints {
		endpointData := endpoint.(map[string]interface{})
		tunnelEndpoint := TunnelEndpoint{
			InterfaceName: getStringOrEmpty(endpointData, "interface_name"),
			DSE:           getStringOrEmpty(endpointData, "dse"),
			IKEVersion:    getStringOrEmpty(endpointData, "ike_version"),
		}

		if ikeSA, ok := endpointData["ike_sa"].([]interface{}); ok && len(ikeSA) > 0 {
			tunnelEndpoint.IKESA = expandIKESA(ikeSA[0].(map[string]interface{}))
		}

		if ipsecSA, ok := endpointData["ipsec_sa"].([]interface{}); ok && len(ipsecSA) > 0 {
			tunnelEndpoint.IPSECSA = expandIPSECSA(ipsecSA[0].(map[string]interface{}))
		}

		if localID, ok := endpointData["local_identifier"].([]interface{}); ok && len(localID) > 0 {
			tunnelEndpoint.LocalIdentifier = expandIdentifier(localID[0].(map[string]interface{}))
		}

		if remoteID, ok := endpointData["remote_identifier"].([]interface{}); ok && len(remoteID) > 0 {
			tunnelEndpoint.RemoteIdentifier = expandIdentifier(remoteID[0].(map[string]interface{}))
		}

		// Expand lifetime for this endpoint
		if lifetime, ok := endpointData["lifetime"].([]interface{}); ok && len(lifetime) > 0 {
			lifetimeData := lifetime[0].(map[string]interface{})
			tunnelEndpoint.Lifetime = &Lifetime{
				SALifetime:  getStringOrEmpty(lifetimeData, "sa_lifetime"),
				IKELifetime: getStringOrEmpty(lifetimeData, "ike_lifetime"),
			}
		}

		tunnelEndpoints = append(tunnelEndpoints, tunnelEndpoint)
	}

	return tunnelEndpoints
}

func expandIKESA(data map[string]interface{}) *IKESA {
	ikesa := &IKESA{
		EncryptionAlgorithms: convertToStringSlice(data["encryption_algorithms"].([]interface{})),
		HashAlgorithms:       convertToStringSlice(data["hash_algorithms"].([]interface{})),
		DHGroups:             convertToStringSlice(data["dh_groups"].([]interface{})),
		RekeyLifetime:        getStringOrEmpty(data, "rekey_lifetime"),
		PreSharedKey:         getStringOrEmpty(data, "pre_shared_key"),
		ReauthLifetime:       getStringOrEmpty(data, "reauth_lifetime"),
		DPDDelay:             getStringOrEmpty(data, "dpd_delay"),
		IKEV1DPDTimeout:      getStringOrEmpty(data, "ikev1_dpd_timeout"),
		IKEInitiator:         getBoolOrDefault(data, "ike_initiator", true),
		AuthType:             getStringOrEmpty(data, "auth_type"),
	}

	if ikesa.AuthType == "certificates" {
		ikesa.LocalIdentityCertificates = getStringOrEmpty(data, "local_identity_certificates")
		if certs, ok := data["remote_ca_certificates"].([]interface{}); ok {
			ikesa.RemoteCACertificates = convertToStringSlice(certs)
		}
	}

	return ikesa
}

func expandIPSECSA(data map[string]interface{}) *IPSECSA {
	return &IPSECSA{
		EncryptionAlgorithms: convertToStringSlice(data["encryption_algorithms"].([]interface{})),
		DHGroups:             convertToStringSlice(data["dh_groups"].([]interface{})),
		RekeyLifetime:        getStringOrEmpty(data, "rekey_lifetime"),
	}
}

func expandIdentifier(data map[string]interface{}) Identifier {
	return Identifier{
		Type:  getStringOrEmpty(data, "type"),
		Value: getStringOrEmpty(data, "value"),
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func flattenSpec(spec *TunnelSpec, d *schema.ResourceData) error {
	if spec == nil {
		return fmt.Errorf("TunnelSpec is nil")
	}

	tunnelMap := make(map[string]interface{})
	tunnelMap["ha_mode"] = spec.HAMode
	tunnelMap["policy_distribution_targets"] = spec.PolicyDistributionTargets
	tunnelMap["disable_tcp_mss_adjust"] = spec.DisableTCPMSSAdjust

	tunnelEndpoints := flattenTunnelEndpoints(spec.TunnelEndpoints, d)
	tunnelMap["tunnel_endpoints"] = tunnelEndpoints

	if spec.Config != nil {
		lifetime := make(map[string]interface{})

		// Preserve existing values if new values are empty
		oldTunnel, ok := d.GetOk("tunnel")
		if ok && len(oldTunnel.([]interface{})) > 0 {
			oldTunnelMap := oldTunnel.([]interface{})[0].(map[string]interface{})
			if oldLifetime, ok := oldTunnelMap["lifetime"].([]interface{}); ok && len(oldLifetime) > 0 {
				oldLifetimeMap := oldLifetime[0].(map[string]interface{})

				if spec.Config.SALifetime != "" {
					lifetime["sa_lifetime"] = spec.Config.SALifetime
				} else if saLifetime, ok := oldLifetimeMap["sa_lifetime"]; ok {
					lifetime["sa_lifetime"] = saLifetime
				}

				if spec.Config.IKELifetime != "" {
					lifetime["ike_lifetime"] = spec.Config.IKELifetime
				} else if ikeLifetime, ok := oldLifetimeMap["ike_lifetime"]; ok {
					lifetime["ike_lifetime"] = ikeLifetime
				}
			}
		} else {
			// If there are no old values, only set the new values if they're not empty
			if spec.Config.SALifetime != "" {
				lifetime["sa_lifetime"] = spec.Config.SALifetime
			}
			if spec.Config.IKELifetime != "" {
				lifetime["ike_lifetime"] = spec.Config.IKELifetime
			}
		}

		// Only add the lifetime block if it's not empty
		if len(lifetime) > 0 {
			tunnelMap["lifetime"] = []interface{}{lifetime}
		}
	}

	if err := d.Set("tunnel", []interface{}{tunnelMap}); err != nil {
		return fmt.Errorf("error setting tunnel: %s", err)
	}

	return nil
}

func flattenTunnelEndpoints(endpoints []TunnelEndpoint, d *schema.ResourceData) []interface{} {
	tunnelEndpoints := make([]interface{}, len(endpoints))
	for i, endpoint := range endpoints {
		endpointMap := make(map[string]interface{})
		endpointMap["interface_name"] = endpoint.InterfaceName
		endpointMap["dse"] = endpoint.DSE
		endpointMap["ike_version"] = endpoint.IKEVersion

		if endpoint.IKESA != nil {
			ikesa := make(map[string]interface{})
			ikesa["encryption_algorithms"] = endpoint.IKESA.EncryptionAlgorithms
			ikesa["hash_algorithms"] = endpoint.IKESA.HashAlgorithms
			ikesa["dh_groups"] = endpoint.IKESA.DHGroups
			ikesa["rekey_lifetime"] = endpoint.IKESA.RekeyLifetime
			ikesa["reauth_lifetime"] = endpoint.IKESA.ReauthLifetime
			ikesa["dpd_delay"] = endpoint.IKESA.DPDDelay
			ikesa["ikev1_dpd_timeout"] = endpoint.IKESA.IKEV1DPDTimeout
			ikesa["ike_initiator"] = endpoint.IKESA.IKEInitiator
			ikesa["auth_type"] = endpoint.IKESA.AuthType
			if endpoint.IKESA.LocalIdentityCertificates != "" {
				ikesa["local_identity_certificates"] = endpoint.IKESA.LocalIdentityCertificates
			}
			if len(endpoint.IKESA.RemoteCACertificates) > 0 {
				ikesa["remote_ca_certificates"] = endpoint.IKESA.RemoteCACertificates
			}

			// Preserve pre_shared_key if it exists in the current state
			if v, ok := d.GetOk(fmt.Sprintf("tunnel.0.tunnel_endpoints.%d.ike_sa.0.pre_shared_key", i)); ok {
				ikesa["pre_shared_key"] = v.(string)
			}

			endpointMap["ike_sa"] = []interface{}{ikesa}
		}

		if endpoint.IPSECSA != nil {
			ipsecsa := make(map[string]interface{})
			ipsecsa["encryption_algorithms"] = endpoint.IPSECSA.EncryptionAlgorithms
			ipsecsa["dh_groups"] = endpoint.IPSECSA.DHGroups
			ipsecsa["rekey_lifetime"] = endpoint.IPSECSA.RekeyLifetime
			endpointMap["ipsec_sa"] = []interface{}{ipsecsa}
		}

		endpointMap["local_identifier"] = []interface{}{
			map[string]interface{}{
				"type":  endpoint.LocalIdentifier.Type,
				"value": endpoint.LocalIdentifier.Value,
			},
		}

		endpointMap["remote_identifier"] = []interface{}{
			map[string]interface{}{
				"type":  endpoint.RemoteIdentifier.Type,
				"value": endpoint.RemoteIdentifier.Value,
			},
		}

		// Add lifetime configuration if it exists
		if endpoint.Lifetime != nil {
			endpointMap["lifetime"] = []interface{}{
				map[string]interface{}{
					"sa_lifetime":  endpoint.Lifetime.SALifetime,
					"ike_lifetime": endpoint.Lifetime.IKELifetime,
				},
			}
		} else {
			// Check if lifetime exists in the current state and preserve it
			if v, ok := d.GetOk(fmt.Sprintf("tunnel.0.tunnel_endpoints.%d.lifetime", i)); ok {
				endpointMap["lifetime"] = v.([]interface{})
			}
		}

		tunnelEndpoints[i] = endpointMap
	}

	return tunnelEndpoints
}

func suppressMissingLifetimeValues(k, old, new string, d *schema.ResourceData) bool {
	// If the new value is empty, suppress the diff
	return new == ""
}

func validateNATRules(rules []interface{}) error {
	for i, r := range rules {
		rule, ok := r.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid rule type at index %d", i)
		}

		// Validate name
		name, ok := rule["name"].(string)
		if !ok || name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}

		// Validate type
		ruleType, ok := rule["type"].(string)
		if !ok || ruleType == "" {
			return fmt.Errorf("rule %d: type is required", i)
		}

		// Validate source
		if err := validateAddressOrCollection(rule, "source", i); err != nil {
			return err
		}

		// Validate destination (optional)
		if destination, ok := rule["destination"].([]interface{}); ok && len(destination) > 0 {
			if err := validateAddressOrCollection(rule, "destination", i); err != nil {
				return err
			}
		}

		// Validate destination_proto_port (optional)
		if destProtoPort, ok := rule["destination_proto_port"].([]interface{}); ok && len(destProtoPort) > 0 {
			dpp := destProtoPort[0].(map[string]interface{})
			protocol, ok := dpp["protocol"].(string)
			if !ok || protocol == "" {
				return fmt.Errorf("rule %d: protocol in destination_proto_port is required if destination_proto_port is specified", i)
			}
			// Ports are optional, no need to validate
		}

		// Validate that at least one of translated_source or translated_destination is specified
		hasTranslatedSource := rule["translated_source"] != nil
		hasTranslatedDest := rule["translated_destination"] != nil

		if !hasTranslatedSource && !hasTranslatedDest {
			return fmt.Errorf("rule %d: either translated_source or translated_destination must be specified", i)
		}

		// Validate translated_destination_port (optional)
		if translatedDestPort, ok := rule["translated_destination_port"].(string); ok && translatedDestPort != "" {
			if !hasTranslatedDest {
				return fmt.Errorf("rule %d: translated_destination is required when translated_destination_port is specified", i)
			}
		}
	}
	return nil
}

func createNatRule(r map[string]interface{}) NatRule {
	natRule := NatRule{
		Name:    r["name"].(string),
		Disable: r["disable"].(bool),
		Type:    r["type"].(string),
	}

	if source, ok := r["source"].([]interface{}); ok && len(source) > 0 {
		setAddressOrCollection(&natRule.Source, source[0].(map[string]interface{}))
	}

	if destination, ok := r["destination"].([]interface{}); ok && len(destination) > 0 {
		setAddressOrCollection(&natRule.Destination, destination[0].(map[string]interface{}))
	}

	if destProtoPort, ok := r["destination_proto_port"].([]interface{}); ok && len(destProtoPort) > 0 {
		dpp := destProtoPort[0].(map[string]interface{})
		natRule.DestinationProtoPort.Protocol = dpp["protocol"].(string)
		if ports, ok := dpp["ports"].(string); ok {
			natRule.DestinationProtoPort.Ports = ports
		}
	}

	if translatedSource, ok := r["translated_source"].([]interface{}); ok && len(translatedSource) > 0 {
		setAddressOrCollection(&natRule.TranslatedSource, translatedSource[0].(map[string]interface{}))
	}

	if translatedDest, ok := r["translated_destination"].([]interface{}); ok && len(translatedDest) > 0 {
		setAddressOrCollection(&natRule.TranslatedDestination, translatedDest[0].(map[string]interface{}))
	} else if natRule.Type == "static" && !natRule.Destination.Any {
		// For static NAT, if translated_destination is not provided but destination is,
		// set translated_destination to be the same as destination
		natRule.TranslatedDestination = natRule.Destination
	}

	if translatedDestPort, ok := r["translated_destination_port"].(string); ok {
		natRule.TranslatedDestinationPort = translatedDestPort
	}

	return natRule
}

func setAddressOrCollection(target *AddressCollection, source map[string]interface{}) {
	if addresses, ok := source["addresses"].([]interface{}); ok && len(addresses) > 0 {
		target.Addresses = expandStringList(addresses)
	}
	if ipCollections, ok := source["ipcollections"].([]interface{}); ok && len(ipCollections) > 0 {
		target.IPCollections = expandStringList(ipCollections)
	}
	target.Any = len(target.Addresses) == 0 && len(target.IPCollections) == 0
}

func validateAddressOrCollection(rule map[string]interface{}, field string, ruleIndex int) error {
	value, ok := rule[field].([]interface{})
	if !ok || len(value) == 0 {
		return fmt.Errorf("rule %d: %s is required", ruleIndex, field)
	}

	valueMap := value[0].(map[string]interface{})
	addresses := valueMap["addresses"].([]interface{})
	ipCollections := valueMap["ipcollections"].([]interface{})

	if len(addresses) == 0 && len(ipCollections) == 0 {
		return fmt.Errorf("rule %d: %s must have either addresses or ipcollections", ruleIndex, field)
	}

	return nil
}

func expandPermissions(permissions []interface{}) []struct {
	ResourceGroup     string   `json:"resource-group"`
	ResourceKind      string   `json:"resource-kind"`
	ResourceNamespace string   `json:"resource-namespace"`
	Actions           []string `json:"actions"`
} {
	result := make([]struct {
		ResourceGroup     string   `json:"resource-group"`
		ResourceKind      string   `json:"resource-kind"`
		ResourceNamespace string   `json:"resource-namespace"`
		Actions           []string `json:"actions"`
	}, len(permissions))

	for i, perm := range permissions {
		p := perm.(map[string]interface{})
		result[i] = struct {
			ResourceGroup     string   `json:"resource-group"`
			ResourceKind      string   `json:"resource-kind"`
			ResourceNamespace string   `json:"resource-namespace"`
			Actions           []string `json:"actions"`
		}{
			ResourceGroup:     p["resource_group"].(string),
			ResourceKind:      p["resource_kind"].(string),
			ResourceNamespace: "*_ALL_*",
			Actions:           expandStringList(p["actions"].([]interface{})),
		}
	}

	return result
}

func flattenPermissions(permissions []struct {
	ResourceGroup     string   `json:"resource-group"`
	ResourceKind      string   `json:"resource-kind"`
	ResourceNamespace string   `json:"resource-namespace"`
	Actions           []string `json:"actions"`
}) []interface{} {
	result := make([]interface{}, len(permissions))

	for i, perm := range permissions {
		result[i] = map[string]interface{}{
			"resource_group": perm.ResourceGroup,
			"resource_kind":  perm.ResourceKind,
			"actions":        perm.Actions,
		}
	}

	return result
}

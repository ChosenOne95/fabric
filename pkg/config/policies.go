/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/policydsl"
)

// GetPoliciesForConsortiums returns a map of policies for channel consortiums.
func GetPoliciesForConsortiums(config cb.Config) (map[string]Policy, error) {
	consortiums, ok := config.ChannelGroup.Groups[ConsortiumsGroupKey]
	if !ok {
		return nil, errors.New("consortiums missing from config")
	}

	return getPolicies(consortiums.Policies)
}

// GetPoliciesForConsortium returns a map of policies for a specific consortium.
func GetPoliciesForConsortium(config cb.Config, consortiumName string) (map[string]Policy, error) {
	consortium, ok := config.ChannelGroup.Groups[ConsortiumsGroupKey].Groups[consortiumName]
	if !ok {
		return nil, fmt.Errorf("consortium %s does not exist in channel config", consortiumName)
	}

	return getPolicies(consortium.Policies)
}

// GetPoliciesForConsortiumOrg returns a map of policies for a specific consortium org.
func GetPoliciesForConsortiumOrg(config cb.Config, consortiumName, orgName string) (map[string]Policy, error) {
	org, ok := config.ChannelGroup.Groups[ConsortiumsGroupKey].Groups[consortiumName].Groups[orgName]
	if !ok {
		return nil, fmt.Errorf("consortium org %s does not exist in channel config", orgName)
	}

	return getPolicies(org.Policies)
}

// GetPoliciesForOrderer returns a map of policies for channel orderer.
func GetPoliciesForOrderer(config cb.Config) (map[string]Policy, error) {
	orderer, ok := config.ChannelGroup.Groups[OrdererGroupKey]
	if !ok {
		return nil, errors.New("orderer missing from config")
	}

	return getPolicies(orderer.Policies)
}

// GetPoliciesForOrdererOrg returns a map of policies for a specific org.
func GetPoliciesForOrdererOrg(config cb.Config, orgName string) (map[string]Policy, error) {
	org, ok := config.ChannelGroup.Groups[OrdererGroupKey].Groups[orgName]
	if !ok {
		return nil, fmt.Errorf("orderer org %s does not exist in channel config", orgName)
	}

	return getPolicies(org.Policies)
}

// GetPoliciesForApplication returns a map of policies for application config group.
func GetPoliciesForApplication(config cb.Config) (map[string]Policy, error) {
	application, ok := config.ChannelGroup.Groups[ApplicationGroupKey]
	if !ok {
		return nil, errors.New("application missing from config")
	}

	return getPolicies(application.Policies)
}

// GetPoliciesForApplicationOrg returns a map of policies for specific application
// organization.
func GetPoliciesForApplicationOrg(config cb.Config, orgName string) (map[string]Policy, error) {
	orgGroup, ok := config.ChannelGroup.Groups[ApplicationGroupKey].Groups[orgName]
	if !ok {
		return nil, fmt.Errorf("application org %s does not exist in channel config", orgName)
	}

	return getPolicies(orgGroup.Policies)
}

// AddApplicationPolicy modifies an existing application policy configuration.
// When the policy exists it will overwrite the existing policy.
func AddApplicationPolicy(config *cb.Config, modPolicy, policyName string, policy Policy) error {
	err := addPolicy(config.ChannelGroup.Groups[ApplicationGroupKey], modPolicy, policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to add policy '%s': %v", policyName, err)
	}

	return nil
}

// RemoveApplicationPolicy removes an existing application policy configuration.
// The policy must exist in the config.
func RemoveApplicationPolicy(config *cb.Config, policyName string) error {
	policies, err := GetPoliciesForApplication(*config)
	if err != nil {
		return err
	}

	return removePolicy(config.ChannelGroup.Groups[ApplicationGroupKey], policyName, policies)
}

// AddApplicationOrgPolicy modifies an existing organization in a application configuration's policies.
// When the policy exists it will overwrite the existing policy.
func AddApplicationOrgPolicy(config *cb.Config, orgName, modPolicy, policyName string, policy Policy) error {
	err := addPolicy(config.ChannelGroup.Groups[ApplicationGroupKey].Groups[orgName], modPolicy, policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to add policy '%s': %v", policyName, err)
	}

	return nil
}

// RemoveApplicationOrgPolicy removes an existing policy from an application organization.
// The removed policy must exist.
func RemoveApplicationOrgPolicy(config *cb.Config, orgName, policyName string) error {
	policies, err := GetPoliciesForApplicationOrg(*config, orgName)
	if err != nil {
		return err
	}

	return removePolicy(config.ChannelGroup.Groups[ApplicationGroupKey].Groups[orgName], policyName, policies)
}

// AddConsortiumOrgPolicy modifies an existing organization in a consortiums configuration's policies.
// When the policy exists it will overwrite the existing policy.
func AddConsortiumOrgPolicy(config *cb.Config, consortiumName, orgName, policyName string, policy Policy) error {
	groupKey := ConsortiumsGroupKey

	consortiumGroup, ok := config.ChannelGroup.Groups[groupKey].Groups[consortiumName]
	if !ok {
		return fmt.Errorf("consortium '%s' does not exist in channel config", consortiumName)
	}

	orgGroup, ok := consortiumGroup.Groups[orgName]
	if !ok {
		return fmt.Errorf("%s org '%s' does not exist in channel config", strings.ToLower(groupKey), orgName)
	}

	err := addPolicy(orgGroup, AdminsPolicyKey, policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to add policy '%s' to consortium org '%s': %v", policyName, orgName, err)
	}

	return nil
}

// RemoveConsortiumOrgPolicy removes an existing policy from an consortiums organization.
// The removed policy must exist however will not error if it does not exist in configuration.
func RemoveConsortiumOrgPolicy(config *cb.Config, consortiumName, orgName, policyName string) error {
	groupKey := ConsortiumsGroupKey

	consortiumGroup, ok := config.ChannelGroup.Groups[groupKey].Groups[consortiumName]
	if !ok {
		return fmt.Errorf("consortium '%s' does not exist in channel config", consortiumName)
	}

	orgGroup, ok := consortiumGroup.Groups[orgName]
	if !ok {
		return fmt.Errorf("%s org '%s' does not exist in channel config", strings.ToLower(groupKey), orgName)
	}

	delete(orgGroup.Policies, policyName)

	return nil
}

// AddOrdererPolicy modifies an existing orderer policy configuration.
// When the policy exists it will overwrite the existing policy.
func AddOrdererPolicy(config *cb.Config, modPolicy, policyName string, policy Policy) error {
	err := addPolicy(config.ChannelGroup.Groups[OrdererGroupKey], modPolicy, policyName, policy)
	if err != nil {
		return fmt.Errorf("failed to add policy '%s': %v", policyName, err)
	}

	return nil
}

// RemoveOrdererPolicy removes an existing orderer policy configuration.
// The policy must exist in the config.
func RemoveOrdererPolicy(config *cb.Config, policyName string) error {
	if policyName == BlockValidationPolicyKey {
		return errors.New("BlockValidation policy must be defined")
	}

	policies, err := GetPoliciesForOrderer(*config)
	if err != nil {
		return err
	}

	return removePolicy(config.ChannelGroup.Groups[OrdererGroupKey], policyName, policies)
}

// AddOrdererOrgPolicy modifies an existing organization in a orderer configuration's policies.
// When the policy exists it will overwrite the existing policy.
func AddOrdererOrgPolicy(config *cb.Config, orgName, modPolicy, policyName string, policy Policy) error {
	return addPolicy(config.ChannelGroup.Groups[OrdererGroupKey].Groups[orgName], modPolicy, policyName, policy)
}

// RemoveOrdererOrgPolicy removes an existing policy from an orderer organization.
// The removed policy must exist however will not error if it does not exist in configuration.
func RemoveOrdererOrgPolicy(config *cb.Config, orgName, policyName string) error {
	policies, err := GetPoliciesForOrdererOrg(*config, orgName)
	if err != nil {
		return err
	}

	return removePolicy(config.ChannelGroup.Groups[OrdererGroupKey].Groups[orgName], policyName, policies)
}

// getPolicies returns a map of Policy from given map of ConfigPolicy in organization config group.
func getPolicies(policies map[string]*cb.ConfigPolicy) (map[string]Policy, error) {
	p := map[string]Policy{}

	for name, policy := range policies {
		switch cb.Policy_PolicyType(policy.Policy.Type) {
		case cb.Policy_IMPLICIT_META:
			imp := &cb.ImplicitMetaPolicy{}
			err := proto.Unmarshal(policy.Policy.Value, imp)
			if err != nil {
				return nil, err
			}

			rule, err := implicitMetaToString(imp)
			if err != nil {
				return nil, err
			}

			p[name] = Policy{
				Type: ImplicitMetaPolicyType,
				Rule: rule,
			}
		case cb.Policy_SIGNATURE:
			sp := &cb.SignaturePolicyEnvelope{}
			err := proto.Unmarshal(policy.Policy.Value, sp)
			if err != nil {
				return nil, err
			}

			rule, err := signatureMetaToString(sp)
			if err != nil {
				return nil, err
			}

			p[name] = Policy{
				Type: SignaturePolicyType,
				Rule: rule,
			}
		default:
			return nil, fmt.Errorf("unknown policy type: %v", policy.Policy.Type)
		}
	}

	return p, nil
}

// implicitMetaToString converts a *cb.ImplicitMetaPolicy to a string representation.
func implicitMetaToString(imp *cb.ImplicitMetaPolicy) (string, error) {
	var args string

	switch imp.Rule {
	case cb.ImplicitMetaPolicy_ANY:
		args += cb.ImplicitMetaPolicy_ANY.String()
	case cb.ImplicitMetaPolicy_ALL:
		args += cb.ImplicitMetaPolicy_ALL.String()
	case cb.ImplicitMetaPolicy_MAJORITY:
		args += cb.ImplicitMetaPolicy_MAJORITY.String()
	default:
		return "", fmt.Errorf("unknown implicit meta policy rule type %v", imp.Rule)
	}

	args = args + " " + imp.SubPolicy

	return args, nil
}

// signatureMetaToString converts a *cb.SignaturePolicyEnvelope to a string representation.
func signatureMetaToString(sig *cb.SignaturePolicyEnvelope) (string, error) {
	var roles []string

	for _, id := range sig.Identities {
		role, err := mspPrincipalToString(id)
		if err != nil {
			return "", err
		}

		roles = append(roles, role)
	}

	return signaturePolicyToString(sig.Rule, roles)
}

// mspPrincipalToString converts a *mb.MSPPrincipal to a string representation.
func mspPrincipalToString(principal *mb.MSPPrincipal) (string, error) {
	switch principal.PrincipalClassification {
	case mb.MSPPrincipal_ROLE:
		var res strings.Builder

		role := &mb.MSPRole{}

		err := proto.Unmarshal(principal.Principal, role)
		if err != nil {
			return "", err
		}

		res.WriteString("'")
		res.WriteString(role.MspIdentifier)
		res.WriteString(".")
		res.WriteString(strings.ToLower(role.Role.String()))
		res.WriteString("'")

		return res.String(), nil
		// TODO: currently fabric only support string to principle convertion for
		// type ROLE. Implement MSPPrinciple to String for types ORGANIZATION_UNIT,
		// IDENTITY, ANONYMITY, and GOMBINED once we have support from fabric.
	case mb.MSPPrincipal_ORGANIZATION_UNIT:
		return "", nil
	case mb.MSPPrincipal_IDENTITY:
		return "", nil
	case mb.MSPPrincipal_ANONYMITY:
		return "", nil
	case mb.MSPPrincipal_COMBINED:
		return "", nil
	default:
		return "", fmt.Errorf("unknown MSP principal classiciation %v", principal.PrincipalClassification)
	}
}

// signaturePolicyToString recursively converts a *cb.SignaturePolicy to a
// string representation.
func signaturePolicyToString(sig *cb.SignaturePolicy, IDs []string) (string, error) {
	switch sig.Type.(type) {
	case *cb.SignaturePolicy_NOutOf_:
		nOutOf := sig.GetNOutOf()

		var policies []string

		var res strings.Builder

		// get gate values
		gate := policydsl.GateOutOf
		if nOutOf.N == 1 {
			gate = policydsl.GateOr
		}

		if nOutOf.N == int32(len(nOutOf.Rules)) {
			gate = policydsl.GateAnd
		}

		if gate == policydsl.GateOutOf {
			policies = append(policies, strconv.Itoa(int(nOutOf.N)))
		}

		// get subpolicies recursively
		for _, rule := range nOutOf.Rules {
			subPolicy, err := signaturePolicyToString(rule, IDs)
			if err != nil {
				return "", err
			}

			policies = append(policies, subPolicy)
		}

		res.WriteString(strings.ToUpper(gate))
		res.WriteString("(")
		res.WriteString(strings.Join(policies, ", "))
		res.WriteString(")")

		return res.String(), nil
	case *cb.SignaturePolicy_SignedBy:
		return IDs[sig.GetSignedBy()], nil
	default:
		return "", fmt.Errorf("unknown signature policy type %v", sig.Type)
	}
}

// TODO: evaluate if modPolicy actually needs to be passed in if all callers pass AdminsPolicyKey.
func addPolicies(cg *cb.ConfigGroup, policyMap map[string]Policy, modPolicy string) error {
	if policyMap == nil {
		return errors.New("no policies defined")
	}

	if _, ok := policyMap[AdminsPolicyKey]; !ok {
		return errors.New("no Admins policy defined")
	}

	if _, ok := policyMap[ReadersPolicyKey]; !ok {
		return errors.New("no Readers policy defined")
	}

	if _, ok := policyMap[WritersPolicyKey]; !ok {
		return errors.New("no Writers policy defined")
	}

	for policyName, policy := range policyMap {
		err := addPolicy(cg, modPolicy, policyName, policy)
		if err != nil {
			return err
		}
	}

	return nil
}

func addPolicy(cg *cb.ConfigGroup, modPolicy, policyName string, policy Policy) error {
	switch policy.Type {
	case ImplicitMetaPolicyType:
		imp, err := implicitMetaFromString(policy.Rule)
		if err != nil {
			return fmt.Errorf("invalid implicit meta policy rule: '%s': %v", policy.Rule, err)
		}

		implicitMetaPolicy, err := proto.Marshal(imp)
		if err != nil {
			return fmt.Errorf("marshaling implicit meta policy: %v", err)
		}

		cg.Policies[policyName] = &cb.ConfigPolicy{
			ModPolicy: modPolicy,
			Policy: &cb.Policy{
				Type:  int32(cb.Policy_IMPLICIT_META),
				Value: implicitMetaPolicy,
			},
		}
	case SignaturePolicyType:
		sp, err := policydsl.FromString(policy.Rule)
		if err != nil {
			return fmt.Errorf("invalid signature policy rule: '%s': %v", policy.Rule, err)
		}

		signaturePolicy, err := proto.Marshal(sp)
		if err != nil {
			return fmt.Errorf("marshaling signature policy: %v", err)
		}

		cg.Policies[policyName] = &cb.ConfigPolicy{
			ModPolicy: modPolicy,
			Policy: &cb.Policy{
				Type:  int32(cb.Policy_SIGNATURE),
				Value: signaturePolicy,
			},
		}
	default:
		return fmt.Errorf("unknown policy type: %s", policy.Type)
	}

	return nil
}

// removePolicy removes an existing policy from an group key organization.
func removePolicy(configGroup *cb.ConfigGroup, policyName string, policies map[string]Policy) error {
	_, exists := policies[policyName]
	if !exists {
		return fmt.Errorf("could not find policy '%s'", policyName)
	}

	delete(configGroup.Policies, policyName)

	return nil
}

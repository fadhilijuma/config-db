package kubernetes

import (
	"strings"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/commons/logger"
	v1 "github.com/flanksource/config-db/api/v1"
	"github.com/flanksource/ketall"
	"github.com/flanksource/ketall/options"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type KubernetesScraper struct {
}

const ExternalTypePrefix = "Kubernetes::"

// Scrape ...
func (kubernetes KubernetesScraper) Scrape(ctx *v1.ScrapeContext, configs v1.ConfigScraper) v1.ScrapeResults {

	results := v1.ScrapeResults{}
	for _, config := range configs.Kubernetes {
		if config.ClusterName == "" {
			logger.Fatalf("clusterName missing from kubernetes configuration")
		}

		opts := options.NewDefaultCmdOptions()
		opts = updateOptions(opts, config)

		objs := ketall.KetAll(opts)

		// {Namespace: {Kind: {Name: ID}}}
		resourceIDMap := make(map[string]map[string]map[string]string)

		for _, obj := range objs {
			if collections.Contains([]string{"Namespace", "Deployment", "Node"}, obj.GetKind()) {
				if resourceIDMap[obj.GetNamespace()] == nil {
					resourceIDMap[obj.GetNamespace()] = make(map[string]map[string]string)
				}
				if resourceIDMap[obj.GetNamespace()][obj.GetKind()] == nil {
					resourceIDMap[obj.GetNamespace()][obj.GetKind()] = make(map[string]string)
				}
				resourceIDMap[obj.GetNamespace()][obj.GetKind()][obj.GetName()] = string(obj.GetUID())
			}
		}

		// Add Cluster object first
		clusterID := "Kubernetes/Cluster/" + config.ClusterName
		results = append(results, v1.ScrapeResult{
			BaseScraper:  config.BaseScraper,
			Name:         config.ClusterName,
			Type:         "Cluster",
			ExternalType: ExternalTypePrefix + "Cluster",
			Config:       make(map[string]string),
			ID:           clusterID,
		})

		resourceIDMap[""]["Cluster"] = make(map[string]string)
		resourceIDMap[""]["Cluster"][config.ClusterName] = clusterID
		// For shorthand
		resourceIDMap[""]["Cluster"]["selfRef"] = clusterID

		for _, obj := range objs {
			var relationships v1.RelationshipResults
			if obj.GetKind() == "Pod" {
				spec := obj.Object["spec"].(map[string]interface{})
				if spec["nodeName"] != nil {
					nodeName := spec["nodeName"].(string)
					nodeID := resourceIDMap[""]["Node"][nodeName]
					relationships = append(relationships, v1.RelationshipResult{
						ConfigExternalID: v1.ExternalID{
							ExternalID:   []string{string(obj.GetUID())},
							ExternalType: ExternalTypePrefix + "Pod",
						},
						RelatedExternalID: v1.ExternalID{
							ExternalID:   []string{nodeID},
							ExternalType: ExternalTypePrefix + "Node",
						},
						Relationship: "NodePod",
					})
				}
			}
			createdAt := obj.GetCreationTimestamp().Time
			parentType, parentExternalID := getKubernetesParent(obj, resourceIDMap)
			results = append(results, v1.ScrapeResult{
				BaseScraper:         config.BaseScraper,
				Name:                obj.GetName(),
				Namespace:           obj.GetNamespace(),
				Type:                obj.GetKind(),
				ExternalType:        ExternalTypePrefix + obj.GetKind(),
				CreatedAt:           &createdAt,
				Config:              *obj,
				ID:                  string(obj.GetUID()),
				Aliases:             getKubernetesAlias(obj),
				ParentExternalID:    parentExternalID,
				ParentExternalType:  ExternalTypePrefix + parentType,
				RelationshipResults: relationships,
			})

		}
	}
	return results
}

func getKubernetesParent(obj *unstructured.Unstructured, resourceIDMap map[string]map[string]map[string]string) (string, string) {
	var parentExternalID, parentConfigType string

	// This will work for pods and replicasets
	if len(obj.GetOwnerReferences()) > 0 {
		ref := obj.GetOwnerReferences()[0]
		if obj.GetKind() == "Pod" {
			// We want pod's parents as Deployments
			if ref.Kind == "ReplicaSet" {
				deployName := extractDeployNameFromReplicaSet(ref.Name)
				parentConfigType = "Deployment"
				parentExternalID = resourceIDMap[obj.GetNamespace()]["Deployment"][deployName]
				return parentConfigType, parentExternalID
			}
		}
		parentConfigType = ref.Kind
		parentExternalID = string(ref.UID)
		return parentConfigType, parentExternalID
	}

	if obj.GetNamespace() != "" {
		parentConfigType = "Namespace"
		parentExternalID = resourceIDMap[""]["Namespace"][obj.GetNamespace()]
		return parentConfigType, parentExternalID
	}

	// Everything which is not namespaced should be mapped to cluster
	parentConfigType = "Cluster"
	parentExternalID = resourceIDMap[""]["Cluster"]["selfRef"]
	return parentConfigType, parentExternalID
}

func getKubernetesAlias(obj *unstructured.Unstructured) []string {
	return []string{strings.Join([]string{"Kubernetes", obj.GetKind(), obj.GetNamespace(), obj.GetName()}, "/")}
}

func updateOptions(opts *options.KetallOptions, config v1.Kubernetes) *options.KetallOptions {
	opts.AllowIncomplete = config.AllowIncomplete
	opts.Namespace = config.Namespace
	opts.Scope = config.Scope
	opts.Selector = config.Selector
	opts.FieldSelector = config.FieldSelector
	opts.UseCache = config.UseCache
	opts.MaxInflight = config.MaxInflight
	opts.Exclusions = config.Exclusions
	opts.Since = config.Since
	//TODO: update kubeconfig reference if provided by user
	// if config.Kubeconfig != nil {
	// 	opts.Kubeconfig = config.Kubeconfig.GetValue()
	// }
	return opts
}

func extractDeployNameFromReplicaSet(rs string) string {
	split := strings.Split(rs, "-")
	split = split[:len(split)-1]
	return strings.Join(split, "-")
}

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flowcontrol

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// These are valid wildcards.
const (
	APIGroupAll    = "*"
	ResourceAll    = "*"
	VerbAll        = "*"
	NonResourceAll = "*"

	NameAll = "*"
)

// System preset priority level names
const (
	PriorityLevelConfigurationNameExempt = "exempt"
)

// Conditions
const (
	FlowSchemaConditionDangling = "Dangling"

	PriorityLevelConfigurationConditionConcurrencyShared = "ConcurrencyShared"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlowSchema defines the schema of a group of flows. Note that a flow is made up of a set of inbound API requests with
// similar attributes and is identified by a pair of strings: the name of the FlowSchema and a "flow distinguisher".
type FlowSchema struct {
	metav1.TypeMeta
	// `metadata` is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta
	// `spec` is the specification of the desired behavior of a FlowSchema.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec FlowSchemaSpec
	// `status` is the current status of a FlowSchema.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status FlowSchemaStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlowSchemaList is a list of FlowSchema objects.
type FlowSchemaList struct {
	metav1.TypeMeta
	// `metadata` is the standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta

	// `items` is a list of FlowSchemas.
	// +listType=set
	Items []FlowSchema
}

// FlowSchemaSpec describes how the FlowSchema's specification looks like.
type FlowSchemaSpec struct {
	// `priorityLevelConfiguration` should reference a PriorityLevelConfiguration in the cluster. If the reference cannot
	// be resolved, the FlowSchema will be ignored and marked as invalid in its status.
	// Required.
	PriorityLevelConfiguration PriorityLevelConfigurationReference
	// `matchingPrecedence` is used to choose among the FlowSchemas that match a given request. The chosen
	// FlowSchema is among those with the numerically lowest (which we take to be logically highest)
	// MatchingPrecedence.  Each MatchingPrecedence value must be non-negative.
	// Note that if the precedence is not specified or zero, it will be set to 1000 as default.
	// +optional
	MatchingPrecedence int32
	// `distinguisherMethod` defines how to compute the flow distinguisher for requests that match this schema.
	// `nil` specifies that the distinguisher is disabled and thus will always be the empty string.
	// +optional
	DistinguisherMethod *FlowDistinguisherMethod
	// `rules` describes which requests will match this flow schema. This FlowSchema matches a request if and only if
	// at least one member of rules matches the request.
	// if it is an empty slice, there will be no requests matching the FlowSchema.
	// +listType=set
	// +optional
	Rules []PolicyRulesWithSubjects
}

// FlowDistinguisherMethodType is the type of flow distinguisher method
type FlowDistinguisherMethodType string

// These are valid flow-distinguisher methods.
const (
	// FlowDistinguisherMethodByUserType specifies that the flow distinguisher is the username in the request.
	// This type is used to provide some insulation between users.
	FlowDistinguisherMethodByUserType FlowDistinguisherMethodType = "ByUser"

	// FlowDistinguisherMethodByNamespaceType specifies that the flow distinguisher is the namespace of the
	// object that the request acts upon. If the object is not namespaced, or if the request is a non-resource
	// request, then the distinguisher will be the empty string. An example usage of this type is to provide
	// some insulation between tenants in a situation where there are multiple tenants and each namespace
	// is dedicated to a tenant.
	FlowDistinguisherMethodByNamespaceType FlowDistinguisherMethodType = "ByNamespace"
)

// FlowDistinguisherMethod specifies the method of a flow distinguisher.
type FlowDistinguisherMethod struct {
	// `type` is the type of flow distinguisher method
	// The supported types are "ByUser" and "ByNamespace".
	// Required.
	Type FlowDistinguisherMethodType
}

// PriorityLevelConfigurationReference contains information that points to the "request-priority" being used.
type PriorityLevelConfigurationReference struct {
	// `name` is the name of the priority level configuration being referenced
	// Required.
	Name string
}

// PolicyRulesWithSubjects prescribes a test that applies to a request to an apiserver. The test considers the subject
// making the request, the verb being requested, and the resource to be acted upon. This PolicyRulesWithSubjects matches
// a request if and only if both (a) at least one member of subjects matches the request and (b) at least one member
// of resourceRules or nonResourceRules matches the request.
type PolicyRulesWithSubjects struct {
	// subjects is the list of normal user, serviceaccount, or group that this rule cares about.
	// There must be at least one member in this slice.
	// A slice that includes both the system:authenticated and system:unauthenticated user groups matches every request.
	// +listType=set
	// Required.
	Subjects []Subject
	// `resourceRules` is a slice of ResourcePolicyRules that identify matching requests according to their verb and the
	// target resource.
	// At least one of `resourceRules` and `nonResourceRules` has to be non-empty.
	// +listType=set
	// +optional
	ResourceRules []ResourcePolicyRule
	// `nonResourceRules` is a list of NonResourcePolicyRules that identify matching requests according to their verb
	// and the target non-resource URL.
	// +listType=set
	// +optional
	NonResourceRules []NonResourcePolicyRule
}

// Subject matches the originator of a request, as identified by the request authentication system. There are three
// ways of matching an originator; by user, group, or service account.
// +union
type Subject struct {
	// Required
	// +unionDiscriminator
	Kind SubjectKind
	// +optional
	User *UserSubject
	// +optional
	Group *GroupSubject
	// +optional
	ServiceAccount *ServiceAccountSubject
}

// SubjectKind is the kind of subject.
type SubjectKind string

// Supported subject's kinds.
const (
	SubjectKindUser           SubjectKind = "User"
	SubjectKindGroup          SubjectKind = "Group"
	SubjectKindServiceAccount SubjectKind = "ServiceAccount"
)

// UserSubject holds detailed information for user-kind subject.
type UserSubject struct {
	// `name` is the username that matches, or "*" to match all usernames.
	// Required.
	Name string
}

// GroupSubject holds detailed information for group-kind subject.
type GroupSubject struct {
	// name is the user group that matches, or "*" to match all user groups.
	// See https://github.com/kubernetes/apiserver/blob/master/pkg/authentication/user/user.go for some
	// well-known group names.
	// Required.
	Name string
}

// ServiceAccountSubject holds detailed information for service-account-kind subject.
type ServiceAccountSubject struct {
	// `namespace` is the namespace of matching ServiceAccount objects.
	// Required.
	Namespace string
	// `name` is the name of matching ServiceAccount objects, or "*" to match regardless of name.
	// Required.
	Name string
}

// ResourcePolicyRule is a predicate that matches some resource requests, testing the request's verb and the target
// resource. A ResourcePolicyRule matches a request if and only if: (a) at least one member
// of verbs matches the request, (b) at least one member of apiGroups matches the request, and (c) at least one member
// of resources matches the request.
type ResourcePolicyRule struct {
	// `verbs` is a list of matching verbs and may not be empty.
	// "*" matches all verbs. if it is present, it must be the only entry.
	// +listType=set
	// Required.
	Verbs []string
	// `apiGroups` is a list of matching API groups and may not be empty.
	// "*" matches all api-groups. if it is present, it must be the only entry.
	// +listType=set
	// Required.
	APIGroups []string
	// `resources` is a list of matching resources (i.e., lowercase and plural) with, if desired, subresource.
	// For example, [ "services", "nodes/status" ].
	// This list may not be empty.
	// "*" matches all resources. if it is present, it must be the only entry.
	// +listType=set
	// Required.
	Resources []string
}

// NonResourcePolicyRule is a predicate that matches non-resource requests according to their verb and the
// target non-resource URL. A NonResourcePolicyRule matches a request if and only if both (a) at least one member
// of verbs matches the request and (b) at least one member of nonResourceURLs matches the request.
type NonResourcePolicyRule struct {
	// `verbs` is a list of matching verbs and may not be empty.
	// "*" matches all verbs. If it is present, it must be the only entry.
	// +listType=set
	// Required.
	Verbs []string
	// `nonResourceURLs` is a set of url prefixes that a user should have access to and may not be empty.
	// For example:
	//   - "/healthz" is legal
	//   - "/hea*" is illegal
	//   - "/hea" is legal but matches nothing
	//   - "/hea/*" also matches nothing
	//   - "/healthz/*" matches all per-component health checks.
	// "*" matches all non-resource urls. if it is present, it must be the only entry.
	// +listType=set
	// Required.
	NonResourceURLs []string
}

// FlowSchemaStatus represents the current state of a FlowSchema.
type FlowSchemaStatus struct {
	// `conditions` is a list of the current states of FlowSchema.
	// +listType=associative
	// +listMapKey=type
	// +optional
	Conditions []FlowSchemaCondition
}

// FlowSchemaCondition describes conditions for a FlowSchema.
type FlowSchemaCondition struct {
	// `type` is the type of the condition.
	// Required.
	Type FlowSchemaConditionType
	// `status` is the status of the condition.
	// Can be True, False, Unknown.
	// Required.
	Status ConditionStatus
	// `lastTransitionTime` is the last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time
	// `reason` is a unique, one-word, CamelCase reason for the condition's last transition.
	Reason string
	// `message` is a human-readable message indicating details about last transition.
	Message string
}

// FlowSchemaConditionType is a valid value for FlowSchemaStatusCondition.Type
type FlowSchemaConditionType string

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityLevelConfiguration represents the configuration of a priority level.
type PriorityLevelConfiguration struct {
	metav1.TypeMeta
	// `metadata` is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta
	// `spec` is the specification of the desired behavior of a "request-priority".
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PriorityLevelConfigurationSpec
	// `status` is the current status of a "request-priority".
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status PriorityLevelConfigurationStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PriorityLevelConfigurationList is a list of PriorityLevelConfiguration objects.
type PriorityLevelConfigurationList struct {
	metav1.TypeMeta
	// `metadata` is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta
	// `items` is a list of request-priorities.
	// +listType=set
	Items []PriorityLevelConfiguration
}

// PriorityLevelConfigurationSpec is specification of a priority level
type PriorityLevelConfigurationSpec struct {
	// `type` indicates whether this priority level does
	// queuing or is exempt.  Valid values are "Queuing" and "Exempt".
	// "Exempt" means that requests of this priority level are not subject
	// to concurrency limits (and thus are never queued) and do not detract
	// from the concurrency available for non-exempt requests. The "Exempt"
	// type is useful for apiserver self-requests and system administrator use.
	// Required.
	Type PriorityLevelQueueingType

	// `queuing` holds the configuration parameters that are
	// only meaningful for a priority level that does queuing (i.e.,
	// is not exempt).  This field must be non-empty if and only if
	// `queuingType` is `"Queuing"`.
	// +optional
	Queuing *QueuingConfiguration
}

// PriorityLevelQueueingType identifies the queuing nature of a priority level
type PriorityLevelQueueingType string

// Supported queuing types.
const (
	// PriorityLevelQueuingTypeQueueing is the PriorityLevelQueueingType for priority levels that queue
	PriorityLevelQueuingTypeQueueing PriorityLevelQueueingType = "Queuing"

	// PriorityLevelQueuingTypeExempt is the PriorityLevelQueueingType for priority levels that are exempt from concurrency controls
	PriorityLevelQueuingTypeExempt PriorityLevelQueueingType = "Exempt"
)

// QueuingConfiguration holds the configuration parameters that are specific to a priority level that is subject to concurrency controls
type QueuingConfiguration struct {
	// `assuredConcurrencyShares` (ACS) must be a positive number. The
	// server's concurrency limit (SCL) is divided among the
	// concurrency-controlled priority levels in proportion to their
	// assured concurrency shares. This produces the assured
	// concurrency value (ACV) for each such priority level:
	//
	//             ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
	//
	// bigger numbers of ACS mean more reserved concurrent requests (at the
	// expense of every other PL).
	// This field has a default value of 30.
	// +optional
	AssuredConcurrencyShares int32

	// `queues` is the number of queues for this priority level. The
	// queues exist independently at each apiserver. The value must be
	// positive.  Setting it to 1 effectively precludes
	// shufflesharding and thus makes the distinguisher method of
	// associated flow schemas irrelevant.  This field has a default
	// value of 64.
	// +optional
	Queues int32

	// `handSize` is a small positive number that configures the
	// shuffle sharding of requests into queues.  When enqueuing a request
	// at this priority level the request's flow identifier (a string
	// pair) is hashed and the hash value is used to shuffle the list
	// of queues and deal a hand of the size specified here.  The
	// request is put into one of the shortest queues in that hand.
	// `handSize` must be no larger than `queues`, and should be
	// significantly smaller (so that a few heavy flows do not
	// saturate most of the queues).  See the user-facing
	// documentation for more extensive guidance on setting this
	// field.  This field has a default value of 8.
	// +optional
	HandSize int32

	// `queueLengthLimit` is the maximum number of requests allowed to
	// be waiting in a given queue of this priority level at a time;
	// excess requests are rejected.  This value must be positive.  If
	// not specified, it will be defaulted to 50.
	// +optional
	QueueLengthLimit int32
}

// PriorityLevelConfigurationConditionType is a valid value for PriorityLevelConfigurationStatusCondition.Type
type PriorityLevelConfigurationConditionType string

// PriorityLevelConfigurationStatus represents the current state of a "request-priority".
type PriorityLevelConfigurationStatus struct {
	// `conditions` is the current state of "request-priority".
	// +listType=associative
	// +listMapKey=type
	// +optional
	Conditions []PriorityLevelConfigurationCondition
}

// PriorityLevelConfigurationCondition defines the condition of priority level.
type PriorityLevelConfigurationCondition struct {
	// `type` is the type of the condition.
	// Required.
	Type PriorityLevelConfigurationConditionType
	// `status` is the status of the condition.
	// Can be True, False, Unknown.
	// Required.
	Status ConditionStatus
	// `lastTransitionTime` is the last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time
	// `reason` is a unique, one-word, CamelCase reason for the condition's last transition.
	Reason string
	// `message` is a human-readable message indicating details about last transition.
	Message string
}

// ConditionStatus is the status of the condition.
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

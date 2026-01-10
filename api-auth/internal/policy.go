package internal

import (
	"encoding/json"
	"fmt"
	"github.com/burkeclove/auth-api/models/requests"
)

// IAMPolicy represents a MinIO/S3 IAM policy
type IAMPolicy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Statement represents a single policy statement
type Statement struct {
	Effect    string                 `json:"Effect"`
	Action    interface{}            `json:"Action"` // string or []string
	Resource  interface{}            `json:"Resource"` // string or []string
	Condition map[string]interface{} `json:"Condition,omitempty"`
}

// BuildIAMPolicy creates an IAM policy from the SigV4 request
func BuildIAMPolicy(req *requests.CreateSigV4Request) (*IAMPolicy, error) {
	policy := &IAMPolicy{
		Version:   "2012-10-17",
		Statement: []Statement{},
	}

	// Add global actions if any
	if len(req.GlobalActions) > 0 {
		stmt := Statement{
			Effect:   "Allow",
			Action:   req.GlobalActions,
			Resource: "*",
		}

		// Add conditions if specified
		if req.Conditions != nil {
			stmt.Condition = buildConditions(req.Conditions)
		}

		policy.Statement = append(policy.Statement, stmt)
	}

	// Add bucket-specific permissions
	for _, bucket := range req.Buckets {
		// Handle bucket-level actions (ListBucket, etc.)
		bucketResource := fmt.Sprintf("arn:clove:s3:::%s", bucket.BucketID)

		// Handle object-level actions (GetObject, PutObject, etc.)
		var objectResources []string
		if len(bucket.Prefixes) > 0 {
			for _, prefix := range bucket.Prefixes {
				objectResources = append(objectResources, fmt.Sprintf("arn:clove:s3:::%s/%s*", bucket.BucketID, prefix))
			}
		} else {
			objectResources = append(objectResources, fmt.Sprintf("arn:clove:s3:::%s/*", bucket.BucketID))
		}

		stmt := Statement{
			Effect: "Allow",
			Action: bucket.Actions,
		}

		// Combine bucket and object resources
		resources := []string{bucketResource}
		resources = append(resources, objectResources...)
		stmt.Resource = resources

		// Add conditions if specified
		if req.Conditions != nil {
			stmt.Condition = buildConditions(req.Conditions)
		}

		policy.Statement = append(policy.Statement, stmt)
	}

	return policy, nil
}

// buildConditions creates IAM policy conditions from request conditions
func buildConditions(cond *requests.Conditions) map[string]interface{} {
	conditions := make(map[string]interface{})

	if len(cond.IPAllowlist) > 0 {
		conditions["IpAddress"] = map[string]interface{}{
			"clove:SourceIp": cond.IPAllowlist,
		}
	}

	if cond.MaxObjectSize > 0 {
		conditions["NumericLessThanEquals"] = map[string]interface{}{
			"s3:content-length": cond.MaxObjectSize,
		}
	}

	return conditions
}

// PolicyToJSON converts an IAM policy to JSON string
func PolicyToJSON(policy *IAMPolicy) (string, error) {
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	return string(policyBytes), nil
}

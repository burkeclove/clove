package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type IAMPolicy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Effect    string      `json:"Effect"`
	Action    interface{} `json:"Action"`    // string or []string
	Resource  interface{} `json:"Resource"`  // string or []string
	Condition map[string]interface{} `json:"Condition,omitempty"`
}

func GetPolicyFromContext(c *gin.Context) (*IAMPolicy, error) {
	policyJSON, exists := c.Get("policy")
	if !exists {
		return nil, fmt.Errorf("policy not found in context")
	}

	policyStr, ok := policyJSON.(string)
	if !ok {
		return nil, fmt.Errorf("policy is not a string")
	}

	var policy IAMPolicy
	if err := json.Unmarshal([]byte(policyStr), &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	return &policy, nil
}

func IsActionAllowed(c *gin.Context, action string, resource string) (bool, error) {
	policy, err := GetPolicyFromContext(c)
	if err != nil {
		return false, err
	}

	for _, stmt := range policy.Statement {
		if stmt.Effect != "Allow" {
			continue
		}

		if !matchesAction(stmt.Action, action) {
			continue
		}

		if matchesResource(stmt.Resource, resource) {
			return true, nil
		}
	}

	return false, nil
}

func matchesAction(stmtAction interface{}, requestedAction string) bool {
	switch v := stmtAction.(type) {
	case string:
		return matchesPattern(v, requestedAction)
	case []interface{}:
		for _, action := range v {
			if actionStr, ok := action.(string); ok {
				if matchesPattern(actionStr, requestedAction) {
					return true
				}
			}
		}
	}
	return false
}

func matchesResource(stmtResource interface{}, requestedResource string) bool {
	switch v := stmtResource.(type) {
	case string:
		return matchesPattern(v, requestedResource)
	case []interface{}:
		for _, res := range v {
			if resStr, ok := res.(string); ok {
				if matchesPattern(resStr, requestedResource) {
					return true
				}
			}
		}
	}
	return false
}

func matchesPattern(pattern string, str string) bool {
	if pattern == str {
		return true
	}

	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(str, prefix)
	}

	return false
}

func RequireAction(action string, resourceFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		resource := resourceFunc(c)

		allowed, err := IsActionAllowed(c, action, resource)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "Failed to evaluate policy"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{
				"error": fmt.Sprintf("Action %s not allowed on resource %s", action, resource),
			})
			return
		}

		c.Next()
	}
}

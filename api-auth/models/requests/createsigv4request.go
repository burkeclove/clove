package requests

type CreateSigV4Request struct {
	OrgID         string      `json:"org_id" binding:"required"`
	GlobalActions []string    `json:"global_actions"`
	Buckets       []BucketScope `json:"buckets"`
	Conditions    *Conditions `json:"conditions"`
}

type BucketScope struct {
	BucketID string   `json:"bucket_id" binding:"required"`
	Actions  []string `json:"actions" binding:"required"`
	Prefixes []string `json:"prefixes"`
}

type Conditions struct {
	IPAllowlist   []string `json:"ip_allowlist"`
	MaxObjectSize int64    `json:"max_object_size"`
}

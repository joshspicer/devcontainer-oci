package main

type LockFileItem struct {
	Target      string
	Ref         string
	Registry    string
	Digest      string
	Annotations map[string]string
	MediaType   string
}

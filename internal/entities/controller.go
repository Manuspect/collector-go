package entities

import "time"

type InfoFile struct {
	Error bool `json:"error"`
	Info  Info `json:"info"`
	Msg   any  `json:"msg"`
}
type Info struct {
	Bucket           string    `json:"bucket"`
	Key              string    `json:"key"`
	ETag             string    `json:"e_tag"`
	Size             int       `json:"size"`
	LastModified     time.Time `json:"last_modified"`
	Location         string    `json:"location"`
	VersionID        string    `json:"version_id"`
	Expiration       time.Time `json:"expiration"`
	ExpirationRuleID string    `json:"expiration_rule_id"`
	ChecksumCRC32    string    `json:"checksum_CRC32"`
	ChecksumCRC32C   string    `json:"checksum_CRC32C"`
	ChecksumSHA1     string    `json:"checksum_SHA1"`
	ChecksumSHA256   string    `json:"checksum_SHA256"`
}

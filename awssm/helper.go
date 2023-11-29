package awssm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// Convert a map to a list of tags
func convertMapToTags(tags map[string]string) []smtypes.Tag {
	var tagList []smtypes.Tag
	for key, value := range tags {
		tag := smtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}
		tagList = append(tagList, tag)
	}
	return tagList
}

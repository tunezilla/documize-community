package storage

import (
	"github.com/documize/community/core/env"
	"github.com/documize/community/domain/attachment"
)

func objectStorer(r *env.Runtime) attachment.ObjectStorer {
	var (
		objectStorer attachment.ObjectStorer
		err          error
	)

	if r.Flags.MinioEndpoint == "" {
		objectStorer, err = attachment.S3Storer(r.Flags.Bucket)
	} else {
		objectStorer, err = attachment.MinioStorer(r.Flags.Bucket, r.Flags.MinioEndpoint)
	}

	if err != nil {
		// doesn't seem to be a simple way to bubble this upwards, so just panic
		panic(err)
	}

	return objectStorer
}

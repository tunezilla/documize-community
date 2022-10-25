package attachment

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/documize/community/domain"
	"github.com/documize/community/model/attachment"
)

type ObjectStorer interface {
	Get(ctx context.Context, rctx domain.RequestContext, a attachment.Attachment) ([]byte, error)
	Put(ctx context.Context, rctx domain.RequestContext, a attachment.Attachment, data []byte) error
}

func attachmentS3Path(ctx domain.RequestContext, a attachment.Attachment) string {
	return fmt.Sprintf("%v/%v/%v", ctx.OrgID, a.RefID, path.Clean(a.Filename))
}

type s3BucketStorer struct {
	Bucket string
	S3     s3iface.S3API
}

func (s *s3BucketStorer) Get(ctx context.Context, rctx domain.RequestContext, a attachment.Attachment) ([]byte, error) {
	filepath := attachmentS3Path(rctx, a)
	res, err := s.S3.GetObject(&s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &filepath,
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *s3BucketStorer) Put(ctx context.Context, rctx domain.RequestContext, a attachment.Attachment, data []byte) error {
	md5Builder := md5.New()
	if _, err := md5Builder.Write(data); err != nil {
		return err
	}
	md5 := base64.StdEncoding.EncodeToString(md5Builder.Sum(nil))
	filepath := attachmentS3Path(rctx, a)

	if _, err := s.S3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:     &s.Bucket,
		Body:       bytes.NewReader(data),
		Key:        &filepath,
		ContentMD5: &md5,
	}); err != nil {
		return err
	}

	return nil
}

func S3Storer(bucket string) (ObjectStorer, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &s3BucketStorer{
		Bucket: bucket,
		S3:     s3.New(sess),
	}, nil
}

func MinioStorer(bucket, minioURL string) (ObjectStorer, error) {
	parsedURL, err := url.Parse(minioURL)
	if err != nil {
		return nil, err
	}

	id := parsedURL.User.Username()
	secret, _ := parsedURL.User.Password()
	parsedURL.User = nil
	endpoint := parsedURL.String()

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Endpoint:         &endpoint,
		Region:           aws.String("us-west-2"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return &s3BucketStorer{
		Bucket: bucket,
		S3:     s3.New(sess),
	}, nil
}

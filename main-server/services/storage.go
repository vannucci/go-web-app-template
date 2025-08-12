package services

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type StorageBackend interface {
    Upload(ctx context.Context, file io.Reader, key string) error
    Download(ctx context.Context, key string) (io.Reader, error)
    Delete(ctx context.Context, key string) error
    GeneratePresignedURL(key string, expiry time.Duration) (string, error)
}

// S3 Storage Implementation
type S3Storage struct {
    client   *s3.S3
    uploader *s3manager.Uploader
    bucket   string
}

func NewS3Storage(region, bucket string) (*S3Storage, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        return nil, err
    }
    
    return &S3Storage{
        client:   s3.New(sess),
        uploader: s3manager.NewUploader(sess),
        bucket:   bucket,
    }, nil
}

func (s *S3Storage) Upload(ctx context.Context, file io.Reader, key string) error {
    _, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
        Body:   file,
    })
    return err
}

func (s *S3Storage) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
    req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })
    
    urlStr, err := req.Presign(expiry)
    if err != nil {
        return "", err
    }
    
    return urlStr, nil
}

// Local Storage Implementation (for development)
type LocalStorage struct {
    basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
    os.MkdirAll(basePath, 0755)
    return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Upload(ctx context.Context, file io.Reader, key string) error {
    path := filepath.Join(l.basePath, key)
    os.MkdirAll(filepath.Dir(path), 0755)
    
    out, err := os.Create(path)
    if err != nil {
        return err
    }
    defer out.Close()
    
    _, err = io.Copy(out, file)
    return err
}

func (l *LocalStorage) Download(ctx context.Context, key string) (io.Reader, error) {
    path := filepath.Join(l.basePath, key)
    return os.Open(path)
}

func (l *LocalStorage) Delete(ctx context.Context, key string) error {
    path := filepath.Join(l.basePath, key)
    return os.Remove(path)
}

func (l *LocalStorage) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
    // For local storage, just return a local URL
    return fmt.Sprintf("/files/%s", key), nil
}

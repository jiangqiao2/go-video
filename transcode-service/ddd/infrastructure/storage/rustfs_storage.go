package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"transcode-service/ddd/domain/gateway"
	"transcode-service/pkg/logger"
)

type RustFSStorage struct {
	endpoint string
	access   string
	secret   string
	region   string
}

func NewRustFSStorage(endpoint, access, secret string) gateway.StorageGateway {
	return &RustFSStorage{endpoint: normalizeEndpoint(endpoint), access: access, secret: secret, region: "us-east-1"}
}

func (s *RustFSStorage) UploadTranscodedFile(ctx context.Context, localPath, objectKey, contentType string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open local file: %w", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	hash, err := sha256FileHex(localPath)
	if err != nil {
		return "", err
	}
	url := s.s3URL("transcode", objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, f)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-amz-content-sha256", hash)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	req.ContentLength = stat.Size()
	s.signS3(req, hash)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("put object failed: status=%d, body=%s", resp.StatusCode, string(b))
	}
	logger.Info("RustFS uploaded file", map[string]interface{}{"object_key": objectKey, "local_path": localPath})
	return objectKey, nil
}

func (s *RustFSStorage) UploadObjects(ctx context.Context, objects []gateway.UploadObject) error {
	for _, obj := range objects {
		if _, err := s.UploadTranscodedFile(ctx, obj.LocalPath, obj.ObjectKey, obj.ContentType); err != nil {
			return err
		}
	}
	return nil
}

func (s *RustFSStorage) DownloadFile(ctx context.Context, objectKey, localPath string) error {
	bucket := inferBucketFromKey(objectKey)
	url := s.s3URL(bucket, objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	s.signS3(req, "UNSIGNED-PAYLOAD")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get object failed: status=%d, body=%s", resp.StatusCode, string(b))
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	logger.Info("RustFS downloaded file", map[string]interface{}{"object_key": objectKey, "local_path": localPath})
	return nil
}

func (s *RustFSStorage) s3URL(bucket, key string) string {
	k := strings.TrimLeft(key, "/")
	return fmt.Sprintf("%s/%s/%s", s.endpoint, bucket, k)
}

func (s *RustFSStorage) signS3(req *http.Request, payloadHash string) {
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	date := t.Format("20060102")
	req.Header.Set("x-amz-date", amzDate)

	u, _ := neturl.Parse(req.URL.String())
	host := u.Host
	req.Header.Set("host", host)

	signed := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if req.Header.Get("content-type") != "" {
		signed = append(signed, "content-type")
	}
	sort.Strings(signed)

	var canonicalHeaders strings.Builder
	for _, h := range signed {
		canonicalHeaders.WriteString(h)
		canonicalHeaders.WriteString(":")
		if h == "host" {
			canonicalHeaders.WriteString(strings.TrimSpace(host))
		} else {
			canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(h)))
		}
		canonicalHeaders.WriteString("\n")
	}
	canonicalURI := u.Path
	canonicalQuery := u.RawQuery
	signedHeaders := strings.Join(signed, ";")
	cr := strings.Join([]string{req.Method, canonicalURI, canonicalQuery, canonicalHeaders.String(), signedHeaders, payloadHash}, "\n")
	crHash := sha256Hex([]byte(cr))

	scope := strings.Join([]string{date, s.region, "s3", "aws4_request"}, "/")
	sts := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, crHash}, "\n")
	kDate := hmacSHA256([]byte("AWS4"+s.secret), date)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	sig := hex.EncodeToString(hmacSHA256(kSigning, sts))
	auth := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s", s.access, scope, signedHeaders, sig)
	req.Header.Set("Authorization", auth)
}

func normalizeEndpoint(e string) string {
	e = strings.TrimSpace(e)
	if e == "" {
		return "http://localhost:9000"
	}
	if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") {
		return strings.TrimRight(e, "/")
	}
	return "http://" + strings.TrimRight(e, "/")
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func sha256FileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	d := sha256.New()
	if _, err := io.Copy(d, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(d.Sum(nil)), nil
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func inferBucketFromKey(key string) string {
	k := strings.TrimLeft(key, "/")
	if strings.HasPrefix(k, "uploads/") || strings.HasPrefix(k, "chunks/") {
		return "uploads"
	}
	if strings.HasPrefix(k, "transcoded/") {
		return "transcode"
	}
	return "uploads"
}

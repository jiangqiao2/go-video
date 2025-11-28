package rustfs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/vo"
	"upload-service/internal/resource"
	"upload-service/pkg/assert"
	"upload-service/pkg/logger"
)

var (
	rustfsServiceOnce sync.Once
	singletonRustFS   gateway.MinioService
)

type RustFSServiceImpl struct {
	endpoint string
	access   string
	secret   string
	region   string
}

func DefaultRustFSService() gateway.MinioService {
	assert.NotCircular()
	rustfsServiceOnce.Do(func() {
		r := resource.DefaultRustFSResource()
		singletonRustFS = &RustFSServiceImpl{
			endpoint: normalizeEndpoint(r.GetEndpoint()),
			access:   r.GetAccessKey(),
			secret:   r.GetSecretKey(),
			region:   "us-east-1",
		}
	})
	return singletonRustFS
}

func (s *RustFSServiceImpl) GenerateStoragePath(ctx context.Context, genStoPathVo *vo.GenerateStoragePathVO) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	fileName := genStoPathVo.FileName()
	ext := ""
	if dot := strings.LastIndex(fileName, "."); dot != -1 {
		ext = fileName[dot+1:]
	}
	return fmt.Sprintf("uploads/%s/%s/%s/%s/%s.%s",
		genStoPathVo.UserUUID(), year, month, day, genStoPathVo.UploadVideoUUID(), ext,
	)
}

func (s *RustFSServiceImpl) GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	return fmt.Sprintf("chunks/%s/%s/%s/%s/chunk_", year, month, day, uploadVideoUUID)
}

func (s *RustFSServiceImpl) UploadChunk(ctx context.Context, minIoChunkVo *vo.MinIoUploadChunkVo) error {
	client := s.httpClientFor(ctx)
	buf, err := io.ReadAll(minIoChunkVo.Reader())
	if err != nil {
		return err
	}
	payloadHash := sha256Hex(buf)
	u := s.s3URL(minIoChunkVo.BucketName(), minIoChunkVo.StoragePath())
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	if ct := minIoChunkVo.ContentType(); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	req.ContentLength = int64(len(buf))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	s.signS3(req, payloadHash)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logger.Errorf("rustfs put object failed status=%d body=%s", resp.StatusCode, string(b))
		return fmt.Errorf("put object failed: status=%d", resp.StatusCode)
	}
	return nil
}

func (s *RustFSServiceImpl) MergeChunk(ctx context.Context, mergeChunkVo *vo.MergeChunkVo) error {
	client := s.httpClientFor(ctx)
	bucket := "uploads"
	destKey := mergeChunkVo.StoragePath()
	uploadID, err := s.initiateMultipartUpload(ctx, client, bucket, destKey)
	if err != nil {
		return err
	}
	parts := make([]s3Part, 0, mergeChunkVo.TotalChunks())
	for i := int64(0); i < mergeChunkVo.TotalChunks(); i++ {
		srcKey := fmt.Sprintf("%s%d", mergeChunkVo.ChunkStoragePath(), i)
		etag, err := s.uploadPartCopy(ctx, client, bucket, destKey, uploadID, int(i+1), srcKey)
		if err != nil {
			_ = s.abortMultipartUpload(ctx, client, bucket, destKey, uploadID)
			return err
		}
		parts = append(parts, s3Part{PartNumber: int(i + 1), ETag: etag})
	}
	if err := s.completeMultipartUpload(ctx, client, bucket, destKey, uploadID, parts); err != nil {
		_ = s.abortMultipartUpload(ctx, client, bucket, destKey, uploadID)
		return err
	}
	return nil
}

type s3Part struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type initiateResult struct {
	UploadId string `xml:"UploadId"`
}

type copyPartResult struct {
	ETag string `xml:"ETag"`
}

func (s *RustFSServiceImpl) initiateMultipartUpload(ctx context.Context, client *http.Client, bucket, key string) (string, error) {
	u := s.s3URL(bucket, key) + "?uploads="
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	s.signS3(req, "UNSIGNED-PAYLOAD")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logger.Errorf("rustfs initiate multipart failed status=%d body=%s", resp.StatusCode, string(b))
		return "", fmt.Errorf("initiate multipart failed: status=%d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var r initiateResult
	if err := xml.Unmarshal(b, &r); err != nil {
		return "", err
	}
	if r.UploadId == "" {
		return "", fmt.Errorf("empty upload id")
	}
	return r.UploadId, nil
}

func (s *RustFSServiceImpl) uploadPartCopy(ctx context.Context, client *http.Client, bucket, key, uploadID string, partNumber int, sourceKey string) (string, error) {
	v := neturl.Values{}
	v.Set("partNumber", strconv.Itoa(partNumber))
	v.Set("uploadId", uploadID)
	u := s.s3URL(bucket, key) + "?" + v.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("x-amz-copy-source", fmt.Sprintf("%s/%s", bucket, strings.TrimLeft(sourceKey, "/")))
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	s.signS3(req, "UNSIGNED-PAYLOAD")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		logger.Errorf("rustfs upload part copy failed status=%d body=%s part=%d", resp.StatusCode, string(b), partNumber)
		return "", fmt.Errorf("upload part copy failed: status=%d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var r copyPartResult
	if err := xml.Unmarshal(b, &r); err != nil {
		return "", err
	}
	etag := strings.TrimSpace(r.ETag)
	if etag == "" {
		return "", fmt.Errorf("empty etag")
	}
	return etag, nil
}

func (s *RustFSServiceImpl) completeMultipartUpload(ctx context.Context, client *http.Client, bucket, key, uploadID string, parts []s3Part) error {
	cr := struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []s3Part `xml:"Part"`
	}{Parts: parts}
	b, err := xml.Marshal(cr)
	if err != nil {
		return err
	}
	vv := neturl.Values{}
	vv.Set("uploadId", uploadID)
	u := s.s3URL(bucket, key) + "?" + vv.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(b))
	ph := sha256Hex(b)
	req.Header.Set("x-amz-content-sha256", ph)
	s.signS3(req, ph)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bb, _ := io.ReadAll(resp.Body)
		logger.Errorf("rustfs complete multipart failed status=%d body=%s", resp.StatusCode, string(bb))
		return fmt.Errorf("complete multipart failed: status=%d", resp.StatusCode)
	}
	return nil
}

func (s *RustFSServiceImpl) abortMultipartUpload(ctx context.Context, client *http.Client, bucket, key, uploadID string) error {
	u := s.s3URL(bucket, key) + "?uploadId=" + neturl.QueryEscape(uploadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	s.signS3(req, "UNSIGNED-PAYLOAD")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *RustFSServiceImpl) DeleteChunks(ctx context.Context, chunkStoragePath string, totalChunks int64) error {
	client := s.httpClientFor(ctx)
	var firstErr error
	for i := int64(0); i < totalChunks; i++ {
		key := fmt.Sprintf("%s%d", chunkStoragePath, i)
		url := s.s3URL("uploads", key)
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
		s.signS3(req, "UNSIGNED-PAYLOAD")
		resp, err := client.Do(req)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if firstErr == nil {
				firstErr = fmt.Errorf("delete failed %s status=%d", key, resp.StatusCode)
			}
		}
	}
	return firstErr
}

func (s *RustFSServiceImpl) GenerateImagePath(ctx context.Context, ivo *vo.GenerateImagePathVO) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	fileName := ivo.FileName()
	ext := ""
	if dot := strings.LastIndex(fileName, "."); dot != -1 {
		ext = fileName[dot+1:]
	}
	category := ivo.Category()
	if category == "" {
		category = "images"
	}
	user := strings.TrimSpace(ivo.UserUUID())
	if user == "" {
		user = "public"
	}
	name := uuid.NewString()
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s.%s", category, user, year, month, day, name, ext)
}

func (s *RustFSServiceImpl) PresignPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	raw := s.s3URL(bucket, key)
	u, err := neturl.Parse(raw)
	if err != nil || u == nil || u.Host == "" {
		return "", fmt.Errorf("invalid s3 url: %s, err=%v", raw, err)
	}
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	date := t.Format("20060102")
	scope := strings.Join([]string{date, s.region, "s3", "aws4_request"}, "/")
	qs := neturl.Values{}
	qs.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	qs.Set("X-Amz-Credential", s.access+"/"+scope)
	qs.Set("X-Amz-Date", amzDate)
	sec := int(expires.Seconds())
	if sec <= 0 || sec > 604800 {
		sec = 900
	}
	qs.Set("X-Amz-Expires", strconv.Itoa(sec))
	qs.Set("X-Amz-SignedHeaders", "host")
	canonicalQuery := canonicalizeQuery(qs.Encode())
	canonicalHeaders := "host:" + u.Host + "\n"
	cr := strings.Join([]string{http.MethodPut, u.Path, canonicalQuery, canonicalHeaders, "host", "UNSIGNED-PAYLOAD"}, "\n")
	crHash := sha256Hex([]byte(cr))
	sts := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, crHash}, "\n")
	kDate := hmacSHA256([]byte("AWS4"+s.secret), date)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	sig := hex.EncodeToString(hmacSHA256(kSigning, sts))
	qs.Set("X-Amz-Signature", sig)
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (s *RustFSServiceImpl) PresignGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 24 * time.Hour
	}
	raw := s.s3URL(bucket, key)
	u, err := neturl.Parse(raw)
	if err != nil || u == nil || u.Host == "" {
		return "", fmt.Errorf("invalid s3 url: %s, err=%v", raw, err)
	}
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	date := t.Format("20060102")
	scope := strings.Join([]string{date, s.region, "s3", "aws4_request"}, "/")
	qs := neturl.Values{}
	qs.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	qs.Set("X-Amz-Credential", s.access+"/"+scope)
	qs.Set("X-Amz-Date", amzDate)
	sec := int(expires.Seconds())
	if sec <= 0 || sec > 604800 {
		sec = 86400
	}
	qs.Set("X-Amz-Expires", strconv.Itoa(sec))
	qs.Set("X-Amz-SignedHeaders", "host")
	canonicalQuery := canonicalizeQuery(qs.Encode())
	canonicalHeaders := "host:" + u.Host + "\n"
	cr := strings.Join([]string{http.MethodGet, u.Path, canonicalQuery, canonicalHeaders, "host", "UNSIGNED-PAYLOAD"}, "\n")
	crHash := sha256Hex([]byte(cr))
	sts := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, crHash}, "\n")
	kDate := hmacSHA256([]byte("AWS4"+s.secret), date)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	sig := hex.EncodeToString(hmacSHA256(kSigning, sts))
	qs.Set("X-Amz-Signature", sig)
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (s *RustFSServiceImpl) HeadObject(ctx context.Context, bucket, key string) (int64, error) {
	client := s.httpClientFor(ctx)
	u := s.s3URL(bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	s.signS3(req, "UNSIGNED-PAYLOAD")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("head object failed: status=%d body=%s", resp.StatusCode, string(b))
	}
	cl := strings.TrimSpace(resp.Header.Get("Content-Length"))
	if cl == "" {
		return 0, fmt.Errorf("missing content-length")
	}
	size, err := strconv.ParseInt(cl, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (s *RustFSServiceImpl) s3URL(bucket, key string) string {
	k := strings.TrimLeft(key, "/")
	return fmt.Sprintf("%s/%s/%s", s.endpoint, bucket, k)
}

func (s *RustFSServiceImpl) signS3(req *http.Request, payloadHash string) {
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
	canonicalQuery := canonicalizeQuery(u.RawQuery)
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

func canonicalizeQuery(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	m, _ := neturl.ParseQuery(raw)
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, k := range keys {
		vs := m[k]
		sort.Strings(vs)
		if len(vs) == 0 {
			if !first {
				b.WriteByte('&')
			} else {
				first = false
			}
			b.WriteString(percentEncode(k))
			b.WriteByte('=')
			continue
		}
		for _, v := range vs {
			if !first {
				b.WriteByte('&')
			} else {
				first = false
			}
			b.WriteString(percentEncode(k))
			b.WriteByte('=')
			b.WriteString(percentEncode(v))
		}
	}
	return b.String()
}

func percentEncode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return b.String()
}

func (s *RustFSServiceImpl) httpClientFor(ctx context.Context) *http.Client {
	d := time.Hour
	if dl, ok := ctx.Deadline(); ok {
		dd := time.Until(dl)
		if dd > 0 {
			d = dd
		}
	}
	return &http.Client{Timeout: d}
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

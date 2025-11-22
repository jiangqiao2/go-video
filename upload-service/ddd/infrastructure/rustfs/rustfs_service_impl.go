package rustfs

import (
    "bytes"
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
    "sync"
    "time"

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
    buf, err := io.ReadAll(minIoChunkVo.Reader())
    if err != nil { return err }
    payloadHash := sha256Hex(buf)
    u := s.s3URL(minIoChunkVo.BucketName(), minIoChunkVo.StoragePath())
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(buf))
    if err != nil { return err }
    if ct := minIoChunkVo.ContentType(); ct != "" { req.Header.Set("Content-Type", ct) }
    req.ContentLength = int64(len(buf))
    req.Header.Set("x-amz-content-sha256", payloadHash)
    s.signS3(req, payloadHash)
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        logger.Errorf("rustfs put object failed", map[string]interface{}{ "status": resp.StatusCode, "body": string(b) })
        return fmt.Errorf("put object failed: status=%d", resp.StatusCode)
    }
    return nil
}

func (s *RustFSServiceImpl) MergeChunk(ctx context.Context, mergeChunkVo *vo.MergeChunkVo) error {
    tempDir := filepath.Join(os.TempDir(), "upload_service_merge")
    if err := os.MkdirAll(tempDir, 0o755); err != nil { return err }
    combined := filepath.Join(tempDir, fmt.Sprintf("%x.tmp", time.Now().UnixNano()))
    out, err := os.Create(combined)
    if err != nil { return err }
    defer func(){ out.Close(); _ = os.Remove(combined) }()

    for i := int64(0); i < mergeChunkVo.TotalChunks(); i++ {
        chunkKey := fmt.Sprintf("%s%d", mergeChunkVo.ChunkStoragePath(), i)
        url := s.s3URL("uploads", chunkKey)
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil { return err }
        req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
        s.signS3(req, "UNSIGNED-PAYLOAD")
        resp, err := http.DefaultClient.Do(req)
        if err != nil { return err }
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            b, _ := io.ReadAll(resp.Body)
            resp.Body.Close()
            logger.Errorf("rustfs get chunk failed", map[string]interface{}{ "key": chunkKey, "status": resp.StatusCode, "body": string(b) })
            return fmt.Errorf("get chunk failed: %s", chunkKey)
        }
        _, err = io.Copy(out, resp.Body)
        resp.Body.Close()
        if err != nil { return err }
    }

    f, err := os.Open(combined)
    if err != nil { return err }
    defer f.Close()

    stat, err := f.Stat()
    if err != nil { return err }
    hash, err := sha256FileHex(combined)
    if err != nil { return err }
    putURL := s.s3URL("uploads", mergeChunkVo.StoragePath())
    putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, f)
    if err != nil { return err }
    putReq.Header.Set("Content-Type", "application/octet-stream")
    putReq.Header.Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
    putReq.ContentLength = stat.Size()
    putReq.Header.Set("x-amz-content-sha256", hash)
    s.signS3(putReq, hash)
    putResp, err := http.DefaultClient.Do(putReq)
    if err != nil { return err }
    defer putResp.Body.Close()
    if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
        b, _ := io.ReadAll(putResp.Body)
        logger.Errorf("rustfs merge put failed", map[string]interface{}{ "status": putResp.StatusCode, "body": string(b) })
        return fmt.Errorf("merge put failed: status=%d", putResp.StatusCode)
    }
    return nil
}

func (s *RustFSServiceImpl) DeleteChunks(ctx context.Context, chunkStoragePath string, totalChunks int64) error {
    var firstErr error
    for i := int64(0); i < totalChunks; i++ {
        key := fmt.Sprintf("%s%d", chunkStoragePath, i)
        url := s.s3URL("uploads", key)
        req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
        if err != nil { if firstErr == nil { firstErr = err }; continue }
        req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
        s.signS3(req, "UNSIGNED-PAYLOAD")
        resp, err := http.DefaultClient.Do(req)
        if err != nil { if firstErr == nil { firstErr = err }; continue }
        resp.Body.Close()
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            if firstErr == nil { firstErr = fmt.Errorf("delete failed %s status=%d", key, resp.StatusCode) }
        }
    }
    return firstErr
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

    signed := []string{"host","x-amz-content-sha256","x-amz-date"}
    if req.Header.Get("content-type") != "" { signed = append(signed, "content-type") }
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
    if e == "" { return "http://localhost:9000" }
    if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") { return strings.TrimRight(e, "/") }
    return "http://" + strings.TrimRight(e, "/")
}

func sha256Hex(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

func sha256FileHex(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil { return "", err }
    defer f.Close()
    d := sha256.New()
    if _, err := io.Copy(d, f); err != nil { return "", err }
    return hex.EncodeToString(d.Sum(nil)), nil
}

func hmacSHA256(key []byte, data string) []byte {
    h := hmac.New(sha256.New, key)
    h.Write([]byte(data))
    return h.Sum(nil)
}
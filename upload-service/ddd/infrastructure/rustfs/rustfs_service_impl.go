package rustfs

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
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
}

func DefaultRustFSService() gateway.MinioService {
    assert.NotCircular()
    rustfsServiceOnce.Do(func() {
        r := resource.DefaultRustFSResource()
        singletonRustFS = &RustFSServiceImpl{
            endpoint: normalizeEndpoint(r.GetEndpoint()),
            access:   r.GetAccessKey(),
            secret:   r.GetSecretKey(),
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
    url := s.buildObjectURL(minIoChunkVo.StoragePath())
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, minIoChunkVo.Reader())
    if err != nil { return err }
    req.Header.Set("Content-Type", minIoChunkVo.ContentType())
    s.applyAuth(req)
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
        url := s.buildObjectURL(chunkKey)
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil { return err }
        s.applyAuth(req)
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

    putURL := s.buildObjectURL(mergeChunkVo.StoragePath())
    putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, f)
    if err != nil { return err }
    putReq.Header.Set("Content-Type", "application/octet-stream")
    s.applyAuth(putReq)
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
        url := s.buildObjectURL(key)
        req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
        if err != nil { if firstErr == nil { firstErr = err }; continue }
        s.applyAuth(req)
        resp, err := http.DefaultClient.Do(req)
        if err != nil { if firstErr == nil { firstErr = err }; continue }
        resp.Body.Close()
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            if firstErr == nil { firstErr = fmt.Errorf("delete failed %s status=%d", key, resp.StatusCode) }
        }
    }
    return firstErr
}

func (s *RustFSServiceImpl) applyAuth(req *http.Request) {
    if s.access != "" || s.secret != "" {
        req.SetBasicAuth(s.access, s.secret)
    }
}

func (s *RustFSServiceImpl) buildObjectURL(objectKey string) string {
    key := strings.TrimLeft(objectKey, "/")
    return fmt.Sprintf("%s/api/v1/objects/%s", s.endpoint, key)
}

func normalizeEndpoint(e string) string {
    e = strings.TrimSpace(e)
    if e == "" { return "http://localhost:9000" }
    if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") { return strings.TrimRight(e, "/") }
    return "http://" + strings.TrimRight(e, "/")
}
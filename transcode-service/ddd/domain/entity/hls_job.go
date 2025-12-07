package entity

import (
	"time"
	"transcode-service/ddd/domain/vo"
)

type HLSJobEntity struct {
	id             uint64
	jobUUID        string
	userUUID       string
	videoUUID      string
	sourceJobUUID  *string
	sourceType     string
	inputPath      string
	outputDir      string
	masterPlaylist *string
	status         string
	progress       int
	config         vo.HLSConfig
	errorMessage   string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewHLSJobEntity(jobUUID, userUUID, videoUUID, inputPath, outputDir string, cfg vo.HLSConfig) *HLSJobEntity {
	now := time.Now()
	return &HLSJobEntity{jobUUID: jobUUID, userUUID: userUUID, videoUUID: videoUUID, inputPath: inputPath, outputDir: outputDir, config: cfg, status: vo.HLSStatusPending.String(), progress: 0, createdAt: now, updatedAt: now}
}

func (e *HLSJobEntity) ID() uint64               { return e.id }
func (e *HLSJobEntity) JobUUID() string          { return e.jobUUID }
func (e *HLSJobEntity) UserUUID() string         { return e.userUUID }
func (e *HLSJobEntity) VideoUUID() string        { return e.videoUUID }
func (e *HLSJobEntity) SourceJobUUID() *string   { return e.sourceJobUUID }
func (e *HLSJobEntity) SourceType() string       { return e.sourceType }
func (e *HLSJobEntity) InputPath() string        { return e.inputPath }
func (e *HLSJobEntity) OutputDir() string        { return e.outputDir }
func (e *HLSJobEntity) MasterPlaylist() *string  { return e.masterPlaylist }
func (e *HLSJobEntity) Status() string           { return e.status }
func (e *HLSJobEntity) Progress() int            { return e.progress }
func (e *HLSJobEntity) CreatedAt() time.Time     { return e.createdAt }
func (e *HLSJobEntity) UpdatedAt() time.Time     { return e.updatedAt }
func (e *HLSJobEntity) GetConfig() *vo.HLSConfig { return &e.config }

func (e *HLSJobEntity) SetStatus(status vo.HLSStatus) {
	e.status = status.String()
	e.updatedAt = time.Now()
}
func (e *HLSJobEntity) SetProgress(p int) {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	e.progress = p
	e.updatedAt = time.Now()
}
func (e *HLSJobEntity) SetInputPath(path string) {
	e.inputPath = path
	e.updatedAt = time.Now()
}
func (e *HLSJobEntity) SetMasterPlaylist(path string) {
	e.masterPlaylist = &path
	e.updatedAt = time.Now()
}
func (e *HLSJobEntity) SetOutputDir(dir string) { e.outputDir = dir; e.updatedAt = time.Now() }
func (e *HLSJobEntity) SetError(msg string)     { e.errorMessage = msg; e.updatedAt = time.Now() }
func (e *HLSJobEntity) SetSource(jobUUID *string, sourceType string) {
	e.sourceJobUUID = jobUUID
	e.sourceType = sourceType
	e.updatedAt = time.Now()
}

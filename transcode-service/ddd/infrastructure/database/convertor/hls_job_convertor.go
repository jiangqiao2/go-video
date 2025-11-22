package convertor

import (
    "transcode-service/ddd/domain/entity"
    "transcode-service/ddd/domain/vo"
    "transcode-service/ddd/infrastructure/database/po"
)

type HLSJobConvertor struct{}

func NewHLSJobConvertor() *HLSJobConvertor { return &HLSJobConvertor{} }

func (c *HLSJobConvertor) ToEntity(poJob *po.HLSJob) *entity.HLSJobEntity {
    if poJob == nil { return nil }
    cfg := vo.DefaultHLSConfig()
    if poJob.ProfilesJSON != nil { _ = cfg.FromJSON(*poJob.ProfilesJSON) }
    cfg.SegmentDuration = poJob.SegmentDuration
    cfg.ListSize = poJob.ListSize
    cfg.Format = poJob.Format
    cfg.SetProgress(poJob.Progress)
    cfg.SetStatus(vo.HLSStatus(poJob.Status))
    if poJob.MasterPlaylist != nil { cfg.SetOutputPath(*poJob.MasterPlaylist) }
    e := entity.NewHLSJobEntity(poJob.JobUUID, poJob.UserUUID, poJob.VideoUUID, poJob.InputPath, poJob.OutputDir, *cfg)
    return e
}

func (c *HLSJobConvertor) ToPO(e *entity.HLSJobEntity) *po.HLSJob {
    var profiles *string
    if e.GetConfig() != nil { if json, err := e.GetConfig().ToJSON(); err == nil { profiles = &json } }
    return &po.HLSJob{
        BaseModel: po.BaseModel{Id: e.ID(), CreatedAt: e.CreatedAt(), UpdatedAt: e.UpdatedAt()},
        JobUUID:   e.JobUUID(),
        UserUUID:  e.UserUUID(),
        VideoUUID: e.VideoUUID(),
        InputPath: e.InputPath(),
        OutputDir: e.OutputDir(),
        MasterPlaylist: e.MasterPlaylist(),
        ProfilesJSON: profiles,
        Status:    e.Status(),
        Progress:  e.Progress(),
        SegmentDuration: e.GetConfig().SegmentDuration,
        ListSize:  e.GetConfig().ListSize,
        Format:    e.GetConfig().Format,
        VariantCount: e.GetConfig().GetResolutionCount(),
    }
}
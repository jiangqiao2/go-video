package vo

type VideoPrivacy struct {
	value string
}

var (
	VideoPrivacyPublic    = VideoPrivacy{"public"}
	VideoPrivacyFollowers = VideoPrivacy{"followers"}
	VideoPrivacyPrivate   = VideoPrivacy{"private"}
)

var videoPrivacySet = []VideoPrivacy{
	VideoPrivacyPublic,
	VideoPrivacyFollowers,
	VideoPrivacyPrivate,
}

func (p VideoPrivacy) Value() string { return p.value }

func (p VideoPrivacy) IsPublic() bool { return p.value == VideoPrivacyPublic.value }

func (p VideoPrivacy) IsFollowers() bool { return p.value == VideoPrivacyFollowers.value }

func (p VideoPrivacy) IsPrivate() bool { return p.value == VideoPrivacyPrivate.value }

func NewVideoPrivacy(value string) VideoPrivacy {
	for _, v := range videoPrivacySet {
		if v.value == value {
			return v
		}
	}
	return VideoPrivacyPublic
}

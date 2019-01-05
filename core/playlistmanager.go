package core

import (
	"fmt"
	"sync"

	"github.com/ericxtang/m3u8"
	"github.com/golang/glog"
	"github.com/livepeer/go-livepeer/drivers"
	ffmpeg "github.com/livepeer/lpms/ffmpeg"
)

const LIVE_LIST_LENGTH uint = 6

//	PlaylistManager manages playlists and data for one video stream, backed by one object storage.
type PlaylistManager interface {
	ManifestID() ManifestID
	// Implicitly creates master and media playlists
	// Inserts in media playlist given a link to a segment
	InsertHLSSegment(profile *ffmpeg.VideoProfile, seqNo uint64, uri string, duration float64) error

	GetHLSMasterPlaylist() *m3u8.MasterPlaylist

	GetHLSMediaPlaylist(rendition string) *m3u8.MediaPlaylist

	GetOSSession() drivers.OSSession

	Cleanup()
}

type BasicPlaylistManager struct {
	storageSession drivers.OSSession
	manifestID     ManifestID
	// Live playlist used for broadcasting
	masterPList *m3u8.MasterPlaylist
	mediaLists  map[string]*m3u8.MediaPlaylist
	mapSync     *sync.RWMutex
}

// NewBasicPlaylistManager create new BasicPlaylistManager struct
func NewBasicPlaylistManager(manifestID ManifestID,
	storageSession drivers.OSSession) *BasicPlaylistManager {

	bplm := &BasicPlaylistManager{
		storageSession: storageSession,
		manifestID:     manifestID,
		masterPList:    m3u8.NewMasterPlaylist(),
		mediaLists:     make(map[string]*m3u8.MediaPlaylist),
		mapSync:        &sync.RWMutex{},
	}
	return bplm
}

func (mgr *BasicPlaylistManager) ManifestID() ManifestID {
	return mgr.manifestID
}

func (mgr *BasicPlaylistManager) Cleanup() {
	mgr.storageSession.EndSession()
}

func (mgr *BasicPlaylistManager) GetOSSession() drivers.OSSession {
	return mgr.storageSession
}

func (mgr *BasicPlaylistManager) getPL(rendition string) *m3u8.MediaPlaylist {
	mgr.mapSync.RLock()
	mpl := mgr.mediaLists[rendition]
	mgr.mapSync.RUnlock()
	return mpl
}

func (mgr *BasicPlaylistManager) createPL(profile *ffmpeg.VideoProfile) *m3u8.MediaPlaylist {
	mpl, err := m3u8.NewMediaPlaylist(LIVE_LIST_LENGTH, LIVE_LIST_LENGTH)
	if err != nil {
		glog.Error(err)
		return nil
	}
	mgr.mapSync.Lock()
	mgr.mediaLists[profile.Name] = mpl
	mgr.mapSync.Unlock()
	vParams := ffmpeg.VideoProfileToVariantParams(*profile)
	url := fmt.Sprintf("%v/%v.m3u8", mgr.manifestID, profile.Name)
	mgr.masterPList.Append(url, mpl, vParams)
	return mpl
}

func (mgr *BasicPlaylistManager) InsertHLSSegment(profile *ffmpeg.VideoProfile, seqNo uint64, uri string,
	duration float64) error {

	mpl := mgr.getPL(profile.Name)
	if mpl == nil {
		mpl = mgr.createPL(profile)
	}
	return mgr.addToMediaPlaylist(uri, seqNo, duration, mpl)
}

func (mgr *BasicPlaylistManager) addToMediaPlaylist(uri string, seqNo uint64, duration float64,
	mpl *m3u8.MediaPlaylist) error {

	mseg := newMediaSegment(uri, seqNo, duration)
	if mpl.Count() >= mpl.WinSize() {
		mpl.Remove()
	}
	if mpl.Count() == 0 {
		mpl.SeqNo = mseg.SeqId
	}
	// XXX This probably should be using mpl.InsertSegment instead
	return mpl.AppendSegment(mseg)
}

// GetHLSMasterPlaylist ..
func (mgr *BasicPlaylistManager) GetHLSMasterPlaylist() *m3u8.MasterPlaylist {
	return mgr.masterPList
}

// GetHLSMediaPlaylist ...
func (mgr *BasicPlaylistManager) GetHLSMediaPlaylist(rendition string) *m3u8.MediaPlaylist {
	return mgr.getPL(rendition)
}

func newMediaSegment(uri string, seqNo uint64, duration float64) *m3u8.MediaSegment {
	return &m3u8.MediaSegment{
		URI:      uri,
		SeqId:    seqNo,
		Duration: duration,
	}
}

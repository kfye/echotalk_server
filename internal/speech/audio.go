package speech

import (
	"bytes"
	"encoding/binary"

	"github.com/echotalk/echotalk_server/internal/pkg/errcode"
)

// ValidateWAV 轻量校验音频为 16K 采样率 / 16bit / 单声道 WAV。
func ValidateWAV(data []byte) error {
	if len(data) < 44 {
		return errcode.ErrAudioFormat
	}
	if !bytes.Equal(data[0:4], []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WAVE")) {
		return errcode.ErrAudioFormat
	}
	numChannels := binary.LittleEndian.Uint16(data[22:24])
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])
	if numChannels != 1 || sampleRate != 16000 || bitsPerSample != 16 {
		return errcode.ErrAudioFormat
	}
	return nil
}

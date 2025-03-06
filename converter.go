package main

import (
	"bytes"
	"fmt"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
	"io"
	"math"
	"math/cmplx"
	"mime/multipart"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
)

func mp3ToWav(file multipart.File) ([]byte, error) {
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MP3: %v", err)
	}
	var wavBuffer bytes.Buffer
	enc := wav.NewEncoder(&wavBuffer, decoder.SampleRate(), 16, 1, 1)

	buf := make([]byte, 1024)
	intBuffer := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  decoder.SampleRate(),
		},
		Data: make([]int, 0, 1024),
	}

	for {

		n, err := decoder.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read MP3 data: %v", err)
		}

		for i := 0; i < n; i += 2 {
			sample := int16(buf[i]) | int16(buf[i+1])<<8
			intBuffer.Data = append(intBuffer.Data, int(sample))
		}
		if err := enc.Write(intBuffer); err != nil {
			return nil, fmt.Errorf("failed to write WAV data: %v", err)
		}
		intBuffer.Data = intBuffer.Data[:0]
	}

	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close WAV encoder: %v", err)
	}

	return wavBuffer.Bytes(), nil
}

func normalize(data []float64) []float64 {
	maxAmplitude := 0.0
	for _, sample := range data {
		if abs := math.Abs(sample); abs > maxAmplitude {
			maxAmplitude = abs
		}
	}
	if maxAmplitude == 0 {
		return data
	}
	normalized := make([]float64, len(data))
	for i, sample := range data {
		normalized[i] = sample / maxAmplitude
	}
	return normalized
}

func computeSpectrogram(data []float64, sampleRate, windowSize, hopSize int) [][]float64 {
	numFrames := (len(data) - windowSize) / hopSize
	spectrogram := make([][]float64, numFrames)

	for i := 0; i < numFrames; i++ {
		start := i * hopSize
		end := start + hopSize

		frame := make([]float64, windowSize)
		copy(frame, data[start:end])
		window.Apply(frame, window.Hann)
		fftOut := fft.FFTReal(frame)
		magnitude := make([]float64, len(fftOut)/2)
		for j := 0; j < len(fftOut); j++ {
			magnitude[j] = cmplx.Abs(fftOut[j])
		}
		spectrogram[i] = magnitude
	}
	return spectrogram
}

package preprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FeatureExtractor interface {
	Extract(ctx context.Context, inputPath string) (*FeatureExtractionOutput, error)
}

type FeatureExtractionOutput struct {
	OutputPath  string   `json:"output_path"`
	FeatureType string   `json:"feature_type"`
	Inputs      []string `json:"inputs"`
	Shape       []int    `json:"shape"`
	DType       string   `json:"dtype"`
	SampleRate  int      `json:"sample_rate"`
	NMels       int      `json:"n_mels"`
}

type PythonFeatureExtractor struct {
	pythonBin   string
	scriptPath  string
	workDir     string
	featureType string
	sampleRate  int
	nMels       int
}

func NewPythonFeatureExtractor(cfg *Config) *PythonFeatureExtractor {
	return &PythonFeatureExtractor{
		pythonBin:   cfg.PythonBin,
		scriptPath:  cfg.ScriptPath,
		workDir:     cfg.WorkDir,
		featureType: cfg.FeatureType,
		sampleRate:  cfg.SampleRate,
		nMels:       cfg.NMels,
	}
}

func (e *PythonFeatureExtractor) Extract(ctx context.Context, inputPath string) (*FeatureExtractionOutput, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "." + e.featureType + ".pt"

	cmd := exec.CommandContext(
		ctx,
		e.pythonBin,
		e.scriptPath,
		"--input", inputPath,
		"--output", outputPath,
		"--feature", e.featureType,
		"--sample-rate", strconv.Itoa(e.sampleRate),
		"--n-mels", strconv.Itoa(e.nMels),
	)
	cmd.Dir = "."
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"run feature extraction script: %w: stdout=%s stderr=%s",
			err,
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}

	var result FeatureExtractionOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &result); err != nil {
		return nil, fmt.Errorf(
			"parse feature extraction result: %w: stdout=%s stderr=%s",
			err,
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}

	return &result, nil
}

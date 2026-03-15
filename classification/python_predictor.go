package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type PredictionOutput struct {
	Label         string             `json:"label"`
	Index         int                `json:"index"`
	Score         float64            `json:"score"`
	Probabilities map[string]float64 `json:"probabilities"`
	ModelName     string             `json:"model_name"`
}

type Predictor interface {
	Predict(ctx context.Context, featurePath string) (*PredictionOutput, error)
}

type PythonPredictor struct {
	pythonBin  string
	scriptPath string
	modelPath  string
	device     string
	labels     []string
}

func NewPythonPredictor(cfg *Config) *PythonPredictor {
	return &PythonPredictor{
		pythonBin:  cfg.PythonBin,
		scriptPath: cfg.ScriptPath,
		modelPath:  cfg.ModelPath,
		device:     cfg.Device,
		labels:     cfg.Labels,
	}
}

func (p *PythonPredictor) Predict(ctx context.Context, featurePath string) (*PredictionOutput, error) {
	args := []string{
		p.scriptPath,
		"--input", featurePath,
		"--model", p.modelPath,
		"--device", p.device,
	}
	if len(p.labels) > 0 {
		args = append(args, "--labels", strings.Join(p.labels, ","))
	}

	cmd := exec.CommandContext(ctx, p.pythonBin, args...)
	cmd.Dir = "."

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"run classification script: %w: stdout=%s stderr=%s",
			err,
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}

	var result PredictionOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &result); err != nil {
		return nil, fmt.Errorf(
			"parse classification result: %w: stdout=%s stderr=%s",
			err,
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}

	return &result, nil
}

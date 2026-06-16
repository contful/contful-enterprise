// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// RuleConfig 规则配置文件结构
type RuleConfig struct {
	Rules []Rule `yaml:"rules"`
}

// Rule 单条异常检测规则
type Rule struct {
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"`
	Enabled     bool                   `yaml:"enabled"`
	Description string                 `yaml:"description"`
	Query       string                 `yaml:"query,omitempty"`
	Window      time.Duration          `yaml:"window,omitempty"`
	Threshold   int                    `yaml:"threshold,omitempty"`
	Score       float64                `yaml:"score,omitempty"`
	Analysis    map[string]bool        `yaml:"analysis,omitempty"`
	Thresholds  map[string]interface{} `yaml:"thresholds,omitempty"`
	// high_frequency
	// permission_escalation
	SensitiveActions []string `yaml:"sensitive_actions,omitempty"`
	// time_series_anomaly
	BaselineWindow  time.Duration `yaml:"baseline_window,omitempty"`
	DeviationFactor float64       `yaml:"deviation_factor,omitempty"`
	// behavior_deviation
	Features []string `yaml:"features,omitempty"`
}

// RuleEngine 规则引擎
type RuleEngine struct {
	mu        sync.RWMutex
	rules     []Rule
	rulePath  string
	watcher   *fsnotify.Watcher
	stopCh    chan struct{}
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(rulePath string) (*RuleEngine, error) {
	engine := &RuleEngine{
		rulePath: rulePath,
		stopCh:   make(chan struct{}),
	}

	if err := engine.LoadRules(); err != nil {
		return nil, err
	}

	return engine, nil
}

// StartWatch 启动文件监听（热加载）
func (e *RuleEngine) StartWatch() error {
	var err error
	e.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 监听规则文件所在目录（兼容文件替换场景）
	dir := e.rulePath
	fi, statErr := os.Stat(dir)
	if statErr != nil {
		return statErr
	}
	if fi.IsDir() {
		// 已经是目录
	} else {
		// 是文件，监听父目录
		dir = dir[:len(dir)-len(fi.Name())-1]
		if dir == "" {
			dir = "."
		}
	}

	if err := e.watcher.Add(dir); err != nil {
		return err
	}

	go e.watchLoop()

	log.Info().
		Str("module", "rule_engine").
		Str("path", e.rulePath).
		Msg("file watcher started for rule hot-reload")

	return nil
}

// StopWatch 停止文件监听
func (e *RuleEngine) StopWatch() {
	if e.watcher != nil {
		e.watcher.Close()
	}
	close(e.stopCh)
}

// watchLoop 文件监听循环
func (e *RuleEngine) watchLoop() {
	// 防抖定时器
	var debounceTimer *time.Timer
	const debounceInterval = 500 * time.Millisecond

	for {
		select {
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceInterval, func() {
					log.Info().
						Str("module", "rule_engine").
						Str("file", event.Name).
						Msg("rule file changed, reloading")
					if err := e.LoadRules(); err != nil {
						log.Error().
							Err(err).
							Str("module", "rule_engine").
							Msg("failed to reload rules")
					}
				})
			}
		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			log.Error().
				Err(err).
				Str("module", "rule_engine").
				Msg("file watcher error")
		case <-e.stopCh:
			return
		}
	}
}

// LoadRules 加载 YAML 规则文件
func (e *RuleEngine) LoadRules() error {
	data, err := os.ReadFile(e.rulePath)
	if err != nil {
		return err
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	e.mu.Lock()
	e.rules = config.Rules
	e.mu.Unlock()

	log.Info().
		Str("module", "rule_engine").
		Int("count", len(config.Rules)).
		Msg("rules loaded")

	return nil
}

// GetRules 获取当前规则列表（只返回已启用的）
func (e *RuleEngine) GetRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Rule, 0, len(e.rules))
	for _, r := range e.rules {
		if r.Enabled {
			result = append(result, r)
		}
	}
	return result
}

// GetRuleByType 根据类型获取规则
func (e *RuleEngine) GetRuleByType(ruleType string) *Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, r := range e.rules {
		if r.Type == ruleType && r.Enabled {
			return &r
		}
	}
	return nil
}
